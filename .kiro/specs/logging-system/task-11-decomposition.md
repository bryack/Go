# Task Decomposition: Add CLI Logging Support

## Overview

Add minimal, optional structured logging to the CLI for debugging internal errors and HTTP communication issues. The CLI will remain clean and user-friendly by default, with logging only enabled via a `--debug` flag. This provides troubleshooting capability without cluttering the user experience.

## Complexity Check & Simplification

**Requirements Analysis:**
- Requirement 7.1: CLI outputs user messages to stdout (✅ already done)
- Requirement 7.2: CLI logs internal errors to stderr with structured format
- Requirement 7.3: CLI supports debug flag for detailed logging
- Requirement 7.4: CLI doesn't log HTTP details at INFO level (✅ already done)
- Requirement 7.5: CLI logs authentication errors at ERROR level

**Current State:**
- CLI has excellent error handling with `NetworkError` and `APIError` types
- User-facing messages are clean with emojis and helpful context
- Only one `log.Fatalf()` call exists (config loading)
- HTTP client already provides detailed error information

**Simple Approach (Recommended):**
- Add optional logger that writes to stderr (disabled by default)
- Add `--debug` flag to enable DEBUG level logging
- Log only: config errors, HTTP connection details, auth token issues
- Keep all existing user-facing error messages unchanged
- Estimated time: 20-30 minutes

**Complex Approach (Not Recommended):**
- Full structured logging integration like the server
- Log every operation and API call
- Complex configuration system for CLI logging
- Estimated time: 2-3 hours

**Recommendation:** Use the simple approach. The CLI is a user-facing tool that should be quiet by default. The existing error handling is already excellent. We only need optional debug logging for troubleshooting HTTP/auth issues.

**Why Simple is Better:**
- CLI users want clean output, not log noise
- Existing error messages are already informative
- Debug flag provides troubleshooting when needed
- Maintains the current excellent UX
- Follows Unix philosophy: quiet by default

## Prerequisites

**Existing Code:**
- `logger/logger.go` - Logger factory (already implemented)
- `logger/config.go` - Logger configuration types (already implemented)
- `cmd/cli/main.go` - CLI entry point
- `cmd/cli/cli.go` - CLI command loop
- `cmd/cli/client/client.go` - HTTP client with error types

**Dependencies:**
- `log/slog` package (standard library, already used in server)
- `github.com/spf13/pflag` (for command-line flags, may need to add)

**Knowledge Required:**
- Understanding of Go's slog package
- Command-line flag parsing in Go
- When to log vs when to show user messages

## Step-by-Step Instructions

### Step 1: Add debug flag and logger initialization to main.go

**File**: `cmd/cli/main.go`

**What to do:**
Add a `--debug` command-line flag and create a logger instance that writes to stderr when debug mode is enabled.

**What to implement:**
- Import the `pflag` package for flag parsing
- Add a `--debug` boolean flag before loading configuration
- Create a logger configuration with:
  - Level: "error" by default, "debug" when flag is set
  - Format: "text" (human-readable for CLI)
  - Output: "stderr" (keep stdout clean for user messages)
  - AddSource: true (helpful for debugging)
  - ServiceName: "task-manager-cli"
  - Environment: "development"
- Create logger using `logger.NewLogger()` with the configuration
- Pass the logger to the CLI struct (will need to update CLI struct)
- Log configuration loading errors using the logger instead of `log.Fatalf()`
- Log successful startup at DEBUG level with server URL

**Why:**
The debug flag gives users control over verbosity. By default, the CLI is quiet (ERROR level), but with `--debug` they can see detailed HTTP communication for troubleshooting. Using stderr keeps stdout clean for user-facing output.

**Expected result:**
- Running `./cli` works as before (quiet, clean output)
- Running `./cli --debug` shows detailed debug logs on stderr
- Config loading errors are logged with structured format

---

### Step 2: Update CLI struct to accept and use logger

**File**: `cmd/cli/cli.go`

**What to do:**
Add a logger field to the CLI struct and update the constructor to accept it.

**What to implement:**
- Add `logger *slog.Logger` field to the `CLI` struct
- Update `NewCLI()` function signature to accept logger parameter
- Store the logger in the CLI struct
- Update the call to `NewCLI()` in main.go to pass the logger

