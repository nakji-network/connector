package kafkautils

import (
	"fmt"
	"reflect"
	"testing"

	"google.golang.org/protobuf/proto"
)

func init() {
	var TopicTypes = map[string]proto.Message{
		"parsley":                            &Petersilie{},
		"blep.test.1_2_3.mycontract_parsley": &Petersilie{},
		"blep.test.3_2_1.mycontract_parsley": &Petersilie{},
	}

	TopicTypeRegistry.Load(TopicTypes)
}

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
				Env:     "test",
				MsgType: "fct",
				Schema:  "blep.test.1_2_3.mycontract_parsley",
				Version: "1_2_3",
				pb:      proto.Clone(&Petersilie{}),
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
