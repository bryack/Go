# Task Decomposition: Implement Graceful Shutdown Sequence

## Overview

Enhance the signal handler to properly coordinate the graceful shutdown sequence with error handling, timeout detection, duration tracking, and shutdown completion logging. This transforms the basic signal capture into a production-ready shutdown mechanism that meets all observability and reliability requirements.

## Implementation Approach

We're refactoring the signal handler goroutine to track shutdown duration, handle errors from `server.Shutdown()`, detect timeout scenarios, and log completion status. The key insight is that we need to wait for shutdown to complete and handle different outcomes before the process exits.

**Complexity Check:**
- **Requirements need**: Error handling, timeout detection, duration logging, completion status
- **Simple approach**: Add error checking, time tracking, and status logging in the signal handler (15-20 min)
- **Complex approach**: Separate shutdown coordinator with state machine, metrics, hooks (45+ min)
- **Recommendation**: Simple approach. Requirements only need error handling and logging. The shutdown sequence is linear: shutdown server → check result → log outcome. No need for complex state management or extensibility hooks.

**Key Concepts:**
- **time.Now()**: Capture start time to calculate shutdown duration
- **context.DeadlineExceeded**: Error returned when shutdown timeout is exceeded
- **server.Shutdown() error**: Can return errors if shutdown fails
- **Structured logging**: Log shutdown outcome with duration and status fields

## Prerequisites

**Existing Code:**
- `cmd/server/main.go` - Signal handler goroutine (lines 443-453)
- `server` variable available for shutdown
- `l` (logger) available for logging
- `cfg.ServerConfig.ShutdownTimeout` configured
- Context with timeout already created

**Dependencies:**
- `time` package - Already imported
- `context` package - Already imported
- `errors` package - Already imported

**Knowledge Required:**
- Understanding that `server.Shutdown()` blocks until shutdown completes or context times out
- `context.DeadlineExceeded` is the error returned on timeout
- `time.Since()` calculates duration from a start time

## Step-by-Step Instructions

### Step 1: Track shutdown start time

**File**: `cmd/server/main.go`

**What to do:**
Add time tracking to measure shutdown duration.

**What to implement:**
- Locate the signal handler goroutine (around line 443)
- After logging "Shutdown signal received", capture the start time
- Add a line: `shutdownStart := time.Now()`
- Place it right after the log statement, before creating the context

**Why:**
We need to track how long shutdown takes to meet requirement 3.2: "log the completion event with duration". This helps operators understand shutdown performance and detect slow shutdowns.

**Expected result:**
Start time is captured. Code compiles without errors.

---

### Step 2: Handle server.Shutdown() error

**File**: `cmd/server/main.go`

**What to do:**
Capture and handle the error returned by `server.Shutdown()`.

**What to implement:**
- Locate the `server.Shutdown(ctx)` call (around line 451)
- Change it to capture the error: `err := server.Shutdown(ctx)`
- This allows us to check if shutdown succeeded or failed

**Why:**
`server.Shutdown()` can return errors if shutdown fails or times out. We need to handle these cases differently to meet requirements 1.4, 1.5, 3.2, and 3.3.

**Expected result:**
Error is captured. Code compiles without errors.

---

### Step 3: Detect and log timeout scenario

**File**: `cmd/server/main.go`

**What to do:**
Check if shutdown timed out and log appropriately.

**What to implement:**
- After capturing the error from `server.Shutdown()`, add an if statement
- Check if the error is a timeout: `if errors.Is(err, context.DeadlineExceeded)`
- If timeout occurred, log a warning with structured fields:
  - Message: "Graceful shutdown timed out"
  - Duration: `time.Since(shutdownStart)`
  - Timeout value: `cfg.ServerConfig.ShutdownTimeout`
- Use `slog.Duration()` for duration fields
- This should be inside the signal handler goroutine

**Why:**
Requirement 3.3 explicitly requires: "IF graceful shutdown times out, THEN THE Server SHALL log a warning with the timeout duration". This helps operators identify when shutdown is taking too long and may need timeout adjustment.

