package main

import (
	"bytes"
	"errors"
	"io"
	"myproject/cmd/cli/auth"
	"myproject/cmd/cli/client"
	"myproject/validation"
	"strings"
	"testing"
)

// TestFormatTask tests the formatTask function
func TestFormatTask(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name     string
		task     client.Task
		expected string
	}{
		{
			name:     "Incomplete task",
			task:     client.Task{ID: 1, Description: "Test task", Done: false},
			expected: "[ ] 1: Test task",
		},
		{
			name:     "Complete task",
			task:     client.Task{ID: 2, Description: "Done task", Done: true},
			expected: "[✓] 2: Done task",
		},
		{
			name:     "Task with ID 0",
			task:     client.Task{ID: 0, Description: "Zero ID task", Done: false},
			expected: "[ ] 0: Zero ID task",
		},
		{
			name:     "Task with empty description",
			task:     client.Task{ID: 5, Description: "", Done: false},
			expected: "[ ] 5: ",
		},
		{
			name:     "Task with long description",
			task:     client.Task{ID: 10, Description: "This is a very long task description that should not be truncated", Done: true},
			expected: "[✓] 10: This is a very long task description that should not be truncated",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			result := formatTask(tc.task)

			// ====Assert====
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// TestNewConsoleInputReader tests the NewConsoleInputReader constructor
func TestNewConsoleInputReader(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name   string
		reader io.Reader
	}{
		{
			name:   "Valid io.Reader with strings.Reader",
			reader: strings.NewReader("test input\n"),
		},
		{
			name:   "Valid io.Reader with empty content",
			reader: strings.NewReader(""),
		},
		{
			name:   "Valid io.Reader with bytes.Buffer",
			reader: bytes.NewBufferString("buffer input\n"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			result := NewConsoleInputReader(tc.reader)

			// ====Assert====
			if result == nil {
				t.Error("Expected non-nil ConsoleInputReader")
			}
			if result.reader == nil {
				t.Error("Expected non-nil buffered reader")
			}
		})
	}
}

// MockInputReader implements InputReader for testing
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
		return "", io.EOF
	}
	input := m.inputs[m.index]
	m.index++

	// Simulate size validation
	if len(input) > maxSize {
		return "", ErrMaxSizeExceeded
	}
	if input == "" {
		return "", ErrEmptyInput
	}

	return input, nil
}

// TestNewCLI tests the NewCLI constructor
func TestNewCLI(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name        string
		input       InputReader
		output      io.Writer
		cfg         *Config
		client      client.TaskClient
		authManager auth.AuthManager
	}{
		{
			name:        "Valid dependencies",
			input:       NewMockInputReader(),
			output:      &bytes.Buffer{},
			cfg:         &Config{ServerURL: "http://localhost:8080"},
			client:      &MockTaskClient{},
			authManager: &MockAuthManager{loadTokenResult: "mock-token"},
		},
		{
			name:        "Different output writer",
			input:       NewMockInputReader(),
			output:      &strings.Builder{},
			cfg:         &Config{ServerURL: "http://example.com"},
			client:      &MockTaskClient{},
			authManager: &MockAuthManager{loadTokenResult: "mock-token"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			result := NewCLI(tc.input, tc.output, tc.cfg, tc.client, tc.authManager)

			// ====Assert====
			if result == nil {
				t.Fatal("Expected non-nil CLI")
			}
			if result.input != tc.input {
				t.Error("Expected input to be set correctly")
			}
			if result.output != tc.output {
				t.Error("Expected output to be set correctly")
			}
			if result.config != tc.cfg {
				t.Error("Expected config to be set correctly")
			}
			if result.client != tc.client {
				t.Error("Expected client to be set correctly")
			}
			if result.authManager != tc.authManager {
				t.Error("Expected authManager to be set correctly")
			}
		})
	}
}

