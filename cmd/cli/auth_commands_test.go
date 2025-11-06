package main

import (
	"bytes"
	"errors"
	"myproject/cmd/cli/client"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockAuthManager is a mock implementation of AuthManager for testing
type MockAuthManager struct {
	loginToken         string
	loginErr           error
	registerToken      string
	registerErr        error
	clearTokenErr      error
	loadTokenResult    string
	loadTokenErr       error
	handleAuthErrToken string
	handleAuthErrErr   error
}

func (m *MockAuthManager) LoadToken() (string, error) {
	return m.loadTokenResult, m.loadTokenErr
}

func (m *MockAuthManager) SaveToken(token string) error {
	return nil
}

func (m *MockAuthManager) ClearToken() error {
	return m.clearTokenErr
}

func (m *MockAuthManager) IsAuthenticated() bool {
	return m.loadTokenResult != ""
}

func (m *MockAuthManager) RequireAuth() (string, error) {
	return m.loadTokenResult, m.loadTokenErr
}

func (m *MockAuthManager) PromptLogin() (string, error) {
	return m.loginToken, m.loginErr
}

func (m *MockAuthManager) PromptRegister() (string, error) {
	return m.registerToken, m.registerErr
}

func (m *MockAuthManager) HandleAuthError() (string, error) {
	return m.handleAuthErrToken, m.handleAuthErrErr
}

// MockTaskClient is a mock implementation of TaskClient for testing
type MockTaskClient struct {
	token            string
	createTaskResult *client.Task
	createTaskErr    error
	getTaskResult    *client.Task
	getTaskErr       error
	updateTaskResult *client.Task
	updateTaskErr    error
	deleteTaskErr    error
	getTasksResult   []client.Task
	getTasksErr      error
}

func (m *MockTaskClient) GetTasks() ([]client.Task, error) {
	return m.getTasksResult, m.getTasksErr
}

func (m *MockTaskClient) GetTask(id int) (*client.Task, error) {
	return m.getTaskResult, m.getTaskErr
}

func (m *MockTaskClient) CreateTask(description string) (*client.Task, error) {
	return m.createTaskResult, m.createTaskErr
}

func (m *MockTaskClient) UpdateTask(id int, description *string, done *bool) (*client.Task, error) {
	return m.updateTaskResult, m.updateTaskErr
}

func (m *MockTaskClient) DeleteTask(id int) error {
	return m.deleteTaskErr
}

func (m *MockTaskClient) Login(email, password string) (string, error) {
	return "", nil
}

func (m *MockTaskClient) Register(email, password string) (string, error) {
	return "", nil
}

func (m *MockTaskClient) SetToken(token string) {
	m.token = token
}

func (m *MockTaskClient) GetServerURL() string {
	return "http://localhost:8080"
}

// TestNewAuthCommands tests that the new authentication commands are recognized as valid
func TestNewAuthCommands(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedCmd Command
		expectedErr error
	}{
		{
			name:        "Login command",
			input:       "login",
			expectedCmd: CommandLogin,
			expectedErr: nil,
		},
		{
			name:        "Register command",
			input:       "register",
			expectedCmd: CommandRegister,
			expectedErr: nil,
		},
		{
			name:        "Logout command",
			input:       "logout",
			expectedCmd: CommandLogout,
			expectedErr: nil,
		},
		{
			name:        "Login command uppercase",
			input:       "LOGIN",
			expectedCmd: CommandLogin,
			expectedErr: nil,
		},
		{
			name:        "Register command mixed case",
			input:       "ReGiStEr",
			expectedCmd: CommandRegister,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := validateCommand(tc.input)

			assert.Equal(t, tc.expectedCmd, cmd)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

// TestCLI_HandleLoginCommand tests the handleLoginCommand method
func TestCLI_HandleLoginCommand(t *testing.T) {
	testCases := []struct {
		name           string
		loginToken     string
		loginErr       error
		expectedOutput string
		expectedErr    bool
	}{
		{
			name:           "Successful login",
			loginToken:     "test-token-123",
			loginErr:       nil,
			expectedOutput: "",
			expectedErr:    false,
		},
		{
			name:           "Failed login",
			loginToken:     "",
			loginErr:       errors.New("invalid credentials"),
			expectedOutput: "",
			expectedErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockAuth := &MockAuthManager{
				loginToken: tc.loginToken,
				loginErr:   tc.loginErr,
			}
			mockClient := &MockTaskClient{}

			cli := NewCLI(
				NewConsoleInputReader(strings.NewReader("")),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				mockAuth,
			)

			err := cli.handleLoginCommand()

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.loginToken, mockClient.token)
			}
		})
	}
}

