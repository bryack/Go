package webserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"myproject/auth"
	"myproject/infrastructure/testhelpers"
	"myproject/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dummyAuthMiddleware = &auth.AuthMiddleware{}
	dummyLogger         = slog.New(slog.NewTextHandler(io.Discard, nil))
)

type StubAuth struct {
	authCalled int
}

func (sa *StubAuth) Authenticate(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sa.authCalled++
		ctx := context.WithValue(r.Context(), auth.UserIDKey, 1)
		r = r.WithContext(ctx)
		handler(w, r)
	}
}

type StubAuthService struct {
	RegisterCalled []RegisterRequest
	LoginCalled    []string
}

func (sas *StubAuthService) Register(email, password string) (token string, err error) {
	sas.RegisterCalled = append(sas.RegisterCalled, RegisterRequest{email, password})
	return "", nil
}

func (sas *StubAuthService) Login(email, password string) (token string, err error) {
	sas.LoginCalled = append(sas.LoginCalled, email)
	return "", nil
}

func TestHealth(t *testing.T) {
	t.Run("returns status healthy", func(t *testing.T) {
		store := &testhelpers.StubTaskStore{}
		authService := &StubAuthService{}
		svr := NewTasksServer(store, authService, dummyAuthMiddleware, dummyLogger)
		request, err := http.NewRequest(http.MethodGet, "/health", nil)
		assert.NoError(t, err)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		var health HealthResponse
		err = json.NewDecoder(response.Body).Decode(&health)
		want := "healthy"

		assert.Equal(t, want, health.Status)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
	})
}

func TestRoot(t *testing.T) {

	t.Run("returns 200 on /", func(t *testing.T) {
		store := &testhelpers.StubTaskStore{}
		authService := &StubAuthService{}
		svr := NewTasksServer(store, authService, dummyAuthMiddleware, dummyLogger)
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.NoError(t, err)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
	})
}

func TestGetTaskByID(t *testing.T) {
	store := &testhelpers.StubTaskStore{
		Tasks: map[int]string{
			1: "task 1",
			2: "task 2",
		},
	}
	authService := &StubAuthService{}
	auth := &StubAuth{}

	tests := []struct {
		name                string
		url                 string
		expectedDescription string
		expectedStatus      int
	}{
		{
			name:                "returns task by ID 1",
			url:                 "/tasks/1",
			expectedDescription: "task 1",
			expectedStatus:      http.StatusOK,
		},
		{
			name:                "returns task by ID 2",
			url:                 "/tasks/2",
			expectedDescription: "task 2",
			expectedStatus:      http.StatusOK,
		},
		{
			name:                "returns 404 on nonexistense task",
			url:                 "/tasks/404",
			expectedDescription: "",
			expectedStatus:      http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		auth.authCalled = 0
		svr := NewTasksServer(store, authService, auth, dummyLogger)
		request := getTaskByIDRequest(t, tt.url)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		task := domain.Task{}
		err := json.NewDecoder(response.Body).Decode(&task)
		assert.NoError(t, err)

		assert.Equal(t, tt.expectedStatus, response.Code)
		assert.Equal(t, tt.expectedDescription, task.Description)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	}
}

func getTaskByIDRequest(t *testing.T, url string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	return request
}

func TestCreateTask(t *testing.T) {
	store := &testhelpers.StubTaskStore{}
	auth := &StubAuth{authCalled: 0}
	authService := &StubAuthService{}
	svr := NewTasksServer(store, authService, auth, dummyLogger)
	t.Run("returns 201 on POST", func(t *testing.T) {
		request := createTaskRequest(t, "task 1")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		task := domain.Task{}
		err := json.NewDecoder(response.Body).Decode(&task)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.Code)
		assert.Equal(t, store.CreateCall[0], task.ID)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	})
	t.Run("returns 400 on empty description", func(t *testing.T) {
		auth.authCalled = 0
		request := createTaskRequest(t, "")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		task := domain.Task{}
		err := json.NewDecoder(response.Body).Decode(&task)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.True(t, task == domain.Task{})
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	})
}

func createTaskRequest(t *testing.T, desription string) *http.Request {
	t.Helper()
	task := domain.Task{Description: desription}
	jsonTask, err := json.Marshal(task)

	request, err := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))
	request.Header.Set("Content-Type", "application/json")
	assert.NoError(t, err)
	return request
}

func TestLoadTasks(t *testing.T) {
	t.Run("returns tasks on GET /tasks", func(t *testing.T) {
		tasksList := []domain.Task{
			{Description: "task 1"},
			{Description: "task 2"},
			{Description: "task 3"},
		}
		store := &testhelpers.StubTaskStore{Tasks: nil, CreateCall: nil, TasksTable: tasksList}
		auth := &StubAuth{authCalled: 0}
		authService := &StubAuthService{}
		svr := NewTasksServer(store, authService, auth, dummyLogger)
		request := loadTasksRequest(t)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		expectedDescription := []string{"task 1", "task 2", "task 3"}
		got := HandleLoadTasksResponse(t, response.Body)
		assert.Equal(t, http.StatusOK, response.Code)
		assert.ElementsMatch(t, expectedDescription, got)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	})
}

func loadTasksRequest(t *testing.T) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, "/tasks", nil)
	assert.NoError(t, err)
	return request
}

