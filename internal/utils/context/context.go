package context

import (
	"context"
)

type ctxKey string

const (
	requestID  ctxKey = "requestID"
	apiVersion ctxKey = "apiVersion"
	actorKey   ctxKey = "actorID"
)

func SetRequestID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, requestID, reqID)
}

func GetRequestID(ctx context.Context) string {
	v, ok := ctx.Value(requestID).(string)
	if !ok {
		return ""
	}

	return v
}

func SetAPIVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, apiVersion, version)
}

// GetAPIVersion returns the api version from the context.
func GetAPIVersion(ctx context.Context) string {
	v, ok := ctx.Value(apiVersion).(string)
	if !ok {
		return ""
	}

	return v
}

// SetActorID stores the authenticated actor's user ID in the context.
func SetActorID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, actorKey, id)
}

// GetActorID returns the authenticated actor's user ID from the context.
func GetActorID(ctx context.Context) string {
	v, ok := ctx.Value(actorKey).(string)
	if !ok {
		return ""
	}

	return v
}
