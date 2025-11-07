package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func TestDefaultValues(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name                string
		expectedPort        int
		expectedHost        string
		expectedDBPath      string
		expectedExpiration  time.Duration
		jwtSecret           string
		expectValidationErr bool
	}{
		{
			name:                "All default values with valid JWT secret",
			expectedPort:        8080,
			expectedHost:        "0.0.0.0",
			expectedDBPath:      "/tmp/data/tasks.db",
			expectedExpiration:  24 * time.Hour,
			jwtSecret:           "this-is-a-test-secret-key-with-32-chars-minimum",
			expectValidationErr: false,
		},
		{
			name:                "Missing JWT secret should fail validation",
			expectedPort:        8080,
			expectedHost:        "0.0.0.0",
			expectedDBPath:      "/tmp/data/tasks.db",
			expectedExpiration:  24 * time.Hour,
			jwtSecret:           "",
			expectValidationErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset pflag for each test to avoid flag already registered errors
			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

			// Clear any environment variables that might interfere
			os.Unsetenv("TASKMANAGER_SERVER_PORT")
			os.Unsetenv("TASKMANAGER_SERVER_HOST")
			os.Unsetenv("TASKMANAGER_DATABASE_PATH")
			os.Unsetenv("TASKMANAGER_JWT_EXPIRATION")
			os.Unsetenv("TASKMANAGER_JWT_SECRET")

			// Create a new viper instance for isolated testing
			v := viper.New()

			// Set default values (same as LoadConfig)
			v.SetDefault("server.port", 8080)
			v.SetDefault("server.host", "0.0.0.0")
			v.SetDefault("database.path", "/tmp/data/tasks.db")
			v.SetDefault("jwt.expiration", "24h")

			// Set JWT secret if provided
			if tc.jwtSecret != "" {
				v.Set("jwt.secret", tc.jwtSecret)
			}

			// Don't read any config file - testing defaults only
			// Don't set any environment variables - already cleared above
			// Don't parse any flags - testing defaults only

			// ====Act====
			var config Config
			err := v.Unmarshal(&config)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			validationErr := config.Validate()

			// ====Assert====
			if tc.expectValidationErr && validationErr == nil {
				t.Error("Expected validation error but got none")
			}

			if !tc.expectValidationErr && validationErr != nil {
				t.Errorf("Expected no validation error but got: %v", validationErr)
			}

			// Verify default values
			if config.ServerConfig.Port != tc.expectedPort {
				t.Errorf("Expected server.port %d, got %d", tc.expectedPort, config.ServerConfig.Port)
			}

			if config.ServerConfig.Host != tc.expectedHost {
				t.Errorf("Expected server.host %q, got %q", tc.expectedHost, config.ServerConfig.Host)
			}

			if config.DatabaseConfig.Path != tc.expectedDBPath {
				t.Errorf("Expected database.path %q, got %q", tc.expectedDBPath, config.DatabaseConfig.Path)
			}

			if config.JWTConfig.Expiration != tc.expectedExpiration {
				t.Errorf("Expected jwt.expiration %v, got %v", tc.expectedExpiration, config.JWTConfig.Expiration)
			}

			if config.JWTConfig.Secret != tc.jwtSecret {
				t.Errorf("Expected jwt.secret %q, got %q", tc.jwtSecret, config.JWTConfig.Secret)
			}
		})
	}
}

