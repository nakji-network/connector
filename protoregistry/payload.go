package protoregistry

import (
	"github.com/nakji-network/connector/kafkautils"

	"google.golang.org/protobuf/types/descriptorpb"
)

type TopicProtoMsg struct {
	MsgType             kafkautils.MsgType                `json:"msg_type" binding:"required"`
	TopicName           string                            `json:"topic" binding:"required"`
	ProtoMsgName        string                            `json:"proto_msg" binding:"required"`
	FileDescriptorProto *descriptorpb.FileDescriptorProto `json:"file_descriptor_proto"`
	Descriptor          []byte                            `json:"descriptor"` // for backward compatibility
}

type TopicSubscription struct {
	TopicProtoMsgs []*TopicProtoMsg
	ShouldUpdate   bool  `json:"should_update"`
	Latest         int64 `json:"latest"`
}
