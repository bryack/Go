package main

import (
	"encoding/json"
	"io"
	"myproject/storage"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type StubTaskStore struct {
	tasks      map[int]string
	createCall []int
	tasksTable []storage.Task
}

func (s *StubTaskStore) GetTaskByID(id int, userID int) (task storage.Task, err error) {
	t := s.tasks[id]
	return storage.Task{Description: t}, nil
}

func (s *StubTaskStore) CreateTask(task storage.Task, userID int) (int, error) {
	s.createCall = append(s.createCall, task.ID)
	return task.ID, nil
}

func (s *StubTaskStore) LoadTasks(userID int) ([]storage.Task, error) {
	return s.tasksTable, nil
}

func (s *StubTaskStore) UpdateTask(task storage.Task, userID int) error {
	return nil
}

func (s *StubTaskStore) DeleteTask(id int, userID int) error {
	return nil
}

func (s *StubTaskStore) Close() error {
	return nil
}

func TestHealth(t *testing.T) {
	t.Run("returns status healthy", func(t *testing.T) {
		store := &StubTaskStore{}
		svr := NewTasksServer(store)
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
		svr := NewTasksServer(store)
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
	svr := NewTasksServer(store)
	t.Run("returns task by ID 1", func(t *testing.T) {
		request := getTaskByIDRequest(t, "/tasks/1")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		task := storage.Task{}
		err := json.NewDecoder(response.Body).Decode(&task)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "task 1", task.Description)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
	})
	t.Run("returns task by ID 2", func(t *testing.T) {
		request := getTaskByIDRequest(t, "/tasks/2")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		task := storage.Task{}
		err := json.NewDecoder(response.Body).Decode(&task)
		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, "task 2", task.Description)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
	})
	t.Run("returns 404", func(t *testing.T) {
		request := getTaskByIDRequest(t, "/tasks/404")
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		assert.Equal(t, http.StatusNotFound, response.Code)
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
	svr := NewTasksServer(store)
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
	})
}

func createTaskRequest(t *testing.T) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodPost, "/tasks", nil)
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
		svr := NewTasksServer(store)
		request := loadTasksRequest(t)
		response := httptest.NewRecorder()

		svr.ServeHTTP(response, request)

		got := loadTasksResponse(t, response.Body)
		assert.Equal(t, http.StatusOK, response.Code)
		assert.Equal(t, expectedTasks, got)
		assert.Equal(t, "application/json", response.Result().Header.Get("content-type"))
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
