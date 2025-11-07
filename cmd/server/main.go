package main

import (
	"errors"
	"fmt"
	"log"
	"myproject/auth"
	"myproject/cmd/server/config"
	"myproject/internal/handlers"
	"myproject/storage"
	"myproject/validation"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"
)

// HealthResponse represents the JSON response structure for health check endpoints.
// Contains service status, timestamp, and service identification.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

// CreateTaskRequest represents the JSON payload for creating new tasks.
// Contains the required task description field.
type CreateTaskRequest struct {
	Description string `json:"description"`
}

// UpdateTaskRequest represents the JSON payload for updating existing tasks.
// All fields are optional pointers to support partial updates.
type UpdateTaskRequest struct {
	Description *string `json:"description,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

// rootHandler serves the API information and available endpoints.
// Returns a JSON response with service description and endpoint list.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Task Manager API",
		"enpoints": []string{
			"Get /health - Health check",
			"Get / - This message",
		},
	}
	handlers.JSONSuccess(w, response)
}

// healthHandler provides service health status information.
// Only accepts GET requests and returns current service status with timestamp.
func healthHandler(w http.ResponseWriter, r *http.Request) {
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

// logRequest is a simple logging middleware
func logRequest(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		handler(w, r)

		duration := time.Since(start)
		log.Printf("%s %s - %v", r.Method, r.URL.Path, duration)
	}
}

// tasksHandler returns a handler function that has access to TaskManager
func tasksHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		userID, err := auth.GetUserIDFromContext(r.Context())
		if err != nil {
			handlers.JSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		switch r.Method {
		case http.MethodGet:
			response, err := s.LoadTasks(userID)
			if err != nil {
				handlers.JSONError(w, http.StatusInternalServerError, "Failed to load tasks")
				return
			}
			handlers.JSONSuccess(w, response)

		case http.MethodPost:
			var taskRequest CreateTaskRequest
			if err := handlers.ParseJSONRequest(w, r, &taskRequest); err != nil {
				return
			}

			desc, err := validation.ValidateTaskDescription(string(taskRequest.Description))
			if err != nil {
				handlers.JSONError(w, http.StatusBadRequest, err.Error())
				return
			}

			newTask := storage.Task{Description: desc, Done: false}
			id, err := s.CreateTask(newTask, userID)
			if err != nil {
				log.Printf("Failed to create task in database: %v", err)
				handlers.JSONError(w, http.StatusInternalServerError, "Failed to create task")
				return
			}
			newTask.ID = id

			handlers.JSONResponse(w, http.StatusCreated, newTask)
		default:
			handlers.HandleMethodNotAllowed(w, []string{"GET", "POST"})
			return
		}
	}
}

// taskHandler returns an HTTP handler for individual task operations by ID.
// Supports GET, PUT, and DELETE methods with automatic storage persistence.
func taskHandler(s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		response := storage.Task{}
		path := r.URL.Path
		id, err := validation.ExtractTaskIDFromPath(path)
		if err != nil {
			handlers.JSONError(w, http.StatusBadRequest, "Invalid ID format")
			return
		}

		userID, err := auth.GetUserIDFromContext(r.Context())
		if err != nil {
			handlers.JSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		switch r.Method {
		case http.MethodGet:
			response, err = s.GetTaskByID(id, userID)
			if err != nil {
				handlers.JSONError(w, http.StatusNotFound, "Task not found")
				return
			}
			handlers.JSONSuccess(w, response)
		case http.MethodPut:
			var taskRequest UpdateTaskRequest
			if err := handlers.ParseJSONRequest(w, r, &taskRequest); err != nil {
				return
			}

			if taskRequest.Description == nil && taskRequest.Done == nil {
				handlers.JSONError(w, http.StatusBadRequest, "At least one field must be provided for update")
				return
			}

			response, err := s.GetTaskByID(id, userID)
			if err != nil {
				handlers.JSONError(w, http.StatusNotFound, "Task not found")
			}

			if taskRequest.Description != nil {
				desc := string(*taskRequest.Description)
				desc, err = validation.ValidateTaskDescription(desc)
				if err != nil {
					handlers.JSONError(w, http.StatusBadRequest, err.Error())
					return
				}
				response.Description = desc
			}

			if taskRequest.Done != nil {
				response.Done = *taskRequest.Done
			}

			if err := s.UpdateTask(response, userID); err != nil {
				handlers.JSONError(w, http.StatusNotFound, "Task not found")
				return
			}

			handlers.JSONSuccess(w, response)

		case http.MethodDelete:
			if err := s.DeleteTask(id, userID); err != nil {
				handlers.JSONError(w, http.StatusNotFound, "Task not found")
				return
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			handlers.HandleMethodNotAllowed(w, []string{"GET", "PUT", "DELETE"})
			return
		}
	}
}

// RegisterHandler creates a new user account with email and password.
// Returns JWT token on success or appropriate error for validation failures.
func RegisterHandler(service auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			handlers.HandleMethodNotAllowed(w, []string{"POST"})
			return
		}

		var registerRequest RegisterRequest
		if err := handlers.ParseJSONRequest(w, r, &registerRequest); err != nil {
			return
		}

		if registerRequest.Email == "" || registerRequest.Password == "" {
			handlers.JSONError(w, http.StatusBadRequest, "Fields must be provided for register")
			return
		}

		token, err := service.Register(registerRequest.Email, registerRequest.Password)
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
}

// LoginHandler authenticates user credentials and returns a JWT token.
// Returns generic error message on failure to prevent user enumeration.
func LoginHandler(service auth.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			handlers.HandleMethodNotAllowed(w, []string{"POST"})
			return
		}

		var loginRequest LoginRequest
		if err := handlers.ParseJSONRequest(w, r, &loginRequest); err != nil {
			return
		}

		if loginRequest.Email == "" || loginRequest.Password == "" {
			handlers.JSONError(w, http.StatusBadRequest, "Fields must be provided for login")
			return
		}

		token, err := service.Login(loginRequest.Email, loginRequest.Password)
		if err != nil {
			handlers.JSONError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		var authResp AuthResponse
		authResp.Email = loginRequest.Email
		authResp.Token = token
		handlers.JSONSuccess(w, authResp)
	}
}

func main() {
	cfg, v, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// Check if --show-config flag was set
	if pflag.Lookup("show-config").Changed && pflag.Lookup("show-config").Value.String() == "true" {
		config.ShowConfig(cfg, v)
		os.Exit(0)
	}

	s, err := storage.NewDatabaseStorage(cfg.DatabaseConfig.Path)
	if err != nil {
		log.Fatal("Failed to initialize database storage:", err)
	}

	jwtService := auth.NewJWTService(cfg.JWTConfig.Secret, cfg.JWTConfig.Expiration)
	authService := auth.NewService(s, jwtService)
	authMiddleware := auth.NewAuthMiddleware(jwtService)

	fmt.Println("🚀 Database storage initialized")
	fmt.Println("🔐 Authentication system initialized")

	http.HandleFunc("/register", logRequest(RegisterHandler(*authService)))
	http.HandleFunc("/login", logRequest(LoginHandler(*authService)))
	http.HandleFunc("/health", logRequest(healthHandler))
	http.HandleFunc("/tasks/", logRequest(authMiddleware.Authenticate(taskHandler(s))))
	http.HandleFunc("/tasks", logRequest(authMiddleware.Authenticate(tasksHandler(s))))
	http.HandleFunc("/", logRequest(rootHandler))

	fmt.Printf("🚀 HTTP Server starting on http://%s:%d\n", cfg.ServerConfig.Host, cfg.ServerConfig.Port)
	fmt.Println("Endpoints:")
	fmt.Println("  GET /")
	fmt.Println("  GET /health")
	fmt.Println("  GET /tasks")
	fmt.Println("  POST /tasks")
	fmt.Println("  GET /tasks/{id}")
	fmt.Println("  PUT /tasks/{id}")
	fmt.Println("  DELETE /tasks/{id}")
	fmt.Println("  POST /register")
	fmt.Println("  POST /login")

	address := fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port)
	log.Fatal(http.ListenAndServe(address, nil))
}
