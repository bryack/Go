# Task Decomposition: Add Shutdown Timeout Configuration

## Overview

Add a configurable shutdown timeout to the server configuration system. This allows operators to control how long the server waits for in-flight requests to complete during graceful shutdown. The implementation follows the existing configuration patterns used for JWT expiration and logging settings.

## Implementation Approach

We're extending the existing Viper-based configuration system by adding a new field to `ServerConfig`. This approach reuses the established patterns for duration configuration (similar to JWT expiration) and follows the same precedence chain: flags > environment variables > config file > defaults.

**Complexity Check:**
- **Requirements need**: Configurable shutdown timeout with 30-second default, validation, and logging
- **Simple approach**: Add field to existing struct, reuse Viper's duration parsing (5-10 min)
- **Complex approach**: Create separate shutdown config struct with multiple timeout options (20+ min)
- **Recommendation**: Simple approach. Requirements only need one timeout value. The existing `ServerConfig` struct is the natural home for this setting since it controls server behavior.

**Key Concepts:**
- **time.Duration**: Go's built-in type for representing time spans, automatically parsed by Viper from strings like "30s" or "1m"
- **Configuration precedence**: Viper's layered config system where flags override env vars, which override config files, which override defaults
- **Validation**: Ensuring the timeout is positive to prevent invalid configurations

## Prerequisites

**Existing Code:**
- `cmd/server/config/config.go` - Configuration loading and validation logic
- `config.yaml` - Example configuration file
- `cmd/server/main.go` - Server initialization where config is loaded

**Dependencies:**
- `github.com/spf13/viper` - Already installed for configuration management
- `github.com/spf13/pflag` - Already installed for command-line flags
- `time` package - Standard library

**Knowledge Required:**
- Understanding of Go's `time.Duration` type
- Familiarity with Viper configuration patterns (already used in the codebase)
- Basic YAML syntax

## Step-by-Step Instructions

### Step 1: Add ShutdownTimeout field to ServerConfig struct

**File**: `cmd/server/config/config.go`

**What to do:**
Add a new field to the `ServerConfig` struct to store the shutdown timeout duration.

**What to implement:**
- Locate the `ServerConfig` struct (around line 26)
- Add a new field named `ShutdownTimeout` of type `time.Duration`
- Add the mapstructure tag `mapstructure:"shutdown_timeout"` to enable YAML binding
- Place it after the `Host` field for logical grouping

**Why:**
The `ServerConfig` struct defines server-level settings. Shutdown timeout is a server behavior setting, so it belongs here. The mapstructure tag tells Viper how to map the YAML key to the struct field.

**Expected result:**
The `ServerConfig` struct now has three fields: `Port`, `Host`, and `ShutdownTimeout`. Code compiles without errors.

---

### Step 2: Set default value for shutdown timeout

**File**: `cmd/server/config/config.go`

**What to do:**
Add a default value for the shutdown timeout in the `LoadConfig` function.

**What to implement:**
- Locate the `LoadConfig` function where defaults are set (around line 42)
- Add a new line: `v.SetDefault("server.shutdown_timeout", "30s")`
- Place it after the other server defaults (`server.port` and `server.host`)
- Use string format "30s" which Viper will automatically parse to `time.Duration`

**Why:**
Setting defaults ensures the configuration always has a valid value even if not specified. The 30-second default meets requirement 4.2 and provides a reasonable balance between allowing requests to complete and not delaying shutdowns excessively.

**Expected result:**
If no shutdown timeout is configured, the system will use 30 seconds. Code compiles without errors.

---

### Step 3: Add command-line flag for shutdown timeout

**File**: `cmd/server/config/config.go`

**What to do:**
Add a command-line flag to allow setting shutdown timeout via CLI.

**What to implement:**
- Locate where flags are defined in `LoadConfig` (around line 60)
- Add a new flag definition: `pflag.String("shutdown-timeout", "30s", "Graceful shutdown timeout")`
- Place it after the server-related flags (`--port`, `--host`)
- Use kebab-case naming convention consistent with other flags

