package main

import (
	"errors"
	"log/slog"
	"myproject/auth"
	"myproject/internal/handlers"
	"myproject/logger"
	"myproject/storage"
	"myproject/validation"
	"net/http"
	"time"
)

type AuthService interface {
	Register(email, password string) (token string, err error)
	Login(email, password string) (token string, err error)
}

type Authenticator interface {
	Authenticate(handler http.HandlerFunc) http.HandlerFunc
}

type TasksServer struct {
	store          storage.Storage
	authService    AuthService
	authMiddleware Authenticator
	logger         *slog.Logger
	http.Handler
}

func NewTasksServer(store storage.Storage, authService AuthService, authMiddleware Authenticator, logger *slog.Logger) *TasksServer {
	ts := &TasksServer{}
	ts.store = store
	ts.authService = authService
	ts.authMiddleware = authMiddleware
	ts.logger = logger
	router := http.NewServeMux()

	router.Handle("/", http.HandlerFunc(ts.rootHandler))
	router.Handle("/health", http.HandlerFunc(ts.healthHandler))
	router.Handle("GET /tasks", ts.authMiddleware.Authenticate(ts.tasksHandler))
	router.Handle("POST /tasks", ts.authMiddleware.Authenticate(ts.tasksHandler))
	router.Handle("GET /tasks/{id}", ts.authMiddleware.Authenticate(ts.taskHandler))
	router.Handle("PUT /tasks/{id}", ts.authMiddleware.Authenticate(ts.taskHandler))
	router.Handle("DELETE /tasks/{id}", ts.authMiddleware.Authenticate(ts.taskHandler))
	router.Handle("POST /register", http.HandlerFunc(ts.registerHandler))
	router.Handle("POST /login", http.HandlerFunc(ts.loginHandler))

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
		ts.processCreateTask(w, r, userID)
	default:
		handlers.HandleMethodNotAllowed(w, []string{"GET", "POST"})
		return
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

func (ts *TasksServer) processCreateTask(w http.ResponseWriter, r *http.Request, userID int) {
	var taskRequest CreateTaskRequest
	if err := handlers.ParseJSONRequest(w, r, &taskRequest); err != nil {
		return
	}

	desc, err := validation.ValidateTaskDescription(string(taskRequest.Description))
	if err != nil {
		ts.logger.Warn("Failed to validate description",
			slog.String(logger.FieldOperation, "tasks_handler"),
			slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
			slog.Int(logger.FieldUserID, userID),
			slog.String("task_description", string(taskRequest.Description)),
			slog.String(logger.FieldError, err.Error()),
		)
		handlers.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	newTask := storage.Task{Description: desc, Done: false}
	id, err := ts.store.CreateTask(newTask, userID)
	if err != nil {
		ts.logger.Error("Failed to create task in database",
			slog.String(logger.FieldOperation, "tasks_handler"),
			slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
			slog.Int(logger.FieldUserID, userID),
			slog.String(logger.FieldError, err.Error()),
		)
		handlers.JSONError(w, http.StatusInternalServerError, "Failed to create task")
		return
	}
	newTask.ID = id

	handlers.JSONResponse(w, http.StatusCreated, newTask)
}

func (ts *TasksServer) taskHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		handlers.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	idStr := r.PathValue("id")
	id, err := validation.ValidateTaskID(idStr)
	if err != nil {
		handlers.JSONError(w, http.StatusBadRequest, "Invalid task ID")
		return
	}
	switch r.Method {
	case http.MethodGet:
		ts.processGetTaskByID(w, r, id, userID)
	case http.MethodPut:
		ts.processUpdateTask(w, r, id, userID)
	case http.MethodDelete:
		ts.processDeleteTask(w, r, id, userID)
	}
}

func (ts *TasksServer) processGetTaskByID(w http.ResponseWriter, r *http.Request, taskID int, userID int) {

	response, err := ts.store.GetTaskByID(taskID, userID)
	if err != nil {
		ts.logger.Warn("Failed to get task by ID from database",
			slog.String(logger.FieldOperation, "task_handler"),
			slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
			slog.Int(logger.FieldUserID, userID),
			slog.Int(logger.FieldTaskID, taskID),
			slog.String(logger.FieldError, err.Error()),
		)
		handlers.JSONError(w, http.StatusNotFound, "Task not found")
		return
	}
	handlers.JSONSuccess(w, response)
}

func (ts *TasksServer) processUpdateTask(w http.ResponseWriter, r *http.Request, taskID int, userID int) {
	var taskRequest UpdateTaskRequest
	if err := handlers.ParseJSONRequest(w, r, &taskRequest); err != nil {
		return
	}

	if taskRequest.Description == nil && taskRequest.Done == nil {
		handlers.JSONError(w, http.StatusBadRequest, "At least one field must be provided for update")
		return
	}

	response, err := ts.store.GetTaskByID(taskID, userID)
	if err != nil {
		ts.logger.Warn("Failed to get task by ID from database to update",
			slog.String(logger.FieldOperation, "task_handler"),
			slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
			slog.Int(logger.FieldUserID, userID),
			slog.Int(logger.FieldTaskID, taskID),
			slog.String(logger.FieldError, err.Error()),
		)
		handlers.JSONError(w, http.StatusNotFound, "Task not found")
		return
	}

	if taskRequest.Description != nil {
		desc := string(*taskRequest.Description)
		desc, err = validation.ValidateTaskDescription(desc)
		if err != nil {
			ts.logger.Warn("Failed to validate description",
				slog.String(logger.FieldOperation, "task_handler"),
				slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
				slog.Int(logger.FieldUserID, userID),
				slog.String("task_description", string(*taskRequest.Description)),
				slog.String(logger.FieldError, err.Error()),
			)
			handlers.JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Description = desc
	}

	if taskRequest.Done != nil {
		response.Done = *taskRequest.Done
	}

	if err := ts.store.UpdateTask(response, userID); err != nil {
		handlers.JSONError(w, http.StatusNotFound, "Task not found")
		return
	}

	handlers.JSONSuccess(w, response)
}

func (ts *TasksServer) processDeleteTask(w http.ResponseWriter, r *http.Request, taskID int, userID int) {
	if err := ts.store.DeleteTask(taskID, userID); err != nil {
		ts.logger.Warn("Failed to get task by ID from database to delete",
			slog.String(logger.FieldOperation, "task_handler"),
			slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
			slog.Int(logger.FieldUserID, userID),
			slog.Int(logger.FieldTaskID, taskID),
			slog.String(logger.FieldError, err.Error()),
		)
		handlers.JSONError(w, http.StatusNotFound, "Task not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)

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

func (ts *TasksServer) registerHandler(w http.ResponseWriter, r *http.Request) {
	var registerRequest RegisterRequest
	if err := handlers.ParseJSONRequest(w, r, &registerRequest); err != nil {
		return
	}
	if registerRequest.Email == "" || registerRequest.Password == "" {
		handlers.JSONError(w, http.StatusBadRequest, "Fields must be provided for register")
		return
	}

	token, err := ts.authService.Register(registerRequest.Email, registerRequest.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail), errors.Is(err, auth.ErrPasswordTooLong), errors.Is(err, auth.ErrPasswordTooShort):
			handlers.JSONError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, auth.ErrEmailAlreadyExists):
			handlers.JSONError(w, http.StatusConflict, err.Error())
		default:
			handlers.JSONError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	var authResp AuthResponse
	authResp.Email = registerRequest.Email
	authResp.Token = token

	handlers.JSONResponse(w, http.StatusCreated, authResp)
}

func (ts *TasksServer) loginHandler(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest
	if err := handlers.ParseJSONRequest(w, r, &loginRequest); err != nil {
		return
	}

	if loginRequest.Email == "" || loginRequest.Password == "" {
		handlers.JSONError(w, http.StatusBadRequest, "Fields must be provided for login")
		return
	}

	token, err := ts.authService.Login(loginRequest.Email, loginRequest.Password)
	if err != nil {
		handlers.JSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	var authResp AuthResponse
	authResp.Email = loginRequest.Email
	authResp.Token = token
	handlers.JSONSuccess(w, authResp)
}
