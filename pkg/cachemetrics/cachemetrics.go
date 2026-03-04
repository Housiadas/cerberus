// Package cachemetrics implements sturdyc's MetricsRecorder and
// DistributedMetricsRecorder interfaces using OpenTelemetry metrics.
//
// This bridges sturdyc's internal cache events into the same OTel pipeline
// used by the rest of Cerberus, so cache hit rates, evictions, refresh
// patterns, and distributed storage interactions all appear alongside your
// HTTP metrics in Prometheus/Grafana.
//
// Usage:
//
//	recorder, err: = cachemetrics.NewMeterRecorder("user-cache")
//	if err != nil { ... }
//
//	cacheClient := sturdyc.New[any](
//	    capacity, numShards, ttl, evictionPercentage,
//	    sturdyc.WithMetrics(recorder),
//	    sturdyc.WithDistributedMetrics(recorder),
//	)
package cachemetrics

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const instrumentationName = "cerberus/cache"

// OTelRecorder implements both sturdyc.MetricsRecorder and
// sturdyc.DistributedMetricsRecorder, forwarding all cache events
// to OpenTelemetry counters, histograms, and gauges.
type OTelRecorder struct {
	attrs []attribute.KeyValue

	// In-memory cache metrics
	cacheHits    metric.Int64Counter
	cacheMisses  metric.Int64Counter
	asyncRefresh metric.Int64Counter
	syncRefresh  metric.Int64Counter
	missingRec   metric.Int64Counter

	// Eviction metrics
	forcedEvictions metric.Int64Counter
	entriesEvicted  metric.Int64Counter

	// Shard distribution
	shardIndex metric.Int64Counter

	// Batch refresh sizes
	batchRefreshSize metric.Int64Histogram

	// Cache size gauge (async observable)
	cacheSizeGauge metric.Int64ObservableGauge
	cacheSizeFn    func() int
	cacheSizeMu    sync.RWMutex

	// Distributed storage metrics
	distHits       metric.Int64Counter
	distMisses     metric.Int64Counter
	distRefreshes  metric.Int64Counter
	distMissingRec metric.Int64Counter
	distFallbacks  metric.Int64Counter
}

// NewMeterRecorder creates a new recorder for a named cache instance.
// The cacheName is used as an attribute on all metrics, so multiple sturdyc
// instances can be distinguished in dashboards.
func NewMeterRecorder(cacheName string) (*OTelRecorder, error) {
	meter := otel.Meter(instrumentationName)
	r := &OTelRecorder{
		attrs: []attribute.KeyValue{
			attribute.String("cache.name", cacheName),
		},
	}

	err := r.initLocalMetrics(meter)
	if err != nil {
		return nil, err
	}

	err = r.initDistributedMetrics(meter)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// =========================================================================
// sturdyc.MetricsRecorder implementation
// =========================================================================

// CacheHit records an in-memory cache hit.
func (r *OTelRecorder) CacheHit() {
	r.cacheHits.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}

// CacheMiss records an in-memory cache miss.
func (r *OTelRecorder) CacheMiss() {
	r.cacheMisses.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}

// AsynchronousRefresh records an asynchronous background refresh.
func (r *OTelRecorder) AsynchronousRefresh() {
	r.asyncRefresh.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}

// SynchronousRefresh records a synchronous refresh on the cache miss path.
func (r *OTelRecorder) SynchronousRefresh() {
	r.syncRefresh.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}

// MissingRecord records a missing record being served from cache.
func (r *OTelRecorder) MissingRecord() {
	r.missingRec.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}

// ForcedEviction records a forced eviction due to the cache being full.
func (r *OTelRecorder) ForcedEviction() {
	r.forcedEvictions.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}

// EntriesEvicted records the number of entries evicted.
func (r *OTelRecorder) EntriesEvicted(count int) {
	r.entriesEvicted.Add(context.Background(), int64(count), metric.WithAttributes(r.attrs...))
}

// ShardIndex records a cache writing for the given shard.
func (r *OTelRecorder) ShardIndex(index int) {
	shardAttrs := append(r.attrs, attribute.Int("cache.shard.index", index)) //nolint:gocritic
	r.shardIndex.Add(context.Background(), 1, metric.WithAttributes(shardAttrs...))
}

// CacheBatchRefreshSize records the size of a batch refresh operation.
func (r *OTelRecorder) CacheBatchRefreshSize(size int) {
	r.batchRefreshSize.Record(context.Background(), int64(size), metric.WithAttributes(r.attrs...))
}

// ObserveCacheSize registers a callback that the OTel SDK will poll to report
// the current number of entries in the cache.
func (r *OTelRecorder) ObserveCacheSize(callback func() int) {
	r.cacheSizeMu.Lock()
	r.cacheSizeFn = callback
	r.cacheSizeMu.Unlock()

	meter := otel.Meter(instrumentationName)

	gauge, err := meter.Int64ObservableGauge("cache.size",
		metric.WithDescription("Current number of entries in the cache"),
		metric.WithUnit("{entry}"),
		metric.WithInt64Callback(func(_ context.Context, o metric.Int64Observer) error {
			r.cacheSizeMu.RLock()
			fn := r.cacheSizeFn
			r.cacheSizeMu.RUnlock()

			if fn != nil {
				o.Observe(int64(fn()), metric.WithAttributes(r.attrs...))
			}

			return nil
		}),
	)
	if err != nil {
		return
	}

	r.cacheSizeGauge = gauge
}

// =========================================================================
// sturdyc.DistributedMetricsRecorder implementation
// =========================================================================

// DistributedCacheHit records a Redis cache hit.
func (r *OTelRecorder) DistributedCacheHit() {
	r.distHits.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}

// DistributedCacheMiss records a Redis cache miss.
func (r *OTelRecorder) DistributedCacheMiss() {
	r.distMisses.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}

// DistributedRefresh records a record retrieved from Redis that needed refresh.
func (r *OTelRecorder) DistributedRefresh() {
	r.distRefreshes.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}

// DistributedMissingRecord records a missing record retrieved from Redis.
func (r *OTelRecorder) DistributedMissingRecord() {
	r.distMissingRec.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}

// DistributedFallback records a fallback to Redis when a refresh failed.
func (r *OTelRecorder) DistributedFallback() {
	r.distFallbacks.Add(context.Background(), 1, metric.WithAttributes(r.attrs...))
}
