package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// getWriter returns an io.Writer based on the output destination string.
// Supports "stdout", "stderr", or a file path with automatic directory creation.
func getWriter(cfg *Config) (io.Writer, error) {
	if len(cfg.Output) == 0 {
		return nil, fmt.Errorf("output destination cannot be empty")
	}

	outputToLower := strings.ToLower(cfg.Output)

	if outputToLower == "stdout" {
		return os.Stdout, nil
	}

	if outputToLower == "stderr" {
		return os.Stderr, nil
	}

	dir := filepath.Dir(cfg.Output)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}

	if cfg.EnableRotation == true {
		lumber := &lumberjack.Logger{
			Filename:   cfg.Output,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   true,
			LocalTime:  true,
		}
		return lumber, nil
	}

	file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file %s: %w", cfg.Output, err)
	}

	return file, nil
}

// createHandler creates and configures a slog.Handler based on the format specified in cfg.
// Supports "json" and "text" formats. Defaults to JSON for invalid formats.
func createHandler(cfg *Config, writer io.Writer) slog.Handler {
	opts := slog.HandlerOptions{
		Level:     parseLevel(cfg.Level),
		AddSource: cfg.AddSource,
	}

	format := strings.ToLower(cfg.Format)

	if format == "json" {
		return slog.NewJSONHandler(writer, &opts)
	}

	if format == "text" {
		return slog.NewTextHandler(writer, &opts)
	}

	fmt.Printf("invalid format: %s, should be 'json' or 'text'\n", format)
	return slog.NewJSONHandler(writer, &opts)
}

// NewLogger creates a new configured slog.Logger instance based on the provided Config.
// It sets up the output destination, format handler, and adds default attributes
// (service name and environment) that appear in all log entries.
// Returns an error if the configuration is invalid or output destination cannot be created.
func NewLogger(cfg *Config) (*slog.Logger, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	writer, err := getWriter(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get writer: %w", err)
	}

	handler := createHandler(cfg, writer)

	logger := slog.New(handler).With(
		slog.String("service", cfg.ServiceName),
		slog.String("environment", cfg.Environment),
	)

	return logger, nil
}

// NewDefault creates a logger with sensible defaults for development.
// Uses text format, info level, and stdout output.
func NewDefault() *slog.Logger {
	logger, _ := NewLogger(&Config{
		Level:       "info",
		Format:      "text",
		Output:      "stderr",
		AddSource:   false,
		ServiceName: "app",
		Environment: "development",
	})
	return logger
}
