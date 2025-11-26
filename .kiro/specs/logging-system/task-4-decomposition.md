# Task Decomposition: Extend Configuration System with Logging Options

## Overview

This task extends the existing configuration system to support logging configuration. You'll add a new `LogConfig` struct to the existing `Config` struct, define logging-specific fields (level, format, output, etc.), set appropriate defaults, add validation, and integrate with the existing viper-based configuration system that supports config files, environment variables, and command-line flags.

## Implementation Approach

We're extending the existing configuration pattern used in `cmd/server/config/config.go`. The current system uses:
1. Viper for configuration management
2. Struct tags (`mapstructure`) for automatic unmarshaling
3. Three-tier configuration precedence: flags > environment variables > config file > defaults
4. Validation in a `Validate()` method
5. A `ShowConfig()` function to display current configuration

The logging configuration will follow the same patterns, adding a new `LogConfig` struct that integrates seamlessly with the existing system. This ensures consistency and makes it easy for operators to configure logging using their preferred method (config file, environment variables, or flags).

**Key Concepts:**
- **Viper**: Go configuration library that handles multiple sources (files, env vars, flags)
- **Mapstructure**: Struct tags that tell viper how to map config keys to struct fields
- **Configuration Precedence**: Order in which config sources override each other
- **Validation**: Ensuring configuration values are valid before use

## Prerequisites

**Existing Code:**
- `cmd/server/config/config.go` - Current configuration system
- `logger/config.go` - Logger Config struct (from Task 1)
- `logger/logger.go` - Logger factory that uses Config

**Dependencies:**
- `github.com/spf13/viper` (already installed)
- `github.com/spf13/pflag` (already installed)

**Knowledge Required:**
- Understanding of the existing config system structure
- Familiarity with viper configuration library
- Understanding of struct tags in Go
- Basic understanding of configuration precedence

## Step-by-Step Instructions

### Step 1: Add LogConfig struct to Config

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add a new field to the main Config struct for logging configuration.

**What to implement:**
- Locate the `Config` struct (currently has ServerConfig, DatabaseConfig, JWTConfig)
- Add a new field: `LogConfig` of type `LogConfig`
- Add the mapstructure tag: `mapstructure:"logging"`
- This field should be added after JWTConfig to maintain logical grouping

**Why:**
The Config struct is the top-level configuration container. Adding LogConfig here makes logging configuration available throughout the application. The mapstructure tag tells viper to look for a "logging" section in config files.

**Expected result:**
The Config struct now has four fields: ServerConfig, DatabaseConfig, JWTConfig, and LogConfig. The file still compiles.

---

### Step 2: Define LogConfig struct with basic fields

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Create the LogConfig struct with fields for basic logging configuration.

**What to implement:**
- Create a new struct named `LogConfig` (after the JWTConfig struct definition)
- Add the following fields with mapstructure tags:
  - `Level` (string) - tag: `mapstructure:"level"` - log level (debug, info, warn, error)
  - `Format` (string) - tag: `mapstructure:"format"` - output format (json, text)
  - `Output` (string) - tag: `mapstructure:"output"` - destination (stdout, stderr, or file path)
  - `AddSource` (bool) - tag: `mapstructure:"add_source"` - include file:line in logs
  - `ServiceName` (string) - tag: `mapstructure:"service_name"` - service identifier
  - `Environment` (string) - tag: `mapstructure:"environment"` - deployment environment
- Each field should have a mapstructure tag matching the field name in snake_case

**Why:**
These fields control the core logging behavior. The mapstructure tags define how config file keys map to struct fields. Using snake_case in tags follows YAML/JSON conventions and matches the logger.Config struct from Task 1.

**Expected result:**
You have a LogConfig struct with six basic fields. The struct compiles and matches the logger.Config structure.

---

### Step 3: Add file rotation fields to LogConfig

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add fields for log file rotation configuration.

**What to implement:**
- Add the following fields to the LogConfig struct (after Environment):
  - `EnableRotation` (bool) - tag: `mapstructure:"enable_rotation"` - enable log rotation
  - `MaxSize` (int) - tag: `mapstructure:"max_size"` - max file size in MB
  - `MaxAge` (int) - tag: `mapstructure:"max_age"` - max age in days
  - `MaxBackups` (int) - tag: `mapstructure:"max_backups"` - max number of old files to keep
