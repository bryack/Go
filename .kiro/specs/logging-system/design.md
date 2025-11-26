# Logging System Design

## Overview

This document describes the design for implementing a structured logging system using Go's standard library `log/slog` package. The system will provide configurable, structured logging with support for multiple output formats (JSON for production, text for development), integration with industry-standard log analysis tools (ELK Stack, Grafana Loki, Datadog), and comprehensive observability features including request tracing and correlation.

The design follows dependency injection principles to ensure testability and maintains compatibility with the existing codebase architecture.

## Architecture

### High-Level Components

```
┌─────────────────────────────────────────────────────────────┐
│                      Application Layer                       │
│  (HTTP Handlers, CLI Commands, Storage Operations)          │
└────────────────┬────────────────────────────────────────────┘
                 │ Uses logger instance
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                    Logger Package                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │   Logger     │  │  Middleware  │  │   Context    │      │
│  │  Factory     │  │   Wrapper    │  │   Helpers    │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└────────────────┬────────────────────────────────────────────┘
                 │ Configures
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                   slog (Standard Library)                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ TextHandler  │  │ JSONHandler  │  │ MultiHandler │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└────────────────┬────────────────────────────────────────────┘
                 │ Writes to
                 ▼
┌─────────────────────────────────────────────────────────────┐
│                      Output Destinations                     │
│     stdout  │  stderr  │  File  │  Log Rotation             │
└─────────────────────────────────────────────────────────────┘
```

### Package Structure

```
myproject/
├── logger/
│   ├── logger.go          # Logger factory and configuration
│   ├── config.go          # Logger configuration types
│   ├── middleware.go      # HTTP logging middleware
│   ├── context.go         # Context helpers for request IDs
│   ├── fields.go          # Standard field names and helpers
│   └── logger_test.go     # Test logger implementation
├── cmd/
│   ├── server/
│   │   ├── main.go        # Initialize logger, inject into handlers
│   │   └── config/
│   │       └── config.go  # Add logging configuration
│   └── cli/
│       └── main.go        # Initialize CLI logger
├── storage/
│   └── database.go        # Add logger field, log operations
├── auth/
│   ├── service.go         # Add logger field, log auth events
│   └── middleware.go      # Add logger field, log auth failures
└── internal/handlers/
    └── handlers.go        # Add logger field, log HTTP operations
```

## Components and Interfaces

### 1. Logger Configuration

**File**: `logger/config.go`

```go
type Config struct {
    Level      string        // "debug", "info", "warn", "error"
    Format     string        // "json", "text"
    Output     string        // "stdout", "stderr", or file path
    AddSource  bool          // Include source file/line for errors
    ServiceName string       // Service identifier for log aggregation
    Environment string       // "development", "production", "staging"
    
    // File output options
    EnableRotation bool
    MaxSize        int          // MB
    MaxAge         int          // days
    MaxBackups     int
}
```

**Integration**: Extends existing `cmd/server/config/config.go` with a `LogConfig` field.

### 2. Logger Factory

**File**: `logger/logger.go`

**Purpose**: Creates and configures slog.Logger instances based on configuration.

**Key Functions**:
- `New(config Config) (*slog.Logger, error)` - Creates configured logger
- `NewDefault() *slog.Logger` - Creates default logger for quick setup
- `NewTest() (*slog.Logger, *TestBuffer)` - Creates test logger with buffer

**Features**:
- Selects appropriate handler (JSON vs Text) based on config
- Sets log level from configuration
- Configures output destination (stdout, file, etc.)
- Adds default fields (service name, environment, version)
- Supports log rotation using `gopkg.in/natefinch/lumberjack.v2`

### 3. Standard Fields

**File**: `logger/fields.go`

**Purpose**: Defines standard field names for consistency across the application and compatibility with log analysis tools.

**Standard Fields**:
```go
const (
    FieldRequestID   = "request_id"
    FieldUserID      = "user_id"
    FieldMethod      = "method"
    FieldPath        = "path"
    FieldStatusCode  = "status_code"
    FieldDuration    = "duration_ms"
    FieldError       = "error"
    FieldOperation   = "operation"
    FieldTaskID      = "task_id"
    FieldEmail       = "email"        // Always masked
    FieldTraceID     = "trace_id"
    FieldSpanID      = "span_id"
)
```

