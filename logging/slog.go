package logging

import (
	"errors"
	"log/slog"
	"os"
)

var logger *slog.Logger

// HandlerType determines which type of Handler should be used for the logger
type HandlerType int

const (
	JSON HandlerType = 1 + iota
	Text
)

// Initiate function provides a new logger instance from the slog package
// with the provided options.
func Initiate(handler HandlerType, level slog.Level) {
	handlerOptions := &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	}
	if handler == JSON {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, handlerOptions))
	} else if handler == Text {
		logger = slog.New(slog.NewTextHandler(os.Stdout, handlerOptions))
	}
}

// Get function returns the logger instance once initialized or returns error, when
// a logger instance could not be created.
func Get() (*slog.Logger, error) {
	if logger == nil {
		return nil, errors.New("logger is not initiated")
	}
	return logger, nil
}
