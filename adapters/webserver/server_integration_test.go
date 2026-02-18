package webserver_test

import (
	"bytes"
	"encoding/json"
	"myproject/adapters/storage"
	"myproject/adapters/webserver"
	"myproject/auth"
	"myproject/internal/domain"
	"myproject/logger"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreatingTasksAndRetrievingThem(t *testing.T) {

	t.Run("successfully creates tasks and retrieves them", func(t *testing.T) {
		server, token := setupIntegrationTest(t)

		server.ServeHTTP(httptest.NewRecorder(), createTaskRequest(t, "integration task 1", token))
		server.ServeHTTP(httptest.NewRecorder(), createTaskRequest(t, "integration task 2", token))
		server.ServeHTTP(httptest.NewRecorder(), createTaskRequest(t, "integration task 3", token))

		response := httptest.NewRecorder()
		server.ServeHTTP(response, loadTasksRequest(t, token))

		assert.Equal(t, http.StatusOK, response.Code)

		got := webserver.HandleLoadTasksResponse(t, response.Body)

		expectedTasks := []string{
			"integration task 1",
			"integration task 2",
			"integration task 3",
		}
		assert.ElementsMatch(t, expectedTasks, got)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
	})
}

func createTaskRequest(t *testing.T, description, token string) *http.Request {
	t.Helper()
	task := domain.Task{Description: description}
	jsonTask, err := json.Marshal(task)

	request, err := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	assert.NoError(t, err)
	return request
}

func loadTasksRequest(t *testing.T, token string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, "/tasks", nil)
	assert.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}

func setupIntegrationTest(t *testing.T) (*webserver.TasksServer, string) {
	testLogger, err := logger.NewLogger(&logger.Config{
		Level:       "error",
		Format:      "text",
		Output:      "stderr",
		ServiceName: "test-service",
		Environment: "test",
	})
	if err != nil {
		t.Fatalf("failed to configure logger: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.NewDatabaseStorage(dbPath, testLogger)
	if err != nil {
		t.Fatalf("failed to create in-memory database: %v", err)
	}

	t.Cleanup(func() {
		store.Close()
	})

	jwtService := auth.NewJWTService("test-secret-key-minimum-32-chars!", 24*time.Hour)
	authService := auth.NewService(store, jwtService, testLogger)
	authMiddleware := auth.NewAuthMiddleware(jwtService, testLogger)

	server := webserver.NewTasksServer(store, authService, authMiddleware, testLogger)

	authService.Register("test@email.com", "password123")
	token, err := authService.Login("test@email.com", "password123")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	return server, token
}

func TestRaceDatabaseStorage(t *testing.T) {
	server, token := setupIntegrationTest(t)

	const concurrentRequests = 100

	var wg sync.WaitGroup

	for range concurrentRequests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response := httptest.NewRecorder()
			server.ServeHTTP(response, createTaskRequest(t, "race task 1", token))
			assert.Equal(t, http.StatusCreated, response.Code)
		}()
	}
	wg.Wait()

	response := httptest.NewRecorder()
	server.ServeHTTP(response, loadTasksRequest(t, token))
	assert.Equal(t, http.StatusOK, response.Code)

	got := webserver.HandleLoadTasksResponse(t, response.Body)
	assert.Equal(t, concurrentRequests, len(got))
}
