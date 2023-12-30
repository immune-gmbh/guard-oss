package telemetry

import (
	"context"
	"os"

	semconv "go.opentelemetry.io/collector/semconv/v1.9.0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

var traceProvider *sdktrace.TracerProvider

func setupTraceOpenTelemetry(endpoint string, token string, tls bool) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithCompressor(gzip.Name),
	}

	if tls {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")))
	} else {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if token != "" {
		headers := map[string]string{
			"lightstep-access-token": token,
		}
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}

	return otlptrace.New(context.Background(), otlptracegrpc.NewClient(opts...))
}

func setupTraceJaeger(endpoint string) (sdktrace.SpanExporter, error) {
	return jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)),
	)
}

func setupTrace(spanExporter sdktrace.SpanExporter, service string, repoUrl string, tag string) (func(), error) {
	defaultTracer = nil
	Version = tag

	var flush func()
	var err error

	attrs := append([]attribute.KeyValue{},
		attribute.String(semconv.AttributeServiceName, service),
		attribute.String(semconv.AttributeServiceVersion, tag),
		attribute.String("code.repository", repoUrl),
	)
	if hostname, err := os.Hostname(); err == nil {
		attrs = append(attrs, attribute.String(semconv.AttributeHostName, hostname))
	}
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes("", attrs...),
	)
	if err != nil {
		return flush, err
	}

	bsp := sdktrace.NewBatchSpanProcessor(spanExporter)
	traceProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(r),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{}, propagation.Baggage{}))

	tracer := otel.Tracer("")
	defaultTracer = &tracer
	flush = func() {
		_ = bsp.Shutdown(context.Background())
		_ = spanExporter.Shutdown(context.Background())
	}

	return flush, nil
}