- These fields are only used when Output is a file path

**Why:**
Log rotation prevents log files from growing indefinitely and consuming all disk space. These settings control when and how logs are rotated. They're optional (only used for file output) but important for production deployments.

**Expected result:**
LogConfig struct now has ten fields total (six basic + four rotation). The struct is complete.

---

### Step 4: Set default values for logging configuration

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add default values for logging configuration in the LoadConfig function.

**What to implement:**
- Locate the `LoadConfig()` function
- Find the section where defaults are set (after `v := viper.New()`)
- Add default values for logging fields:
  - `v.SetDefault("logging.level", "info")` - reasonable default for production
  - `v.SetDefault("logging.format", "json")` - machine-readable for production
  - `v.SetDefault("logging.output", "stdout")` - standard output
  - `v.SetDefault("logging.add_source", false)` - disabled by default (performance)
  - `v.SetDefault("logging.service_name", "task-manager-api")` - service identifier
  - `v.SetDefault("logging.environment", "production")` - assume production
  - `v.SetDefault("logging.enable_rotation", false)` - disabled by default
  - `v.SetDefault("logging.max_size", 100)` - 100 MB
  - `v.SetDefault("logging.max_age", 30)` - 30 days
  - `v.SetDefault("logging.max_backups", 5)` - keep 5 old files
- Add these after the existing defaults

**Why:**
Defaults ensure the application works without configuration. These defaults are production-ready: JSON format for log aggregation, INFO level to reduce noise, stdout for container compatibility. The rotation defaults are conservative to prevent disk space issues.

**Expected result:**
All logging fields have sensible defaults. The application can run without any logging configuration.

---

### Step 5: Define command-line flags for logging

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add command-line flags for logging configuration.

**What to implement:**
- Locate the flag definitions in `LoadConfig()` (after `pflag.String("config", ...)`)
- Add flags for logging configuration:
  - `pflag.String("log-level", "info", "Log level (debug, info, warn, error)")`
  - `pflag.String("log-format", "json", "Log format (json, text)")`
  - `pflag.String("log-output", "stdout", "Log output (stdout, stderr, or file path)")`
  - `pflag.Bool("log-add-source", false, "Include source file and line in logs")`
  - `pflag.String("log-service-name", "task-manager-api", "Service name for logs")`
  - `pflag.String("log-environment", "production", "Environment name (development, staging, production)")`
- Add these after the existing flags
- Note: File rotation flags are typically not needed as command-line flags (config file is better)

**Why:**
Command-line flags provide the highest precedence in the configuration system. They're useful for quick testing and overriding config files. The flag names use kebab-case (log-level) which is standard for CLI flags.

**Expected result:**
You can pass logging configuration via command-line flags. Running with `--help` shows the new logging flags.

---

### Step 6: Bind flags to configuration keys

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Connect the command-line flags to viper configuration keys.

**What to implement:**
- Locate the flag binding section in `LoadConfig()` (after `v.BindPFlag("jwt.secret", ...)`)
- Add bindings for logging flags:
  - `v.BindPFlag("logging.level", pflag.Lookup("log-level"))`
  - `v.BindPFlag("logging.format", pflag.Lookup("log-format"))`
  - `v.BindPFlag("logging.output", pflag.Lookup("log-output"))`
  - `v.BindPFlag("logging.add_source", pflag.Lookup("log-add-source"))`
  - `v.BindPFlag("logging.service_name", pflag.Lookup("log-service-name"))`
  - `v.BindPFlag("logging.environment", pflag.Lookup("log-environment"))`
- Add these after the existing bindings

**Why:**
Binding connects flags to config keys so viper knows which flag corresponds to which config field. This enables the precedence system: flags override environment variables and config files.

**Expected result:**
Command-line flags now override other configuration sources. The precedence system works correctly.

---

### Step 7: Add validation for log level

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add validation logic for the log level field.

**What to implement:**
- Locate the `Validate()` method on the Config struct
- Add validation for log level (after JWT validation):
  - Convert level to lowercase for comparison
  - Check if level is one of: "debug", "info", "warn", "error"
  - If invalid, append error: `fmt.Errorf("logging.level must be one of [debug, info, warn, error], got %s", config.LogConfig.Level)`
- Use the existing error accumulation pattern (append to errs slice)

