package logger

import (
	"fmt"
	"log/slog"
	"strings"
)

type Config struct {
	Level       string `mapstructure:"level"`        // log level: "debug", "info", "warn", or "error"
	Format      string `mapstructure:"format"`       // output format: "json" or "text"
	Output      string `mapstructure:"output"`       // output destination: "stdout", "stderr", or a file path
	AddSource   bool   `mapstructure:"add_source"`   // whether to include source file and line number in logs
	ServiceName string `mapstructure:"service_name"` // identifier for the service (e.g., "task-manager-api")
	Environment string `mapstructure:"environment"`  // deployment environment: "development", "production", "staging"
}

func parseLevel(levelStr string) (level slog.Level) {
	levelStrToLow := strings.ToLower(levelStr)

	if err := level.UnmarshalText([]byte(levelStrToLow)); err != nil {
		fmt.Printf("Incorrect level '%s', using INFO\n", levelStr)
		return slog.LevelInfo
	}
	return level
}
