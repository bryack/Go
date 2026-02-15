package storage

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateConnection(t *testing.T) {
	cfg := &ConnectionConfig{}
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	db, err := CreateConnection(cfg, dbPath)

	t.Cleanup(func() {
		if db != nil {
			db.Close()
		}
	})
	t.Run("should_successfully_connect_and_ping_db", func(t *testing.T) {
		assert.NoError(t, err)
		assert.NotNil(t, db)
		err = db.Ping()
		assert.NoError(t, err, "Database must be reachable via Ping")
	})

	t.Run("foreign keys constraints", func(t *testing.T) {
		setupQuery := `
            CREATE TABLE users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                email TEXT NOT NULL UNIQUE,
                password_hash TEXT NOT NULL,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            );
            CREATE TABLE tasks (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id INTEGER NOT NULL,
                description TEXT NOT NULL,
                done BOOLEAN NOT NULL DEFAULT FALSE,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY(user_id) REFERENCES users(id)
            );`
		_, err = db.Exec(setupQuery)
		assert.NoError(t, err)

		taskInsertQuery := `INSERT INTO tasks (description, user_id) VALUES (?, ?)`
		_, err = db.Exec(taskInsertQuery, "task 1", "999")
		assert.Error(t, err, "should failed with FOREIGN KEY constraint")
	})
}
