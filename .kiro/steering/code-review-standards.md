---
inclusion: always
---

# Code Review Standards for To-Do List Project

This document defines the code review requirements and patterns that must be followed in this Go project.

## Code Review Focus

When conducting code reviews, prioritize actionable feedback:

- **Brief acknowledgment**: A couple of short sentences acknowledging strengths is sufficient
- **Focus on improvements**: Spend most time on:
  - Identified errors and bugs
  - Inaccuracies or potential issues
  - Specific suggestions for fixes and optimizations
- **Goal**: Make information about what needs to be changed as clear and helpful as possible

Avoid lengthy lists of what was done correctly. The developer knows what they implemented - they need to know what needs fixing.

## 1. Stateless Services

**Rule**: Ensure `TaskManager` is stateless; all state lives in `Storage` (SQLite as single source of truth).

- ❌ **Avoid**: In-memory slices, mutexes, or any state stored in service structs
- ✅ **Prefer**: All data operations go through the `Storage` layer
- **Rationale**: Single source of truth prevents data inconsistency and race conditions

```go
// ❌ Bad: Stateful TaskManager
type TaskManager struct {
    tasks []Task
    mu    sync.Mutex
}

// ✅ Good: Stateless TaskManager (current pattern)
type TaskManager struct {
    s      storage.Storage
    writer io.Writer
}
```

## 2. Granular CRUD Operations

**Rule**: Use individual DB operations (`CreateTask`, `UpdateTask`, `DeleteTask`) instead of bulk `SaveTasks`.

- ✅ **Required**: Validate with `RowsAffected()` for 404 on zero rows
- ❌ **Avoid**: Bulk save/load operations that obscure individual failures
- **Rationale**: Granular operations provide better error handling and HTTP status codes

```go
// ✅ Good: Check rows affected
result, err := db.Exec("UPDATE tasks SET ... WHERE id = ?", id)
if err != nil {
    return fmt.Errorf("update failed: %w", err)
}
rows, _ := result.RowsAffected()
if rows == 0 {
    return ErrTaskNotFound // Return 404
}
```

## 3. Dependency Injection

**Rule**: Pass `Storage` to `TaskManager` and handlers via constructor; avoid globals.

- ✅ **Required**: Constructor functions that accept dependencies
- ❌ **Avoid**: Global variables, package-level state
- **Rationale**: Testability, flexibility, and explicit dependencies

```go
// ✅ Good: Dependency injection (current pattern)
func NewTaskManager(s storage.Storage, writer io.Writer) *TaskManager {
    return &TaskManager{s: s, writer: writer}
}

// Handler functions receive storage directly
func tasksHandler(s storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Use s for operations
    }
}
```

## 4. CLI Structure

**Rule**: Use `CLI` struct with methods (`RunLoop`, `promptForTaskID`); keep `main.go` minimal.

- ✅ **Required**: CLI struct with methods for different operations
- ✅ **Required**: Minimal `main.go` that just initializes and runs
- **Rationale**: Testability and separation of concerns

```go
// ✅ Good: Structured CLI (current pattern)
type CLI struct {
    input       InputReader
    output      io.Writer
    taskManager *task.TaskManager
    storage     *storage.DatabaseStorage
}

func (cli *CLI) RunLoop() { /* ... */ }
func (cli *CLI) promptForTaskID(prompt string) (int, error) { /* ... */ }
func (cli *CLI) handleAddCommand() error { /* ... */ }

// main.go stays minimal (current pattern)
func main() {
    s, err := storage.NewDatabaseStorage(dbPath)
    // ... error handling
    cli := NewCLI(NewConsoleInputReader(os.Stdin), os.Stdout, 
                  task.NewTaskManager(s, os.Stdout), s)
    cli.RunLoop()
}
```

## 5. Integration Tests

**Rule**: Use in-memory SQLite (`:memory:`) for DB tests; use `sqlmock` for HTTP handler tests.

- ✅ **Required**: In-memory databases for integration tests (faster than temp files)
- ✅ **Required**: Use `sqlmock` for HTTP handler tests
- ✅ **Required**: Helper functions like `seedTask` for test data setup
- **Rationale**: Isolated, repeatable tests without side effects

```go
// ✅ Good: In-memory DB (current pattern)
func TestAddTask(t *testing.T) {
    s, err := storage.NewDatabaseStorage(":memory:")
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }
    tm := NewTaskManager(s, &strings.Builder{})
    
    // Seed initial data if needed
    for _, it := range initialTasks {
        seedTask(t, s, it)
    }
    
    // ... test code
}

// Helper for seeding test data
func seedTask(t *testing.T, s storage.Storage, task storage.Task) {
    _, err := s.CreateTask(task, testUserID)
    if err != nil {
        t.Fatalf("Failed to create task: %v", err)
    }
}
```

## 6. Error Handling

**Rule**: Wrap errors with context (`fmt.Errorf("%w: %v", err, msg)`); return proper HTTP statuses.

