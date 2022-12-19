package ethereum

import (
	"github.com/nakji-network/connector/examples/ethereum/chain"

	"google.golang.org/protobuf/proto"
)

var protos = []proto.Message{
	&chain.Block{},
	&chain.Transaction{},
}