**Why:**
Command-line flags provide the highest precedence in the configuration system, allowing operators to override settings at runtime without modifying files. This is useful for testing different timeout values.

**Expected result:**
Users can run `go run cmd/server/main.go --shutdown-timeout=45s` to set a custom timeout. Code compiles without errors.

---

### Step 4: Bind flag to configuration key

**File**: `cmd/server/config/config.go`

**What to do:**
Connect the command-line flag to the Viper configuration system.

**What to implement:**
- Locate the flag binding section in `LoadConfig` (around line 105)
- Add a new binding: `v.BindPFlag("server.shutdown_timeout", pflag.Lookup("shutdown-timeout"))`
- Place it after the other server flag bindings
- Ensure the first parameter matches the mapstructure path and the second matches the flag name

**Why:**
Binding connects the flag to Viper's configuration tree, enabling the precedence system to work correctly. Without this, the flag would be ignored.

**Expected result:**
The flag value is now accessible through Viper and will be unmarshaled into the struct. Code compiles without errors.

---

### Step 5: Add validation for shutdown timeout

**File**: `cmd/server/config/config.go`

**What to do:**
Add validation logic to ensure the shutdown timeout is positive.

**What to implement:**
- Locate the `Validate` method on the `Config` struct (around line 125)
- Add a validation check after the server port validation
- Check if `config.ServerConfig.ShutdownTimeout <= 0`
- If invalid, append an error: `fmt.Errorf("server.shutdown_timeout must be positive, got %v", config.ServerConfig.ShutdownTimeout)`
- Use the same error collection pattern as other validations

**Why:**
Validation prevents invalid configurations from causing runtime issues. A zero or negative timeout would break the graceful shutdown logic. Requirement 4.3 explicitly requires this validation.

**Expected result:**
Starting the server with `--shutdown-timeout=0s` or `--shutdown-timeout=-5s` will fail with a clear error message. Valid positive durations are accepted.

---

### Step 6: Update ShowConfig to display shutdown timeout

**File**: `cmd/server/config/config.go`

**What to do:**
Add the shutdown timeout to the configuration display output.

**What to implement:**
- Locate the `ShowConfig` function (around line 230)
- Add a new printf line after the server port display
- Format: `fmt.Printf("server.shutdown_timeout: %s (%s)\n", cfg.ServerConfig.ShutdownTimeout, getSource(v, "server.shutdown_timeout"))`
- This shows both the value and where it came from (flag/env/file/default)

**Why:**
The `--show-config` flag is used for debugging configuration issues. Including shutdown timeout helps operators verify their settings are correct.

**Expected result:**
Running `go run cmd/server/main.go --show-config` displays the shutdown timeout value and its source.

---

### Step 7: Update getSource function to support shutdown timeout flag

**File**: `cmd/server/config/config.go`

**What to do:**
Add the shutdown timeout flag mapping to the source detection function.

**What to implement:**
- Locate the `getSource` function (around line 195)
- Find the `flagMap` variable
- Add a new entry: `"server.shutdown_timeout": "shutdown-timeout"`
- Place it after the other server entries for consistency

**Why:**
The `getSource` function determines whether a config value came from a flag, env var, config file, or default. Adding this mapping ensures `--show-config` correctly identifies the source.

**Expected result:**
When using `--show-config`, the shutdown timeout line shows the correct source (e.g., "flag" when set via CLI).

---

### Step 8: Add shutdown timeout to example config.yaml

**File**: `config.yaml`

**What to do:**
Document the new configuration option in the example config file.

**What to implement:**
- Locate the `server:` section (around line 4)
- Add a new line after `port: 8080`
- Add: `shutdown_timeout: "30s"` with a comment explaining the setting
- Include a comment describing what the timeout controls and typical values
- Mention that it can be specified as "30s", "1m", "90s", etc.

**Why:**
The config.yaml file serves as documentation for operators. Including the new setting with clear comments helps users understand what it does and how to configure it.

