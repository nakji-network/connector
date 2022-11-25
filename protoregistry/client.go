package protoregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/nakji-network/connector/kafkautils"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
)

// TopicTypes is a map where topic schemas are keys and proto.Message are values.
type TopicTypes map[string]proto.Message

// RegisterDynamicTopics registers kafka topic and protobuf type mappings
func RegisterDynamicTopics(host string, topicTypes map[string]proto.Message, msgType kafkautils.MsgType) error {
	tpmList := buildTopicProtoMsgs(topicTypes, msgType)

	u := url.URL{Scheme: "http", Host: host, Path: "/v1/register"}

	b, err := json.Marshal(tpmList)
	if err != nil {
		return err
	}

	res, err := http.Post(u.String(), "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	statusOK := res.StatusCode >= 200 && res.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("call failed, err: %v", string(b))
	}

	return nil
}

func buildTopicProtoMsgs(topicTypes TopicTypes, msgType kafkautils.MsgType) []*TopicProtoMsg {
	var tpmList []*TopicProtoMsg

	for k, v := range topicTypes {
		md := v.ProtoReflect().Descriptor()

		fdp := protodesc.ToFileDescriptorProto(md.ParentFile())

		tpm := &TopicProtoMsg{
			MsgType:             msgType,
			TopicName:           k,
			ProtoMsgName:        string(md.FullName()),
			FileDescriptorProto: fdp,
			Descriptor:          nil,
		}

		tpmList = append(tpmList, tpm)
	}
	return tpmList
}
