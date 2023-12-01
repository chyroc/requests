package requests

import (
	"context"
	"log"
	"os"
)

type Logger interface {
	Info(ctx context.Context, format string, v ...interface{})
	Error(ctx context.Context, format string, v ...interface{})
}

func StdoutLogger() Logger {
	return newStdoutLogger()
}

func DiscardLogger() Logger {
	return newDiscardLogger()
}

// stdout logger
type stdoutLogger struct {
	logger *log.Logger
}

func (r *stdoutLogger) Info(ctx context.Context, format string, v ...interface{}) {
	r.logger.Printf(format, v...)
}

func (r *stdoutLogger) Error(ctx context.Context, format string, v ...interface{}) {
	r.logger.Printf(format, v...)
}

func newStdoutLogger() Logger {
	return &stdoutLogger{
		logger: log.New(os.Stdout, "[requests] ", log.LstdFlags),
	}
}

// discard logger
type discardLogger struct{}

func (r *discardLogger) Info(ctx context.Context, format string, v ...interface{}) {
}

func (r *discardLogger) Error(ctx context.Context, format string, v ...interface{}) {
}

func newDiscardLogger() Logger {
	return &discardLogger{}
}
