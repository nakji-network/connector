package monitor

import (
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func TestMessageCarrier_Get(t *testing.T) {
	type args struct {
		msg *kafka.Message
	}
	tests := []struct {
		name string
		args args
		key  string
		want string
	}{
		{"get val1",
			args{
				&kafka.Message{
					Headers: []kafka.Header{
						{
							Key:   "key1",
							Value: []byte("value1"),
						},
						{
							Key:   "key2",
							Value: []byte("value2"),
						},
					},
				},
			},
			"key1",
			"value1",
		},
		{"get val2",
			args{
				&kafka.Message{
					Headers: []kafka.Header{
						{
							Key:   "key1",
							Value: []byte("value1"),
						},
						{
							Key:   "key2",
							Value: []byte("value2"),
						},
					},
				},
			},
			"key2",
			"value2",
		},
		{"get empty val",
			args{
				&kafka.Message{
					Headers: []kafka.Header{
						{
							Key:   "key1",
							Value: []byte("value1"),
						},
						{
							Key:   "key2",
							Value: []byte("value2"),
						},
					},
				},
			},
			"key3",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := MessageCarrier{
				msg: tt.args.msg,
			}
			if got := c.Get(tt.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessageCarrier_Set(t *testing.T) {
	type args struct {
		msg *kafka.Message
	}
	tests := []struct {
		name string
		args args
		key  string
		val  string
		want int
	}{
		{"set header1",
			args{
				&kafka.Message{},
			},
			"key1",
			"value1",
			1,
		},
		{"set header2",
			args{
				&kafka.Message{
					Headers: []kafka.Header{
						{
							Key:   "key1",
							Value: []byte("value1"),
						},
					},
				},
			},
			"key2",
			"value2",
			2,
		},
		{"override header1",
			args{
				&kafka.Message{
					Headers: []kafka.Header{
						{
							Key:   "key1",
							Value: []byte("value1"),
						},
					},
				},
			},
			"key1",
			"value3",
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := MessageCarrier{
				msg: tt.args.msg,
			}
			c.Set(tt.key, tt.val)
			if len(tt.args.msg.Headers) != tt.want {
				t.Errorf("len(headers) = %v, want %v", len(tt.args.msg.Headers), tt.want)
			}
			for _, header := range tt.args.msg.Headers {
				if header.Key == tt.key && string(header.Value) != tt.val {
					t.Errorf("val = %v, want %v", string(header.Value), tt.val)
				}
			}
		})
	}
}

func TestMessageCarrier_Keys(t *testing.T) {
	type args struct {
		msg *kafka.Message
	}
	tests := []struct {
		name string
		args args
		key  string
		val  string
		want int
	}{
		{"get keys",
			args{
				&kafka.Message{
					Headers: []kafka.Header{
						{
							Key:   "key1",
							Value: []byte("value1"),
						},
						{
							Key:   "key2",
							Value: []byte("value2"),
						},
						{
							Key:   "key3",
							Value: []byte("value3"),
						},
					},
				},
			},
			"key1",
			"value1",
			3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := MessageCarrier{
				msg: tt.args.msg,
			}
			keys := c.Keys()
			if len(keys) != tt.want {
				t.Errorf("len(keys) = %v, want %v", len(tt.args.msg.Headers), tt.want)
			}
			for i, header := range tt.args.msg.Headers {
				if header.Key != keys[i] {
					t.Errorf("key = %v, want %v", keys[i], header.Key)
				}
			}
		})
	}
}
