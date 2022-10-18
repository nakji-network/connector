package compound

import (
	"context"
	"time"

	"github.com/nakji-network/connector/chain/ethereum"
	nakjicommon "github.com/nakji-network/connector/common"
	"github.com/nakji-network/connector/kafkautils"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
)

type Connector struct {
	*ethereum.Connector
	Contracts map[string]*Contract
}

func NewConnector() *Connector {
	addresses := GetAddresses(ContractAddresses)
	contracts := BuildContracts(ContractAddresses)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ec := ethereum.NewConnector(ctx, addresses, "ethereum")

	return &Connector{
		Connector: ec,
		Contracts: contracts,
	}
}

func (c *Connector) Start() {

	// Register topic and protobuf type mappings
	c.RegisterProtos(kafkautils.MsgTypeFct, protos...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Sub.Subscribe(ctx)

	for {
		select {
		//	Listen to event logs
		case evLog := <-c.Sub.Logs():
			if msg := c.parse(kafkautils.MsgTypeFct, evLog); msg != nil {
				c.EventSink <- msg
			}
		case <-c.Sub.Done():
			log.Info().Msg("connector shutdown")
			return
		case err := <-c.Sub.Err():
			log.Fatal().Err(err)
		}
	}
}

func (c *Connector) parse(msgType kafkautils.MsgType, vLog types.Log) *kafkautils.Message {
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

	bt, err := c.Sub.GetBlockTime(context.Background(), vLog)
	if err != nil {
		log.Error().Str("contract name", contractName).Err(err).Msg("Failed to retrieve timestamp")
	}
	timestamp := nakjicommon.UnixToTimestampPb(int64(bt * 1000))

	msg := eventParser.Get(abiEvent.Name, contractAbi, vLog, timestamp)
	if msg == nil {
		log.Warn().Str("event", abiEvent.Name).Msg("event is not defined")
		return nil
	}

	return &kafkautils.Message{
		MsgType:  msgType,
		ProtoMsg: msg,
	}
}
