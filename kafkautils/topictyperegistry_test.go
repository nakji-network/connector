package kafkautils

import (
	"testing"

	"google.golang.org/protobuf/proto"
)

var testTopicTypes = map[string]proto.Message{
	"nakji.common.0_0_0.market_trade":    &Parsley{},
	"nakji.common.0_0_0.market_trade-*":  &Parsley{},
	"nakji.common.0_0_0.market_ohlc":     &Parsley{},
	"parsley":                            &Petersilie{},
	"blep.test.1_2_3.mycontract_parsley": &Petersilie{},
	"blep.test.3_2_1.mycontract_parsley": &Petersilie{},
}

func TestSet(t *testing.T) {
	for _, testCase := range []string{
		"satoshi.common.0_0_0.bitcoin_tx",
		"satoshi.common.0_0_0.bitcoin_block",
	} {
		TopicTypeRegistry.Set(testCase, &Parsley{})
	}
}

func TestGet(t *testing.T) {
	for _, testCase := range []struct {
		input string
		want  proto.Message
	}{
		{
			input: "nakji.common.0_0_0.market_trade",
			want:  &Parsley{},
		},
		{
			input: "satoshi.common.0_0_0.bitcoin_trade",
			want:  nil,
		},
	} {
		res := TopicTypeRegistry.Get(testCase.input)
		if (testCase.want != nil && res == nil) || (testCase.want == nil && res != nil) {
			t.Error("topic type registry Get failed.", "got:", res, "want:", testCase.want, "input", testCase.input)
		}
	}
}

func TestGetActiveSchemas(t *testing.T) {
	for _, testCase := range []struct {
		input []string
		want  map[string]bool
	}{
		{
			input: []string{
				"nakji.common.0_0_0.market_trade",
			},
			want: map[string]bool{
				"nakji.common.0_0_0.market_trade":   true,
				"nakji.common.0_0_0.market_trade-*": true,
			},
		},
		{
			input: []string{
				"nakji",
			},
			want: map[string]bool{},
		},
	} {
		res := GetActiveSchemas(testCase.input)
		for k, v := range res {
			if testCase.want[k] != v {
				t.Error("topic type registry GetActiveTopics failed.", "got:", res, "want:", testCase.want)
			}
		}
	}
}
