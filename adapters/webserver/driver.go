package webserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"myproject/internal/domain"
	"net/http"
)

type Driver struct {
	BaseURL string
	Client  *http.Client
}

type APIError struct {
	Message string `json:"error"`
}

func (a APIError) Error() string {
	return a.Message
}

func (d Driver) Register(email, password string) error {
	body := RegisterRequest{
		Email:    email,
		Password: password,
	}

	response, err := post(d, body, "/register", "")
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return checkStatus(response)
}

func (d Driver) Login(email, password string) (token string, err error) {
	body := LoginRequest{
		Email:    email,
		Password: password,
	}

	response, err := post(d, body, "/login", "")
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if err = checkStatus(response); err != nil {
		return "", err
	}

	result := AuthResponse{}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode login response: %w", err)
	}

	return result.Token, nil
}

func (d Driver) CreateTask(token, description string) (taskID int, err error) {
	body := CreateTaskRequest{
		Description: description,
	}

	response, err := post(d, body, "/tasks", token)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()

	if err = checkStatus(response); err != nil {
		return 0, err
	}

	result := domain.Task{}
	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode create task response: %w", err)
	}

	return result.ID, nil
}

func post[T any](d Driver, body T, url, token string) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, d.BaseURL+url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create POST request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	return d.Client.Do(request)
}

func checkStatus(response *http.Response) error {
	if response.StatusCode < 400 {
		return nil
	}

	var apiErr APIError
	if err := json.NewDecoder(response.Body).Decode(&apiErr); err != nil {
		return fmt.Errorf("unexpected status: %d", response.StatusCode)
	}

	return fmt.Errorf("unexpected status %d: %w", response.StatusCode, apiErr)
}
