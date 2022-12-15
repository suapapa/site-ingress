package main

// import (
// 	"context"
// 	"os"

// 	"go.opentelemetry.io/otel"
// 	"go.opentelemetry.io/otel/attribute"
// 	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
// 	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
// 	"go.opentelemetry.io/otel/metric/global"
// 	"go.opentelemetry.io/otel/propagation"
// 	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
// 	"go.opentelemetry.io/otel/sdk/resource"
// 	sdktrace "go.opentelemetry.io/otel/sdk/trace"
// 	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
// 	"go.opentelemetry.io/otel/trace"
// )

// var (
// 	tracer trace.Tracer
// )

// func initTracerProvider(ctx context.Context, url string) *sdktrace.TracerProvider {
// 	// ctx := context.Background()

// 	exporter, err := otlptracegrpc.New(
// 		ctx,
// 		otlptracegrpc.WithEndpoint(url),
// 		otlptracegrpc.WithInsecure(),
// 	)
// 	if err != nil {
// 		log.Fatalf("new otlp trace grpc exporter failed: %v", err)
// 	}
// 	tp := sdktrace.NewTracerProvider(
// 		sdktrace.WithBatcher(exporter),
// 		sdktrace.WithResource(resource.NewWithAttributes(
// 			semconv.SchemaURL,
// 			semconv.ServiceNameKey.String("homin-dev"),
// 			attribute.String("name", programName),
// 			attribute.String("ver", programVer),
// 			attribute.String("k8s-node", os.Getenv("K8S_NODE_NAME")),
// 		)),
// 	)
// 	otel.SetTracerProvider(tp)
// 	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
// 	return tp
// }

// func initMeterProvider(ctx context.Context, url string) *sdkmetric.MeterProvider {
// 	// ctx := context.Background()

// 	exporter, err := otlpmetricgrpc.New(
// 		ctx,
// 		otlpmetricgrpc.WithEndpoint(url),
// 		otlpmetricgrpc.WithInsecure(),
// 	)
// 	if err != nil {
// 		log.Fatalf("new otlp metric grpc exporter failed: %v", err)
// 	}

// 	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)))
// 	global.SetMeterProvider(mp)
// 	return mp
// }
