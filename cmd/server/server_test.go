package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"myproject/auth"
	"myproject/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dummyAuthMiddleware = &auth.AuthMiddleware{}
	dummyLogger         = slog.New(slog.NewTextHandler(io.Discard, nil))
)

type StubTaskStore struct {
	tasks      map[int]string
	createCall []int
	tasksTable []storage.Task
}

func (s *StubTaskStore) GetTaskByID(id int, userID int) (task storage.Task, err error) {
	t, ok := s.tasks[id]
	if !ok {
		return storage.Task{}, fmt.Errorf("Task not found")
	}
	return storage.Task{ID: id, Description: t}, nil
}

func (s *StubTaskStore) CreateTask(task storage.Task, userID int) (int, error) {
	s.createCall = append(s.createCall, task.ID)
	return task.ID, nil
}

func (s *StubTaskStore) LoadTasks(userID int) ([]storage.Task, error) {
	return s.tasksTable, nil
}

func (s *StubTaskStore) UpdateTask(task storage.Task, userID int) error {
	s.tasks[task.ID] = task.Description
	return nil
}

func (s *StubTaskStore) DeleteTask(id int, userID int) error {
	delete(s.tasks, id)
	return nil
}

func (s *StubTaskStore) Close() error {
	return nil
}

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

func TestHealth(t *testing.T) {
	t.Run("returns status healthy", func(t *testing.T) {
		store := &StubTaskStore{}
		svr := NewTasksServer(store, dummyAuthMiddleware, dummyLogger)
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
		store := &StubTaskStore{}
		svr := NewTasksServer(store, dummyAuthMiddleware, dummyLogger)
		request, err := http.NewRequest(http.MethodGet, "/", nil)
		assert.NoError(t, err)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		assert.Equal(t, http.StatusOK, response.Code)
	})
}

func TestGetTaskByID(t *testing.T) {
	store := &StubTaskStore{
		tasks: map[int]string{
			1: "task 1",
			2: "task 2",
		},
	}

	t.Run("returns task by ID 1", func(t *testing.T) {
		auth := &StubAuth{}
		svr := NewTasksServer(store, auth, dummyLogger)
		request := getTaskByIDRequest(t, "/tasks/1")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		task := storage.Task{}
		err := json.NewDecoder(response.Body).Decode(&task)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "task 1", task.Description)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	})
	t.Run("returns task by ID 2", func(t *testing.T) {
		auth := &StubAuth{}
		svr := NewTasksServer(store, auth, dummyLogger)
		request := getTaskByIDRequest(t, "/tasks/2")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		task := storage.Task{}
		err := json.NewDecoder(response.Body).Decode(&task)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "task 2", task.Description)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	})
	t.Run("returns 404", func(t *testing.T) {
		auth := &StubAuth{}
		svr := NewTasksServer(store, auth, dummyLogger)
		request := getTaskByIDRequest(t, "/tasks/404")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		assert.Equal(t, http.StatusNotFound, response.Code)
		assert.Equal(t, 1, auth.authCalled)
	})
}

func getTaskByIDRequest(t *testing.T, url string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, url, nil)
	assert.NoError(t, err)
	return request
}

func TestCreateTask(t *testing.T) {
	store := &StubTaskStore{}
	auth := &StubAuth{authCalled: 0}
	svr := NewTasksServer(store, auth, dummyLogger)
	t.Run("returns 201 on POST", func(t *testing.T) {
		request := createTaskRequest(t)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		task := storage.Task{}
		err := json.NewDecoder(response.Body).Decode(&task)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, response.Code)
		assert.Equal(t, store.createCall[0], task.ID)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	})
}

func createTaskRequest(t *testing.T) *http.Request {
	t.Helper()
	task := storage.Task{Description: "task 1"}
	jsonTask, err := json.Marshal(task)

	request, err := http.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(jsonTask))
	request.Header.Set("Content-Type", "application/json")
	assert.NoError(t, err)
	return request
}

func TestLoadTasks(t *testing.T) {
	t.Run("returns tasks on GET /tasks", func(t *testing.T) {
		expectedTasks := []storage.Task{
			{ID: 1, Description: "task 1"},
			{ID: 2, Description: "task 2"},
			{ID: 3, Description: "task 3"},
		}
		store := &StubTaskStore{nil, nil, expectedTasks}
		auth := &StubAuth{authCalled: 0}
		svr := NewTasksServer(store, auth, dummyLogger)
		request := loadTasksRequest(t)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		got := loadTasksResponse(t, response.Body)
		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, expectedTasks, got)
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

func loadTasksResponse(t testing.TB, body io.Reader) (tasks []storage.Task) {
	t.Helper()
	err := json.NewDecoder(body).Decode(&tasks)

	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of Player, '%v'", body, err)
	}

	return
}

func TestUpdateTask(t *testing.T) {
	store := &StubTaskStore{
		tasks: map[int]string{
			1: "task 1",
			2: "task 2",
		},
	}
	t.Run("update task 1", func(t *testing.T) {
		auth := &StubAuth{authCalled: 0}
		svr := NewTasksServer(store, auth, dummyLogger)

		request := updateTaskRequest(t)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		assert.Equal(t, "new task 1", store.tasks[1])
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
		assert.Equal(t, 1, auth.authCalled)
	})
}

func updateTaskRequest(t *testing.T) *http.Request {
	t.Helper()
	task := storage.Task{ID: 1, Description: "new task 1"}
	jsonTask, err := json.Marshal(task)
	assert.NoError(t, err)

	request, err := http.NewRequest(http.MethodPut, "/tasks/1", bytes.NewReader(jsonTask))
	request.Header.Set("Content-Type", "application/json")
	assert.NoError(t, err)
	return request
}

func TestDeleteTask(t *testing.T) {
	store := &StubTaskStore{
		tasks: map[int]string{
			1: "task 1",
			2: "task 2",
		},
	}
	t.Run("update task 1", func(t *testing.T) {
		auth := &StubAuth{authCalled: 0}
		svr := NewTasksServer(store, auth, dummyLogger)

		request := deleteTaskRequest(t)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		_, ok := store.tasks[1]
		assert.True(t, !ok)

		assert.Equal(t, 1, auth.authCalled)
	})
}

func deleteTaskRequest(t *testing.T) *http.Request {
	t.Helper()

	request, err := http.NewRequest(http.MethodDelete, "/tasks/1", nil)
	assert.NoError(t, err)
	return request
}
