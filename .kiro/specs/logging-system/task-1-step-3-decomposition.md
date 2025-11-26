# Step 3 Decomposition: Implement Output Destination Handler

## Overview

This step creates a function that determines where log messages should be written. The function takes an output destination string (like "stdout", "stderr", or a file path) and returns an `io.Writer` that the logger will use. This abstraction allows the same logger code to write to different destinations based on configuration.

## Implementation Approach

We're implementing a factory function that returns different `io.Writer` implementations based on the input string. Go's `io.Writer` interface is powerful because it abstracts away the details of where data goes - whether it's the console, a file, a network connection, or anything else.

The function needs to handle three cases:
1. Standard output (stdout) - for normal console output
2. Standard error (stderr) - for error streams, common in containers
3. File paths - for persistent logging to disk

For file paths, we need to handle several edge cases: creating parent directories, opening files in append mode, and setting appropriate permissions.

**Key Concepts:**
- **io.Writer interface**: Go's standard interface for writing bytes
- **os.Stdout/os.Stderr**: Pre-defined writers for console output
- **File permissions**: Unix-style permissions (0644 = rw-r--r--)
- **Append mode**: Opening files without truncating existing content

## Prerequisites

**Existing Code:**
- `logger/config.go` - Config struct with Output field

**Dependencies:**
- `io` package (standard library)
- `os` package (standard library)
- `path/filepath` package (standard library)

**Knowledge Required:**
- Understanding of Go's io.Writer interface
- Basic file operations in Go
- Unix file permissions (octal notation)

## Detailed Implementation Steps

### Sub-step 3.1: Create the function signature

**File**: `logger/logger.go` (create new file)

**What to do:**
Create the basic function structure with proper signature and imports.

**What to implement:**
- Add package declaration: `package logger`
- Import required packages: `io`, `os`, `path/filepath`, `fmt`
- Create function named `getWriter` that:
  - Accepts one parameter: `output string`
  - Returns two values: `io.Writer` and `error`
- Add a function comment explaining what it does

**Why:**
The function signature defines the contract. Returning an error allows callers to handle failures gracefully. The io.Writer return type provides flexibility - any type that implements Write([]byte) (int, error) can be returned.

**Expected result:**
You have a `logger/logger.go` file with a function skeleton that compiles (even if it just returns nil, nil for now).

---

### Sub-step 3.2: Handle stdout case

**File**: `logger/logger.go` (modify existing)

**What to do:**
Add logic to detect and handle the "stdout" output destination.

**What to implement:**
- Check if the `output` parameter equals "stdout" (case-sensitive)
- If it matches, return `os.Stdout` and `nil` error
- `os.Stdout` is a pre-defined `*os.File` that implements `io.Writer`
- This should be a simple if statement at the start of the function

**Why:**
stdout is the standard output stream. In Unix/Linux, this is file descriptor 1. It's the default destination for program output and is commonly used in development and when piping output to other programs.

**Expected result:**
When you call `getWriter("stdout")`, it returns os.Stdout without error.

---

### Sub-step 3.3: Handle stderr case

**File**: `logger/logger.go` (modify existing)

**What to do:**
Add logic to detect and handle the "stderr" output destination.

**What to implement:**
- Check if the `output` parameter equals "stderr" (case-sensitive)
- If it matches, return `os.Stderr` and `nil` error
- `os.Stderr` is a pre-defined `*os.File` that implements `io.Writer`
- This should be another if statement after the stdout check

**Why:**
stderr is the standard error stream (file descriptor 2). It's commonly used for error messages and logging because it's separate from stdout. In containerized environments (Docker, Kubernetes), stderr is often captured separately for log aggregation.

**Expected result:**
When you call `getWriter("stderr")`, it returns os.Stderr without error.

---

### Sub-step 3.4: Extract directory path from file path

**File**: `logger/logger.go` (modify existing)

**What to do:**
For file paths, extract the directory portion to check if it exists.

