// Package ethereum provides a base ethereum connector as well as some functionalities
// such as eth block headers and logs subscription etc.
//
// It also works for other evm-compatible chains as long as their ethclient implement the ETHClient interface.
//
// Users only need to embed it into their connectors in order to use it.
// See more examples at: https://github.com/nakji-network/connector/tree/main/examples

package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/nakji-network/connector"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

type ETHClient interface {
	FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error)
	SubscribeFilterLogs(context.Context, ethereum.FilterQuery, chan<- types.Log) (ethereum.Subscription, error)
}

type Connector struct {
	*connector.Connector
	Sub ISubscription
}

const chain = "ethereum"

// NewConnector returns an ethereum connector connected to websockets RPC
func NewConnector(ctx context.Context, addresses []common.Address) *Connector {
	c, err := connector.NewConnector()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate nakji connector")
	}

	// Read config from config yaml under `rpcs.ethereum.full`
	rpcs := c.RPCMap[chain].Full

	// go-ethereum client only supports 1 rpc connection currently, so we do this hack
	var RPCURL string
	for _, u := range rpcs {
		if strings.HasPrefix(u, "ws") {
			RPCURL = u
			break
		}
	}
	log.Info().
		Str("chain", chain).
		Str("url", RPCURL).
		Msg("connecting to RPC")
	client, err := ethclient.DialContext(ctx, RPCURL)
	if err != nil {
		log.Fatal().Err(err).Msg("RPC connection error")
	}

	sub, err := NewSubscription(client, chain, addresses)
	if err != nil {
		log.Fatal().Err(err).Str("chain", chain).Msg("subscription error")
	}

	return &Connector{
		c,
		sub,
	}
}

// ChunkedSubscribeFilterLogs allows to subscribe to addresses in chunks to avoid memory issues.
// It also aggregates error log for each subscription in a common errChan error channel.
// Resulting `ethereum.Subsription` objects are returned as an array for later use by the caller.
// Default chunksize = 7350 for quiknode.
func ChunkedSubscribeFilterLogs(ctx context.Context, client ETHClient, addresses []common.Address, logChan chan<- types.Log, errChan chan<- error, subs []ethereum.Subscription) ([]ethereum.Subscription, error) {
	if subs == nil {
		subs = make([]ethereum.Subscription, 0)
	}

	// Split filterlog subscriptions into 7350 contracts due to json-rpc limitation (undocumented)
	const chunkSize int = 7350

	if len(addresses) > chunkSize {
		s, err := ChunkedSubscribeFilterLogs(ctx, client, addresses[:chunkSize], logChan, errChan, subs)
		if err != nil {
			return nil, err
		}
		return ChunkedSubscribeFilterLogs(ctx, client, addresses[chunkSize:], logChan, errChan, s)
	}

	q := ethereum.FilterQuery{
		Addresses: addresses,
	}

	sub, err := client.SubscribeFilterLogs(ctx, q, logChan)
	if err != nil {
		return nil, err
	}

	//	Aggregate errors in one channel
	go func() {
		errChan <- <-sub.Err()
	}()

	subs = append(subs, sub)

	return subs, nil
}

//	ChunkedFilterLogs queries the blockchain for past events in batches.
//	It slices addresses and total number of blocks with pre-defined batch size.
//	The results are later fed into a log chan that was provided by the caller.
//	Failed query intervals are fed into another channel to allow the caller to retry later.
func ChunkedFilterLogs(ctx context.Context, client ETHClient, addresses []common.Address, fromBlock, toBlock uint64, logChan chan<- types.Log, failedQueries []ethereum.FilterQuery) ([]ethereum.FilterQuery, error) {
	// Split filterlog queries into 7350 contracts due to json-rpc limitation (undocumented)
	const (
		addressChunkSize int    = 7350
		blockChunkSize   uint64 = 50
	)

	var err error

	if failedQueries == nil {
		failedQueries = make([]ethereum.FilterQuery, 0)
	}

	if len(addresses) > addressChunkSize {
		failedQueries, err = ChunkedFilterLogs(ctx, client, addresses[:addressChunkSize], fromBlock, toBlock, logChan, failedQueries)
		if err != nil {
			return nil, err
		}
		return ChunkedFilterLogs(ctx, client, addresses[addressChunkSize:], fromBlock, toBlock, logChan, failedQueries)
	}

	if toBlock-fromBlock > blockChunkSize {
		failedQueries, err = ChunkedFilterLogs(ctx, client, addresses, toBlock-blockChunkSize, toBlock, logChan, failedQueries)
		if err != nil {
			return nil, err
		}
		return ChunkedFilterLogs(ctx, client, addresses, fromBlock, toBlock-blockChunkSize-1, logChan, failedQueries)
	}

	log.Debug().Str("from", fmt.Sprint(fromBlock)).Str("to", fmt.Sprint(toBlock)).Msg("retrieving historical events...")

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: addresses,
	}

	logs, err := client.FilterLogs(ctx, query)
	if err != nil {
		if strings.Contains(err.Error(), "read limit") {
			// error -> websocket: read limit exceeded

			log.Error().Err(err).Msg("hit RPC rate limit")
			return nil, err
		}
		log.Error().Err(err).Str("from", fmt.Sprint(fromBlock)).Str("to", fmt.Sprint(toBlock)).Msg("skipping failed backfill interval...")
		failedQueries = append(failedQueries, query)
	}

	for _, l := range logs {
		logChan <- l
	}

	return failedQueries, err
}

//	Backfill queries past blocks for the events emitted by the given contract addresses and feeds these events into the event log chan.
func Backfill(ctx context.Context, client *ethclient.Client, addresses []common.Address, logs chan types.Log, fromBlock uint64, toBlock uint64) error {
	defer close(logs)

	if fromBlock == toBlock {
		return nil
	}

	if toBlock == 0 {
		var err error
		toBlock, err = client.BlockNumber(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to get block number")
		}
	}

	if fromBlock >= toBlock {
		return nil
	}

	//	Store failed queries for retry
	failedQueries, err := ChunkedFilterLogs(ctx, client, addresses, fromBlock, toBlock, logs, nil)
	if err != nil {
		return err
	}

	for _, q := range failedQueries {
		//	Retry failed queries one more time
		fq, err := ChunkedFilterLogs(ctx, client, q.Addresses, q.FromBlock.Uint64(), q.ToBlock.Uint64(), logs, nil)
		if err != nil {
			return err
		}
		for _, q2 := range fq {
			log.Error().Str("from", fmt.Sprint(q2.FromBlock)).Str("to", fmt.Sprint(q2.ToBlock)).Msg("aborting failed backfill interval.")
		}
	}
	return nil
}