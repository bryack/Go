# Task Decomposition: Implement Context Helpers for Request Correlation

## Overview

This task creates a context management system for storing and retrieving request-scoped data like request IDs, trace IDs, and logger instances. You'll implement functions that attach this data to Go's context.Context and retrieve it later, enabling log correlation across the entire request lifecycle. This is essential for distributed tracing and debugging complex request flows.

## Implementation Approach

We're leveraging Go's `context.Context` package to propagate request-scoped data through the call stack. Context is Go's standard way to pass request-scoped values, cancellation signals, and deadlines across API boundaries and between goroutines.

The approach uses context keys (unexported types) to store values safely without collisions. Each piece of data (request ID, trace ID, logger) gets its own key type and pair of functions (With* to store, Get* to retrieve). This pattern is idiomatic Go and prevents key collisions between packages.

The request ID generation uses a simple but effective approach: combining timestamp with random data to ensure uniqueness across distributed systems. This ID will appear in every log entry related to a request, making it easy to trace the entire request flow.

**Key Concepts:**
- **Context Propagation**: Passing data through function calls without explicit parameters
- **Context Keys**: Type-safe keys that prevent collisions between packages
- **Request Correlation**: Linking related log entries using unique identifiers
- **Distributed Tracing**: Following requests across multiple services using trace IDs

## Prerequisites

**Existing Code:**
- `logger/logger.go` - Logger factory functions
- `logger/fields.go` - Field name constants (FieldRequestID, FieldTraceID, etc.)

**Dependencies:**
- `context` package (standard library)
- `log/slog` package (standard library)
- `crypto/rand` package (standard library) - for random ID generation
- `encoding/hex` package (standard library) - for encoding random bytes
- `fmt` package (standard library)
- `time` package (standard library) - for timestamp in IDs

**Knowledge Required:**
- Understanding of Go's context.Context
- Context value storage and retrieval patterns
- Understanding of request correlation in distributed systems
- Basic understanding of unique ID generation

## Step-by-Step Instructions

### Step 1: Create context.go file with package structure

**File**: `logger/context.go` (create new file)

**What to do:**
Create a new file to hold context helper functions for request correlation.

**What to implement:**
- Add package declaration: `package logger`
- Import required packages: `context`, `log/slog`, `crypto/rand`, `encoding/hex`, `fmt`, `time`
- Add a package-level comment explaining:
  - Context helpers for request-scoped data
  - Functions for storing/retrieving request IDs, trace IDs, and loggers
  - Used for log correlation across request lifecycle
- Keep the comment concise (2-3 sentences)

**Why:**
Separating context helpers into their own file keeps the code organized. Context management is a distinct concern from logger creation and field definitions, so it deserves its own file.

**Expected result:**
You have a `logger/context.go` file with package declaration and imports. The file compiles without errors.

---

### Step 2: Define context key types

**File**: `logger/context.go` (modify existing)

**What to do:**
Create unexported types to use as context keys.

**What to implement:**
- Define three unexported (lowercase) types:
  - `type contextKey int` - base type for all keys
- Define three constants of type contextKey:
  - `requestIDKey contextKey = iota` - key for request ID
  - `traceIDKey` - key for trace ID
  - `loggerKey` - key for logger instance
- Add a comment explaining why these are unexported (prevents key collisions)

**Why:**
Using unexported types as context keys is a Go best practice. It prevents other packages from accidentally using the same key value and overwriting your data. Each key is unique to this package because the type itself is unexported.

**Expected result:**
You have three context key constants that can be used to store/retrieve values safely.

---

### Step 3: Implement request ID generation

**File**: `logger/context.go` (modify existing)

**What to do:**
Create a function that generates unique request IDs.

**What to implement:**
- Create a function named `generateRequestID`
- Return type: `string`
- No parameters needed
- Implementation approach:
  - Get current Unix timestamp in milliseconds
  - Generate 8 random bytes using `crypto/rand.Read`
  - Encode random bytes to hex string
  - Combine timestamp and random hex: "req_" + timestamp + "_" + hex
  - Handle errors from crypto/rand.Read (return a fallback ID if needed)
- Example output: "req_1699459845123_a1b2c3d4e5f6g7h8"
- Add function documentation explaining the format and uniqueness guarantee

**Why:**
Request IDs must be unique across all requests in all instances of your service. Combining timestamp (for temporal ordering) with random data (for uniqueness) is a simple, effective approach. The "req_" prefix makes it easy to identify request IDs in logs. Crypto/rand provides cryptographically secure randomness.