func TestConfigurationPrecedence(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name           string
		setupFunc      func(v *viper.Viper, tmpFile *os.File) error
		expectedPort   int
		expectedSource string
	}{
		{
			name: "Flag value wins over all other sources",
			setupFunc: func(v *viper.Viper, tmpFile *os.File) error {
				// Set default
				v.SetDefault("server.port", 8080)

				// Set config file
				v.SetConfigFile(tmpFile.Name())
				if err := v.ReadInConfig(); err != nil {
					return err
				}

				// Set environment variable
				os.Setenv("TASKMANAGER_SERVER_PORT", "9500")

				v.AutomaticEnv()
				v.SetEnvPrefix("TASKMANAGER")
				v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
				v.BindEnv("server.port")

				// Set flag (highest priority) - use Set to simulate flag value
				v.Set("server.port", 10000)

				return nil
			},
			expectedPort:   10000,
			expectedSource: "flag",
		},
		{
			name: "Env value wins when flag is not set",
			setupFunc: func(v *viper.Viper, tmpFile *os.File) error {
				// Set default
				v.SetDefault("server.port", 8080)

				// Set config file
				v.SetConfigFile(tmpFile.Name())
				if err := v.ReadInConfig(); err != nil {
					return err
				}

				// Set environment variable (should override file)
				os.Setenv("TASKMANAGER_SERVER_PORT", "9500")

				v.AutomaticEnv()
				v.SetEnvPrefix("TASKMANAGER")
				v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
				v.BindEnv("server.port")

				// Don't set flag - env should win

				return nil
			},
			expectedPort:   9500,
			expectedSource: "env",
		},
		{
			name: "File value wins when flag and env are not set",
			setupFunc: func(v *viper.Viper, tmpFile *os.File) error {
				// Set default
				v.SetDefault("server.port", 8080)

				// Set config file
				v.SetConfigFile(tmpFile.Name())
				if err := v.ReadInConfig(); err != nil {
					return err
				}

				// Don't set environment variable
				// Don't set flag

				return nil
			},
			expectedPort:   9000,
			expectedSource: "file",
		},
		{
			name: "Default value wins when no other sources are set",
			setupFunc: func(v *viper.Viper, tmpFile *os.File) error {
				// Set default
				v.SetDefault("server.port", 8080)

				// Don't set config file
				// Don't set environment variable
				// Don't set flag

				return nil
			},
			expectedPort:   8080,
			expectedSource: "default",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset pflag for each test
			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

			// Clear all environment variables
			os.Unsetenv("TASKMANAGER_SERVER_PORT")
			os.Unsetenv("TASKMANAGER_SERVER_HOST")
			os.Unsetenv("TASKMANAGER_DATABASE_PATH")
			os.Unsetenv("TASKMANAGER_JWT_EXPIRATION")
			os.Unsetenv("TASKMANAGER_JWT_SECRET")

			// Create a temporary config file
			configContent := `
server:
  port: 9000
  host: 0.0.0.0
database:
  path: /tmp/data/tasks.db
jwt:
  secret: config-file-secret-key-with-32-chars
  expiration: 24h
`
			tmpFile, err := os.CreateTemp("", "config-*.yaml")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(configContent); err != nil {
				t.Fatalf("Failed to write config: %v", err)
			}
			tmpFile.Close()

			// Create a new viper instance
			v := viper.New()

			// ====Act====
			err = tc.setupFunc(v, tmpFile)
			if err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			var config Config
			err = v.Unmarshal(&config)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			// Clean up environment variables after test
			os.Unsetenv("TASKMANAGER_SERVER_PORT")

			// ====Assert====
			if config.ServerConfig.Port != tc.expectedPort {
				t.Errorf("Expected server.port %d from %s, got %d", tc.expectedPort, tc.expectedSource, config.ServerConfig.Port)
			}
		})
	}
}

