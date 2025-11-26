# Implementation Plan

- [x] 1. Add Viper and pflag dependencies
  - Add `github.com/spf13/viper` to go.mod
  - Add `github.com/spf13/pflag` to go.mod
  - Run `go mod tidy` to download dependencies
  - _Requirements: 1.1, 1.2_

- [x] 2. Create config package structure
  - [x] 2.1 Create `cmd/server/config/config.go` with Config structs
    - Define `Config` struct with Server, Database, and JWT nested structs
    - Define `ServerConfig` struct with Port and Host fields
    - Define `DatabaseConfig` struct with Path field
    - Define `JWTConfig` struct with Secret and Expiration fields
    - Add mapstructure tags for Viper unmarshaling
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [x] 2.2 Implement configuration loading function
    - Create `LoadConfig()` function that returns `*Config` and error
    - Initialize Viper instance
    - Call `setDefaults()` to set default values
    - Configure config file search paths (current dir, /etc/taskmanager, ~/.taskmanager)
    - Read config file with error handling (ignore file not found)
    - Set up environment variable support with TASKMANAGER prefix
    - Bind command-line flags using pflag
    - Call `handleLegacyEnvVars()` for backward compatibility
    - Unmarshal Viper config into Config struct
    - Call `Validate()` on config before returning
    - _Requirements: 1.1, 2.1, 2.2, 2.3, 3.1, 3.2, 3.3, 7.1, 7.2, 8.1, 8.2_

  - [x] 2.3 Implement default values function
    - Create `setDefaults()` function accepting Viper instance
    - Set server.port default to 8080
    - Set server.host default to "0.0.0.0"
    - Set database.path default to "./data/tasks.db"
    - Set jwt.expiration default to 24 hours
    - Add comments noting jwt.secret has no default (must be provided)
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

  - [x] 2.4 Implement command-line flag binding
    - Create `bindFlags()` function accepting Viper instance
    - Define --port flag with default 8080 and description
    - Define --host flag with default "0.0.0.0" and description
    - Define --db-path flag with default "./data/tasks.db" and description
    - Define --jwt-secret flag with empty default and description
    - Define --jwt-expiration flag with default 24h and description
    - Define --config flag for custom config file path
    - Define --show-config flag to display configuration
    - Parse flags and bind to Viper
    - _Requirements: 1.2, 1.3, 1.4, 1.5, 6.1, 6.2_

- [x] 3. Implement configuration validation
  - [x] 3.1 Create Validate method on Config struct
    - Validate server port is between 1 and 65535
    - Validate database path is not empty
    - Call `validateDatabasePath()` helper function
    - Validate JWT secret is not empty
    - Validate JWT secret is at least 32 characters
    - Validate JWT expiration is positive duration
    - Collect all validation errors and return combined error
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [x] 3.2 Create database path validation helper
    - Create `validateDatabasePath()` function accepting path string
    - Extract directory from database path
    - Check if directory exists, create if not (with os.MkdirAll)
    - Test directory writability by creating temporary test file
    - Clean up test file
    - Return descriptive error if validation fails
    - _Requirements: 4.3_

- [x] 4. Implement legacy environment variable support
  - Create `handleLegacyEnvVars()` function accepting Viper instance
  - Define mapping of legacy env vars to new config keys (JWT_SECRET_KEY -> jwt.secret, TASK_DB_PATH -> database.path, PORT -> server.port)
  - Check each legacy env var and map to new key if set
  - Only use legacy value if new key is not already set (respect precedence)
  - Log deprecation warning when legacy env var is detected
  - _Requirements: 8.1, 8.2, 8.3_

