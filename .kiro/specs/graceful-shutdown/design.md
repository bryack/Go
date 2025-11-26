# Design Document: Graceful Shutdown

## Overview

This design implements graceful shutdown for the Task Manager API server using Go's `context` package and `http.Server.Shutdown()` method. The implementation captures OS signals (SIGINT, SIGTERM), initiates an orderly shutdown sequence, allows in-flight requests to complete within a timeout period, and ensures proper cleanup of database connections and other resources.

## Architecture

### High-Level Flow

```
OS Signal (SIGINT/SIGTERM)
    ↓
Signal Handler Goroutine
    ↓
Initiate Shutdown Sequence
    ↓
┌─────────────────────────────────┐
│ 1. Stop accepting new requests  │
│ 2. Wait for in-flight requests  │
│ 3. Close database connections   │
│ 4. Log shutdown completion      │
│ 5. Exit with appropriate code   │
└─────────────────────────────────┘
```

### Component Interaction

```
main()
  ├─> http.Server (with explicit configuration)
  ├─> Signal Handler Goroutine (listens for SIGINT/SIGTERM)
  ├─> Shutdown Context (with timeout)
  └─> Resource Cleanup (database, logger)
```

## Components and Interfaces

### 1. HTTP Server Configuration

**Current State:**
```go
log.Fatal(http.ListenAndServe(address, nil))
```

**New State:**
The server will be explicitly configured as an `http.Server` instance to enable calling `Shutdown()`:

```go
server := &http.Server{
    Addr:    address,
    Handler: nil, // Uses DefaultServeMux
}
```

This allows us to:
- Call `server.Shutdown(ctx)` for graceful shutdown
- Configure server-level timeouts if needed
- Have a reference to the server for shutdown operations

### 2. Signal Handler

**Purpose:** Capture OS termination signals and trigger shutdown sequence.

**Implementation Approach:**
- Use `signal.Notify()` to capture SIGINT and SIGTERM
- Run in a separate goroutine to avoid blocking main execution
- Trigger shutdown when signal is received
- Handle second signal as force quit

**Key Behaviors:**
- First signal: Initiates graceful shutdown
- Second signal: Forces immediate exit with code 1
- Logs the signal type received

### 3. Shutdown Sequence

**Purpose:** Coordinate orderly shutdown of all server components.

**Steps:**
1. Log shutdown initiation with signal type
2. Create context with timeout for shutdown operations
3. Call `server.Shutdown(ctx)` to stop accepting new requests and wait for in-flight requests
4. Close database storage connection
5. Log shutdown completion or timeout
6. Exit with appropriate status code

**Timeout Handling:**
- Use `context.WithTimeout()` to enforce maximum shutdown duration
- Default timeout: 30 seconds (configurable)
- If timeout expires, force shutdown and log warning

### 4. Database Connection Cleanup

**Current Interface:**
The `DatabaseStorage` already has a `Close()` method that should be called during shutdown.

**Integration:**
- Call `storage.Close()` after `server.Shutdown()` completes
- Log any errors from database close operation
- Ensure close is called even if server shutdown times out

### 5. Configuration Extension

**New Configuration Field:**
Add `ShutdownTimeout` to the server configuration:

```go
type ServerConfig struct {
    Host            string
    Port            int
    ShutdownTimeout time.Duration // New field
}
```

**Default Value:** 30 seconds

**Validation:** Must be positive duration

## Data Models

### Configuration Changes

**File:** `cmd/server/config/config.go`

Add `ShutdownTimeout` field to `ServerConfig`:
- Type: `time.Duration`
- Default: `30 * time.Second`
- YAML key: `shutdown_timeout`
- Validation: Must be > 0

### Exit Codes

Define exit code constants for clarity:
- `ExitSuccess = 0` - Clean shutdown
- `ExitShutdownTimeout = 1` - Shutdown timeout exceeded
- `ExitShutdownError = 1` - Error during shutdown

## Error Handling

### Signal Handling Errors

**Scenario:** Signal notification setup fails
**Handling:** Log fatal error and exit immediately (should never happen in practice)

### Server Shutdown Errors

**Scenario:** `server.Shutdown(ctx)` returns error
**Handling:** 
- Log error with details
- Continue with resource cleanup
- Exit with code 1

### Database Close Errors

**Scenario:** `storage.Close()` returns error
**Handling:**
- Log error with details
- Continue with shutdown
- Exit with code 1

### Timeout Errors

**Scenario:** Shutdown context deadline exceeded
**Handling:**
- Log warning with timeout duration
- Force shutdown
- Exit with code 1

## Testing Strategy

### Unit Tests

**File:** `cmd/server/main_test.go`

Test scenarios:
1. Signal handler captures SIGINT correctly
2. Signal handler captures SIGTERM correctly
3. Second signal forces immediate exit
4. Shutdown timeout is respected
5. Database close is called during shutdown

**Approach:**
- Use channels to simulate signal delivery
- Mock database storage with close tracking
- Use test server for HTTP server testing
- Verify log output contains expected messages

### Integration Tests

**Approach:**
- Start real server in test
- Send actual OS signals
- Verify server stops accepting new requests
- Verify in-flight requests complete
- Verify clean exit

