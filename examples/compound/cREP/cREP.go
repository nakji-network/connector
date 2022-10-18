package cREP

import (
	"github.com/nakji-network/connector/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventParser struct{}

func (ep *EventParser) Get(eventName string, contractAbi *abi.ABI, evLog types.Log, timestamp *timestamppb.Timestamp) proto.Message {
	switch eventName {
	case "Mint":
		event := new(CREPMint)
		if err := common.UnpackLog(*contractAbi, event, eventName, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack Mint event error")
			return nil
		}
		return &Mint{
			Ts:         timestamp,
			Block:      evLog.BlockNumber,
			Idx:        uint64(evLog.Index),
			Tx:         evLog.TxHash.Bytes(),
			Minter:     event.Minter.Bytes(),
			MintAmount: event.MintAmount.Bytes(),
			MintTokens: event.MintTokens.Bytes(),
		}
	case "Redeem":
		event := new(CREPRedeem)
		if err := common.UnpackLog(*contractAbi, event, eventName, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack Redeem event error")
			return nil
		}
		return &Redeem{
			Ts:           timestamp,
			Block:        evLog.BlockNumber,
			Idx:          uint64(evLog.Index),
			Tx:           evLog.TxHash.Bytes(),
			Redeemer:     event.Redeemer.Bytes(),
			RedeemAmount: event.RedeemAmount.Bytes(),
			RedeemTokens: event.RedeemTokens.Bytes(),
		}
	case "Borrow":
		event := new(CREPBorrow)
		if err := common.UnpackLog(*contractAbi, event, eventName, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack Borrow event error")
			return nil
		}
		return &Borrow{
			Ts:             timestamp,
			Block:          evLog.BlockNumber,
			Idx:            uint64(evLog.Index),
			Tx:             evLog.TxHash.Bytes(),
			Borrower:       event.Borrower.Bytes(),
			BorrowAmount:   event.BorrowAmount.Bytes(),
			AccountBorrows: event.AccountBorrows.Bytes(),
			TotalBorrows:   event.TotalBorrows.Bytes(),
		}
	case "RepayBorrow":
		event := new(CREPRepayBorrow)
		if err := common.UnpackLog(*contractAbi, event, eventName, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack RepayBorrow event error")
			return nil
		}
		return &RepayBorrow{
			Ts:             timestamp,
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
		event := new(CREPLiquidateBorrow)
		if err := common.UnpackLog(*contractAbi, event, eventName, evLog); err != nil {
			log.Error().Err(err).Msg("Unpack LiquidateBorrow event error")
			return nil
		}
		return &LiquidateBorrow{
			Ts:               timestamp,
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
	return nil
}
