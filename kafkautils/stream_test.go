package kafkautils

import (
	"testing"
)

func TestNewSchema(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		input string
		want  *StreamName
	}{
		{
			input: "satoshi.common.0_0_0.bitcoin_block",
			want: &StreamName{
				Author:    "satoshi",
				Namespace: "common",
				Version:   "0_0_0",
				Contract:  "bitcoin",
				Event:     "block",
				Period:    "",
			},
		},
		{
			input: "vitalik.common.0_0_0.ethereum_tx",
			want: &StreamName{
				Author:    "vitalik",
				Namespace: "common",
				Version:   "0_0_0",
				Contract:  "ethereum",
				Event:     "tx",
				Period:    "",
			},
		},
		{
			input: "nakji.common.0_0_0.liquiditypool_reserve-1h",
			want: &StreamName{
				Author:    "nakji",
				Namespace: "common",
				Version:   "0_0_0",
				Contract:  "liquiditypool",
				Event:     "reserve",
				Period:    "1h",
			},
		},
		// {
		// 	input: "wrong.schema",
		// 	want:  nil,
		// },
	} {
		res, _ := NewSchema(testCase.input)
		if (res == nil && testCase.want == nil) && testCase.want.Author != res.Author ||
			testCase.want.Namespace != res.Namespace || testCase.want.Version != res.Version ||
			testCase.want.Contract != res.Contract || testCase.want.Event != res.Event ||
			testCase.want.Period != res.Period {
			t.Error("schema NewSchema failed.", "got:", res, "want:", testCase.want)
		}
	}
}

func TestIsValid(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		input *StreamName
		want  bool
	}{
		{
			input: &StreamName{
				Author:    "satoshi",
				Namespace: "common",
				Version:   "0_0_0",
				Contract:  "bitcoin",
				Event:     "*",
			},
			want: true,
		},
		{
			input: &StreamName{
				Author:    "vitalik",
				Namespace: "ethereum",
				Version:   "0_0_0",
				Contract:  "*",
				Event:     "*",
			},
			want: true,
		},
		{
			input: &StreamName{
				Author:    "nakji",
				Namespace: "common",
				Version:   "",
				Contract:  "liquiditypool",
				Event:     "",
				Period:    "1h",
			},
			want: false,
		},
	} {
		res := testCase.input.isValid()
		if testCase.want != res {
			t.Error("schema isValid failed.", "got:", res, "want:", testCase.want)
		}
	}
}

func TestHasSchema(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		receiver *StreamName
		input    string
		want     bool
	}{
		{
			receiver: &StreamName{
				Author:    "satoshi",
				Namespace: "common",
				Version:   "0_0_0",
				Contract:  "bitcoin",
				Event:     "*",
			},
			input: "satoshi.common.0_0_0.bitcoin_block",
			want:  true,
		},
		{
			receiver: &StreamName{
				Author:    "vitalik",
				Namespace: "common",
				Version:   "0_0_0",
				Contract:  "*",
				Event:     "*",
			},
			input: "vitalik.common.0_0_0.ethereum_tx",
			want:  true,
		},
		{
			receiver: &StreamName{
				Author:    "nakji",
				Namespace: "common",
				Version:   "0_0_0",
				Contract:  "liquiditypool",
				Event:     "reserve",
				Period:    "1h",
			},
			input: "nakji.common.0_0_0.liquiditypool_change",
			want:  false,
		},
	} {
		res := testCase.receiver.hasSchema(testCase.input)
		if testCase.want != res {
			t.Error("schema", testCase.input, "hasSchema failed.", "got:", res, "want:", testCase.want)
		}
	}
}
