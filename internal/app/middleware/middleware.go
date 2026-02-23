// Package middleware package provides level middleware support.
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Housiadas/cerberus/internal/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_roles_permissions_usecase"
	"github.com/Housiadas/cerberus/internal/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"go.opentelemetry.io/otel/trace"
)

const (
	contentTypeKey  = "Content-Type"
	contentTypeJSON = "application/json"
)

type Config struct {
	Log                  logger.Logger
	Tracer               trace.Tracer
	Tx                   *pgsql.DBBeginner
	AuthUseCase          *auth_usecase.UseCase
	UserUseCase          *user_usecase.UseCase
	UserRolesPermissions *user_roles_permissions_usecase.UseCase
}

type Middleware struct {
	Log     logger.Logger
	Tracer  trace.Tracer
	Tx      *pgsql.DBBeginner
	UseCase UseCase
}

type UseCase struct {
	Auth                 *auth_usecase.UseCase
	User                 *user_usecase.UseCase
	UserRolesPermissions *user_roles_permissions_usecase.UseCase
}

func New(cfg Config) *Middleware {
	return &Middleware{
		UseCase: UseCase{
			Auth:                 cfg.AuthUseCase,
			User:                 cfg.UserUseCase,
			UserRolesPermissions: cfg.UserRolesPermissions,
		},
		Log:    cfg.Log,
		Tracer: cfg.Tracer,
		Tx:     cfg.Tx,
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

// ResponseRecorder a custom http.ResponseWriter to capture the response before it's sent to the client.
// We are capturing the result of the handler to the middleware.
type ResponseRecorder struct {
	http.ResponseWriter

	statusCode int
	body       bytes.Buffer
}

func (rec *ResponseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// Write Capture the response body.
func (rec *ResponseRecorder) Write(b []byte) (int, error) {
	rec.body.Write(b)

	write, err := rec.ResponseWriter.Write(b)
	if err != nil {
		return 0, fmt.Errorf("write response error: %w", err)
	}

	return write, nil
}