**Helper Functions**:
- `MaskEmail(email string) string` - Masks email for privacy (user@example.com → u***r@example.com)
- `MaskToken(token string) string` - Masks tokens for security
- `WithRequestID(ctx context.Context, requestID string) context.Context`
- `GetRequestID(ctx context.Context) string`

### 4. HTTP Logging Middleware

**File**: `logger/middleware.go`

**Purpose**: Replaces the current `logRequest` function in `cmd/server/main.go` with structured logging.

**Key Functions**:
- `LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler`
- Generates unique request ID for each request
- Logs request start with method, path, user agent
- Logs request completion with status code, duration, bytes written
- Logs errors with full context
- Adds request ID to context for downstream logging

**Features**:
- Request ID generation using UUID or similar
- Response writer wrapper to capture status code
- Duration tracking with high precision
- User ID extraction from context (if authenticated)
- Panic recovery with stack trace logging

### 5. Context Helpers

**File**: `logger/context.go`

**Purpose**: Manage request-scoped data (request IDs, trace IDs) in context.

**Key Functions**:
```go
func WithRequestID(ctx context.Context, requestID string) context.Context
func GetRequestID(ctx context.Context) string
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context
func FromContext(ctx context.Context) *slog.Logger
func WithTraceID(ctx context.Context, traceID string) context.Context
func GetTraceID(ctx context.Context) string
```

### 6. Test Logger

**File**: `logger/logger_test.go`

**Purpose**: Provides test utilities for capturing and asserting log output.

**Key Types**:
```go
type TestBuffer struct {
    entries []LogEntry
    mu      sync.Mutex
}

type LogEntry struct {
    Level   slog.Level
    Message string
    Attrs   map[string]any
}
```

**Key Functions**:
- `NewTest() (*slog.Logger, *TestBuffer)` - Creates test logger
- `(tb *TestBuffer) Entries() []LogEntry` - Returns captured entries
- `(tb *TestBuffer) Contains(level, message string) bool` - Assertion helper
- `(tb *TestBuffer) Reset()` - Clears buffer for next test

## Data Models

### Log Entry Structure (JSON Format)

```json
{
  "timestamp": "2024-11-08T10:30:45.123Z",
  "level": "INFO",
  "message": "HTTP request completed",
  "service": "task-manager-api",
  "environment": "production",
  "request_id": "req_abc123xyz",
  "user_id": 42,
  "method": "POST",
  "path": "/tasks",
  "status_code": 201,
  "duration_ms": 45.2,
  "source": "cmd/server/main.go:123"
}
```

### Log Entry Structure (Text Format)

```
2024-11-08T10:30:45.123Z INFO HTTP request completed service=task-manager-api environment=production request_id=req_abc123xyz user_id=42 method=POST path=/tasks status_code=201 duration_ms=45.2
```

## Integration Points

### 1. Server Main (`cmd/server/main.go`)

**Changes**:
1. Load logger configuration from config file
2. Create logger instance using `logger.New()`
3. Replace `logRequest` middleware with `logger.LoggingMiddleware`
4. Inject logger into handlers, storage, and auth service
5. Log startup events (server start, database connection, etc.)
6. Log shutdown events (graceful shutdown, cleanup)

**Example**:
```go
func main() {
    cfg, v, err := config.LoadConfig()
    if err != nil {
        log.Fatal("Failed to load config: ", err)
    }
    
    // Create logger
    logger, err := logger.New(cfg.LogConfig)
    if err != nil {
        log.Fatal("Failed to create logger: ", err)
    }
    
    logger.Info("Starting Task Manager API",
        slog.String("version", "1.0.0"),
        slog.String("environment", cfg.LogConfig.Environment))
    
    // Initialize storage with logger
    s, err := storage.NewDatabaseStorage(cfg.DatabaseConfig.Path, logger)
    // ... rest of initialization
}
```

