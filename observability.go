package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

const (
	OtelMetricExportInterval       = 1 * time.Minute
	OtelTraceExportInterval        = 1 * time.Minute
	OtelLogExportInterval          = 10 * time.Second
	OtelProtocolHTTP               = "http"
	OtelProtocolGRPC               = "grpc"
	OtelSpanLimitAttributePerEvent = 16
	OtelSpanLimitEventCount        = 64
)

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context, config OtelConfig, serviceResource *resource.Resource) (*trace.TracerProvider, error) {
	var exporter trace.SpanExporter
	var err error

	switch config.OtelProtocol() {
	case OtelProtocolGRPC:
		exporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithEndpointURL(config.TracesURL()))
	case OtelProtocolHTTP:
		exporter, err = otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(config.TracesURL()))
	default:
		return nil, fmt.Errorf("unknown observability protocol %q", config.Protocol)
	}

	if err != nil {
		return nil, err
	}

	spanLimits := trace.SpanLimits{
		AttributePerEventCountLimit: OtelSpanLimitAttributePerEvent,
		EventCountLimit:             OtelSpanLimitEventCount,
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithResource(serviceResource),
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(OtelTraceExportInterval),
		),
		trace.WithRawSpanLimits(spanLimits),
		trace.WithSampler(trace.AlwaysSample()),
	)

	return traceProvider, nil
}

func newMeterProvider(ctx context.Context, config OtelConfig, serviceResource *resource.Resource) (*metric.MeterProvider, error) {
	var exporter metric.Exporter
	var err error

	switch config.OtelProtocol() {
	case OtelProtocolGRPC:
		exporter, err = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpointURL(config.MetricsURL()))
	case OtelProtocolHTTP:
		exporter, err = otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(config.MetricsURL()))
	default:
		return nil, fmt.Errorf("unknown observability protocol %q", config.Protocol)
	}

	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(serviceResource),
		metric.WithReader(metric.NewPeriodicReader(exporter,
			metric.WithInterval(OtelMetricExportInterval))),
	)

	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context, config OtelConfig, serviceResource *resource.Resource) (*log.LoggerProvider, error) {
	var exporter log.Exporter
	var err error

	switch config.OtelProtocol() {
	case OtelProtocolGRPC:
		exporter, err = otlploggrpc.New(ctx, otlploggrpc.WithEndpointURL(config.LogsURL()))
	case OtelProtocolHTTP:
		exporter, err = otlploghttp.New(ctx, otlploghttp.WithEndpointURL(config.LogsURL()))
	default:
		return nil, fmt.Errorf("unknown observability protocol %q", config.Protocol)
	}

	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithResource(serviceResource),
		log.WithProcessor(log.NewBatchProcessor(exporter,
			log.WithExportInterval(OtelLogExportInterval))),
	)

	return loggerProvider, nil
}

func newServiceResource(ctx context.Context, hostData HostData) (*resource.Resource, error) {
	providerBinary, err := os.Executable()
	if err != nil {
		return nil, err
	}

	return resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(hostData.ProviderKey),
			semconv.HostIDKey.String(hostData.HostID),
			semconv.ServiceInstanceIDKey.String(hostData.InstanceID),
			semconv.ProcessExecutableNameKey.String(filepath.Base(providerBinary)),
		),
	)
}
