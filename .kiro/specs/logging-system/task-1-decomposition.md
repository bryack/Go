# Task Decomposition: Create Logger Package Foundation

## Overview

This task establishes the foundational logger package that will be used throughout the application. You'll create the core configuration types, implement a factory function that creates properly configured slog.Logger instances, and provide helper functions for common use cases. This foundation supports both JSON and text output formats, multiple output destinations (stdout, stderr, files), and configurable log levels.

## Implementation Approach

We're building a factory pattern around Go's standard library `log/slog` package. The factory approach allows us to centralize logger configuration and create consistent logger instances across the application. This is preferable to directly instantiating slog.Logger everywhere because it:

1. Provides a single point of configuration
2. Ensures consistent default settings
3. Makes testing easier with dedicated test logger creation
4. Allows runtime configuration changes

The implementation uses slog's handler pattern where different handlers (TextHandler, JSONHandler) format the output differently. We'll select the appropriate handler based on configuration and wrap it with level filtering.

**Key Concepts:**
- **Handler Pattern**: slog uses handlers to determine output format and destination
- **Log Levels**: Hierarchical filtering (DEBUG < INFO < WARN < ERROR)
- **Factory Pattern**: Centralized creation of configured objects
- **Dependency Injection**: Logger instances passed to components that need them

## Prerequisites

**Existing Code:**
- None - this is a new package

**Dependencies:**
- Go 1.21+ (for `log/slog` standard library package)
- No external dependencies needed for core functionality

**Knowledge Required:**
- Understanding of Go package structure and organization
- Familiarity with Go's `log/slog` package basics
- Understanding of log levels (DEBUG, INFO, WARN, ERROR)
- Basic understanding of the factory pattern

## Step-by-Step Instructions

### Step 1: Create package directory and config types

**File**: `logger/config.go` (create new file)

**What to do:**
Create the configuration structure that will control logger behavior.

**What to implement:**
- Define a `Config` struct with the following fields:
  - `Level` (string) - log level: "debug", "info", "warn", or "error"
  - `Format` (string) - output format: "json" or "text"
  - `Output` (string) - output destination: "stdout", "stderr", or a file path
  - `AddSource` (bool) - whether to include source file and line number in logs
  - `ServiceName` (string) - identifier for the service (e.g., "task-manager-api")
  - `Environment` (string) - deployment environment: "development", "production", "staging"
- Add struct tags for mapstructure if you want viper compatibility later
- Consider adding comments explaining each field's purpose and valid values

**Why:**
The Config struct is the contract between your application's configuration system and the logger. It defines all the knobs you can turn to control logging behavior. Separating this into its own file keeps concerns organized.

**Expected result:**
You have a `logger/config.go` file with a well-documented Config struct. The file compiles without errors.

---

### Step 2: Implement log level parsing

**File**: `logger/config.go` (modify existing)

**What to do:**
Add a function that converts string log levels to slog.Level constants.

**What to implement:**
- Create a function that accepts a string (e.g., "debug", "info") and returns `slog.Level`
- Handle case-insensitive input (convert to lowercase)
- Map strings to slog constants:
  - "debug" → `slog.LevelDebug`
  - "info" → `slog.LevelInfo`
  - "warn" → `slog.LevelWarn`
  - "error" → `slog.LevelError`
- Return a default level (INFO) for invalid input
- Consider returning an error for invalid levels instead of silently defaulting

**Why:**
Configuration typically comes from files or environment variables as strings, but slog needs typed Level constants. This function bridges that gap and provides validation.

**Expected result:**
You can convert string log levels to slog.Level. Invalid inputs are handled gracefully.

---

### Step 3: Implement output destination handler

**File**: `logger/logger.go` (create new file)

**What to do:**
Create a function that determines where logs should be written based on configuration.

**What to implement:**
- Create a function that accepts the `Output` string from Config
- Return an `io.Writer` based on the output value:
  - "stdout" → return `os.Stdout`
  - "stderr" → return `os.Stderr`
  - Any other value → treat as file path, open/create the file
- For file paths, open the file in append mode with appropriate permissions (0644)
- Create parent directories if they don't exist
- Return an error if file creation fails
- Consider whether to buffer the writer for performance

**Why:**
Different environments need logs in different places. Development might use stdout, production might use files, and containers often use stderr. This abstraction lets configuration control the destination.

**Expected result:**
You can get an io.Writer for any valid output destination. Files are created if needed, and errors are returned for invalid paths.

---

### Step 4: Implement handler selection logic

**File**: `logger/logger.go` (modify existing)

**What to do:**
Create a function that selects and configures the appropriate slog handler based on format.

