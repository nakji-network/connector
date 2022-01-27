package kafkautils

import (
	"fmt"

	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

// Print proto message as human readable text
func DebugPrint(topic Topic, key Key, protoMsg proto.Message) {
	fmt.Printf("%s: %s = %s\n", topic, key.String(), prototext.Format(protoMsg))
}
