package connector

import (
	"context"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/nakji-network/connector/chain"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rs/zerolog/log"
)

type ISubscription interface {
	Done() <-chan bool
	GetBlockTime(types.Log) (uint64, error)
	Err() <-chan error
	Headers() chan *types.Header
	Logs() chan types.Log
	Unsubscribe()
}

type Subscription struct {
	addresses         []common.Address
	backfillOnce      sync.Once
	cache             *lru.Cache
	client            *ethclient.Client
	context           context.Context
	connector         *Connector
	done              chan bool
	errchan           chan error
	fromBlock         uint64
	headers           chan *types.Header
	interrupt         chan os.Signal
	isHeaderRequired  bool
	latestBlockNumber *big.Int
	logs              chan types.Log

	//	Blockchain network to connect to
	network string

	//	Number of blocks from latest block number to retrieve historical events
	numBlocks uint64
}

const allowedBlocksBehind uint64 = 60

//	NewSubscription	connects to given endpoints and subscribes to blockchain.
func NewSubscription(ctx context.Context, connector *Connector, network string, addresses []common.Address, fromBlock uint64, numBlocks uint64) (*Subscription, error) {
	s := Subscription{
		addresses:        addresses,
		done:             make(chan bool, 1),
		client:           connector.ChainClients.Ethereum(ctx, network),
		connector:        connector,
		context:          ctx,
		errchan:          make(chan error, 1),
		fromBlock:        fromBlock,
		headers:          make(chan *types.Header),
		interrupt:        make(chan os.Signal, 1),
		isHeaderRequired: false,
		logs:             make(chan types.Log),
		network:          network,
		numBlocks:        numBlocks,
	}

	//	Create cache for storing block timestamp
	cache, err := lru.New(1280)
	if err != nil {
		return nil, err
	}
	s.cache = cache

	signal.Notify(s.interrupt, os.Interrupt)

	go func() {
		select {
		case <-s.interrupt:
			s.Unsubscribe()
		case <-s.context.Done():
			s.Unsubscribe()
		}
	}()

	go s.subscribeHeaders()
	go s.subscribeLogs()

	return &s, nil
}

//	Unsubscribe closes subscriptions and open channels.
func (s *Subscription) Unsubscribe() {
	log.Info().Str("network", s.network).Msg("shutting down subscription")
	s.done <- true
	close(s.headers)
	close(s.logs)
}

// GetBlockTime retrieves block time from cache.
func (s *Subscription) GetBlockTime(vLog types.Log) (uint64, error) {
	hash := vLog.BlockHash
	val, hit := s.cache.Get(hash.Hex())
	if !hit {
		ts, err := s.getBlockTimeFromChain(hash)
		if err != nil {
			return 0, err
		}
		s.cache.Add(hash.Hex(), ts)
		return ts, nil
	}
	return val.(uint64), nil
}

func (s *Subscription) Done() <-chan bool {
	return s.done
}

func (s *Subscription) Err() <-chan error {
	return s.errchan
}

func (s *Subscription) Headers() chan *types.Header {
	s.isHeaderRequired = true
	return s.headers
}

func (s *Subscription) Logs() chan types.Log {
	return s.logs
}

//	getBlockTimeFromChain queries the blockchain and retrieves block time.
func (s *Subscription) getBlockTimeFromChain(blockHash common.Hash) (uint64, error) {
	header, err := s.client.HeaderByHash(s.context, blockHash)
	if err != nil {
		if header != nil {
			log.Error().Err(err).Uint64("block", header.Number.Uint64()).Msg("failed to retrieve header")
		}
		return 0, err
	}
	return header.Time, nil
}

//	subscribeHeaders subscribes each websocket client to block headers and extracts block time as each header is received.
func (s *Subscription) subscribeHeaders() {
	log.Info().Str("network", s.network).Msg("subscribing to headers..")

	headers := make(chan *types.Header)
	hs, err := s.client.SubscribeNewHead(s.context, headers)
	if err != nil {
		log.Error().Err(err).Msg("failed to subscribe to block headers")
		s.errchan <- err
		return
	}
	defer hs.Unsubscribe()
	defer close(headers)

	for {
		select {
		case err := <-hs.Err():
			log.Error().Err(err).Msg("header subscription failed")

			if isRetryable(err) {
				s.subscribeHeaders()
			} else {
				s.errchan <- err
			}
			return

		case header := <-headers:

			if s.latestBlockNumber == nil {
				s.latestBlockNumber = header.Number
				s.backfillOnce.Do(func() {
					ok := s.backfill()
					if !ok {
						log.Fatal().Msg("latest block number not found")
					}
				})
			}

			blockNumber := header.Number.Uint64()
			diff := blockNumber - s.latestBlockNumber.Uint64()
			if diff > 0 && diff > allowedBlocksBehind {
				s.subscribeHeaders()
			} else if blockNumber > s.latestBlockNumber.Uint64() {
				log.Debug().Str("block", header.Number.String()).Str("network", s.network).Uint64("ts", header.Time).Msg("header received")
				s.latestBlockNumber = header.Number
			}

			s.cache.ContainsOrAdd(header.Hash().Hex(), header.Time)
			if s.isHeaderRequired {
				s.headers <- header
			}
		}
	}
}

//	subscribeHeaders subscribes each websocket client to block headers and extracts block time as each header is received.
func (s *Subscription) subscribeLogs() {
	log.Info().Str("network", s.network).Msg("subscribing to event logs..")

	q := ethereum.FilterQuery{
		Addresses: s.addresses,
	}

	logch := make(chan types.Log)
	subs, subErrChan, err := chain.ChunkedSubscribeFilterLogs(s.context, s.client, q, logch, 0)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe to event logs")
	}
	for _, sub := range subs {
		defer sub.Unsubscribe()
	}
	defer close(logch)

	tWait := time.Second
	tMin := time.Second
	tMax := time.Second * 16

	for {
		select {
		case <-s.Done():
			s.done <- true
			return

		case err = <-subErrChan:
			log.Error().Err(err).Msg("event log subscription failed")

			if isRetryable(err) {
				s.subscribeLogs()
			} else {
				s.errchan <- err
			}
			return

		case vLog := <-logch:
			_, err := s.GetBlockTime(vLog)
			for err != nil {
				log.Debug().Uint64("block", vLog.BlockNumber).Str("network", s.network).Msg("waiting for block timestamp")
				time.Sleep(tWait)
				tWait *= 2
				_, err = s.GetBlockTime(vLog)
				if tWait > tMax {
					log.Warn().Uint64("block", vLog.BlockNumber).Str("network", s.network).Msg("block timestamp not available")
					break
				}
			}
			tWait = tMin
			s.logs <- vLog
		}
	}
}

//	backfill queries past blocks for the events emitted by the given contract addresses and feeds these events into the event log chan
func (s *Subscription) backfill() bool {
	if s.latestBlockNumber == nil {
		return false
	}

	if s.fromBlock == 0 && s.numBlocks == 0 {
		return true
	}

	var numBlocks uint64
	if s.numBlocks == 0 {
		if s.latestBlockNumber.Uint64() > s.fromBlock {
			numBlocks = s.latestBlockNumber.Uint64() - s.fromBlock
		} else {
			numBlocks = s.latestBlockNumber.Uint64()
		}
	} else {
		numBlocks = s.numBlocks
	}

	query := ethereum.FilterQuery{Addresses: s.addresses}
	errchan := make(chan error)
	chain.ChunkedFilterLogs(s.context, s.client, query, s.latestBlockNumber.Int64(), int64(numBlocks), s.logs, errchan)

	return true
}

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
