package kafkautils

import (
	"testing"

	"google.golang.org/protobuf/proto"
)

var testTopicTypes = map[string]proto.Message{
	"nakji.common.0_0_0.market_trade":   &Petersilie{},
	"nakji.common.0_0_0.market_trade-*": &Petersilie{},
	"nakji.common.0_0_0.market_ohlc":    &Petersilie{},
}

func TestSet(t *testing.T) {
	t.Parallel()

	for _, testCase := range []string{
		"satoshi1=",
		"satoshi.common.0_0_0.bitcoin_block",
	} {
		mockTTR.Set(testCase, &Petersilie{})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		input string
		want  proto.Message
	}{
		{
			input: "nakji.common.0_0_0.market_trade",
			want:  &Petersilie{},
		},
		{
			input: "satoshi.common.0_0_0.bitcoin_trade",
			want:  nil,
		},
	} {
		res := mockTTR.Get(testCase.input)
		if testCase.want != res {
			t.Error("topic type registry Get failed.", "got:", res, "want:", testCase.want)
		}
	}
}

func TestGetActiveTopics(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		input []string
		want  map[string]bool
	}{
		{
			input: nil,
			want: map[string]bool{
				".fct.nakji.common.0_0_0.market_trade":   true,
				".fct.nakji.common.0_0_0.market_trade-*": true,
				".fct.nakji.common.0_0_0.market_ohlc":    true,
			},
		},
		{
			input: []string{},
			want: map[string]bool{
				".fct.nakji.common.0_0_0.market_trade":   true,
				".fct.nakji.common.0_0_0.market_trade-*": true,
				".fct.nakji.common.0_0_0.market_ohlc":    true,
			},
		},
		{
			input: []string{
				"nakji.common.0_0_0.market_trade",
			},
			want: map[string]bool{
				".fct.nakji.common.0_0_0.market_trade": true,
			},
		},
		{
			input: []string{
				"nakji",
			},
			want: map[string]bool{},
		},
	} {
		res := GetActiveTopics(testCase.input)
		for k, v := range res {
			if testCase.want[k] != v {
				t.Error("topic type registry GetActiveTopics failed.", "got:", res, "want:", testCase.want)
			}
		}
	}
}
