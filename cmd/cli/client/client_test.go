package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHTTPClient_HandleErrorResponse_401 tests that 401 responses return AuthError
func TestHTTPClient_HandleErrorResponse_401(t *testing.T) {
	// Create a test server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Unauthorized"})
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	client.SetToken("invalid-token")

	// Try to get tasks, should return AuthError
	_, err := client.GetTasks()

	assert.Error(t, err)
	assert.True(t, IsAuthError(err), "Expected AuthError for 401 response")

	authErr, ok := err.(*AuthError)
	assert.True(t, ok, "Error should be of type *AuthError")
	assert.Contains(t, authErr.Message, "Authentication required")
}

// TestHTTPClient_HandleErrorResponse_404 tests that 404 responses return APIError
func TestHTTPClient_HandleErrorResponse_404(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Task not found"})
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	client.SetToken("valid-token")

	// Try to get a non-existent task
	_, err := client.GetTask(999)

	assert.Error(t, err)
	assert.False(t, IsAuthError(err), "404 should not return AuthError")

	apiErr, ok := err.(*APIError)
	assert.True(t, ok, "Error should be of type *APIError")
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Contains(t, apiErr.Message, "Task not found")
}

// TestHTTPClient_HandleErrorResponse_500 tests that 500 responses return APIError
func TestHTTPClient_HandleErrorResponse_500(t *testing.T) {
	// Create a test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Internal server error"})
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL)
	client.SetToken("valid-token")

	// Try to get tasks
	_, err := client.GetTasks()

	assert.Error(t, err)
	assert.False(t, IsAuthError(err), "500 should not return AuthError")

	apiErr, ok := err.(*APIError)
	assert.True(t, ok, "Error should be of type *APIError")
	assert.Equal(t, http.StatusInternalServerError, apiErr.StatusCode)
	assert.Contains(t, apiErr.Message, "Server error")
}

// TestIsAuthError tests the IsAuthError helper function
func TestIsAuthError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "AuthError returns true",
			err:      &AuthError{Message: "token expired"},
			expected: true,
		},
		{
			name:     "APIError returns false",
			err:      &APIError{StatusCode: 404, Message: "not found"},
			expected: false,
		},
		{
			name:     "NetworkError returns false",
			err:      &NetworkError{URL: "http://localhost", Err: http.ErrServerClosed},
			expected: false,
		},
		{
			name:     "Nil error returns false",
			err:      nil,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsAuthError(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
