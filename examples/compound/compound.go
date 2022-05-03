package compound

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/nakji-network/connector/protoregistry"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nakji-network/connector"
	"github.com/nakji-network/connector/examples/compound/ctoken"
)

type Connector struct {
	*connector.Connector
	ContractAddresses []common.Address
	Chain             string // chain override, since ChainClients.Ethereum supports overriding with any evm chain
}

const (
	namespace = "compound"
)

func (c *Connector) Start() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	logs := make(chan types.Log)
	defer close(logs)

	// Initialize CEther ABI for reading logs
	contractAbi, err := abi.JSON(strings.NewReader(ctoken.CompoundABI))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read CEther abi")
	}

	// Register topic and protobuf type mappings
	tt := c.buildTopicTypes()
	err = c.ProtoRegistryCli.RegisterDynamicTopics(tt, c.MsgType)
	if err != nil {
		log.Error().Err(err).Msg("failed to register dynamic topics")
	}

	// Get the initialized Ethereum client. For more Nakji supported clients see connector/chain/
	client := c.ChainClients.Ethereum(context.Background(), c.Chain)

	// TODO: Subscribe to headers and store timestamps
	headers := make(chan *types.Header)
	_, err = client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe headers")
	}

	sub := c.CEtherLogsListener(client, logs)

	for {
		select {
		case header := <-headers:
			log.Debug().
				Str("block", header.Number.String()).
				Uint64("ts", header.Time).
				Msg("header received")
		case err = <-sub.Err():
			log.Fatal().Err(err)
		case evLog := <-logs:
			msg, err := c.ProcessLogEvent(contractAbi, evLog)
			if err != nil {
				log.Error().Err(err).Msg("failed to process log event")
			}
			// Not sure what value needs to be passed as subject
			err = c.ProduceMessage(namespace, evLog.Address.Hex(), msg)
			if err != nil {
				log.Error().Err(err).Msg("Kafka write proto")
			}
			// Commit Kafka Transaction
			err = c.Producer.CommitTransaction(nil)
			if err != nil {
				log.Error().Err(err).Msg("Processor: Failed to commit transaction")

				err = c.Producer.AbortTransaction(nil)
				if err != nil {
					log.Fatal().Err(err).Msg("")
				}
			}
			log.Info().Interface("msg", msg).Msg("message delivered")
			// Start a new transaction
			err = c.BeginTransaction()
			if err != nil {
				log.Fatal().Err(err).Msg("")
			}
		case <-interrupt:
			log.Info().Msg("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			client.Close()
			c.Producer.Close()
			return
		}
	}

}

// CEtherLogsListener listens to all contracts that emit CEther related events.
func (c *Connector) CEtherLogsListener(client *ethclient.Client, logs chan types.Log) ethereum.Subscription {
	query := ethereum.FilterQuery{
		Addresses: c.ContractAddresses,
	}

	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal().Err(err).Msg("subscribing CEther filter logs failed")
	}

	return sub
}

