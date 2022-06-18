package common

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
)

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