**Why:**
Invalid log levels would cause runtime errors when creating the logger. Validating early provides clear error messages and prevents the application from starting with invalid configuration.

**Expected result:**
Invalid log levels are rejected with a clear error message. Valid levels pass validation.

---

### Step 8: Add validation for log format

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add validation logic for the log format field.

**What to implement:**
- In the `Validate()` method, after log level validation:
  - Convert format to lowercase for comparison
  - Check if format is one of: "json", "text"
  - If invalid, append error: `fmt.Errorf("logging.format must be one of [json, text], got %s", config.LogConfig.Format)`

**Why:**
Only JSON and text formats are supported by the logger. Validating ensures users don't specify unsupported formats.

**Expected result:**
Invalid formats are rejected. Valid formats (json, text) pass validation.

---

### Step 9: Add validation for output destination

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add validation logic for the output destination field.

**What to implement:**
- In the `Validate()` method, after format validation:
  - Check if output is empty
  - If empty, append error: `fmt.Errorf("logging.output cannot be empty")`
  - For file paths (not "stdout" or "stderr"):
    - Extract directory using `filepath.Dir(output)`
    - Check if directory is writable (similar to validateDatabasePath)
    - Consider creating a helper function `validateLogPath` similar to `validateDatabasePath`
- Note: Full file path validation can be complex; basic checks are sufficient

**Why:**
Empty output would cause logger creation to fail. For file paths, we want to catch permission issues early rather than at runtime. However, full validation is tricky (file might not exist yet), so basic checks are sufficient.

**Expected result:**
Empty output is rejected. File paths are checked for basic validity.

---

### Step 10: Add validation for service name and environment

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add validation for required string fields.

**What to implement:**
- In the `Validate()` method, after output validation:
  - Check if ServiceName is empty
  - If empty, append error: `fmt.Errorf("logging.service_name cannot be empty")`
  - Check if Environment is empty
  - If empty, append error: `fmt.Errorf("logging.environment cannot be empty")`
- These fields are required for log aggregation and filtering

**Why:**
Service name and environment are critical for log aggregation tools. They allow filtering logs by service and environment. Empty values would make logs difficult to identify and filter.

**Expected result:**
Empty service name or environment is rejected. Non-empty values pass validation.

---

### Step 11: Add validation for rotation settings

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add validation for log rotation configuration.

**What to implement:**
- In the `Validate()` method, after service name/environment validation:
  - If EnableRotation is true:
    - Check MaxSize > 0, append error if not: `fmt.Errorf("logging.max_size must be positive when rotation is enabled, got %d", config.LogConfig.MaxSize)`
    - Check MaxAge > 0, append error if not: `fmt.Errorf("logging.max_age must be positive when rotation is enabled, got %d", config.LogConfig.MaxAge)`
    - Check MaxBackups >= 0, append error if not: `fmt.Errorf("logging.max_backups must be non-negative, got %d", config.LogConfig.MaxBackups)`
- Only validate rotation settings if rotation is enabled

**Why:**
Invalid rotation settings (negative or zero values) would cause issues with the rotation library. Validating ensures sensible values. MaxBackups can be 0 (no backups kept), but MaxSize and MaxAge must be positive.

**Expected result:**
Invalid rotation settings are rejected when rotation is enabled. Valid settings pass validation.

---

### Step 12: Update ShowConfig to display logging configuration

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add logging configuration to the ShowConfig output.

**What to implement:**
- Locate the `ShowConfig()` function
- After the JWT configuration output, add logging configuration:
  - `fmt.Printf("logging.level: %s (%s)\n", cfg.LogConfig.Level, getSource(v, "logging.level"))`
  - `fmt.Printf("logging.format: %s (%s)\n", cfg.LogConfig.Format, getSource(v, "logging.format"))`
  - `fmt.Printf("logging.output: %s (%s)\n", cfg.LogConfig.Output, getSource(v, "logging.output"))`
  - `fmt.Printf("logging.add_source: %v (%s)\n", cfg.LogConfig.AddSource, getSource(v, "logging.add_source"))`
  - `fmt.Printf("logging.service_name: %s (%s)\n", cfg.LogConfig.ServiceName, getSource(v, "logging.service_name"))`
  - `fmt.Printf("logging.environment: %s (%s)\n", cfg.LogConfig.Environment, getSource(v, "logging.environment"))`
- Optionally add rotation settings if EnableRotation is true

