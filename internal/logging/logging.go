package logging

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Level string

	// Format: txext | json
	Format string

	Silent bool

	Debug bool

	LogFile string
}

type cleanupFn func()

func Init(cfg Config) cleanupFn {
	level := parseLevel(cfg.Level)
	if cfg.Debug {
		level = slog.LevelDebug
	}

	var (
		w      io.Writer
		closer io.Closer
		flush  func() error
	)

	if cfg.Silent {
		w = io.Discard
	} else {
		w = os.Stderr
	}

	if cfg.LogFile != "" && !cfg.Silent {
		_ = os.MkdirAll(filepath.Dir(cfg.LogFile), 0o755)

		file, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open log file failed: %v (fallback to stderr)\n", err)
		} else {
			closer = file
			bw := bufio.NewWriterSize(file, 10*1024)
			flush = bw.Flush

			w = io.MultiWriter(os.Stderr, bw)
		}
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05"))
				return a
			}

			if a.Key == slog.LevelKey {
				lv := a.Value.Any().(slog.Level)
				a.Value = slog.StringValue(strings.ToUpper(lv.String()))
				return a
			}

			if a.Key == slog.SourceKey {
				if src, ok := a.Value.Any().(*slog.Source); ok && src != nil {
					src.File = filepath.Base(src.File)
					a.Value = slog.AnyValue(src)
				}
				return a
			}

			return a
		},
	}

	var h slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		h = slog.NewJSONHandler(w, opts)
	default:
		h = slog.NewTextHandler(w, opts)
	}

	logger := slog.New(h)

	slog.SetDefault(logger)

	return func() {
		if flush != nil {
			_ = flush()
		}
		if closer != nil {
			_ = closer.Close()
		}
	}
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
