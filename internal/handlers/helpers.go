package handlers

import (
	"encoding/json"
	"net/http"
)

func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func JSONError(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := map[string]string{
		"error": message,
	}
	JSONResponse(w, statusCode, errorResponse)
}

func JSONSuccess(w http.ResponseWriter, data interface{}) {
	JSONResponse(w, http.StatusOK, data)
}
