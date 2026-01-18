package metrics

import (
	"context"
	"runtime"
)

const CtxKey = "metricsCtx"

// Set sets the metrics data into the context.
func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, CtxKey, &m)
}

// AddGoroutines refreshes the goroutine metric.
func AddGoroutines(ctx context.Context) int64 {
	if v, ok := ctx.Value(CtxKey).(*metrics); ok {
		g := int64(runtime.NumGoroutine())
		v.goroutines.Set(g)

		return g
	}

	return 0
}

// AddRequests increments the request metric by 1.
func AddRequests(ctx context.Context) int64 {
	v, ok := ctx.Value(CtxKey).(*metrics)
	if ok {
		v.requests.Add(1)

		return v.requests.Value()
	}

	return 0
}

// AddErrors increments the errors metric by 1.
func AddErrors(ctx context.Context) int64 {
	if v, ok := ctx.Value(CtxKey).(*metrics); ok {
		v.errors.Add(1)

		return v.errors.Value()
	}

	return 0
}

// AddPanics increments the panics metric by 1.
func AddPanics(ctx context.Context) int64 {
	if v, ok := ctx.Value(CtxKey).(*metrics); ok {
		v.panics.Add(1)

		return v.panics.Value()
	}

	return 0
}
