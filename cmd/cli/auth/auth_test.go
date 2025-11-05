package auth

import (
	"bytes"
	"errors"
	"myproject/cmd/cli/client"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockInputReader is a mock implementation of InputReader for testing
type MockInputReader struct {
	inputs []string
	index  int
}

func NewMockInputReader(inputs ...string) *MockInputReader {
	return &MockInputReader{
		inputs: inputs,
		index:  0,
	}
}

func (m *MockInputReader) ReadInput(maxSize int) (string, error) {
	if m.index >= len(m.inputs) {
		return "", errors.New("no more inputs")
	}
	input := m.inputs[m.index]
	m.index++
	return input, nil
}

// MockTaskClient is a mock implementation of TaskClient for testing
type MockTaskClient struct {
	loginEmail    string
	loginPassword string
	loginToken    string
	loginErr      error

	registerEmail    string
	registerPassword string
	registerToken    string
	registerErr      error
}

func (m *MockTaskClient) Login(email, password string) (string, error) {
	m.loginEmail = email
	m.loginPassword = password
	return m.loginToken, m.loginErr
}

func (m *MockTaskClient) Register(email, password string) (string, error) {
	m.registerEmail = email
	m.registerPassword = password
	return m.registerToken, m.registerErr
}

func (m *MockTaskClient) GetTasks() ([]client.Task, error)                    { return nil, nil }
func (m *MockTaskClient) GetTask(id int) (*client.Task, error)                { return nil, nil }
func (m *MockTaskClient) CreateTask(description string) (*client.Task, error) { return nil, nil }
func (m *MockTaskClient) UpdateTask(id int, description *string, done *bool) (*client.Task, error) {
	return nil, nil
}
func (m *MockTaskClient) DeleteTask(id int) error { return nil }
func (m *MockTaskClient) SetToken(token string)   {}
func (m *MockTaskClient) GetServerURL() string    { return "http://localhost:8080" }

// TestFileAuthManager_HandleAuthError tests the HandleAuthError method
func TestFileAuthManager_HandleAuthError(t *testing.T) {
	testCases := []struct {
		name          string
		inputs        []string
		loginToken    string
		loginErr      error
		registerToken string
		registerErr   error
		expectedToken string
		expectedErr   bool
	}{
		{
			name:          "Successful re-authentication via login",
			inputs:        []string{"1", "test@example.com", "password123"},
			loginToken:    "new-token-123",
			loginErr:      nil,
			expectedToken: "new-token-123",
			expectedErr:   false,
		},
		{
			name:          "Successful re-authentication via register",
			inputs:        []string{"2", "new@example.com", "password123", "password123"},
			registerToken: "new-token-456",
			registerErr:   nil,
			expectedToken: "new-token-456",
			expectedErr:   false,
		},
		{
			name:        "User chooses to exit",
			inputs:      []string{"3"},
			expectedErr: true,
		},
		{
			name:        "Invalid choice",
			inputs:      []string{"invalid"},
			expectedErr: true,
		},
		{
			name:        "Failed login",
			inputs:      []string{"1", "test@example.com", "wrongpassword"},
			loginErr:    errors.New("invalid credentials"),
			expectedErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockInput := NewMockInputReader(tc.inputs...)
			mockClient := &MockTaskClient{
				loginToken:    tc.loginToken,
				loginErr:      tc.loginErr,
				registerToken: tc.registerToken,
				registerErr:   tc.registerErr,
			}

			// Create a temporary auth manager (token path doesn't matter for this test)
			authMgr := &FileAuthManager{
				tokenPath: "/tmp/test-token",
				client:    mockClient,
				input:     mockInput,
				output:    output,
			}

			token, err := authMgr.HandleAuthError()

			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedToken, token)
			}

			// Verify output contains expected messages
			outputStr := output.String()
			assert.Contains(t, outputStr, "session has expired")
			assert.Contains(t, outputStr, "Choose an option")
		})
	}
}

// TestFileAuthManager_HandleAuthError_ClearsToken tests that HandleAuthError clears the token
func TestFileAuthManager_HandleAuthError_ClearsToken(t *testing.T) {
	output := &bytes.Buffer{}
	mockInput := NewMockInputReader("1", "test@example.com", "password123")
	mockClient := &MockTaskClient{
		loginToken: "new-token",
		loginErr:   nil,
	}

	// Create a temporary token file
	tmpDir := t.TempDir()
	tokenPath := tmpDir + "/token"

	// Write an old token
	authMgr := &FileAuthManager{
		tokenPath: tokenPath,
		client:    mockClient,
		input:     mockInput,
		output:    output,
	}

	// Save an old token
	err := authMgr.SaveToken("old-token")
	assert.NoError(t, err)

	// Verify token exists
	oldToken, err := authMgr.LoadToken()
	assert.NoError(t, err)
	assert.Equal(t, "old-token", oldToken)

	// Call HandleAuthError
	newToken, err := authMgr.HandleAuthError()
	assert.NoError(t, err)
	assert.Equal(t, "new-token", newToken)

	// Verify the new token was saved
	savedToken, err := authMgr.LoadToken()
	assert.NoError(t, err)
	assert.Equal(t, "new-token", savedToken)
}
