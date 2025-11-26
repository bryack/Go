# Step 4 Decomposition: Implement Handler Selection Logic

## Overview

This step creates a function that selects and configures the appropriate slog handler based on the desired output format (JSON or text). The handler is responsible for formatting log messages before they're written to the output destination. You'll configure the handler with options like log level filtering and source code location inclusion.

## Implementation Approach

We're building a factory function that creates slog handlers. Go's slog package provides two built-in handlers:
- `JSONHandler` - outputs structured JSON logs (machine-readable, for production)
- `TextHandler` - outputs human-readable text logs (for development)

Both handlers accept `HandlerOptions` that control behavior like minimum log level and whether to include source file/line information. The function will:
1. Create the appropriate handler options from Config
2. Select the handler type based on format
3. Return a configured handler ready to use

This is a key abstraction point - by returning `slog.Handler` interface, we can easily add custom handlers later without changing calling code.

**Key Concepts:**
- **Handler Interface**: slog.Handler defines how logs are formatted and written
- **HandlerOptions**: Configuration struct for handler behavior
- **Level Filtering**: Handlers only process logs at or above the configured level
- **Source Location**: File and line number where log was called (useful for debugging)

## Prerequisites

**Existing Code:**
- `logger/config.go` - Config struct and parseLevel function
- `logger/logger.go` - getWriter function

**Dependencies:**
- `log/slog` package (standard library)
- `strings` package (for case-insensitive format comparison)

**Knowledge Required:**
- Understanding of Go interfaces
- Familiarity with slog.Handler interface
- Understanding of log levels and filtering

## Detailed Implementation Steps

### Sub-step 4.1: Create the function signature

**File**: `logger/logger.go` (modify existing)

**What to do:**
Add a new function that creates and configures slog handlers.

**What to implement:**
- Create a function named `createHandler`
- Accept two parameters:
  - `cfg Config` - the logger configuration
  - `writer io.Writer` - where the handler should write output
- Return one value: `slog.Handler`
- Add a function comment explaining its purpose
- Import `log/slog` if not already imported

**Why:**
This function bridges configuration and handler creation. By accepting both Config and io.Writer, it has everything needed to create a fully configured handler. Returning the interface type (slog.Handler) rather than concrete types provides flexibility.

**Expected result:**
You have a function skeleton that compiles. It can return nil for now.

---

### Sub-step 4.2: Create HandlerOptions from Config

**File**: `logger/logger.go` (modify existing)

**What to do:**
Build the slog.HandlerOptions struct from your Config values.

