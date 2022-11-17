package monitor

import (
	"context"
	"strconv"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/unit"
)

type latencyHistogram struct {
	meter                metric.Meter
	histogramName        string
	histogramDescription string
	connectorName        string
	env                  string
	value                int64
}

const (
	// Baggage keys
	LatencyRpcKey                 = "latencyrpc"
	LatencyConnectorKey           = "latencyconn"
	LatencyKafkaProduceKey        = "latencykafkaprod"
	LatencyStreamserverConsumeKey = "latencyssconsume"
)

// ExportLatencyMetrics exports all latency metrics from baggage in ctx
func ExportLatencyMetrics(ctx context.Context, meter metric.Meter, connectorNmae string, env string) {
	latency := latencyHistogram{
		meter:         meter,
		connectorName: connectorName,
		env:           env,
		value:         0,
	}

	recordRpcLatency(ctx, latency)
	recordConnectorLatency(ctx, latency)
	recordCoreLatency(ctx, latency)
	recordSystemLatency(ctx, latency)
	recordEndLatency(ctx, latency)
}

func recordRpcLatency(ctx context.Context, latency latencyHistogram) {
	latency.histogramName = "latency.rpc.histogram"
	latency.histogramDescription = "Latency from block time to connector reception"

	rpcObservation := getBaggageLatency(ctx, LatencyRpcKey)
	connObservation := getBaggageLatency(ctx, LatencyConnectorKey)

	if rpcObservation > 0 && connObservation > 0 {
		latency.value = connObservation - rpcObservation
	}

	recordLatencyHistogram(ctx, latency)
}

func recordConnectorLatency(ctx context.Context, latency latencyHistogram) {
	latency.histogramName = "latency.connector.histogram"
	latency.histogramDescription = "Latency from connector reception to kafka produce"

	connObservation := getBaggageLatency(ctx, LatencyConnectorKey)
	kafkaProduceObservation := getBaggageLatency(ctx, LatencyKafkaProduceKey)

	if kafkaProduceObservation > 0 && connObservation > 0 {
		latency.value = kafkaProduceObservation - connObservation
	}

	recordLatencyHistogram(ctx, latency)
}

func recordCoreLatency(ctx context.Context, latency latencyHistogram) {
	latency.histogramName = "latency.core.histogram"
	latency.histogramDescription = "Latency from kafka to streamserver"

	kafkaProduceObservation := getBaggageLatency(ctx, LatencyKafkaProduceKey)
	ssConsumeObservation := getBaggageLatency(ctx, LatencyStreamserverConsumeKey)

	if ssConsumeObservation > 0 && kafkaProduceObservation > 0 {
		latency.value = ssConsumeObservation - kafkaProduceObservation
	}

	recordLatencyHistogram(ctx, latency)
}

func recordSystemLatency(ctx context.Context, latency latencyHistogram) {
	latency.histogramName = "latency.system.histogram"
	latency.histogramDescription = "Latency from connector to streamserver"

	connObservation := getBaggageLatency(ctx, LatencyConnectorKey)
	ssConsumeObservation := getBaggageLatency(ctx, LatencyStreamserverConsumeKey)

	if ssConsumeObservation > 0 && connObservation > 0 {
		latency.value = ssConsumeObservation - connObservation
	}

	recordLatencyHistogram(ctx, latency)
}

func recordEndLatency(ctx context.Context, latency latencyHistogram) {
	latency.histogramName = "latency.e2e.histogram"
	latency.histogramDescription = "Latency from block time to streamserver"

	rpcObservation := getBaggageLatency(ctx, LatencyRpcKey)
	ssConsumeObservation := getBaggageLatency(ctx, LatencyStreamserverConsumeKey)

	if ssConsumeObservation > 0 && rpcObservation > 0 {
		latency.value = ssConsumeObservation - rpcObservation
	}

	recordLatencyHistogram(ctx, latency)
}

func recordLatencyHistogram(ctx context.Context, latency latencyHistogram) {
	if latency.value > 0 {
		// Convert latency from microseconds to float milliseconds
		latencyValue := float64(latency.value) / 1000
		histogram, err := latency.meter.SyncFloat64().Histogram(
			latency.histogramName,
			instrument.WithUnit(unit.Milliseconds),
			instrument.WithDescription(latency.histogramDescription),
		)
		if err != nil {
			log.Error().Err(err).Str("connector", latency.connectorName).Str("histogram", latency.histogramName).Msg("Unable to create histogram")
		}
		attributes := []attribute.KeyValue{
			attribute.String("Connector", latency.connectorName),
			attribute.String("Env", latency.env),
		}
		histogram.Record(ctx, latencyValue, attributes...)
	}
}

func getBaggageLatency(ctx context.Context, key string) int64 {
	bag := baggage.FromContext(ctx)

	mem := bag.Member(key)
	ts, err := strconv.Atoi(mem.Value())
	// Baggage key does not exist
	if err != nil {
		return 0
	}
	if ts > 0 {
		return int64(ts)
	} else {
		return 0
	}
}
