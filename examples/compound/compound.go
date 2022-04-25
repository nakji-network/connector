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
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nakji-network/connector"
	"github.com/nakji-network/connector/examples/compound/ctoken"
	"github.com/nakji-network/connector/kafkautils"
)

type Connector struct {
	*connector.Connector
	Topics            map[string]kafkautils.Topic
	ContractAddresses []common.Address
	Chain             string // chain override, since ChainClients.Ethereum supports overriding with any evm chain
}

const (
	Namespace = "compound"
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

	// Get the initialized Ethereum client. For more Nakji supported clients see connector/chain/
	client := c.ChainClients.Ethereum(context.Background(), c.Chain)

	// Subscribe to headers
	headers := make(chan *types.Header)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe headers")
	}
	sub = c.CEtherLogsListener(client, logs)

	sink, err := c.MakeQueueTransactionSink()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe headers")
	}

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
			c.ProcessLogEvent(contractAbi, evLog, sink)
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

func (c *Connector) ProcessLogEvent(contractAbi abi.ABI, evLog types.Log, sink chan *kafkautils.Message) {
	// TODO: Add timestamp from block since logs don't include timestamp
	// Writes to out chan
	err := c.WriteEventToChan(contractAbi, evLog, sink)
	if err != nil {
		log.Error().Err(err).
			Interface("evLog", evLog).
			Msg("processLogEvent error")
	}
}

// WriteEventToChan parses the event and writes it to the channel for kafka.
func (c *Connector) WriteEventToChan(contractAbi abi.ABI, evLog types.Log, out chan<- *kafkautils.Message) error {
	ev, err := contractAbi.EventByID(evLog.Topics[0])
	if err != nil {
		log.Warn().Err(err).Msg("EventByID error, skipping")
		return err
	}

	if ev == nil {
		log.Warn().Msg("ignore if event id isn't defined in a partial ABI")
		return nil
	}

	switch ev.Name {
	case "Mint":
		event := new(ctoken.CompoundMint)
		if err := UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack Mint event error")
			return nil
		}
		out <- &kafkautils.Message{
			Topic: c.Topics["mint"],
			Key:   kafkautils.NewKey(Namespace, evLog.Address.Hex()),
			ProtoMsg: &ctoken.Mint{
				Ts:         timestamppb.Now(),
				Block:      evLog.BlockNumber,
				Idx:        uint64(evLog.Index),
				Tx:         evLog.TxHash.Bytes(),
				Minter:     event.Minter.Bytes(),
				MintAmount: event.MintAmount.Bytes(),
				MintTokens: event.MintTokens.Bytes(),
			},
		}
	case "Redeem":
		event := new(ctoken.CompoundRedeem)
		if err := UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack Redeem event error")
			return nil
		}
		out <- &kafkautils.Message{
			Topic: c.Topics["redeem"],
			Key:   kafkautils.NewKey(Namespace, evLog.Address.Hex()),
			ProtoMsg: &ctoken.Redeem{
				Ts:           timestamppb.Now(),
				Block:        evLog.BlockNumber,
				Idx:          uint64(evLog.Index),
				Tx:           evLog.TxHash.Bytes(),
				Redeemer:     event.Redeemer.Bytes(),
				RedeemAmount: event.RedeemAmount.Bytes(),
				RedeemTokens: event.RedeemTokens.Bytes(),
			},
		}
	case "Borrow":
		event := new(ctoken.CompoundBorrow)
		if err := UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack Borrow event error")
			return nil
		}
		out <- &kafkautils.Message{
			Topic: c.Topics["borrow"],
			Key:   kafkautils.NewKey(Namespace, evLog.Address.Hex()),
			ProtoMsg: &ctoken.Borrow{
				Ts:             timestamppb.Now(),
				Block:          evLog.BlockNumber,
				Idx:            uint64(evLog.Index),
				Tx:             evLog.TxHash.Bytes(),
				Borrower:       event.Borrower.Bytes(),
				BorrowAmount:   event.BorrowAmount.Bytes(),
				AccountBorrows: event.AccountBorrows.Bytes(),
				TotalBorrows:   event.TotalBorrows.Bytes(),
			},
		}
	case "RepayBorrow":
		event := new(ctoken.CompoundRepayBorrow)
		if err := UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack RepayBorrow event error")
			return nil
		}
		out <- &kafkautils.Message{
			Topic: c.Topics["repayborrow"],
			Key:   kafkautils.NewKey(Namespace, evLog.Address.Hex()),
			ProtoMsg: &ctoken.RepayBorrow{
				Ts:             timestamppb.Now(),
				Block:          evLog.BlockNumber,
				Idx:            uint64(evLog.Index),
				Tx:             evLog.TxHash.Bytes(),
				Payer:          event.Payer.Bytes(),
				Borrower:       event.Borrower.Bytes(),
				RepayAmount:    event.RepayAmount.Bytes(),
				AccountBorrows: event.AccountBorrows.Bytes(),
				TotalBorrows:   event.TotalBorrows.Bytes(),
			},
		}
	case "LiquidateBorrow":
		event := new(ctoken.CompoundLiquidateBorrow)
		if err := UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack LiquidateBorrow event error")
			return nil
		}
		out <- &kafkautils.Message{
			Topic: c.Topics["liquidateborrow"],
			Key:   kafkautils.NewKey(Namespace, evLog.Address.Hex()),
			ProtoMsg: &ctoken.LiquidateBorrow{
				Ts:               timestamppb.Now(),
				Block:            evLog.BlockNumber,
				Idx:              uint64(evLog.Index),
				Tx:               evLog.TxHash.Bytes(),
				Liquidator:       event.Liquidator.Bytes(),
				Borrower:         event.Borrower.Bytes(),
				RepayAmount:      event.RepayAmount.Bytes(),
				CTokenCollateral: event.CTokenCollateral.Bytes(),
				SeizeTokens:      event.SeizeTokens.Bytes(),
			},
		}
	}

	return nil
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