### 2. Storage Layer (`storage/database.go`)

**Changes**:
1. Add `logger *slog.Logger` field to `DatabaseStorage`
2. Update `NewDatabaseStorage` to accept logger parameter
3. Log database operations at DEBUG level
4. Log errors at ERROR level with operation context
5. Log migration events at INFO level

**Logging Points**:
- Database connection open/close
- Migration start/complete/failure
- CRUD operations (at DEBUG level)
- Query errors with SQL context
- Row count for bulk operations

### 3. Auth Service (`auth/service.go`)

**Changes**:
1. Add `logger *slog.Logger` field to `Service`
2. Update `NewService` to accept logger parameter
3. Log registration attempts (with masked email)
4. Log login attempts and outcomes
5. Log JWT validation failures

**Logging Points**:
- User registration (success/failure)
- Login attempts (success/failure, masked email)
- Password validation failures
- JWT token generation
- JWT token validation failures

### 4. Auth Middleware (`auth/middleware.go`)

**Changes**:
1. Add `logger *slog.Logger` field to `AuthMiddleware`
2. Update `NewAuthMiddleware` to accept logger parameter
3. Log authentication failures with reason
4. Log missing/invalid tokens

**Logging Points**:
- Missing Authorization header
- Invalid token format
- Expired tokens
- Invalid signatures
- User ID extraction from token

### 5. HTTP Handlers (`internal/handlers/handlers.go`)

**Changes**:
1. Create handler struct to hold logger
2. Update handler functions to methods on struct
3. Log request processing steps
4. Log validation errors
5. Log business logic errors

**Logging Points**:
- Request parsing errors
- Validation failures
- Task not found errors
- Permission denied errors
- Internal errors with context

### 6. CLI (`cmd/cli/main.go`)

**Changes**:
1. Create logger with text format for stderr
2. Set log level to ERROR by default (quiet mode)
3. Add `--debug` flag to enable DEBUG level
4. Log internal errors only (not user-facing messages)
5. Keep stdout clean for user interaction

**Logging Points**:
- HTTP client errors (connection failures)
- Authentication errors
- Configuration loading errors
- Internal CLI errors

## Error Handling

### Error Logging Strategy

1. **Log at the point of handling, not at the point of creation**
   - Errors should be wrapped as they propagate up
   - Log once at the handler/service boundary
   - Avoid duplicate logging

2. **Include context in error logs**
   - Operation being performed
   - Input parameters (sanitized)
   - User ID (if available)
   - Request ID (if available)

3. **Use appropriate log levels**
   - ERROR: Unexpected errors, failures that need attention
   - WARN: Expected errors (validation, not found), auth failures
   - INFO: Normal operations, state changes
   - DEBUG: Detailed operation traces, query details

### Example Error Logging

```go
// In handler
task, err := s.GetTaskByID(id, userID)
if err != nil {
    if errors.Is(err, storage.ErrTaskNotFound) {
        logger.Warn("Task not found",
            slog.Int("task_id", id),
            slog.Int("user_id", userID),
            slog.String("request_id", GetRequestID(r.Context())))
        handlers.JSONError(w, http.StatusNotFound, "Task not found")
        return
    }
    logger.Error("Failed to get task",
        slog.Int("task_id", id),
        slog.Int("user_id", userID),
        slog.String("error", err.Error()),
        slog.String("request_id", GetRequestID(r.Context())))
    handlers.JSONError(w, http.StatusInternalServerError, "Internal error")
    return
}
```

## Testing Strategy

### Unit Testing

1. **Test logger creation and configuration**
   - Verify correct handler selection (JSON vs Text)
   - Verify log level filtering
   - Verify output destination

2. **Test middleware**
   - Verify request ID generation
   - Verify request/response logging
   - Verify duration calculation
   - Verify panic recovery

3. **Test context helpers**
   - Verify request ID storage/retrieval
   - Verify logger storage/retrieval
   - Verify trace ID handling

### Integration Testing

