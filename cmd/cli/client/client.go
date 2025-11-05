package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TaskClient defines the interface for interacting with the task management API
type TaskClient interface {
	// Task operations
	GetTasks() ([]Task, error)
	GetTask(id int) (*Task, error)
	CreateTask(description string) (*Task, error)
	UpdateTask(id int, description *string, done *bool) (*Task, error)
	DeleteTask(id int) error

	// Authentication
	Login(email, password string) (string, error)
	Register(email, password string) (string, error)

	// Configuration
	SetToken(token string)
	GetServerURL() string
}

// HTTPClient implements TaskClient using HTTP requests
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// Task represents a task in the system
type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

// AuthRequest represents login/register request payload
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

// CreateTaskRequest represents task creation request
type CreateTaskRequest struct {
	Description string `json:"description"`
}

// UpdateTaskRequest represents task update request
type UpdateTaskRequest struct {
	Description *string `json:"description,omitempty"`
	Done        *bool   `json:"done,omitempty"`
}

// ErrorResponse represents an error response from the server
type ErrorResponse struct {
	Error string `json:"error"`
}

// NetworkError represents a network connectivity error
type NetworkError struct {
	URL string
	Err error
}

func (e *NetworkError) Error() string {
	return fmt.Sprintf("cannot connect to server at %s: %v", e.URL, e.Err)
}

// APIError represents an HTTP error response from the API
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return e.Message
}

// AuthError represents an authentication error (401 Unauthorized)
// This error type signals that the stored token is invalid and re-authentication is required
type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

// IsAuthError checks if an error is an authentication error
func IsAuthError(err error) bool {
	_, ok := err.(*AuthError)
	return ok
}

// NewHTTPClient creates a new HTTP client with the specified base URL
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the authentication token for subsequent requests
func (c *HTTPClient) SetToken(token string) {
	c.token = token
}

// GetServerURL returns the configured server URL
func (c *HTTPClient) GetServerURL() string {
	return c.baseURL
}

// doRequest performs an HTTP request with JSON encoding/decoding
func (c *HTTPClient) doRequest(method, path string, body, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &NetworkError{
			URL: c.baseURL,
			Err: err,
		}
	}
	defer resp.Body.Close()

	// Handle error responses
	if resp.StatusCode >= 400 {
		return c.handleErrorResponse(resp)
	}

	// Decode successful response
	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// handleErrorResponse parses and returns appropriate errors for HTTP error responses
func (c *HTTPClient) handleErrorResponse(resp *http.Response) error {
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		// If we can't decode the error response, use status text
		errResp.Error = resp.Status
	}

	// Handle 401 Unauthorized - return AuthError to trigger re-authentication
	if resp.StatusCode == http.StatusUnauthorized {
		return &AuthError{
			Message: "Authentication required: token is invalid or expired",
		}
	}

	// Handle specific status codes
	switch {
	case resp.StatusCode >= 500:
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Server error (%d), please try again later", resp.StatusCode),
		}
	case resp.StatusCode >= 400:
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    errResp.Error,
		}
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    errResp.Error,
	}
}

// Login authenticates a user and returns a JWT token
func (c *HTTPClient) Login(email, password string) (string, error) {
	req := AuthRequest{
		Email:    email,
		Password: password,
	}

	var resp AuthResponse
	if err := c.doRequest(http.MethodPost, "/login", req, &resp); err != nil {
		return "", err
	}

	return resp.Token, nil
}

// Register creates a new user account and returns a JWT token
func (c *HTTPClient) Register(email, password string) (string, error) {
	req := AuthRequest{
		Email:    email,
		Password: password,
	}

	var resp AuthResponse
	if err := c.doRequest(http.MethodPost, "/register", req, &resp); err != nil {
		return "", err
	}

	return resp.Token, nil
}

// GetTasks retrieves all tasks for the authenticated user
func (c *HTTPClient) GetTasks() ([]Task, error) {
	var tasks []Task
	if err := c.doRequest(http.MethodGet, "/tasks", nil, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTask retrieves a specific task by ID
func (c *HTTPClient) GetTask(id int) (*Task, error) {
	var task Task
	path := fmt.Sprintf("/tasks/%d", id)
	if err := c.doRequest(http.MethodGet, path, nil, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateTask creates a new task with the given description
func (c *HTTPClient) CreateTask(description string) (*Task, error) {
	req := CreateTaskRequest{
		Description: description,
	}

	var task Task
	if err := c.doRequest(http.MethodPost, "/tasks", req, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask updates a task's description and/or done status
func (c *HTTPClient) UpdateTask(id int, description *string, done *bool) (*Task, error) {
	req := UpdateTaskRequest{
		Description: description,
		Done:        done,
	}

	var task Task
	path := fmt.Sprintf("/tasks/%d", id)
	if err := c.doRequest(http.MethodPut, path, req, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

// DeleteTask deletes a task by ID
func (c *HTTPClient) DeleteTask(id int) error {
	path := fmt.Sprintf("/tasks/%d", id)
	return c.doRequest(http.MethodDelete, path, nil, nil)
}
