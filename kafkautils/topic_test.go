package kafkautils

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Masterminds/semver"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestParseTopic(t *testing.T) {
	type args struct {
		s string
		e []string
	}
	tests := []struct {
		name    string
		args    args
		want    Topic
		wantErr error
	}{
		{
			name: "parse new style topic",
			args: args{s: ".fct.blep.test.1_2_3.mycontract_parsley", e: []string{"test"}},
			want: Topic{
				Env:           "test",
				MsgType:       "fct",
				Author:        "blep",
				ConnectorName: "test",
				Version:       semver.MustParse("1.2.3"),
				EventName:     "mycontract_parsley",
				pb:            proto.Clone(&Petersilie{}),
			},
			wantErr: nil,
		},
		{
			name:    "parse invalid topic",
			args:    args{s: ".invalid.fct.blep_test.3_2_1.parsley", e: []string{"test"}},
			want:    Topic{},
			wantErr: fmt.Errorf("cannot find topic schema in type registry: %s", "fct.blep_test.3_2_1.parsley"),
		},
		{
			name:    "parse topic without env",
			args:    args{s: ".fct.blep.test.3_2_1.mycontract_parsley", e: []string{}},
			want:    Topic{},
			wantErr: fmt.Errorf("invalid env (empty)"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTopic(tt.args.s, tt.args.e...)

			if err != nil && err.Error() != tt.wantErr.Error() {
				t.Errorf("ParseTopic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTopic() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getEventName(t *testing.T) {
	type args struct {
		fn protoreflect.FullName
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "short name",
			args: args{
				fn: "chain.Transaction",
			},
			want: "chain_Transaction",
		},
		{
			name: "long name",
			args: args{
				fn: "nakji.evm.chain.Block",
			},
			want: "chain_Block",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEventName(tt.args.fn); got != tt.want {
				t.Errorf("getEventName() = %v, want %v", got, tt.want)
			}
		})
	}
}
