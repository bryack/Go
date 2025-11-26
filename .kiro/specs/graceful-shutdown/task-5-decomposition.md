# Task Decomposition: Implement Database Cleanup During Shutdown

## Overview

Add proper database connection cleanup to the graceful shutdown sequence. Currently, the server shuts down gracefully but never closes the database connection, which can lead to resource leaks, connection pool exhaustion, and potential data integrity issues with SQLite. This task ensures the database connection is properly closed after the HTTP server completes its shutdown.

## Implementation Approach

We're adding a `Close()` method to `DatabaseStorage` that wraps the underlying `sql.DB.Close()` call, then calling it in the shutdown sequence after `server.Shutdown()` completes. This follows the standard Go pattern of explicit resource cleanup and ensures proper ordering: first stop accepting requests, then close database connections.

The implementation is straightforward because Go's `database/sql` package already provides the `Close()` method on `*sql.DB`. We just need to expose it through our `DatabaseStorage` interface and call it at the right time in the shutdown sequence.

**Complexity Check:**
- **Requirements need**: Close database connections during shutdown (Requirement 2.1, 2.2, 2.3, 2.4)
- **Simple approach**: Add Close() method to DatabaseStorage, call it after server.Shutdown() (10 min)
- **Complex approach**: Implement connection tracking, graceful drain, timeout handling (45+ min)
- **Recommendation**: Start simple. The requirements just need the database closed properly. The `sql.DB.Close()` method already handles waiting for connections to be released. No need for custom tracking or complex drain logic.

**Key Concepts:**
- **Resource cleanup order**: HTTP server stops first (no new requests), then database closes (no new queries)
- **sql.DB.Close()**: Blocks until all connections are returned to the pool or closed
- **Error handling**: Database close errors should be logged but shouldn't prevent process exit

## Prerequisites

**Existing Code:**
- `storage/database.go` - DatabaseStorage struct with db field
- `cmd/server/main.go` - Graceful shutdown sequence (tasks 1-4 complete)
- `logger` package - Structured logging for shutdown events

**Dependencies:**
- `database/sql` package (already imported)
- `log/slog` package (already imported)

**Knowledge Required:**
- Understanding of Go's `defer` statement and resource cleanup patterns
- Basic understanding of database connection lifecycle
- Familiarity with the existing shutdown sequence in main.go

## Step-by-Step Instructions

### Step 1: Add Close method to DatabaseStorage

**File**: `storage/database.go`

**What to do:**
Add a `Close()` method to the `DatabaseStorage` struct that properly closes the database connection.

**What to implement:**
- **Option A (Simpler)**: Just add the method to `DatabaseStorage` struct
- **Option B (More complete)**: Add `Close() error` to the `Storage` interface, then implement it on `DatabaseStorage`
- Recommendation: Use Option B to keep the interface complete, but Option A works fine too
- The method should be public (capitalized `Close`) and return an error
- Call `ds.db.Close()` to close the underlying database connection
- Add structured logging before attempting to close (debug level)
- Add structured logging after successful close (info level)
- Add structured logging if close fails (error level)
- Return any error from `db.Close()` wrapped with context
- Use the existing logger field (`ds.logger`) for all log statements
- Include the operation name in logs using `logger.FieldOperation` constant

**Why:**
The `DatabaseStorage` struct owns the `*sql.DB` connection, so it should be responsible for closing it. This follows Go's principle of "whoever opens it, closes it." The `db` field is unexported, so you cannot call `s.db.Close()` directly from main.go. Adding it to the interface (Option B) makes the contract complete and helps with future testing/mocking.

**Expected result:**
Code compiles. DatabaseStorage now has a Close() method that can be called during shutdown.

---

### Step 2: Call storage.Close() in shutdown sequence

**File**: `cmd/server/main.go`

**What to do:**
Add a call to close the database storage after the HTTP server completes its shutdown.

**What to implement:**
- Locate the shutdown goroutine (the one that handles `<-shutdownChan`)
- After the `server.Shutdown(ctx)` call completes (after the if/else block that handles shutdown errors)
- Before the `os.Exit(0)` call
- Add a call to `s.Close()` where `s` is the storage instance
- Check if the close operation returns an error
- If error is not nil, log it as an error with structured fields (operation, error message)
- If error is not nil, change the exit code from 0 to 1
- If close succeeds, log it as info with structured fields (operation, status)
- Use the existing logger instance `l` for all logging
- Ensure the close is called even if server shutdown times out or errors

**Why:**
The database must be closed after the server stops accepting requests but before the process exits. This ensures all in-flight requests complete their database operations before we close the connection. Calling it in the shutdown goroutine ensures it happens during graceful shutdown, not during normal operation.

**Expected result:**
Server compiles and starts successfully. When you send SIGINT/SIGTERM, you see database close logs in the output.

---

### Step 3: Update exit code logic

**File**: `cmd/server/main.go`

**What to do:**
Modify the shutdown sequence to track whether any errors occurred and exit with the appropriate code.

