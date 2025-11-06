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

	"github.com/stretchr/testify/assert"
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

// TestCLI_handleAddCommand tests the handleAddCommand method
func TestCLI_handleAddCommand(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name             string
		input            string
		createTaskResult *client.Task
		createTaskErr    error
		expectedErr      error
		expectedContains string
	}{
		{
			name:  "Valid task description",
			input: "Buy groceries",
			createTaskResult: &client.Task{
				ID:          1,
				Description: "Buy groceries",
				Done:        false,
			},
			createTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "✅ Task added (ID: 1)",
		},
		{
			name:  "Task description with whitespace",
			input: "  Clean room  ",
			createTaskResult: &client.Task{
				ID:          2,
				Description: "Clean room",
				Done:        false,
			},
			createTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "✅ Task added (ID: 2)",
		},
		{
			name:             "Empty input",
			input:            "",
			createTaskResult: nil,
			createTaskErr:    nil,
			expectedErr:      ErrEmptyInput,
			expectedContains: "",
		},
		{
			name:             "Input exceeds max size",
			input:            strings.Repeat("a", 201),
			createTaskResult: nil,
			createTaskErr:    nil,
			expectedErr:      ErrMaxSizeExceeded,
			expectedContains: "",
		},
		{
			name:             "Client CreateTask fails with generic error",
			input:            "Valid task",
			createTaskResult: nil,
			createTaskErr:    errors.New("database error"),
			expectedErr:      nil, // Check manually that error is wrapped
			expectedContains: "",
		},
		{
			name:             "Network error from client",
			input:            "Valid task",
			createTaskResult: nil,
			createTaskErr: &client.NetworkError{
				URL: "http://localhost:8080",
				Err: errors.New("connection refused"),
			},
			expectedErr:      &client.NetworkError{},
			expectedContains: "",
		},
		{
			name:             "API error from client",
			input:            "Valid task",
			createTaskResult: nil,
			createTaskErr: &client.APIError{
				StatusCode: 401,
				Message:    "Unauthorized",
			},
			expectedErr:      &client.APIError{},
			expectedContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockClient := &MockTaskClient{
				createTaskResult: tc.createTaskResult,
				createTaskErr:    tc.createTaskErr,
			}
			cli := NewCLI(
				NewMockInputReader(tc.input),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				&MockAuthManager{loadTokenResult: "mock-token"},
			)

			// ====Act====
			err := cli.handleAddCommand()

			// ====Assert====
			if tc.expectedErr != nil {
				assert.Error(t, err, "Expected an error but got nil")

				// Check if it's the expected error type
				switch tc.expectedErr.(type) {
				case *client.NetworkError:
					var netErr *client.NetworkError
					assert.ErrorAs(t, err, &netErr, "Expected NetworkError")
				case *client.APIError:
					var apiErr *client.APIError
					assert.ErrorAs(t, err, &apiErr, "Expected APIError")
				default:
					assert.ErrorIs(t, err, tc.expectedErr, "Expected specific error")
				}
			} else if tc.name == "Client CreateTask fails with generic error" {
				// Special case: verify error is wrapped with context
				assert.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), "adding task: creation failed", "Error should contain context")
				assert.Contains(t, err.Error(), "database error", "Error should contain original error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			// Verify output contains expected message
			if tc.expectedContains != "" {
				assert.Contains(t, output.String(), tc.expectedContains, "Output should contain expected message")
			}

			// Verify prompt was displayed
			assert.Contains(t, output.String(), "Enter task description:", "Prompt should be displayed")
		})
	}
}

