# Task Decomposition: Implement HTTP Logging Middleware

## Overview

This task completes the HTTP middleware that automatically logs all incoming requests and their responses with structured data. The middleware generates unique request IDs for correlation, captures timing information, handles panics gracefully, and enriches logs with contextual information like user IDs and HTTP details.

## What's Already Implemented

✅ **From logger/context.go (Task 3):**
- `GenerateRequestID()` - Creates unique request IDs with timestamp + random data
- `WithRequestID()` / `GetRequestID()` - Store/retrieve request IDs from context
- `WithLogger()` / `FromContext()` - Store/retrieve logger from context
- `WithRequestFields()` - Create logger with request ID attached

✅ **From logger/middleware.go (Partial):**
- `recoverPanic()` - Panic recovery with stack trace logging
- `LoggingMiddleware()` - Started but incomplete (missing response capture and completion logging)

✅ **From auth/middleware.go:**
- `GetUserIDFromContext()` - Retrieves authenticated user ID from context

## What Needs to Be Completed

The middleware is partially implemented but missing:
1. **Response writer wrapper** to capture status codes and bytes written
2. **Request completion logging** with status code and duration
3. **Proper log level selection** based on status code (INFO/WARN/ERROR)
4. **User ID extraction** and inclusion in logs (optional, when authenticated)
5. **Integration with server's main.go** - create logger and replace old middleware

## Simplified Task Overview

Since much of the foundation is already in place, this task is now simplified to:

### Option A: Minimal Implementation (Recommended - 10-15 minutes)
**Step 1**: Complete the LoggingMiddleware function with basic completion logging
**Step 2**: Update main.go to create logger and use new middleware

**Pros**: Simple, meets requirements (status code + duration)
**Cons**: Status code will always be logged as 200 (or you can omit it)

### Option B: Full Implementation (20-35 minutes)
**Step 1**: Create a responseWriter wrapper struct
**Step 2**: Complete the LoggingMiddleware function to use the wrapper
**Step 3**: Update main.go to create logger and use new middleware

**Pros**: Accurate status codes, bytes written
**Cons**: More code to maintain

**Recommendation**: Start with Option A. The requirements only mandate status code and duration. You can always add the wrapper later if you need accurate status codes for debugging.

## Implementation Approach

We're completing a standard Go HTTP middleware pattern. The middleware intercepts requests before they reach handlers and responses before they're sent to clients. We need a custom response writer wrapper to capture status codes and byte counts that aren't normally accessible in middleware.

**Key Concepts:**
- **Response Writer Wrapping**: Custom type that intercepts WriteHeader and Write calls
- **Request Context**: Using context.Context to propagate request IDs through the call chain
- **Panic Recovery**: Already implemented with defer/recover pattern
- **Log Level Selection**: INFO for 2xx/3xx, WARN for 4xx, ERROR for 5xx

## Prerequisites

**Existing Code:**
- `logger/logger.go` - Logger factory and slog.Logger creation ✅
- `logger/context.go` - Context helper functions (GenerateRequestID, WithRequestID, etc.) ✅
- `logger/fields.go` - Standard field name constants ✅
- `logger/middleware.go` - Partial implementation (panic recovery, request start logging) ✅
- `auth/middleware.go` - GetUserIDFromContext function ✅
- `cmd/server/main.go` - Current `logRequest` middleware that will be replaced

**Dependencies:**
- `log/slog` package (standard library) ✅
- `net/http` package (standard library) ✅
- `time` package (standard library) ✅
- `runtime/debug` package (standard library) ✅

**Knowledge Required:**
- Go HTTP middleware pattern (func(http.Handler) http.Handler)
- http.ResponseWriter interface and how to wrap it
- Context propagation in HTTP handlers (already used in auth middleware)

## Step-by-Step Instructions

### Option A: Minimal Implementation Steps

#### Step 1: Complete the LoggingMiddleware function (simple version)

**File**: `logger/middleware.go` (modify existing)

**What to do:**
Complete the middleware without a response writer wrapper.

**What to implement:**
- Add defer for panic recovery: `defer recoverPanic(logger, w, r)`
- Call the next handler: `next.ServeHTTP(w, r)`
- After handler returns:
  - Calculate duration: `duration := time.Since(start).Milliseconds()`
  - Try to extract user ID (optional): `userID, _ := auth.GetUserIDFromContext(r.Context())`
  - Log completion at INFO level with fields: request_id, method, path, duration_ms
  - Optionally include user_id if it was found (check if userID != 0)

**Why:**
This meets the requirements (Requirement 3.2: log duration) without overengineering. Status codes can be added later if needed.

**Expected result:**
Middleware logs request start and completion with duration. Simple and maintainable.

---

#### Step 2: Update main.go

(Same as Option B Step 3 below)

---

### Option B: Full Implementation Steps

#### Step 1: Implement response writer wrapper

**File**: `logger/middleware.go`

**What to do:**
Create a custom type that wraps http.ResponseWriter to capture status codes.

**What to implement:**
- Define a struct type `responseWriter` that embeds http.ResponseWriter
- Add fields: statusCode (int, default 200), wroteHeader (bool)
- Implement WriteHeader method to capture status code
- Implement Write method to set default status 200 if WriteHeader wasn't called

**Why:**
To capture accurate status codes for better observability and log level selection.

**Expected result:**
Custom responseWriter type that captures status codes.

---

#### Step 2: Complete the LoggingMiddleware function (full version)

**File**: `logger/middleware.go` (modify existing)

**What to do:**
Complete the LoggingMiddleware function with response writer wrapper.

