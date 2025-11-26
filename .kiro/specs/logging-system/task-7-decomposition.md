# Task Decomposition: Integrate logger into server main

## Overview

This task replaces the current `fmt.Println` startup messages in `cmd/server/main.go` with structured logging using the logger instance that's already been created. The logger configuration loading and middleware integration are already complete, so this task focuses solely on adding structured log entries for server lifecycle events (startup, initialization, shutdown).

## Complexity Check

**Requirements need**: Log server startup events with structured fields for observability (Requirements 1.1, 1.2, 1.3, 2.1, 2.2, 3.1, 3.2, 10.1, 10.2, 11.3)

**Current state**: 
- Logger is already created and configured ✅
- LoggingMiddleware is already applied to all routes ✅
- Startup messages use `fmt.Println` (unstructured) ❌

**Simple approach**: Replace `fmt.Println` with `logger.Info()` calls (5-10 min)

**Complex approach**: Create startup event types, structured event logging system (30+ min)

**Recommendation**: Use the simple approach. The requirements only need structured logging of startup events, not a complex event system. The logger is already configured and working.

## Prerequisites

**Existing Code:**
- `cmd/server/main.go` - Server initialization and startup
- `logger/logger.go` - Logger factory (already implemented)
- `logger/config.go` - Logger configuration (already implemented)
- `logger/middleware.go` - HTTP logging middleware (already implemented)
- `cmd/server/config/config.go` - Configuration loading (already implemented)

**Dependencies:**
- `log/slog` package (standard library, already in use)

**Knowledge Required:**
- Understanding of slog structured logging API
- Familiarity with the existing server startup flow
- Understanding of log levels (INFO for normal events, ERROR for failures)

## Step-by-Step Instructions

### Step 1: Add structured logging for logger initialization

**File**: `cmd/server/main.go`

**What to do:**
Add a log entry immediately after the logger is successfully created to confirm logging is operational.

**What to implement:**
- After `logger.NewLogger(&cfg.LogConfig)` succeeds, add a logger call
- Use INFO level
- Include fields about the logger configuration:
  - Log level
  - Log format (JSON or text)
  - Output destination
  - Service name
- This is the first "real" structured log entry that will appear

**Why:**
This confirms the logging system is working correctly and shows what logging configuration is active. It's especially useful when debugging why logs aren't appearing as expected.

**Expected result:**
The first structured log entry appears showing the logger is initialized with its configuration.

---

### Step 2: Replace database initialization message with structured logging

**File**: `cmd/server/main.go`

**What to do:**
Replace the `fmt.Println("🚀 Database storage initialized")` line with a structured log entry.

**What to implement:**
- Remove the `fmt.Println` call
- Add a `logger.Info()` call instead
- Include structured fields:
  - Database path
  - Message indicating successful initialization
- Use the same INFO level since this is a normal startup event

**Why:**
Database initialization is a critical step. Structured logging allows operators to see the database path being used and correlate database issues with the configuration. The emoji is nice for console output but doesn't work well in log aggregation tools.

**Expected result:**
Database initialization is logged with structured fields instead of a plain text message.

---

### Step 3: Replace authentication system initialization message with structured logging

**File**: `cmd/server/main.go`

**What to do:**
Replace the `fmt.Println("🔐 Authentication system initialized")` line with a structured log entry.

**What to implement:**
- Remove the `fmt.Println` call
- Add a `logger.Info()` call instead
- Include structured fields:
  - JWT expiration duration
  - Message indicating successful initialization
- Do NOT log the JWT secret (security risk)
- Use INFO level

**Why:**
Authentication system initialization is critical for security. Logging it confirms the auth system is ready and shows the token expiration configuration, which is useful for troubleshooting authentication issues.

**Expected result:**
Auth system initialization is logged with structured fields showing the JWT configuration (except the secret).

---

### Step 4: Replace server startup messages with structured logging

**File**: `cmd/server/main.go`

**What to do:**
Replace the multiple `fmt.Printf` and `fmt.Println` calls that show the server address and endpoints with structured log entries.

**What to implement:**
- Remove all the `fmt.Printf` and `fmt.Println` calls that show:
  - "🚀 HTTP Server starting on..."
  - "Endpoints:" list
- Add a single `logger.Info()` call for server startup
- Include structured fields:
  - Server address (host:port)
  - List of available endpoints (as a slice or comma-separated string)
  - Environment name
  - Service version (if available)
- Use INFO level

**Why:**
Server startup is the final initialization step. Structured logging makes it easy to search logs for "when did the server start" and "what endpoints are available". The endpoint list is useful for documentation and troubleshooting.

**Expected result:**
Server startup is logged once with all relevant information in structured fields, replacing multiple console print statements.

---

### Step 5: Add structured logging for fatal errors

**File**: `cmd/server/main.go`

**What to do:**
Replace the `log.Fatal()` calls with structured logging before exiting.

