package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const MinJWTSecretLength = 32

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

func LoadConfig() (*Config, *viper.Viper, error) {
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
	v.BindPFlag("database.path", pflag.Lookup("db-path"))
	v.BindPFlag("jwt.expiration", pflag.Lookup("jwt-expiration"))
	v.BindPFlag("jwt.secret", pflag.Lookup("jwt-secret"))

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

func (config *Config) Validate() error {
	var errs []error
	if config.ServerConfig.Port < 1 || config.ServerConfig.Port > 65535 {
		errs = append(errs, fmt.Errorf("server.port must be between 1 and 65535, got %d", config.ServerConfig.Port))
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

	return errors.Join(errs...)
}

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

func maskSensitive(scrt string) string {
	if len(scrt) <= 4 {
		return "****"
	}

	return scrt[0:2] + "****" + scrt[len(scrt)-2:]
}

func getSource(v *viper.Viper, key string) string {
	flagMap := map[string]string{
		"server.port":    "port",
		"server.host":    "host",
		"database.path":  "db-path",
		"jwt.secret":     "jwt-secret",
		"jwt.expiration": "jwt-expiration",
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

func ShowConfig(cfg *Config, v *viper.Viper) {
	fmt.Println("Current Configuration:")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Printf("server.host: %s (%s)\n", cfg.ServerConfig.Host, getSource(v, "server.host"))
	fmt.Printf("server.port: %d (%s)\n", cfg.ServerConfig.Port, getSource(v, "server.port"))
	fmt.Printf("database.path: %s (%s)\n", cfg.DatabaseConfig.Path, getSource(v, "database.path"))
	fmt.Printf("jwt.secret: %s (%s)\n", maskSensitive(cfg.JWTConfig.Secret), getSource(v, "jwt.secret"))
	fmt.Printf("jwt.expiration: %s (%s)\n", cfg.JWTConfig.Expiration, getSource(v, "jwt.expiration"))
	fmt.Println()
	fmt.Println("Configuration Precedence: flags > env > config file > defaults")
}