// TestCLI_HandleRegisterCommand tests the handleRegisterCommand method
func TestCLI_HandleRegisterCommand(t *testing.T) {
	testCases := []struct {
		name           string
		registerToken  string
		registerErr    error
		expectedOutput string
		expectedErr    bool
	}{
		{
			name:           "Successful registration",
			registerToken:  "test-token-456",
			registerErr:    nil,
			expectedOutput: "",
			expectedErr:    false,
		},
		{
			name:           "Failed registration",
			registerToken:  "",
			registerErr:    errors.New("email already exists"),
			expectedOutput: "",
			expectedErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockAuth := &MockAuthManager{
				registerToken: tc.registerToken,
				registerErr:   tc.registerErr,
			}
			mockClient := &MockTaskClient{}

			cli := NewCLI(
				NewConsoleInputReader(strings.NewReader("")),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				mockAuth,
			)

			err := cli.handleRegisterCommand()

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.registerToken, mockClient.token)
			}
		})
	}
}

// TestCLI_HandleLogoutCommand tests the handleLogoutCommand method
func TestCLI_HandleLogoutCommand(t *testing.T) {
	testCases := []struct {
		name           string
		clearTokenErr  error
		expectedOutput string
		expectedErr    bool
	}{
		{
			name:           "Successful logout",
			clearTokenErr:  nil,
			expectedOutput: "✅ Logged out successfully\n👋 Bye!\n",
			expectedErr:    false,
		},
		{
			name:           "Failed logout",
			clearTokenErr:  errors.New("failed to delete token"),
			expectedOutput: "",
			expectedErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockAuth := &MockAuthManager{
				clearTokenErr: tc.clearTokenErr,
			}
			mockClient := &MockTaskClient{}

			cli := NewCLI(
				NewConsoleInputReader(strings.NewReader("")),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				mockAuth,
			)

			err := cli.handleLogoutCommand()

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedOutput, output.String())
			}
		})
	}
}

// TestCLI_HandleAuthError tests the 401 re-authentication handling
func TestCLI_HandleAuthError(t *testing.T) {
	testCases := []struct {
		name               string
		err                error
		handleAuthErrToken string
		handleAuthErrErr   error
		expectedResult     bool
		expectedToken      string
	}{
		{
			name:               "Non-auth error returns false",
			err:                errors.New("some other error"),
			handleAuthErrToken: "",
			handleAuthErrErr:   nil,
			expectedResult:     false,
			expectedToken:      "",
		},
		{
			name:               "Auth error with successful re-authentication",
			err:                &client.AuthError{Message: "token expired"},
			handleAuthErrToken: "new-token-789",
			handleAuthErrErr:   nil,
			expectedResult:     true,
			expectedToken:      "new-token-789",
		},
		{
			name:               "Auth error with failed re-authentication",
			err:                &client.AuthError{Message: "token expired"},
			handleAuthErrToken: "",
			handleAuthErrErr:   errors.New("re-auth failed"),
			expectedResult:     false,
			expectedToken:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockAuth := &MockAuthManager{
				handleAuthErrToken: tc.handleAuthErrToken,
				handleAuthErrErr:   tc.handleAuthErrErr,
			}
			mockClient := &MockTaskClient{}

			cli := NewCLI(
				NewConsoleInputReader(strings.NewReader("")),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				mockAuth,
			)

			result := cli.handleAuthError(tc.err)

			assert.Equal(t, tc.expectedResult, result)
			if tc.expectedToken != "" {
				assert.Equal(t, tc.expectedToken, mockClient.token)
			}
		})
	}
}

// TestClient_IsAuthError tests the IsAuthError helper function
func TestClient_IsAuthError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "AuthError returns true",
			err:      &client.AuthError{Message: "token expired"},
			expected: true,
		},
		{
			name:     "APIError returns false",
			err:      &client.APIError{StatusCode: 404, Message: "not found"},
			expected: false,
		},
		{
			name:     "NetworkError returns false",
			err:      &client.NetworkError{URL: "http://localhost", Err: errors.New("connection refused")},
			expected: false,
		},
		{
			name:     "Generic error returns false",
			err:      errors.New("some error"),
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
			result := client.IsAuthError(tc.err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
