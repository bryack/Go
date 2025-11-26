# Task Decomposition: Integrate Logger into Storage Layer

## Overview

Add structured logging to the storage layer to provide visibility into database operations, errors, and performance. This enables debugging of data issues, monitoring of query performance, and tracking of user-specific database operations for security and compliance.

## Implementation Approach

We're adding a logger field to the `DatabaseStorage` struct and logging at appropriate points in the database lifecycle. The approach uses dependency injection to pass the logger instance during storage initialization, following the existing pattern used for database connections.

**Complexity Check:**
- **Requirements need**: Log database operations at DEBUG level, errors at ERROR level, migrations at INFO level, include user ID in logs
- **Simple approach**: Add logger field, pass it through constructor, add log statements at key points (5-10 log calls total)
- **Complex approach**: Create database operation wrappers, query interceptors, automatic parameter logging
- **Recommendation**: Use simple approach. Requirements only need visibility into operations and errors, not query analysis or performance profiling. Add straightforward log calls at operation boundaries.

**Key Concepts:**
- **Dependency Injection**: Logger is passed to `NewDatabaseStorage` and stored as a field, making it available to all methods
- **Log Levels**: DEBUG for normal operations (queries), INFO for lifecycle events (migrations), ERROR for failures
- **Contextual Logging**: Include operation type, user ID, and relevant IDs (task ID) in all log entries
- **User ID Tracking**: All CRUD operations already accept userID parameter, making it easy to include in logs

## Prerequisites

**Existing Code:**
- `storage/database.go` - DatabaseStorage struct and all CRUD methods
- `storage/connection.go` - Database connection management
- `storage/migrations.go` - Database migration system
- `logger/logger.go` - Logger factory and configuration
- `logger/fields.go` - Standard field name constants
- `cmd/server/main.go` - Server initialization where storage is created

**Dependencies:**
- `log/slog` package (already imported in logger package)
- Logger instance created in main.go (already implemented)

**Knowledge Required:**
- Understanding of Go struct fields and methods
- Familiarity with slog's structured logging API (`slog.String()`, `slog.Int()`, etc.)
- Understanding of when to use different log levels (DEBUG, INFO, ERROR)

## Step-by-Step Instructions

### Step 1: Add logger field to DatabaseStorage struct

**File**: `storage/database.go`

**What to do:**
Add a logger field to the `DatabaseStorage` struct so all methods can access it.

**What to implement:**
- Add a `logger *slog.Logger` field to the `DatabaseStorage` struct
- Place it after the existing `db` and `migrator` fields
- Required import: `log/slog` (add to imports at top of file)

**Why:**
The logger needs to be accessible to all DatabaseStorage methods. Storing it as a field follows the dependency injection pattern already used for the database connection.

**Expected result:**
Code compiles. DatabaseStorage struct now has a logger field.

---

### Step 2: Update NewDatabaseStorage to accept and store logger

**File**: `storage/database.go`

**What to do:**
Modify the `NewDatabaseStorage` function signature to accept a logger parameter and store it in the struct.

**What to implement:**
- Add `logger *slog.Logger` as the second parameter to `NewDatabaseStorage` (after `dbPath string`)
- Store the logger in the DatabaseStorage struct when creating the instance
- The function signature should become: `func NewDatabaseStorage(dbPath string, logger *slog.Logger) (*DatabaseStorage, error)`

**Why:**
This follows the dependency injection pattern. The logger is created in main.go and passed down to components that need it, making the code testable and flexible.

**Expected result:**
Code compiles. NewDatabaseStorage now accepts a logger parameter.

---

### Step 3: Log database connection event

**File**: `storage/database.go`

**What to do:**
Add a log entry when the database connection is successfully established.

**What to implement:**
- After the `CreateConnection` call succeeds (before checking for errors), add an INFO level log
- Log message: "Database connection established"
- Include fields: database path using `slog.String("db_path", dbPath)`
- Use the logger instance: `logger.Info(...)`

**Why:**
Connection events are important for troubleshooting startup issues and understanding when the database becomes available. INFO level is appropriate because this is a significant lifecycle event.

**Expected result:**
When the server starts, you'll see a log entry showing the database connection was established with the file path.

---

### Step 4: Log migration events

**File**: `storage/database.go`

**What to do:**
Add logging before and after migration execution to track schema changes.

**What to implement:**
- Before calling `migrator.ApplyMigrations()`, log at INFO level: "Applying database migrations"
- After successful migration (before checking error), log at INFO level: "Database migrations completed"
- If migration fails, the error will be returned and logged by the caller (no additional logging needed here)

