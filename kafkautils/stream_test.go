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
				Subject:   "bitcoin",
				Event:     "*",
			},
		},
		{
			input: "vitalik.common.0_0_0.ethereum_tx",
			want: &StreamName{
				Author:    "vitalik",
				Namespace: "ethereum",
				Version:   "0_0_0",
				Subject:   "*",
				Event:     "*",
			},
		},
		{
			input: "nakji.common.0_0_0.liquiditypool_reserve-1h",
			want: &StreamName{
				Author:    "nakji",
				Namespace: "common",
				Version:   "0_0_0",
				Subject:   "liquiditypool",
				Event:     "reserve",
				Period:    "1h",
			},
		},
		{
			input: "wrong.schema",
			want:  nil,
		},
	} {
		res, _ := NewSchema(testCase.input)
		if testCase.want != res {
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
				Subject:   "bitcoin",
				Event:     "*",
			},
			want: true,
		},
		{
			input: &StreamName{
				Author:    "vitalik",
				Namespace: "ethereum",
				Version:   "0_0_0",
				Subject:   "*",
				Event:     "*",
			},
			want: true,
		},
		{
			input: &StreamName{
				Author:    "nakji",
				Namespace: "common",
				Version:   "",
				Subject:   "liquiditypool",
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
				Subject:   "bitcoin",
				Event:     "*",
			},
			input: "satoshi.common.0_0_0.bitcoin_block",
			want:  true,
		},
		{
			receiver: &StreamName{
				Author:    "vitalik",
				Namespace: "ethereum",
				Version:   "0_0_0",
				Subject:   "*",
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
				Subject:   "liquiditypool",
				Event:     "reserve",
				Period:    "1h",
			},
			input: "nakji.common.0_0_0.liquiditypool_change",
			want:  false,
		},
	} {
		res := testCase.receiver.hasSchema(testCase.input)
		if testCase.want != res {
			t.Error("schema hasSchema failed.", "got:", res, "want:", testCase.want)
		}
	}
}
