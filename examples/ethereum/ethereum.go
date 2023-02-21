// Package ethereum follows https://goethereumbook.org/block-subscribe/ to
// subscribe to new Blocks and Transactions and writes the results to Nakji.

// It also works for other evm-compatible chains, as long as they use the ethclient.Client.
package ethereum

import (
	"context"
	"math/big"

	"github.com/nakji-network/connector/chain/ethereum"
	"github.com/nakji-network/connector/common"
	"github.com/nakji-network/connector/examples/ethereum/chain"
	"github.com/nakji-network/connector/kafkautils"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
)

type Connector struct {
	*ethereum.Connector // embed Nakji ethereum.Connector into your custom connector to get access to all its methods

	Blockchain string

	// Parameters for historical data
	FromBlock uint64
	NumBlocks uint64

	// Any additional config vars from the config yaml, using functions from Viper (https://pkg.go.dev/github.com/spf13/viper#readme-getting-values-from-viper)
	// This is namespaced via connector id (author-name-version)
	// CustomOption: c.Config.GetString("custom_option"),
}

func NewConnector(blockchain string, fromBlock, numBlocks uint64) *Connector {

	ec := ethereum.NewConnector(context.Background(), nil, blockchain)

	return &Connector{
		Connector: ec,
		FromBlock: fromBlock,
		NumBlocks: numBlocks,
	}
}

func (c *Connector) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	go c.listenCloseSignal(cancel)

	go c.backfill(ctx, cancel, c.FromBlock, c.NumBlocks)

	//	Only subscribe to the blockchain events when it is not a backfill job
	if c.FromBlock == 0 && c.NumBlocks == 0 {

		// Backfill last 100 blocks at every start
		go c.backfill(ctx, nil, 0, 100)

		// Listen live data
		go c.listenBlocks(ctx, cancel)
	}

	<-ctx.Done()
	c.Sub.Close()
}

// backfill queries for historical data and pushes them to Kafka.
func (c *Connector) backfill(ctx context.Context, cancel context.CancelFunc, fromBlock, numBlocks uint64) {
	if fromBlock == 0 && numBlocks == 0 {
		return
	}

	// Calculate block interval for historical data
	startingBlock := fromBlock
	toBlock, err := c.Client.BlockNumber(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get current block number")
	}

	if fromBlock > 0 && numBlocks > 0 {
		lastBlock := fromBlock + numBlocks

		if lastBlock < toBlock {
			toBlock = lastBlock
		}

	} else if numBlocks > 0 {
		startingBlock = toBlock - numBlocks
	}

	for startingBlock < toBlock {
		block, err := c.Client.BlockByNumber(ctx, big.NewInt(int64(toBlock)))
		if err != nil {
			log.Error().Err(err).Msg("failed to retrieve block")
		}

		err = c.process(block)
		if err != nil {
			log.Error().Err(err).Uint64("block", block.Number().Uint64()).Msg("failed to process block")
		}
		toBlock--
	}

	if cancel != nil {
		log.Info().Msg("backfill completed. shutting down connector.")
		cancel()
	}
}

// listenBlocks subscribes to live data and pushes incoming logs to Kafka.
func (c *Connector) listenBlocks(ctx context.Context, cancel context.CancelFunc) {
	// Register topic and protobuf type mappings
	c.RegisterProtos(kafkautils.MsgTypeFct, protos...)

	c.Sub.Subscribe(ctx)

	for h := range c.Sub.Headers() {
		block, err := c.Client.BlockByNumber(ctx, h.Number)
		if err != nil {
			block, err = c.Client.BlockByHash(ctx, h.Hash())
			if err != nil {
				log.Error().Err(err).Msg("failed to retrieve block")
				continue
			}
		}

		err = c.process(block)
		if err != nil {
			log.Error().Err(err).Uint64("block", block.Number().Uint64()).Msg("failed to process block")
		}
	}
}

// listenCloseSignal signals the program to terminate.
func (c *Connector) listenCloseSignal(cancel context.CancelFunc) {
	select {
	//	Listen to error channel
	case err := <-c.Sub.Err():
		log.Error().Err(err).Str("network", c.Blockchain).Msg("subscription failed")
		cancel()

	case <-c.Sub.Done():
		cancel()
	}
}

func (c *Connector) process(block *types.Block) error {

	header := block.Header()
	messages := make([]*kafkautils.Message, len(block.Transactions())+1)

	for i, t := range block.Transactions() {
		ts := common.UnixToTimestampPb(int64(header.Time))
		messages[i] = &kafkautils.Message{
			MsgType:  kafkautils.MsgTypeFct,
			ProtoMsg: chain.ParseTransaction(t, ts),
		}
	}

	messages[len(messages)-1] = &kafkautils.Message{
		MsgType:  kafkautils.MsgTypeFct,
		ProtoMsg: chain.ParseHeader(header),
	}

	return c.ProduceWithTransaction(messages)
}
