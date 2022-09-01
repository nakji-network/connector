package chain

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	// "github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

type MockETHClient struct{}
type MockSubscription struct{}

var logs []types.Log = []types.Log{
	{
		Address:     common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"),
		BlockHash:   common.HexToHash("0x11a8fad69e2a6ceb2782045dfdf889217c1b893fb96bdda96d524aa1b32022af"),
		BlockNumber: 13145843,
		// Data:        hexutil.MustDecode("0x000000000000000000000041529125421212024814400"),
		Index:   47,
		TxIndex: 35,
		TxHash:  common.HexToHash("0xf4f60cf66e67aa31ec2d7ca032803e53b81d33d8f4dc69f45a3d1257f14002d3"),
		Topics: []common.Hash{
			common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
			common.HexToHash("0x0000000000000000000000000cfeb7b8b2cf70e9e6fe768e43b8efbe640cc9ff"),
			common.HexToHash("0x0000000000000000000000003c9ff3cc55c82c82f4921083c1f32211d58225f5"),
		},
	},
}

func (MockETHClient) FilterLogs(context.Context, ethereum.FilterQuery) ([]types.Log, error) {
	return logs, nil
}
func (MockETHClient) SubscribeFilterLogs(context.Context, ethereum.FilterQuery, chan<- types.Log) (ethereum.Subscription, error) {
	return MockSubscription{}, nil
}

func (MockSubscription) Err() <-chan error { return nil }
func (MockSubscription) Unsubscribe()      {}

func TestChunkedSubscribeFilterLogs(t *testing.T) {
	mockETHClient := MockETHClient{}
	addr := make([]common.Address, 9000)
	logch := make(chan types.Log)
	errch := make(chan error)

	subs, err := ChunkedSubscribeFilterLogs(context.Background(), mockETHClient, addr, logch, errch, nil)
	if err != nil {
		t.Error("test ChunkedSubscribeFilterLogs failed. error: ", err)
	}
	if len(subs) != 2 {
		t.Errorf("test ChunkedSubscribeFilterLogs failed. want: %d, got: %d", 2, len(subs))
	}
}

func TestChunkedFilterLogs(t *testing.T) {
	mockETHClient := MockETHClient{}
	addr := make([]common.Address, 8000)
	const fromBlock uint64 = 10
	const toBlock uint64 = 20000
	logch := make(chan types.Log)

	go func() {
		i := 0
		for range logch {
			i++
		}
		if i != len(logs) {
			t.Errorf("test ChunkedFilterLogs failed. want: %d, got: %d", len(logs), i)
		}
	}()

	ChunkedFilterLogs(context.Background(), mockETHClient, addr, fromBlock, toBlock, logch, nil)
}
