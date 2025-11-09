package logger

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
)

// Config holds logger configuration including level, format, output destination,
// service name, and environment for structured logging.
type Config struct {
	Level       string `mapstructure:"level"`        // log level: "debug", "info", "warn", or "error"
	Format      string `mapstructure:"format"`       // output format: "json" or "text"
	Output      string `mapstructure:"output"`       // output destination: "stdout", "stderr", or a file path
	AddSource   bool   `mapstructure:"add_source"`   // whether to include source file and line number in logs
	ServiceName string `mapstructure:"service_name"` // identifier for the service (e.g., "task-manager-api")
	Environment string `mapstructure:"environment"`  // deployment environment: "development", "production", "staging"
}

func (cfg *Config) Validate() error {
	var errs []error
	validLevels := []string{"debug", "info", "warn", "error"}
	if !slices.Contains(validLevels, strings.ToLower(cfg.Level)) {
		errs = append(errs, fmt.Errorf("invalid level '%s', should be 'debug', 'info', 'warn', 'error'", cfg.Level))
	}

	format := strings.ToLower(cfg.Format)
	if format != "json" && format != "text" {
		errs = append(errs, fmt.Errorf("invalid format: %s, should be 'json' or 'text'", format))
	}

	if len(cfg.Output) == 0 {
		errs = append(errs, fmt.Errorf("output required"))
	}

	if len(cfg.ServiceName) == 0 {
		errs = append(errs, fmt.Errorf("service name required"))
	}

	return errors.Join(errs...)
}

// parseLevel converts a string log level to slog.Level.
// Returns INFO level for invalid input.
func parseLevel(levelStr string) (level slog.Level) {
	levelStrToLow := strings.ToLower(levelStr)

	if err := level.UnmarshalText([]byte(levelStrToLow)); err != nil {
		fmt.Printf("Incorrect level '%s', using INFO\n", levelStr)
		return slog.LevelInfo
	}
	return level
}
