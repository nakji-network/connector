package monitor

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const DefaultTracerName = "github.com/nakji-network/connector/monitor"

// InitTracerProvider creates and registers a new TracerProvider globally.
func InitTracerProvider(ctx context.Context, host, name, version, env string, sampleRatio float64) (*trace.TracerProvider, error) {
	conn, err := grpc.DialContext(ctx, host, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to opentelemetry collector")
		return nil, err
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		log.Error().Err(err).Msg("failed to set up a traceExporter")
		return nil, err
	}

	// Set up a resource
	r, err := newResource(name, version, env)
	if err != nil {
		log.Error().Err(err).Msg("failed to set up a resource")
		return nil, err
	}

	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(sampleRatio))),
		trace.WithBatcher(traceExporter),
		trace.WithResource(r),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	log.Info().Msg("initialized the trace provider")
	return tp, err
}

func InitMeterProvider(ctx context.Context, host string) (*metric.MeterProvider, error) {
	conn, err := grpc.DialContext(ctx, host, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to opentelemetry collector")
		return nil, err
	}
	exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		log.Error().Err(err).Msg("failed to set up metric exporter")
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))

	return meterProvider, nil
}

// newResource returns a resource describing this service.
func newResource(name, version, env string) (*resource.Resource, error) {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersionKey.String(version),
			semconv.DeploymentEnvironmentKey.String(env),
		),
	)
	return r, err
}

var _ propagation.TextMapCarrier = (*MessageCarrier)(nil)

// MessageCarrier injects and extracts traces from a *kafka.Message.
type MessageCarrier struct {
	msg *kafka.Message
}

// NewMessageCarrier creates a new MessageCarrier.
func NewMessageCarrier(msg *kafka.Message) MessageCarrier {
	return MessageCarrier{msg: msg}
}

// Get retrieves a single value for a given key.
func (c MessageCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

// Set sets a header.
func (c MessageCarrier) Set(key, val string) {
	// Ensure uniqueness of keys
	for i := 0; i < len(c.msg.Headers); i++ {
		if c.msg.Headers[i].Key == key {
			c.msg.Headers = append(c.msg.Headers[:i], c.msg.Headers[i+1:]...)
			i--
		}
	}
	c.msg.Headers = append(c.msg.Headers, kafka.Header{
		Key:   key,
		Value: []byte(val),
	})
}

// Keys returns a slice of all key identifiers in the carrier.
func (c MessageCarrier) Keys() []string {
	out := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		out[i] = h.Key
	}
	return out
}