**What to implement:**
- If output is not "stdout" or "stderr", treat it as a file path
- Use `filepath.Dir(output)` to extract the directory portion
- Store this in a variable (e.g., `dir`)
- For example: "/var/log/app/app.log" → "/var/log/app"
- Handle the special case where output is just a filename (e.g., "app.log") - Dir returns "."

**Why:**
Before creating a file, we need to ensure its parent directory exists. filepath.Dir is the standard way to extract the directory portion of a path in a cross-platform way.

**Expected result:**
You can extract the directory from any file path string.

---

### Sub-step 3.5: Create parent directories if needed

**File**: `logger/logger.go` (modify existing)

**What to do:**
Ensure the directory exists before trying to create the log file.

**What to implement:**
- Use `os.MkdirAll(dir, 0755)` to create the directory and any missing parents
- The permission `0755` means: owner can read/write/execute, others can read/execute
- Check the error returned by MkdirAll
- If the error is not nil, wrap it with context and return it
- Use `fmt.Errorf("failed to create log directory %s: %w", dir, err)` for wrapping
- Note: MkdirAll succeeds if the directory already exists (idempotent)

**Why:**
If you try to create "/var/log/app/app.log" but "/var/log/app" doesn't exist, the file creation will fail. MkdirAll creates all necessary parent directories in one call. The 0755 permission is standard for directories - it allows the owner full access and others to read and traverse.

**Expected result:**
Directories are created as needed. If directory creation fails (e.g., permission denied), an error is returned with context.

---

### Sub-step 3.6: Open the log file

**File**: `logger/logger.go` (modify existing)

**What to do:**
Open the log file for writing, creating it if it doesn't exist.

