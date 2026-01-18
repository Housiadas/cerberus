package logger

import (
	"context"
	"log/slog"
	"time"
)

// Level represents different logging levels.
type Level slog.Level

// A set of possible logging levels.
const (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo  = Level(slog.LevelInfo)
	LevelWarn  = Level(slog.LevelWarn)
	LevelError = Level(slog.LevelError)
)

// Record represents the data that is being logged.
type Record struct {
	Time       time.Time
	Message    string
	Level      Level
	Attributes map[string]any
}

func toRecord(record slog.Record) Record {
	atts := make(map[string]any, record.NumAttrs())

	f := func(attr slog.Attr) bool {
		atts[attr.Key] = attr.Value.Any()

		return true
	}
	record.Attrs(f)

	return Record{
		Time:       record.Time,
		Message:    record.Message,
		Level:      Level(record.Level),
		Attributes: atts,
	}
}

// EventFn is a function to be executed when configured against a log level.
type EventFn func(ctx context.Context, r Record)

// Events contain an assignment of an event function to a log level.
type Events struct {
	Debug EventFn
	Info  EventFn
	Warn  EventFn
	Error EventFn
}
