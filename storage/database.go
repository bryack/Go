package storage

import (
	"database/sql"
	"myproject/task"
	"os"
	"time"
)

// DatabaseStorage provides SQLite-based task persistence with automatic schema management.
// It implements the Storage interface and handles database connections and migrations.
type DatabaseStorage struct {
	db       *sql.DB
	migrator *Migrator
}

// GetDatabasePath returns the database file path from environment or default location.
// Checks TASK_DB_PATH environment variable, falls back to "./tasks.db".
func GetDatabasePath() string {
	if dbPath := os.Getenv("TASK_DB_PATH"); dbPath != "" {
		return dbPath
	}
	return "./tasks.db"
}

// NewDatabaseStorage creates a new database storage instance with automatic setup.
// It handles connection pooling, schema migrations, and JSON data migration.
func NewDatabaseStorage(dbPath string) (*DatabaseStorage, error) {
	config := ConnectionConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 15 * time.Minute,
	}
	db, err := CreateConnection(&config, dbPath)
	if err != nil {
		return nil, mapSQLiteError(err)
	}

	migrator := NewMigratorWithDefaults(db)

	if err := migrator.ApplyMigrations(); err != nil {
		return nil, err
	}

	// Create storage instance
	storage := &DatabaseStorage{
		db:       db,
		migrator: migrator,
	}
	return storage, nil
}

// LoadTasks retrieves all tasks from the database ordered by ID.
// Returns an empty slice if no tasks exist, never returns nil.
func (ds *DatabaseStorage) LoadTasks() ([]task.Task, error) {
	query := "SELECT id, description, done FROM tasks ORDER BY id"
	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, mapSQLiteError(err)
	}

	defer rows.Close()
	var tasks []task.Task
	for rows.Next() {
		var task task.Task
		if err := rows.Scan(&task.ID, &task.Description, &task.Done); err != nil {
			return nil, mapSQLiteError(err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, mapSQLiteError(err)
	}

	return tasks, nil
}

// SaveTasks replaces all tasks in the database with the provided task slice.
// Uses a transaction to ensure atomic replacement of all task data.
func (ds *DatabaseStorage) SaveTasks(tasks []task.Task) error {
	tx, err := ds.db.Begin()
	if err != nil {
		return mapSQLiteError(err)
	}

	if _, err = tx.Exec("DELETE FROM tasks"); err != nil {
		tx.Rollback()
		return mapSQLiteError(err)
	}

	if len(tasks) == 0 {
		return nil
	}

	query := "INSERT INTO tasks (id, description, done) VALUES "
	values := make([]interface{}, 0, len(tasks)*3)

	for i, task := range tasks {
		if i > 0 {
			query += ","
		}
		query += "(?, ?, ?)"

		values = append(values, task.ID, task.Description, task.Done)
	}

	if _, err := tx.Exec(query, values...); err != nil {
		tx.Rollback()
		return mapSQLiteError(err)
	}

	return tx.Commit()
}
