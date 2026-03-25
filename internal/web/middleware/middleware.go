// Package middleware package provides level middleware support.
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_roles_permissions_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"
)

const (
	contentTypeKey  = "Content-Type"
	contentTypeJSON = "application/json"
)

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

type Config struct {
	Log                  logger
	Tracer               trace.Tracer
	Meter                metric.Meter
	AuthUseCase          *auth_usecase.UseCase
	UserUseCase          *user_usecase.UseCase
	UserRolesPermissions *user_roles_permissions_usecase.UseCase
}

type Middleware struct {
	log         logger
	tracer      trace.Tracer
	meter       metric.Meter
	useCase     useCase
	permSflight singleflight.Group
	metrics     middlewareMetrics
}

type useCase struct {
	auth                 *auth_usecase.UseCase
	user                 *user_usecase.UseCase
	userRolesPermissions *user_roles_permissions_usecase.UseCase
}

func New(ctx context.Context, cfg Config) *Middleware {
	middlewareMetrics := initOTelMetrics(ctx, cfg.Meter, cfg.Log)

	return &Middleware{
		useCase: useCase{
			auth:                 cfg.AuthUseCase,
			user:                 cfg.UserUseCase,
			userRolesPermissions: cfg.UserRolesPermissions,
		},
		log:     cfg.Log,
		tracer:  cfg.Tracer,
		meter:   cfg.Meter,
		metrics: middlewareMetrics,
	}
}

func (m *Middleware) Error(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set(contentTypeKey, contentTypeJSON)
	w.WriteHeader(statusCode)

	err = json.NewEncoder(w).Encode(err)
	if err != nil {
		return
	}
}

// responseRecorder a custom http.ResponseWriter to capture
// the response before it's sent to the client.
// We are capturing the result of the handler to the middleware.
type responseRecorder struct {
	http.ResponseWriter

	statusCode int
	body       bytes.Buffer
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// Write Capture the response body.
func (rec *responseRecorder) Write(b []byte) (int, error) {
	rec.body.Write(b)

	write, err := rec.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("write response error: %w", err)
	}

	return write, nil
}
