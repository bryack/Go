package main

import (
	"bytes"
	"errors"
	"myproject/cmd/cli/client"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCLI_HandleError_NetworkError tests that NetworkError is displayed with user-friendly message
func TestCLI_HandleError_NetworkError(t *testing.T) {
	output := &bytes.Buffer{}
	cli := NewCLI(
		nil,
		output,
		nil,
		nil,
		nil,
	)

	netErr := &client.NetworkError{
		URL: "http://localhost:8080",
		Err: errors.New("connection refused"),
	}

	cli.handleError(netErr, "Test operation")

	expected := "❌ Test operation: Cannot connect to server at http://localhost:8080\n   Please check that the server is running and the URL is correct\n"
	assert.Equal(t, expected, output.String())
}

// TestCLI_HandleError_APIError tests that APIError is displayed with server error message
func TestCLI_HandleError_APIError(t *testing.T) {
	testCases := []struct {
		name           string
		apiError       *client.APIError
		context        string
		expectedOutput string
	}{
		{
			name: "404 Not Found",
			apiError: &client.APIError{
				StatusCode: 404,
				Message:    "Task not found",
			},
			context:        "Get task",
			expectedOutput: "❌ Get task: Task not found\n",
		},
		{
			name: "400 Bad Request",
			apiError: &client.APIError{
				StatusCode: 400,
				Message:    "Invalid task description",
			},
			context:        "Create task",
			expectedOutput: "❌ Create task: Invalid task description\n",
		},
		{
			name: "500 Internal Server Error",
			apiError: &client.APIError{
				StatusCode: 500,
				Message:    "Server error (500), please try again later",
			},
			context:        "Update task",
			expectedOutput: "❌ Update task: Server error (500), please try again later\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			cli := NewCLI(
				nil,
				output,
				nil,
				nil,
				nil,
			)

			cli.handleError(tc.apiError, tc.context)

			assert.Equal(t, tc.expectedOutput, output.String())
		})
	}
}

// TestCLI_HandleError_GenericError tests that generic errors are displayed with standard format
func TestCLI_HandleError_GenericError(t *testing.T) {
	output := &bytes.Buffer{}
	cli := NewCLI(
		nil,
		output,
		nil,
		nil,
		nil,
	)

	genericErr := errors.New("some generic error")

	cli.handleError(genericErr, "Generic operation")

	expected := "Generic operation: some generic error\n"
	assert.Equal(t, expected, output.String())
}

// TestCLI_HandleError_WrappedNetworkError tests that wrapped NetworkError is properly detected
func TestCLI_HandleError_WrappedNetworkError(t *testing.T) {
	output := &bytes.Buffer{}
	cli := NewCLI(
		nil,
		output,
		nil,
		nil,
		nil,
	)

	netErr := &client.NetworkError{
		URL: "http://localhost:8080",
		Err: errors.New("connection timeout"),
	}

	// Wrap the error to simulate real-world scenario
	wrappedErr := errors.Join(netErr)

	cli.handleError(wrappedErr, "List tasks")

	// Should still detect and format as NetworkError
	assert.Contains(t, output.String(), "Cannot connect to server at http://localhost:8080")
	assert.Contains(t, output.String(), "Please check that the server is running")
}
