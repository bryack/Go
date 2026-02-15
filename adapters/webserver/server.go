package webserver

import (
	"errors"
	"log/slog"
	"myproject/application"
	"myproject/auth"
	infraErrors "myproject/infrastructure/errors"
	"myproject/internal/domain"
	"myproject/internal/handlers"
	"myproject/logger"
	"myproject/validation"
	"net/http"
	"time"
)

// HealthResponse represents the JSON response for health check endpoints.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// CreateTaskRequest represents the JSON payload for creating new tasks.
type CreateTaskRequest struct {
	Description string `json:"description"`
}

// UpdateTaskRequest represents the JSON payload for updating tasks with optional fields.
type UpdateTaskRequest struct {
	Description *string `json:"description,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}

// RegisterRequest represents the JSON payload for user registration.
// Contains email and password fields for creating a new account.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the JSON payload for user authentication.
// Contains email and password credentials for login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents the JSON response for successful authentication.
// Contains the JWT token and associated email address.
type AuthResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

type AuthService interface {
	Register(email, password string) (token string, err error)
	Login(email, password string) (token string, err error)
}

type Authenticator interface {
	Authenticate(handler http.HandlerFunc) http.HandlerFunc
}

type TasksServer struct {
	store          domain.Storage
	service        *application.Service
	authService    AuthService
	authMiddleware Authenticator
	logger         *slog.Logger
	http.Handler
}

func NewTasksServer(store domain.Storage, authService AuthService, authMiddleware Authenticator, l *slog.Logger) *TasksServer {
	ts := &TasksServer{}
	ts.store = store
	ts.authService = authService
	ts.authMiddleware = authMiddleware
	ts.service = application.NewService(store)
	ts.logger = l
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

	ts.Handler = logger.LoggingMiddleware(l)(router)
	return ts
}

// rootHandler serves the API information and available endpoints.
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

// tasksHandler handles GET (list all tasks) and POST (create task) requests.
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

	task, err := ts.service.CreateTask(taskRequest.Description, userID)
	if err != nil {
		ts.handleCreateTaskError(w, r, userID, err)
		return
	}

	handlers.JSONResponse(w, http.StatusCreated, task)
}

func (ts *TasksServer) handleCreateTaskError(w http.ResponseWriter, r *http.Request, userID int, err error) {
	if errors.Is(err, infraErrors.ErrDescriptionRequired) || errors.Is(err, infraErrors.ErrDescriptionTooLong) || errors.Is(err, infraErrors.ErrEmptyFieldsToUpdate) {
		ts.logTaskError(r, slog.LevelWarn, "Failed to validate description", userID, 0, err)
		handlers.JSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	ts.logTaskError(r, slog.LevelError, "Failed to create task in database", userID, 0, err)
	handlers.JSONError(w, http.StatusInternalServerError, "Failed to create task")
}

// taskHandler handles GET, PUT, and DELETE operations for individual tasks by ID.
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
		ts.logTaskError(r, slog.LevelWarn, "Failed to get task by ID from database", userID, taskID, err)
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

	task, err := ts.service.UpdateTask(taskID, userID, taskRequest.Description, taskRequest.Done)
	if err != nil {
		ts.handleUpdateTaskError(w, r, userID, taskID, err)
		return
	}

	handlers.JSONSuccess(w, task)
}

func (ts *TasksServer) handleUpdateTaskError(w http.ResponseWriter, r *http.Request, userID, taskID int, err error) {
	switch {
	case errors.Is(err, infraErrors.ErrDescriptionRequired),
		errors.Is(err, infraErrors.ErrDescriptionTooLong),
		errors.Is(err, infraErrors.ErrEmptyFieldsToUpdate):
		ts.logTaskError(r, slog.LevelWarn, "Failed to validate description", userID, taskID, err)
		handlers.JSONError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, infraErrors.ErrTaskNotFound):
		ts.logTaskError(r, slog.LevelWarn, "Failed to get task by ID from database to update", userID, taskID, err)
		handlers.JSONError(w, http.StatusNotFound, "Task not found")
	default:
		ts.logTaskError(r, slog.LevelError, "Failed to update task in database", userID, taskID, err)
		handlers.JSONError(w, http.StatusInternalServerError, "Failed to update task")
	}
}

func (ts *TasksServer) processDeleteTask(w http.ResponseWriter, r *http.Request, taskID, userID int) {
	if err := ts.store.DeleteTask(taskID, userID); err != nil {
		ts.logTaskError(r, slog.LevelWarn, "Failed to delete task from database", userID, taskID, err)
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

// RegisterHandler creates a new user account and returns a JWT token.
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
			ts.logger.Error("Registration failed",
				slog.String(logger.FieldOperation, "register_handler"),
				slog.String(logger.FieldError, err.Error()),
			)
			handlers.JSONError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	var authResp AuthResponse
	authResp.Email = registerRequest.Email
	authResp.Token = token

	handlers.JSONResponse(w, http.StatusCreated, authResp)
}

// LoginHandler authenticates user credentials and returns a JWT token.
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
		ts.logger.Warn("Login failed",
			slog.String(logger.FieldOperation, "login_handler"),
			slog.String("email", loginRequest.Email),
			slog.String(logger.FieldError, err.Error()),
		)
		handlers.JSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	var authResp AuthResponse
	authResp.Email = loginRequest.Email
	authResp.Token = token
	handlers.JSONSuccess(w, authResp)
}

func (ts *TasksServer) logTaskError(r *http.Request, level slog.Level, msg string, userID, taskID int, err error) {
	ts.logger.Log(r.Context(), level, msg,
		slog.String(logger.FieldOperation, "task_handler"),
		slog.String(logger.FieldRequestID, logger.GetRequestID(r.Context())),
		slog.Int(logger.FieldUserID, userID),
		slog.Int(logger.FieldTaskID, taskID),
		slog.String(logger.FieldError, err.Error()),
	)
}