// TestCLI_handleStatusCommand tests the handleStatusCommand method
func TestCLI_handleStatusCommand(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name             string
		taskIDInput      string
		statusInput      string
		getTaskResult    *client.Task
		getTaskErr       error
		updateTaskResult *client.Task
		updateTaskErr    error
		expectedErr      error
		expectedContains string
	}{
		{
			name:        "Change status to done",
			taskIDInput: "1",
			statusInput: "done",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Test task",
				Done:        false,
			},
			getTaskErr: nil,
			updateTaskResult: &client.Task{
				ID:          1,
				Description: "Test task",
				Done:        true,
			},
			updateTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "✅ Task (ID: 1) status is has changed",
		},
		{
			name:        "Change status to undone",
			taskIDInput: "2",
			statusInput: "undone",
			getTaskResult: &client.Task{
				ID:          2,
				Description: "Completed task",
				Done:        true,
			},
			getTaskErr: nil,
			updateTaskResult: &client.Task{
				ID:          2,
				Description: "Completed task",
				Done:        false,
			},
			updateTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "✅ Task (ID: 2) status is has changed",
		},
		{
			name:             "Invalid task ID - non-numeric",
			taskIDInput:      "abc",
			statusInput:      "done",
			getTaskResult:    nil,
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      validation.ErrInvalidTaskID,
			expectedContains: "",
		},
		{
			name:             "Task not found",
			taskIDInput:      "999",
			statusInput:      "done",
			getTaskResult:    nil,
			getTaskErr:       errors.New("task not found"),
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      nil, // Will check error message contains text
			expectedContains: "",
		},
		{
			name:        "Invalid status - not done or undone",
			taskIDInput: "1",
			statusInput: "completed",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Test task",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      ErrInvalidStatus,
			expectedContains: "",
		},
		{
			name:        "Invalid status - empty input",
			taskIDInput: "1",
			statusInput: "",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Test task",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      ErrEmptyInput,
			expectedContains: "",
		},
		{
			name:        "Status input too long",
			taskIDInput: "1",
			statusInput: "verylongstatus",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Test task",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      ErrMaxSizeExceeded,
			expectedContains: "",
		},
		{
			name:        "Client UpdateTask fails",
			taskIDInput: "1",
			statusInput: "done",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Test task",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    errors.New("database error"),
			expectedErr:      nil, // Will check error is wrapped
			expectedContains: "",
		},
		{
			name:        "Network error from client",
			taskIDInput: "1",
			statusInput: "done",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Test task",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr: &client.NetworkError{
				URL: "http://localhost:8080",
				Err: errors.New("connection refused"),
			},
			expectedErr:      &client.NetworkError{},
			expectedContains: "",
		},
		{
			name:        "API error from client",
			taskIDInput: "1",
			statusInput: "done",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Test task",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr: &client.APIError{
				StatusCode: 403,
				Message:    "Forbidden",
			},
			expectedErr:      &client.APIError{},
			expectedContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockClient := &MockTaskClient{
				getTaskResult:    tc.getTaskResult,
				getTaskErr:       tc.getTaskErr,
				updateTaskResult: tc.updateTaskResult,
				updateTaskErr:    tc.updateTaskErr,
			}
			cli := NewCLI(
				NewMockInputReader(tc.taskIDInput, tc.statusInput),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				&MockAuthManager{loadTokenResult: "mock-token"},
			)

			// ====Act====
			err := cli.handleStatusCommand()

			// ====Assert====
			if tc.expectedErr != nil {
				assert.Error(t, err, "Expected an error but got nil")

				// Check if it's the expected error type
				switch tc.expectedErr.(type) {
				case *client.NetworkError:
					var netErr *client.NetworkError
					assert.ErrorAs(t, err, &netErr, "Expected NetworkError")
				case *client.APIError:
					var apiErr *client.APIError
					assert.ErrorAs(t, err, &apiErr, "Expected APIError")
				default:
					assert.ErrorIs(t, err, tc.expectedErr, "Expected specific error")
				}
			} else if tc.name == "Task not found" {
				// Special case: verify error is wrapped with context
				assert.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), "updating status: task id validation failed", "Error should contain context")
			} else if tc.name == "Client UpdateTask fails" {
				// Special case: verify error is wrapped with context
				assert.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), "updating status for task id", "Error should contain context")
				assert.Contains(t, err.Error(), "database error", "Error should contain original error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			// Verify output contains expected message
			if tc.expectedContains != "" {
				assert.Contains(t, output.String(), tc.expectedContains, "Output should contain expected message")
			}

			// Verify prompts were displayed (for successful cases)
			if tc.expectedErr == nil && tc.name != "Task not found" && tc.name != "Client UpdateTask fails" {
				assert.Contains(t, output.String(), "Enter task ID to change status:", "Task ID prompt should be displayed")
				assert.Contains(t, output.String(), "Enter new status 'done' // 'undone'", "Status prompt should be displayed")
			}
		})
	}
}

