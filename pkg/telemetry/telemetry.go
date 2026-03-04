// Package telemetry provides OpenTelemetry initialization and shutdown for
// Cerberus. It configures tracing and metrics via OTLP, and exposes helpers
// for injecting trace context into requests.
package telemetry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const defaultTraceID = "00000000000000000000000000000000"

var (
	errCastTracer = errors.New("error casting tracer")
	errCastMeter  = errors.New("error casting meter")
)

// Config defines the information needed to initialize tracing and metrics.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Namespace      string
	// OTLPEndpoint is the gRPC endpoint for the OTel collector, e.g. "localhost:4317".
	// When empty, no telemetry is exported
	OTLPEndpoint string
	// ExcludedRoutes lists HTTP paths that should never be traced (e.g., health checks).
	ExcludedRoutes map[string]struct{}
	// TraceSampleRate controls the fraction of traces sampled (0.0–1.0).
	// Defaults to 1.0 when not set.
	TraceSampleRate float64
	// MetricInterval is how often metrics are exported. Defaults is 30s.
	MetricInterval time.Duration
}

// Provider wraps the OTel SDK providers for clean lifecycle management.
type Provider struct {
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
}

// New initializes OpenTelemetry tracing and metrics with OTLP exporters and
// returns a Provider. The caller must call Provider.Shutdown on application exit.
// When cfg.OTLPEndpoint is empty, noop providers are used.
func New(ctx context.Context, cfg Config) (*Provider, error) {
	if cfg.TraceSampleRate == 0 {
		cfg.TraceSampleRate = 1.0
	}

	if cfg.MetricInterval == 0 {
		cfg.MetricInterval = 30 * time.Second
	}

	res, err := newResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating otel resource: %w", err)
	}

	tp, err := newTracerProvider(ctx, cfg, res)
	if err != nil {
		return nil, fmt.Errorf("error creating tracer provider: %w", err)
	}

	mp, err := newMeterProvider(ctx, cfg, res)
	if err != nil {
		_ = tp.Shutdown(ctx)

		return nil, fmt.Errorf("error creating meter provider: %w", err)
	}

	// registered global tracer and meter providers
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{
		tracerProvider: tp,
		meterProvider:  mp,
	}, nil
}

// TracerProvider returns the underlying tracer.
func (p *Provider) TracerProvider() trace.TracerProvider {
	return p.tracerProvider
}

// MeterProvider returns the underlying meter.
func (p *Provider) MeterProvider() metric.MeterProvider {
	return p.meterProvider
}

// Shutdown flushes and shuts down all telemetry providers.
func (p *Provider) Shutdown(ctx context.Context) error {
	errTracer := p.shutdownTracer(ctx)
	errMeter := p.shutdownMeter(ctx)

	return errors.Join(errTracer, errMeter)
}

// Shutdown flushes and shuts down tracer.
func (p *Provider) shutdownTracer(ctx context.Context) error {
	tp, ok := p.tracerProvider.(*sdktrace.TracerProvider)
	if !ok {
		return errCastTracer
	}

	err := tp.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("error shutting down tracer: %w", err)
	}

	return nil
}

// Shutdown flushes and shuts down meter.
func (p *Provider) shutdownMeter(ctx context.Context) error {
	mp, ok := p.meterProvider.(*sdkmetric.MeterProvider)
	if !ok {
		return errCastMeter
	}

	err := mp.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("shutting down meter: %w", err)
	}

	return nil
}

func newResource(cfg Config) (*resource.Resource, error) {
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceNamespace(cfg.Namespace),
			semconv.ServiceVersion(cfg.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error merging otel resources: %w", err)
	}

	return res, nil
}
