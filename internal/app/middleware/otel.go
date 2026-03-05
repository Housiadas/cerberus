package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	"go.opentelemetry.io/otel/trace"
)

type middlewareMetrics struct {
	requestCount    metric.Int64Counter
	requestDuration metric.Float64Histogram
	requestsActive  metric.Int64UpDownCounter
}

// Otel records HTTP server metrics (request count, duration, active requests).
func (m *Middleware) Otel() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// route
			attrs := []attribute.KeyValue{
				semconv.HTTPRequestMethodKey.String(r.Method),
				semconv.URLPath(r.URL.Path),
			}

			// Start a server span.
			ctx, span := m.tracer.Start(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path),
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(attrs...),
			)
			defer span.End()

			// inject trace to context after span is started so trace ID is real
			ctx = telemetry.InjectTracing(ctx, m.tracer)

			// Track in-flight requests.
			if m.metrics.requestsActive != nil {
				m.metrics.requestsActive.Add(ctx, 1, metric.WithAttributes(attrs...))
				defer m.metrics.requestsActive.Add(ctx, -1, metric.WithAttributes(attrs...))
			}

			// response recorder
			rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(ctx))

			responseAttrs := make([]attribute.KeyValue, len(attrs)+1)
			copy(responseAttrs, attrs)
			responseAttrs[len(attrs)] = semconv.HTTPResponseStatusCode(rec.statusCode)

			if m.metrics.requestCount != nil {
				m.metrics.requestCount.Add(ctx, 1, metric.WithAttributes(responseAttrs...))
			}

			duration := time.Since(start)
			if m.metrics.requestDuration != nil {
				m.metrics.requestDuration.Record(
					ctx,
					duration.Seconds(),
					metric.WithAttributes(responseAttrs...),
				)
			}

			span.SetAttributes(semconv.HTTPResponseStatusCode(rec.statusCode))

			if rec.statusCode >= 500 {
				span.SetStatus(codes.Error, http.StatusText(rec.statusCode))
			}
		})
	}
}

// initOTelMetrics initializes the OTel HTTP server metrics used by Otel().
// Errors are non-fatal: if a metric fails to create, it is left nil and
// silently skipped in the middleware.
func initOTelMetrics(
	ctx context.Context,
	meter metric.Meter,
	log logger.Logger,
) middlewareMetrics {
	requestCount, err := meter.Int64Counter("http.server.request.total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		log.Error(ctx, "metric int64counter error", "msg", err)
	}

	requestDuration, err := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(
			0.005,
			0.01,
			0.025,
			0.05,
			0.1,
			0.25,
			0.5,
			1,
			2.5,
			5,
			10,
		),
	)
	if err != nil {
		log.Error(ctx, "metric float64Histogram error", "msg", err)
	}

	requestsActive, err := meter.Int64UpDownCounter("http.server.active_requests",
		metric.WithDescription("Number of in-flight HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		log.Error(ctx, "metric int64downCounter error", "msg", err)
	}

	return middlewareMetrics{
		requestCount:    requestCount,
		requestDuration: requestDuration,
		requestsActive:  requestsActive,
	}
}
