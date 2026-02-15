package storage

import (
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var dummyLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestNewMigratorWithDefaults(t *testing.T) {

	t.Run("version 4. user's tasks delete on cascade", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")

		store, err := NewDatabaseStorage(dbPath, dummyLogger)
		if err != nil {
			t.Fatal("failed to create DatabaseStorage")
		}
		t.Cleanup(func() {
			if store.db != nil {
				store.db.Close()
			}
		})

		query := `INSERT INTO users(email, password_hash) VALUES (?, ?)`
		res, err := store.db.Exec(query, "test@email.com", "password_hash")

		id, err := res.LastInsertId()
		assert.NoError(t, err)

		query = `INSERT INTO tasks(user_id, description) VALUES (?, ?)`
		_, err = store.db.Exec(query, id, "task 1")
		assert.NoError(t, err)

		query = `DELETE FROM users WHERE id = ?`
		_, err = store.db.Exec(query, id)
		assert.NoError(t, err)

		var count int
		query = `SELECT COUNT(*) FROM tasks`
		store.db.QueryRow(query).Scan(&count)
		assert.True(t, count == 0, "Tasks should be deleted automatically by cascade")
	})
}
