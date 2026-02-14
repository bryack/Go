package storage

import (
	"errors"

	"modernc.org/sqlite"
)

var (
	ErrDatabaseConnection  = errors.New("database connection failed")
	ErrMigrationFailed     = errors.New("database migration failed")
	ErrConstraintViolation = errors.New("database constraint violation")
	ErrDatabaseLocked      = errors.New("database is locked")
	ErrDiskFull            = errors.New("database disk is full")
)

// mapSQLiteError converts SQLite-specific errors to custom error types.
// It uses string matching to identify common SQLite error conditions.
func mapSQLiteError(err error) error {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() {
		case 5: // SQLITE_BUSY
			return ErrDatabaseLocked
		case 19: // SQLITE_CONSTRAINT
			return ErrConstraintViolation
		case 13: // SQLITE_FULL
			return ErrDiskFull
		default:
			return ErrDatabaseConnection
		}
	}
	return ErrDatabaseConnection
}