// TestCLI_handleClearCommand tests the handleClearCommand method
func TestCLI_handleClearCommand(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name             string
		taskIDInput      string
		getTaskResult    *client.Task
		getTaskErr       error
		updateTaskResult *client.Task
		updateTaskErr    error
		expectedErr      error
		expectedContains string
	}{
		{
			name:        "Successfully clear task description",
			taskIDInput: "1",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task to be cleared",
				Done:        false,
			},
			getTaskErr: nil,
			updateTaskResult: &client.Task{
				ID:          1,
				Description: "",
				Done:        false,
			},
			updateTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "✅ Task (ID: 1) description cleared!",
		},
		{
			name:        "Clear already empty description",
			taskIDInput: "2",
			getTaskResult: &client.Task{
				ID:          2,
				Description: "",
				Done:        true,
			},
			getTaskErr: nil,
			updateTaskResult: &client.Task{
				ID:          2,
				Description: "",
				Done:        true,
			},
			updateTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "✅ Task (ID: 2) description cleared!",
		},
		{
			name:             "Invalid task ID - non-numeric",
			taskIDInput:      "abc",
			getTaskResult:    nil,
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      validation.ErrInvalidTaskID,
			expectedContains: "",
		},
		{
			name:             "Invalid task ID - zero",
			taskIDInput:      "0",
			getTaskResult:    nil,
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      validation.ErrInvalidTaskID,
			expectedContains: "",
		},
		{
			name:             "Task not found",
			taskIDInput:      "999",
			getTaskResult:    nil,
			getTaskErr:       errors.New("task not found"),
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      nil, // Will check error message contains text
			expectedContains: "",
		},
		{
			name:        "Client UpdateTask fails",
			taskIDInput: "1",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task description",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    errors.New("database error"),
			expectedErr:      nil, // Will check error is wrapped
			expectedContains: "",
		},
		{
			name:        "Network error from client",
			taskIDInput: "1",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task description",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr: &client.NetworkError{
				URL: "http://localhost:8080",
				Err: errors.New("connection refused"),
			},
			expectedErr:      &client.NetworkError{},
			expectedContains: "",
		},
		{
			name:        "API error from client",
			taskIDInput: "1",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task description",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr: &client.APIError{
				StatusCode: 404,
				Message:    "Task not found",
			},
			expectedErr:      &client.APIError{},
			expectedContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockClient := &MockTaskClient{
				getTaskResult:    tc.getTaskResult,
				getTaskErr:       tc.getTaskErr,
				updateTaskResult: tc.updateTaskResult,
				updateTaskErr:    tc.updateTaskErr,
			}
			cli := NewCLI(
				NewMockInputReader(tc.taskIDInput),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				&MockAuthManager{loadTokenResult: "mock-token"},
			)

			// ====Act====
			err := cli.handleClearCommand()

			// ====Assert====
			if tc.expectedErr != nil {
				assert.Error(t, err, "Expected an error but got nil")

				// Check if it's the expected error type
				switch tc.expectedErr.(type) {
				case *client.NetworkError:
					var netErr *client.NetworkError
					assert.ErrorAs(t, err, &netErr, "Expected NetworkError")
				case *client.APIError:
					var apiErr *client.APIError
					assert.ErrorAs(t, err, &apiErr, "Expected APIError")
				default:
					assert.ErrorIs(t, err, tc.expectedErr, "Expected specific error")
				}
			} else if tc.name == "Task not found" {
				// Special case: verify error is wrapped with context
				assert.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), "clearing task description: task id validation failed", "Error should contain context")
			} else if tc.name == "Client UpdateTask fails" {
				// Special case: verify error is wrapped with context
				assert.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), "clearing task description for task id", "Error should contain context")
				assert.Contains(t, err.Error(), "database error", "Error should contain original error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			// Verify output contains expected message
			if tc.expectedContains != "" {
				assert.Contains(t, output.String(), tc.expectedContains, "Output should contain expected message")
			}

			// Verify prompt was displayed (for successful cases)
			if tc.expectedErr == nil && tc.name != "Task not found" && tc.name != "Client UpdateTask fails" {
				assert.Contains(t, output.String(), "Enter task ID you want to clear description", "Task ID prompt should be displayed")
			}
		})
	}
}

