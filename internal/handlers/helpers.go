package handlers

import (
	"encoding/json"
	"net/http"
)

// JSONResponse sends a JSON response with the given status code
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// JSONError sends a JSON error response
func JSONError(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := map[string]string{
		"error": message,
	}
	JSONResponse(w, statusCode, errorResponse)
}

func JSONSuccess(w http.ResponseWriter, data interface{}) {
	JSONResponse(w, http.StatusOK, data)
}

// HandleMethodNotAllowed handles unsupported HTTP methods
func HandleMethodNotAllowed(w http.ResponseWriter, allowedMethods []string) {
	w.Header().Set("Allow", joinMethods(allowedMethods))
	JSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

// Helper function to join methods
func joinMethods(methods []string) string {
	result := ""
	for i, method := range methods {
		if i > 0 {
			result += ", "
		}
		result += method
	}
	return result
}
