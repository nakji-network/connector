package compound

import (
	"context"
	"os"
	"os/signal"
	"strings"

	"github.com/nakji-network/connector"
	"github.com/nakji-network/connector/common"
	"github.com/nakji-network/connector/examples/compound/ctoken"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Connector struct {
	*connector.Connector
	ContractAddresses []ethcommon.Address
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
	contractAbi1, err := abi.JSON(strings.NewReader(ctoken.CompoundABI))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to read CEther abi")
	}
	contractAbi2, err := abi.JSON(strings.NewReader(ctoken.CtokenMetaData.ABI))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read CEther abi")
	}

	contractAbis := []abi.ABI{contractAbi1, contractAbi2}

	// Register topic and protobuf type mappings
	protos := []proto.Message{
		&ctoken.Mint{},
		&ctoken.Redeem{},
		&ctoken.Borrow{},
		&ctoken.RepayBorrow{},
		&ctoken.LiquidateBorrow{},
	}

	c.RegisterProtos(protos...)

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
			msg, err := c.ProcessLogEvent(contractAbis, evLog)
			if err != nil {
				log.Error().Err(err).Msg("failed to process log event")
				continue
			}
			if msg == nil {
				log.Warn().Msg("empty message")
				continue
			}
			// Not sure what value needs to be passed as subject
			err = c.ProduceAndCommitMessage(namespace, evLog.Address.Hex(), msg)
			if err != nil {
				log.Error().Err(err).Msg("Kafka write proto")
			}
		case <-interrupt:
			log.Info().Msg("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			client.Close()
			c.ProducerInterface.Close()
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

func (c *Connector) ProcessLogEvent(contractAbis []abi.ABI, evLog types.Log) (proto.Message, error) {
	// TODO: Add timestamp from block since logs don't include timestamp
	var ev *abi.Event
	var err error
	var idx int

	for i, cAbi := range contractAbis {
		ev, err = cAbi.EventByID(evLog.Topics[0])
		if err == nil {
			idx = i
			break
		}
	}

	if ev == nil {
		log.Warn().Msg("ignore if event id isn't defined in a partial ABI")
		return nil, err
	}

	contractAbi := contractAbis[idx]

	var msg proto.Message

	switch ev.Name {
	case "Mint":
		event := new(ctoken.CompoundMint)
		if err := common.UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
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
	case "Redeem":
		event := new(ctoken.CompoundRedeem)
		if err := common.UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
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
	case "Borrow":
		event := new(ctoken.CompoundBorrow)
		if err := common.UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
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
	case "RepayBorrow":
		event := new(ctoken.CompoundRepayBorrow)
		if err := common.UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
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
	case "LiquidateBorrow":
		event := new(ctoken.CompoundLiquidateBorrow)
		if err := common.UnpackLog(contractAbi, event, ev.Name, evLog); err != nil {
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
	}

	return msg, nil
}

func ConvertRawAddress(rawAddresses ...string) []ethcommon.Address {
	var addresses []ethcommon.Address

	for _, addr := range rawAddresses {
		addresses = append(addresses, ethcommon.HexToAddress(addr))
	}

	return addresses
}