// TestCLI_handleUpdateCommand tests the handleUpdateCommand method
func TestCLI_handleUpdateCommand(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name             string
		taskIDInput      string
		descriptionInput string
		getTaskResult    *client.Task
		getTaskErr       error
		updateTaskResult *client.Task
		updateTaskErr    error
		expectedErr      error
		expectedContains string
	}{
		{
			name:             "Successfully update task description",
			taskIDInput:      "1",
			descriptionInput: "Updated description",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Old description",
				Done:        false,
			},
			getTaskErr: nil,
			updateTaskResult: &client.Task{
				ID:          1,
				Description: "Updated description",
				Done:        false,
			},
			updateTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "✅ Task (ID: 1) updated",
		},
		{
			name:             "Update with whitespace trimming",
			taskIDInput:      "2",
			descriptionInput: "  New description  ",
			getTaskResult: &client.Task{
				ID:          2,
				Description: "Old description",
				Done:        true,
			},
			getTaskErr: nil,
			updateTaskResult: &client.Task{
				ID:          2,
				Description: "New description",
				Done:        true,
			},
			updateTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "✅ Task (ID: 2) updated",
		},
		{
			name:             "Invalid task ID - non-numeric",
			taskIDInput:      "abc",
			descriptionInput: "New description",
			getTaskResult:    nil,
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      validation.ErrInvalidTaskID,
			expectedContains: "",
		},
		{
			name:             "Task not found",
			taskIDInput:      "999",
			descriptionInput: "New description",
			getTaskResult:    nil,
			getTaskErr:       errors.New("task not found"),
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      nil, // Will check error message contains text
			expectedContains: "",
		},
		{
			name:             "Empty new description",
			taskIDInput:      "1",
			descriptionInput: "",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Old description",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      ErrEmptyInput,
			expectedContains: "",
		},
		{
			name:             "New description too long",
			taskIDInput:      "1",
			descriptionInput: strings.Repeat("a", 201),
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Old description",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      ErrMaxSizeExceeded,
			expectedContains: "",
		},
		{
			name:             "Description unchanged - same as current",
			taskIDInput:      "1",
			descriptionInput: "Same description",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Same description",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    nil,
			expectedErr:      ErrDescUnchanged,
			expectedContains: "",
		},
		{
			name:             "Client UpdateTask fails",
			taskIDInput:      "1",
			descriptionInput: "New description",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Old description",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr:    errors.New("database error"),
			expectedErr:      nil, // Will check error is wrapped
			expectedContains: "",
		},
		{
			name:             "Network error from client",
			taskIDInput:      "1",
			descriptionInput: "New description",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Old description",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr: &client.NetworkError{
				URL: "http://localhost:8080",
				Err: errors.New("connection refused"),
			},
			expectedErr:      &client.NetworkError{},
			expectedContains: "",
		},
		{
			name:             "API error from client",
			taskIDInput:      "1",
			descriptionInput: "New description",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Old description",
				Done:        false,
			},
			getTaskErr:       nil,
			updateTaskResult: nil,
			updateTaskErr: &client.APIError{
				StatusCode: 400,
				Message:    "Invalid description",
			},
			expectedErr:      &client.APIError{},
			expectedContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockClient := &MockTaskClient{
				getTaskResult:    tc.getTaskResult,
				getTaskErr:       tc.getTaskErr,
				updateTaskResult: tc.updateTaskResult,
				updateTaskErr:    tc.updateTaskErr,
			}
			cli := NewCLI(
				NewMockInputReader(tc.taskIDInput, tc.descriptionInput),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				&MockAuthManager{loadTokenResult: "mock-token"},
			)

			// ====Act====
			err := cli.handleUpdateCommand()

			// ====Assert====
			if tc.expectedErr != nil {
				assert.Error(t, err, "Expected an error but got nil")

				// Check if it's the expected error type
				switch tc.expectedErr.(type) {
				case *client.NetworkError:
					var netErr *client.NetworkError
					assert.ErrorAs(t, err, &netErr, "Expected NetworkError")
				case *client.APIError:
					var apiErr *client.APIError
					assert.ErrorAs(t, err, &apiErr, "Expected APIError")
				default:
					assert.ErrorIs(t, err, tc.expectedErr, "Expected specific error")
				}
			} else if tc.name == "Task not found" {
				// Special case: verify error is wrapped with context
				assert.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), "updating task description: task id validation failed", "Error should contain context")
			} else if tc.name == "Client UpdateTask fails" {
				// Special case: verify error is wrapped with context
				assert.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), "updating task description for task id", "Error should contain context")
				assert.Contains(t, err.Error(), "database error", "Error should contain original error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			// Verify output contains expected message
			if tc.expectedContains != "" {
				assert.Contains(t, output.String(), tc.expectedContains, "Output should contain expected message")
			}

			// Verify prompts were displayed (for successful cases)
			if tc.expectedErr == nil && tc.name != "Task not found" && tc.name != "Client UpdateTask fails" {
				assert.Contains(t, output.String(), "Enter task ID to update:", "Task ID prompt should be displayed")
				assert.Contains(t, output.String(), "Enter new description:", "Description prompt should be displayed")
			}
		})
	}
}