**Why:**
The CLI needs access to the logger to log internal errors and debug information. Dependency injection makes the code testable and follows the existing pattern used in the server.

**Expected result:**
- CLI struct has logger field
- NewCLI accepts logger parameter
- Code compiles without errors

---

### Step 3: Add debug logging for HTTP client operations

**File**: `cmd/cli/client/client.go`

**What to do:**
Add optional logger to HTTPClient and log HTTP request/response details at DEBUG level.

**What to implement:**
- Add `logger *slog.Logger` field to `HTTPClient` struct
- Update `NewHTTPClient()` to accept optional logger parameter (can be nil)
- In `doRequest()` method, add debug logging before making HTTP request:
  - Log method, URL, and whether auth token is present
  - Only log if logger is not nil
- In `doRequest()` method, add debug logging after receiving response:
  - Log status code and response time
  - Only log if logger is not nil
- In `handleErrorResponse()`, add error-level logging for 5xx errors:
  - Log status code and error message
  - Only log if logger is not nil
- Update the call to `NewHTTPClient()` in main.go to pass the logger

**Why:**
When users encounter connection issues or API errors, debug logs showing the actual HTTP requests help diagnose problems. Logging only when logger is provided keeps the client usable in tests without logging noise.

**Expected result:**
- With `--debug` flag, users see HTTP request/response details
- Without flag, HTTP operations are silent (existing behavior)
- Error responses are logged at ERROR level for troubleshooting

---

### Step 4: Add logging for authentication operations

**File**: `cmd/cli/auth/auth.go`

**What to do:**
Add optional logger to auth manager and log authentication events.

**What to implement:**
- Add `logger *slog.Logger` field to `FileAuthManager` struct
- Update `NewFileAuthManager()` to accept optional logger parameter (can be nil)
- Log token loading failures at DEBUG level (file not found is normal)
- Log token validation failures at WARN level
- Log successful authentication at DEBUG level
- Log token save failures at ERROR level
- Only log if logger is not nil

**Why:**
Authentication issues are common troubleshooting scenarios. Logging token operations helps users understand why authentication is failing without exposing sensitive token data.

**Expected result:**
- With `--debug`, users see authentication flow details
- Token validation failures are logged for troubleshooting
- Sensitive data (actual tokens, passwords) is never logged

---

### Step 5: Add selective logging in CLI command handlers

**File**: `cmd/cli/cli.go`

**What to do:**
Add debug logging for command execution flow, but keep user-facing messages unchanged.

**What to implement:**
- In `RunLoop()`, log each command execution at DEBUG level before processing
- In error handlers, log internal errors at ERROR level before showing user message
- Log re-authentication attempts at DEBUG level
- Do NOT change any existing user-facing messages (fmt.Fprintf to output)
- Only add logger calls that provide additional debug context
- All logging should check if logger is not nil before logging

**Why:**
Debug logs help trace command execution flow when troubleshooting. Keeping user messages unchanged preserves the excellent UX. The separation between logs (stderr) and messages (stdout) is maintained.

**Expected result:**
- User-facing output is unchanged
- With `--debug`, command flow is visible in logs
- Logs and user messages are clearly separated (stderr vs stdout)

---

### Step 6: Update CLI initialization in main.go

**File**: `cmd/cli/main.go`

**What to do:**
Wire up all the logger dependencies when creating CLI components.

**What to implement:**
- Pass logger to `NewHTTPClient()` when creating HTTP client
- Pass logger to `NewFileAuthManager()` when creating auth manager
- Pass logger to `NewCLI()` when creating CLI instance
- Log any initialization errors at ERROR level
- Log successful initialization at DEBUG level with configuration details

**Why:**
This completes the dependency injection chain, ensuring all components have access to the logger when debug mode is enabled.

**Expected result:**
- All components receive logger instance
- Debug mode shows initialization flow
- Application compiles and runs correctly

## Verification

### Compile Check
```bash
cd cmd/cli
go build
```
**Expected**: No compilation errors

### Test Normal Operation (Quiet Mode)
```bash
./cli
```
**Expected**: 
- Clean output with no log messages
- User-facing messages appear normally
- Commands work as before

### Test Debug Mode
```bash
./cli --debug
```
**Expected**:
- Debug logs appear on stderr (can redirect: `2>debug.log`)
- User messages still appear on stdout
- Logs show: config loading, HTTP requests, auth operations

