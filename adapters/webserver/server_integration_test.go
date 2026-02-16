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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreatingTasksAndRetrievingThem(t *testing.T) {
	expectedTasks := []domain.Task{
		{ID: 1, Description: "task 1"},
		{ID: 2, Description: "task 2"},
		{ID: 3, Description: "task 3"},
	}

	server, token := setupIntegrationTest(t)

	server.ServeHTTP(httptest.NewRecorder(), createTaskRequest(t, "task 1", token))
	server.ServeHTTP(httptest.NewRecorder(), createTaskRequest(t, "task 2", token))
	server.ServeHTTP(httptest.NewRecorder(), createTaskRequest(t, "task 3", token))

	response := httptest.NewRecorder()
	server.ServeHTTP(response, loadTasksRequest(t, token))

	assert.Equal(t, http.StatusOK, response.Code)

	got := webserver.LoadTasksResponse(t, response.Body)
	assert.Equal(t, expectedTasks, got)
	assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
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
		Level:       "info",
		Format:      "text",
		Output:      "stderr",
		ServiceName: "test-service",
		Environment: "test",
	})
	if err != nil {
		t.Fatalf("failed to configure logger: %v", err)
	}

	store, err := storage.NewDatabaseStorage(":memory:", testLogger)
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