func TestEnvironmentVariableMapping(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name                string
		envVars             map[string]string
		expectedPort        int
		expectedHost        string
		expectedDBPath      string
		expectedJWTSecret   string
		expectedExpiration  time.Duration
		expectValidationErr bool
	}{
		{
			name: "TASKMANAGER_ prefixed env vars override defaults",
			envVars: map[string]string{
				"TASKMANAGER_SERVER_PORT":    "9090",
				"TASKMANAGER_SERVER_HOST":    "127.0.0.1",
				"TASKMANAGER_DATABASE_PATH":  "/tmp/custom/path/tasks.db",
				"TASKMANAGER_JWT_SECRET":     "custom-secret-key-with-32-chars-min",
				"TASKMANAGER_JWT_EXPIRATION": "48h",
			},
			expectedPort:        9090,
			expectedHost:        "127.0.0.1",
			expectedDBPath:      "/tmp/custom/path/tasks.db",
			expectedJWTSecret:   "custom-secret-key-with-32-chars-min",
			expectedExpiration:  48 * time.Hour,
			expectValidationErr: false,
		},
		{
			name: "Nested key mapping with underscores (TASKMANAGER_SERVER_PORT)",
			envVars: map[string]string{
				"TASKMANAGER_SERVER_PORT": "3000",
				"TASKMANAGER_JWT_SECRET":  "another-secret-key-with-32-chars",
			},
			expectedPort:        3000,
			expectedHost:        "0.0.0.0",            // default
			expectedDBPath:      "/tmp/data/tasks.db", // default
			expectedJWTSecret:   "another-secret-key-with-32-chars",
			expectedExpiration:  24 * time.Hour, // default
			expectValidationErr: false,
		},
		{
			name: "Partial env vars with defaults for missing values",
			envVars: map[string]string{
				"TASKMANAGER_JWT_SECRET":  "partial-secret-key-with-32-chars-",
				"TASKMANAGER_SERVER_HOST": "localhost",
			},
			expectedPort:        8080, // default
			expectedHost:        "localhost",
			expectedDBPath:      "/tmp/data/tasks.db", // default
			expectedJWTSecret:   "partial-secret-key-with-32-chars-",
			expectedExpiration:  24 * time.Hour, // default
			expectValidationErr: false,
		},
		{
			name: "All env vars set with valid values",
			envVars: map[string]string{
				"TASKMANAGER_SERVER_PORT":    "8888",
				"TASKMANAGER_SERVER_HOST":    "0.0.0.0",
				"TASKMANAGER_DATABASE_PATH":  "/tmp/test/tasks.db",
				"TASKMANAGER_JWT_SECRET":     "all-env-vars-secret-key-32-chars",
				"TASKMANAGER_JWT_EXPIRATION": "12h",
			},
			expectedPort:        8888,
			expectedHost:        "0.0.0.0",
			expectedDBPath:      "/tmp/test/tasks.db",
			expectedJWTSecret:   "all-env-vars-secret-key-32-chars",
			expectedExpiration:  12 * time.Hour,
			expectValidationErr: false,
		},
		{
			name: "Invalid port from env var should fail validation",
			envVars: map[string]string{
				"TASKMANAGER_SERVER_PORT": "99999",
				"TASKMANAGER_JWT_SECRET":  "invalid-port-secret-key-32-chars",
			},
			expectedPort:        99999,
			expectedHost:        "0.0.0.0",
			expectedDBPath:      "/tmp/data/tasks.db",
			expectedJWTSecret:   "invalid-port-secret-key-32-chars",
			expectedExpiration:  24 * time.Hour,
			expectValidationErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset pflag for each test
			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

			// Clear all environment variables first
			os.Unsetenv("TASKMANAGER_SERVER_PORT")
			os.Unsetenv("TASKMANAGER_SERVER_HOST")
			os.Unsetenv("TASKMANAGER_DATABASE_PATH")
			os.Unsetenv("TASKMANAGER_JWT_EXPIRATION")
			os.Unsetenv("TASKMANAGER_JWT_SECRET")

			// Set test-specific environment variables
			for key, value := range tc.envVars {
				os.Setenv(key, value)
			}

			// Clean up environment variables after test
			defer func() {
				for key := range tc.envVars {
					os.Unsetenv(key)
				}
			}()

			// Create a new viper instance
			v := viper.New()

			// Set defaults
			v.SetDefault("server.port", 8080)
			v.SetDefault("server.host", "0.0.0.0")
			v.SetDefault("database.path", "/tmp/data/tasks.db")
			v.SetDefault("jwt.expiration", "24h")

			// Configure environment variable support (same as LoadConfig)
			v.AutomaticEnv()
			v.SetEnvPrefix("TASKMANAGER")
			v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

			// Explicitly bind environment variables for nested keys
			v.BindEnv("server.port")
			v.BindEnv("server.host")
			v.BindEnv("database.path")
			v.BindEnv("jwt.secret")
			v.BindEnv("jwt.expiration")

			// ====Act====
			var config Config
			err := v.Unmarshal(&config)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			validationErr := config.Validate()

			// ====Assert====
			if tc.expectValidationErr && validationErr == nil {
				t.Error("Expected validation error but got none")
			}

			if !tc.expectValidationErr && validationErr != nil {
				t.Errorf("Expected no validation error but got: %v", validationErr)
			}

			// Verify values from environment variables
			if config.ServerConfig.Port != tc.expectedPort {
				t.Errorf("Expected server.port %d, got %d", tc.expectedPort, config.ServerConfig.Port)
			}

			if config.ServerConfig.Host != tc.expectedHost {
				t.Errorf("Expected server.host %q, got %q", tc.expectedHost, config.ServerConfig.Host)
			}

			if config.DatabaseConfig.Path != tc.expectedDBPath {
				t.Errorf("Expected database.path %q, got %q", tc.expectedDBPath, config.DatabaseConfig.Path)
			}

			if config.JWTConfig.Secret != tc.expectedJWTSecret {
				t.Errorf("Expected jwt.secret %q, got %q", tc.expectedJWTSecret, config.JWTConfig.Secret)
			}

			if config.JWTConfig.Expiration != tc.expectedExpiration {
				t.Errorf("Expected jwt.expiration %v, got %v", tc.expectedExpiration, config.JWTConfig.Expiration)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name        string
		config      Config
		expectedErr bool
		errContains string
	}{
		{
			name: "Valid configuration",
			config: Config{
				ServerConfig: ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
				},
				DatabaseConfig: DatabaseConfig{
					Path: "/tmp/test-valid/tasks.db",
				},
				JWTConfig: JWTConfig{
					Secret:     "this-is-a-valid-secret-key-with-32-characters",
					Expiration: 24 * time.Hour,
				},
			},
			expectedErr: false,
			errContains: "",
		},
		{
			name: "Invalid port",
			config: Config{
				ServerConfig: ServerConfig{
					Port: 99999,
					Host: "0.0.0.0",
				},
				DatabaseConfig: DatabaseConfig{
					Path: "/tmp/test-port/tasks.db",
				},
				JWTConfig: JWTConfig{
					Secret:     "this-is-a-valid-secret-key-with-32-characters",
					Expiration: 24 * time.Hour,
				},
			},
			expectedErr: true,
			errContains: "server.port must be between 1 and 65535",
		},
		{
			name: "Empty database path",
			config: Config{
				ServerConfig: ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
				},
				DatabaseConfig: DatabaseConfig{
					Path: "",
				},
				JWTConfig: JWTConfig{
					Secret:     "this-is-a-valid-secret-key-with-32-characters",
					Expiration: 24 * time.Hour,
				},
			},
			expectedErr: true,
			errContains: "database path required",
		},
		{
			name: "Non-writable database directory",
			config: Config{
				ServerConfig: ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
				},
				DatabaseConfig: DatabaseConfig{
					Path: "/root/restricted/tasks.db",
				},
				JWTConfig: JWTConfig{
					Secret:     "this-is-a-valid-secret-key-with-32-characters",
					Expiration: 24 * time.Hour,
				},
			},
			expectedErr: true,
			errContains: "validate database path",
		},
		{
			name: "Empty JWT secret",
			config: Config{
				ServerConfig: ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
				},
				DatabaseConfig: DatabaseConfig{
					Path: "/tmp/test-empty-secret/tasks.db",
				},
				JWTConfig: JWTConfig{
					Secret:     "",
					Expiration: 24 * time.Hour,
				},
			},
			expectedErr: true,
			errContains: "jwt secret required",
		},
		{
			name: "JWT secret too short",
			config: Config{
				ServerConfig: ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
				},
				DatabaseConfig: DatabaseConfig{
					Path: "/tmp/test-short-secret/tasks.db",
				},
				JWTConfig: JWTConfig{
					Secret:     "short",
					Expiration: 24 * time.Hour,
				},
			},
			expectedErr: true,
			errContains: "secret must be at least 32 symbols",
		},
		{
			name: "JWT secret exactly 32 chars - valid",
			config: Config{
				ServerConfig: ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
				},
				DatabaseConfig: DatabaseConfig{
					Path: "/tmp/test-32-chars/tasks.db",
				},
				JWTConfig: JWTConfig{
					Secret:     "12345678901234567890123456789012",
					Expiration: 24 * time.Hour,
				},
			},
			expectedErr: false,
			errContains: "",
		},
		{
			name: "Invalid JWT expiration",
			config: Config{
				ServerConfig: ServerConfig{
					Port: 8080,
					Host: "0.0.0.0",
				},
				DatabaseConfig: DatabaseConfig{
					Path: "/tmp/test-expiration/tasks.db",
				},
				JWTConfig: JWTConfig{
					Secret:     "this-is-a-valid-secret-key-with-32-characters",
					Expiration: 0,
				},
			},
			expectedErr: true,
			errContains: "expiration must be positive",
		},
		{
			name: "Multiple validation errors",
			config: Config{
				ServerConfig: ServerConfig{
					Port: 0,
					Host: "0.0.0.0",
				},
				DatabaseConfig: DatabaseConfig{
					Path: "",
				},
				JWTConfig: JWTConfig{
					Secret:     "short",
					Expiration: -1 * time.Hour,
				},
			},
			expectedErr: true,
			errContains: "server.port must be between 1 and 65535",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			err := tc.config.Validate()

			// ====Assert====
			if tc.expectedErr && err == nil {
				t.Error("Expected validation error but got none")
			}

			if !tc.expectedErr && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
			}

			if tc.expectedErr && err != nil {
				if !strings.Contains(err.Error(), tc.errContains) {
					t.Errorf("Expected error to contain %q, but got: %v", tc.errContains, err)
				}
			}
		})
	}
}

