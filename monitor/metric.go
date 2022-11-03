package monitor

import (
	"context"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
)

type Histogram struct {
	name        string
	description string
}

const (
	// Baggage keys
	LatencyRpcKey                 = "latencyrpc"
	LatencyConnectorKey           = "latencyconn"
	LatencyKafkaProduceKey        = "latencykafkaprod"
	LatencyStreamserverConsumeKey = "latencyssconsume"

	// Metric keys
	rpcLatency    = 0
	connLatency   = 1
	coreLatency   = 2
	systemLatency = 3
	e2eLatency    = 4
)

var (
	latencyObservations = map[string]int64{
		LatencyRpcKey:                 0,
		LatencyConnectorKey:           0,
		LatencyKafkaProduceKey:        0,
		LatencyStreamserverConsumeKey: 0,
	}
	latencyMetrics = map[int]int64{
		rpcLatency:    0,
		connLatency:   0,
		coreLatency:   0,
		systemLatency: 0,
		e2eLatency:    0,
	}
	histograms = map[int]Histogram{
		rpcLatency:    {name: "latency.rpc.histogram", description: "Latency from block time to connector reception"},
		connLatency:   {name: "latency.connector.histogram", description: "Latency from connector reception to kafka produce"},
		coreLatency:   {name: "latency.core.histogram", description: "Latency from kafka to streamserver"},
		systemLatency: {name: "latency.system.histogram", description: "Latency from connector to streamserver"},
		e2eLatency:    {name: "latency.e2e.histogram", description: "Latency from block time to streamserver"},
	}
)

// ExportLatencyMetrics exports latency metrics from baggage in ctx
func ExportLatencyMetrics(ctx context.Context, meter metric.Meter, connName string, env string) {
	bag := baggage.FromContext(ctx)

	// Extract latency observations from baggage
	for key := range latencyObservations {
		latencyObservations[key] = getBaggageLatency(bag, key)
	}

	// Get observations
	rpcObservation := latencyObservations[LatencyRpcKey]
	connObservation := latencyObservations[LatencyConnectorKey]
	kafkaProduceObservation := latencyObservations[LatencyKafkaProduceKey]
	ssConsumeObservation := latencyObservations[LatencyStreamserverConsumeKey]

	// Derive latency metrics from observations
	if rpcObservation > 0 && connObservation > 0 {
		latencyMetrics[rpcLatency] = connObservation - rpcObservation
	} else {
		latencyMetrics[rpcLatency] = 0
	}

	if kafkaProduceObservation > 0 && connObservation > 0 {
		latencyMetrics[connLatency] = kafkaProduceObservation - connObservation
	} else {
		latencyMetrics[connLatency] = 0
	}

	if ssConsumeObservation > 0 && kafkaProduceObservation > 0 {
		latencyMetrics[coreLatency] = ssConsumeObservation - kafkaProduceObservation
	} else {
		latencyMetrics[coreLatency] = 0
	}

	if ssConsumeObservation > 0 && connObservation > 0 {
		latencyMetrics[systemLatency] = ssConsumeObservation - connObservation
	} else {
		latencyMetrics[systemLatency] = 0
	}

	if ssConsumeObservation > 0 && rpcObservation > 0 {
		latencyMetrics[e2eLatency] = ssConsumeObservation - rpcObservation
	} else {
		latencyMetrics[e2eLatency] = 0
	}

	// Record histograms
	for key, hist := range histograms {
		latency := latencyMetrics[key]
		if latency > 0 {
			histogramMetric, err := meter.SyncInt64().Histogram(
				hist.name,
				instrument.WithUnit("microseconds"),
				instrument.WithDescription(hist.description),
			)
			if err != nil {
				log.Error().Err(err).Str("histogram", hist.name).Msg("Unable to create histogram")
			}
			histogramMetric.Record(ctx, latency, attribute.String("Connector", connName), attribute.String("Env", env))
		}
	}
}

func getBaggageLatency(bag baggage.Baggage, key string) int64 {
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

// ApiUsageMiddleware can be added to gin engine to export metrics for api calls by origin, token, and path.
func ApiUsageMiddleware(name string, meter metric.Meter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Start: staging-only
		// These debug messages are only for testing how traefik interacts with requests
		// These will be removed
		log.Debug().Str("authorization", ctx.Request.Header.Get("Authorization")).Str("origin", ctx.Request.Header.Get("Origin")).Msg("auth headers")
		log.Debug().Str("clientIP", ctx.ClientIP()).Str("remoteIP", ctx.RemoteIP()).Str("remoteaddr", ctx.Request.RemoteAddr).Msg("request IP and addr")
		log.Debug().Str("referer", ctx.Request.Referer()).Msg("Referer")
		// End: staging-only
		origin, token := getApiAuthHeaders(ctx)
		path := ctx.Request.URL.EscapedPath()
		counter, err := meter.SyncInt64().Counter(name)
		if err != nil {
			log.Err(err).Msg("Unable to record usage observation")
		}
		counter.Add(context.TODO(), 1, attribute.String("origin", origin), attribute.String("token", token), attribute.String("path", path))

		// serve the request to the next middleware
		ctx.Next()
	}
}

// getApiAuthHeaders gets Origin and Authorization headers to be used as attributes in usage metrics.
// Note that while getApiAuthHeaders checks that both Origin and Authorization headers are set, it does
// not validate that the returned origin and token exist in the database.
func getApiAuthHeaders(ctx *gin.Context) (string, string) {
	origin := ctx.Request.Header.Get("Origin")
	token := ctx.Request.Header.Get("Authorization")
	if origin == "" || token == "" {
		return "anonymous", "none"
	}

	// Remove Bearer prefix
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	} else {
		return "anonymous", "none"
	}

	return origin, token
}
