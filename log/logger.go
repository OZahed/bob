package log

import (
	"fmt"
	"log/slog"
	"runtime"
)

const (
	maxDepthOfLogger = 25
	runtimeMain      = "runtime.main"
)

// Logger is a wrapper around the slog logger from the slog package.
type Logger struct {
	*slog.Logger
	stackSkip int
}

type loggerOption struct {
	stackSkip int
}

type loggerOptionFunc func(*loggerOption)

func WithSkipStack(skip int) loggerOptionFunc {
	return func(opt *loggerOption) {
		opt.stackSkip = skip + 1
	}
}

// NewLogger returns a new instance of the Logger.
// Example
//
//		slg := log.NewSlog(log.WithHandlerType(log.JsonHandler), log.WithLevel("debug"), log.WithAlwaysUTC(true))
//	 	lgr := log.NewLogger(slg, WithSkipStack(1))
//		lgr.ErrorWithStack("error message", "key", "value")
func NewLogger(slg *slog.Logger, opts ...loggerOptionFunc) *Logger {
	option := &loggerOption{
		stackSkip: 1,
	}

	for _, opt := range opts {
		opt(option)
	}

	return &Logger{
		Logger:    slg,
		stackSkip: option.stackSkip,
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

		stackFrameInfo = fmt.Sprintf("%s%s\n\t%s:%d\n", stackFrameInfo, file, funcName, line)
	}

	return stackFrameInfo
}

// ErrorWithStack logs error with the called stack frames during the call to the function.
func (l Logger) ErrorWithStack(msg string, args ...any) {
	stacks := getStackFrame(l.stackSkip)
	args = append(args, "stack", stacks)
	l.Error(msg, args...)
}

// DebugWithStack logs error with the called stack frames during the call to the function.
func (l Logger) DebugWithStack(msg string, args ...any) {
	stacks := getStackFrame(l.stackSkip)
	args = append(args, "stack", stacks)
	l.Debug(msg, args...)
}

func (l Logger) WarnWithStack(msg string, args ...any) {
	stacks := getStackFrame(l.stackSkip)
	args = append(args, "stack", stacks)
	l.Warn(msg, args...)
}
