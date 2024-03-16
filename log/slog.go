package log

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

// HandlerType determines which type of Handler should be used for the logger
type HandlerType int

const (
	JsonHandler HandlerType = 1 + iota
	TextHandler
)

const (
	maxDepthOfLogger         = 25
	runtimeMain              = "runtime.main"
	replaceAttrFunctionStack = 7
)

// slogOptions is a configuration struct for the ReplaceAttr function
type slogOptions struct {
	// HandlerType is the type of handler to be used for the logger
	HandlerType     HandlerType
	ReplaceAttrFunc func(groups []string, a slog.Attr) slog.Attr
	// Level is the level of logging to be used for the logger
	// Possible levels are "info", "warn", "warning", "error", "err", "debug"
	Level string

	// SkipStack is the number of stack frames to skip when logging 1 is the default
	SkipStack int
	// AddStack is a flag to determine if the stack should be added to the log
	AddStack bool

	// ReplaceAttrEnable is a flag to determine if the ReplaceAttr function should be enabled
	ReplaceAttrEnable bool
	// AlwaysUTC is a flag to determine if the time should always be in UTC regardless of your system timezone
	AlwaysUTC bool
}

type slogOptionFunc func(*slogOptions)

func WithLevel(level string) slogOptionFunc {
	return func(cfg *slogOptions) {
		cfg.Level = level
	}
}

func WithHandlerType(handlerType HandlerType) slogOptionFunc {
	return func(cfg *slogOptions) {
		cfg.HandlerType = handlerType
	}
}

func WithAlwaysUTC(enable bool) slogOptionFunc {
	return func(cfg *slogOptions) {
		cfg.ReplaceAttrEnable = true
		cfg.AlwaysUTC = enable
	}
}

func WithReplceAttrFunc(replaceAttr func(groups []string, a slog.Attr) slog.Attr) slogOptionFunc {
	return func(cfg *slogOptions) {
		cfg.ReplaceAttrEnable = true
		cfg.ReplaceAttrFunc = replaceAttr

	}
}

func WithStackFrame() slogOptionFunc {
	return func(cfg *slogOptions) {
		cfg.ReplaceAttrEnable = true
		cfg.SkipStack = replaceAttrFunctionStack
		cfg.AddStack = true
	}
}

// NewSlog function provides a new logger instance from the slog package
// with the provided options.
func NewSlog(opts ...slogOptionFunc) *slog.Logger {
	// Default Options
	opt := slogOptions{
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

func makeReplaceAttr(cfg slogOptions) func(groups []string, a slog.Attr) slog.Attr {
	if !cfg.ReplaceAttrEnable {
		return nil
	}

	if cfg.ReplaceAttrFunc != nil {
		return cfg.ReplaceAttrFunc
	}

	return func(groups []string, a slog.Attr) slog.Attr {
		switch {
		case a.Key == slog.TimeKey && cfg.AlwaysUTC:
			a.Value = slog.TimeValue(a.Value.Time().UTC())
		case cfg.HandlerType == JsonHandler && a.Value.Kind() == slog.KindDuration:
			a.Value = slog.StringValue(a.Value.Duration().String())
		case cfg.AddStack && a.Key == slog.SourceKey:
			src := a.Value.Any().(*slog.Source)
			stack := getStackFrame(cfg.SkipStack)
			return slog.Group(slog.SourceKey,
				"caller", fmt.Sprintf("%s:%d\n\t%s:%d", src.File, src.Line, src.Function, src.Line),
				"callerStack", stack)
		}

		return a
	}
}

func getStackFrame(depth int) (stackFrameInfo string) {
	for i := depth; i < maxDepthOfLogger; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		funcInfo := runtime.FuncForPC(pc)
		funcName := funcInfo.Name()

		if funcName == runtimeMain {
			break
		}

		stackFrameInfo = fmt.Sprintf("%s%s:%d\n\t%s:%d\n", stackFrameInfo, file, line, funcName, line)
	}

	return stackFrameInfo
}
