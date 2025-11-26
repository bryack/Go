# Task Decomposition: Implement Signal Handler Goroutine

## Overview

Implement a signal handler that captures OS termination signals (SIGINT from Ctrl+C and SIGTERM from container orchestrators) and initiates the graceful shutdown sequence. The handler runs in a separate goroutine to avoid blocking the main server, and supports force-quit on a second signal for emergency situations.

## Implementation Approach

We're using Go's `os/signal` package to capture SIGINT and SIGTERM signals in a dedicated goroutine. When a signal is received, we'll trigger the shutdown sequence. A second signal forces immediate exit.

**Complexity Check:**
- **Requirements need**: Capture SIGINT/SIGTERM, log signal type, handle second signal as force quit
- **Simple approach**: One goroutine with signal channel, basic shutdown trigger (10-15 min)
- **Complex approach**: Multiple signal handlers, signal buffering, graceful degradation states (30+ min)
- **Recommendation**: Simple approach. Requirements only need basic signal capture and shutdown trigger. The complex patterns add no value for this use case - we just need to detect the signal and start shutdown.

**Key Concepts:**
- **os/signal**: Go's package for handling OS signals
- **signal.Notify()**: Registers a channel to receive specific signals
- **Goroutine**: Separate execution thread that doesn't block main
- **Buffered channel**: Channel with capacity to prevent signal loss
- **SIGINT**: Signal sent by Ctrl+C (interrupt)
- **SIGTERM**: Signal sent by container orchestrators for graceful termination

## Prerequisites

**Existing Code:**
- `cmd/server/main.go` - Server initialization with explicit `http.Server` instance (line 434-438)
- `server` variable is available for calling `Shutdown()`
- `l` (logger) variable is available for logging
- `cfg` variable has the shutdown timeout configuration

**Dependencies:**
- `os/signal` package - Standard library
- `syscall` package - Standard library (for SIGINT, SIGTERM constants)
- `context` package - Standard library (will be used in task 4)

**Knowledge Required:**
- Understanding that goroutines run concurrently with main
- Signal channels receive OS signals asynchronously
- Buffered channels prevent blocking when capacity isn't exceeded

## Step-by-Step Instructions

### Step 1: Import required packages

**File**: `cmd/server/main.go`

**What to do:**
Add the necessary imports for signal handling at the top of the file.

**What to implement:**
- Locate the import block (around line 3-19)
- Add `"os/signal"` to the imports
- Add `"syscall"` to the imports
- Add `"context"` to the imports (will be used in this and next tasks)
- Keep imports organized (standard library packages first, then third-party, then local)

**Why:**
- `os/signal` provides the `Notify()` function to register for signals
- `syscall` provides the SIGINT and SIGTERM constants
- `context` will be used to create the shutdown timeout context

**Expected result:**
Code compiles without errors. New imports are available for use.

---

### Step 2: Create signal channel and register for signals

**File**: `cmd/server/main.go`

**What to do:**
Set up a channel to receive OS signals and register for SIGINT and SIGTERM.

**What to implement:**
- Locate the end of the `main()` function, just before the `server.ListenAndServe()` call (around line 434)
- Create a buffered channel of type `os.Signal` with capacity 1
- Name it `sigChan` or `shutdownChan`
- Call `signal.Notify()` with the channel and the signals to capture: `syscall.SIGINT` and `syscall.SIGTERM`
- Use buffer size of 1 to prevent signal loss if a signal arrives before the goroutine reads it

**Why:**
The buffered channel ensures that if a signal arrives before our goroutine is ready to read it, the signal won't be lost. We register for both SIGINT (Ctrl+C) and SIGTERM (container stop) to handle both development and production scenarios.

**Expected result:**
Signal channel is created and registered. Code compiles without errors.

---

### Step 3: Launch signal handler goroutine

**File**: `cmd/server/main.go`

**What to do:**
Create a goroutine that waits for signals and initiates shutdown.

