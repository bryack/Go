# Task Decomposition: Implement Log File Rotation Support

## Overview

Add log file rotation using the lumberjack library to prevent log files from growing indefinitely. The configuration infrastructure is already complete - this task only requires integrating lumberjack into the existing `getWriter()` function when file output and rotation are enabled.

## Complexity Check & Simplification

**Requirements Analysis:**
- Requirement 8.5: "THE Application SHALL support log rotation configuration (max size, max age, max backups)"
- Configuration fields already exist: `EnableRotation`, `MaxSize`, `MaxAge`, `MaxBackups`
- Validation already implemented in `logger/config.go`

**Current State:**
- ✅ Config struct has all rotation fields
- ✅ Validation checks rotation settings
- ✅ `getWriter()` creates files with append mode
- ❌ No actual rotation - files grow indefinitely

**Simple Approach (Recommended - 10-15 minutes):**
- Add lumberjack dependency: `go get gopkg.in/natefinch/lumberjack.v2`
- Modify `getWriter()` to wrap file writer with lumberjack when rotation enabled
- ~15 lines of code total

**Complex Approach (Not Recommended):**
- Implement custom rotation logic with goroutines, file monitoring, compression
- 2-3 hours of work, error-prone, reinventing the wheel

**Recommendation:** Use lumberjack. It's the standard Go solution, battle-tested, and requires minimal code.

## Implementation Approach

We're modifying the existing `getWriter()` function in `logger/logger.go` to use lumberjack when:
1. Output is a file path (not stdout/stderr)
2. `EnableRotation` is true in the config

Lumberjack automatically handles:
- Rotating files when they reach MaxSize
- Deleting old files based on MaxAge
- Keeping only MaxBackups old files
- Compressing rotated files (optional)

**Key Concepts:**
- **Lumberjack**: Drop-in replacement for `*os.File` that implements `io.Writer` with automatic rotation
- **Rotation Triggers**: Files rotate when size limit is reached
- **Cleanup**: Old files are automatically deleted based on age and backup count
- **Zero Downtime**: Rotation happens transparently without blocking writes

## Prerequisites

**Existing Code:**
- `logger/logger.go` - Has `getWriter()` function that creates file writers
- `logger/config.go` - Has `Config` struct with rotation fields and validation ✅
- `cmd/server/config/config.go` - Config system with rotation defaults ✅

**Dependencies:**
- `gopkg.in/natefinch/lumberjack.v2` - Need to add this

**Knowledge Required:**
- Understanding of `io.Writer` interface
- Basic understanding of how lumberjack works (it's simple)

## Step-by-Step Instructions

### Step 1: Add lumberjack dependency

**What to do:**
Add the lumberjack library to your project.

**What to implement:**
- Run: `go get gopkg.in/natefinch/lumberjack.v2`
- This will update your `go.mod` file

**Why:**
Lumberjack is the standard Go library for log rotation. It's maintained, well-tested, and used by thousands of projects.

**Expected result:**
Dependency is added to `go.mod`. You can import it in your code.

---

### Step 2: Import lumberjack in logger.go

**File**: `logger/logger.go`

**What to do:**
Add the lumberjack import to the existing imports.

**What to implement:**
- Add to imports: `"gopkg.in/natefinch/lumberjack.v2"`

**Why:**
You need to import the package to use it.

**Expected result:**
File compiles with the new import.

---

### Step 3: Modify getWriter to accept Config parameter

**File**: `logger/logger.go`

**What to do:**
Change `getWriter()` signature to accept the full Config so it can check rotation settings.

**What to implement:**
- Change function signature from `getWriter(output string)` to `getWriter(cfg *Config)`
- Update the function body to use `cfg.Output` instead of `output` parameter
- Update the call to `getWriter()` in `NewLogger()` to pass `cfg` instead of `cfg.Output`

**Why:**
The function needs access to rotation settings (`EnableRotation`, `MaxSize`, etc.) which are in the Config struct.

**Expected result:**
Function signature changed. Code still compiles. Behavior unchanged (for now).

---

### Step 4: Add lumberjack writer for file rotation

**File**: `logger/logger.go` (in the `getWriter` function)

**What to do:**
After the file path handling (after directory creation), add logic to use lumberjack when rotation is enabled.

**What to implement:**
- After the `os.MkdirAll` call (which creates directories)
- Before the `os.OpenFile` call
- Add a check: if `cfg.EnableRotation` is true, create and return a lumberjack writer instead
- Create lumberjack.Logger with:
  - `Filename`: `cfg.Output` (the file path)
  - `MaxSize`: `cfg.MaxSize` (in megabytes)
  - `MaxAge`: `cfg.MaxAge` (in days)
  - `MaxBackups`: `cfg.MaxBackups` (number of old files to keep)
  - `Compress`: `false` (optional - set to true if you want gzip compression)
- Return the lumberjack.Logger (it implements io.Writer)
- Keep the existing `os.OpenFile` code for when rotation is disabled

**Why:**
Lumberjack handles all rotation logic automatically. When the file reaches MaxSize, it renames it with a timestamp and creates a new file. Old files are cleaned up based on MaxAge and MaxBackups.

**Expected result:**
When rotation is enabled, lumberjack is used. When disabled, regular file writing is used. Both implement io.Writer so the rest of the code doesn't change.

---

### Step 5: Test the implementation

**What to do:**
Verify that rotation works correctly.

**What to implement:**
Test scenarios:
1. **Rotation disabled**: Logs go to regular file, no rotation
2. **Rotation enabled**: Logs rotate when size limit reached
3. **Old file cleanup**: Old files are deleted based on MaxAge/MaxBackups

**Why:**
Ensure the implementation works correctly and doesn't break existing functionality.

**Expected result:**
Rotation works as configured. No breaking changes to existing behavior.

---

## Complete Implementation Example (Reference Only)

Here's what the modified `getWriter()` function should look like (don't copy this - implement it yourself):

