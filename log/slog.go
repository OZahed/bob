package log

import (
	"log/slog"
	"os"
	"strings"
)

// HandlerType determines which type of Handler should be used for the logger
type HandlerType int

const (
	JsonHandler HandlerType = 1 + iota
	TextHandler
)

// loggerOption is a configuration struct for the ReplaceAttr function
type loggerOption struct {
	// HandlerType is the type of handler to be used for the logger
	HandlerType     HandlerType
	ReplaceAttrFunc func(groups []string, a slog.Attr) slog.Attr
	// Level is the level of logging to be used for the logger
	// Possible levels are "info", "warn", "warning", "error", "err", "debug"
	Level string

	// ReplaceAttrEnable is a flag to determine if the ReplaceAttr function should be enabled
	ReplaceAttrEnable bool
	// AlwaysUTC is a flag to determine if the time should always be in UTC regardless of your system timezone
	AlwaysUTC bool
}

type LoggerOpt func(*loggerOption)

func WithLevel(level string) LoggerOpt {
	return func(cfg *loggerOption) {
		cfg.Level = level
	}
}

func WithHandlerType(handlerType HandlerType) LoggerOpt {
	return func(cfg *loggerOption) {
		cfg.HandlerType = handlerType
	}
}

func WithAlwaysUTC(enable bool) LoggerOpt {
	return func(cfg *loggerOption) {
		cfg.ReplaceAttrEnable = true
		cfg.AlwaysUTC = enable
	}
}

func WithReplceAttrFunc(replaceAttr func(groups []string, a slog.Attr) slog.Attr) LoggerOpt {
	return func(cfg *loggerOption) {
		cfg.ReplaceAttrEnable = true
		cfg.ReplaceAttrFunc = replaceAttr

	}
}

// NewSlog function provides a new logger instance from the slog package
// with the provided options.
func NewSlog(opts ...LoggerOpt) *slog.Logger {
	// Default Options
	opt := loggerOption{
		HandlerType:       TextHandler,
		Level:             "debug",
		ReplaceAttrEnable: false,
	}

	for _, o := range opts {
		o(&opt)
	}

	var handlerFunc slog.Handler
	handlerOptions := &slog.HandlerOptions{
		AddSource:   true,
		Level:       getLoggerLevel(opt.Level),
		ReplaceAttr: makeReplaceAttr(opt),
	}
	switch opt.HandlerType {
	case JsonHandler:
		handlerFunc = slog.NewJSONHandler(os.Stdout, handlerOptions)
	default:
		handlerFunc = slog.NewTextHandler(os.Stdout, handlerOptions)
	}

	return slog.New(handlerFunc)
}

func getLoggerLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "info", "information":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}

func makeReplaceAttr(cfg loggerOption) func(groups []string, a slog.Attr) slog.Attr {
	if !cfg.ReplaceAttrEnable {
		return nil
	}

	if cfg.ReplaceAttrFunc != nil {
		return cfg.ReplaceAttrFunc
	}

	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey && cfg.AlwaysUTC {
			a.Value = slog.TimeValue(a.Value.Time().UTC())
		}

		// On slog Text handler duration is already in text format
		if cfg.HandlerType == JsonHandler && a.Value.Kind() == slog.KindDuration {
			a.Value = slog.StringValue(a.Value.Duration().String())
		}

		return a
	}
}
