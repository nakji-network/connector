// Helpers for Blep's kafka topic naming scheme
// Design document: https://docs.google.com/spreadsheets/d/1PmYvbw8LiBYYooAINrm4_lGWiewKA-yq-zCGbXQVNfE/edit#gid=0
package kafkautils

import (
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

type Env string

const (
	Prod    Env = "prod"
	Staging     = "staging"
	Dev         = "dev"
	Test        = "test"
)

const (
	TopicContextSeparator   string = "."
	TopicContractSeparator  string = "_"
	TopicAggregateSeparator string = "-"
	TopicWildcardSuffix     string = "-*"
	TopicNumSegments        int    = 4
)

type MsgType string

const (
	Fct MsgType = "fct"
	Cdc         = "cdc"
	Cmd         = "cmd"
	Sys         = "sys"
)

type Topic struct {
	Env           Env
	MsgType       MsgType
	Author        string
	ConnectorName string
	Version       *semver.Version
	EventName     string
	pb            proto.Message // create an empty protobuf struct instance, filled upon UnmarshalProto
}

// String generates the topic string
func (t Topic) String() string {
	return strings.Join([]string{string(t.Env), string(t.MsgType), t.Schema()}, TopicContextSeparator)
}

// Schema generates the schema string
func (t Topic) Schema() string {
	return strings.Join([]string{
		t.Author,
		t.ConnectorName,
		strings.ReplaceAll(t.Version.String(), ".", "_"),
		t.EventName,
	}, TopicContextSeparator)
}

func NewTopic(en Env, ty MsgType, author, connectorName string, version *semver.Version, msg proto.Message) Topic {
	return Topic{
		Env:           en,
		MsgType:       ty,
		Author:        author,
		ConnectorName: connectorName,
		Version:       version,
		EventName:     strings.ReplaceAll(string(msg.ProtoReflect().Descriptor().FullName()), ".", "_"),
		pb:            msg,
	}
}

// ParseTopic parses topic string to Topic struct.
// topic strings that start with . (eg .fct.nakji.ethereum.0_0_0.chain_block) get set `dev` prefix.
// Use second argument to override env (only for initialization at start of program)
func ParseTopic(s string, e ...string) (Topic, error) {
	p := strings.Split(s, TopicContextSeparator)

	if len(p) != 6 {
		return Topic{}, fmt.Errorf("cannot parse topic, does not have 6 segments: %s", s)
	}

	schema := strings.SplitAfterN(s, TopicContextSeparator, 3)[2]
	version, err := semver.NewVersion(strings.ReplaceAll(p[4], "_", "."))
	if err != nil {
		return Topic{}, err
	}

	pbType := TopicTypeRegistry.Get(schema)
	if pbType == nil {
		return Topic{}, fmt.Errorf("cannot find topic schema in type registry: %s", schema)
	}

	res := Topic{
		Env:           Env(p[0]),
		MsgType:       MsgType(p[1]),
		Author:        p[2],
		ConnectorName: p[3],
		Version:       version,
		EventName:     p[5],
		pb:            proto.Clone(pbType),
	}

	// override env
	if len(e) == 1 {
		res.Env = Env(e[0])
	}

	if res.Env == "" {
		return Topic{}, fmt.Errorf("invalid env (empty)")
	}

	return res, nil
}

func MustParseTopicsMap(m map[string]string, e ...string) map[string]Topic {
	topics := make(map[string]Topic)
	for k, v := range m {
		topics[k] = MustParseTopic(v, e[0])
	}
	return topics
}

func MustParseTopics(s []string, e ...string) []Topic {
	topics := make([]Topic, len(s))
	for i, t := range s {
		topics[i] = MustParseTopic(t, e[0])
	}
	return topics
}

func MustParseTopic(s string, e ...string) Topic {
	t, err := ParseTopic(s, e...)
	if err != nil {
		log.Warn().Err(err).Msg("")
	}
	return t
}

func TopicsStrings(topics []Topic) []string {
	res := make([]string, len(topics))
	for i, t := range topics {
		res[i] = t.String()
	}
	return res
}

// UnmarshalProto converts protobuf bytes -> struct
func (t *Topic) UnmarshalProto(data []byte) (proto.Message, error) {
	if t.pb == nil {
		return nil, fmt.Errorf("cannot unmarshal proto for topic %s", t)
	}
	return t.pb, proto.Unmarshal(data, t.pb)
}
