// Package logging configures structured logging useing log/slog (Go 1.21+).
package logging

import (
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

type contextKey string

const loggerKey = contextKey("logger")

var (
	defaultLogger     *slog.Logger
	defaultLoggerOnce sync.Once
)

func NewLogger(level, mode string) *slog.Logger {
	minLevel := levelToSlogLevel(level)
	opts := &slog.HandlerOptions{
		Level:     minLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Key = "timestamp"
					a.Value = slog.StringValue(t.UTC().Format("2006-01-02T15:04:05.000Z"))
					return a
				}
				a.Key = "timestamp"

			case slog.LevelKey:
				a.Key = "level"
				a.Value = slog.StringValue(strings.ToLower(a.Value.String()))

			case slog.MessageKey:
				a.Key = "message"

			case slog.SourceKey:
				a.Key = "caller"
			}
			return a
		},
	}
	var h slog.Handler
	if mode == "dev" {
		h = slog.NewTextHandler(os.Stderr, opts)
	} else {
		h = slog.NewJSONHandler(os.Stderr, opts)
	}

	return slog.New(h)
}

func levelToSlogLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	default:
		return slog.LevelInfo
	}
}
