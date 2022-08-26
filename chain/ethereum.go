package chain

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

// Ethereum returns an ethereum client connected to websockets RPC
func (c Clients) Ethereum(ctx context.Context, chainOverride ...string) *ethclient.Client {
	chain := "ethereum"
	if len(chainOverride) > 0 && chainOverride[0] != "" {
		chain = chainOverride[0]
	}

	// Read config from config yaml under `rpcs.ethereum`
	rpcs := c.rpcMap[chain].Full

	// go-ethereum client only supports 1 rpc connection currently so we do this hack
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
	return client
}

// ChunkedSubscribeFilterLogs allows to subscribe to addresses in chunks to avoid memory issues.
// It also aggregates error log for each subscription in a common errChan error channel.
// Resulting `ethereum.Subsription` objects are returned as an array for later use by the caller.
// Default chunksize = 7350 for quiknode.
func ChunkedSubscribeFilterLogs(
	ctx context.Context,
	client *ethclient.Client,
	addresses []ethcommon.Address,
	logChan chan<- types.Log,
	errChan chan<- error,
	subs []ethereum.Subscription) (
	[]ethereum.Subscription, error) {

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
func ChunkedFilterLogs(
	ctx context.Context,
	client *ethclient.Client,
	addresses []ethcommon.Address,
	fromBlock uint64,
	toBlock uint64,
	logChan chan<- types.Log,
	failedQueries []ethereum.FilterQuery) []ethereum.FilterQuery {

	// Split filterlog queries into 7350 contracts due to json-rpc limitation (undocumented)
	const (
		addressChunkSize int    = 7350
		blockChunkSize   uint64 = 50
	)

	if failedQueries == nil {
		failedQueries = make([]ethereum.FilterQuery, 0)
	}

	if len(addresses) > addressChunkSize {
		failedQueries = ChunkedFilterLogs(ctx, client, addresses[:addressChunkSize], fromBlock, toBlock, logChan, failedQueries)
		return ChunkedFilterLogs(ctx, client, addresses[addressChunkSize:], fromBlock, toBlock, logChan, failedQueries)
	}

	if toBlock-fromBlock > blockChunkSize {
		failedQueries = ChunkedFilterLogs(ctx, client, addresses, toBlock-blockChunkSize, toBlock, logChan, failedQueries)
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
		log.Error().Err(err).Str("from", fmt.Sprint(fromBlock)).Str("to", fmt.Sprint(toBlock)).Msg("skipping failed backfill interval...")
		failedQueries = append(failedQueries, query)
	}

	for _, l := range logs {
		logChan <- l
	}
	return failedQueries
}
