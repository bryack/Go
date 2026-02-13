package main

import (
	"log/slog"
	"myproject/auth"
	"myproject/internal/handlers"
	"myproject/storage"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Authenticator interface {
	Authenticate(handler http.HandlerFunc) http.HandlerFunc
}

type TasksServer struct {
	store          storage.Storage
	authMiddleware Authenticator
	logger         *slog.Logger
	http.Handler
}

func NewTasksServer(store storage.Storage, authMiddleware Authenticator, logger *slog.Logger) *TasksServer {
	ts := &TasksServer{}
	ts.store = store
	ts.authMiddleware = authMiddleware
	ts.logger = logger
	router := http.NewServeMux()

	router.Handle("/", http.HandlerFunc(ts.rootHandler))
	router.Handle("/health", http.HandlerFunc(ts.healthHandler))
	router.Handle("/tasks", ts.authMiddleware.Authenticate(ts.tasksHandler))
	router.Handle("/tasks/", ts.authMiddleware.Authenticate(http.HandlerFunc(ts.taskHandler)))

	ts.Handler = router
	return ts
}

func (ts *TasksServer) rootHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Task Manager API",
		"enpoints": []string{
			"Get /health - Health check",
			"GET /tasks - Get tasks",
			"POST /tasks - Add task",
			"GET /tasks/{id} - Get task",
			"PUT /tasks/{id} - Update task",
			"DELETE /tasks/{id} - Delete task",
			"POST /register - Register user",
			"POST /login - Login user",
			"Get / - This message",
		},
	}
	handlers.JSONSuccess(w, response)
}

func (ts *TasksServer) tasksHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		handlers.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	switch r.Method {
	case http.MethodGet:
		ts.processLoadTasks(w, userID)
	case http.MethodPost:
		ts.processCreateTask(w)
	}
}

func (ts *TasksServer) processLoadTasks(w http.ResponseWriter, userID int) {
	response, err := ts.store.LoadTasks(userID)
	if err != nil {
		handlers.JSONError(w, http.StatusInternalServerError, "Failed to load tasks")
		return
	}
	handlers.JSONSuccess(w, response)
}

func (ts *TasksServer) taskHandler(w http.ResponseWriter, r *http.Request) {
	ts.processGetTaskByID(w, r)
}

func (ts *TasksServer) processCreateTask(w http.ResponseWriter) {
	newTask := storage.Task{ID: 1}
	id, err := ts.store.CreateTask(newTask, 1)
	if err != nil {
		handlers.JSONError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}
	newTask.ID = id

	handlers.JSONResponse(w, http.StatusCreated, newTask)
}

func (ts *TasksServer) processGetTaskByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/tasks/"))
	if err != nil {
		handlers.JSONError(w, http.StatusBadRequest, "Task not found")
		return
	}
	response, err := ts.store.GetTaskByID(id, 1)
	if err != nil {
		handlers.JSONError(w, http.StatusNotFound, "Task not found")
		return
	}
	if response.Description == "" {
		handlers.JSONError(w, http.StatusNotFound, "Task not found")
		return
	}
	handlers.JSONSuccess(w, response)
}

// healthHandler provides service health status information.
func (ts *TasksServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handlers.HandleMethodNotAllowed(w, []string{"GET"})
		return
	}
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "task-manager-api",
	}
	handlers.JSONSuccess(w, response)
}
