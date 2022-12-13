package chain

import (
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/nakji-network/connector/common"
	"github.com/rs/zerolog/log"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func ParseHeader(header *types.Header) *Block {
	return &Block{
		Ts:         common.UnixToTimestampPb(int64(header.Time)),
		Hash:       header.Hash().Hex(),
		Difficulty: header.Difficulty.Uint64(),
		Number:     header.Number.Uint64(),
		GasLimit:   header.GasLimit,
		GasUsed:    header.GasUsed,
		Nonce:      header.Nonce.Uint64(),
	}
}

func ParseTransaction(tx *types.Transaction, timestamp *timestamppb.Timestamp) *Transaction {
	V, R, S := tx.RawSignatureValues()

	// handle nil recipients for contract creations
	recipient := []byte{}
	if tx.To() != nil {
		recipient = tx.To().Bytes()
	}

	// Get Sender (.From()) address
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		log.Warn().Err(err).
			Interface("tx", tx).
			Msg("UnmarshallEthTransaction .AsMessage error")
		return nil
	}

	return &Transaction{
		Ts:           timestamp,
		From:         from.Bytes(),
		Hash:         tx.Hash().Hex(),
		Size:         float64(tx.Size()),
		AccountNonce: tx.Nonce(),
		Price:        tx.GasPrice().Uint64(),
		GasLimit:     tx.Gas(),
		Recipient:    recipient,
		Amount:       tx.Value().Uint64(),
		Payload:      tx.Data(),
		V:            V.Uint64(),
		R:            R.Uint64(),
		S:            S.Uint64(),
	}
}