```go
func getWriter(cfg *Config) (io.Writer, error) {
    if len(cfg.Output) == 0 {
        return nil, fmt.Errorf("output destination cannot be empty")
    }

    outputToLower := strings.ToLower(cfg.Output)

    if outputToLower == "stdout" {
        return os.Stdout, nil
    }

    if outputToLower == "stderr" {
        return os.Stderr, nil
    }

    // File path - create directory
    dir := filepath.Dir(cfg.Output)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return nil, fmt.Errorf("failed to create log directory %s: %w", dir, err)
    }

    // Use lumberjack for rotation if enabled
    if cfg.EnableRotation {
        return &lumberjack.Logger{
            Filename:   cfg.Output,
            MaxSize:    cfg.MaxSize,    // megabytes
            MaxAge:     cfg.MaxAge,     // days
            MaxBackups: cfg.MaxBackups, // number of backups
            Compress:   false,          // don't compress by default
        }, nil
    }

    // Regular file without rotation
    file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return nil, fmt.Errorf("failed to create log file %s: %w", cfg.Output, err)
    }

    return file, nil
}
```

## Verification

### Compile Check
```bash
go build ./...
```
**Expected**: No compilation errors

### Test Rotation Disabled (Default)
```bash
# Run with file output but rotation disabled
./server --log-output=/tmp/app.log

# Make some requests to generate logs
curl http://localhost:8080/health

# Check the file
ls -lh /tmp/app.log
```
**Expected**: Single log file, no rotation files

### Test Rotation Enabled
```bash
# Create config with rotation enabled
cat > test-rotation.yaml << 'EOF'
logging:
  level: info
  format: json
  output: /tmp/rotated.log
  enable_rotation: true
  max_size: 1        # 1 MB for testing
  max_age: 7
  max_backups: 3
EOF

# Run server
./server --config test-rotation.yaml

# Generate logs until rotation happens (write >1MB of logs)
for i in {1..10000}; do
  curl http://localhost:8080/health
done

# Check for rotated files
ls -lh /tmp/rotated*.log
```
**Expected**: Multiple files: `rotated.log`, `rotated-2024-11-12T10-30-00.000.log`, etc.

### Test Old File Cleanup
```bash
# With max_backups: 3, only 3 old files should be kept
# After generating many rotations, verify only 3 backups exist
ls -1 /tmp/rotated*.log | wc -l
```
**Expected**: 4 files total (1 current + 3 backups)

## Common Pitfalls

### Pitfall 1: Forgetting to pass Config to getWriter
**Symptom**: Compilation error "not enough arguments"
**Fix**: Update the call in `NewLogger()` to pass `cfg` instead of `cfg.Output`

### Pitfall 2: Not checking EnableRotation flag
**Symptom**: Rotation happens even when disabled
**Fix**: Only create lumberjack.Logger when `cfg.EnableRotation` is true

### Pitfall 3: Wrong MaxSize units
**Symptom**: Files rotate too early or too late
**Fix**: MaxSize is in megabytes. Config validation already ensures it's positive.

### Pitfall 4: File handle leaks
**Symptom**: Too many open files error
**Fix**: Lumberjack handles this automatically. No need to close the writer.

### Pitfall 5: Rotation not working in tests
**Symptom**: Tests don't see rotation
**Fix**: Use small MaxSize (1 MB) for testing. Generate enough logs to trigger rotation.

## Learning Resources

### Essential Reading
- [Lumberjack Documentation](https://pkg.go.dev/gopkg.in/natefinch/lumberjack.v2) - Official docs with examples
- [Go io.Writer Interface](https://pkg.go.dev/io#Writer) - Understanding the interface lumberjack implements

### Additional Resources (Optional)
- [Log Rotation Best Practices](https://www.loggly.com/ultimate-guide/managing-log-files/) - When and why to rotate logs
- [Lumberjack GitHub](https://github.com/natefinch/lumberjack) - Source code and issues

## Real-World Context

### Why Log Rotation Matters

Without rotation:
- Log files grow indefinitely
- Disk space fills up
- Application crashes when disk is full
- Old logs are never cleaned up

With rotation:
- Files stay manageable size
- Old logs are automatically cleaned up
- Disk space is controlled
- No manual intervention needed

### Production Settings

Typical production settings:
- **MaxSize**: 100 MB (balance between rotation frequency and file size)
- **MaxAge**: 30 days (compliance/retention requirements)
- **MaxBackups**: 5-10 (enough history without wasting space)
- **Compress**: true (save disk space for old logs)

### Container Environments

Note: In containerized environments (Docker, Kubernetes), you typically:
- Log to stdout/stderr (not files)
- Let the container runtime handle log collection
- Use external log aggregation (ELK, Loki, CloudWatch)

File rotation is mainly for:
- Traditional server deployments
- Development environments
- Edge cases where file logging is required

## Summary

This is a simple task that adds important production functionality:

**What you're doing:**
- Adding lumberjack dependency (1 command)
- Modifying `getWriter()` to use lumberjack when rotation enabled (~15 lines)
- Testing that it works

**What you're NOT doing:**
- Implementing custom rotation logic
- Managing file handles manually
- Writing complex goroutines
- Dealing with edge cases (lumberjack handles them)

**Time estimate:** 10-15 minutes for an experienced Go developer

**Value:** Prevents disk space issues in production, meets Requirement 8.5