**Expected result:**
Users examining config.yaml see the shutdown_timeout option with helpful documentation.

---

### Step 9: Log shutdown timeout at server startup

**File**: `cmd/server/main.go`

**What to do:**
Add a log entry showing the configured shutdown timeout when the server starts.

**What to implement:**
- Locate the server initialization logging section (around line 430)
- After the "HTTP Server initialized" log entry, add a new log line
- Use structured logging: `l.Info("Graceful shutdown configured", slog.Duration("shutdown_timeout", cfg.ServerConfig.ShutdownTimeout))`
- Place it before the `http.ListenAndServe` call

**Why:**
Requirement 4.4 explicitly requires logging the configured timeout at startup. This helps operators verify the setting is correct and aids in troubleshooting shutdown behavior.

**Expected result:**
Server startup logs include a line showing the shutdown timeout value. The server starts successfully.

## Verification

### Compile Check
```bash
go build ./cmd/server
```
**Expected**: No compilation errors

### Configuration Loading Test
```bash
# Test default value
go run cmd/server/main.go --show-config | grep shutdown_timeout
# Expected output: server.shutdown_timeout: 30s (default)

# Test flag override
go run cmd/server/main.go --shutdown-timeout=45s --show-config | grep shutdown_timeout
# Expected output: server.shutdown_timeout: 45s (flag)

# Test environment variable
export TASKMANAGER_SERVER_SHUTDOWN_TIMEOUT=60s
go run cmd/server/main.go --show-config | grep shutdown_timeout
# Expected output: server.shutdown_timeout: 1m0s (env)
unset TASKMANAGER_SERVER_SHUTDOWN_TIMEOUT
```

### Validation Test
```bash
# Test invalid timeout (should fail)
go run cmd/server/main.go --shutdown-timeout=0s
# Expected: Error message about shutdown_timeout must be positive

go run cmd/server/main.go --shutdown-timeout=-5s
# Expected: Error message about shutdown_timeout must be positive
```

### YAML Configuration Test
1. Edit `config.yaml` and add `shutdown_timeout: "45s"` under the `server:` section
2. Run: `go run cmd/server/main.go --show-config | grep shutdown_timeout`
3. **Expected output**: `server.shutdown_timeout: 45s (config file)`

### Startup Logging Test
```bash
go run cmd/server/main.go
# Expected: Log output includes "Graceful shutdown configured" with shutdown_timeout value
```

## Common Pitfalls

### Pitfall 1: Forgetting the mapstructure tag
**Symptom**: Config file values are ignored, always uses default
**Fix**: Ensure the struct field has `` `mapstructure:"shutdown_timeout"` `` tag

### Pitfall 2: Mismatched key names
**Symptom**: Flag or env var doesn't override the value
**Fix**: Verify consistency:
- Struct tag: `shutdown_timeout` (snake_case)
- Flag name: `shutdown-timeout` (kebab-case)
- Env var: `TASKMANAGER_SERVER_SHUTDOWN_TIMEOUT` (uppercase with underscores)
- Viper key: `server.shutdown_timeout` (dot notation)

### Pitfall 3: Wrong duration format in YAML
**Symptom**: Parsing error or unexpected value
**Fix**: Use quoted strings with units: `"30s"`, `"1m"`, `"90s"`. Don't use bare numbers or missing quotes.

### Pitfall 4: Validation error message doesn't show the value
**Symptom**: Error message is unclear about what value was invalid
**Fix**: Include `%v` in the error format string to show the actual value received

## Learning Resources

### Essential Reading
- [Viper Documentation](https://github.com/spf13/viper) - Configuration management library used in this project
- [Go time.Duration](https://pkg.go.dev/time#Duration) - Understanding Go's duration type and parsing

### Additional Resources (Optional)
- [12-Factor App Config](https://12factor.net/config) - Best practices for application configuration
- [Viper Duration Parsing](https://github.com/spf13/viper#working-with-flags) - How Viper handles duration strings