1. **Test with in-memory buffer**
   - Use `TestBuffer` to capture logs
   - Assert log entries contain expected fields
   - Assert log levels are correct
   - Assert sensitive data is masked

2. **Test with real handlers**
   - Create test server with logging middleware
   - Make HTTP requests
   - Verify logs contain request details
   - Verify error logs contain context

### Example Test

```go
func TestLoggingMiddleware(t *testing.T) {
    logger, buf := logger.NewTest()
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    middleware := logger.LoggingMiddleware(logger)
    wrappedHandler := middleware(handler)
    
    req := httptest.NewRequest("GET", "/tasks", nil)
    rec := httptest.NewRecorder()
    
    wrappedHandler.ServeHTTP(rec, req)
    
    entries := buf.Entries()
    assert.Equal(t, 2, len(entries)) // Start and complete
    assert.Equal(t, "HTTP request completed", entries[1].Message)
    assert.Equal(t, 200, entries[1].Attrs["status_code"])
}
```

## Configuration Integration

### Extend `cmd/server/config/config.go`

Add logging configuration to the existing Config struct:

```go
type Config struct {
    ServerConfig   ServerConfig   `mapstructure:"server"`
    DatabaseConfig DatabaseConfig `mapstructure:"database"`
    JWTConfig      JWTConfig      `mapstructure:"jwt"`
    LogConfig      LogConfig      `mapstructure:"logging"`  // NEW
}

type LogConfig struct {
    Level       string `mapstructure:"level"`
    Format      string `mapstructure:"format"`
    Output      string `mapstructure:"output"`
    AddSource   bool   `mapstructure:"add_source"`
    ServiceName string `mapstructure:"service_name"`
    Environment string `mapstructure:"environment"`
    
    // File rotation
    EnableRotation bool `mapstructure:"enable_rotation"`
    MaxSize        int  `mapstructure:"max_size"`
    MaxAge         int  `mapstructure:"max_age"`
    MaxBackups     int  `mapstructure:"max_backups"`
}
```

### Configuration File Example (`config.yaml`)

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  path: "./data/tasks.db"

jwt:
  secret: "your-secret-key-here-min-32-chars"
  expiration: "24h"

logging:
  level: "info"              # debug, info, warn, error
  format: "json"             # json, text
  output: "stdout"           # stdout, stderr, or file path
  add_source: true           # Include file:line for errors
  service_name: "task-manager-api"
  environment: "production"  # development, staging, production
  
  # File rotation (only used if output is a file path)
  enable_rotation: true
  max_size: 100              # MB
  max_age: 30                # days
  max_backups: 5
```

### Environment Variables

```bash
TASKMANAGER_LOGGING_LEVEL=debug
TASKMANAGER_LOGGING_FORMAT=json
TASKMANAGER_LOGGING_OUTPUT=/var/log/taskmanager/app.log
TASKMANAGER_LOGGING_ENVIRONMENT=production
```

### Command-Line Flags

```bash
--log-level=debug
--log-format=json
--log-output=/var/log/app.log
```

## Log Analysis Tool Integration

### ELK Stack (Elasticsearch, Logstash, Kibana)

**Configuration**:
- Use JSON format (`format: "json"`)
- Output to stdout or file
- Logstash reads from file or stdout
- Elasticsearch indexes logs
- Kibana visualizes

**Field Mapping**:
- `timestamp` → `@timestamp` (Logstash filter)
- `level` → `log.level`
- `service` → `service.name`
- All custom fields preserved

**Example Logstash Config**:
```ruby
input {
  file {
    path => "/var/log/taskmanager/app.log"
    codec => "json"
  }
}

filter {
  mutate {
    rename => { "timestamp" => "@timestamp" }
  }
}

output {
  elasticsearch {
    hosts => ["localhost:9200"]
    index => "taskmanager-%{+YYYY.MM.dd}"
  }
}
```

### Grafana Loki

**Configuration**:
- Use JSON format
- Add Promtail agent to ship logs
- Configure labels for filtering

**Promtail Config**:
```yaml
clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: taskmanager
    static_configs:
      - targets:
          - localhost
        labels:
          job: taskmanager
          environment: production
          __path__: /var/log/taskmanager/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            service: service
            request_id: request_id
      - labels:
          level:
          service:
