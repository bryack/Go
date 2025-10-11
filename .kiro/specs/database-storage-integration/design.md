# Database Storage Integration Design

## Overview

This design replaces the current JSON file storage with SQLite database storage while maintaining the existing Storage interface. The implementation will use Go's `database/sql` package with the `modernc.io/sqlite` driver for a pure Go solution that requires no CGO dependencies. The design emphasizes connection management, schema migrations, and error handling to create a production-ready storage layer.

## Architecture

### Database Layer Structure
```
storage/
├── database.go          # SQLite implementation of Storage interface
├── migrations.go        # Schema migration management
├── connection.go        # Database connection and pool management
└── errors.go           # Database-specific error types
```

### Database Schema
```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_done ON tasks(done);
CREATE INDEX idx_tasks_created_at ON tasks(created_at);
```

### Migration System
```sql
CREATE TABLE schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## Components and Interfaces

### DatabaseStorage Implementation
```go
type DatabaseStorage struct {
    db *sql.DB
    migrator *Migrator
}

func NewDatabaseStorage(dbPath string) (*DatabaseStorage, error)
func (d *DatabaseStorage) LoadTasks() ([]task.Task, error)
func (d *DatabaseStorage) SaveTasks(tasks []task.Task) error
func (d *DatabaseStorage) Close() error
```

### Connection Management
```go
type ConnectionManager struct {
    db *sql.DB
    config ConnectionConfig
}

type ConnectionConfig struct {
    MaxOpenConns    int
    MaxIdleConns    int
    ConnMaxLifetime time.Duration
    ConnMaxIdleTime time.Duration
}
```

### Migration System
```go
type Migrator struct {
    db *sql.DB
    migrations []Migration
}

type Migration struct {
    Version int
    Name    string
    Up      string
    Down    string
}
```

## Data Models

### Task Table Mapping
The existing `task.Task` struct maps directly to the database schema:
- `ID` → `id` (INTEGER PRIMARY KEY)
- `Description` → `description` (TEXT NOT NULL)
- `Done` → `done` (BOOLEAN NOT NULL)

Additional database fields for auditing:
- `created_at` (DATETIME) - automatically set on insert
- `updated_at` (DATETIME) - automatically updated on modification

### Query Patterns
```go
// Load all tasks
SELECT id, description, done FROM tasks ORDER BY id

// Save tasks (replace all)
BEGIN TRANSACTION;
DELETE FROM tasks;
INSERT INTO tasks (id, description, done) VALUES (?, ?, ?);
COMMIT;

// Individual operations (future enhancement)
INSERT INTO tasks (description, done) VALUES (?, ?)
UPDATE tasks SET description = ?, done = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?
DELETE FROM tasks WHERE id = ?
```

## Error Handling

### Database-Specific Errors
```go
var (
    ErrDatabaseConnection = errors.New("database connection failed")
    ErrMigrationFailed   = errors.New("database migration failed")
    ErrConstraintViolation = errors.New("database constraint violation")
    ErrDatabaseLocked    = errors.New("database is locked")
    ErrDiskFull         = errors.New("database disk full")
)
```

### Error Classification
1. **Connection Errors**: Network issues, file permissions, disk space
2. **Migration Errors**: Schema version conflicts, SQL syntax errors
3. **Query Errors**: Constraint violations, data type mismatches
4. **Concurrency Errors**: Lock timeouts, busy database

### Retry Strategy
- Connection failures: 3 retries with exponential backoff (1s, 2s, 4s)
- Lock timeouts: 5 retries with 100ms intervals
- Migration failures: No retry (fail fast)

## Testing Strategy

### Unit Tests
- Database connection establishment and configuration
- Migration system functionality (up/down migrations)
- CRUD operations with various data scenarios
- Error handling for different failure modes
- Connection pool behavior under load

### Integration Tests
- Full Storage interface compatibility with existing code
- Concurrent access scenarios with multiple goroutines
- Database file creation and schema initialization
- Migration from JSON storage to database storage
- Performance comparison with JSON storage

### Test Database Management
- Use in-memory SQLite (`:memory:`) for fast unit tests
- Use temporary files for integration tests
- Automatic cleanup of test databases
- Isolated test environments to prevent interference

## Performance Considerations

### Connection Pooling
- MaxOpenConns: 25 (suitable for typical web applications)
- MaxIdleConns: 5 (balance between resource usage and performance)
- ConnMaxLifetime: 1 hour (prevent stale connections)
- ConnMaxIdleTime: 15 minutes (release unused connections)

### Query Optimization
- Use prepared statements for repeated queries
- Implement proper indexing on frequently queried columns
- Use transactions for batch operations
- Consider connection reuse for multiple operations

### Database File Management
- Default location: `./tasks.db` (same directory as executable)
- WAL mode for better concurrent access
- Automatic VACUUM on startup if database size > 10MB
- Configurable database path via environment variable

## Migration Path

### Backward Compatibility
1. Keep existing JSON storage as fallback option
2. Provide migration utility to convert JSON data to database
3. Support both storage types during transition period
4. Environment variable to choose storage backend

### Data Migration Process
1. Detect existing `tasks.json` file
2. Load tasks using existing JSON storage
3. Initialize database and create schema
4. Insert all tasks into database
5. Backup original JSON file as `tasks.json.backup`
6. Switch to database storage for future operations

## Security Considerations

### SQL Injection Prevention
- Use parameterized queries exclusively
- Validate all input data before database operations
- Implement proper escaping for dynamic query construction

### File System Security
- Set appropriate file permissions (0600) for database file
- Validate database file path to prevent directory traversal
- Handle database file access errors gracefully

### Data Integrity
- Use database transactions for multi-step operations
- Implement foreign key constraints where applicable
- Regular integrity checks during application startup