**Expected result:**
Timeout is detected and logged with duration. Code compiles without errors.

---

### Step 4: Log other shutdown errors

**File**: `cmd/server/main.go`

**What to do:**
Handle non-timeout errors from shutdown.

**What to implement:**
- After the timeout check, add an `else if err != nil` block
- Log an error with structured fields:
  - Message: "Server shutdown failed"
  - Error: `err.Error()`
  - Duration: `time.Since(shutdownStart)`
- Use `slog.String("error", err.Error())` for the error field

**Why:**
Requirement 3.4 requires: "IF any resource cleanup fails during shutdown, THEN THE Server SHALL log an error with failure details". This catches any unexpected shutdown errors.

**Expected result:**
Non-timeout errors are logged. Code compiles without errors.

---

### Step 5: Log successful shutdown completion

**File**: `cmd/server/main.go`

**What to do:**
Log when shutdown completes successfully.

**What to implement:**
- After the error handling blocks, add an `else` block for the success case
- Log an info message with structured fields:
  - Message: "Server shutdown completed successfully"
  - Duration: `time.Since(shutdownStart)`
  - Status: "success"
- Use `slog.String("status", "success")` for the status field

**Why:**
Requirement 3.2 explicitly requires: "WHEN graceful shutdown completes successfully, THE Server SHALL log the completion event with duration". This confirms shutdown worked as expected.

**Expected result:**
Successful shutdown is logged with duration. Code compiles without errors.

---

### Step 6: Remove log.Fatal wrapper

**File**: `cmd/server/main.go`

**What to do:**
Change the server startup to not use `log.Fatal()`.

**What to implement:**
- Locate the `log.Fatal(server.ListenAndServe())` line (around line 461)
- Replace it with two lines:
  - `err := server.ListenAndServe()`
  - `if err != nil && err != http.ErrServerClosed { ... }`
- In the if block, log the error and exit with code 1
- Use `l.Error()` to log the error
- Call `os.Exit(1)` to exit with error code

**Why:**
`log.Fatal()` always exits with code 1, even on successful shutdown. We need to remove it so we can control the exit code based on shutdown outcome (task 6). The check for `http.ErrServerClosed` is important because that's the expected error when `Shutdown()` is called.

**Expected result:**
Server startup no longer uses `log.Fatal()`. Code compiles without errors.

---

### Step 7: Add temporary exit after shutdown

**File**: `cmd/server/main.go`

**What to do:**
Add a temporary exit point after shutdown completes.

**What to implement:**
- At the end of the signal handler goroutine, after all the logging
- Add `os.Exit(0)` as a temporary measure
- This will be replaced with proper exit code logic in task 6
- Add a comment: `// TODO: Task 6 - implement proper exit code handling`

**Why:**
Without this, the program would continue running after shutdown completes (the main goroutine would still be blocked). This is a temporary solution until task 6 implements proper exit code handling based on shutdown outcome.

**Expected result:**
Program exits after shutdown completes. Code compiles without errors.

## Verification

### Compile Check
```bash
go build ./cmd/server
```
**Expected**: No compilation errors

### Successful Shutdown Test
```bash
# Start the server
go run cmd/server/main.go

# Press Ctrl+C once
# Expected logs:
# - "Shutdown signal received" with signal type
# - "Server shutdown completed successfully" with duration
# - Process exits
```

### Timeout Test (requires modification)
```bash
# Temporarily change shutdown timeout to 1 second in config.yaml:
# shutdown_timeout: "1s"

# Start server and make a long-running request in another terminal
# Then press Ctrl+C

# Expected logs:
# - "Shutdown signal received"
# - "Graceful shutdown timed out" with duration and timeout value
# - Process exits
```

### Error Handling Test
```bash
# Start the server
go run cmd/server/main.go

# Verify it starts successfully
curl http://localhost:8080/health

# Press Ctrl+C
# Expected: Clean shutdown with success log
```

