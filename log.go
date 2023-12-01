package requests

import (
	"context"
	"log"
	"os"
)

type Logger interface {
	Info(ctx context.Context, format string, v ...any)
	Error(ctx context.Context, format string, v ...any)
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

func (r *stdoutLogger) Info(ctx context.Context, format string, v ...any) {
	r.logger.Printf(format, v...)
}

func (r *stdoutLogger) Error(ctx context.Context, format string, v ...any) {
	r.logger.Printf(format, v...)
}

func newStdoutLogger() Logger {
	return &stdoutLogger{
		logger: log.New(os.Stdout, "[requests] ", log.LstdFlags),
	}
}

// discard logger
type discardLogger struct{}

func (r *discardLogger) Info(ctx context.Context, format string, v ...any) {
}

func (r *discardLogger) Error(ctx context.Context, format string, v ...any) {
}

func newDiscardLogger() Logger {
	return &discardLogger{}
}
