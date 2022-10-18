package ethereum

import (
	"google.golang.org/protobuf/proto"
)

var protos = []proto.Message{
	&Block{},
	&Transaction{},
	&Block0{},
}
