package main

import (
	"fmt"
	"log"
	"myproject/internal/handlers"
	"myproject/storage"
	"myproject/task"
	"net/http"
	"os"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
}

type CreateTaskRequest struct {
	Description string `json:"description"`
}

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
func tasksHandler(tm *task.TaskManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			response := tm.GetTasks()
			handlers.JSONSuccess(w, response)
		case http.MethodPost:
			response := tm.GetTasks()
			handlers.JSONSuccess(w, response)
		default:
			handlers.HandleMethodNotAllowed(w, []string{"GET", "POST"})
			return
		}
	}
}

func main() {
	tm := task.NewTaskManager(os.Stdout)
	var s storage.Storage = storage.JsonStorage{}

	loadedTask, err := s.LoadTasks()
	if err == nil {
		tm.SetTasks(loadedTask)
		fmt.Println("Loaded existing tasks")
	} else {
		fmt.Println("Starting with empty task list")
	}

	http.HandleFunc("/health", logRequest(healthHandler))
	http.HandleFunc("/", logRequest(rootHandler))
	http.HandleFunc("/tasks", logRequest(tasksHandler(tm)))

	fmt.Println("🚀 HTTP Server starting on http://localhost:8080")
	fmt.Println("Endpoints:")
	fmt.Println("  GET http://localhost:8080/")
	fmt.Println("  GET http://localhost:8080/health")
	fmt.Println("  GET http://localhost:8080/tasks")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
