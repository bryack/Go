// Package config provides configuration management for the task manager server.
// It supports loading configuration from files, environment variables, and command-line flags.
package config

import (
	"errors"
	"fmt"
	"myproject/logger"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// MinJWTSecretLength is the minimum required length for JWT secret keys.
const MinJWTSecretLength = 32

// Config holds all application configuration settings.
type Config struct {
	ServerConfig   ServerConfig   `mapstructure:"server"`
	DatabaseConfig DatabaseConfig `mapstructure:"database"`
	JWTConfig      JWTConfig      `mapstructure:"jwt"`
	LogConfig      logger.Config  `mapstructure:"logging"`
}

// ServerConfig contains HTTP server configuration.
type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	Host            string        `mapstructure:"host"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
}

// DatabaseConfig contains database connection settings.
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// JWTConfig contains JWT authentication settings.
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
}

// LoadConfig loads configuration from files, environment variables, and flags.
// Returns the parsed config, viper instance, and any error encountered.
func LoadConfig() (*Config, *viper.Viper, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.shutdown_timeout", "30s")
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "15s")
	v.SetDefault("server.idle_timeout", "2s")
	v.SetDefault("database.path", "./data/tasks.db")
	v.SetDefault("jwt.expiration", "24h")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stderr")
	v.SetDefault("logging.add_source", false)
	v.SetDefault("logging.service_name", "task-manager-api")
	v.SetDefault("logging.environment", "production")

	// Define and parse flags first (before reading config file)
	pflag.String("config", "", "Path to config file")
	pflag.Bool("show-config", false, "Display current configuration and exit")
	pflag.Int("port", 8080, "Server port")
	pflag.String("host", "0.0.0.0", "Server host")
	pflag.String("shutdown-timeout", "30s", "Graceful shutdown timeout")
	pflag.String("read-timeout", "15s", "Server ReadTimeout")
	pflag.String("write-timeout", "15s", "Server WriteTimeout")
	pflag.String("idle-timeout", "2s", "Server IdleTimeout")
	pflag.String("db-path", "./data/tasks.db", "Database path")
	pflag.String("jwt-expiration", "24h", "JWT expiration")
	pflag.String("jwt-secret", "", "JWT Secret")
	pflag.String("log-level", "info", "Log level (debug, info, warn, error)")
	pflag.String("log-format", "json", "Log format (json, text)")
	pflag.String("log-output", "stderr", "Log output (stdout, stderr, or file path)")
	pflag.Bool("log-add-source", false, "Include source file and line in logs")
	pflag.String("log-service-name", "task-manager-api", "Service name for logs")
	pflag.String("log-environment", "production", "Environment name (development, staging, production)")
	pflag.Parse()

	// Check if custom config file was specified
	configFile := pflag.Lookup("config").Value.String()
	if configFile != "" {
		// Use the specified config file
		v.SetConfigFile(configFile)
	} else {
		// Use default search paths
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/taskmanager/")
		v.AddConfigPath("$HOME/.taskmanager/")
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found is OK, continue with defaults and env vars
	}

	// Set up environment variables
	v.AutomaticEnv()
	v.SetEnvPrefix("TASKMANAGER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind flags to config keys (except --config and --show-config which are handled separately)
	v.BindPFlag("server.port", pflag.Lookup("port"))
	v.BindPFlag("server.host", pflag.Lookup("host"))
	v.BindPFlag("server.shutdown_timeout", pflag.Lookup("shutdown-timeout"))
	v.BindPFlag("server.read_timeout", pflag.Lookup("read-timeout"))
	v.BindPFlag("server.write_timeout", pflag.Lookup("write-timeout"))
	v.BindPFlag("server.idle_timeout", pflag.Lookup("idle-timeout"))
	v.BindPFlag("database.path", pflag.Lookup("db-path"))
	v.BindPFlag("jwt.expiration", pflag.Lookup("jwt-expiration"))
	v.BindPFlag("jwt.secret", pflag.Lookup("jwt-secret"))
	v.BindPFlag("logging.level", pflag.Lookup("log-level"))
	v.BindPFlag("logging.format", pflag.Lookup("log-format"))
	v.BindPFlag("logging.output", pflag.Lookup("log-output"))
	v.BindPFlag("logging.add_source", pflag.Lookup("log-add-source"))
	v.BindPFlag("logging.service_name", pflag.Lookup("log-service-name"))
	v.BindPFlag("logging.environment", pflag.Lookup("log-environment"))

	// Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, v, nil
}

// Validate checks all configuration values for correctness.
// Returns a combined error if any validation fails.
func (config *Config) Validate() error {
	var errs []error
	if config.ServerConfig.Port < 1 || config.ServerConfig.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be between 1 and 65535, got %d", config.ServerConfig.Port))
	}

	if config.ServerConfig.ShutdownTimeout <= 0 {
		errs = append(errs, fmt.Errorf("server.shutdown_timeout must be positive, got %v", config.ServerConfig.ShutdownTimeout))
	}

	if len(config.DatabaseConfig.Path) == 0 {
		errs = append(errs, fmt.Errorf("database path required"))
	}

	err := validateDatabasePath(config.DatabaseConfig.Path)
	if err != nil {
		err = fmt.Errorf("validate database path '%s' failed: %w", config.DatabaseConfig.Path, err)
		errs = append(errs, err)
	}

	if len(config.JWTConfig.Secret) == 0 {
		errs = append(errs, fmt.Errorf("jwt secret required"))
	} else if len(config.JWTConfig.Secret) < MinJWTSecretLength {
		errs = append(errs, fmt.Errorf("secret must be at least 32 symbols, got %d", len(config.JWTConfig.Secret)))
	}

	if config.JWTConfig.Expiration <= 0 {
		errs = append(errs, fmt.Errorf("expiration must be positive, got %v", config.JWTConfig.Expiration))
	}

	if err := config.LogConfig.Validate(); err != nil {
		errs = append(errs, fmt.Errorf("validate log config failed: %w", err))
	}

	return errors.Join(errs...)
}

// validateDatabasePath ensures the database directory exists and is writable.
func validateDatabasePath(path string) error {
	dir := filepath.Dir(path)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("creating directory %s failed: %w", dir, err)
			}
		}
	}

	file, err := os.CreateTemp(dir, "test*.txt")
	if err != nil {
		return fmt.Errorf("creating temp file in directory %s failed: %w", dir, err)
	}
	defer file.Close()
	defer os.Remove(file.Name())

	if _, err := file.WriteString("test data"); err != nil {
		return fmt.Errorf("writing to test file in directory %s failed: %w", dir, err)
	}
	return nil
}

// maskSensitive obscures sensitive values for display purposes.
func maskSensitive(scrt string) string {
	if len(scrt) <= 4 {
		return "****"
	}

	return scrt[0:2] + "****" + scrt[len(scrt)-2:]
}

// getSource determines where a configuration value came from (flag, env, config file, or default).
func getSource(v *viper.Viper, key string) string {
	flagMap := map[string]string{
		"server.port":             "port",
		"server.host":             "host",
		"server.shutdown_timeout": "shutdown-timeout",
		"server.read_timeout":     "read-timeout",
		"server.write_timeout":    "write-timeout",
		"server.idle_timeout":     "idle-timeout",
		"database.path":           "db-path",
		"jwt.secret":              "jwt-secret",
		"jwt.expiration":          "jwt-expiration",
		"logging.level":           "log-level",
		"logging.format":          "log-format",
		"logging.output":          "log-output",
		"logging.add_source":      "log-add-source",
		"logging.service_name":    "log-service-name",
		"logging.environment":     "log-environment",
	}

	if flagName, exists := flagMap[key]; exists {
		if flag := pflag.Lookup(flagName); flag != nil && flag.Changed {
			return "flag"
		}
	}

	envKey := "TASKMANAGER_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
	if os.Getenv(envKey) != "" {
		return "env"
	}

	if v.ConfigFileUsed() != "" && v.InConfig(key) {
		return "config file"
	}

	return "default"
}

// ShowConfig displays the current configuration with source information for each value.
func ShowConfig(cfg *Config, v *viper.Viper) {
	fmt.Println("Current Configuration:")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Printf("server.host: %s (%s)\n", cfg.ServerConfig.Host, getSource(v, "server.host"))
	fmt.Printf("server.port: %d (%s)\n", cfg.ServerConfig.Port, getSource(v, "server.port"))
	fmt.Printf("server.shutdown_timeout: %s (%s)\n", cfg.ServerConfig.ShutdownTimeout, getSource(v, "server.shutdown_timeout"))
	fmt.Printf("server.read_timeout: %s (%s)\n", cfg.ServerConfig.ReadTimeout, getSource(v, "server.read_timeout"))
	fmt.Printf("server.write_timeout: %s (%s)\n", cfg.ServerConfig.WriteTimeout, getSource(v, "server.write_timeout"))
	fmt.Printf("server.idle_timeout: %s (%s)\n", cfg.ServerConfig.IdleTimeout, getSource(v, "server.idle_timeout"))
	fmt.Printf("database.path: %s (%s)\n", cfg.DatabaseConfig.Path, getSource(v, "database.path"))
	fmt.Printf("jwt.secret: %s (%s)\n", maskSensitive(cfg.JWTConfig.Secret), getSource(v, "jwt.secret"))
	fmt.Printf("jwt.expiration: %s (%s)\n", cfg.JWTConfig.Expiration, getSource(v, "jwt.expiration"))
	fmt.Printf("logging.level: %s (%s)\n", cfg.LogConfig.Level, getSource(v, "logging.level"))
	fmt.Printf("logging.format: %s (%s)\n", cfg.LogConfig.Format, getSource(v, "logging.format"))
	fmt.Printf("logging.output: %s (%s)\n", cfg.LogConfig.Output, getSource(v, "logging.output"))
	fmt.Printf("logging.add_source: %v (%s)\n", cfg.LogConfig.AddSource, getSource(v, "logging.add_source"))
	fmt.Printf("logging.service_name: %s (%s)\n", cfg.LogConfig.ServiceName, getSource(v, "logging.service_name"))
	fmt.Printf("logging.environment: %s (%s)\n", cfg.LogConfig.Environment, getSource(v, "logging.environment"))
	fmt.Println()
	fmt.Println("Configuration Precedence: flags > env > config file > defaults")
}
