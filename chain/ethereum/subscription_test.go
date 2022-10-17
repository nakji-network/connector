package ethereum

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	lru "github.com/hashicorp/golang-lru"
)

func TestGetBlockTime(t *testing.T) {
	t.Parallel()

	c, err := lru.New(1280)
	if err != nil {
		t.Error("Cache initialization failed.", err)
	}

	s := Subscription{
		cache: c,
	}

	var wantedTS uint64 = 1633956315
	s.cache.Add("0x534cc509df37f6b5fbf10691db4d70bb6d2199fbaccca5494d7da7fce489c85c", wantedTS)
	vLog := types.Log{BlockHash: common.HexToHash("0x534cc509df37f6b5fbf10691db4d70bb6d2199fbaccca5494d7da7fce489c85c")}
	ts, err := s.GetBlockTime(context.Background(), vLog)
	if err != nil {
		t.Error("GetBlockTime failed.", err)
	}

	if ts != wantedTS {
		t.Error("Event log KEY parse failed.", "got:", ts, "want:", wantedTS)
	}
}

func TestClose(t *testing.T) {
	t.Parallel()

	s := Subscription{
		done:    make(chan bool),
		headers: make(chan *types.Header),
		client:  &ethclient.Client{},
		inLogs:  make(chan types.Log),
		outLogs: make(chan types.Log),
		inErr:   make(chan error),
		outErr:  make(chan error),
	}

	go func() {
		<-s.Done()
	}()

	s.Close()
}
