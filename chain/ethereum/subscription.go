//	It takes care of most of the heavy lifting interacting with a blockchain chain.
//	It connects to the chain for real-time data and provides the results through channels.
//	It also handles websocket disconnections

package ethereum

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/nakji-network/connector/kafkautils"
	"github.com/nakji-network/connector/monitor"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
)

const (
	spanName      = "connector -> kafka"
	blockSpanName = "block time -> connector"
	eventName     = "receive event log"
)

type ISubscription interface {
	AddAddress(common.Address, kafkautils.MsgType)
	Done() <-chan bool
	Err() <-chan error
	GetBlockTime(context.Context, types.Log) (uint64, error)
	TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, error)
	TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error)
	Headers() <-chan *types.Header
	Logs() <-chan Log
	Subscribe(context.Context)
	Close()
}

type Subscription struct {
	interrupt   chan os.Signal //	Shutdown signal for connector
	done        chan bool      //	Channel to signal ongoing subscriptions
	resubscribe chan bool      //	Channel to signal for resubscribing to logs

	//	Blockchain chain
	chain     string
	addresses []common.Address
	client    *ethclient.Client

	//	Network subscription
	headers           chan *types.Header //	Block header channel
	inLogs            chan types.Log     //	Event logs coming from the chain
	outLogs           chan Log           //	Event logs pushed to the connector
	inErr             chan error         //	Aggregate channel for errors coming from the chain
	outErr            chan error         //	Aggregate channel for errors sent to connector
	cache             *lru.Cache         //	Store timestamps for block numbers
	txCache           *lru.Cache         // Store transactions for hashes
	isHeaderRequired  bool               //	Flag to release block headers, if user wants them
	latestBlockNumber *big.Int
}

type Log struct {
	types.Log
	Span    trace.Span
	Baggage baggage.Baggage
}

// NewSubscription	connects to given endpoints and subscribes to blockchain.
func NewSubscription(client *ethclient.Client, chain string, addresses []common.Address) (*Subscription, error) {
	s := Subscription{
		addresses:        addresses,
		client:           client,
		chain:            chain,
		done:             make(chan bool, 1),
		interrupt:        make(chan os.Signal, 1),
		resubscribe:      make(chan bool, 1),
		inErr:            make(chan error, 1),
		outErr:           make(chan error, 1),
		isHeaderRequired: false,
	}

	//	Create cache for storing block timestamp
	cache, err := lru.New(1280)
	if err != nil {
		return nil, err
	}
	s.cache = cache

	// Create cache for storing transactions
	txCache, err := lru.New(1280)
	if err != nil {
		return nil, err
	}
	s.txCache = txCache

	signal.Notify(s.interrupt, os.Interrupt)

	go func() {
		<-s.interrupt
		s.done <- true
	}()

	return &s, nil
}

func (s *Subscription) Subscribe(ctx context.Context) {
	log.Info().Msg("subscribing to headers and logs")

	s.headers = make(chan *types.Header, 10)
	s.inLogs = make(chan types.Log, 10000)
	s.outLogs = make(chan Log, 10000)

	go s.subscribeHeaders(ctx)
	go s.subscribeLogs(ctx)
}

func (s *Subscription) AddAddress(address common.Address, msgType kafkautils.MsgType) {
	log.Debug().Str("address", address.Hex()).Msg("adding new address")
	s.addresses = append(s.addresses, address)
	if msgType == kafkautils.MsgTypeFct {
		s.resubscribe <- true
	}
}

func (s *Subscription) Done() <-chan bool {
	return s.done
}

func (s *Subscription) Err() <-chan error {
	return s.outErr
}

// GetBlockTime retrieves block time from cache.
func (s *Subscription) GetBlockTime(ctx context.Context, vLog types.Log) (uint64, error) {
	hash := vLog.BlockHash
	val, hit := s.cache.Get(hash.Hex())
	if !hit {
		ts, err := s.getBlockTimeFromChain(ctx, hash)
		if err != nil {
			return 0, err
		}
		s.cache.Add(hash.Hex(), ts)
		return ts, nil
	}
	return val.(uint64), nil
}

func (s *Subscription) Headers() <-chan *types.Header {
	s.isHeaderRequired = true
	return s.headers
}

func (s *Subscription) Logs() <-chan Log {
	return s.outLogs
}

// Close closes subscriptions and open channels.
func (s *Subscription) Close() {
	log.Info().Str("chain", s.chain).Msg("shutting down subscription")
	s.done <- true
	if s.headers != nil {
		close(s.headers)
	}
	if s.inLogs != nil {
		close(s.inLogs)
		close(s.outLogs)
	}
	close(s.inErr)
	close(s.outErr)
}

// getBlockTimeFromChain queries the blockchain and retrieves block time.
func (s *Subscription) getBlockTimeFromChain(ctx context.Context, blockHash common.Hash) (uint64, error) {
	header, err := s.client.HeaderByHash(ctx, blockHash)
	if err != nil {
		return 0, err
	}
	return header.Time, nil
}

