package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func configureTelemetry(ctx context.Context) *sdktrace.TracerProvider {
	exporter, _ := otlptracegrpc.New(ctx)
	res, _ := resource.New(
		ctx,
		resource.WithTelemetrySDK(),
		resource.WithFromEnv(),
		resource.WithAttributes(
			semconv.ServiceName("nomad-listener"),
			semconv.ServiceVersion(version()),
		),
	)

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider
}

func traceError(span trace.Span, err error) error {
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())

	return err
}

func traceParent(span trace.Span) string {

	sc := span.SpanContext()
	flags := sc.TraceFlags() & trace.FlagsSampled

	traceID := sc.TraceID()
	spanID := sc.SpanID()
	flagByte := [1]byte{byte(flags)}

	return fmt.Sprintf("00-%s-%s-%s",
		traceID,
		spanID,
		hex.EncodeToString(flagByte[:]),
	)

}