// TestCLI_handleDeleteCommand tests the handleDeleteCommand method
func TestCLI_handleDeleteCommand(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name             string
		taskIDInput      string
		confirmInput     string
		getTaskResult    *client.Task
		getTaskErr       error
		deleteTaskErr    error
		expectedErr      error
		expectedContains string
	}{
		{
			name:         "Successfully delete task with 'y' confirmation",
			taskIDInput:  "1",
			confirmInput: "y",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task to delete",
				Done:        false,
			},
			getTaskErr:       nil,
			deleteTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "✅ Task (ID: 1) deleted",
		},
		{
			name:         "Successfully delete task with uppercase 'Y' confirmation",
			taskIDInput:  "2",
			confirmInput: "Y",
			getTaskResult: &client.Task{
				ID:          2,
				Description: "Another task",
				Done:        true,
			},
			getTaskErr:       nil,
			deleteTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "✅ Task (ID: 2) deleted",
		},
		{
			name:         "Cancel deletion with 'n' confirmation",
			taskIDInput:  "3",
			confirmInput: "n",
			getTaskResult: &client.Task{
				ID:          3,
				Description: "Task not to delete",
				Done:        false,
			},
			getTaskErr:       nil,
			deleteTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "Deletion canceled",
		},
		{
			name:         "Cancel deletion with uppercase 'N' confirmation",
			taskIDInput:  "4",
			confirmInput: "N",
			getTaskResult: &client.Task{
				ID:          4,
				Description: "Another task not to delete",
				Done:        true,
			},
			getTaskErr:       nil,
			deleteTaskErr:    nil,
			expectedErr:      nil,
			expectedContains: "Deletion canceled",
		},
		{
			name:             "Invalid task ID - non-numeric",
			taskIDInput:      "abc",
			confirmInput:     "y",
			getTaskResult:    nil,
			getTaskErr:       nil,
			deleteTaskErr:    nil,
			expectedErr:      validation.ErrInvalidTaskID,
			expectedContains: "",
		},
		{
			name:             "Task not found",
			taskIDInput:      "999",
			confirmInput:     "y",
			getTaskResult:    nil,
			getTaskErr:       errors.New("task not found"),
			deleteTaskErr:    nil,
			expectedErr:      nil, // Will check error message contains text
			expectedContains: "",
		},
		{
			name:         "Invalid confirmation - not y or n",
			taskIDInput:  "1",
			confirmInput: "yes",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task",
				Done:        false,
			},
			getTaskErr:       nil,
			deleteTaskErr:    nil,
			expectedErr:      ErrInvalidConfirmChoice,
			expectedContains: "",
		},
		{
			name:         "Invalid confirmation - empty input",
			taskIDInput:  "1",
			confirmInput: "",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task",
				Done:        false,
			},
			getTaskErr:       nil,
			deleteTaskErr:    nil,
			expectedErr:      ErrEmptyInput,
			expectedContains: "",
		},
		{
			name:         "Confirmation input too long",
			taskIDInput:  "1",
			confirmInput: "verylonginput",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task",
				Done:        false,
			},
			getTaskErr:       nil,
			deleteTaskErr:    nil,
			expectedErr:      ErrMaxSizeExceeded,
			expectedContains: "",
		},
		{
			name:         "Client DeleteTask fails",
			taskIDInput:  "1",
			confirmInput: "y",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task",
				Done:        false,
			},
			getTaskErr:       nil,
			deleteTaskErr:    errors.New("database error"),
			expectedErr:      nil, // Will check error is wrapped
			expectedContains: "",
		},
		{
			name:         "Network error from client",
			taskIDInput:  "1",
			confirmInput: "y",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task",
				Done:        false,
			},
			getTaskErr: nil,
			deleteTaskErr: &client.NetworkError{
				URL: "http://localhost:8080",
				Err: errors.New("connection refused"),
			},
			expectedErr:      &client.NetworkError{},
			expectedContains: "",
		},
		{
			name:         "API error from client",
			taskIDInput:  "1",
			confirmInput: "y",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Task",
				Done:        false,
			},
			getTaskErr: nil,
			deleteTaskErr: &client.APIError{
				StatusCode: 403,
				Message:    "Forbidden",
			},
			expectedErr:      &client.APIError{},
			expectedContains: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockClient := &MockTaskClient{
				getTaskResult: tc.getTaskResult,
				getTaskErr:    tc.getTaskErr,
				deleteTaskErr: tc.deleteTaskErr,
			}
			cli := NewCLI(
				NewMockInputReader(tc.taskIDInput, tc.confirmInput),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				&MockAuthManager{loadTokenResult: "mock-token"},
			)

			// ====Act====
			err := cli.handleDeleteCommand()

			// ====Assert====
			if tc.expectedErr != nil {
				assert.Error(t, err, "Expected an error but got nil")

				// Check if it's the expected error type
				switch tc.expectedErr.(type) {
				case *client.NetworkError:
					var netErr *client.NetworkError
					assert.ErrorAs(t, err, &netErr, "Expected NetworkError")
				case *client.APIError:
					var apiErr *client.APIError
					assert.ErrorAs(t, err, &apiErr, "Expected APIError")
				default:
					assert.ErrorIs(t, err, tc.expectedErr, "Expected specific error")
				}
			} else if tc.name == "Task not found" {
				// Special case: verify error is wrapped with context
				assert.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), "deleting task: id validation failed", "Error should contain context")
			} else if tc.name == "Client DeleteTask fails" {
				// Special case: verify error is wrapped with context
				assert.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), "deleting task id", "Error should contain context")
				assert.Contains(t, err.Error(), "database error", "Error should contain original error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			// Verify output contains expected message
			if tc.expectedContains != "" {
				assert.Contains(t, output.String(), tc.expectedContains, "Output should contain expected message")
			}

			// Verify prompts were displayed (for cases that get past task ID validation)
			if tc.getTaskResult != nil {
				assert.Contains(t, output.String(), "Enter task ID to delete task:", "Task ID prompt should be displayed")
				assert.Contains(t, output.String(), "Enter y/N:", "Confirmation prompt should be displayed")
			}
		})
	}
}