func (c *Connector) ProcessLogEvent(contractAbi abi.ABI, evLog types.Log) (proto.Message, error) {
	// TODO: Add timestamp from block since logs don't include timestamp
	ev, err := contractAbi.EventByID(evLog.Topics[0])
	if err != nil {
		log.Warn().Err(err).Msg("EventByID error, skipping")
		return nil, err
	}

	if ev == nil {
		log.Warn().Msg("ignore if event id isn't defined in a partial ABI")
		return nil, err
	}

	var msg proto.Message

	switch ev.Name {
	case "Mint":
		event := new(ctoken.CompoundMint)
		if err := UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack Mint event error")
			return nil, err
		}
		msg = &ctoken.Mint{
			Ts:         timestamppb.Now(),
			Block:      evLog.BlockNumber,
			Idx:        uint64(evLog.Index),
			Tx:         evLog.TxHash.Bytes(),
			Minter:     event.Minter.Bytes(),
			MintAmount: event.MintAmount.Bytes(),
			MintTokens: event.MintTokens.Bytes(),
		}

		break
	case "Redeem":
		event := new(ctoken.CompoundRedeem)
		if err := UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack Redeem event error")
			return nil, err
		}
		msg = &ctoken.Redeem{
			Ts:           timestamppb.Now(),
			Block:        evLog.BlockNumber,
			Idx:          uint64(evLog.Index),
			Tx:           evLog.TxHash.Bytes(),
			Redeemer:     event.Redeemer.Bytes(),
			RedeemAmount: event.RedeemAmount.Bytes(),
			RedeemTokens: event.RedeemTokens.Bytes(),
		}

		break
	case "Borrow":
		event := new(ctoken.CompoundBorrow)
		if err := UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack Borrow event error")
			return nil, err
		}
		msg = &ctoken.Borrow{
			Ts:             timestamppb.Now(),
			Block:          evLog.BlockNumber,
			Idx:            uint64(evLog.Index),
			Tx:             evLog.TxHash.Bytes(),
			Borrower:       event.Borrower.Bytes(),
			BorrowAmount:   event.BorrowAmount.Bytes(),
			AccountBorrows: event.AccountBorrows.Bytes(),
			TotalBorrows:   event.TotalBorrows.Bytes(),
		}

		break
	case "RepayBorrow":
		event := new(ctoken.CompoundRepayBorrow)
		if err := UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack RepayBorrow event error")
			return nil, err
		}
		msg = &ctoken.RepayBorrow{
			Ts:             timestamppb.Now(),
			Block:          evLog.BlockNumber,
			Idx:            uint64(evLog.Index),
			Tx:             evLog.TxHash.Bytes(),
			Payer:          event.Payer.Bytes(),
			Borrower:       event.Borrower.Bytes(),
			RepayAmount:    event.RepayAmount.Bytes(),
			AccountBorrows: event.AccountBorrows.Bytes(),
			TotalBorrows:   event.TotalBorrows.Bytes(),
		}

		break
	case "LiquidateBorrow":
		event := new(ctoken.CompoundLiquidateBorrow)
		if err := UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack LiquidateBorrow event error")
			return nil, err
		}
		msg = &ctoken.LiquidateBorrow{
			Ts:               timestamppb.Now(),
			Block:            evLog.BlockNumber,
			Idx:              uint64(evLog.Index),
			Tx:               evLog.TxHash.Bytes(),
			Liquidator:       event.Liquidator.Bytes(),
			Borrower:         event.Borrower.Bytes(),
			RepayAmount:      event.RepayAmount.Bytes(),
			CTokenCollateral: event.CTokenCollateral.Bytes(),
			SeizeTokens:      event.SeizeTokens.Bytes(),
		}

		break
	}

	return msg, nil
}

func (c *Connector) buildTopicTypes() protoregistry.TopicTypes {
	tt := make(map[string]proto.Message)

	tt[c.GenerateTopicFromProto(&ctoken.Mint{}).Schema()] = &ctoken.Mint{}
	tt[c.GenerateTopicFromProto(&ctoken.Redeem{}).Schema()] = &ctoken.Redeem{}
	tt[c.GenerateTopicFromProto(&ctoken.Borrow{}).Schema()] = &ctoken.Borrow{}
	tt[c.GenerateTopicFromProto(&ctoken.RepayBorrow{}).Schema()] = &ctoken.RepayBorrow{}
	tt[c.GenerateTopicFromProto(&ctoken.LiquidateBorrow{}).Schema()] = &ctoken.LiquidateBorrow{}

	return tt
}

func ConvertRawAddress(rawAddresses ...string) []common.Address {
	var addresses []common.Address

	for _, addr := range rawAddresses {
		addresses = append(addresses, common.HexToAddress(addr))
	}

	return addresses
}

// UnpackLog is copied from https://github.com/ethereum/go-ethereum/blob/c2d2f4ed8f232bb11663a1b01a2e578aa22f24bd/accounts/abi/bind/base.go#L350
func UnpackLog(contractAbi abi.ABI, out interface{}, event string, log types.Log) error {
	if len(log.Data) > 0 {
		if err := contractAbi.UnpackIntoInterface(out, event, log.Data); err != nil {
			return err
		}
	}
	var indexed abi.Arguments
	for _, arg := range contractAbi.Events[event].Inputs {
		if arg.Indexed {
			indexed = append(indexed, arg)
		}
	}
	return abi.ParseTopics(out, indexed, log.Topics[1:])
}