**Why:**
Migrations are critical operations that change the database schema. Logging them helps track when schema changes occur and aids in troubleshooting migration failures.

**Expected result:**
During server startup, you'll see log entries showing migration start and completion.

---

### Step 5: Log CRUD operations at DEBUG level

**File**: `storage/database.go`

**What to do:**
Add DEBUG level logging to all CRUD methods to track database operations during development and troubleshooting.

**What to implement:**
For each CRUD method (`CreateTask`, `UpdateTask`, `DeleteTask`, `GetTaskByID`, `LoadTasks`), add a DEBUG log at the start of the method:

- **CreateTask**: Log "Creating task" with fields: user_id, description (first 50 chars)
- **UpdateTask**: Log "Updating task" with fields: task_id, user_id, done status
- **DeleteTask**: Log "Deleting task" with fields: task_id, user_id
- **GetTaskByID**: Log "Fetching task" with fields: task_id, user_id
- **LoadTasks**: Log "Loading tasks" with field: user_id

Use the standard field names from `logger/fields.go`:
- `logger.FieldUserID` for user_id
- `logger.FieldTaskID` for task_id
- `logger.FieldOperation` for operation type

**Why:**
DEBUG level logs provide detailed operation traces useful during development and troubleshooting. They're disabled in production by default (INFO level) but can be enabled when investigating issues. Including user_id in every log enables security auditing and user-specific debugging.

**Expected result:**
When log level is set to DEBUG, you'll see detailed logs for every database operation with relevant context.

---

### Step 6: Log database errors at ERROR level

**File**: `storage/database.go`

**What to do:**
Add ERROR level logging when database operations fail with unexpected errors.

**What to implement:**
For each CRUD method, add ERROR logging when database operations fail (but NOT for expected errors like ErrTaskNotFound):

- After database operations that return errors (Exec, Query, QueryRow, Scan)
- Before returning the error to the caller
- Log message should describe the operation that failed
- Include fields: operation type, user_id, task_id (if applicable), error message
- Use `slog.String("error", err.Error())` for the error
- Do NOT log when returning `ErrTaskNotFound` (this is an expected error, not a system error)

Example locations:
- **CreateTask**: After `db.Exec` fails or `LastInsertId` fails
- **UpdateTask**: After `db.Exec` fails (but not when rowsAffected is 0)
- **DeleteTask**: After `db.Exec` fails (but not when rowsAffected is 0)
- **GetTaskByID**: After `QueryRow.Scan` fails (but not for `sql.ErrNoRows`)
- **LoadTasks**: After `db.Query` fails, `rows.Scan` fails, or `rows.Err` returns error

**Why:**
ERROR level logs indicate unexpected failures that need attention. We log database errors because they might indicate connection issues, disk problems, or SQL errors. We don't log ErrTaskNotFound because it's an expected business logic error (like a 404), not a system error.

**Expected result:**
When database operations fail unexpectedly, you'll see ERROR logs with full context to help diagnose the issue.

---

### Step 7: Log RowsAffected for update and delete operations

**File**: `storage/database.go`

**What to do:**
Add DEBUG level logging to show how many rows were affected by UPDATE and DELETE operations.

**What to implement:**
In `UpdateTask` and `DeleteTask` methods:
- After successfully getting `rowsAffected` from the result
- Before checking if rowsAffected is 0
- Log at DEBUG level with message: "Database operation completed"
- Include fields: operation type, task_id, user_id, rows_affected

Use `slog.Int64("rows_affected", rowsAffected)` for the count.

**Why:**
Knowing how many rows were affected helps diagnose issues where operations succeed but don't modify data as expected. This is especially useful for debugging user ownership issues (where user_id doesn't match).

**Expected result:**
When DEBUG logging is enabled, you'll see how many rows each UPDATE and DELETE operation affected.

---

### Step 8: Update main.go to pass logger to storage

**File**: `cmd/server/main.go`

**What to do:**
Update the storage initialization to pass the logger instance.

**What to implement:**
- Find the line where `storage.NewDatabaseStorage` is called
- Add the logger instance as the second parameter: `storage.NewDatabaseStorage(cfg.DatabaseConfig.Path, l)`
- The logger `l` is already created earlier in main.go

**Why:**
This completes the dependency injection chain. The logger created in main.go is now passed to the storage layer, enabling all the logging we added in previous steps.

**Expected result:**
Server compiles and starts successfully. Storage layer now logs operations.

---

### Step 9: Update user storage methods with logging

**File**: `storage/user.go`

**What to do:**
Add the same logging pattern to user-related storage methods.

