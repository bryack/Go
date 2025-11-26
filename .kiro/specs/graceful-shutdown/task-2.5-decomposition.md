# Task Decomposition: Configure HTTP Server Timeouts for Reliable Shutdown

## Overview

Add HTTP server timeouts to the explicit `http.Server` configuration to ensure graceful shutdown can complete reliably. Without these timeouts, HTTP keep-alive connections can remain open indefinitely, preventing `server.Shutdown()` from completing within the shutdown timeout period. This task is critical for the graceful shutdown feature to work as designed.

## Implementation Approach

We're adding three timeout fields to the existing `http.Server` struct initialization: `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`. The most critical is `IdleTimeout`, which forces idle keep-alive connections to close after a specified duration.

**Complexity Check:**
- **Requirements need**: Ensure server.Shutdown() completes within shutdown timeout (Requirements 1.2, 1.3, 1.4)
- **Simple approach**: Add three timeout fields to existing http.Server struct (2-3 min)
- **Complex approach**: Make timeouts configurable, add validation, create separate timeout config (20+ min)
- **Recommendation**: Simple approach with hardcoded values. The requirements don't ask for configurable HTTP timeouts, only shutdown timeout. These are implementation details that ensure the feature works. Hardcoded values (15s read/write, 60s idle) are industry-standard and work for 99% of use cases.

**Key Concepts:**
- **ReadTimeout**: Prevents slow-read attacks where clients send data very slowly
- **WriteTimeout**: Prevents slow-write issues where responses take too long to send
- **IdleTimeout**: Closes idle keep-alive connections, enabling graceful shutdown
- **HTTP Keep-Alive**: HTTP/1.1 feature where connections stay open for reuse

## Prerequisites

**Existing Code:**
- `cmd/server/main.go` - Explicit `http.Server` initialization (task 2 complete, around line 404-408)
- Server is already configured with `Addr` and `Handler` fields

**Dependencies:**
- `time` package - Already imported
- `net/http` package - Already imported

**Knowledge Required:**
- Understanding that HTTP keep-alive keeps connections open between requests
- `IdleTimeout` is the key to allowing graceful shutdown to complete
- These timeouts don't affect request processing, only connection lifecycle

## Step-by-Step Instructions

### Step 1: Add ReadTimeout to http.Server

**File**: `cmd/server/main.go`

**What to do:**
Add the `ReadTimeout` field to the existing `http.Server` struct initialization.

**What to implement:**
- Locate the `http.Server` struct initialization (around line 404-408)
- Add a new field: `ReadTimeout: 15 * time.Second`
- Place it after the `Handler` field
- Use 15 seconds as the value (industry standard for API servers)

**Why:**
`ReadTimeout` limits how long the server waits to read the entire request (headers + body). This prevents slow-read attacks where malicious clients send data very slowly to tie up server resources. 15 seconds is sufficient for normal API requests while protecting against abuse.

**Expected result:**
Code compiles. Server starts successfully. Requests still work normally.

---

### Step 2: Add WriteTimeout to http.Server

**File**: `cmd/server/main.go`

**What to do:**
Add the `WriteTimeout` field to the `http.Server` struct.

**What to implement:**
- Add a new field: `WriteTimeout: 15 * time.Second`
- Place it after the `ReadTimeout` field
- Use 15 seconds as the value

**Why:**
`WriteTimeout` limits how long the server takes to write the response. This prevents issues where response writing hangs or takes too long. For API servers returning JSON, 15 seconds is more than sufficient.

**Expected result:**
Code compiles. Server starts successfully. Responses are sent normally.

---

### Step 3: Add IdleTimeout to http.Server

**File**: `cmd/server/main.go`

**What to do:**
Add the `IdleTimeout` field to the `http.Server` struct. This is the **critical timeout** for graceful shutdown.

**What to implement:**
- Add a new field: `IdleTimeout: 20 * time.Second`
- Place it after the `WriteTimeout` field
- Use 20 seconds as the value (shorter than the 30s shutdown timeout)
- **Note**: You could use 60s if you increase shutdown timeout to 90s in config

**Why:**
`IdleTimeout` is the **key to making graceful shutdown work**. Without it, HTTP keep-alive connections stay open indefinitely. When `server.Shutdown()` is called, it waits for all connections to close. Without `IdleTimeout`, keep-alive connections never close, causing shutdown to hang until the shutdown context timeout expires.

60 seconds is chosen because:
- It's long enough to allow connection reuse for multiple requests (good for performance)
- It's a common industry standard for API servers
- **Important**: This is LONGER than the default shutdown timeout (30s), which means you should either:
  - Reduce `IdleTimeout` to 20s (recommended for faster shutdown)
  - Or increase shutdown timeout in config to 90s (if you want longer connection reuse)

**Expected result:**
Code compiles. Server starts successfully. Idle connections close after 60 seconds. Graceful shutdown completes quickly.

---

### Step 4: Verify graceful shutdown works

**File**: `cmd/server/main.go`

**What to do:**
Test that graceful shutdown now completes successfully without hanging.

**What to implement:**
- No code changes needed for this step
- This is a verification step

**Why:**
We need to verify that the timeouts solve the hanging shutdown issue. With `IdleTimeout` configured, `server.Shutdown()` should complete within a few milliseconds (or seconds if requests are in-flight), not hang until the shutdown timeout.

**Expected result:**
- Server starts successfully
- Requests work normally
- Ctrl+C triggers shutdown that completes in < 1 second (if no active requests)
- Logs show "Server shutdown completed successfully" with short duration

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
- No timeout-related errors
- Logs show normal startup messages

