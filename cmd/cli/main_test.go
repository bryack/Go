package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// TestCommand_isValid tests the isValid method for Command type
func TestCommand_isValid(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name     string
		command  Command
		expected bool
	}{
		{
			name:     "Valid command - add",
			command:  CommandAdd,
			expected: true,
		},
		{
			name:     "Valid command - status",
			command:  CommandStatus,
			expected: true,
		},
		{
			name:     "Valid command - list",
			command:  CommandList,
			expected: true,
		},
		{
			name:     "Valid command - process",
			command:  CommandProcess,
			expected: true,
		},
		{
			name:     "Valid command - clear",
			command:  CommandClear,
			expected: true,
		},
		{
			name:     "Valid command - help",
			command:  CommandHelp,
			expected: true,
		},
		{
			name:     "Valid command - exit",
			command:  CommandExit,
			expected: true,
		},
		{
			name:     "Valid command - update",
			command:  CommandUpdate,
			expected: true,
		},
		{
			name:     "Valid command - delete",
			command:  CommandDelete,
			expected: true,
		},
		{
			name:     "Valid command - login",
			command:  CommandLogin,
			expected: true,
		},
		{
			name:     "Valid command - register",
			command:  CommandRegister,
			expected: true,
		},
		{
			name:     "Valid command - logout",
			command:  CommandLogout,
			expected: true,
		},
		{
			name:     "Invalid command - empty string",
			command:  Command(""),
			expected: false,
		},
		{
			name:     "Invalid command - unknown",
			command:  Command("unknown"),
			expected: false,
		},
		{
			name:     "Invalid command - typo",
			command:  Command("addd"),
			expected: false,
		},
		{
			name:     "Invalid command - uppercase",
			command:  Command("ADD"),
			expected: false,
		},
		{
			name:     "Invalid command - mixed case",
			command:  Command("Add"),
			expected: false,
		},
		{
			name:     "Invalid command - with spaces",
			command:  Command("add task"),
			expected: false,
		},
		{
			name:     "Invalid command - special characters",
			command:  Command("add!"),
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			result := tc.command.isValid()

			// ====Assert====
			if result != tc.expected {
				t.Errorf("Expected isValid() to return %v for command %q, got %v", tc.expected, tc.command, result)
			}
		})
	}
}

// TestValidateCommand tests the validateCommand function
func TestValidateCommand(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name            string
		input           string
		expectedCommand Command
		expectedErr     error
	}{
		{
			name:            "Valid lowercase command",
			input:           "add",
			expectedCommand: CommandAdd,
			expectedErr:     nil,
		},
		{
			name:            "Valid uppercase command",
			input:           "LIST",
			expectedCommand: CommandList,
			expectedErr:     nil,
		},
		{
			name:            "Valid mixed case command",
			input:           "StAtUs",
			expectedCommand: CommandStatus,
			expectedErr:     nil,
		},
		{
			name:            "All commands are valid - help",
			input:           "help",
			expectedCommand: CommandHelp,
			expectedErr:     nil,
		},
		{
			name:            "All commands are valid - exit",
			input:           "exit",
			expectedCommand: CommandExit,
			expectedErr:     nil,
		},
		{
			name:            "All commands are valid - login",
			input:           "login",
			expectedCommand: CommandLogin,
			expectedErr:     nil,
		},
		{
			name:            "Invalid - empty string",
			input:           "",
			expectedCommand: "",
			expectedErr:     ErrInvalidCommand,
		},
		{
			name:            "Invalid - unknown command",
			input:           "unknown",
			expectedCommand: "",
			expectedErr:     ErrInvalidCommand,
		},
		{
			name:            "Invalid - typo",
			input:           "addd",
			expectedCommand: "",
			expectedErr:     ErrInvalidCommand,
		},
		{
			name:            "Invalid - partial command",
			input:           "ad",
			expectedCommand: "",
			expectedErr:     ErrInvalidCommand,
		},
		{
			name:            "Invalid - with spaces",
			input:           "add task",
			expectedCommand: "",
			expectedErr:     ErrInvalidCommand,
		},
		{
			name:            "Invalid - with leading space",
			input:           " add",
			expectedCommand: "",
			expectedErr:     ErrInvalidCommand,
		},
		{
			name:            "Invalid - special characters",
			input:           "add!",
			expectedCommand: "",
			expectedErr:     ErrInvalidCommand,
		},
		{
			name:            "Invalid - similar word",
			input:           "adding",
			expectedCommand: "",
			expectedErr:     ErrInvalidCommand,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			command, err := validateCommand(tc.input)

			// ====Assert====
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
			if command != tc.expectedCommand {
				t.Errorf("Expected command %q, got %q", tc.expectedCommand, command)
			}
		})
	}
}

