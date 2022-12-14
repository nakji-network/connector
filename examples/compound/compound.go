package compound

import (
	"context"

	"github.com/nakji-network/connector/chain/ethereum"
	nakjicommon "github.com/nakji-network/connector/common"
	"github.com/nakji-network/connector/kafkautils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
)

type Connector struct {
	*ethereum.Connector

	Addresses  []common.Address
	Contracts  map[string]*Contract
	Blockchain string
	FromBlock  uint64
	NumBlocks  uint64
}

func NewConnector(blockchain string, fromBlock, numBlocks uint64) *Connector {
	addresses := GetAddresses(ContractAddresses)
	contracts := BuildContracts(ContractAddresses)

	ec := ethereum.NewConnector(context.Background(), addresses, blockchain)

	return &Connector{
		Connector:  ec,
		Addresses:  addresses,
		Contracts:  contracts,
		Blockchain: blockchain,
		FromBlock:  fromBlock,
		NumBlocks:  numBlocks,
	}
}

func (c *Connector) Start() {
	ctx, cancel := context.WithCancel(context.Background())

	go c.listenCloseSignal(cancel)

	c.RegisterProtos(kafkautils.MsgTypeBf, protos...)

	go c.backfill(ctx, cancel, c.FromBlock, c.NumBlocks)

	//	Only subscribe to the blockchain events when it is not a backfill job
	if c.FromBlock == 0 && c.NumBlocks == 0 {

		// Backfill last 100 blocks at every start
		go c.backfill(ctx, nil, 0, 100)

		// Listen live data
		go c.listenLogs(ctx, cancel)
	}

	<-ctx.Done()
	c.Sub.Close()
}

func (c *Connector) backfill(ctx context.Context, cancel context.CancelFunc, fromBlock, numBlocks uint64) {
	if fromBlock == 0 && numBlocks == 0 {
		return
	}

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

	var blockNumber uint64

	messages := make([]*kafkautils.Message, 0)
	if logs, err := ethereum.HistoricalEvents(ctx, c.Client, c.Addresses, startingBlock, toBlock); err == nil {
		for bfLog := range logs {

			if msg := c.parse(kafkautils.MsgTypeBf, ethereum.Log{Log: bfLog}); msg != nil {
				messages = append(messages, msg)

				if blockNumber != bfLog.BlockNumber {
					c.ProduceWithTransaction(messages)
					blockNumber = bfLog.BlockNumber
					messages = make([]*kafkautils.Message, 0)
				}
			}
		}
	}

	if cancel != nil {
		log.Info().Msg("backfill completed. shutting down connector.")
		cancel()
	}
}

func (c *Connector) listenLogs(ctx context.Context, cancel context.CancelFunc) {

	// Register topic and protobuf type mappings
	c.RegisterProtos(kafkautils.MsgTypeFct, protos...)

	c.Sub.Subscribe(ctx)

	var blockNumber uint64

	messages := make([]*kafkautils.Message, 0)
	for vLog := range c.Sub.Logs() {

		if msg := c.parse(kafkautils.MsgTypeFct, vLog); msg != nil {
			messages = append(messages, msg)

			if blockNumber != vLog.BlockNumber {
				c.ProduceWithTransaction(messages)
				blockNumber = vLog.BlockNumber
				messages = make([]*kafkautils.Message, 0)
			}
		}
	}
}

func (c *Connector) listenCloseSignal(cancel context.CancelFunc) {
	select {
	//	Listen to error channel
	case err := <-c.Sub.Err():
		log.Error().Err(err).Str("blockchain", c.Blockchain).Msg("subscription failed")
		cancel()

	case <-c.Sub.Done():
		cancel()
	}
}

func (c *Connector) parse(msgType kafkautils.MsgType, vLog ethereum.Log) *kafkautils.Message {
	address := vLog.Address.String()
	if c.Contracts[address] == nil {
		log.Info().Str("address", address).Msg("Event from unsupported address")
		return nil
	}

	contract := c.Contracts[address]

	contractAbi := contract.ABI
	contractName := contract.Name
	eventParser := contract.Pmg

	abiEvent, err := contractAbi.EventByID(vLog.Topics[0])
	if err != nil {
		log.Warn().Str("contract name", contractName).Err(err).Msg("Failed to get event from ABI")
		return nil
	}

	bt, err := c.Sub.GetBlockTime(context.Background(), vLog.Log)
	if err != nil {
		log.Error().Str("contract name", contractName).Err(err).Msg("Failed to retrieve timestamp")
	}
	timestamp := nakjicommon.UnixToTimestampPb(int64(bt * 1000))

	msg := eventParser.Get(abiEvent.Name, contractAbi, vLog.Log, timestamp)
	if msg == nil {
		log.Warn().Str("event", abiEvent.Name).Msg("event is not defined")
		return nil
	}

	return &kafkautils.Message{
		MsgType:  msgType,
		ProtoMsg: msg,
	}
}