**What to implement:**
- After creating the signal channel, launch a goroutine using `go func() { ... }()`
- Inside the goroutine, use a `select` statement or simple channel receive to wait for the first signal
- When a signal is received, log the shutdown initiation with the signal type
- Use structured logging: `l.Info("Shutdown signal received", slog.String("signal", sig.String()))`
- After logging, call `server.Shutdown()` with a context (we'll create the context in this step)
- Create a context with timeout using `context.WithTimeout(context.Background(), cfg.ServerConfig.ShutdownTimeout)`
- Store the returned context and cancel function
- Defer the cancel function to clean up resources
- Pass the context to `server.Shutdown(ctx)`
- For now, ignore the error from `Shutdown()` (will be handled in task 4)

**Why:**
The goroutine runs concurrently with the main server, allowing it to listen for signals without blocking. The context with timeout ensures shutdown doesn't hang forever. Logging the signal type helps operators understand what triggered the shutdown.

**Expected result:**
Goroutine is launched and waits for signals. Code compiles without errors.

---

### Step 4: Implement second signal force quit

**File**: `cmd/server/main.go`

**What to do:**
Add logic to handle a second signal as an immediate force quit.

**What to implement:**
- After the first signal is received and logged, set up another signal receive
- Use a `select` statement with two cases:
  - Case 1: Another signal from `sigChan` - force quit
  - Case 2: A channel that never sends (or just don't have a second case initially)
- In the force quit case, log a warning: `l.Warn("Force shutdown signal received, exiting immediately")`
- Call `os.Exit(1)` to immediately terminate
- This should be in a separate goroutine or after initiating shutdown

**Why:**
If graceful shutdown hangs or takes too long, operators need an escape hatch. A second Ctrl+C forces immediate exit. This is standard behavior in production servers (nginx, etc.) and meets requirement 5.4.

**Expected result:**
First signal initiates graceful shutdown. Second signal forces immediate exit with code 1.

---

### Step 5: Keep server.ListenAndServe() blocking

**File**: `cmd/server/main.go`

**What to do:**
Ensure the main goroutine still blocks on `server.ListenAndServe()`.

**What to implement:**
- Keep the existing `log.Fatal(server.ListenAndServe())` line
- This should be the last line in `main()`
- The signal handler goroutine runs concurrently and will call `server.Shutdown()` when needed
- When `Shutdown()` is called, `ListenAndServe()` will return, and `log.Fatal()` will execute

**Why:**
The main goroutine must block somewhere, or the program exits immediately. `ListenAndServe()` is the natural blocking point. When shutdown is triggered, it returns and allows cleanup to proceed.

**Expected result:**
Server starts and blocks on `ListenAndServe()`. Signal handler runs concurrently.

---

### Step 6: Test signal handling behavior

**File**: `cmd/server/main.go`

**What to do:**
Verify the signal handler works correctly.

**What to implement:**
- No code changes needed for this step
- This is a verification step

**Why:**
We need to ensure the signal handler correctly captures signals and initiates shutdown.

**Expected result:**
- Server starts successfully
- Ctrl+C triggers shutdown (logs show "Shutdown signal received")
- Second Ctrl+C forces immediate exit (logs show "Force shutdown signal received")

## Verification

### Compile Check
```bash
go build ./cmd/server
```
**Expected**: No compilation errors

### Basic Signal Test
```bash
# Start the server
go run cmd/server/main.go

# Press Ctrl+C once
# Expected: Log shows "Shutdown signal received signal=interrupt"
# Expected: Server begins shutdown process
```

### Force Quit Test
```bash
# Start the server
go run cmd/server/main.go

# Press Ctrl+C once (initiates graceful shutdown)
# Quickly press Ctrl+C again (within 1-2 seconds)
# Expected: Log shows "Force shutdown signal received, exiting immediately"
# Expected: Server exits immediately with code 1
```

### SIGTERM Test
```bash
# Terminal 1: Start the server
go run cmd/server/main.go &
SERVER_PID=$!

# Terminal 2: Send SIGTERM
kill -TERM $SERVER_PID
# Expected: Log shows "Shutdown signal received signal=terminated"
# Expected: Server begins shutdown process
```

### Verify Server Still Works
```bash
# Start the server
go run cmd/server/main.go

# In another terminal, test endpoints still work
curl http://localhost:8080/health
# Expected: {"status":"healthy",...}

# Stop with Ctrl+C
# Expected: Graceful shutdown initiated
```

## Common Pitfalls

### Pitfall 1: Unbuffered signal channel
**Symptom**: Signals are sometimes missed, especially under load
**Fix**: Use buffered channel with capacity 1: `make(chan os.Signal, 1)`

### Pitfall 2: Not calling signal.Notify()
**Symptom**: Signals are not captured, Ctrl+C doesn't work
**Fix**: Ensure `signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)` is called before the goroutine

### Pitfall 3: Goroutine exits immediately
**Symptom**: Signal handler doesn't work, server exits on startup
**Fix**: Ensure the goroutine blocks on receiving from the signal channel (use `<-sigChan` or `select`)

### Pitfall 4: Forgetting to create context with timeout
**Symptom**: Compilation error about undefined context
**Fix**: Create context before calling `server.Shutdown()`: `ctx, cancel := context.WithTimeout(context.Background(), cfg.ServerConfig.ShutdownTimeout)`

### Pitfall 5: Not deferring cancel function
**Symptom**: Context leak warning or resource not released
**Fix**: Add `defer cancel()` right after creating the context

### Pitfall 6: Blocking main goroutine in signal handler
**Symptom**: Server doesn't start, or starts but doesn't accept requests
**Fix**: Ensure signal handler is in a goroutine (`go func() { ... }()`), not in main flow

## Learning Resources

### Essential Reading
- [os/signal Package](https://pkg.go.dev/os/signal) - Official documentation for signal handling
- [Go Concurrency Patterns](https://go.dev/blog/pipelines) - Understanding goroutines and channels
- [Context Package](https://pkg.go.dev/context) - Understanding context for cancellation and timeouts

### Additional Resources (Optional)
- [Graceful Shutdown in Go](https://medium.com/honestbee-tw-engineer/gracefully-shutdown-in-go-http-server-5f5e6b83da5a) - Practical guide to graceful shutdown
- [Unix Signals](https://man7.org/linux/man-pages/man7/signal.7.html) - Understanding SIGINT, SIGTERM, and other signals

## Notes

**Implementation Pattern:**

The typical pattern looks like this:
```go
// Create signal channel
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

// Launch signal handler goroutine
go func() {
    // Wait for first signal
    sig := <-sigChan
    l.Info("Shutdown signal received", slog.String("signal", sig.String()))
    
    // Create shutdown context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), cfg.ServerConfig.ShutdownTimeout)
    defer cancel()
    
    // Initiate graceful shutdown
    server.Shutdown(ctx)
    
    // Wait for second signal (force quit)
    sig = <-sigChan
    l.Warn("Force shutdown signal received, exiting immediately")
    os.Exit(1)
}()

// Main goroutine blocks here
log.Fatal(server.ListenAndServe())
```

**What this enables:**
- Requirement 1.1: Server responds to SIGINT and SIGTERM
- Requirement 3.1: Logs shutdown initiation with signal type
- Requirement 5.4: Second signal forces immediate exit with code 1
- Foundation for task 4: Shutdown sequence can now be triggered
- Foundation for task 5: Database cleanup can happen after shutdown
- Foundation for task 6: Exit code handling can be implemented

**What this doesn't handle yet:**
- Shutdown timeout handling (task 4)
- Error handling from `server.Shutdown()` (task 4)
- Database cleanup (task 5)
- Proper exit codes for different scenarios (task 6)
- Logging shutdown completion (task 4)

These will be implemented in subsequent tasks.
