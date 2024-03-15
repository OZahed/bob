package logging

import (
	"fmt"
	"log/slog"
	"runtime"
)

const maxDepthOfLogger = 25

// Logger is a wrapper around the slog logger from the slog package.
type Logger struct {
	*slog.Logger
	stackSkip int
}

func NewLogger(logger *slog.Logger, stackSkip int) *Logger {
	return &Logger{logger, stackSkip}
}

func traverseStackFrames(depth int) (stackFrameInfo string) {
STACK_FRAME:

	if depth >= maxDepthOfLogger {
		return stackFrameInfo
	}

	pc, file, line, ok := runtime.Caller(depth)

	if !ok {
		return stackFrameInfo
	}

	funcInfo := runtime.FuncForPC(pc)
	funcName := funcInfo.Name()

	if funcName == "runtime.main" {
		return stackFrameInfo
	}

	stackFrameInfo = fmt.Sprintf("%s%s\n\t%s:%d\n", stackFrameInfo, file, funcName, line)

	depth++
	goto STACK_FRAME
}

// ErrorWithStack logs error with the called stack frames during the call to the function.
func (l Logger) ErrorWithStack(msg string, args ...any) {
	stacks := traverseStackFrames(l.stackSkip)
	args = append(args, "stack", stacks)
	l.Error(msg, args...)
}

// DebugWithStack logs error with the called stack frames during the call to the function.
func (l Logger) DebugWithStack(msg string, args ...any) {
	stacks := traverseStackFrames(l.stackSkip)
	args = append(args, "stack", stacks)
	l.Debug(msg, args...)
}