### Duration Logging Test
```bash
# Start server
go run cmd/server/main.go

# Wait a few seconds, then Ctrl+C
# Expected: Log shows duration (should be very small, like "duration":0.001s)
```

## Common Pitfalls

### Pitfall 1: Not checking for http.ErrServerClosed
**Symptom**: Error log appears even on successful shutdown
**Fix**: Check `if err != nil && err != http.ErrServerClosed` - `ErrServerClosed` is expected when `Shutdown()` is called

### Pitfall 2: Forgetting to capture start time
**Symptom**: Compilation error about undefined `shutdownStart`
**Fix**: Add `shutdownStart := time.Now()` before creating the context

### Pitfall 3: Using errors.Is incorrectly
**Symptom**: Timeout not detected properly
**Fix**: Use `errors.Is(err, context.DeadlineExceeded)` not `err == context.DeadlineExceeded`

### Pitfall 4: Logging duration without time.Since()
**Symptom**: Duration shows as a large number instead of readable format
**Fix**: Use `slog.Duration("duration", time.Since(shutdownStart))` which formats it properly

### Pitfall 5: Not exiting after shutdown
**Symptom**: Program hangs after shutdown completes
**Fix**: Add `os.Exit(0)` at the end of the signal handler goroutine (temporary until task 6)

### Pitfall 6: Blocking main goroutine incorrectly
**Symptom**: Server doesn't start or exits immediately
**Fix**: Ensure `server.ListenAndServe()` is still called in the main goroutine, not in a goroutine

## Learning Resources

### Essential Reading
- [http.Server.Shutdown Documentation](https://pkg.go.dev/net/http#Server.Shutdown) - Understanding shutdown behavior and errors
- [context.DeadlineExceeded](https://pkg.go.dev/context#pkg-variables) - Understanding timeout detection
- [time.Since](https://pkg.go.dev/time#Since) - Calculating durations

### Additional Resources (Optional)
- [Structured Logging with slog](https://pkg.go.dev/log/slog) - Best practices for structured logging
- [Go Error Handling](https://go.dev/blog/error-handling-and-go) - Understanding error checking patterns

## Notes

**Implementation Pattern:**

The enhanced signal handler looks like this:
```go
go func() {
    sig := <-shutdownChan
    l.Info("Shutdown signal received", slog.String("signal", sig.String()))
    
    shutdownStart := time.Now()
    ctx, cancel := context.WithTimeout(context.Background(), cfg.ServerConfig.ShutdownTimeout)
    defer cancel()
    
    err := server.Shutdown(ctx)
    
    if errors.Is(err, context.DeadlineExceeded) {
        l.Warn("Graceful shutdown timed out",
            slog.Duration("duration", time.Since(shutdownStart)),
            slog.Duration("timeout", cfg.ServerConfig.ShutdownTimeout),
        )
    } else if err != nil {
        l.Error("Server shutdown failed",
            slog.String("error", err.Error()),
            slog.Duration("duration", time.Since(shutdownStart)),
        )
    } else {
        l.Info("Server shutdown completed successfully",
            slog.Duration("duration", time.Since(shutdownStart)),
            slog.String("status", "success"),
        )
    }
    
    // TODO: Task 6 - implement proper exit code handling
    os.Exit(0)
}()
```

**What this achieves:**
- ✅ Requirement 1.4: Detects when shutdown exceeds timeout
- ✅ Requirement 1.5: Logs shutdown completion status
- ✅ Requirement 3.2: Logs completion with duration on success
- ✅ Requirement 3.3: Logs warning with timeout duration on timeout
- ✅ Requirement 3.4: Logs errors with details on failure
- ✅ Foundation for task 6: Exit code handling can now be based on shutdown outcome

**What this doesn't handle yet:**
- Database cleanup (task 5)
- Proper exit codes based on outcome (task 6)
- Exit code 0 vs 1 based on success/failure (task 6)

These will be implemented in subsequent tasks. For now, we always exit with code 0 as a temporary measure.
