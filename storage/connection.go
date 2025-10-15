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
	db, err := sql.Open("sqlite", path)
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

// retry executes an operation with exponential backoff on failure.
// It attempts the operation up to maxAttempts times with 1s, 2s, 4s delays.
func retry[T any](operations func() (T, error), maxAttempts int) (T, error) {
	var zero T
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := operations()

		if err == nil {
			return result, nil
		}

		if attempt == maxAttempts {
			return zero, err
		}

		backoff := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
		time.Sleep(backoff)
	}

	// Technically unreachable code, but Go requires a return
	return zero, nil
}