**What to implement:**
- Create a function that accepts Config and an io.Writer
- Return a `slog.Handler` based on the Format field:
  - "json" → create `slog.NewJSONHandler` with the writer
  - "text" → create `slog.NewTextHandler` with the writer
- Configure handler options:
  - Set the log level using `slog.HandlerOptions.Level`
  - Set `AddSource` based on config (shows file:line for logs)
  - Consider adding timestamp formatting options
- Default to JSON format if format is invalid
- The handler options should be created from the Config values

**Why:**
Different environments benefit from different formats. JSON is machine-readable for log aggregation tools, while text is human-readable for development. The handler encapsulates all formatting logic.

**Expected result:**
You can create a properly configured slog.Handler for either JSON or text format with the correct log level and source settings.

---

### Step 5: Implement the main logger factory function

**File**: `logger/logger.go` (modify existing)

**What to do:**
Create the primary factory function that ties everything together.

**What to implement:**
- Create a `New` function that accepts a Config struct
- Return `(*slog.Logger, error)`
- Inside the function:
  - Call your output destination function to get the io.Writer
  - Call your handler selection function to get the slog.Handler
  - Create a new slog.Logger using `slog.New(handler)`
  - Add default attributes to the logger (service name, environment)
  - Return the logger and any errors encountered
- Handle all errors from helper functions appropriately
- Consider validating the Config before processing

**Why:**
This is the main entry point for creating loggers. It orchestrates all the configuration steps and returns a ready-to-use logger. Centralizing this logic ensures consistency across the application.

**Expected result:**
You can call `logger.New(config)` and get a fully configured slog.Logger with the specified format, level, output destination, and default fields. Errors are returned for invalid configurations.

---

### Step 6: Implement NewDefault helper function

**File**: `logger/logger.go` (modify existing)

**What to do:**
Create a convenience function for quick logger creation with sensible defaults.

**What to implement:**
- Create a `NewDefault` function that takes no parameters
- Return `*slog.Logger` (no error since defaults are always valid)
- Inside the function:
  - Create a Config with default values:
    - Level: "info"
    - Format: "text"
    - Output: "stdout"
    - AddSource: false
    - ServiceName: "app" or similar generic name
    - Environment: "development"
  - Call your `New` function with this config
  - Since defaults are valid, you can ignore/panic on errors
- This should be a one-liner or very simple function

**Why:**
During development or testing, you often just want a logger quickly without configuring everything. This provides a fast path while still using the same underlying factory logic.

**Expected result:**
You can call `logger.NewDefault()` and immediately get a working logger with reasonable defaults for development use.

---

### Step 7: Add default attributes to loggers

**File**: `logger/logger.go` (modify existing)

**What to do:**
Enhance the logger factory to include default attributes in all log entries.

**What to implement:**
- In your `New` function, after creating the base logger:
  - Use `logger.With()` to add default attributes
  - Add service name from config as an attribute
  - Add environment from config as an attribute
  - Consider adding a version field (can be hardcoded for now)
- These attributes will appear in every log entry automatically
- Use appropriate slog attribute types (slog.String, etc.)

**Why:**
When logs are aggregated from multiple services, you need to identify which service and environment they came from. Adding these as default attributes ensures they're always present without manual inclusion in every log call.

**Expected result:**
Every log entry from your logger includes service name and environment fields automatically. You can verify this by creating a logger and logging a test message.

---

### Step 8: Add package-level documentation

**File**: `logger/logger.go` (modify existing)

**What to do:**
Add comprehensive package documentation at the top of the file.

**What to implement:**
- Add a package comment block explaining:
  - What the logger package does
  - How to create a logger (basic example)
  - The different configuration options available
  - When to use New vs NewDefault
- Follow Go documentation conventions (package comment before package declaration)
- Include a simple usage example in the comment
- Keep it concise but informative

**Why:**
Good documentation helps other developers (and future you) understand how to use the package. Package-level docs appear in godoc and IDE tooltips.

**Expected result:**
Running `go doc logger` shows helpful documentation. The package purpose and basic usage are clear.

---

### Step 9: Add validation to Config

**File**: `logger/config.go` (modify existing)

**What to do:**
Add a validation method to ensure Config values are valid before use.

**What to implement:**
- Create a `Validate()` method on the Config struct
- Return an error if validation fails
- Check that:
  - Level is one of: "debug", "info", "warn", "error" (case-insensitive)
  - Format is one of: "json", "text"
  - Output is not empty
  - ServiceName is not empty (required for log aggregation)
- Use descriptive error messages that explain what's wrong
- Consider using `errors.Join()` to return multiple validation errors at once

**Why:**
Catching configuration errors early prevents runtime surprises. Clear validation errors help users fix their configuration quickly.

**Expected result:**
Invalid configurations are rejected with clear error messages. The `New` function can call `Validate()` before proceeding.

---

