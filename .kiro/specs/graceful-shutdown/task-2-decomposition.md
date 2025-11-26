# Task Decomposition: Refactor Server Initialization to Use Explicit http.Server

## Overview

Refactor the server initialization in `main.go` to use an explicit `http.Server` instance instead of the convenience function `http.ListenAndServe()`. This change is a prerequisite for implementing graceful shutdown, as it provides a reference to the server that can be shut down via the `Shutdown(ctx)` method when OS signals are received.

## Implementation Approach

We're replacing the blocking `http.ListenAndServe()` call with an explicit `http.Server` struct. This is a minimal refactoring that changes how the server is started but maintains all existing functionality.

**Complexity Check:**
- **Requirements need**: Ability to call `server.Shutdown(ctx)` for graceful shutdown
- **Simple approach**: Create `http.Server` struct, call `ListenAndServe()` method (2-3 min)
- **Complex approach**: Add custom server timeouts, connection tracking, custom listeners (15+ min)
- **Recommendation**: Simple approach. Requirements only need the ability to shut down gracefully. The `http.Server` struct with default settings provides everything needed. Custom timeouts and advanced features can be added later if needed.

**Key Concepts:**
- **http.Server**: Go's HTTP server type that provides lifecycle control methods like `Shutdown()`
- **Blocking call**: `server.ListenAndServe()` blocks the main goroutine until the server stops
- **DefaultServeMux**: The global HTTP request multiplexer used by `http.Handle()` calls

## Prerequisites

**Existing Code:**
- `cmd/server/main.go` - Server initialization and route registration (line 434 has the current `http.ListenAndServe()` call)
- All route handlers are already registered using `http.Handle()` which uses `DefaultServeMux`

**Dependencies:**
- `net/http` package - Already imported

**Knowledge Required:**
- Understanding that `http.ListenAndServe(addr, nil)` is equivalent to creating an `http.Server` with that address and calling its `ListenAndServe()` method
- The `nil` handler parameter means "use `DefaultServeMux`"

## Step-by-Step Instructions

### Step 1: Create explicit http.Server instance

**File**: `cmd/server/main.go`

**What to do:**
Replace the single-line `http.ListenAndServe()` call with an explicit `http.Server` struct initialization.

**What to implement:**
- Locate the current server start line (line 434): `log.Fatal(http.ListenAndServe(address, nil))`
- Replace it with a multi-line server initialization
- Create a variable named `server` of type `*http.Server`
- Set the `Addr` field to the `address` variable (same as before)
- Set the `Handler` field to `nil` (which means use `DefaultServeMux`, same as before)
- Call `server.ListenAndServe()` method instead of the package-level function
- Keep the `log.Fatal()` wrapper for now (will be removed in later tasks)

**Why:**
The explicit `http.Server` struct gives us a reference to the server instance. This reference is essential for calling `server.Shutdown(ctx)` in later tasks. Setting `Handler` to `nil` maintains the current behavior of using `DefaultServeMux`, which all our `http.Handle()` calls register routes with.

**Expected result:**
Server starts exactly as before. All endpoints work identically. The only difference is we now have a `server` variable we can reference.

---

### Step 2: Verify server starts and all endpoints work

**File**: `cmd/server/main.go`

**What to do:**
Test that the refactored server initialization works correctly.

**What to implement:**
- No code changes needed for this step
- This is a verification step to ensure the refactoring didn't break anything

**Why:**
Since we're making a structural change, we need to verify that all existing functionality still works. The refactoring should be transparent to the application's behavior.

**Expected result:**
- Server compiles without errors
- Server starts successfully
- All endpoints respond correctly (/, /health, /tasks, /register, /login)
- Logs show the same startup messages as before

## Verification

### Compile Check
```bash
go build ./cmd/server
```
**Expected**: No compilation errors

### Server Startup Test
```bash
go run cmd/server/main.go
```
**Expected**: 
- Server starts successfully
- Logs show "HTTP Server initialized" message
- No error messages

### Endpoint Verification
```bash
# Terminal 1: Start the server
go run cmd/server/main.go

# Terminal 2: Test endpoints
curl http://localhost:8080/
# Expected: {"message":"Task Manager API","enpoints":[...]}

curl http://localhost:8080/health
# Expected: {"status":"healthy","timestamp":"...","service":"task-manager-api"}

# Test that server still blocks (Ctrl+C to stop)
# Expected: Server continues running until you press Ctrl+C
```

### Behavior Verification
```bash
# Verify server still blocks the main goroutine
go run cmd/server/main.go &
SERVER_PID=$!

# Server should be running
ps -p $SERVER_PID
# Expected: Process is running

# Kill the server
kill $SERVER_PID
```

## Common Pitfalls

### Pitfall 1: Forgetting to set Handler to nil
**Symptom**: Routes registered with `http.Handle()` don't work, 404 errors
**Fix**: Ensure `Handler: nil` is set in the `http.Server` struct. This tells the server to use `DefaultServeMux`.

### Pitfall 2: Using wrong variable name for address
**Symptom**: Compilation error about undefined variable
**Fix**: The `address` variable is already defined on line 433. Use that exact variable name in the `Addr` field.

### Pitfall 3: Removing log.Fatal too early
**Symptom**: Server exits immediately after starting
**Fix**: Keep `log.Fatal()` wrapping the `server.ListenAndServe()` call for now. It will be removed in task 3 when we implement proper signal handling.

### Pitfall 4: Trying to add server timeouts now
**Symptom**: Unnecessary complexity, potential issues with long-running requests
**Fix**: Don't add `ReadTimeout`, `WriteTimeout`, or other timeout fields yet. The default behavior is fine. These can be added later if needed.

## Learning Resources

### Essential Reading
- [http.Server Documentation](https://pkg.go.dev/net/http#Server) - Official documentation for the Server type
- [http.ListenAndServe Source](https://cs.opensource.google/go/go/+/refs/tags/go1.21.0:src/net/http/server.go;l=3221) - Shows that `ListenAndServe()` just creates a Server and calls its method

### Additional Resources (Optional)
- [Graceful Shutdown in Go](https://pkg.go.dev/net/http#Server.Shutdown) - Official documentation for the Shutdown method we'll use in later tasks
- [Go HTTP Server Lifecycle](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) - Understanding server lifecycle and timeouts

## Notes

**Why this is a separate task:**
- This refactoring is a prerequisite for tasks 3-6
- It's a minimal, low-risk change that can be tested independently
- It maintains 100% backward compatibility with existing behavior
- It provides a clear checkpoint before implementing the more complex signal handling logic

**What this enables:**
- Task 3: Signal handler can call `server.Shutdown(ctx)`
- Task 4: Graceful shutdown sequence can coordinate server shutdown
- Task 5: Database cleanup can happen after `server.Shutdown()` completes
- Task 6: Exit code handling can be implemented after shutdown completes

**What this doesn't change:**
- Route registration (still uses `http.Handle()`)
- Handler functions (no changes needed)
- Middleware (still works the same way)
- Server behavior (still blocks on `ListenAndServe()`)
- Error handling (still uses `log.Fatal()` for now)
