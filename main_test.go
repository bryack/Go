package main

import (
	"errors"
	"strings"
	"testing"
)

// TestReadInput tests the readInput function with various input scenarios.
// Covers valid input, whitespace handling, size limits, empty input, and edge cases.
func TestReadInput(t *testing.T) {
	// ====Arrange====
	var lenInput = 50

	testCases := []struct {
		name        string
		input       string
		lenInput    int
		expectedStr string
		expectedErr error
	}{
		{
			name:        "valid string",
			input:       "task 1\n",
			expectedStr: "task 1",
			lenInput:    lenInput,
			expectedErr: nil,
		},
		{
			name:        "string with spaces",
			input:       " task 1 \n",
			lenInput:    lenInput,
			expectedStr: "task 1",
			expectedErr: nil,
		},
		{
			name:        "empty string",
			input:       "\n",
			lenInput:    lenInput,
			expectedStr: "",
			expectedErr: ErrEmptyInput,
		},
		{
			name:        "string with spaces only",
			input:       "   \n",
			lenInput:    lenInput,
			expectedStr: "",
			expectedErr: ErrEmptyInput,
		},
		{
			name:        "string more than maxSize",
			input:       "string more than maxSize\n",
			lenInput:    5,
			expectedStr: "",
			expectedErr: ErrMaxSizeExceeded,
		},
		{
			name:        "string with special characters",
			input:       "#@`[]$%^*\n",
			lenInput:    lenInput,
			expectedStr: "#@`[]$%^*",
			expectedErr: nil,
		},
		{
			name:        "exactly at max size",
			input:       "12345\n",
			lenInput:    5,
			expectedStr: "12345",
			expectedErr: nil,
		},
		{
			name:        "input with carriage return",
			input:       "task\r\n",
			lenInput:    lenInput,
			expectedStr: "task",
			expectedErr: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			testInput := strings.NewReader(tc.input)
			str, err := readInput(testInput, tc.lenInput)

			// ====Assert====
			if str != tc.expectedStr {
				t.Errorf("Expected '%s', got '%s'", tc.expectedStr, str)
			}

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected '%v', got '%v'", tc.expectedErr, err)
			}
		})
	}
}
