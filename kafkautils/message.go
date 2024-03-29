package kafkautils

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// Message is a kafka.Message wrapper to store protoMsg
// this way, the kafka event stays together
type Message struct {
	*kafka.Message
	Topic    Topic
	Key      Key
	MsgType  MsgType
	ProtoMsg proto.Message
	Span     trace.Span
	Baggage  baggage.Baggage
}

type MsgType string

const (
	MsgTypeFct MsgType = "fct"
	MsgTypeBf  MsgType = "bf" //	backfill
	MsgTypeCdc MsgType = "cdc"
	MsgTypeCmd MsgType = "cmd"
	MsgTypeSys MsgType = "sys"
)

// FieldNames gets the field names from the ProtoMsg
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