**Expected result:**
You can generate unique request IDs. Each call returns a different ID. The format is consistent and recognizable.

---

### Step 4: Implement WithRequestID function

**File**: `logger/context.go` (modify existing)

**What to do:**
Create a function that stores a request ID in context.

**What to implement:**
- Create a function named `WithRequestID`
- Accept two parameters:
  - `ctx context.Context` - the parent context
  - `requestID string` - the request ID to store
- Return: `context.Context` - new context with request ID
- Implementation:
  - Use `context.WithValue(ctx, requestIDKey, requestID)` to store the ID
  - Return the new context
- Add function documentation explaining:
  - What it does (stores request ID in context)
  - When to use it (at the start of request handling)
  - That it returns a new context (doesn't modify the original)

**Why:**
Context values are immutable - WithValue creates a new context that wraps the parent. This function provides a clean API for storing request IDs without exposing the internal key type. It's the standard Go pattern for context value storage.

**Expected result:**
You can store a request ID in a context. The original context is unchanged. The returned context contains the request ID.

---

### Step 5: Implement GetRequestID function

**File**: `logger/context.go` (modify existing)

**What to do:**
Create a function that retrieves a request ID from context.

**What to implement:**
- Create a function named `GetRequestID`
- Accept one parameter: `ctx context.Context`
- Return: `string` - the request ID, or empty string if not found
- Implementation:
  - Use `ctx.Value(requestIDKey)` to retrieve the value
  - Type assert to string: `id, ok := ctx.Value(requestIDKey).(string)`
  - If ok is true, return the id
  - If ok is false (not found or wrong type), return empty string ""
- Add function documentation explaining:
  - What it does (retrieves request ID from context)
  - Return value (empty string if not found)
  - When to use it (when logging within request handlers)

**Why:**
Context.Value returns interface{}, so we need to type assert to get the actual string. Returning empty string for missing values is a safe default - it won't cause panics and is easy to check. This function provides a clean API for retrieving request IDs.

**Expected result:**
You can retrieve request IDs from context. Missing IDs return empty string without panicking.

---

### Step 6: Implement WithTraceID and GetTraceID functions

**File**: `logger/context.go` (modify existing)

**What to do:**
Create functions for storing and retrieving trace IDs (similar to request ID functions).

**What to implement:**
- Create `WithTraceID` function:
  - Accept `ctx context.Context` and `traceID string`
  - Return `context.Context`
  - Use `context.WithValue(ctx, traceIDKey, traceID)`
  - Add documentation explaining trace IDs are for distributed tracing
- Create `GetTraceID` function:
  - Accept `ctx context.Context`
  - Return `string`
  - Retrieve and type assert like GetRequestID
  - Return empty string if not found
  - Add documentation

**Why:**
Trace IDs are used in distributed systems to follow requests across multiple services. They're typically generated by API gateways or load balancers and propagated through headers. Storing them in context makes them available for logging throughout the request lifecycle.

**Expected result:**
You can store and retrieve trace IDs from context using the same pattern as request IDs.

---

### Step 7: Implement WithLogger function

**File**: `logger/context.go` (modify existing)

**What to do:**
Create a function that stores a logger instance in context.

**What to implement:**
- Create a function named `WithLogger`
- Accept two parameters:
  - `ctx context.Context` - the parent context
  - `logger *slog.Logger` - the logger to store
- Return: `context.Context` - new context with logger
- Implementation:
  - Use `context.WithValue(ctx, loggerKey, logger)` to store the logger
  - Return the new context
- Add function documentation explaining:
  - What it does (stores logger in context)
  - Why this is useful (logger with request-specific fields can be propagated)
  - When to use it (after creating a logger with request ID attached)

**Why:**
Storing the logger in context allows you to create a logger with request-specific fields (like request ID) once at the start of request handling, then retrieve it anywhere in the call stack. This avoids passing logger as a parameter to every function.

**Expected result:**
You can store a logger instance in context. The logger can be retrieved later in the call stack.

---

### Step 8: Implement FromContext function

**File**: `logger/context.go` (modify existing)

**What to do:**
Create a function that retrieves a logger from context with a fallback.

**What to implement:**
- Create a function named `FromContext`
- Accept one parameter: `ctx context.Context`
- Return: `*slog.Logger` - never nil
- Implementation:
  - Use `ctx.Value(loggerKey)` to retrieve the logger
  - Type assert to `*slog.Logger`
  - If found and type assertion succeeds, return the logger
  - If not found or wrong type, return `slog.Default()` (the default global logger)
- Add function documentation explaining:
  - What it does (retrieves logger from context)
  - Fallback behavior (returns default logger if not found)
  - That it never returns nil (safe to use without checking)

**Why:**
Always returning a valid logger (never nil) makes this function safe to use without error checking. The fallback to slog.Default() ensures logging always works even if no logger was stored in context. This is a defensive programming practice that prevents panics.

**Expected result:**
You can retrieve a logger from context. If no logger is stored, you get the default logger. The function never returns nil.

---

### Step 9: Implement GenerateRequestID exported function

**File**: `logger/context.go` (modify existing)

**What to do:**
Create an exported version of the request ID generator for external use.

**What to implement:**
- Create a function named `GenerateRequestID` (exported, starts with capital G)
- No parameters
- Return: `string` - a unique request ID
- Implementation:
  - Simply call and return the result of `generateRequestID()`
- Add function documentation explaining:
  - What it does (generates unique request ID)
  - The format of generated IDs
  - When to use it (at the start of request handling, in middleware)
  - That IDs are unique across distributed systems

**Why:**
The exported function provides a public API for generating request IDs. This will be used by the HTTP middleware to generate IDs for incoming requests. Keeping the internal implementation separate allows you to change the generation algorithm without breaking the public API.

**Expected result:**
External packages can generate request IDs by calling `logger.GenerateRequestID()`.

---

### Step 10: Add helper function for creating request-scoped logger

**File**: `logger/context.go` (modify existing)

**What to do:**
Create a convenience function that creates a logger with request ID already attached.

**What to implement:**
- Create a function named `WithRequestFields`
- Accept two parameters:
  - `logger *slog.Logger` - base logger
  - `requestID string` - request ID to attach
- Return: `*slog.Logger` - new logger with request ID field
- Implementation:
  - Use `logger.With(slog.String(FieldRequestID, requestID))` to create new logger
  - Return the new logger
- Add function documentation explaining:
  - What it does (creates logger with request ID attached)
  - That all logs from returned logger will include request ID
  - When to use it (in middleware after generating request ID)

**Why:**
This helper simplifies a common pattern: creating a logger with request ID attached. Instead of manually calling logger.With() everywhere, you use this helper. The returned logger automatically includes the request ID in every log entry, ensuring correlation.

**Expected result:**
You can create a logger with request ID attached. All logs from this logger include the request ID field.

---

### Step 11: Add package documentation

**File**: `logger/context.go` (modify existing)

**What to do:**
Add comprehensive documentation at the top of the file.

**What to implement:**
- Add a comment block before the package declaration explaining:
  - Purpose: Context helpers for request correlation
  - What's provided: Functions for storing/retrieving request IDs, trace IDs, loggers
  - Usage pattern: Generate ID → Store in context → Retrieve when logging
  - Benefits: Log correlation, distributed tracing support
- Include a brief usage example in the comment
- Keep it concise but informative (4-6 sentences)

**Why:**
Good documentation helps developers understand how to use the package. It appears in godoc and IDE tooltips. A usage example is especially valuable for showing the intended workflow.

**Expected result:**
The file has clear documentation explaining its purpose and usage pattern.

---

## Complete File Structure

Here's what the file structure should look like (without showing actual code):

```
logger/context.go
├── Package comment with usage example
├── Package declaration
├── Imports (context, log/slog, crypto/rand, encoding/hex, fmt, time)
├── Context key type and constants
│   ├── type contextKey int
│   ├── requestIDKey
│   ├── traceIDKey
│   └── loggerKey
├── generateRequestID (internal)
│   ├── Get timestamp
│   ├── Generate random bytes
│   ├── Combine into ID string
│   └── Handle errors
├── GenerateRequestID (exported)
│   └── Calls generateRequestID
├── WithRequestID
│   └── Stores request ID in context
├── GetRequestID
│   └── Retrieves request ID from context
├── WithTraceID
│   └── Stores trace ID in context
├── GetTraceID
│   └── Retrieves trace ID from context
├── WithLogger
│   └── Stores logger in context
├── FromContext
│   └── Retrieves logger from context (with fallback)
└── WithRequestFields
    └── Creates logger with request ID attached
```

## Verification

### Compile Check
```bash
go build ./logger
```
**Expected**: No compilation errors

### Test Request ID Generation
```bash
# Create a test file
cat > test_context.go << 'EOF'
package main

import (
    "fmt"
    "yourproject/logger"
)

func main() {
    // Generate multiple IDs
    for i := 0; i < 5; i++ {
        id := logger.GenerateRequestID()
        fmt.Println("Generated ID:", id)
    }
}
EOF

go run test_context.go
```

**Expected output**: Five unique request IDs with format "req_timestamp_randomhex"

### Test Context Storage and Retrieval
```bash
# Create a test file
cat > test_context_storage.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "yourproject/logger"
)

func main() {
    // Create base context
    ctx := context.Background()
    
    // Store request ID
    requestID := logger.GenerateRequestID()
    ctx = logger.WithRequestID(ctx, requestID)
    
    // Retrieve request ID
    retrieved := logger.GetRequestID(ctx)
    fmt.Printf("Stored: %s\n", requestID)
    fmt.Printf("Retrieved: %s\n", retrieved)
    fmt.Printf("Match: %v\n", requestID == retrieved)
    
    // Test missing request ID
    emptyCtx := context.Background()
    missing := logger.GetRequestID(emptyCtx)
    fmt.Printf("Missing ID: '%s' (empty: %v)\n", missing, missing == "")
}
EOF

go run test_context_storage.go
```

**Expected output**:
```
Stored: req_1699459845123_a1b2c3d4e5f6g7h8
Retrieved: req_1699459845123_a1b2c3d4e5f6g7h8
Match: true
Missing ID: '' (empty: true)
```

### Test Trace ID Functions
```bash
# Create a test file
cat > test_trace.go << 'EOF'
package main

import (
    "context"
    "fmt"
    "yourproject/logger"
)

func main() {
    ctx := context.Background()
    
    // Store and retrieve trace ID
    traceID := "trace-abc-123-def-456"
    ctx = logger.WithTraceID(ctx, traceID)
    retrieved := logger.GetTraceID(ctx)
    
    fmt.Printf("Stored: %s\n", traceID)
    fmt.Printf("Retrieved: %s\n", retrieved)
    fmt.Printf("Match: %v\n", traceID == retrieved)
}
EOF

go run test_trace.go
```

**Expected output**: Trace ID is stored and retrieved correctly

### Test Logger Storage and Retrieval
```bash
# Create a test file
cat > test_logger_context.go << 'EOF'
package main

import (
    "context"
    "yourproject/logger"
)

func main() {
    ctx := context.Background()
    
    // Create a logger
    log := logger.NewDefault()
    
    // Store in context
    ctx = logger.WithLogger(ctx, log)
    
    // Retrieve from context
    retrieved := logger.FromContext(ctx)
    retrieved.Info("Testing logger from context")
    
    // Test fallback (no logger in context)
    emptyCtx := context.Background()
    fallback := logger.FromContext(emptyCtx)
    fallback.Info("Testing fallback logger")
}
EOF

go run test_logger_context.go
```

**Expected**: Both log messages appear (first from stored logger, second from default logger)

### Test WithRequestFields Helper
```bash
# Create a test file
cat > test_request_fields.go << 'EOF'
package main

import (
    "yourproject/logger"
)

func main() {
    // Create base logger
    log := logger.NewDefault()
    
    // Generate request ID
    requestID := logger.GenerateRequestID()
    
    // Create logger with request ID
    requestLog := logger.WithRequestFields(log, requestID)
    
    // All logs from requestLog will include request_id
    requestLog.Info("Processing request")
    requestLog.Info("Request completed")
}
EOF

go run test_request_fields.go
```

**Expected**: Both log messages include the request_id field automatically

### Integration Test: Full Request Flow
```bash
# Create a test file simulating request handling
cat > test_full_flow.go << 'EOF'
package main

import (
    "context"
    "yourproject/logger"
)

func handleRequest(ctx context.Context) {
    // Get logger from context
    log := logger.FromContext(ctx)
    
    // Get request ID for additional logging
    requestID := logger.GetRequestID(ctx)
    
    log.Info("Handling request", "step", "start")
    
    // Simulate calling another function
    processTask(ctx)
    
    log.Info("Request completed", "step", "end")
}

func processTask(ctx context.Context) {
    // Logger and request ID are available here too
    log := logger.FromContext(ctx)
    log.Info("Processing task", "step", "middle")
}

func main() {
    // Simulate middleware: generate request ID
    requestID := logger.GenerateRequestID()
    
    // Create logger with request ID
    log := logger.NewDefault()
    requestLog := logger.WithRequestFields(log, requestID)
    
    // Create context with request ID and logger
    ctx := context.Background()
    ctx = logger.WithRequestID(ctx, requestID)
    ctx = logger.WithLogger(ctx, requestLog)
    
    // Handle request
    handleRequest(ctx)
}
EOF

go run test_full_flow.go
```

**Expected**: All three log messages include the same request_id field, demonstrating correlation

## Common Pitfalls

### Pitfall 1: Using exported types as context keys
**Symptom**: Key collisions with other packages, values get overwritten
**Fix**: Always use unexported types for context keys: `type contextKey int` not `type ContextKey int`

### Pitfall 2: Not handling missing values from context
**Symptom**: Panic when type asserting nil values
**Fix**: Always check the ok value from type assertion: `id, ok := ctx.Value(key).(string); if !ok { return "" }`

### Pitfall 3: Modifying the original context
**Symptom**: Confusion about why context values aren't available
**Fix**: Remember context.WithValue returns a NEW context - you must use the returned value: `ctx = logger.WithRequestID(ctx, id)`

### Pitfall 4: Returning nil from FromContext
**Symptom**: Panic when trying to use the logger
**Fix**: Always return a valid logger, use slog.Default() as fallback: `if logger == nil { return slog.Default() }`

### Pitfall 5: Not generating unique request IDs
**Symptom**: Multiple requests have the same ID, can't correlate logs
**Fix**: Use crypto/rand for randomness and include timestamp: `crypto/rand.Read(bytes)` not `math/rand`

### Pitfall 6: Forgetting to propagate context
**Symptom**: Request ID or logger not available in nested functions
**Fix**: Always pass context as first parameter to functions: `func processTask(ctx context.Context)`

## Learning Resources

### Essential Reading
- [Go Context Package](https://pkg.go.dev/context) - Official documentation for context
- [Go Blog: Context](https://go.dev/blog/context) - Understanding context patterns and best practices
- [Context Keys Best Practices](https://go.dev/blog/context-keys) - How to use context keys safely

### Additional Resources
- [Distributed Tracing](https://opentelemetry.io/docs/concepts/observability-primer/#distributed-traces) - Understanding trace IDs and spans
- [Request Correlation](https://www.honeycomb.io/blog/request-correlation-with-unique-ids) - Why request IDs matter
- [crypto/rand vs math/rand](https://pkg.go.dev/crypto/rand) - Why to use crypto/rand for IDs

## Real-World Context

### Why Request IDs Matter

In production systems, a single user action might trigger:
- HTTP request to API
- Database queries
- Cache lookups
- External API calls
- Background job creation

Without request IDs, you have hundreds of log entries with no way to know which ones are related. With request IDs, you can filter logs to see the entire flow of a single request.

### Distributed Tracing Integration

Trace IDs are typically generated by:
- API gateways (AWS API Gateway, Kong)
- Load balancers (ALB, NGINX)
- Service meshes (Istio, Linkerd)
- APM tools (Datadog, New Relic)

They're propagated through HTTP headers (X-Trace-Id, X-Request-Id) and stored in context for logging. This allows you to trace requests across multiple services.

### Context Propagation Pattern

The standard Go pattern for request handling:

1. Middleware generates request ID
2. Middleware creates logger with request ID
3. Middleware stores both in context
4. Handler retrieves logger from context
5. Handler passes context to service layer
6. Service layer retrieves logger from context
7. All logs include request ID automatically

This pattern is used by major Go projects like Kubernetes, Docker, and Prometheus.

## Testing Your Implementation

After implementing, verify:

1. **Request ID generation**: IDs are unique and well-formatted
2. **Context storage**: Values can be stored and retrieved
3. **Type safety**: Wrong types don't cause panics
4. **Fallback behavior**: Missing values return safe defaults
5. **Logger propagation**: Logger with request ID works through call stack

## Next Steps

After completing this task, you'll have a complete context management system. Task 4 will extend the configuration system to support logging configuration, and Task 5 will implement the HTTP middleware that uses these context helpers to generate request IDs and propagate them through requests.

The combination of Tasks 1-3 gives you:
- Task 1: Logger factory and configuration
- Task 2: Standardized field names and masking
- Task 3: Context management for request correlation
- Task 4: Will add configuration integration
- Task 5: Will add HTTP middleware that ties everything together
