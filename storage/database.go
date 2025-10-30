package storage

import (
	"database/sql"
	"errors"
	"os"
	"time"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

// Task represents a single task with unique ID, description, and completion status.
// All fields are JSON-serializable for API responses.
type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

// Storage defines the interface for task persistence operations.
type Storage interface {
	LoadTasks() ([]Task, error)
	GetTaskByID(id int) (task Task, err error)
	CreateTask(task Task) (int, error)
	UpdateTask(task Task) error
	DeleteTask(id int) error
	SaveTasks(tasks []Task) error
}

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

// CreateTask inserts a new task into the database and returns the generated ID.
// The database AUTOINCREMENT feature assigns the ID automatically.
// Timestamps (created_at, updated_at) are set to current time on creation.
func (ds *DatabaseStorage) CreateTask(task Task) (int, error) {
	result, err := ds.db.Exec(
		"INSERT INTO tasks (description, done, created_at, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		task.Description, task.Done,
	)
	if err != nil {
		return 0, mapSQLiteError(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, mapSQLiteError(err)
	}
	return int(id), nil
}

func (ds *DatabaseStorage) UpdateTask(task Task) error {
	result, err := ds.db.Exec(
		"UPDATE tasks SET description = ?, done = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		task.Description, task.Done, task.ID,
	)
	if err != nil {
		return mapSQLiteError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return mapSQLiteError(err)
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (ds *DatabaseStorage) DeleteTask(id int) error {
	result, err := ds.db.Exec(
		"DELETE FROM tasks WHERE id = ?",
		id,
	)
	if err != nil {
		return mapSQLiteError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return mapSQLiteError(err)
	}

	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (ds *DatabaseStorage) GetTaskByID(id int) (task Task, err error) {
	err = ds.db.QueryRow(
		"SELECT id, description, done FROM tasks WHERE id = ?",
		id,
	).Scan(&task.ID, &task.Description, &task.Done)

	if err != nil {
		if err == sql.ErrNoRows {
			return task, ErrTaskNotFound
		}
		return task, mapSQLiteError(err)
	}

	return task, nil
}

// LoadTasks retrieves all tasks from the database ordered by ID.
// Returns an empty slice if no tasks exist, never returns nil.
func (ds *DatabaseStorage) LoadTasks() ([]Task, error) {
	query := "SELECT id, description, done FROM tasks ORDER BY id"
	rows, err := ds.db.Query(query)
	if err != nil {
		return nil, mapSQLiteError(err)
	}

	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
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
func (ds *DatabaseStorage) SaveTasks(tasks []Task) error {
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
