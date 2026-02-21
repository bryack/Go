package storage

import (
	"database/sql"
	"math"
	"time"

	_ "modernc.org/sqlite"
)

// ConnectionConfig defines database connection pool settings.
// All fields control connection lifecycle and resource limits.
type ConnectionConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// ConnectionManager holds a database connection and its configuration.
// It provides a structured way to manage database resources.
type ConnectionManager struct {
	db     *sql.DB
	config ConnectionConfig
}

// CreateConnection establishes a SQLite database connection with retry logic.
// It applies connection pool settings and tests connectivity before returning.
func CreateConnection(config *ConnectionConfig, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_busy_timeout=5000")
	if err != nil {
		return nil, mapSQLiteError(err)
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, mapSQLiteError(err)
	}

	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	return db, nil
}

// Retry executes an operation with exponential backoff on failure.
// It attempts the operation up to maxAttempts times with 1ms, 5s, 25s delays.
func Retry[T any](operations func() (T, error), maxAttempts int) (T, error) {
	var zero T
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := operations()

		if err == nil {
			return result, nil
		}

		if attempt == maxAttempts {
			return zero, err
		}

		backoff := time.Duration(math.Pow(5, float64(attempt-1))) * time.Millisecond
		time.Sleep(backoff)
	}

	// Technically unreachable code, but Go requires a return
	return zero, nil
}