**What to implement:**
- Create wrapped writer: `wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}`
- Add defer for panic recovery: `defer recoverPanic(logger, w, r)`
- Call handler with wrapped writer: `next.ServeHTTP(wrapped, r)`
- After handler returns:
  - Calculate duration: `duration := time.Since(start).Milliseconds()`
  - Extract user ID (optional): `userID, _ := auth.GetUserIDFromContext(r.Context())`
  - Select log level based on `wrapped.statusCode`
  - Log completion with: request_id, method, path, status_code, duration_ms
  - Include user_id if found

**Why:**
Captures accurate status codes for better observability and proper log level selection.

**Expected result:**
Each request logs completion with accurate status code and appropriate log level.

---

### Step 3 (Both Options): Update server main to use new middleware

**File**: `cmd/server/main.go` (modify existing)

**What to do:**
Replace the existing `logRequest` middleware with the new `LoggingMiddleware` and create a logger instance.

**What to implement:**
- After loading config, create a logger instance using the LogConfig:
  - Call `logger.NewLogger(&cfg.LogConfig)` directly (cfg.LogConfig is already logger.Config type)
  - Handle any errors from logger creation
- Remove or comment out the old `logRequest` function definition (lines 93-101)
- Replace all calls to `logRequest(handler)` with `logger.LoggingMiddleware(log)(handler)` where `log` is your logger instance
- Update all route registrations (there are 6 routes that use logRequest)
- Consider middleware order: logging should be outermost, then auth middleware

**Why:**
This integrates the new structured logging middleware into the application. Creating the logger from config ensures it uses the configured format, level, and output. Removing the old middleware prevents duplicate logging. Proper middleware ordering ensures request IDs are generated before other middleware runs.

**Expected result:**
Server creates a logger from config, uses new logging middleware, and removes old logRequest. All HTTP requests are logged with structured data in the configured format.

---

## Verification

### Compile Check
```bash
go build ./...
```
**Expected**: No compilation errors. The middleware should compile cleanly.

### Manual Testing

1. Start the server with default config (JSON format):
```bash
go run cmd/server/main.go
```

2. Make a request to an unauthenticated endpoint:
```bash
curl -v http://localhost:8080/health
```
**Expected logs**: Two JSON log entries:
- "HTTP request started" with request_id, method=GET, path=/health
- "HTTP request completed" with same request_id, status_code=200, duration_ms

3. Make a request that requires authentication (will fail):
```bash
curl -v http://localhost:8080/tasks
```
**Expected logs**: Request started and completed with status_code=401, logged at WARN level

4. Register a user and make an authenticated request:
```bash
# Register
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Save the token from response, then:
curl http://localhost:8080/tasks \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```
**Expected logs**: Request completed with status_code=200, includes user_id field

5. Test with text format for easier reading:
```bash
# Run with text format
go run cmd/server/main.go --log-format=text --log-level=debug
```
**Expected**: Human-readable log output with key=value pairs

6. Verify request ID correlation:
```bash
# Start server and make a request, then check logs
# The same request_id should appear in both "started" and "completed" logs
```

### Test Panic Recovery (Optional)

1. Temporarily add a panic to a handler (e.g., in healthHandler):
```go
func healthHandler(w http.ResponseWriter, r *http.Request) {
    panic("test panic")
    // ... rest of handler
}
```

2. Make a request to that handler:
```bash
curl -v http://localhost:8080/health
```
**Expected**: 
- Server logs ERROR with panic message and stack trace
- Client receives 500 Internal Server Error
- Server continues running (doesn't crash)

3. Remove the test panic

## Common Pitfalls

### Pitfall 1: Not checking if WriteHeader was already called
**Symptom**: Error message "http: superfluous response.WriteHeader call" in logs
**Fix**: In the responseWriter wrapper's Write method, check if WriteHeader was called. If not, set status to 200 before calling the wrapped Write. Track this with a boolean field `wroteHeader`.

### Pitfall 2: Forgetting to use the wrapped response writer
**Symptom**: Status code is always 0 or 200 in logs, even for errors
**Fix**: Ensure you pass the wrapped responseWriter to `next.ServeHTTP()`, not the original http.ResponseWriter. The wrapper must intercept all calls.

### Pitfall 3: Request context not propagated
**Symptom**: Request ID is not available in downstream handlers or logs
**Fix**: The current code already does `r = r.WithContext(ctx)` which is correct. Make sure you pass this updated `r` to `next.ServeHTTP()`.

### Pitfall 4: Forgetting to import auth package
**Symptom**: Compilation error when trying to use `auth.GetUserIDFromContext`
**Fix**: Add `"myproject/auth"` to the imports in middleware.go

### Pitfall 5: Duration calculation is inaccurate
**Symptom**: Duration is always 0 or very small
**Fix**: The `start` variable is already declared. Make sure you calculate duration AFTER calling `next.ServeHTTP()` using `time.Since(start).Milliseconds()`

## Learning Resources

### Essential Reading
- [Writing HTTP Middleware in Go](https://www.alexedwards.net/blog/making-and-using-middleware) - Comprehensive guide to the middleware pattern
- [Go http.ResponseWriter Interface](https://pkg.go.dev/net/http#ResponseWriter) - Understanding what methods need to be implemented
- [slog Structured Logging](https://pkg.go.dev/log/slog) - Official documentation for structured logging

### Additional Resources (Optional)
- [Panic and Recover in Go](https://go.dev/blog/defer-panic-and-recover) - Deep dive into panic recovery patterns
- [Context in Go](https://go.dev/blog/context) - Understanding context propagation
