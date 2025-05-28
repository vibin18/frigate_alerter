package logger

import (
	"io"
	"log/slog"
	"os"
)

var (
	// DefaultLogger is the default logger instance
	DefaultLogger *slog.Logger
)

// Config holds the logger configuration
type Config struct {
	// Level is the minimum log level to log
	Level slog.Level
	// Output is where logs are written
	Output io.Writer
	// AddSource adds source code location to log entries
	AddSource bool
}

// init initializes the default logger
func init() {
	// Create a default logger that logs to stdout with default settings
	DefaultLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))

	// Set as default logger
	slog.SetDefault(DefaultLogger)
}

// Configure configures the default logger with the provided options
func Configure(cfg Config) {
	// Create a handler with the provided options
	handler := slog.NewJSONHandler(cfg.Output, &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	})

	// Create a new logger with the handler
	DefaultLogger = slog.New(handler)

	// Set as default logger
	slog.SetDefault(DefaultLogger)
}