**What to implement:**
Apply the same logging approach to user storage methods:
- **CreateUser**: DEBUG log "Creating user" with masked email
- **GetUserByEmail**: DEBUG log "Fetching user by email" with masked email
- **GetUserByID**: DEBUG log "Fetching user by ID" with user_id
- **EmailExists**: DEBUG log "Checking email existence" with masked email
- Add ERROR logs for database failures (not for ErrUserNotFound)

Use `logger.MaskEmail(email)` when logging email addresses for privacy.

**Why:**
User operations need the same visibility as task operations. Masking emails protects user privacy in logs while still providing useful debugging information.

**Expected result:**
User-related database operations are logged with the same detail as task operations.

## Verification

### Compile Check
```bash
go build ./...
```
**Expected**: No compilation errors

### Run Server with DEBUG Logging
```bash
# Update config.yaml to set log level to debug
# Or use environment variable
TASKMANAGER_LOGGING_LEVEL=debug go run cmd/server/main.go
```
**Expected**: 
- See "Database connection established" log on startup
- See "Applying database migrations" and "Database migrations completed" logs
- See DEBUG logs for each database operation when you make API calls

### Test Database Operations
```bash
# Start the server
go run cmd/server/main.go

# In another terminal, register and login
TOKEN=$(curl -s -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')

# Create a task (should see DEBUG logs for CreateTask)
curl -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"description":"Test task"}'

# Get tasks (should see DEBUG logs for LoadTasks)
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer $TOKEN"

# Update task (should see DEBUG logs for UpdateTask with rows_affected)
curl -X PUT http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"done":true}'

# Delete task (should see DEBUG logs for DeleteTask with rows_affected)
curl -X DELETE http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer $TOKEN"
```

**Expected logs** (with DEBUG level enabled):
```
INFO Database connection established db_path=./data/tasks.db
INFO Applying database migrations
INFO Database migrations completed
DEBUG Creating task user_id=1 description="Test task"
DEBUG Database operation completed operation=create task_id=1 user_id=1
DEBUG Loading tasks user_id=1
DEBUG Updating task task_id=1 user_id=1 done=true
DEBUG Database operation completed operation=update task_id=1 user_id=1 rows_affected=1
DEBUG Deleting task task_id=1 user_id=1
DEBUG Database operation completed operation=delete task_id=1 user_id=1 rows_affected=1
```

### Test Error Logging
```bash
# Cause a database error by using an invalid database path
TASKMANAGER_DATABASE_PATH=/invalid/path/tasks.db go run cmd/server/main.go
```
**Expected**: See ERROR log with database connection failure details

### Run Tests
```bash
go test ./storage -v
```
**Expected**: All existing tests pass (they'll need to be updated to pass a logger to NewDatabaseStorage)

## Common Pitfalls

### Pitfall 1: Forgetting to import log/slog
**Symptom**: Compilation error "undefined: slog"
**Fix**: Add `"log/slog"` to the imports at the top of `storage/database.go` and `storage/user.go`

### Pitfall 2: Logging ErrTaskNotFound at ERROR level
**Symptom**: Logs filled with ERROR entries for normal 404 responses
**Fix**: Only log at ERROR level for unexpected database errors. ErrTaskNotFound is expected and should not be logged (or logged at DEBUG/WARN level by the handler)

### Pitfall 3: Not updating all NewDatabaseStorage calls
**Symptom**: Compilation errors in tests or other files that create storage
**Fix**: Search for all calls to `NewDatabaseStorage` and update them to pass a logger. For tests, you can use `slog.Default()` or create a test logger with `logger.NewDefault()`

### Pitfall 4: Logging before checking for nil logger
**Symptom**: Panic when logger is nil
**Fix**: Ensure logger is always passed to NewDatabaseStorage. If you want to make it optional, add nil checks: `if ds.logger != nil { ds.logger.Debug(...) }`

### Pitfall 5: Forgetting to mask sensitive data
**Symptom**: Emails or other PII appear in logs unmasked
**Fix**: Always use `logger.MaskEmail(email)` when logging email addresses. Never log passwords or tokens.

## Learning Resources

### Essential Reading
- [Go slog Package Documentation](https://pkg.go.dev/log/slog) - Official documentation for structured logging
- [Structured Logging Best Practices](https://www.honeycomb.io/blog/structured-logging-and-your-team) - Understanding when and what to log
- [Log Levels Guide](https://www.loggly.com/ultimate-guide/logging-levels/) - Choosing appropriate log levels

### Additional Resources (Optional)
- [Dependency Injection in Go](https://blog.drewolson.org/dependency-injection-in-go) - Understanding the pattern used here
- [Observability Engineering](https://www.oreilly.com/library/view/observability-engineering/9781492076438/) - Deep dive into logging, metrics, and tracing
