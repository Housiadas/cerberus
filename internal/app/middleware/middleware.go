// Package middleware provides level middleware support.
package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"

	"github.com/Housiadas/cerberus/internal/app/usecase/auth_usecase"
	"github.com/Housiadas/cerberus/internal/app/usecase/user_usecase"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
	"github.com/Housiadas/cerberus/pkg/web"
)

var (
	// ErrInvalidID represents a condition where the id is not an uuid.
	ErrInvalidID = errors.New("ID is not in its proper form")

	group = singleflight.Group{}
)

type Config struct {
	Log         *logger.Logger
	Tracer      trace.Tracer
	Tx          *pgsql.DBBeginner
	AuthUseCase *auth_usecase.UseCase
	UserUseCase *user_usecase.UseCase
}

type Middleware struct {
	Tracer  trace.Tracer
	Log     *logger.Logger
	Tx      *pgsql.DBBeginner
	UseCase UseCase
}

type UseCase struct {
	Auth *auth_usecase.UseCase
	User *user_usecase.UseCase
}

func New(cfg Config) *Middleware {
	return &Middleware{
		UseCase: UseCase{
			Auth: cfg.AuthUseCase,
			User: cfg.UserUseCase,
		},
		Log:    cfg.Log,
		Tracer: cfg.Tracer,
		Tx:     cfg.Tx,
	}
}

func (m *Middleware) Error(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(err); err != nil {
		return
	}
	return
}

// ResponseRecorder a custom http.ResponseWriter to capture the response before it's sent to the client.
// We are capturing the result of the handler to the middleware
type ResponseRecorder struct {
	statusCode int
	body       bytes.Buffer
	http.ResponseWriter
}

func (rec *ResponseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// Capture the response body
func (rec *ResponseRecorder) Write(b []byte) (int, error) {
	rec.body.Write(b)
	return rec.ResponseWriter.Write(b)
}

func checkIsError(e web.Encoder) error {
	err, hasError := e.(error)
	if hasError {
		return err
	}

	return nil
}
