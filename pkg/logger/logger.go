// Package logger provides support for initializing the log system.
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

// TraceIDFn represents a function that can return the trace id from the specified context.
type TraceIDFn func(ctx context.Context) string

// RequestIDFn represents a function that can return the request id from the specified context.
type RequestIDFn func(ctx context.Context) string

type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Debugc(ctx context.Context, caller int, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Infoc(ctx context.Context, caller int, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Warnc(ctx context.Context, caller int, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Errorc(ctx context.Context, caller int, msg string, args ...any)
}

// Service represents a logger for logging information.
type Service struct {
	discard   bool
	handler   slog.Handler
	traceID   string
	requestID string
}

// New constructs a new log for application use.
func New(
	writerIO io.Writer,
	minLevel Level,
	serviceName string,
	traceIDFn string,
	requestIDFn string,
) *Service {
	return new(writerIO, minLevel, serviceName, traceIDFn, requestIDFn, Events{})
}

// NewWithEvents constructs a new log for application use with events.
func NewWithEvents(
	w io.Writer,
	minLevel Level,
	serviceName string,
	traceIDFn string,
	requestIDFn string,
	events Events,
) *Service {
	return new(w, minLevel, serviceName, traceIDFn, requestIDFn, events)
}

// NewWithHandler returns a new log for application use with the underlying
// handler.
func NewWithHandler(h slog.Handler) *Service {
	return &Service{handler: h}
}

// NewStdLogger returns a standard library Service that wraps the slog Service.
func NewStdLogger(logger *Service, level Level) *log.Logger {
	return slog.NewLogLogger(logger.handler, slog.Level(level))
}

// Debug logs at LevelDebug with the given context.
func (log *Service) Debug(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelDebug, 3, msg, args...)
}

// Debugc logs the information at the specified call stack position.
func (log *Service) Debugc(ctx context.Context, caller int, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelDebug, caller, msg, args...)
}

// Info logs at LevelInfo with the given context.
func (log *Service) Info(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelInfo, 3, msg, args...)
}

// Infoc logs the information at the specified call stack position.
func (log *Service) Infoc(ctx context.Context, caller int, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelInfo, caller, msg, args...)
}

// Warn logs at LevelWarn with the given context.
func (log *Service) Warn(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelWarn, 3, msg, args...)
}

// Warnc logs the information at the specified call stack position.
func (log *Service) Warnc(ctx context.Context, caller int, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelWarn, caller, msg, args...)
}

// Error logs at LevelError with the given context.
func (log *Service) Error(ctx context.Context, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelError, 3, msg, args...)
}

// Errorc logs the information at the specified call stack position.
func (log *Service) Errorc(ctx context.Context, caller int, msg string, args ...any) {
	if log.discard {
		return
	}

	log.write(ctx, LevelError, caller, msg, args...)
}

func (log *Service) write(
	ctx context.Context,
	level Level,
	caller int,
	msg string,
	args ...any,
) {
	slogLevel := slog.Level(level)

	if !log.handler.Enabled(ctx, slogLevel) {
		return
	}

	var pcs [1]uintptr

	runtime.Callers(caller, pcs[:])

	slogRec := slog.NewRecord(time.Now(), slogLevel, msg, pcs[0])

	if log.traceID != "" {
		args = append(args, "trace_id", log.traceID)
	}

	if log.requestID != "" {
		args = append(args, "request_id", log.requestID)
	}

	slogRec.Add(args...)

	log.handler.Handle(ctx, slogRec)
}

func new(
	w io.Writer,
	minLevel Level,
	serviceName string,
	traceID string,
	requestID string,
	events Events,
) *Service {
	// Convert the file name to just the name.ext when this key/value is logged.
	f := func(_ []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)

				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}

	// Construct the slog JSON handler for use.
	handler := slog.Handler(
		slog.NewJSONHandler(
			w,
			&slog.HandlerOptions{AddSource: true, Level: slog.Level(minLevel), ReplaceAttr: f},
		),
	)

	// If events are to be processed, wrap the JSON handler around the custom
	// log handler.
	if events.Debug != nil || events.Info != nil || events.Warn != nil || events.Error != nil {
		handler = newLogHandler(handler, events)
	}

	// Attributes to add to every log.
	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}

	// Add those attributes and capture the final handler.
	handler = handler.WithAttrs(attrs)

	return &Service{
		discard:   w == io.Discard,
		handler:   handler,
		traceID:   traceID,
		requestID: requestID,
	}
}