### Request Handling Test
```bash
# Terminal 1: Start the server
go run cmd/server/main.go

# Terminal 2: Make requests
curl http://localhost:8080/health
# Expected: {"status":"healthy",...}

curl http://localhost:8080/
# Expected: {"message":"Task Manager API",...}

# Make multiple requests to test keep-alive
for i in {1..5}; do curl http://localhost:8080/health; done
# Expected: All requests succeed
```

### Graceful Shutdown Test (The Critical Test)
```bash
# Terminal 1: Start the server
go run cmd/server/main.go

# Terminal 2: Make a request to establish a connection
curl http://localhost:8080/health

# Terminal 1: Press Ctrl+C
# Expected logs:
# - "Shutdown signal received signal=interrupt"
# - "Server shutdown completed successfully duration=<very short time>"
# - NOT "Graceful shutdown timed out"

# Check exit code
echo $?
# Expected: 0
```

### Shutdown with Active Request Test
```bash
# Terminal 1: Start the server
go run cmd/server/main.go

# Terminal 2: Make a request that takes time (if you have one)
# Or just send shutdown immediately after a request

# Terminal 1: Press Ctrl+C while request is processing
# Expected: Request completes, then shutdown succeeds quickly
```

### Idle Connection Test
```bash
# Terminal 1: Start the server
go run cmd/server/main.go

# Terminal 2: Open a connection and leave it idle
curl http://localhost:8080/health

# Wait 65 seconds (longer than IdleTimeout)

# Terminal 2: Try another request on same connection
curl http://localhost:8080/health
# Expected: New connection is established (old one was closed)

# Terminal 1: Press Ctrl+C
# Expected: Shutdown completes quickly
```

## Common Pitfalls

### Pitfall 1: Setting IdleTimeout longer than shutdown timeout
**Symptom**: Shutdown still times out after 30 seconds
**Fix**: `IdleTimeout` (60s) is LONGER than shutdown timeout (30s), which is problematic. Idle connections won't close until 60s, but shutdown only waits 30s. Solutions:
- **Recommended**: Change `IdleTimeout: 20 * time.Second` (shorter than 30s shutdown timeout)
- **Alternative**: Increase shutdown timeout to 90s in config.yaml: `shutdown_timeout: "90s"`

### Pitfall 2: Setting timeouts too short
**Symptom**: Legitimate requests fail with timeout errors
**Fix**: 15s for read/write is generous for API requests. If you have long-running operations, consider increasing `WriteTimeout` or handling them asynchronously.

### Pitfall 3: Forgetting to import time package
**Symptom**: Compilation error about undefined `time`
**Fix**: Ensure `"time"` is in the imports (it should already be there)

### Pitfall 4: Not testing with actual HTTP client
**Symptom**: Shutdown seems to work in tests but hangs in production
**Fix**: Test with real HTTP clients (curl, Postman, browsers) that use keep-alive connections

## Learning Resources

### Essential Reading
- [http.Server Timeouts](https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/) - Comprehensive guide to Go HTTP server timeouts
- [http.Server Documentation](https://pkg.go.dev/net/http#Server) - Official documentation for timeout fields
- [HTTP Keep-Alive](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Keep-Alive) - Understanding HTTP keep-alive connections

### Additional Resources (Optional)
- [Graceful Shutdown Best Practices](https://medium.com/honestbee-tw-engineer/gracefully-shutdown-in-go-http-server-5f5e6b83da5a) - Why timeouts matter for shutdown
- [Go HTTP Server Patterns](https://www.alexedwards.net/blog/configuring-timeouts) - Production-ready server configuration

## Notes

**Why These Specific Values:**

- **ReadTimeout: 15s** - Most API requests complete in < 5s. 15s provides buffer for slow networks while preventing abuse.
- **WriteTimeout: 15s** - JSON responses are typically small and fast. 15s is generous for API responses.
- **IdleTimeout: 60s** - Balances connection reuse (good for performance) with timely connection closure (good for shutdown). **WARNING**: This is longer than the default 30s shutdown timeout, which can cause shutdown timeouts. Consider using 20s instead, or increase shutdown timeout to 90s.

**What This Fixes:**

Before this task:
```
1. Client makes request to server
2. Request completes, but HTTP keep-alive keeps connection open
3. Operator sends SIGTERM
4. server.Shutdown() waits for connection to close
5. Connection never closes (no timeout)
6. Shutdown hangs until shutdown context timeout (30s)
7. Logs show "Graceful shutdown timed out"
```

After this task:
```
1. Client makes request to server
2. Request completes, connection stays open for reuse
3. After 60s of idle time, IdleTimeout closes the connection
4. Operator sends SIGTERM
5. server.Shutdown() waits for connections to close
6. No idle connections exist (or they close within 60s)
7. Shutdown completes in milliseconds
8. Logs show "Server shutdown completed successfully"
```

**Production Considerations:**

**Recommended Configuration:**
```go
ReadTimeout:  15 * time.Second,
WriteTimeout: 15 * time.Second,
IdleTimeout:  20 * time.Second,  // Shorter than 30s shutdown timeout
```

If you need different values:
- **Long-running requests**: Increase `WriteTimeout` or use streaming responses
- **Slow clients**: Increase `ReadTimeout` (but be aware of DoS risks)
- **More connection reuse**: Increase `IdleTimeout` to 60s AND increase shutdown timeout to 90s in config
- **Faster shutdown**: Keep `IdleTimeout` at 20s or reduce to 10s

**Critical Rule**: `IdleTimeout` MUST be shorter than shutdown timeout, or shutdown will always time out waiting for idle connections to close.