**Challenges:**
- Requires actual process signal handling
- May be flaky in CI environments
- Consider marking as optional or manual

### Manual Testing

**Test Cases:**
1. Start server, send SIGINT (Ctrl+C), verify clean shutdown
2. Start server with active request, send SIGTERM, verify request completes
3. Start server, send SIGINT twice quickly, verify force quit
4. Configure short timeout, send signal during long request, verify timeout

## Implementation Notes

### Goroutine Management

The signal handler will run in a goroutine. The main goroutine will block on `server.ListenAndServe()` until shutdown is initiated.

**Flow:**
```
main goroutine: server.ListenAndServe() [blocks]
signal goroutine: <-sigChan [blocks until signal]
    → triggers server.Shutdown()
    → main goroutine unblocks
    → cleanup and exit
```

### Context Usage

Two contexts are involved:
1. **Shutdown Context:** Created with timeout, passed to `server.Shutdown()`
2. **Request Contexts:** Existing request contexts remain independent

The shutdown context controls how long we wait for in-flight requests, not how long individual requests can run.

### Logging During Shutdown

All shutdown-related logs should use structured logging with:
- `operation: "graceful_shutdown"`
- `signal: "SIGINT"` or `"SIGTERM"`
- `duration: <time taken>`
- `status: "success"`, `"timeout"`, or `"error"`

### Configuration Loading

The shutdown timeout should be loaded from config with fallback to default:
- Check YAML config file
- Check environment variable `SHUTDOWN_TIMEOUT`
- Fall back to 30 seconds

## Design Decisions

### Why `http.Server.Shutdown()` instead of manual tracking?

**Decision:** Use Go's built-in `http.Server.Shutdown()` method.

**Rationale:**
- Battle-tested implementation
- Handles edge cases (keep-alive connections, websockets, etc.)
- Simpler than manual request tracking
- Standard Go pattern

### Why 30 seconds default timeout?

**Decision:** Default shutdown timeout of 30 seconds.

**Rationale:**
- Most API requests complete in < 5 seconds
- Provides buffer for slow requests or database operations
- Long enough to be safe, short enough to not delay deployments
- Can be overridden via configuration

### Why close database after server shutdown?

**Decision:** Close database connections after `server.Shutdown()` completes.

**Rationale:**
- In-flight requests may still need database access
- Ensures all requests complete before closing DB
- Prevents "database closed" errors during shutdown

### Why support force quit on second signal?

**Decision:** Second signal forces immediate exit.

**Rationale:**
- Gives operators escape hatch if graceful shutdown hangs
- Standard behavior in many servers (nginx, etc.)
- Prevents indefinite hangs

## Alternatives Considered

### Alternative 1: Manual Request Tracking

**Approach:** Use sync.WaitGroup to track active requests.

**Rejected Because:**
- Reinvents the wheel
- More complex and error-prone
- Doesn't handle all edge cases
- `http.Server.Shutdown()` already does this

### Alternative 2: No Timeout

**Approach:** Wait indefinitely for requests to complete.

**Rejected Because:**
- Could hang forever on stuck requests
- Delays deployments and restarts
- No escape hatch for operators
- Timeout is standard practice

### Alternative 3: Immediate Shutdown

**Approach:** Close server immediately on signal.

**Rejected Because:**
- Drops active requests
- Can cause data loss
- Poor user experience
- Violates requirements

## Migration Path

This is a new feature with no breaking changes:
1. Add configuration field (optional, has default)
2. Modify main() to use explicit server and signal handling
3. No API changes
4. No database schema changes
5. Backward compatible with existing deployments

## Security Considerations

### Signal Handling

- Only handle SIGINT and SIGTERM (standard termination signals)
- Don't handle SIGKILL (can't be caught anyway)
- Second signal forces exit to prevent DoS via signal spam

### Resource Cleanup

- Ensure database connections close to prevent connection leaks
- Don't expose internal state in shutdown logs
- Exit codes don't leak sensitive information

## Performance Considerations

### Shutdown Performance

- Shutdown time = min(in-flight request time, timeout)
- Typical shutdown: < 5 seconds
- Worst case: 30 seconds (timeout)
- No performance impact during normal operation

### Memory Usage

- Signal handler goroutine: negligible overhead
- No additional memory allocation during shutdown
- Existing request memory is freed as requests complete

## Monitoring and Observability

### Metrics to Track

- Shutdown duration
- Number of in-flight requests at shutdown
- Shutdown success vs timeout rate
- Database close errors

### Log Events

1. **Shutdown Initiated:** Signal type, timestamp
2. **Shutdown Complete:** Duration, status
3. **Shutdown Timeout:** Timeout value, in-flight requests
4. **Database Close:** Success or error
5. **Force Quit:** Second signal received

### Example Log Output

```
INFO: Shutdown initiated signal=SIGTERM
INFO: Server shutdown complete duration=2.3s status=success
INFO: Database connections closed
INFO: Graceful shutdown complete exit_code=0
```

## Future Enhancements

Potential improvements not in current scope:
- Metrics endpoint for in-flight request count
- Health check returns 503 during shutdown
- Drain mode (stop accepting new requests before shutdown)
- Configurable shutdown hooks for custom cleanup
- Graceful shutdown for background workers (if added later)
