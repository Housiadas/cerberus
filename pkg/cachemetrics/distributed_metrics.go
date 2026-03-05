package cachemetrics

import (
	"fmt"

	"go.opentelemetry.io/otel/metric"
)

func (r *OTelRecorder) initDistributedMetrics(meter metric.Meter) error {
	var err error

	r.distHits, err = meter.Int64Counter("cache.distributed.hit",
		metric.WithDescription("Distributed storage (Redis) cache hits"),
		metric.WithUnit("{hit}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.distributed.hit counter: %w", err)
	}

	r.distMisses, err = meter.Int64Counter("cache.distributed.miss",
		metric.WithDescription("Distributed storage (Redis) cache misses"),
		metric.WithUnit("{miss}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.distributed.miss counter: %w", err)
	}

	r.distRefreshes, err = meter.Int64Counter("cache.distributed.refresh",
		metric.WithDescription("Records retrieved from distributed storage that needed refresh"),
		metric.WithUnit("{refresh}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.distributed.refresh counter: %w", err)
	}

	r.distMissingRec, err = meter.Int64Counter("cache.distributed.missing_record",
		metric.WithDescription("Missing records retrieved from distributed storage"),
		metric.WithUnit("{record}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.distributed.missing_record counter: %w", err)
	}

	r.distFallbacks, err = meter.Int64Counter("cache.distributed.fallback",
		metric.WithDescription("Fallbacks to distributed storage when refresh failed"),
		metric.WithUnit("{fallback}"),
	)
	if err != nil {
		return fmt.Errorf("creating cache.distributed.fallback counter: %w", err)
	}

	return nil
}