- [x] 5. Implement configuration display functionality
  - [x] 5.1 Create Display method on Config struct
    - Create copy of config to avoid modifying original
    - Mask JWT secret using `maskSensitive()` helper
    - Marshal config to YAML format
    - Return formatted string
    - _Requirements: 6.2, 6.4, 9.1, 9.2_

  - [x] 5.2 Create sensitive value masking helper
    - Create `maskSensitive()` function accepting string value
    - Return "****" for values 4 characters or less
    - Return first 2 chars + "****" + last 2 chars for longer values
    - _Requirements: 9.1_

  - [x] 5.3 Create ShowConfig function with source tracking
    - Create `ShowConfig()` function accepting Config and Viper instance
    - Print header "Current Configuration"
    - Display each config value with its source (flag/env/file/default)
    - Use `getSource()` helper to determine source for each value
    - Mask sensitive values in display
    - Print configuration precedence order at bottom
    - _Requirements: 6.2, 6.3, 6.4_

  - [x] 5.4 Create configuration source detection helper
    - Create `getSource()` function accepting Viper instance and key
    - Check if value came from command-line flag
    - Check if value came from environment variable
    - Check if value came from config file
    - Return "default" if none of the above
    - _Requirements: 6.3_

- [x] 6. Update server main.go to use new config system
  - Import the new config package
  - Replace hardcoded values with config loading at startup
  - Call `config.LoadConfig()` and handle errors
  - Check for --show-config flag and display config if set
  - Use `cfg.Database.Path` for database initialization
  - Use `cfg.JWT.Secret` and `cfg.JWT.Expiration` for JWT service
  - Use `cfg.Server.Host` and `cfg.Server.Port` for HTTP server
  - Update startup log messages to show configured values
  - Remove direct os.Getenv calls for JWT_SECRET_KEY and TASK_DB_PATH
  - _Requirements: 1.1, 1.6, 6.1, 6.2_

- [ ] 7. Create example configuration file
  - Create `cmd/server/config.yaml.example` file
  - Include all configuration sections (server, database, jwt)
  - Add comments explaining each setting
  - Show example values
  - Include note about sensitive values and environment variables
  - _Requirements: 2.1, 2.2, 2.5, 5.4_

- [x] 8. Update Dockerfile for new config system
  - Update environment variable names in comments (document both legacy and new)
  - Add note about TASKMANAGER_ prefix for new env vars
  - Keep existing env vars for backward compatibility
  - Add example of mounting config file as volume
  - Update CMD to show flag usage examples in comments
  - _Requirements: 3.1, 8.1_

- [ ] 9. Write comprehensive tests
  - [x] 9.1 Test default values
    - Create test that loads config with no sources
    - Verify all default values are set correctly
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

  - [ ] 9.2 Test config file loading
    - Create temporary YAML config file
    - Load config and verify values from file
    - Test with JSON config file
    - Test with invalid syntax and verify error
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

  - [x] 9.3 Test environment variable mapping
    - Set TASKMANAGER_ prefixed env vars
    - Load config and verify env vars override defaults
    - Test nested key mapping (TASKMANAGER_SERVER_PORT)
    - _Requirements: 3.1, 3.2, 3.3, 3.4_

  - [ ] 9.4 Test command-line flag parsing
    - Parse flags programmatically in test
    - Load config and verify flags override env vars
    - Test all flag types (int, string, duration)
    - _Requirements: 1.2, 1.3, 1.4, 1.5_

  - [x] 9.5 Test configuration precedence
    - Set same value in all sources (default, file, env, flag)
    - Verify flag value wins
    - Remove flag, verify env value wins
    - Remove env, verify file value wins
    - Remove file, verify default value wins
    - _Requirements: 7.1, 7.2, 7.3_

  - [x] 9.6 Test validation logic
    - Test invalid port (0, -1, 99999)
    - Test empty database path
    - Test non-writable database directory
    - Test empty JWT secret
    - Test short JWT secret (less than 32 chars)
    - Test negative JWT expiration
    - Verify descriptive error messages
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

  - [ ] 9.7 Test legacy environment variables
    - Set JWT_SECRET_KEY and verify mapping
    - Set TASK_DB_PATH and verify mapping
    - Set PORT and verify mapping
    - Verify deprecation warnings are logged
    - Verify new env vars take precedence over legacy
    - _Requirements: 8.1, 8.2, 8.3_

  - [x] 9.8 Test sensitive value masking
    - Test maskSensitive with various string lengths
    - Test Display method masks JWT secret
    - Test ShowConfig masks sensitive values
    - _Requirements: 9.1, 9.2_

  - [x] 9.9 Test configuration display
    - Test ShowConfig output format
    - Test source detection for each config value
    - Verify human-readable output
    - _Requirements: 6.2, 6.3, 6.4_