// TestCLI_handleListCommand tests the handleListCommand method
func TestCLI_handleListCommand(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name             string
		getTasksResult   []client.Task
		getTasksErr      error
		expectedErr      error
		expectedContains []string
	}{
		{
			name: "Successfully list multiple tasks",
			getTasksResult: []client.Task{
				{ID: 1, Description: "Buy groceries", Done: false},
				{ID: 2, Description: "Clean room", Done: true},
				{ID: 3, Description: "Write report", Done: false},
			},
			getTasksErr: nil,
			expectedErr: nil,
			expectedContains: []string{
				"=== Your Tasks ===",
				"[ ] 1: Buy groceries",
				"[✓] 2: Clean room",
				"[ ] 3: Write report",
				"==================",
			},
		},
		{
			name: "Successfully list single task",
			getTasksResult: []client.Task{
				{ID: 1, Description: "Single task", Done: false},
			},
			getTasksErr: nil,
			expectedErr: nil,
			expectedContains: []string{
				"=== Your Tasks ===",
				"[ ] 1: Single task",
				"==================",
			},
		},
		{
			name: "List completed tasks only",
			getTasksResult: []client.Task{
				{ID: 5, Description: "Completed task 1", Done: true},
				{ID: 10, Description: "Completed task 2", Done: true},
			},
			getTasksErr: nil,
			expectedErr: nil,
			expectedContains: []string{
				"=== Your Tasks ===",
				"[✓] 5: Completed task 1",
				"[✓] 10: Completed task 2",
				"==================",
			},
		},
		{
			name: "List tasks with empty descriptions",
			getTasksResult: []client.Task{
				{ID: 1, Description: "", Done: false},
				{ID: 2, Description: "Normal task", Done: true},
			},
			getTasksErr: nil,
			expectedErr: nil,
			expectedContains: []string{
				"=== Your Tasks ===",
				"[ ] 1: ",
				"[✓] 2: Normal task",
				"==================",
			},
		},
		{
			name: "List tasks with long descriptions",
			getTasksResult: []client.Task{
				{ID: 1, Description: "This is a very long task description that should be displayed completely without truncation", Done: false},
				{ID: 2, Description: "Another long description with many words to test the display formatting", Done: true},
			},
			getTasksErr: nil,
			expectedErr: nil,
			expectedContains: []string{
				"=== Your Tasks ===",
				"[ ] 1: This is a very long task description that should be displayed completely without truncation",
				"[✓] 2: Another long description with many words to test the display formatting",
				"==================",
			},
		},
		{
			name:           "Empty task list",
			getTasksResult: []client.Task{},
			getTasksErr:    nil,
			expectedErr:    nil,
			expectedContains: []string{
				"No tasks found",
			},
		},
		{
			name:           "Nil task list",
			getTasksResult: nil,
			getTasksErr:    nil,
			expectedErr:    nil,
			expectedContains: []string{
				"No tasks found",
			},
		},
		{
			name:             "Client GetTasks fails with generic error",
			getTasksResult:   nil,
			getTasksErr:      errors.New("database error"),
			expectedErr:      nil, // Will check error is wrapped
			expectedContains: []string{},
		},
		{
			name:           "Network error from client",
			getTasksResult: nil,
			getTasksErr: &client.NetworkError{
				URL: "http://localhost:8080",
				Err: errors.New("connection refused"),
			},
			expectedErr:      &client.NetworkError{},
			expectedContains: []string{},
		},
		{
			name:           "API error from client - 401 Unauthorized",
			getTasksResult: nil,
			getTasksErr: &client.APIError{
				StatusCode: 401,
				Message:    "Unauthorized",
			},
			expectedErr:      &client.APIError{},
			expectedContains: []string{},
		},
		{
			name:           "API error from client - 500 Internal Server Error",
			getTasksResult: nil,
			getTasksErr: &client.APIError{
				StatusCode: 500,
				Message:    "Internal server error",
			},
			expectedErr:      &client.APIError{},
			expectedContains: []string{},
		},
		{
			name: "List tasks with special characters in descriptions",
			getTasksResult: []client.Task{
				{ID: 1, Description: "Task with emoji 🎉", Done: false},
				{ID: 2, Description: "Task with symbols: @#$%^&*()", Done: true},
				{ID: 3, Description: "Task with quotes \"test\"", Done: false},
			},
			getTasksErr: nil,
			expectedErr: nil,
			expectedContains: []string{
				"=== Your Tasks ===",
				"[ ] 1: Task with emoji 🎉",
				"[✓] 2: Task with symbols: @#$%^&*()",
				"[ ] 3: Task with quotes \"test\"",
				"==================",
			},
		},
		{
			name: "List tasks with large ID numbers",
			getTasksResult: []client.Task{
				{ID: 999999, Description: "Task with large ID", Done: false},
				{ID: 1000000, Description: "Task with even larger ID", Done: true},
			},
			getTasksErr: nil,
			expectedErr: nil,
			expectedContains: []string{
				"=== Your Tasks ===",
				"[ ] 999999: Task with large ID",
				"[✓] 1000000: Task with even larger ID",
				"==================",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockClient := &MockTaskClient{
				getTasksResult: tc.getTasksResult,
				getTasksErr:    tc.getTasksErr,
			}
			cli := NewCLI(
				NewMockInputReader(),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				&MockAuthManager{loadTokenResult: "mock-token"},
			)

			// ====Act====
			err := cli.handleListCommand()

			// ====Assert====
			if tc.expectedErr != nil {
				assert.Error(t, err, "Expected an error but got nil")

				// Check if it's the expected error type
				switch tc.expectedErr.(type) {
				case *client.NetworkError:
					var netErr *client.NetworkError
					assert.ErrorAs(t, err, &netErr, "Expected NetworkError")
				case *client.APIError:
					var apiErr *client.APIError
					assert.ErrorAs(t, err, &apiErr, "Expected APIError")
				default:
					assert.ErrorIs(t, err, tc.expectedErr, "Expected specific error")
				}
			} else if tc.name == "Client GetTasks fails with generic error" {
				// Special case: verify error is wrapped with context
				assert.Error(t, err, "Expected error but got nil")
				assert.Contains(t, err.Error(), "failed to retrieve tasks", "Error should contain context")
				assert.Contains(t, err.Error(), "database error", "Error should contain original error")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			// Verify output contains expected messages
			result := output.String()
			for _, expected := range tc.expectedContains {
				assert.Contains(t, result, expected, "Output should contain expected message")
			}

			// Verify "No tasks found" is NOT shown when tasks exist
			if len(tc.getTasksResult) > 0 && tc.getTasksErr == nil {
				assert.NotContains(t, result, "No tasks found", "Should not show 'No tasks found' when tasks exist")
			}

			// Verify task headers are NOT shown when no tasks exist
			if (len(tc.getTasksResult) == 0 || tc.getTasksResult == nil) && tc.getTasksErr == nil {
				assert.NotContains(t, result, "=== Your Tasks ===", "Should not show task header when no tasks exist")
				assert.NotContains(t, result, "==================", "Should not show task footer when no tasks exist")
			}
		})
	}
}

