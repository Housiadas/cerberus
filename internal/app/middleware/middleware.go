// Package middleware provides cli level middleware support.
package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"

	"github.com/Housiadas/cerberus/internal/core/service/product_core"
	"github.com/Housiadas/cerberus/internal/core/service/user_core"
	"github.com/Housiadas/cerberus/pkg/logger"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

var (
	// ErrInvalidID represents a condition where the id is not an uuid.
	ErrInvalidID = errors.New("ID is not in its proper form")

	group = singleflight.Group{}
)

type Config struct {
	Log     *logger.Logger
	Tracer  trace.Tracer
	Tx      *pgsql.DBBeginner
	User    *user_core.Core
	Product *product_core.Core
}

type Middleware struct {
	Core   Core
	Log    *logger.Logger
	Tracer trace.Tracer
	Tx     *pgsql.DBBeginner
}

type Core struct {
	User    *user_core.Core
	Product *product_core.Core
}

func New(cfg Config) *Middleware {
	return &Middleware{
		Core: Core{
			User:    cfg.User,
			Product: cfg.Product,
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
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
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
