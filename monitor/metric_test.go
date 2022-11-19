package monitor

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/baggage"
)

func TestGetBaggageLatency(t *testing.T) {
	type args struct {
		ctx context.Context
		key string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			"normal input",
			args{
				createBaggage("latency", "3459"),
				"latency",
			},
			3459,
		},
		{
			"negative input",
			args{
				createBaggage("latency", "-99999"),
				"latency",
			},
			0,
		},
		{
			"empty value",
			args{
				createBaggage("empty", "3459"),
				"latency",
			},
			0,
		},
		{
			"large value",
			args{
				createBaggage("latency", "922337203685477580"),
				"latency",
			},
			922337203685477580,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBaggageLatency(tt.args.ctx, tt.args.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createBaggage(key string, value string) context.Context {
	member, _ := baggage.NewMember(key, value)
	bag, _ := baggage.New(member)
	ctx := baggage.ContextWithBaggage(context.TODO(), bag)
	return ctx
}
