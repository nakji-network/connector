package monitor

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func CreateTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}

func StartSpan(ctx context.Context, tr trace.Tracer, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	newCtx, span := tr.Start(ctx, name, opts...)
	return newCtx, span
}

func EndSpan(span trace.Span, err error, kv ...attribute.KeyValue) {
	span.SetAttributes(kv...)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}

	span.End()
}
