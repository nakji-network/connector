package monitor

import (
	"context"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/baggage"
)

// Takes in baggage member key and value, creates baggage, adds to context, and returns context.
func NewLatencyBaggage(ctx context.Context, memberKey string, memberVal time.Time) context.Context {
	// Baggage member values must be strings
	latencyStr := strconv.Itoa(int(memberVal.UnixMilli()))

	bag := baggage.FromContext(ctx)

	member, err := baggage.NewMember(memberKey, latencyStr)
	if err != nil {
		log.Err(err).Msg("Unable to create baggage member from latency obseration")
	}

	bag, err = bag.SetMember(member)
	if err != nil {
		log.Err(err).Msg("Unable to create baggage from latency obseration")
	}

	ctx = baggage.ContextWithBaggage(ctx, bag)

	return ctx
}
