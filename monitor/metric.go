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

// ExportLatencyMetrics exports latency metrics from baggage in ctx

func RecordRpcLatency(ctx context.Context, meter metric.Meter, connName string, env string) {
	rpcLatency := latencyHistogram{
		meter:                meter,
		histogramName:        "latency.rpc.histogram",
		histogramDescription: "Latency from block time to connector reception",
		connectorName:        connName,
		env:                  env,
		value:                0,
	}

	rpcObservation := getBaggageLatency(ctx, LatencyRpcKey)
	connObservation := getBaggageLatency(ctx, LatencyConnectorKey)

	if rpcObservation > 0 && connObservation > 0 {
		rpcLatency.value = connObservation - rpcObservation
	}

	recordLatencyHistogram(ctx, rpcLatency)
}

func RecordConnectorLatency(ctx context.Context, meter metric.Meter, connName string, env string) {
	connectorLatency := latencyHistogram{
		meter:                meter,
		histogramName:        "latency.connector.histogram",
		histogramDescription: "Latency from connector reception to kafka produce",
		connectorName:        connName,
		env:                  env,
		value:                0,
	}

	connObservation := getBaggageLatency(ctx, LatencyConnectorKey)
	kafkaProduceObservation := getBaggageLatency(ctx, LatencyKafkaProduceKey)

	if kafkaProduceObservation > 0 && connObservation > 0 {
		connectorLatency.value = kafkaProduceObservation - connObservation
	}

	recordLatencyHistogram(ctx, connectorLatency)
}

func RecordCoreLatency(ctx context.Context, meter metric.Meter, connName string, env string) {
	coreLatency := latencyHistogram{
		meter:                meter,
		histogramName:        "latency.core.histogram",
		histogramDescription: "Latency from kafka to streamserver",
		connectorName:        connName,
		env:                  env,
		value:                0,
	}

	kafkaProduceObservation := getBaggageLatency(ctx, LatencyKafkaProduceKey)
	ssConsumeObservation := getBaggageLatency(ctx, LatencyStreamserverConsumeKey)

	if ssConsumeObservation > 0 && kafkaProduceObservation > 0 {
		coreLatency.value = ssConsumeObservation - kafkaProduceObservation
	}

	recordLatencyHistogram(ctx, coreLatency)
}

func RecordSystemLatency(ctx context.Context, meter metric.Meter, connName string, env string) {
	systemLatency := latencyHistogram{
		meter:                meter,
		histogramName:        "latency.system.histogram",
		histogramDescription: "Latency from connector to streamserver",
		connectorName:        connName,
		env:                  env,
		value:                0,
	}

	connObservation := getBaggageLatency(ctx, LatencyConnectorKey)
	ssConsumeObservation := getBaggageLatency(ctx, LatencyStreamserverConsumeKey)

	if ssConsumeObservation > 0 && connObservation > 0 {
		systemLatency.value = ssConsumeObservation - connObservation
	}

	recordLatencyHistogram(ctx, systemLatency)
}

func RecordEndLatency(ctx context.Context, meter metric.Meter, connName string, env string) {
	e2eLatency := latencyHistogram{
		meter:                meter,
		histogramName:        "latency.e2e.histogram",
		histogramDescription: "Latency from block time to streamserver",
		connectorName:        connName,
		env:                  env,
		value:                0,
	}

	rpcObservation := getBaggageLatency(ctx, LatencyRpcKey)
	ssConsumeObservation := getBaggageLatency(ctx, LatencyStreamserverConsumeKey)

	if ssConsumeObservation > 0 && rpcObservation > 0 {
		e2eLatency.value = ssConsumeObservation - rpcObservation
	}

	recordLatencyHistogram(ctx, e2eLatency)
}

func recordLatencyHistogram(ctx context.Context, obs latencyHistogram) {
	if obs.value > 0 {
		// Convert latency from microseconds to float milliseconds
		latency := float64(obs.value) / 1000
		histogram, err := obs.meter.SyncFloat64().Histogram(
			obs.histogramName,
			instrument.WithUnit(unit.Milliseconds),
			instrument.WithDescription(obs.histogramDescription),
		)
		if err != nil {
			log.Error().Err(err).Str("connector", obs.connectorName).Str("histogram", obs.histogramName).Msg("Unable to create histogram")
		}
		histogram.Record(ctx, latency, attribute.String("Connector", obs.connectorName), attribute.String("Env", obs.env))
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