// TestCLI_promptForTaskWithDisplay tests the promptForTaskWithDisplay method
func TestCLI_promptForTaskWithDisplay(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name             string
		taskIDInput      string
		prompt           string
		getTaskResult    *client.Task
		getTaskErr       error
		expectedID       int
		expectedTask     *client.Task
		expectedErr      error
		expectedContains string
	}{
		{
			name:        "Successfully retrieve and display task",
			taskIDInput: "1",
			prompt:      "Enter task ID:\n",
			getTaskResult: &client.Task{
				ID:          1,
				Description: "Test task",
				Done:        false,
			},
			getTaskErr:       nil,
			expectedID:       1,
			expectedTask:     &client.Task{ID: 1, Description: "Test task", Done: false},
			expectedErr:      nil,
			expectedContains: "Current task: '[ ] 1: Test task'",
		},
		{
			name:        "Display completed task",
			taskIDInput: "2",
			prompt:      "Enter task ID:\n",
			getTaskResult: &client.Task{
				ID:          2,
				Description: "Completed task",
				Done:        true,
			},
			getTaskErr:       nil,
			expectedID:       2,
			expectedTask:     &client.Task{ID: 2, Description: "Completed task", Done: true},
			expectedErr:      nil,
			expectedContains: "Current task: '[✓] 2: Completed task'",
		},
		{
			name:             "Invalid task ID - non-numeric",
			taskIDInput:      "abc",
			prompt:           "Enter task ID:\n",
			getTaskResult:    nil,
			getTaskErr:       nil,
			expectedID:       0,
			expectedTask:     nil,
			expectedErr:      validation.ErrInvalidTaskID,
			expectedContains: "",
		},
		{
			name:             "Invalid task ID - zero",
			taskIDInput:      "0",
			prompt:           "Enter task ID:\n",
			getTaskResult:    nil,
			getTaskErr:       nil,
			expectedID:       0,
			expectedTask:     nil,
			expectedErr:      validation.ErrInvalidTaskID,
			expectedContains: "",
		},
		{
			name:             "Invalid task ID - negative",
			taskIDInput:      "-5",
			prompt:           "Enter task ID:\n",
			getTaskResult:    nil,
			getTaskErr:       nil,
			expectedID:       0,
			expectedTask:     nil,
			expectedErr:      validation.ErrInvalidTaskID,
			expectedContains: "",
		},
		{
			name:             "Task not found",
			taskIDInput:      "999",
			prompt:           "Enter task ID:\n",
			getTaskResult:    nil,
			getTaskErr:       errors.New("task not found"),
			expectedID:       0,
			expectedTask:     nil,
			expectedErr:      nil, // Will check error is returned
			expectedContains: "",
		},
		{
			name:          "Network error from client",
			taskIDInput:   "1",
			prompt:        "Enter task ID:\n",
			getTaskResult: nil,
			getTaskErr: &client.NetworkError{
				URL: "http://localhost:8080",
				Err: errors.New("connection refused"),
			},
			expectedID:       0,
			expectedTask:     nil,
			expectedErr:      &client.NetworkError{},
			expectedContains: "",
		},
		{
			name:          "API error from client",
			taskIDInput:   "1",
			prompt:        "Enter task ID:\n",
			getTaskResult: nil,
			getTaskErr: &client.APIError{
				StatusCode: 404,
				Message:    "Task not found",
			},
			expectedID:       0,
			expectedTask:     nil,
			expectedErr:      &client.APIError{},
			expectedContains: "",
		},
		{
			name:        "Task with empty description",
			taskIDInput: "5",
			prompt:      "Enter task ID:\n",
			getTaskResult: &client.Task{
				ID:          5,
				Description: "",
				Done:        false,
			},
			getTaskErr:       nil,
			expectedID:       5,
			expectedTask:     &client.Task{ID: 5, Description: "", Done: false},
			expectedErr:      nil,
			expectedContains: "Current task: '[ ] 5: '",
		},
		{
			name:        "Task with long description",
			taskIDInput: "10",
			prompt:      "Enter task ID:\n",
			getTaskResult: &client.Task{
				ID:          10,
				Description: "This is a very long task description that should be displayed completely",
				Done:        true,
			},
			getTaskErr:       nil,
			expectedID:       10,
			expectedTask:     &client.Task{ID: 10, Description: "This is a very long task description that should be displayed completely", Done: true},
			expectedErr:      nil,
			expectedContains: "Current task: '[✓] 10: This is a very long task description that should be displayed completely'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockClient := &MockTaskClient{
				getTaskResult: tc.getTaskResult,
				getTaskErr:    tc.getTaskErr,
			}
			cli := NewCLI(
				NewMockInputReader(tc.taskIDInput),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				&MockAuthManager{loadTokenResult: "mock-token"},
			)

			// ====Act====
			id, task, err := cli.promptForTaskWithDisplay(tc.prompt)

			// ====Assert====
			if tc.expectedErr != nil {
				assert.Error(t, err, "Expected an error but got nil")

				// Check if it's the expected error type
				switch tc.expectedErr.(type) {
				case *client.NetworkError:
					var netErr *client.NetworkError
					assert.ErrorAs(t, err, &netErr, "Expected NetworkError")
				case *client.APIError:
					var apiErr *client.APIError
					assert.ErrorAs(t, err, &apiErr, "Expected APIError")
				default:
					assert.ErrorIs(t, err, tc.expectedErr, "Expected specific error")
				}
			} else if tc.name == "Task not found" {
				// Special case: verify error is returned
				assert.Error(t, err, "Expected error but got nil")
			} else {
				assert.NoError(t, err, "Expected no error")
			}

			// Verify returned ID
			assert.Equal(t, tc.expectedID, id, "Task ID should match expected")

			// Verify returned task
			if tc.expectedTask != nil {
				assert.NotNil(t, task, "Task should not be nil")
				assert.Equal(t, tc.expectedTask.ID, task.ID, "Task ID should match")
				assert.Equal(t, tc.expectedTask.Description, task.Description, "Task description should match")
				assert.Equal(t, tc.expectedTask.Done, task.Done, "Task done status should match")
			} else {
				assert.Nil(t, task, "Task should be nil on error")
			}

			// Verify output contains expected message
			if tc.expectedContains != "" {
				assert.Contains(t, output.String(), tc.expectedContains, "Output should contain expected task display")
			}

			// Verify prompt was displayed
			assert.Contains(t, output.String(), tc.prompt, "Prompt should be displayed")
		})
	}
}
