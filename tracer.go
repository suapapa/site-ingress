package main

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// tracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Jaeger exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
// func tracerProvider(url string) (*sdktrace.TracerProvider, error) {
// 	// Create the Jaeger exporter
// 	// exp, err := jaeger.New(jaeger.WithCollectorEndpoint()) //  "http://localhost:14268/api/traces"
// 	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
// 	if err != nil {
// 		return nil, err
// 	}
// 	tp := sdktrace.NewTracerProvider(
// 		// Always be sure to batch in production.
// 		sdktrace.WithBatcher(exp),
// 		// Record information about this application in a Resource.
// 		sdktrace.WithResource(resource.NewWithAttributes(
// 			semconv.SchemaURL,
// 			semconv.ServiceNameKey.String("homin-dev"),
// 			attribute.String("name", programName),
// 			attribute.String("ver", programVer),
// 			// attribute.Int64("ID", id),
// 		)),
// 	)
// 	return tp, nil
// }

var (
	tracer trace.Tracer
)

func initTracerProvider(ctx context.Context, url string) *sdktrace.TracerProvider {
	// ctx := context.Background()

	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(url),
	)
	if err != nil {
		log.Fatalf("new otlp trace grpc exporter failed: %v", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		// sdktrace.WithResource(resource.NewWithAttributes(
		// 	semconv.SchemaURL,
		// 	semconv.ServiceNameKey.String("homin-dev"),
		// 	attribute.String("name", programName),
		// 	attribute.String("ver", programVer),
		// )),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}

func initMeterProvider(ctx context.Context, url string) *sdkmetric.MeterProvider {
	// ctx := context.Background()

	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(url),
	)
	if err != nil {
		log.Fatalf("new otlp metric grpc exporter failed: %v", err)
	}

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)))
	global.SetMeterProvider(mp)
	return mp
}
