package chain

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/nakji-network/connector/common"

	"github.com/ethereum/go-ethereum"
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

// Subscribe pairs in address chunks. Default chunksize = 7350 for quiknode.
func ChunkedSubscribeFilterLogs(ctx context.Context, client *ethclient.Client, filterQuery ethereum.FilterQuery, ch chan<- types.Log, chunksize int) (
	[]ethereum.Subscription, <-chan error, error) {
	// Split filterlog subscriptions into 7350 contracts due to json-rpc limitation (undocumented)
	if chunksize == 0 {
		chunksize = 7350
	}

	queries := []ethereum.FilterQuery{}

	for i := 0; i < len(filterQuery.Addresses); i += chunksize {
		chunkQuery := filterQuery
		upperBound := i + chunksize
		if upperBound > len(filterQuery.Addresses) {
			upperBound = len(filterQuery.Addresses)
		}
		chunkQuery.Addresses = filterQuery.Addresses[i:upperBound]
		queries = append(queries, chunkQuery)
	}

	subs := make([]ethereum.Subscription, len(queries))
	errcs := make([]<-chan error, len(queries))
	for i, query := range queries {
		sub, err := client.SubscribeFilterLogs(ctx, query, ch)
		if err != nil {
			return nil, nil, err
		}
		subs[i] = sub
		errcs[i] = sub.Err()
		log.Info().
			Int("addresses", len(query.Addresses)).
			Int("totalAddresses", len(filterQuery.Addresses)).
			Msg("Listening to logs")
	}

	return subs, common.MergeErrChans(errcs...), nil
}

//	ChunkedFilterLogs queries the blockchain for past events in batches.
//	It slices addresses and total number of blocks with pre-defined batch size.
//	The results are later fed into a log chan that was provided by the caller.
//	Any occcuring error is also fed into an error chan that was provided by the caller.
func ChunkedFilterLogs(ctx context.Context, client *ethclient.Client, q ethereum.FilterQuery, latestBlockNumber int64, backFillNumBlocks int64, logch chan<- types.Log, errch chan<- error) {

	const (
		addressChunkSize       = 7350
		blockChunkSize   int64 = 50
	)

	startBlockNumber := latestBlockNumber - backFillNumBlocks

	toBlock := latestBlockNumber - 1
	fromBlock := toBlock - backFillNumBlocks
	if backFillNumBlocks > blockChunkSize {
		fromBlock = toBlock - blockChunkSize
	}

	// Address chunks move forward; block chunks move backwards to get latest blocks first.
	go func() {
		for toBlock > fromBlock {
			log.Info().Str("from", fmt.Sprint(fromBlock)).Str("to", fmt.Sprint(toBlock)).Msg("Retrieving historical events")

			startIndex := 0
			endIndex := startIndex + len(q.Addresses)
			if addressChunkSize < len(q.Addresses) {
				endIndex = startIndex + addressChunkSize
			}

			for startIndex < len(q.Addresses) {
				query := ethereum.FilterQuery{
					FromBlock: big.NewInt(fromBlock),
					ToBlock:   big.NewInt(toBlock),
					Addresses: q.Addresses[startIndex:endIndex],
				}

				logs, err := client.FilterLogs(ctx, query)
				if err != nil {
					errch <- err
				}

				for _, l := range logs {
					if l.Removed {
						continue
					}
					logch <- l
				}

				startIndex = endIndex
				endIndex += addressChunkSize
				if endIndex > len(q.Addresses) {
					endIndex = len(q.Addresses)
				}
			}

			toBlock = fromBlock - 1
			fromBlock = fromBlock - blockChunkSize
			if fromBlock < startBlockNumber {
				fromBlock = startBlockNumber
			}
		}
	}()
}
