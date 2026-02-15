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

func TestUpdateTask(t *testing.T) {

	t.Run("successfully updates task for valid user", func(t *testing.T) {
		store := setupTestStore(t)
		userID := createTestUser(t, store)

		task := domain.Task{Description: "task 1"}
		taskID, err := store.CreateTask(task, userID)
		assert.NoError(t, err)

		task.Description = "new task description"
		task.Done = true
		task.ID = taskID
		err = store.UpdateTask(task, userID)
		assert.NoError(t, err)

		description, done := getTaskDescriptionAndDone(t, store, taskID)
		assert.Equal(t, "new task description", description)
		assert.True(t, done)
	})
	t.Run("fails when task belongs to different user", func(t *testing.T) {
		store := setupTestStore(t)
		userID := createTestUser(t, store)
		task := domain.Task{Description: "task 1"}
		taskID, err := store.CreateTask(task, userID)
		assert.NoError(t, err)
		task = domain.Task{ID: taskID, Description: "new task description"}

		userID = createTestUser(t, store)
		err = store.UpdateTask(task, userID)
		assert.Error(t, err)
	})
	t.Run("fails when task does not exist", func(t *testing.T) {
		store := setupTestStore(t)
		userID := createTestUser(t, store)
		task := domain.Task{ID: 99999, Description: "task 1"}
		err := store.UpdateTask(task, userID)
		assert.Error(t, err)
	})
}

func TestDeleteTask(t *testing.T) {

	t.Run("successfully deletes task for valid user", func(t *testing.T) {
		store := setupTestStore(t)
		userID := createTestUser(t, store)

		task := domain.Task{Description: "task 1"}
		taskID, err := store.CreateTask(task, userID)
		assert.NoError(t, err)

		err = store.DeleteTask(taskID, userID)
		assert.NoError(t, err)

		var count int
		err = store.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
	})
	t.Run("fails when task belongs to different user", func(t *testing.T) {
		store := setupTestStore(t)
		userID := createTestUser(t, store)
		task := domain.Task{Description: "task 1"}
		taskID, err := store.CreateTask(task, userID)
		assert.NoError(t, err)

		userID = createTestUser(t, store)
		err = store.DeleteTask(taskID, userID)
		assert.Error(t, err)

		var count int
		err = store.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
	})
	t.Run("fails when task does not exist", func(t *testing.T) {
		store := setupTestStore(t)
		userID := createTestUser(t, store)

		err := store.DeleteTask(99999, userID)
		assert.Error(t, err)
	})
}

func TestGetTaskByID(t *testing.T) {

	t.Run("successfully gets task for valid user", func(t *testing.T) {
		store := setupTestStore(t)
		userID := createTestUser(t, store)

		task := domain.Task{Description: "task 1"}
		taskID, err := store.CreateTask(task, userID)
		assert.NoError(t, err)

		task, err = store.GetTaskByID(taskID, userID)
		assert.NoError(t, err)

		assert.Equal(t, "task 1", task.Description)
		assert.Equal(t, taskID, task.ID)
		assert.False(t, task.Done)
	})
	t.Run("fails when task belongs to different user", func(t *testing.T) {
		store := setupTestStore(t)
		userID := createTestUser(t, store)
		task := domain.Task{Description: "task 1"}
		taskID, err := store.CreateTask(task, userID)
		assert.NoError(t, err)

		userID = createTestUser(t, store)
		task, err = store.GetTaskByID(taskID, userID)
		assert.Error(t, err)
	})
	t.Run("fails when task does not exist", func(t *testing.T) {
		store := setupTestStore(t)
		userID := createTestUser(t, store)

		_, err := store.GetTaskByID(99999, userID)
		assert.Error(t, err)
	})
}