// TestSuggestCommand tests the suggestCommand function
func TestSuggestCommand(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name            string
		input           string
		expectedCommand Command
	}{
		{
			name:            "Exact match returns command",
			input:           "add",
			expectedCommand: CommandAdd,
		},
		{
			name:            "Single letter prefix match",
			input:           "l",
			expectedCommand: CommandList,
		},
		{
			name:            "Partial prefix match",
			input:           "sta",
			expectedCommand: CommandStatus,
		},
		{
			name:            "Ambiguous prefix returns first match (login before logout)",
			input:           "log",
			expectedCommand: CommandLogin,
		},
		{
			name:            "Longer prefix disambiguates (logout)",
			input:           "logo",
			expectedCommand: CommandLogout,
		},
		{
			name:            "Empty string matches first command",
			input:           "",
			expectedCommand: CommandAdd,
		},
		{
			name:            "No match for typo",
			input:           "addd",
			expectedCommand: "",
		},
		{
			name:            "No match for unknown prefix",
			input:           "xyz",
			expectedCommand: "",
		},
		{
			name:            "No match for input with spaces",
			input:           "add task",
			expectedCommand: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			result := suggestCommand(tc.input)

			// ====Assert====
			if result != tc.expectedCommand {
				t.Errorf("Expected suggestion %q for input %q, got %q", tc.expectedCommand, tc.input, result)
			}
		})
	}
}

// TestCLI_RunLoop tests the RunLoop method
func TestCLI_RunLoop(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name               string
		inputs             []string
		expectedContains   []string
		expectedNotContain []string
	}{
		{
			name:   "Exit command terminates loop",
			inputs: []string{"exit"},
			expectedContains: []string{
				"=== Available Commands ===",
				"Enter command:",
				"👋 Bye!",
			},
		},
		{
			name:   "Logout command terminates loop",
			inputs: []string{"logout"},
			expectedContains: []string{
				"=== Available Commands ===",
				"Enter command:",
				"✅ Logged out successfully",
				"👋 Bye!",
			},
		},
		{
			name:   "Help command displays help menu",
			inputs: []string{"help", "exit"},
			expectedContains: []string{
				"=== Available Commands ===",
				"add      - Add a new task",
				"status   - Change task status",
				"list     - Show all tasks",
			},
		},
		{
			name:   "Invalid command shows suggestion",
			inputs: []string{"ad", "exit"},
			expectedContains: []string{
				"❌ Unknown command: 'ad', maybe you wanted: 'add'",
			},
		},
		{
			name:   "Invalid command without suggestion shows error",
			inputs: []string{"xyz", "exit"},
			expectedContains: []string{
				"Command validate error",
				"Type 'help' to see available commands",
			},
			expectedNotContain: []string{
				"maybe you wanted",
			},
		},
		{
			name:   "Process command shows unavailable message",
			inputs: []string{"process", "exit"},
			expectedContains: []string{
				"⚠️  Process command not available in client mode",
			},
		},
		{
			name:   "Multiple commands execute in sequence",
			inputs: []string{"help", "process", "exit"},
			expectedContains: []string{
				"=== Available Commands ===",
				"⚠️  Process command not available in client mode",
				"👋 Bye!",
			},
		},
		{
			name:   "Case insensitive commands work",
			inputs: []string{"HELP", "ExIt"},
			expectedContains: []string{
				"=== Available Commands ===",
				"👋 Bye!",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output := &bytes.Buffer{}
			mockClient := &MockTaskClient{}
			mockAuth := &MockAuthManager{
				loadTokenResult: "mock-token",
			}

			cli := NewCLI(
				NewMockInputReader(tc.inputs...),
				output,
				&Config{ServerURL: "http://localhost:8080"},
				mockClient,
				mockAuth,
			)

			// ====Act====
			cli.RunLoop()

			// ====Assert====
			result := output.String()

			for _, expected := range tc.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput: %s", expected, result)
				}
			}

			for _, notExpected := range tc.expectedNotContain {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected output NOT to contain %q, but it did.\nOutput: %s", notExpected, result)
				}
			}
		})
	}
}
