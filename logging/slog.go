package logging

import (
	"log/slog"
	"os"
	"time"
)

// HandlerType determines which type of Handler should be used for the logger
type HandlerType int

const (
	JSON HandlerType = 1 + iota
	Text
)

// NewSlog function provides a new logger instance from the slog package
// with the provided options.
func NewSlog(handler HandlerType, level slog.Level, name string) *slog.Logger {
	var handlerFunc slog.Handler

	handlerOptions := &slog.HandlerOptions{
		AddSource:   true,
		Level:       level,
		ReplaceAttr: WithUTCTime,
	}
	switch handler {
	case JSON:
		handlerFunc = slog.NewJSONHandler(os.Stdout, handlerOptions)
	default:
		handlerFunc = slog.NewTextHandler(os.Stdout, handlerOptions)
	}

	if name != "" {
		handlerFunc = handlerFunc.WithAttrs([]slog.Attr{slog.String("name", name)})
	}

	return slog.New(handlerFunc)
}

// WithUTCTime replaces the time attr value from Unix timestamps to UTC date
// and duration to a string representation of it.
func WithUTCTime(groups []string, a slog.Attr) slog.Attr {
	switch a.Key {
	case slog.TimeKey:
		a.Value = slog.TimeValue(time.Now().UTC())
	case slog.KindDuration.String():
		a.Value = slog.StringValue(a.Value.Duration().String())
	}

	return a
}