// subscribeHeaders subscribes each websocket client to block headers and extracts block time as each header is received.
func (s *Subscription) subscribeHeaders(ctx context.Context) {
	log.Debug().Str("chain", s.chain).Msg("subscribing to headers..")

	headers := make(chan *types.Header)
	hs, err := s.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		log.Error().Err(err).Msg("failed to subscribe to block headers")
		s.outErr <- err
		return
	}
	defer hs.Unsubscribe()
	defer close(headers)

	for {
		select {
		case <-s.Done():
			s.done <- true
			return

		case err := <-hs.Err():
			if isRetryable(err) {
				log.Warn().Err(err).Msg("header subscription failed, retrying...")
				go s.subscribeHeaders(ctx)
			} else {
				log.Error().Err(err).Msg("header subscription failed")
				s.outErr <- err
			}
			return

		case header := <-headers:
			//	Start a backfill when there are missing blocks
			if s.latestBlockNumber != nil && header.Number.Uint64()-s.latestBlockNumber.Uint64() > 1 {
				go func(c context.Context, cl *ethclient.Client, adr []common.Address, fromBl, toBl uint64) {
					if backfillLogs, err := HistoricalEvents(c, cl, adr, fromBl, toBl); err == nil {
						for bfLog := range backfillLogs {
							s.inLogs <- bfLog
						}
					}
				}(ctx, s.client, s.addresses, s.latestBlockNumber.Uint64(), header.Number.Uint64())
			}

			s.cache.ContainsOrAdd(header.Hash().Hex(), header.Time)
			if s.isHeaderRequired {
				s.headers <- header
			}

			log.Debug().Str("block", header.Number.String()).Str("chain", s.chain).Uint64("ts", header.Time).Msg("header received")
			s.latestBlockNumber = header.Number
		}
	}
}

// subscribeHeaders subscribes each websocket client to block headers and extracts block time as each header is received.
func (s *Subscription) subscribeLogs(ctx context.Context) {
	log.Debug().Str("chain", s.chain).Msg("subscribing to event logs..")

	subs, err := ChunkedSubscribeFilterLogs(ctx, s.client, s.addresses, s.inLogs, s.inErr, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe to event logs")
	}
	for _, sub := range subs {
		defer sub.Unsubscribe()
	}

	tWait := time.Second
	tMin := time.Second
	tMax := time.Second * 16

	for {
		select {
		case <-s.resubscribe:
			go s.subscribeLogs(ctx)
			return

		case <-s.Done():
			s.done <- true
			return

		case err = <-s.inErr:
			if isRetryable(err) {
				log.Warn().Err(err).Msg("event log subscription failed, retrying...")
				go s.subscribeLogs(ctx)
			} else {
				log.Error().Err(err).Msg("event log subscription failed")
				s.outErr <- err
			}
			return

		case vLog := <-s.inLogs:
			// Time that logs are received. This is the end time for RPC latency
			rcvTime := time.Now()
			// Add rcvTime to baggage as connector latency observation
			spanCtx, bag := monitor.NewLatencyBaggage(ctx, monitor.LatencyConnectorKey, rcvTime)

			// Connector latency span begin
			tr := monitor.CreateTracer(monitor.DefaultTracerName)
			spanCtx, span := monitor.StartSpan(
				spanCtx,
				tr,
				spanName,
				trace.WithSpanKind(trace.SpanKindProducer),
			)
			span.AddEvent(eventName)

			blockTime, err := s.GetBlockTime(ctx, vLog)
			for err != nil {
				log.Debug().Uint64("block", vLog.BlockNumber).Str("chain", s.chain).Msg("waiting for block timestamp")
				time.Sleep(tWait)
				tWait *= 2
				blockTime, err = s.GetBlockTime(ctx, vLog)
				if tWait > tMax {
					log.Warn().Uint64("block", vLog.BlockNumber).Str("chain", s.chain).Msg("block timestamp not available")
					break
				}
			}
			tWait = tMin

			// RPC latency span - derived from block time & rcvTime
			if blockTime > 0 {
				spanStart := time.UnixMilli(int64(blockTime * 1000))
				_, rpcSpan := monitor.StartSpan(
					spanCtx,
					tr,
					blockSpanName,
					trace.WithTimestamp(spanStart),
				)
				rpcSpan.End(trace.WithTimestamp(rcvTime))
				// Add block time to baggage as rpc latency observation
				spanCtx, bag = monitor.NewLatencyBaggage(spanCtx, monitor.LatencyRpcKey, spanStart)
			}

			s.outLogs <- Log{vLog, span, bag}
		}
	}
}

// isRetryable checks the websocket disconnection error to see if connector can recover.
func isRetryable(err error) bool {
	// error 1: Message timed out
	// error 2: Connection reset by peer
	// error 3: websocket: close 1006 (abnormal closure)
	// error 4: unexpected EOF
	// error 5: websocket: close 1001 (going away): upstream went away
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "timed") ||
		strings.Contains(err.Error(), "reset") ||
		strings.Contains(err.Error(), "1006") ||
		strings.Contains(err.Error(), "EOF") ||
		strings.Contains(err.Error(), "1001")
}

func (s *Subscription) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, error) {
	if tx, ok := s.txCache.Get(hash.Hex()); ok {
		return tx.(*types.Transaction), nil
	}
	tx, _, err := s.client.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	s.txCache.Add(hash.Hex(), tx)
	return tx, nil
}

func (s *Subscription) TransactionSender(ctx context.Context, tx *types.Transaction, block common.Hash, index uint) (common.Address, error) {
	return s.client.TransactionSender(ctx, tx, block, index)
}
