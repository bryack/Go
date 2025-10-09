package main

import (
	"errors"
	"fmt"
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

func TestIsValidCommand(t *testing.T) {
	copyValidCommands := make([]Command, len(validCommands))
	copy(copyValidCommands, validCommands)

	for _, validCmd := range copyValidCommands {
		t.Run(fmt.Sprintf("Valid command %s", validCmd), func(t *testing.T) {
			if !validCmd.isValid() {
				t.Errorf("Command '%s' should be valid but isValid() returned false", validCmd)
			}
		})
	}

	invalidCommands := []Command{
		Command(""),
		Command("unknown"),
		Command("ADD"),
		Command("add task"),
		Command("#@`[]$%^*"),
		Command("    "),
	}

	for _, invalidCmd := range invalidCommands {
		t.Run(fmt.Sprintf("Invalid command: %s", invalidCmd), func(t *testing.T) {
			if invalidCmd.isValid() {
				t.Errorf("Command '%s' should be invalid but isValid() returned true", invalidCmd)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedCmd Command
		expectedErr error
	}{
		{
			name:        "Valid command",
			input:       "add",
			expectedCmd: CommandAdd,
			expectedErr: nil,
		},
		{
			name:        "Mixed case",
			input:       "ADd",
			expectedCmd: CommandAdd,
			expectedErr: nil,
		},
		{
			name:        "Invalid command",
			input:       "unknown",
			expectedCmd: "",
			expectedErr: ErrInvalidCommand,
		},
		{
			name:        "Empty command",
			input:       "",
			expectedCmd: "",
			expectedErr: ErrInvalidCommand,
		},
		{
			name:        "Empty command with spaces",
			input:       "     ",
			expectedCmd: "",
			expectedErr: ErrInvalidCommand,
		},
		{
			name:        "Command with special characters",
			input:       "#@`[]$%^*",
			expectedCmd: "",
			expectedErr: ErrInvalidCommand,
		},
		{
			name:        "Command with Unicode characters",
			input:       "✅",
			expectedCmd: "",
			expectedErr: ErrInvalidCommand,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := validateCommand(tc.input)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected %v, got %v", tc.expectedErr, err)
			}

			if cmd != tc.expectedCmd {
				t.Errorf("Expected %s, got %s", tc.expectedCmd, cmd)
			}
		})
	}
}
