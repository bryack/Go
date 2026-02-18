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
}

func (d Driver) Register(email, password string) error {
	body := RegisterRequest{
		Email:    email,
		Password: password,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal json body for email %q: %w", email, err)
	}

	response, err := http.Post(d.BaseURL+"/register", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to post register request for email %q: %w", email, err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status: %d", response.StatusCode)
	}

	return fmt.Errorf("not implemented")
}

func (d Driver) Login(email, password string) (token string, err error) {
	body := LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json body for email %q: %w", email, err)
	}

	response, err := http.Post(d.BaseURL+"/login", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to post login request for email %q: %w", email, err)
	}
	defer response.Body.Close()

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
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal json body with description %q: %w", description, err)
	}

	response, err := http.Post(d.BaseURL+"/tasks", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("failed to post create task request: %w", err)
	}

	result := domain.Task{}
	if err = json.NewDecoder(response.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode create task response: %w", err)
	}

	return result.ID, nil
}