// TestConsoleInputReader_ReadInput tests the ReadInput method with validation logic
func TestConsoleInputReader_ReadInput(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name        string
		input       string
		maxSize     int
		expected    string
		expectedErr error
	}{
		{
			name:        "Valid input within size limit",
			input:       "test input\n",
			maxSize:     20,
			expected:    "test input",
			expectedErr: nil,
		},
		{
			name:        "Input exceeds max size",
			input:       "this is a very long input that exceeds the limit\n",
			maxSize:     10,
			expected:    "",
			expectedErr: ErrMaxSizeExceeded,
		},
		{
			name:        "Empty input after trimming",
			input:       "\n",
			maxSize:     20,
			expected:    "",
			expectedErr: ErrEmptyInput,
		},
		{
			name:        "Input with leading and trailing whitespace",
			input:       "  test input  \n",
			maxSize:     20,
			expected:    "test input",
			expectedErr: nil,
		},
		{
			name:        "Input with only whitespace",
			input:       "     \n",
			maxSize:     20,
			expected:    "",
			expectedErr: ErrEmptyInput,
		},
		{
			name:        "Input exactly at max size",
			input:       "1234567890\n",
			maxSize:     10,
			expected:    "1234567890",
			expectedErr: nil,
		},
		{
			name:        "Input one character over max size",
			input:       "12345678901\n",
			maxSize:     10,
			expected:    "",
			expectedErr: ErrMaxSizeExceeded,
		},
		{
			name:        "EOF - empty reader",
			input:       "",
			maxSize:     20,
			expected:    "",
			expectedErr: io.EOF,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create real ConsoleInputReader with test input
			reader := NewConsoleInputReader(strings.NewReader(tc.input))

			// ====Act====
			result, err := reader.ReadInput(tc.maxSize)

			// ====Assert====
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// TestCLI_showHelp tests the showHelp method
func TestCLI_showHelp(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name             string
		expectedCommands []string
	}{
		{
			name: "Displays all available commands",
			expectedCommands: []string{
				"add",
				"status",
				"list",
				"process",
				"clear",
				"update",
				"delete",
				"login",
				"register",
				"logout",
				"help",
				"exit",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			cli := NewCLI(
				NewMockInputReader(),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				&MockTaskClient{},
				&MockAuthManager{loadTokenResult: "mock-token"},
			)

			// ====Act====
			cli.showHelp()

			// ====Assert====
			result := output.String()

			// Verify help header is present
			if !strings.Contains(result, "=== Available Commands ===") {
				t.Error("Expected help header to be present")
			}

			// Verify all commands are listed
			for _, cmd := range tc.expectedCommands {
				if !strings.Contains(result, cmd) {
					t.Errorf("Expected command %q to be listed in help", cmd)
				}
			}

			// Verify help footer is present
			if !strings.Contains(result, "==========================") {
				t.Error("Expected help footer to be present")
			}
		})
	}
}

// TestCLI_handleError tests the handleError method with different error types
func TestCLI_handleError(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name             string
		err              error
		context          string
		expectedContains []string
	}{
		{
			name:    "EOF error",
			err:     io.EOF,
			context: "Input error",
			expectedContains: []string{
				"Input error",
				"input interrupted by user",
			},
		},
		{
			name: "NetworkError",
			err: &client.NetworkError{
				URL: "http://localhost:8080",
				Err: errors.New("connection refused"),
			},
			context: "Connection error",
			expectedContains: []string{
				"❌",
				"Connection error",
				"Cannot connect to server at http://localhost:8080",
				"Please check that the server is running",
			},
		},
		{
			name: "APIError",
			err: &client.APIError{
				StatusCode: 500,
				Message:    "Internal server error",
			},
			context: "API error",
			expectedContains: []string{
				"❌",
				"API error",
				"Internal server error",
			},
		},
		{
			name:    "Generic error",
			err:     errors.New("something went wrong"),
			context: "Operation failed",
			expectedContains: []string{
				"Operation failed",
				"something went wrong",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			cli := NewCLI(
				NewMockInputReader(),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				&MockTaskClient{},
				&MockAuthManager{loadTokenResult: "mock-token"},
			)

			// ====Act====
			cli.handleError(tc.err, tc.context)

			// ====Assert====
			result := output.String()

			for _, expected := range tc.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q, got: %s", expected, result)
				}
			}
		})
	}
}

// TestCLI_promptForTaskID tests the promptForTaskID method
func TestCLI_promptForTaskID(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name        string
		input       string
		prompt      string
		expectedID  int
		expectedErr error
	}{
		{
			name:        "Valid task ID",
			input:       "5",
			prompt:      "Enter task ID:\n",
			expectedID:  5,
			expectedErr: nil,
		},
		{
			name:        "Invalid ID - non-numeric",
			input:       "abc",
			prompt:      "Enter task ID:\n",
			expectedID:  0,
			expectedErr: validation.ErrInvalidTaskID,
		},
		{
			name:        "Invalid ID - negative number",
			input:       "-1",
			prompt:      "Enter task ID:\n",
			expectedID:  0,
			expectedErr: validation.ErrInvalidTaskID,
		},
		{
			name:        "Invalid ID - zero",
			input:       "0",
			prompt:      "Enter task ID:\n",
			expectedID:  0,
			expectedErr: validation.ErrInvalidTaskID,
		},
		{
			name:        "Input too long",
			input:       "12345678901",
			prompt:      "Enter task ID:\n",
			expectedID:  0,
			expectedErr: ErrMaxSizeExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			cli := NewCLI(
				NewMockInputReader(tc.input),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				&MockTaskClient{},
				&MockAuthManager{loadTokenResult: "mock-token"},
			)

			// ====Act====
			id, err := cli.promptForTaskID(tc.prompt)

			// ====Assert====
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if id != tc.expectedID {
				t.Errorf("Expected ID %d, got %d", tc.expectedID, id)
			}

			// Verify prompt was written to output
			if !strings.Contains(output.String(), tc.prompt) {
				t.Errorf("Expected prompt %q to be written to output", tc.prompt)
			}
		})
	}
}