### Test Debug with Command
```bash
./cli --debug
# Then type: list
```
**Expected**:
- Debug log shows HTTP GET request to /tasks
- Debug log shows response status code
- User sees task list on stdout (clean format)

### Test Error Scenarios
```bash
# Test with server down
./cli --debug
# Try any command
```
**Expected**:
- ERROR level log shows connection failure details
- User sees friendly error message
- Both stderr (logs) and stdout (messages) provide useful info

### Test Without Debug Flag
```bash
./cli
# Try commands
```
**Expected**:
- No debug logs appear
- Only user-facing messages visible
- Clean, quiet operation (current behavior preserved)

## Common Pitfalls

### Pitfall 1: Logging to stdout instead of stderr
**Symptom**: Log messages mixed with user-facing output, messy display
**Fix**: Always configure logger with `Output: "stderr"` in CLI. User messages use `cli.output` (stdout), logs use stderr.

### Pitfall 2: Logging sensitive data (tokens, passwords)
**Symptom**: Security risk, tokens visible in debug logs
**Fix**: Never log actual token values or passwords. Log only "token present: true/false" or use `logger.MaskToken()` if you must log partial data.

### Pitfall 3: Changing existing user-facing messages
**Symptom**: User experience degraded, messages less friendly
**Fix**: Keep all existing `fmt.Fprintf(cli.output, ...)` calls unchanged. Only ADD logger calls for debug/error scenarios. Logs are supplementary, not replacements.

### Pitfall 4: Nil pointer dereference when logger is nil
**Symptom**: Panic when running without debug flag
**Fix**: Always check `if logger != nil` before calling logger methods, OR ensure logger is always created (even if at ERROR level).

### Pitfall 5: Over-logging in CLI
**Symptom**: Debug mode too noisy, hard to find useful information
**Fix**: Be selective. Log only: HTTP requests/responses, auth operations, config loading, and actual errors. Don't log every variable or function call.

## Learning Resources

### Essential Reading
- [Go slog Package](https://pkg.go.dev/log/slog) - Official documentation for structured logging
- [pflag Package](https://pkg.go.dev/github.com/spf13/pflag) - POSIX/GNU-style command-line flags

### Additional Resources (Optional)
- [12 Factor App: Logs](https://12factor.net/logs) - Best practices for application logging
- [Unix Philosophy: Rule of Silence](http://www.catb.org/~esr/writings/taoup/html/ch01s06.html#id2878450) - Why CLIs should be quiet by default

## Implementation Notes

### Design Decisions

**Why stderr for logs?**
- Unix convention: stdout for output, stderr for diagnostics
- Allows users to redirect separately: `./cli > tasks.txt 2> debug.log`
- Keeps user-facing output clean and pipeable

**Why ERROR level by default?**
- CLI should be quiet during normal operation
- Only show logs when something is actually wrong
- Debug flag explicitly enables verbose mode

**Why optional logger in components?**
- Makes components testable without logging noise
- Allows gradual adoption
- Doesn't break existing code

**Why not full structured logging like server?**
- CLI is user-facing, not a long-running service
- No need for log aggregation or analysis
- Simplicity and UX are more important than comprehensive logging

### Testing Strategy

When implementing, test these scenarios:
1. **Default mode**: No logs, clean output
2. **Debug mode**: Logs visible, still usable
3. **Error scenarios**: Helpful logs + user messages
4. **Redirection**: `./cli > out.txt 2> err.txt` works correctly
5. **No server**: Connection errors logged and displayed properly

### Future Enhancements (Not in Scope)

These are explicitly NOT part of this task:
- Log file output (CLI is short-lived, stderr is sufficient)
- Log rotation (not needed for CLI)
- JSON format (text is more readable for CLI debugging)
- Log levels per component (too complex for CLI)
- Integration with log aggregation tools (CLI is not a service)

## Summary

This task adds **minimal, optional** debug logging to the CLI without changing its excellent user experience. The `--debug` flag provides troubleshooting capability when needed, while default operation remains clean and quiet. This follows the Unix philosophy of being quiet by default while providing verbose modes for debugging.

The implementation is intentionally simple because:
1. The CLI already has great error handling
2. Users want clean output, not log noise
3. Debug logs are for troubleshooting, not normal operation
4. The existing UX should be preserved

Total estimated implementation time: 20-30 minutes for an experienced Go developer.
