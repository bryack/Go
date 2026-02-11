package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const jsonContentType = "application/json"

// JSONResponse sends a JSON response with the given status code
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", jsonContentType)
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

func ParseJSONRequest(w http.ResponseWriter, r *http.Request, target interface{}) error {
	if r.Header.Get("Content-Type") != jsonContentType {
		JSONError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return errors.New("Invalid content type")
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return err
	}
	err = json.Unmarshal(body, target)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid JSON format")
		return err
	}

	return nil
}
