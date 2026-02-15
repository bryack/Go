package storage

import (
	"myproject/internal/domain"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTask(t *testing.T) {

	t.Run("successfully creates task for valid user", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "test.db")

		store, err := NewDatabaseStorage(dbPath, dummyLogger)
		if err != nil {
			t.Fatalf("failed to create DatabaseStorage: %v", err)
		}

		t.Cleanup(func() {
			if store.db != nil {
				store.db.Close()
			}
		})

		query := `INSERT INTO users(email, password_hash) VALUES(?, ?)`
		res, err := store.db.Exec(query, "testCreated@email.com", "password_hash")
		assert.NoError(t, err)
		userID, err := res.LastInsertId()
		assert.NoError(t, err)

		task := domain.Task{Description: "task 1"}
		taskID, err := store.CreateTask(task, int(userID))
		assert.NoError(t, err)

		var description string
		var done bool
		err = store.db.QueryRow("SELECT description, done FROM tasks WHERE id = ?", taskID).Scan(&description, &done)
		assert.NoError(t, err)
		assert.Equal(t, "task 1", description)
		assert.False(t, done)
		assert.NotZero(t, taskID)
	})
}
