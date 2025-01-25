package middleware

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

//func initTracer(tempoEndpoint string) (*sdktrace.TracerProvider, error) {
//	ctx := context.Background()
//
//	// Create the gRPC exporter
//	exporter, err := otlptraces.New(ctx,
//		otlptracegrpc.NewClient(
//			otlptracegrpc.WithInsecure(),
//			otlptracegrpc.WithEndpoint(tempoEndpoint),
//			// In Docker Compose, "tempo:4317" is typically how you refer to the Tempo service
//		))
//	if err != nil {
//		return nil, err
//	}
//
//	// Build a Resource to describe this service
//	res, err := resource.New(ctx,
//		resource.WithAttributes(
//			attribute.String("service.name", "score-app"),
//		),
//	)
//	if err != nil {
//		return nil, err
//	}
//
//	// Create a trace provider with the exporter
//	tp := sdktrace.NewTracerProvider(
//		sdktrace.WithBatcher(exporter),
//		sdktrace.WithResource(res),
//	)
//
//	// Set as global trace provider
//	otel.SetTracerProvider(tp)
//
//	return tp, nil
//}

// Console Exporter, only for testing
func NewConsoleExporter() (oteltrace.SpanExporter, error) {
	return stdouttrace.New()
}

// OTLP Exporter
func NewOTLPExporter(ctx context.Context) (oteltrace.SpanExporter, error) {
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "tempo:4318"
	}

	// Change default HTTPS -> HTTP
	insecureOpt := otlptracehttp.WithInsecure()

	endpointOpt := otlptracehttp.WithEndpoint(otlpEndpoint)

	pathUrl := otlptracehttp.WithURLPath("/v1/traces")
	return otlptracehttp.New(ctx, insecureOpt, endpointOpt, pathUrl)
}

func NewTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	// Ensure default SDK resources and the required service name are set.
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("score-app"),
		),
	)

	if err != nil {
		panic(err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}