**Why:**
ShowConfig helps users understand their current configuration and where values came from (flag, env, config file, or default). Adding logging configuration makes it easy to debug configuration issues.

**Expected result:**
Running with `--show-config` displays logging configuration with sources. Users can see their effective logging configuration.

---

### Step 13: Update getSource function for logging keys

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add logging configuration keys to the getSource function's flag map.

**What to implement:**
- Locate the `getSource()` function
- Find the `flagMap` variable
- Add entries for logging flags:
  - `"logging.level": "log-level"`
  - `"logging.format": "log-format"`
  - `"logging.output": "log-output"`
  - `"logging.add_source": "log-add-source"`
  - `"logging.service_name": "log-service-name"`
  - `"logging.environment": "log-environment"`
- Add these after the existing entries

**Why:**
The getSource function determines where a config value came from. Adding logging keys ensures ShowConfig correctly identifies when logging values come from flags vs other sources.

**Expected result:**
ShowConfig correctly identifies the source of logging configuration values.

---

### Step 14: Create helper function to convert to logger.Config

**File**: `cmd/server/config/config.go` (modify existing)

**What to do:**
Add a method to convert the config LogConfig to logger.Config.

**What to implement:**
- Create a method on LogConfig: `func (lc *LogConfig) ToLoggerConfig() logger.Config`
- Return a logger.Config struct with fields mapped from LogConfig:
  - Level: lc.Level
  - Format: lc.Format
  - Output: lc.Output
  - AddSource: lc.AddSource
  - ServiceName: lc.ServiceName
  - Environment: lc.Environment
  - EnableRotation: lc.EnableRotation
  - MaxSize: lc.MaxSize
  - MaxAge: lc.MaxAge
  - MaxBackups: lc.MaxBackups
- This provides a clean conversion between config types
- You'll need to import the logger package: `import "yourproject/logger"`

**Why:**
The config package uses its own LogConfig struct (for viper integration), while the logger package has its own Config struct. This method provides a clean conversion between them, keeping the packages decoupled.

**Expected result:**
You can easily convert config.LogConfig to logger.Config for creating loggers.

---

## Complete Implementation Flow

Here's how the configuration system works after these changes:

1. **Application starts** → LoadConfig() is called
2. **Defaults are set** → All logging fields have default values
3. **Config file is read** → YAML values override defaults
4. **Environment variables are read** → Env vars override config file
5. **Command-line flags are parsed** → Flags override everything
6. **Configuration is validated** → Invalid values are rejected
7. **Logger is created** → Using cfg.LogConfig.ToLoggerConfig()

## Verification

### Compile Check
```bash
go build ./cmd/server
```
**Expected**: No compilation errors

### Test Default Configuration
```bash
# Run with defaults
./server --show-config
```
**Expected output** (includes logging section):
```
Current Configuration:
=====================

server.host: 0.0.0.0 (default)
server.port: 8080 (default)
database.path: ./data/tasks.db (default)
jwt.secret: **** (default)
jwt.expiration: 24h0m0s (default)
logging.level: info (default)
logging.format: json (default)
logging.output: stdout (default)
logging.add_source: false (default)
logging.service_name: task-manager-api (default)
logging.environment: production (default)

Configuration Precedence: flags > env > config file > defaults
```

### Test Config File
```bash
# Create test config file
cat > test-config.yaml << 'EOF'
logging:
  level: debug
  format: text
  output: /tmp/app.log
  add_source: true
  service_name: test-service
  environment: development
  enable_rotation: true
  max_size: 50
  max_age: 7
  max_backups: 3
EOF

./server --config test-config.yaml --show-config
```
**Expected**: Logging values come from "config file"

### Test Environment Variables
```bash
# Test environment variable override
export TASKMANAGER_LOGGING_LEVEL=debug
export TASKMANAGER_LOGGING_FORMAT=text
./server --show-config
```
**Expected**: Level and format show "(env)" as source

### Test Command-Line Flags
```bash
# Test flag override
./server --log-level=debug --log-format=text --show-config
```
**Expected**: Level and format show "(flag)" as source

### Test Validation
```bash
# Test invalid log level
./server --log-level=invalid
```
**Expected**: Error message about invalid log level

```bash
# Test invalid format
./server --log-format=xml
```
**Expected**: Error message about invalid format