- ✅ **Required**: Wrap errors to preserve error chain
- ✅ **Required**: Return 404 if task/user not found
- ✅ **Required**: Return 400 for validation errors
- ✅ **Required**: Return 500 for internal errors
- **Rationale**: Better debugging and proper REST semantics

```go
// ✅ Good: Error wrapping and proper status codes
if err := storage.GetTask(id); err != nil {
    if errors.Is(err, ErrTaskNotFound) {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }
    http.Error(w, fmt.Errorf("failed to get task: %w", err).Error(), 
               http.StatusInternalServerError)
    return
}
```

## 7. Database Migrations

**Rule**: Separate schema changes (v1 tasks, v2 users, v3 associations); use `IF NOT EXISTS` for checks.

- ✅ **Required**: Version-numbered migrations (v1, v2, v3, etc.)
- ✅ **Required**: Use `IF NOT EXISTS` for idempotent table creation
- ✅ **Required**: Track applied migrations in `schema_migrations` table
- ❌ **Avoid**: Using `COUNT(*)` for existence checks
- **Rationale**: Safe, repeatable migrations that can run multiple times

```go
// ✅ Good: Versioned migrations (current pattern)
Migration{
    Version: 1,
    Name:    "create_tasks_table",
    Up: `
        CREATE TABLE tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            description TEXT NOT NULL,
            done BOOLEAN NOT NULL DEFAULT FALSE,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX idx_tasks_done ON tasks(done);
    `,
    Down: `DROP TABLE IF EXISTS tasks;`,
}

Migration{
    Version: 2,
    Name:    "create_users_table",
    Up: `
        CREATE TABLE users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            email TEXT NOT NULL UNIQUE,
            password_hash TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX idx_users_email ON users(email);
    `,
}

Migration{
    Version: 3,
    Name:    "task_user_association_table",
    Up: `
        ALTER TABLE tasks ADD COLUMN user_id INTEGER;
        CREATE INDEX idx_tasks_user_id ON tasks(user_id);
    `,
}
```

## 8. Auth Readiness

**Rule**: JWT middleware validates tokens; all Storage methods require `userID` parameter.

- ✅ **Required**: All Storage CRUD methods accept `userID` parameter
- ✅ **Required**: Include `user_id` in WHERE clauses for all queries
- ✅ **Required**: Extract `userID` from context in handlers using `auth.GetUserIDFromContext`
- ✅ **Required**: Protect endpoints with `authMiddleware.Authenticate` wrapper
- **Rationale**: Security and multi-tenancy support

```go
// ✅ Good: Storage methods with userID (current pattern)
func (ds *DatabaseStorage) GetTaskByID(id int, userID int) (Task, error) {
    var task Task
    err := ds.db.QueryRow(
        "SELECT id, description, done FROM tasks WHERE id = ? AND user_id = ?",
        id, userID,
    ).Scan(&task.ID, &task.Description, &task.Done)
    if err == sql.ErrNoRows {
        return task, ErrTaskNotFound
    }
    return task, mapSQLiteError(err)
}

func (ds *DatabaseStorage) UpdateTask(task Task, userID int) error {
    result, err := ds.db.Exec(
        "UPDATE tasks SET description = ?, done = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?",
        task.Description, task.Done, task.ID, userID,
    )
    // Check RowsAffected for 404
    if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
        return ErrTaskNotFound
    }
    return mapSQLiteError(err)
}

// ✅ Good: Handler extracts userID from context (current pattern)
func tasksHandler(s storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, err := auth.GetUserIDFromContext(r.Context())
        if err != nil {
            handlers.JSONError(w, http.StatusBadRequest, err.Error())
            return
        }
        
        tasks, err := s.LoadTasks(userID)
        // ... handle response
    }
}

// ✅ Good: Protected routes (current pattern in main.go)
http.HandleFunc("/tasks", logRequest(authMiddleware.Authenticate(tasksHandler(s))))
```

## Review Checklist

When reviewing code, verify:

- [ ] Services are stateless (no in-memory state)
- [ ] CRUD operations are granular with `RowsAffected()` checks
- [ ] Dependencies are injected via constructors
- [ ] CLI code is structured with methods, minimal main.go
- [ ] Tests use in-memory SQLite (`:memory:`)
- [ ] Errors are wrapped with context (`fmt.Errorf`)
- [ ] HTTP status codes are appropriate (404, 400, 500)
- [ ] Migrations are versioned and idempotent
- [ ] User ownership is validated in queries
- [ ] All Storage method calls include `userID` parameter

## Known Issues to Fix

**Current compilation errors** (as of latest check):
- `task/operations.go`: TaskManager methods calling Storage without `userID` parameter
- `task/operations_test.go`: Test helper `seedTask` and test cases missing `userID` parameter
- `cmd/cli/cli.go`: CLI methods calling Storage without `userID` parameter

These need to be updated to match the auth-ready Storage interface that requires `userID` for all operations.