```

**Query Examples**:
```
{service="task-manager-api"} |= "error"
{service="task-manager-api", level="ERROR"} | json | request_id="req_123"
```

### Datadog

**Configuration**:
- Use JSON format
- Install Datadog agent
- Configure log collection

**Datadog Agent Config** (`/etc/datadog-agent/conf.d/taskmanager.d/conf.yaml`):
```yaml
logs:
  - type: file
    path: /var/log/taskmanager/app.log
    service: task-manager-api
    source: go
    sourcecategory: sourcecode
```

**Features**:
- Automatic parsing of JSON logs
- APM integration with trace IDs
- Alerting on error rates
- Log-based metrics

## Performance Considerations

### 1. Lazy Evaluation

slog supports lazy evaluation of log arguments:

```go
// Bad: Always evaluates expensive operation
logger.Debug("Task details", slog.Any("task", expensiveSerialize(task)))

// Good: Only evaluates if DEBUG is enabled
logger.Debug("Task details", slog.Any("task", task))
```

### 2. Structured Fields vs String Formatting

```go
// Bad: String formatting overhead
logger.Info(fmt.Sprintf("User %d created task %d", userID, taskID))

// Good: Structured fields, no formatting
logger.Info("Task created",
    slog.Int("user_id", userID),
    slog.Int("task_id", taskID))
```

### 3. Buffered Writes

- Use buffered I/O for file output
- slog handles this automatically
- Lumberjack provides buffering for rotation

### 4. Async Logging (Future Enhancement)

For very high-throughput scenarios, consider async logging:
- Use a channel-based buffer
- Background goroutine writes to disk
- Trade-off: Potential log loss on crash

## Security Considerations

### 1. Sensitive Data Masking

**Always mask**:
- Passwords (never log)
- JWT tokens (mask or omit)
- Email addresses (partial masking)
- API keys (mask)
- Credit card numbers (never log)

**Implementation**:
```go
func MaskEmail(email string) string {
    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return "***"
    }
    username := parts[0]
    if len(username) <= 2 {
        return "***@" + parts[1]
    }
    return username[0:1] + "***" + username[len(username)-1:] + "@" + parts[1]
}
```

### 2. Log Injection Prevention

- slog automatically escapes special characters in JSON
- No manual escaping needed
- Structured logging prevents injection attacks

### 3. Access Control

- Restrict log file permissions (0600 or 0640)
- Use separate log files for different sensitivity levels
- Rotate and archive logs securely

### 4. Compliance

- GDPR: Mask PII, implement log retention policies
- PCI-DSS: Never log payment card data
- HIPAA: Encrypt logs containing health information

## Migration Path

### Phase 1: Foundation (Tasks 1-3)
1. Create logger package with configuration
2. Extend config system with logging options
3. Create logger factory and test utilities

### Phase 2: Core Integration (Tasks 4-6)
4. Implement HTTP logging middleware
5. Integrate logger into storage layer
6. Integrate logger into auth service

### Phase 3: Handlers and CLI (Tasks 7-8)
7. Update HTTP handlers with logging
8. Add CLI logging support

### Phase 4: Testing and Documentation (Tasks 9-10)
9. Write comprehensive tests
10. Update documentation and examples

## Future Enhancements

1. **Distributed Tracing Integration**
   - OpenTelemetry integration
   - Automatic trace/span ID propagation
   - Correlation with metrics

2. **Metrics Integration**
   - Log-based metrics (error rates, latency percentiles)
   - Prometheus exporter
   - Custom business metrics

3. **Advanced Filtering**
   - Dynamic log level changes (via API)
   - Per-user log level overrides
   - Sampling for high-volume logs

4. **Log Aggregation**
   - Built-in log shipping
   - Direct integration with cloud providers
   - Real-time log streaming

5. **Structured Error Types**
   - Error codes for categorization
   - Error severity levels
   - Automatic error grouping