**What to implement:**
- Create a variable of type `slog.HandlerOptions` (it's a struct)
- Set the `Level` field by calling your `parseLevel` function with `cfg.Level`
- Set the `AddSource` field to `cfg.AddSource`
- The HandlerOptions struct has other fields, but these two are the most important
- Store this in a variable (e.g., `opts`)

**Why:**
HandlerOptions controls handler behavior. The Level field determines which logs are processed (e.g., if set to INFO, DEBUG logs are ignored). AddSource adds file:line information to logs, which is invaluable for debugging but adds overhead.

**Expected result:**
You have a configured HandlerOptions struct ready to pass to handler constructors.

---

### Sub-step 4.3: Normalize the format string

**File**: `logger/logger.go` (modify existing)

**What to do:**
Convert the format string to lowercase for case-insensitive comparison.

**What to implement:**
- Use `strings.ToLower(cfg.Format)` to normalize the format
- Store the result in a variable (e.g., `format`)
- This allows users to specify "JSON", "json", or "Json" - all work the same

**Why:**
Configuration often comes from files or environment variables where case might vary. Normalizing ensures consistent behavior regardless of how users specify the format.

**Expected result:**
You have a lowercase format string ready for comparison.

---

### Sub-step 4.4: Handle JSON format case

**File**: `logger/logger.go` (modify existing)

**What to do:**
Create and return a JSONHandler when JSON format is requested.

**What to implement:**
- Check if the normalized format equals "json"
- If true, create a new JSONHandler using `slog.NewJSONHandler(writer, &opts)`
  - First parameter is the io.Writer (where to write)
  - Second parameter is pointer to HandlerOptions (use `&opts`)
- Return the handler immediately
- Note: NewJSONHandler returns `*slog.JSONHandler` which implements `slog.Handler`

**Why:**
JSON format is ideal for production because it's machine-readable. Log aggregation tools (ELK, Loki, Datadog) can parse JSON logs automatically. Each log entry is a complete JSON object with structured fields.

**Expected result:**
When format is "json", a JSONHandler is created and returned with the configured options.

---

### Sub-step 4.5: Handle text format case

**File**: `logger/logger.go` (modify existing)

**What to do:**
Create and return a TextHandler when text format is requested.

**What to implement:**
- Check if the normalized format equals "text"
- If true, create a new TextHandler using `slog.NewTextHandler(writer, &opts)`
  - First parameter is the io.Writer
  - Second parameter is pointer to HandlerOptions
- Return the handler immediately
- Note: NewTextHandler returns `*slog.TextHandler` which implements `slog.Handler`

**Why:**
Text format is human-readable and great for development. It's easier to read in a terminal than JSON. The format is key=value pairs, which is still somewhat structured but more readable than JSON.

**Expected result:**
When format is "text", a TextHandler is created and returned with the configured options.

---

### Sub-step 4.6: Add default case for invalid formats

**File**: `logger/logger.go` (modify existing)

**What to do:**
Handle the case where format is neither "json" nor "text".

**What to implement:**
- If format doesn't match "json" or "text", default to JSON
- You can either:
  - Return a JSONHandler as the default, OR
  - Print a warning and return JSONHandler
- Consider using `fmt.Printf` to warn about invalid format
- Return `slog.NewJSONHandler(writer, &opts)`

**Why:**
Invalid configuration shouldn't cause a crash. Defaulting to JSON is reasonable because it's the production-ready format. A warning helps users fix their configuration.

**Expected result:**
Invalid formats are handled gracefully with a default handler and optional warning.

---

### Sub-step 4.7: Add function documentation

**File**: `logger/logger.go` (modify existing)

**What to do:**
Add a clear comment above the createHandler function.

**What to implement:**
- Add a comment block explaining:
  - What the function does
  - What formats are supported
  - What happens with invalid formats
  - What the HandlerOptions control
- Follow Go documentation conventions (comment starts with function name)

**Why:**
Good documentation helps other developers understand the function's purpose and behavior. It also appears in IDE tooltips and godoc.

**Expected result:**
The function has clear documentation that explains its behavior.

---

## Complete Function Flow

Here's how the function should flow (without showing actual code):

1. Create HandlerOptions struct with Level and AddSource from Config
2. Normalize format string to lowercase
3. If format == "json" → return NewJSONHandler
4. If format == "text" → return NewTextHandler
5. Otherwise → warn and return NewJSONHandler (default)

## Verification

### Test JSON Handler Creation
```bash
# Create a test file
cat > test_handler.go << 'EOF'
package main

import (
    "os"
    "yourproject/logger"
    "log/slog"
)

func main() {
    cfg := logger.Config{
        Level:     "info",
        Format:    "json",
        AddSource: false,
    }
    
    handler := logger.CreateHandler(cfg, os.Stdout)
    log := slog.New(handler)
    log.Info("Testing JSON handler", slog.String("key", "value"))
}
EOF

go run test_handler.go
```
**Expected**: JSON output like `{"time":"...","level":"INFO","msg":"Testing JSON handler","key":"value"}`

### Test Text Handler Creation
```bash
# Modify test to use text format
cfg.Format = "text"
go run test_handler.go
```
**Expected**: Text output like `time=... level=INFO msg="Testing JSON handler" key=value`

### Test Level Filtering
```bash
# Test that DEBUG logs are filtered when level is INFO
cfg := logger.Config{
    Level:  "info",
    Format: "text",
}
handler := logger.CreateHandler(cfg, os.Stdout)
log := slog.New(handler)
log.Debug("This should not appear")
log.Info("This should appear")
```
**Expected**: Only the INFO message appears

### Test AddSource Option
```bash
# Test source location inclusion
cfg := logger.Config{
    Level:     "info",
    Format:    "text",
    AddSource: true,
}
handler := logger.CreateHandler(cfg, os.Stdout)
log := slog.New(handler)
log.Info("Testing source location")
```
**Expected**: Output includes source file and line number

### Test Case Insensitivity
```bash
# Test various format cases
formats := []string{"JSON", "json", "Json", "TEXT", "text", "Text"}
for _, fmt := range formats {
    cfg.Format = fmt
    handler := logger.CreateHandler(cfg, os.Stdout)
    // Should work for all cases
}
```
**Expected**: All format variations work correctly

### Test Invalid Format
```bash
# Test invalid format
cfg.Format = "invalid"
handler := logger.CreateHandler(cfg, os.Stdout)
// Should default to JSON and possibly print warning
```
**Expected**: Defaults to JSON handler, possibly with warning message

## Common Pitfalls

### Pitfall 1: Forgetting to pass pointer to HandlerOptions
**Symptom**: Compilation error or unexpected behavior
**Fix**: Use `&opts` not `opts` when calling NewJSONHandler/NewTextHandler

### Pitfall 2: Not normalizing format string
**Symptom**: "JSON" doesn't work, only "json" works
**Fix**: Use `strings.ToLower(cfg.Format)` before comparison

### Pitfall 3: Returning nil for invalid formats
**Symptom**: Panic when trying to use the handler
**Fix**: Always return a valid handler, default to JSON for invalid formats

### Pitfall 4: Not setting Level in HandlerOptions
**Symptom**: All logs appear regardless of configured level
**Fix**: Set `opts.Level = parseLevel(cfg.Level)`

### Pitfall 5: Confusing handler types
**Symptom**: Trying to return *slog.JSONHandler when function returns slog.Handler
**Fix**: Both *slog.JSONHandler and *slog.TextHandler implement slog.Handler interface, so they can be returned directly

## Learning Resources

### Essential Reading
- [slog.Handler Interface](https://pkg.go.dev/log/slog#Handler) - Understanding the Handler interface
- [slog.HandlerOptions](https://pkg.go.dev/log/slog#HandlerOptions) - Configuration options for handlers
- [slog.NewJSONHandler](https://pkg.go.dev/log/slog#NewJSONHandler) - JSON handler documentation
- [slog.NewTextHandler](https://pkg.go.dev/log/slog#NewTextHandler) - Text handler documentation

### Additional Resources
- [Go Interfaces](https://go.dev/tour/methods/9) - Understanding Go interfaces
- [Structured Logging Best Practices](https://www.honeycomb.io/blog/structured-logging-and-your-team) - Why structured logging matters

## Example Output Formats

### JSON Format
```json
{"time":"2024-11-09T10:30:45.123Z","level":"INFO","msg":"User logged in","user_id":42,"email":"user@example.com"}
```

### Text Format
```
time=2024-11-09T10:30:45.123Z level=INFO msg="User logged in" user_id=42 email=user@example.com
```

### With AddSource=true
```
time=2024-11-09T10:30:45.123Z level=INFO source=main.go:45 msg="User logged in" user_id=42
```

## Testing Your Implementation

After implementing, verify:

```bash
# Compile check
go build ./logger

# Quick test of both formats
# (Create a simple test program that tries both JSON and text)
```

## Next Steps

After completing Step 4, you'll have the handler creation logic. Step 5 will combine everything (getWriter + createHandler) into the main `New` factory function that creates complete logger instances.

The combination of Steps 3 and 4 gives you:
- Step 3: Where to write (io.Writer)
- Step 4: How to format (slog.Handler)
- Step 5: Will tie them together into a complete logger
