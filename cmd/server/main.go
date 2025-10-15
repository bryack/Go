package main

import (
	"fmt"
	"log"
	"myproject/internal/handlers"
	"myproject/storage"
	"myproject/task"
	"myproject/validation"
	"net/http"
	"os"
	"time"
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
func tasksHandler(tm *task.TaskManager, s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		switch r.Method {
		case http.MethodGet:
			response := tm.GetTasks()
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

			id := tm.AddTask(desc)
			response, err := tm.GetTaskByID(id)
			if err != nil {
				handlers.JSONError(w, http.StatusInternalServerError, "Failed to retrieve created task")
			}

			if err := s.SaveTasks(tm.GetTasks()); err != nil {
				log.Printf("Failed to save tasks: %w", err)
			}

			handlers.JSONResponse(w, http.StatusCreated, response)
		default:
			handlers.HandleMethodNotAllowed(w, []string{"GET", "POST"})
			return
		}
	}
}

func taskHandler(tm *task.TaskManager, s storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		response := task.Task{}
		path := r.URL.Path
		id, err := validation.ExtractTaskIDFromPath(path)
		if err != nil {
			handlers.JSONError(w, http.StatusBadRequest, "Invalid ID format")
			return
		}
		switch r.Method {
		case http.MethodGet:
			response, err = tm.GetTaskByID(id)
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

			if taskRequest.Description != nil {
				desc := string(*taskRequest.Description)
				desc, err = validation.ValidateTaskDescription(desc)
				if err != nil {
					handlers.JSONError(w, http.StatusBadRequest, err.Error())
					return
				}

				if err := tm.UpdateTaskDescription(id, desc); err != nil {
					handlers.JSONError(w, http.StatusNotFound, "Task not found")
					return
				}
			}

			if taskRequest.Done != nil {
				if err := tm.UpdateTaskStatus(id, *taskRequest.Done); err != nil {
					handlers.JSONError(w, http.StatusNotFound, "Task not found")
					return
				}
			}

			response, err = tm.GetTaskByID(id)
			if err != nil {
				handlers.JSONError(w, http.StatusNotFound, "Task not found")
				return
			}

			if err := s.SaveTasks(tm.GetTasks()); err != nil {
				log.Printf("Failed to save tasks: %w", err)
			}

			handlers.JSONSuccess(w, response)

		case http.MethodDelete:
			if err := tm.DeleteTask(id); err != nil {
				handlers.JSONError(w, http.StatusNotFound, "Task not found")
				return
			}

			if err := s.SaveTasks(tm.GetTasks()); err != nil {
				log.Printf("Failed to save tasks: %w", err)
			}

			w.WriteHeader(http.StatusNoContent)

		default:
			handlers.HandleMethodNotAllowed(w, []string{"GET", "PUT", "DELETE"})
			return
		}
	}
}

func main() {
	tm := task.NewTaskManager(os.Stdout)

	dbPath := storage.GetDatabasePath()
	s, err := storage.NewDatabaseStorage(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database storage:", err)
	}

	fmt.Println("🚀 Database storage initialized")

	http.HandleFunc("/health", logRequest(healthHandler))
	http.HandleFunc("/tasks/", logRequest(taskHandler(tm, s)))
	http.HandleFunc("/tasks", logRequest(tasksHandler(tm, s)))
	http.HandleFunc("/", logRequest(rootHandler))

	fmt.Println("🚀 HTTP Server starting on http://localhost:8080")
	fmt.Println("Endpoints:")
	fmt.Println("  GET http://localhost:8080/")
	fmt.Println("  GET http://localhost:8080/health")
	fmt.Println("  GET http://localhost:8080/tasks")
	fmt.Println("  POST http://localhost:8080/tasks")
	fmt.Println("  GET http://localhost:8080/tasks/{id}")
	fmt.Println("  PUT http://localhost:8080/tasks/{id}")
	fmt.Println("  DELETE http://localhost:8080/tasks/{id}")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
