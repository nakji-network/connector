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
func ExportLatencyMetrics(ctx context.Context, meter metric.Meter, connectorName string, env string) {
	for _, latency := range []latencyHistogram{
		getRPCLatency(ctx, connectorName, env),
		getConnectorLatency(ctx, connectorName, env),
		getCoreLatency(ctx, connectorName, env),
		getSystemLatency(ctx, connectorName, env),
		getEndLatency(ctx, connectorName, env),
	} {
		recordLatencyHistogram(ctx, meter, latency)
	}
}

func getRPCLatency(ctx context.Context, connectorName string, env string) latencyHistogram {
	latency := latencyHistogram{
		histogramName:        "latency.rpc.histogram",
		histogramDescription: "Latency from block time to connector reception",
		connectorName:        connectorName,
		env:                  env,
		value:                0,
	}

	rpcObservation := getBaggageLatency(ctx, LatencyRpcKey)
	connObservation := getBaggageLatency(ctx, LatencyConnectorKey)

	if rpcObservation > 0 && connObservation > 0 {
		latency.value = connObservation - rpcObservation
	}

	return latency
}

func getConnectorLatency(ctx context.Context, connectorName string, env string) latencyHistogram {
	latency := latencyHistogram{
		histogramName:        "latency.connector.histogram",
		histogramDescription: "Latency from connector reception to kafka produce",
		connectorName:        connectorName,
		env:                  env,
		value:                0,
	}

	connObservation := getBaggageLatency(ctx, LatencyConnectorKey)
	kafkaProduceObservation := getBaggageLatency(ctx, LatencyKafkaProduceKey)

	if kafkaProduceObservation > 0 && connObservation > 0 {
		latency.value = kafkaProduceObservation - connObservation
	}

	return latency
}

func getCoreLatency(ctx context.Context, connectorName string, env string) latencyHistogram {
	latency := latencyHistogram{
		histogramName:        "latency.core.histogram",
		histogramDescription: "Latency from kafka to streamserver",
		connectorName:        connectorName,
		env:                  env,
		value:                0,
	}

	kafkaProduceObservation := getBaggageLatency(ctx, LatencyKafkaProduceKey)
	ssConsumeObservation := getBaggageLatency(ctx, LatencyStreamserverConsumeKey)

	if ssConsumeObservation > 0 && kafkaProduceObservation > 0 {
		latency.value = ssConsumeObservation - kafkaProduceObservation
	}

	return latency
}

func getSystemLatency(ctx context.Context, connectorName string, env string) latencyHistogram {
	latency := latencyHistogram{
		histogramName:        "latency.system.histogram",
		histogramDescription: "Latency from connector to streamserver",
		connectorName:        connectorName,
		env:                  env,
		value:                0,
	}

	connObservation := getBaggageLatency(ctx, LatencyConnectorKey)
	ssConsumeObservation := getBaggageLatency(ctx, LatencyStreamserverConsumeKey)

	if ssConsumeObservation > 0 && connObservation > 0 {
		latency.value = ssConsumeObservation - connObservation
	}

	return latency
}

func getEndLatency(ctx context.Context, connectorName string, env string) latencyHistogram {
	latency := latencyHistogram{
		histogramName:        "latency.e2e.histogram",
		histogramDescription: "Latency from block time to streamserver",
		connectorName:        connectorName,
		env:                  env,
		value:                0,
	}

	rpcObservation := getBaggageLatency(ctx, LatencyRpcKey)
	ssConsumeObservation := getBaggageLatency(ctx, LatencyStreamserverConsumeKey)

	if ssConsumeObservation > 0 && rpcObservation > 0 {
		latency.value = ssConsumeObservation - rpcObservation
	}

	return latency
}

func recordLatencyHistogram(ctx context.Context, meter metric.Meter, latency latencyHistogram) {
	if latency.value > 0 {
		// Convert latency from microseconds to float milliseconds
		latencyValue := float64(latency.value) / 1000
		histogram, err := meter.SyncFloat64().Histogram(
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