### Step 10: Implement basic error handling

**File**: `logger/logger.go` (modify existing)

**What to do:**
Ensure all error paths are properly handled throughout the package.

**What to implement:**
- In the `New` function, check for errors from:
  - Config validation
  - Output destination creation (file opening)
  - Any other operations that can fail
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Return errors to the caller rather than panicking
- Consider what should happen if logger creation fails (fallback logger?)
- Ensure file handles are closed on error paths if needed

**Why:**
Robust error handling prevents crashes and provides useful debugging information. Wrapped errors preserve the error chain for better diagnostics.

**Expected result:**
All error paths are handled. Errors include context about what operation failed. No resource leaks on error paths.

---

## Verification

### Compile Check
```bash
go build ./logger
```
**Expected**: No compilation errors. The package builds successfully.

### Basic Functionality Test
Create a simple test file or main.go to verify:

```bash
# Create a test file
cat > test_logger.go << 'EOF'
package main

import "yourproject/logger"

func main() {
    // Test default logger
    log := logger.NewDefault()
    log.Info("Testing default logger")
    
    // Test configured logger
    cfg := logger.Config{
        Level:       "debug",
        Format:      "json",
        Output:      "stdout",
        ServiceName: "test-service",
        Environment: "development",
    }
    log2, err := logger.New(cfg)
    if err != nil {
        panic(err)
    }
    log2.Debug("Testing configured logger")
    log2.Info("This is an info message")
}
EOF

go run test_logger.go
```

**Expected output**: 
- Default logger outputs text format to stdout
- Configured logger outputs JSON format with debug and info messages
- JSON output includes service name and environment fields

### Validation Test
Test that invalid configurations are rejected:

```bash
# Test with invalid level
cfg := logger.Config{
    Level:       "invalid",
    Format:      "json",
    Output:      "stdout",
    ServiceName: "test",
    Environment: "dev",
}
_, err := logger.New(cfg)
// Should return validation error
```

**Expected**: Validation errors are returned for invalid configurations.

### File Output Test
Test that file output works:

```bash
# Test file output
cfg := logger.Config{
    Level:       "info",
    Format:      "json",
    Output:      "/tmp/test.log",
    ServiceName: "test",
    Environment: "dev",
}
log, _ := logger.New(cfg)
log.Info("Test message")

# Check file was created
cat /tmp/test.log
```

**Expected**: Log file is created at specified path with JSON content.

## Common Pitfalls

### Pitfall 1: Forgetting to handle case-insensitive log levels
**Symptom**: Config with "INFO" (uppercase) doesn't work, only "info" works
**Fix**: Convert level strings to lowercase before comparison: `strings.ToLower(level)`

### Pitfall 2: Not creating parent directories for log files
**Symptom**: Error "no such file or directory" when output path is "/var/log/app/app.log" and /var/log/app doesn't exist
**Fix**: Use `os.MkdirAll(filepath.Dir(path), 0755)` before opening the file

### Pitfall 3: File handle leaks on error paths
**Symptom**: Too many open files error after repeated failures
**Fix**: Use `defer file.Close()` immediately after successfully opening a file, or ensure cleanup in error paths

### Pitfall 4: Not validating Config before use
**Symptom**: Cryptic errors or panics when invalid config is used
**Fix**: Call `config.Validate()` at the start of the `New` function and return early if validation fails

### Pitfall 5: Forgetting to add default attributes
**Symptom**: Service name and environment don't appear in logs
**Fix**: Use `logger.With(slog.String("service", cfg.ServiceName), ...)` to add default attributes after creating the base logger

## Learning Resources

### Essential Reading
- [Go slog Package Documentation](https://pkg.go.dev/log/slog) - Official documentation with examples and API reference
- [Go Blog: Structured Logging with slog](https://go.dev/blog/slog) - Official introduction to slog with design rationale and patterns
- [Effective Go: Errors](https://go.dev/doc/effective_go#errors) - Best practices for error handling in Go

### Additional Resources (Optional)
- [slog Handler Guide](https://github.com/golang/example/tree/master/slog-handler-guide) - Deep dive into creating custom handlers
- [Factory Pattern in Go](https://refactoring.guru/design-patterns/factory-method/go/example) - Understanding the factory pattern
- [Go io.Writer Interface](https://pkg.go.dev/io#Writer) - Understanding the Writer interface used for output

## Next Steps

After completing this task, you'll have a solid logger package foundation. The next task (Task 2) will build on this by adding standard field definitions and helper functions for common logging patterns like masking sensitive data.

You can test your implementation by:
1. Creating loggers with different configurations
2. Verifying output format (JSON vs text)
3. Checking that log levels filter correctly
4. Confirming default attributes appear in all logs
5. Testing file output and error handling
