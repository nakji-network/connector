package kafkautils

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"google.golang.org/protobuf/proto"
)

// kafka.Message wrapper to store protoMsg
// this way, the kafka event stays together
type Message struct {
	*kafka.Message
	Topic    Topic
	Key      Key
	ProtoMsg proto.Message
}

// Get the field names from the ProtoMsg
func (m *Message) FieldNames() []string {
	fieldDescs := m.ProtoMsg.ProtoReflect().Descriptor().Fields()

	// get all column names
	fields := make([]string, fieldDescs.Len())
	for i := 0; i < fieldDescs.Len(); i++ {
		fd := fieldDescs.Get(i)
		fields[i] = string(fd.Name())
	}
	return fields
}
