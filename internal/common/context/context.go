package context

import (
	"context"
)

type ctxKey string

const (
	requestID  ctxKey = "requestID"
	apiVersion ctxKey = "apiVersion"
	userIDKey  ctxKey = "userIDKey"
	userKey    ctxKey = "userKey"
	roleIDKey  ctxKey = "roleIDKey"
	roleKey    ctxKey = "roleKey"
)

func SetRequestID(ctx context.Context, reqId string) context.Context {
	return context.WithValue(ctx, requestID, reqId)
}

func GetRequestID(ctx context.Context) string {
	v, ok := ctx.Value(requestID).(string)
	if !ok {
		return ""
	}
	return v
}

func SetApiVersion(ctx context.Context, version string) context.Context {
	return context.WithValue(ctx, apiVersion, version)
}

func GetApiVersion(ctx context.Context) string {
	v, ok := ctx.Value(apiVersion).(string)
	if !ok {
		return "v1"
	}
	return v
}