**What to implement:**
- Before each `log.Fatal()` call, add a `logger.Error()` call
- Include structured fields with error context:
  - Error message
  - What operation failed (config loading, logger creation, database init)
  - Any relevant parameters
- Keep the `log.Fatal()` or `os.Exit(1)` call after logging
- Use ERROR level

**Why:**
Fatal errors need to be logged with full context before the application exits. This ensures the error is captured in log aggregation systems and provides enough information for troubleshooting. Using structured logging makes it easier to alert on specific failure types.

**Expected result:**
Fatal errors are logged with structured context before the application exits.

## Verification

### Compile Check
```bash
go build ./cmd/server
```
**Expected**: No compilation errors

### Run Server and Check Logs
```bash
# Run with text format for easy reading
./server --log-format=text --log-level=info

# Or with JSON format for structured output
./server --log-format=json --log-level=info
```

**Expected output (text format)**:
```
2024-11-10T10:30:00.000Z INFO Logger initialized service=task-manager-api environment=production level=info format=json output=stdout
2024-11-10T10:30:00.001Z INFO Database initialized service=task-manager-api environment=production database_path=./data/tasks.db
2024-11-10T10:30:00.002Z INFO Authentication system initialized service=task-manager-api environment=production jwt_expiration=24h
2024-11-10T10:30:00.003Z INFO Server starting service=task-manager-api environment=production address=0.0.0.0:8080 endpoints=[/health,/tasks,/register,/login]
```

**Expected output (JSON format)**:
```json
{"timestamp":"2024-11-10T10:30:00.000Z","level":"INFO","message":"Logger initialized","service":"task-manager-api","environment":"production","level":"info","format":"json","output":"stdout"}
{"timestamp":"2024-11-10T10:30:00.001Z","level":"INFO","message":"Database initialized","service":"task-manager-api","environment":"production","database_path":"./data/tasks.db"}
{"timestamp":"2024-11-10T10:30:00.002Z","level":"INFO","message":"Authentication system initialized","service":"task-manager-api","environment":"production","jwt_expiration":"24h"}
{"timestamp":"2024-11-10T10:30:00.003Z","level":"INFO","message":"Server starting","service":"task-manager-api","environment":"production","address":"0.0.0.0:8080","endpoints":["GET /health","POST /tasks","POST /register","POST /login"]}
```

### Test with Different Log Levels
```bash
# Test with DEBUG level (should see all logs)
./server --log-level=debug

# Test with ERROR level (should only see errors, no startup logs)
./server --log-level=error
```

**Expected**: Log output respects the configured log level.

### Test Error Logging
```bash
# Test with invalid config to trigger error logging
./server --jwt-secret=short

# Or with invalid database path
./server --db-path=/invalid/path/tasks.db
```

**Expected**: Structured error logs appear before the application exits.

### Manual Testing
1. Start the server: `go run cmd/server/main.go`
2. Check that startup logs appear in structured format
3. Make an HTTP request: `curl http://localhost:8080/health`
4. Verify that HTTP request logs appear (from LoggingMiddleware)
5. Stop the server with Ctrl+C
6. Verify logs are properly formatted and contain all expected fields

## Common Pitfalls

### Pitfall 1: Forgetting to check for nil logger
**Symptom**: Panic or nil pointer dereference when using the logger
**Fix**: Always check that logger creation succeeded before using it. The error handling after `logger.NewLogger()` should prevent this, but be careful not to use the logger if creation failed.

### Pitfall 2: Logging sensitive information
**Symptom**: JWT secret or passwords appear in logs
**Fix**: Never log the JWT secret, passwords, or tokens. Use the `logger.MaskToken()` function if you need to log token-related information. Only log non-sensitive configuration like expiration duration.

### Pitfall 3: Too many log fields
**Symptom**: Log entries are cluttered with unnecessary fields
**Fix**: Only include fields that are useful for troubleshooting or monitoring. For example, don't log every single configuration value - focus on the most important ones like addresses, paths, and durations.

### Pitfall 4: Using fmt.Println instead of logger
**Symptom**: Some startup messages still appear as plain text instead of structured logs
**Fix**: Search for all `fmt.Println` and `fmt.Printf` calls in main.go and replace them with appropriate `logger.Info()` or `logger.Error()` calls. The only exception might be the `--show-config` output which is meant for human reading.

## Learning Resources

### Essential Reading
- [Go slog Package Documentation](https://pkg.go.dev/log/slog) - Official documentation for structured logging
- [Structured Logging Best Practices](https://www.honeycomb.io/blog/structured-logging-and-your-team) - Why structured logging matters
- [The Twelve-Factor App: Logs](https://12factor.net/logs) - Best practices for application logging

### Additional Resources (Optional)
- [Effective Go: Logging](https://go.dev/doc/effective_go#logging) - Go's approach to logging
- [OpenTelemetry Logging](https://opentelemetry.io/docs/specs/otel/logs/) - Standards for observability
