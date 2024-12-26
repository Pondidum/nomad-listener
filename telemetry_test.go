package main

import (
	"context"
	"fmt"
	"testing"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func TestTraceParent(t *testing.T) {
	tp := tracesdk.NewTracerProvider()
	_, span := tp.Tracer("tests").Start(context.Background(), "test_trace_parent")

	traceId := span.SpanContext().TraceID().String()
	spanId := span.SpanContext().SpanID().String()

	id := traceParent(span)

	if id != fmt.Sprintf("00-%s-%s-01", traceId, spanId) {
		t.Error(id)
	}
}
