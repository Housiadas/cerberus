package cachemetrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

func (r *OTelRecorder) initLocalMetrics(meter metric.Meter) error {
	var err error

	r.cacheHits, err = meter.Int64Counter("cache.hit",
		metric.WithDescription("Number of in-memory cache hits"),
		metric.WithUnit("{hit}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.hit counter: %w", err)
	}

	r.cacheMisses, err = meter.Int64Counter("cache.miss",
		metric.WithDescription("Number of in-memory cache misses"),
		metric.WithUnit("{miss}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.miss counter: %w", err)
	}

	r.asyncRefresh, err = meter.Int64Counter("cache.refresh.async",
		metric.WithDescription("Number of asynchronous background refreshes"),
		metric.WithUnit("{refresh}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.refresh.async counter: %w", err)
	}

	r.syncRefresh, err = meter.Int64Counter("cache.refresh.sync",
		metric.WithDescription("Number of synchronous refreshes (cache miss path)"),
		metric.WithUnit("{refresh}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.refresh.sync counter: %w", err)
	}

	r.missingRec, err = meter.Int64Counter("cache.missing_record",
		metric.WithDescription("Number of times a missing record was served from cache"),
		metric.WithUnit("{record}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.missing_record counter: %w", err)
	}

	r.forcedEvictions, err = meter.Int64Counter("cache.eviction.forced",
		metric.WithDescription("Number of forced evictions (cache full)"),
		metric.WithUnit("{eviction}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.eviction.forced counter: %w", err)
	}

	r.entriesEvicted, err = meter.Int64Counter("cache.eviction.entries",
		metric.WithDescription("Total entries evicted"),
		metric.WithUnit("{entry}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.eviction.entries counter: %w", err)
	}

	r.shardIndex, err = meter.Int64Counter("cache.shard.write",
		metric.WithDescription("Write distribution across shards"),
		metric.WithUnit("{write}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.shard.write counter: %w", err)
	}

	r.batchRefreshSize, err = meter.Int64Histogram("cache.batch.refresh_size",
		metric.WithDescription("Size of batch refresh operations"),
		metric.WithUnit("{record}"),
		metric.WithExplicitBucketBoundaries(1, 5, 10, 25, 50, 100, 250, 500),
	)
	if err != nil {
		return fmt.Errorf("creating cache.batch.refresh_size histogram: %w", err)
	}

	return nil
}