```bash
# Test empty service name
cat > invalid-config.yaml << 'EOF'
logging:
  service_name: ""
EOF

./server --config invalid-config.yaml
```
**Expected**: Error message about empty service name

### Test Conversion to Logger Config
```bash
# Create a simple test program
cat > test_conversion.go << 'EOF'
package main

import (
    "fmt"
    "yourproject/cmd/server/config"
    "yourproject/logger"
)

func main() {
    cfg, _, err := config.LoadConfig()
    if err != nil {
        panic(err)
    }
    
    loggerCfg := cfg.LogConfig.ToLoggerConfig()
    fmt.Printf("Logger Config: %+v\n", loggerCfg)
    
    // Try creating a logger
    log, err := logger.New(loggerCfg)
    if err != nil {
        panic(err)
    }
    
    log.Info("Test message from converted config")
}
EOF

go run test_conversion.go
```
**Expected**: Logger is created successfully and logs a test message

## Common Pitfalls

### Pitfall 1: Forgetting mapstructure tags
**Symptom**: Config file values don't load, always using defaults
**Fix**: Ensure every LogConfig field has a `mapstructure:"field_name"` tag

### Pitfall 2: Mismatched flag names and config keys
**Symptom**: Flags don't override config file values
**Fix**: Ensure BindPFlag uses the correct config key (e.g., "logging.level") and flag name (e.g., "log-level")

### Pitfall 3: Not converting to lowercase in validation
**Symptom**: "INFO" is rejected but "info" works
**Fix**: Use `strings.ToLower()` before comparing level and format values

### Pitfall 4: Forgetting to add keys to getSource flagMap
**Symptom**: ShowConfig shows wrong source for logging values
**Fix**: Add all logging config keys to the flagMap in getSource()

### Pitfall 5: Not handling empty strings in validation
**Symptom**: Empty service name or environment causes runtime errors
**Fix**: Check for empty strings in Validate() and return clear error messages

### Pitfall 6: Validating rotation settings when rotation is disabled
**Symptom**: Validation fails even though rotation isn't being used
**Fix**: Only validate rotation settings when EnableRotation is true

## Learning Resources

### Essential Reading
- [Viper Documentation](https://github.com/spf13/viper) - Configuration management library
- [Pflag Documentation](https://github.com/spf13/pflag) - POSIX/GNU-style command-line flags
- [Go Struct Tags](https://go.dev/wiki/Well-known-struct-tags) - Understanding struct tags

### Additional Resources
- [12-Factor App: Config](https://12factor.net/config) - Best practices for application configuration
- [Configuration Precedence Patterns](https://blog.gopheracademy.com/advent-2014/configuration-with-fangs/) - Understanding config hierarchies

## Real-World Context

### Why This Configuration Structure?

The three-tier precedence (flags > env > config file > defaults) is an industry standard because:

1. **Flags**: Highest precedence for quick testing and overrides
2. **Environment Variables**: Standard in containerized environments (Docker, Kubernetes)
3. **Config Files**: Best for complex, persistent configuration
4. **Defaults**: Ensure application works out of the box

### Environment Variable Naming

The `TASKMANAGER_LOGGING_LEVEL` format follows conventions:
- Prefix prevents collisions with other applications
- Uppercase is standard for environment variables
- Underscores replace dots (dots aren't valid in env var names)

### Why Validate Configuration?

Validation at startup prevents runtime errors and provides clear error messages. It's better to fail fast with a clear message than to start the application and fail mysteriously later.

## Testing Your Implementation

After implementing, verify:

1. **Compilation**: `go build ./cmd/server` succeeds
2. **Defaults work**: Run without config shows sensible defaults
3. **Config file works**: Values from YAML are loaded correctly
4. **Environment variables work**: Env vars override config file
5. **Flags work**: Flags override everything
6. **Validation works**: Invalid values are rejected with clear errors
7. **ShowConfig works**: Displays logging config with correct sources
8. **Conversion works**: ToLoggerConfig() creates valid logger.Config

## Next Steps

After completing this task, you'll have a complete configuration system for logging. Task 5 will use this configuration to create logger instances in the server's main.go file.

The combination of Tasks 1-4 gives you:
- Task 1: Logger factory and core functionality
- Task 2: Standard fields and masking functions
- Task 3: Context management for request correlation
- Task 4: Configuration system integration
- Task 5: Will tie everything together in the server initialization
