package storage

import (
	"database/sql"
)

const (
	createSchemaMigrationsTable = `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version INTEGER PRIMARY KEY,
            applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );`
)

// Migration represents a database schema change with version control.
// It contains SQL statements for both applying and rolling back the change.
type Migration struct {
	Version int
	Name    string
	Up      string // SQL for applying the migration
	Down    string // SQL for rolling back the migration
}

// Migrator manages database schema migrations and tracks applied versions.
// It provides methods to apply, rollback, and query migration status.
type Migrator struct {
	db         *sql.DB
	migrations []Migration
}

// NewMigrator creates a new migration manager for the given database connection.
// It initializes an empty migration list ready for adding schema changes.
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:         db,
		migrations: make([]Migration, 0),
	}
}

func NewMigratorWithDefaults(db *sql.DB) *Migrator {
	migrator := NewMigrator(db)

	initialMigration := Migration{
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
            CREATE INDEX idx_tasks_created_at ON tasks(created_at);
        `,
		Down: `
            DROP INDEX IF EXISTS idx_tasks_created_at;
            DROP INDEX IF EXISTS idx_tasks_done;
            DROP TABLE IF EXISTS tasks;
        `,
	}

	migrator.AddMigration(initialMigration)
	return migrator
}

func (m *Migrator) ApplyMigrations() error {
	if _, err := m.db.Exec(createSchemaMigrationsTable); err != nil {
		return mapSQLiteError(err)
	}

	current, err := m.GetCurrentVersion()
	if err != nil {
		return mapSQLiteError(err)
	}

	// Find pending migrations
	var pendingMigrations []Migration
	for _, migration := range m.migrations {
		if migration.Version > current {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	if len(pendingMigrations) == 0 {
		return nil
	}

	for _, migration := range pendingMigrations {
		tx, err := m.db.Begin()
		if err != nil {
			return mapSQLiteError(err)
		}

		_, err = tx.Exec(migration.Up)
		if err != nil {
			tx.Rollback()
			return mapSQLiteError(err)
		}

		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", migration.Version)
		if err != nil {
			tx.Rollback()
			return mapSQLiteError(err)
		}

		tx.Commit()
	}

	return nil
}

func (m *Migrator) GetCurrentVersion() (int, error) {
	if _, err := m.db.Exec(createSchemaMigrationsTable); err != nil {
		return 0, mapSQLiteError(err)
	}

	var version sql.NullInt64
	err := m.db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, mapSQLiteError(err)
	}

	if !version.Valid {
		return 0, nil
	}
	return int(version.Int64), nil
}

func (m *Migrator) AddMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
}
