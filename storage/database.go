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

	// // Attempt automatic JSON migration (non-fatal if it fails)
	// if err := storage.autoMigrateFromJSON(dbPath); err != nil {
	// 	log.Printf("Warning: JSON migration failed: %v", err)
	// 	// Don't return error - migration failure shouldn't prevent database usage
	// }

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

// // autoMigrateFromJSON automatically migrates tasks from JSON file to database if conditions are met.
// // It only migrates if: 1) JSON file exists, 2) Database is empty, 3) JSON file has tasks
// func (ds *DatabaseStorage) autoMigrateFromJSON(dbPath string) error {
// 	// Determine JSON file path (same directory as database)
// 	dbDir := filepath.Dir(dbPath)
// 	jsonPath := filepath.Join(dbDir, "tasks.json")

// 	// Check if JSON file exists
// 	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
// 		// No JSON file to migrate, this is normal
// 		return nil
// 	}

// 	fmt.Println("📦 Found existing tasks.json, checking if migration is needed...")

// 	// Load tasks from JSON using existing JsonStorage
// 	jsonStorage := JsonStorage{}
// 	tasks, err := jsonStorage.LoadTasks()
// 	if err != nil {
// 		log.Printf("Warning: Could not read tasks.json for migration: %v", err)
// 		return nil // Don't fail, just skip migration
// 	}

// 	if len(tasks) == 0 {
// 		fmt.Println("📦 JSON file is empty, no migration needed")
// 		return nil
// 	}

// 	// Check if database already has tasks (don't overwrite existing data)
// 	existingTasks, err := ds.LoadTasks()
// 	if err != nil {
// 		return err
// 	}

// 	if len(existingTasks) > 0 {
// 		fmt.Printf("📦 Database already has %d tasks, skipping migration\n", len(existingTasks))
// 		return nil
// 	}

// 	// Perform migration: JSON → Database
// 	fmt.Printf("📦 Migrating %d tasks from JSON to database...\n", len(tasks))

// 	if err := ds.SaveTasks(tasks); err != nil {
// 		return fmt.Errorf("failed to save tasks to database during migration: %w", err)
// 	}

// 	// Backup original JSON file
// 	backupPath := jsonPath + ".backup"
// 	if err := os.Rename(jsonPath, backupPath); err != nil {
// 		log.Printf("Warning: Could not backup JSON file: %v", err)
// 		// Don't fail migration for backup issues
// 	} else {
// 		fmt.Printf("📁 Backed up original file as %s\n", filepath.Base(backupPath))
// 	}

// 	fmt.Printf("✅ Successfully migrated %d tasks from JSON to database\n", len(tasks))
// 	return nil
// }
