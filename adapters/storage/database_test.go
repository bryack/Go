package storage

import (
	"fmt"
	"myproject/internal/domain"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateTask(t *testing.T) {

	t.Run("successfully creates task for valid user", func(t *testing.T) {
		store := setupTestStore(t)
		userID := createTestUser(t, store)

		task := domain.Task{Description: "task 1"}
		taskID, err := store.CreateTask(task, userID)
		assert.NoError(t, err)

		description, done := getTaskDescriptionAndDone(t, store, taskID)
		assert.Equal(t, "task 1", description)
		assert.False(t, done)
		assert.NotZero(t, taskID)
	})
	t.Run("fails when user does not exist", func(t *testing.T) {
		store := setupTestStore(t)
		task := domain.Task{Description: "task 1"}
		_, err := store.CreateTask(task, 99999)
		assert.Error(t, err)
	})
}

func setupTestStore(t *testing.T) *DatabaseStorage {
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
	return store
}

func getTaskDescriptionAndDone(t *testing.T, store *DatabaseStorage, taskID int) (description string, done bool) {
	t.Helper()
	err := store.db.QueryRow("SELECT description, done FROM tasks WHERE id = ?", taskID).Scan(&description, &done)
	assert.NoError(t, err)
	return description, done
}

func createTestUser(t *testing.T, store *DatabaseStorage) (userID int) {
	t.Helper()
	query := `INSERT INTO users(email, password_hash) VALUES(?, ?)`
	email := fmt.Sprintf("test-%d@email.com", time.Now().UnixNano())
	res, err := store.db.Exec(query, email, "password_hash")
	assert.NoError(t, err)
	id, err := res.LastInsertId()
	assert.NoError(t, err)
	return int(id)
}