**What to implement:**
- At the start of the shutdown goroutine, create a variable to track exit code (initialize to 0)
- When `server.Shutdown(ctx)` returns an error (either timeout or other error), set exit code to 1
- When `s.Close()` returns an error, set exit code to 1
- Replace the hardcoded `os.Exit(0)` at the end with `os.Exit(exitCode)` using the tracked variable
- Ensure the exit code reflects the worst outcome (if both server shutdown and db close fail, still exit with 1)

**Why:**
Exit codes communicate success or failure to orchestration tools (Docker, Kubernetes, systemd). A non-zero exit code indicates something went wrong during shutdown, which helps operators detect and troubleshoot issues. This satisfies requirements 5.2 and 5.3.

**Expected result:**
Server exits with code 0 on clean shutdown, code 1 if any errors occur during shutdown.

---

### Step 4: Handle force quit scenario

**File**: `cmd/server/main.go`

**What to do:**
Ensure the second signal handler (force quit) also attempts to close the database before exiting.

**What to implement:**
- Locate the second goroutine that handles force quit (`<-shutdownChan` that calls `os.Exit(1)`)
- Before calling `os.Exit(1)`, add a call to `s.Close()`
- Ignore any errors from close (we're force quitting anyway)
- Keep the log message about force shutdown
- Keep the exit code as 1 (force quit is always an error condition)

**Why:**
Even during force quit, we should attempt to close the database cleanly. The `Close()` call will return quickly if connections are already closed or will force-close them. This gives us the best chance of clean shutdown even in the force quit scenario. This satisfies requirement 2.4 (ensure close is called before process exit).

**Expected result:**
When you send two signals quickly (Ctrl+C twice), the server attempts to close the database before force exiting.

## Verification

### Compile Check
```bash
go build ./...
```
**Expected**: No compilation errors

### Run Tests
```bash
# Test the storage layer
go test ./storage -v

# Test the server (if tests exist)
go test ./cmd/server -v
```
**Expected**: All tests pass

### Manual Testing - Normal Shutdown

1. Start the server:
```bash
go run cmd/server/main.go
```

2. In another terminal, send SIGTERM:
```bash
pkill -SIGTERM server
# or find the PID and: kill -TERM <pid>
```

3. **Expected logs** (in order):
```
INFO: Shutdown signal received signal=SIGTERM
INFO: Server shutdown completed successfully duration=<time> status=success
INFO: Database connections closed
INFO: Graceful shutdown complete
```

4. **Expected exit code**: 0
```bash
echo $?  # Should print 0
```

### Manual Testing - Shutdown with Active Request

1. Start the server:
```bash
go run cmd/server/main.go
```

2. In another terminal, start a long request (if you have one), or just send shutdown immediately:
```bash
# Send shutdown signal
pkill -SIGTERM server
```

3. **Expected behavior**: 
   - Server stops accepting new requests
   - Existing requests complete
   - Database closes after requests finish
   - Clean exit with code 0

### Manual Testing - Force Quit

1. Start the server:
```bash
go run cmd/server/main.go
```

2. Send SIGTERM twice quickly:
```bash
pkill -SIGTERM server
pkill -SIGTERM server
```

3. **Expected logs**:
```
INFO: Shutdown signal received signal=SIGTERM
WARN: Force shutdown signal received, exiting immediately
```

4. **Expected exit code**: 1
```bash
echo $?  # Should print 1
```

### Manual Testing - Database Close Error

This is harder to test manually, but you can verify the code path:

1. Review the code to ensure database close errors are logged
2. Verify exit code is set to 1 when close fails
3. Consider adding a temporary test that forces a close error (close the db twice)

## Common Pitfalls

### Pitfall 1: Calling Close() before server.Shutdown() completes
**Symptom**: "database is closed" errors in logs during shutdown, requests fail
**Fix**: Ensure `s.Close()` is called AFTER the `server.Shutdown(ctx)` call and its error handling, not before or in parallel

### Pitfall 2: Not checking the error from Close()
**Symptom**: Database close failures are silent, exit code is always 0
**Fix**: Always check `if err := s.Close(); err != nil` and log the error, set exit code to 1

### Pitfall 3: Forgetting to close in force quit path
**Symptom**: Force quit leaves database connections open, potential file locks on SQLite
**Fix**: Add `s.Close()` call in the second signal handler goroutine before `os.Exit(1)`

### Pitfall 4: Using defer for Close() in main()
**Symptom**: Close() is called after os.Exit(), which means it never runs
**Fix**: Don't use `defer s.Close()` in main(). Call it explicitly in the shutdown sequence before os.Exit()

## Learning Resources

### Essential Reading
- [database/sql Close documentation](https://pkg.go.dev/database/sql#DB.Close) - Official docs on DB.Close() behavior
- [Go database/sql tutorial](https://go.dev/doc/database/manage-connections) - Understanding connection lifecycle
- [Graceful shutdown in Go](https://pkg.go.dev/net/http#Server.Shutdown) - Context for the shutdown sequence

### Additional Resources (Optional)
- [SQLite and Go](https://github.com/mattn/go-sqlite3#connection-string) - SQLite-specific connection behavior
- [Resource cleanup patterns in Go](https://go.dev/blog/defer-panic-and-recover) - Understanding defer and explicit cleanup