func TestMaskSensitive(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Short string (4 chars or less)",
			input:    "abc",
			expected: "****",
		},
		{
			name:     "Boundary case (5 characters)",
			input:    "abcde",
			expected: "ab****de",
		},
		{
			name:     "Long secret (32+ characters)",
			input:    "this-is-a-valid-secret-key-with-32-characters",
			expected: "th****rs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			result := maskSensitive(tc.input)

			// ====Assert====
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestShowConfigMasksSensitiveValues(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name              string
		jwtSecret         string
		expectedMasked    string
		shouldContainPort bool
	}{
		{
			name:              "Short secret is fully masked",
			jwtSecret:         "abc",
			expectedMasked:    "****",
			shouldContainPort: true,
		},
		{
			name:              "Long secret shows first and last 2 chars",
			jwtSecret:         "this-is-a-very-long-secret-key-with-32-chars",
			expectedMasked:    "th****rs",
			shouldContainPort: true,
		},
		{
			name:              "32 character secret is properly masked",
			jwtSecret:         "12345678901234567890123456789012",
			expectedMasked:    "12****12",
			shouldContainPort: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset pflag for each test
			pflag.CommandLine = pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)

			// Clear environment variables
			os.Unsetenv("TASKMANAGER_SERVER_PORT")
			os.Unsetenv("TASKMANAGER_SERVER_HOST")
			os.Unsetenv("TASKMANAGER_DATABASE_PATH")
			os.Unsetenv("TASKMANAGER_JWT_EXPIRATION")
			os.Unsetenv("TASKMANAGER_JWT_SECRET")

			// Create a new viper instance
			v := viper.New()

			// Set defaults
			v.SetDefault("server.port", 8080)
			v.SetDefault("server.host", "0.0.0.0")
			v.SetDefault("database.path", "/tmp/data/tasks.db")
			v.SetDefault("jwt.expiration", "24h")
			v.Set("jwt.secret", tc.jwtSecret)

			var config Config
			err := v.Unmarshal(&config)
			if err != nil {
				t.Fatalf("Failed to unmarshal config: %v", err)
			}

			// Capture output from ShowConfig
			var output strings.Builder
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// ====Act====
			ShowConfig(&config, v)

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = oldStdout
			buf := make([]byte, 4096)
			n, _ := r.Read(buf)
			output.WriteString(string(buf[:n]))

			// ====Assert====
			outputStr := output.String()

			// Verify the masked secret appears in output
			if !strings.Contains(outputStr, tc.expectedMasked) {
				t.Errorf("Expected output to contain masked secret %q, but got:\n%s", tc.expectedMasked, outputStr)
			}

			// Verify the actual secret does NOT appear in output
			if strings.Contains(outputStr, tc.jwtSecret) {
				t.Errorf("Output should not contain actual secret %q, but it does:\n%s", tc.jwtSecret, outputStr)
			}

			// Verify other config values are present
			if tc.shouldContainPort && !strings.Contains(outputStr, "8080") {
				t.Errorf("Expected output to contain port 8080, but got:\n%s", outputStr)
			}

			// Verify the configuration precedence message is present
			if !strings.Contains(outputStr, "Configuration Precedence") {
				t.Errorf("Expected output to contain precedence message, but got:\n%s", outputStr)
			}
		})
	}
}