**What to implement:**
- Use `os.OpenFile(output, flags, 0644)` to open the file
- Set flags to: `os.O_CREATE | os.O_WRONLY | os.O_APPEND`
  - `O_CREATE`: Create file if it doesn't exist
  - `O_WRONLY`: Open for writing only
  - `O_APPEND`: Append to end of file (don't truncate)
- The permission `0644` means: owner can read/write, others can read
- Store the returned `*os.File` in a variable
- Check for errors and return them with context if they occur

**Why:**
The flags control how the file is opened. O_APPEND is crucial - without it, each logger creation would erase previous logs. O_CREATE means we don't need to check if the file exists first. The 0644 permission is standard for log files - readable by all, writable only by owner.

**Expected result:**
The file is opened (or created) in append mode. The returned *os.File implements io.Writer.

---

### Sub-step 3.7: Return the file writer

**File**: `logger/logger.go` (modify existing)

**What to do:**
Return the opened file as an io.Writer.

**What to implement:**
- Return the `*os.File` from OpenFile and `nil` error
- `*os.File` implements `io.Writer`, so it can be returned directly
- This should be the final return statement in the function

**Why:**
The *os.File type implements io.Writer through its Write method. By returning it as io.Writer, we maintain the abstraction - callers don't need to know they're writing to a file specifically.

**Expected result:**
The function returns a valid io.Writer that writes to the specified file.

---

### Sub-step 3.8: Add error handling for edge cases

**File**: `logger/logger.go` (modify existing)

**What to do:**
Handle potential edge cases and invalid inputs.

**What to implement:**
- Check if `output` is an empty string at the start of the function
- If empty, return an error: `fmt.Errorf("output destination cannot be empty")`
- Consider checking for obviously invalid paths (though OS will catch most)
- Ensure all error returns include context about what failed
- Make sure no code path returns (nil, nil) - that would cause a panic later

**Why:**
Empty output strings would cause confusing errors later. Catching them early with a clear message helps debugging. All errors should have context so users know what went wrong and how to fix it.

**Expected result:**
Invalid inputs are rejected with clear error messages. All error paths return non-nil errors.

---

## Complete Function Flow

Here's how the function should flow (without showing actual code):

1. Check if output is empty → return error
2. Check if output == "stdout" → return os.Stdout, nil
3. Check if output == "stderr" → return os.Stderr, nil
4. Otherwise, treat as file path:
   - Extract directory path
   - Create directories if needed → return error if fails
   - Open file with create/append flags → return error if fails
   - Return file writer, nil

## Verification

### Test with stdout
```bash
# Create a test file
cat > test_writer.go << 'EOF'
package main

import (
    "yourproject/logger"
    "fmt"
)

func main() {
    w, err := logger.GetWriter("stdout")
    if err != nil {
        panic(err)
    }
    fmt.Fprintf(w, "Testing stdout\n")
}
EOF

go run test_writer.go
```
**Expected**: "Testing stdout" appears in console

### Test with stderr
```bash
# Modify test to use stderr
go run test_writer.go 2>&1 | cat
```
**Expected**: Output appears (stderr is redirected)

### Test with file path
```bash
# Test file creation
cat > test_file.go << 'EOF'
package main

import (
    "yourproject/logger"
    "fmt"
)

func main() {
    w, err := logger.GetWriter("/tmp/test.log")
    if err != nil {
        panic(err)
    }
    fmt.Fprintf(w, "Testing file output\n")
}
EOF

go run test_file.go
cat /tmp/test.log
```
**Expected**: File created at /tmp/test.log with content

### Test with nested directory
```bash
# Test directory creation
w, err := logger.GetWriter("/tmp/logs/app/test.log")
# Should create /tmp/logs/app/ directories
ls -la /tmp/logs/app/
```
**Expected**: Directories created, file exists

### Test error cases
```bash
# Test empty output
w, err := logger.GetWriter("")
# Should return error

# Test invalid path (if on Unix)
w, err := logger.GetWriter("/root/test.log")
# Should return permission error (unless running as root)
```
**Expected**: Appropriate errors returned

## Common Pitfalls

### Pitfall 1: Forgetting O_APPEND flag
**Symptom**: Each time you create a logger, previous log content is erased
**Fix**: Include `os.O_APPEND` in the OpenFile flags: `os.O_CREATE | os.O_WRONLY | os.O_APPEND`

### Pitfall 2: Wrong file permissions
**Symptom**: File created but other users/processes can't read it, or file is executable
**Fix**: Use `0644` for files (rw-r--r--) and `0755` for directories (rwxr-xr-x)

### Pitfall 3: Not creating parent directories
**Symptom**: Error "no such file or directory" when path has non-existent parents
**Fix**: Use `os.MkdirAll(dir, 0755)` before opening the file

### Pitfall 4: Not handling empty output string
**Symptom**: Confusing error messages or unexpected behavior
**Fix**: Check for empty string at the start: `if output == "" { return nil, fmt.Errorf("...") }`

### Pitfall 5: Returning (nil, nil) on error paths
**Symptom**: Panic: "invalid memory address or nil pointer dereference" when using the writer
**Fix**: Always return a non-nil error when returning nil writer: `return nil, fmt.Errorf("...")`

### Pitfall 6: Not closing files (future consideration)
**Symptom**: File handle leaks over time
**Note**: For this step, we're not implementing Close() yet. The caller will be responsible for closing. In a production system, you might want to track opened files and provide a cleanup function.

## Learning Resources

### Essential Reading
- [Go io.Writer Interface](https://pkg.go.dev/io#Writer) - Understanding the Writer interface
- [Go os Package](https://pkg.go.dev/os) - File operations and standard streams
- [Go filepath Package](https://pkg.go.dev/path/filepath) - Cross-platform path manipulation

### Additional Resources
- [Unix File Permissions](https://en.wikipedia.org/wiki/File-system_permissions#Numeric_notation) - Understanding octal permissions
- [File Descriptors](https://en.wikipedia.org/wiki/File_descriptor) - Understanding stdin/stdout/stderr

## Testing Your Implementation

After implementing, test each case:

```bash
# Compile check
go build ./logger

# Quick manual test
go run -c '
package main
import "yourproject/logger"
func main() {
    w, _ := logger.GetWriter("stdout")
    w.Write([]byte("test\n"))
}
'
```

## Next Steps

After completing Step 3, you'll have a working output destination handler. Step 4 will use this function to create the appropriate slog handler (JSON or text) that writes to your chosen destination.

The combination of Steps 3 and 4 gives you the core of the logger factory: determining where to write (Step 3) and how to format (Step 4).