func HandleLoadTasksResponse(t testing.TB, body io.Reader) (descriptions []string) {
	t.Helper()
	tasks := []domain.Task{}
	err := json.NewDecoder(body).Decode(&tasks)

	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of Tasks, '%v'", body, err)
	}
	descriptions = make([]string, len(tasks))
	for i, task := range tasks {
		descriptions[i] = task.Description
	}

	return
}

func TestUpdateTask(t *testing.T) {
	store := &testhelpers.StubTaskStore{
		Tasks: map[int]string{
			1: "task 1",
			2: "task 2",
		},
	}
	authService := &StubAuthService{}

	t.Run("update task 1", func(t *testing.T) {
		auth := &StubAuth{authCalled: 0}
		svr := NewTasksServer(store, authService, auth, dummyLogger)

		request := updateTaskRequest(t, "/tasks/1", "new task 1")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		assert.Equal(t, "new task 1", store.Tasks[1])
		assert.Equal(t, http.StatusOK, response.Code)

		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	})
	t.Run("returns 400 on empty description", func(t *testing.T) {
		auth := &StubAuth{authCalled: 0}
		svr := NewTasksServer(store, authService, auth, dummyLogger)

		request := updateTaskRequest(t, "/tasks/1", "")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		assert.Equal(t, http.StatusBadRequest, response.Code)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	})
	t.Run("returns 404, if task not found", func(t *testing.T) {
		auth := &StubAuth{authCalled: 0}
		svr := NewTasksServer(store, authService, auth, dummyLogger)

		request := updateTaskRequest(t, "/tasks/404", "new task 404")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		assert.Equal(t, "new task 1", store.Tasks[1])
		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	})
}

func updateTaskRequest(t *testing.T, url, description string) *http.Request {
	t.Helper()
	task := domain.Task{ID: 1, Description: description}
	jsonTask, err := json.Marshal(task)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonTask))
	request.Header.Set("Content-Type", "application/json")
	assert.NoError(t, err)
	return request
}

func TestDeleteTask(t *testing.T) {
	store := &testhelpers.StubTaskStore{
		Tasks: map[int]string{
			1: "task 1",
			2: "task 2",
		},
	}
	t.Run("delete task 1", func(t *testing.T) {
		auth := &StubAuth{authCalled: 0}
		authService := &StubAuthService{}
		svr := NewTasksServer(store, authService, auth, dummyLogger)

		request := deleteTaskRequest(t)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		_, ok := store.Tasks[1]
		assert.True(t, !ok)
		assert.Equal(t, http.StatusNoContent, response.Code)
		assert.Equal(t, 1, auth.authCalled)
	})
}

func deleteTaskRequest(t *testing.T) *http.Request {
	t.Helper()

	request, err := http.NewRequest(http.MethodDelete, "/tasks/1", nil)
	assert.NoError(t, err)
	return request
}

func TestRegister(t *testing.T) {

	t.Run("register test email", func(t *testing.T) {
		store := &testhelpers.StubTaskStore{}
		auth := &StubAuth{}
		authService := &StubAuthService{}
		svr := NewTasksServer(store, authService, auth, dummyLogger)

		request := registerRequest(t)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		assert.Equal(t, http.StatusCreated, response.Code)
		assert.Equal(t, RegisterRequest{"test@email.com", "test_pass"}, authService.RegisterCalled[0])
	})
}

func registerRequest(t *testing.T) *http.Request {
	t.Helper()
	reg := RegisterRequest{
		Email:    "test@email.com",
		Password: "test_pass",
	}
	jsonUser, err := json.Marshal(reg)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonUser))
	request.Header.Set("Content-Type", "application/json")
	assert.NoError(t, err)
	return request
}

func TestLogin(t *testing.T) {

	t.Run("login test email", func(t *testing.T) {
		store := &testhelpers.StubTaskStore{}
		auth := &StubAuth{}
		authService := &StubAuthService{}
		authService.RegisterCalled = []RegisterRequest{{"test@email.com", "test_pass"}}
		svr := NewTasksServer(store, authService, auth, dummyLogger)

		request := loginRequest(t)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "test@email.com", authService.LoginCalled[0])
	})
}

func loginRequest(t *testing.T) *http.Request {
	t.Helper()
	reg := RegisterRequest{
		Email:    "test@email.com",
		Password: "test_pass",
	}
	jsonUser, err := json.Marshal(reg)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonUser))
	request.Header.Set("Content-Type", "application/json")
	assert.NoError(t, err)
	return request
}

func TestLoggingMiddleware(t *testing.T) {
	var logBuffer bytes.Buffer
	testLogger := slog.New(slog.NewJSONHandler(&logBuffer, nil))
	store := &testhelpers.StubTaskStore{}
	authService := &StubAuthService{}
	auth := &StubAuth{}

	svr := NewTasksServer(store, authService, auth, testLogger)

	request, err := http.NewRequest(http.MethodGet, "/health", nil)
	assert.NoError(t, err)
	response := httptest.NewRecorder()

	svr.ServeHTTP(response, request)

	assert.Contains(t, logBuffer.String(), "HTTP request started")
	assert.Contains(t, logBuffer.String(), "HTTP request completed")
}
