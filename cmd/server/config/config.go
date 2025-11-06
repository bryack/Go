package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Config struct {
	ServerConfig   ServerConfig   `mapstructure:"server"`
	DatabaseConfig DatabaseConfig `mapstructure:"database"`
	JWTConfig      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	Expiration time.Duration `mapstructure:"expiration"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("database.path", "./data/tasks.db")
	v.SetDefault("jwt.expiration", "24h")

	// Define and parse flags first (before reading config file)
	pflag.String("config", "", "Path to config file")
	pflag.Bool("show-config", false, "Display current configuration and exit")
	pflag.Int("port", 8080, "Server port")
	pflag.String("host", "0.0.0.0", "Server host")
	pflag.String("db-path", "./data/tasks.db", "Database path")
	pflag.String("jwt-expiration", "24h", "JWT expiration")
	pflag.String("jwt-secret", "", "JWT Secret")
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
			return nil, fmt.Errorf("failed to read config: %w", err)
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
	v.BindPFlag("database.path", pflag.Lookup("db-path"))
	v.BindPFlag("jwt.expiration", pflag.Lookup("jwt-expiration"))
	v.BindPFlag("jwt.secret", pflag.Lookup("jwt-secret"))

	// Unmarshal config into struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Validate configuration (will be implemented in task 3.1)
	// if err := config.Validate(); err != nil {
	// 	return nil, fmt.Errorf("config validation failed: %w", err)
	// }

	return &config, nil
}

func (config *Config) Validate() error {
	var errs []error
	if config.ServerConfig.Port < 1 || config.ServerConfig.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be between 1 and 65535, got %d", config.ServerConfig.Port))
	}

	if len(config.DatabaseConfig.Path) == 0 {
		errs = append(errs, fmt.Errorf("database path required"))
	}

	// err := validateDatabasePath(config.DatabaseConfig.Path)
	// if err != nil {
	// 	err = fmt.Errorf("validate database path '%s' failed: %w", config.DatabaseConfig.Path, err)
	// 	errs = append(errs, err)
	// }

	if len(config.JWTConfig.Secret) == 0 {
		errs = append(errs, fmt.Errorf("jwt secret required"))
	} else if len(config.JWTConfig.Secret) < 32 {
		errs = append(errs, fmt.Errorf("secret must be at least 32 symbols, got %d", len(config.JWTConfig.Secret)))
	}

	if config.JWTConfig.Expiration <= 0 {
		errs = append(errs, fmt.Errorf("expiration must be positive, got %v", config.JWTConfig.Expiration))
	}

	return errors.Join(errs...)
}
