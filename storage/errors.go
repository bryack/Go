package storage

import "errors"

var (
	ErrDatabaseConnection  = errors.New("database connection failed")
	ErrMigrationFailed     = errors.New("database migration failed")
	ErrConstraintViolation = errors.New("database constraint violation")
	ErrDatabaseLocked      = errors.New("database is locked")
	ErrDiskFull            = errors.New("database disk is full")
)
