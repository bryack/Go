package main

import (
	"fmt"
	"log"
	"myproject/internal/handlers"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
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
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "task-manager-api",
	}
	handlers.JSONSuccess(w, response)
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", rootHandler)

	fmt.Println("🚀 HTTP Server starting on http://localhost:8080")
	fmt.Println("Endpoints:")
	fmt.Println("  GET http://localhost:8080/")
	fmt.Println("  GET http://localhost:8080/health")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
