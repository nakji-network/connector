package protoregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"google.golang.org/protobuf/proto"

	"github.com/nakji-network/connector/kafkautils"
)

type Client struct {
	host string
}

func NewClient(host string) *Client {
	return &Client{host}
}

// TopicTypes is a map where topic schemas are keys and proto.Message are values.
type TopicTypes map[string]proto.Message

// RegisterDynamicTopics registers kafka topic and protobuf type mappings
func (c *Client) RegisterDynamicTopics(topicTypes TopicTypes, msgType kafkautils.MsgType) error {
	tpmList := buildTopicProtoMsgs(topicTypes, msgType)

	u := url.URL{Scheme: "http", Host: c.host, Path: "/v1/register"}

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

type TopicProtoMsg struct {
	MsgType      kafkautils.MsgType `json:"msg_type" binding:"required"`
	TopicName    string             `json:"topic" binding:"required"`
	ProtoMsgName string             `json:"proto_msg" binding:"required"`
}

type TopicSubscription struct {
	TopicProtoMsgs []TopicProtoMsg
	ShouldUpdate   bool `json:"should_update"`
}

func buildTopicProtoMsgs(topicTypes TopicTypes, msgType kafkautils.MsgType) []TopicProtoMsg {
	var tpmList []TopicProtoMsg

	for k, v := range topicTypes {
		t := reflect.TypeOf(v)
		pmn := t.String()

		if t.Kind() == reflect.Ptr {
			elem := t.Elem()
			pmn = elem.String()
		}

		tpm := TopicProtoMsg{
			MsgType:      msgType,
			TopicName:    k,
			ProtoMsgName: pmn,
		}

		tpmList = append(tpmList, tpm)
	}
	return tpmList
}
