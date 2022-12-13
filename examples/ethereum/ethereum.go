// Package ethereum follows https://goethereumbook.org/block-subscribe/ to
// subscribe to new Blocks and Transactions and writes the results to Nakji.

// It also works for other evm-compatible chains, as long as they use the ethclient.Client.
package ethereum

import (
	"context"
	"math/big"
	"strings"

	"github.com/nakji-network/connector"
	"github.com/nakji-network/connector/chain/ethereum"
	"github.com/nakji-network/connector/common"
	"github.com/nakji-network/connector/examples/ethereum/chain"
	"github.com/nakji-network/connector/kafkautils"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

type Connector struct {
	*connector.Connector // embed Nakji connector.Connector into your custom connector to get access to all its methods

	// Any additional custom connections not supported natively by Nakji, replace it as you see fit.
	// eg: client: DogecoinClient(context.Background()),
	client *ethclient.Client

	// Subsrciption module handles various tasks while listening to live data from EVM blockchains
	sub ethereum.ISubscription

	// Any additional config vars from the config yaml, using functions from Viper (https://pkg.go.dev/github.com/spf13/viper#readme-getting-values-from-viper)
	// This is namespaced via connector id (author-name-version)
	// CustomOption: c.Config.GetString("custom_option"),
}

func NewConnector() *Connector {
	c, err := connector.NewConnector()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to instantiate nakji connector")
	}

	// Read config from config yaml under `rpcs.[chain].full`
	rpcs := c.RPCMap[c.Blockchain].Full

	// go-ethereum client only supports 1 rpc connection currently, so we do this hack
	var RPCURL string
	for _, u := range rpcs {
		if strings.HasPrefix(u, "ws") {
			RPCURL = u
			break
		}
	}
	log.Info().Str("chain", c.Blockchain).Str("url", RPCURL).Msg("connecting to RPC")

	client, err := ethclient.DialContext(context.Background(), RPCURL)
	if err != nil {
		log.Fatal().Err(err).Msg("RPC connection error")
	}

	sub, err := ethereum.NewSubscription(client, c.Blockchain, nil)
	if err != nil {
		log.Fatal().Err(err).Str("chain", c.Blockchain).Msg("subscription failed")
	}

	// Register topic and protobuf type mappings
	c.RegisterProtos(kafkautils.MsgTypeFct, protos...)

	return &Connector{
		Connector: c,
		client:    client,
		sub:       sub,
	}
}

func (c *Connector) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	go c.backfill(ctx, cancel)

	//	Only subscribe to the blockchain events when it is not a backfill job
	if c.Backfill == nil {
		go c.listenBlocks(ctx, cancel)
	}

	<-ctx.Done()
	c.sub.Close()
}

func (c *Connector) backfill(ctx context.Context, cancel context.CancelFunc) {

	if c.Backfill == nil {
		return
	}

	c.RegisterProtos(kafkautils.MsgTypeBf, protos...)

	fromBlock := c.Backfill.FromBlock
	toBlock, err := c.client.BlockNumber(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get current block number")
	}

	if c.Backfill.FromBlock > 0 && c.Backfill.NumBlocks > 0 {
		lastBlock := c.Backfill.FromBlock + c.Backfill.NumBlocks

		if lastBlock < toBlock {
			toBlock = lastBlock
		}

	} else if c.Backfill.NumBlocks > 0 {
		fromBlock = toBlock - c.Backfill.NumBlocks
	}

	for fromBlock < toBlock {
		block, err := c.client.BlockByNumber(ctx, big.NewInt(int64(toBlock)))
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

func (c *Connector) listenBlocks(ctx context.Context, cancel context.CancelFunc) {
	c.sub.Subscribe(ctx)

	go c.listenCloseSignal(cancel)

	for h := range c.sub.Headers() {
		block, err := c.client.BlockByNumber(ctx, h.Number)
		if err != nil {
			block, err = c.client.BlockByHash(ctx, h.Hash())
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

func (c *Connector) listenCloseSignal(cancel context.CancelFunc) {
	select {
	//	Listen to error channel
	case err := <-c.sub.Err():
		log.Error().Err(err).Str("network", c.Blockchain).Msg("subscription failed")
		cancel()

	case <-c.sub.Done():
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
