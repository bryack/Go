package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"myproject/task"
	"myproject/validation"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			c := NewConsoleInputReader(testInput)
			str, err := c.ReadInput(tc.lenInput)

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

func TestCLI_PromptForTaskID(t *testing.T) {
	// ====Arrange====
	prompt := "Enter task ID:\n"
	testCases := []struct {
		name        string
		input       string
		expectedID  int
		expectedErr error
	}{
		// Valid inputs
		{
			name:        "valid id",
			input:       "1\n",
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name:        "valid id with spaces",
			input:       " 1 \n",
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name:        "input at max size limit",
			input:       "1234567890\n",
			expectedID:  1234567890,
			expectedErr: nil,
		},
		// Input size issues
		{
			name:        "input over max size",
			input:       "11111111111\n",
			expectedID:  0,
			expectedErr: ErrMaxSizeExceeded,
		},
		// Invalid format
		{
			name:        "empty input",
			input:       "\n",
			expectedID:  0,
			expectedErr: ErrEmptyInput,
		},
		{
			name:        "input with spaces only",
			input:       "      \n",
			expectedID:  0,
			expectedErr: ErrEmptyInput,
		},
		{
			name:        "zero ID",
			input:       "0\n",
			expectedID:  0,
			expectedErr: validation.ErrInvalidTaskID,
		},
		{
			name:        "negative ID",
			input:       "-1\n",
			expectedID:  0,
			expectedErr: validation.ErrInvalidTaskID,
		},
		{
			name:        "decimal number",
			input:       "1.5\n",
			expectedID:  0,
			expectedErr: validation.ErrInvalidTaskID,
		},
		{
			name:        "input with special characters",
			input:       "#@`[]$%^*\n",
			expectedID:  0,
			expectedErr: validation.ErrInvalidTaskID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeInput := strings.NewReader(tc.input)
			output := &bytes.Buffer{}
			cli := NewCLI(
				NewConsoleInputReader(fakeInput),
				output,
				nil,
				nil,
			)

			// ==== ACT ====
			id, err := cli.promptForTaskID(prompt)

			// === ASSERT ===
			assert.Equal(t, tc.expectedID, id)
			assert.ErrorIs(t, tc.expectedErr, err)
			assert.Equal(t, prompt, output.String())
		})
	}
}

func TestCLI_PromptForTaskWithDisplay(t *testing.T) {
	// ====Arrange====
	prompt := "Enter task ID:\n"
	testCases := []struct {
		name           string
		input          string
		initialTasks   []task.Task
		expectedID     int
		expectedTask   task.Task
		expectedErr    error
		expectedPrompt string
	}{
		// Valid inputs
		{
			name:           "existent task in one-task list",
			input:          "1\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedID:     1,
			expectedTask:   task.Task{ID: 1, Description: "task 1"},
			expectedErr:    nil,
			expectedPrompt: "Enter task ID:\nCurrent task: '[  ] ID: 1, Description: task 1'\n",
		},
		{
			name:           "existent task (valid id with spaces) in multiple tasks list",
			input:          " 3 \n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedID:     3,
			expectedTask:   task.Task{ID: 3, Description: "task 3", Done: true},
			expectedErr:    nil,
			expectedPrompt: "Enter task ID:\nCurrent task: '[✓ ] ID: 3, Description: task 3'\n",
		},
		{
			name:           "Very large ID",
			input:          "9999999999\n",
			initialTasks:   []task.Task{{ID: 9999999999, Description: "task 9999999999", Done: false}},
			expectedID:     9999999999,
			expectedTask:   task.Task{ID: 9999999999, Description: "task 9999999999"},
			expectedErr:    nil,
			expectedPrompt: "Enter task ID:\nCurrent task: '[  ] ID: 9999999999, Description: task 9999999999'\n",
		},
		{
			name:           "Existent task with very long description",
			input:          "3\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.", Done: true}},
			expectedID:     3,
			expectedTask:   task.Task{ID: 3, Description: "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.", Done: true},
			expectedErr:    nil,
			expectedPrompt: "Enter task ID:\nCurrent task: '[✓ ] ID: 3, Description: Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.'\n",
		},
		{
			name:           "Existent task with empty description",
			input:          "3\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "", Done: true}},
			expectedID:     3,
			expectedTask:   task.Task{ID: 3, Done: true},
			expectedErr:    nil,
			expectedPrompt: "Enter task ID:\nCurrent task: '[✓ ] ID: 3, Description: '\n",
		},
		{
			name:           "Existent task with Unicode and special characters",
			input:          "3\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "Buy 🍞 and 🥛#@`[]$%^*", Done: true}},
			expectedID:     3,
			expectedTask:   task.Task{ID: 3, Description: "Buy 🍞 and 🥛#@`[]$%^*", Done: true},
			expectedErr:    nil,
			expectedPrompt: "Enter task ID:\nCurrent task: '[✓ ] ID: 3, Description: Buy 🍞 and 🥛#@`[]$%^*'\n",
		},
		// Input size issues
		{
			name:           "input over max size",
			input:          "11111111111\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedID:     0,
			expectedTask:   task.Task{},
			expectedErr:    ErrMaxSizeExceeded,
			expectedPrompt: prompt,
		},
		// Invalid format
		{
			name:           "reader with no content",
			input:          "",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedID:     0,
			expectedTask:   task.Task{},
			expectedErr:    io.EOF,
			expectedPrompt: prompt,
		},
		{
			name:           "empty input",
			input:          "\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedID:     0,
			expectedTask:   task.Task{},
			expectedErr:    ErrEmptyInput,
			expectedPrompt: prompt,
		},
		{
			name:           "input with spaces only",
			input:          "      \n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedID:     0,
			expectedTask:   task.Task{},
			expectedErr:    ErrEmptyInput,
			expectedPrompt: prompt,
		},
		{
			name:           "zero ID",
			input:          "0\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedID:     0,
			expectedTask:   task.Task{},
			expectedErr:    validation.ErrInvalidTaskID,
			expectedPrompt: prompt,
		},
		{
			name:           "negative ID",
			input:          "-1\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedID:     0,
			expectedTask:   task.Task{},
			expectedErr:    validation.ErrInvalidTaskID,
			expectedPrompt: prompt,
		},
		{
			name:           "decimal number",
			input:          "1.5\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedID:     0,
			expectedTask:   task.Task{},
			expectedErr:    validation.ErrInvalidTaskID,
			expectedPrompt: prompt,
		},
		{
			name:           "input with special characters",
			input:          "#@`[]$%^*\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedID:     0,
			expectedTask:   task.Task{},
			expectedErr:    validation.ErrInvalidTaskID,
			expectedPrompt: prompt,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeInput := strings.NewReader(tc.input)
			output := &bytes.Buffer{}
			taskManager := task.NewTaskManager(output)
			cli := NewCLI(
				NewConsoleInputReader(fakeInput),
				output,
				taskManager,
				nil,
			)
			cli.taskManager.SetTasks(tc.initialTasks)

			// ==== ACT ====
			id, task, err := cli.promptForTaskWithDisplay(prompt)

			// === ASSERT ===
			assert.Equal(t, tc.expectedID, id)
			assert.Equal(t, tc.expectedTask, task)
			assert.ErrorIs(t, tc.expectedErr, err)
			assert.Equal(t, tc.expectedPrompt, output.String())
		})
	}
}

func TestCLI_HandleAddCommand(t *testing.T) {
	// ====Arrange====
	prompt := "Enter task description:\n"
	testCases := []struct {
		name           string
		input          string
		initialTasks   []task.Task
		expectedPrompt string
		expectedTasks  []task.Task
		expectedErr    error
	}{
		// Valid inputs
		{
			name:           "Add task to empty list",
			input:          "task 1\n",
			initialTasks:   []task.Task{},
			expectedPrompt: "Enter task description:\n✅ Task added (ID: 1)\n",
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:    nil,
		},
		{
			name:           "Add task (description with spaces) to non-empty list",
			input:          " task 2 \n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedPrompt: "Enter task description:\n✅ Task added (ID: 2)\n",
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}},
			expectedErr:    nil,
		},
		{
			name:           "Add long description",
			input:          "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec qu\n",
			initialTasks:   []task.Task{},
			expectedPrompt: "Enter task description:\n✅ Task added (ID: 1)\n",
			expectedTasks:  []task.Task{{ID: 1, Description: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec qu", Done: false}},
			expectedErr:    nil,
		},
		{
			name:           "Add task with Unicode and special characters",
			input:          "Купить 🥛 & 🍞 в магазине \"Пятёрочка\"\n",
			initialTasks:   []task.Task{},
			expectedPrompt: "Enter task description:\n✅ Task added (ID: 1)\n",
			expectedTasks:  []task.Task{{ID: 1, Description: "Купить 🥛 & 🍞 в магазине \"Пятёрочка\"", Done: false}},
			expectedErr:    nil,
		},
		// Input size issues
		{
			name:           "Add description over max size",
			input:          "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam\n",
			initialTasks:   []task.Task{},
			expectedPrompt: prompt,
			expectedTasks:  []task.Task{},
			expectedErr:    ErrMaxSizeExceeded,
		},
		// Invalid format
		{
			name:           "Reader with no content",
			input:          "",
			initialTasks:   []task.Task{},
			expectedPrompt: prompt,
			expectedTasks:  []task.Task{},
			expectedErr:    io.EOF,
		},
		{
			name:           "Add empty description",
			input:          "\n",
			initialTasks:   []task.Task{},
			expectedPrompt: prompt,
			expectedTasks:  []task.Task{},
			expectedErr:    ErrEmptyInput,
		},
		{
			name:           "Add description with spaces only",
			input:          "         \n",
			initialTasks:   []task.Task{},
			expectedPrompt: prompt,
			expectedTasks:  []task.Task{},
			expectedErr:    ErrEmptyInput,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeInput := strings.NewReader(tc.input)
			output := &bytes.Buffer{}
			taskManager := task.NewTaskManager(output)

			cli := NewCLI(
				NewConsoleInputReader(fakeInput),
				output,
				taskManager,
				nil,
			)
			cli.taskManager.SetTasks(tc.initialTasks)

			// ==== ACT ====
			err := cli.handleAddCommand()

			// === ASSERT ===
			assert.Equal(t, tc.expectedTasks, cli.taskManager.GetTasks())
			assert.Equal(t, tc.expectedPrompt, output.String())
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCLI_HandleStatusCommand(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name           string
		input          string
		initialTasks   []task.Task
		expectedTasks  []task.Task
		expectedPrompt string
		expectedErr    error
	}{
		// Valid inputs
		{
			name:           "Mark task done in one-task list",
			input:          "1\ndone\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: true}},
			expectedPrompt: "Enter task ID to change status:\nCurrent task: '[  ] ID: 1, Description: task 1'\nEnter new status 'done' // 'undone'\n✅ Task (ID: 1) status is has changed\n",
			expectedErr:    nil,
		},
		{
			name:           "Mark task undone in one-task list",
			input:          "1\nundone\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: true}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedPrompt: "Enter task ID to change status:\nCurrent task: '[✓ ] ID: 1, Description: task 1'\nEnter new status 'done' // 'undone'\n✅ Task (ID: 1) status is has changed\n",
			expectedErr:    nil,
		},
		{
			name:           "Mark specific task done in multiple tasks",
			input:          "3\ndone\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}, {ID: 4, Description: "task 4", Done: false}},
			expectedPrompt: "Enter task ID to change status:\nCurrent task: '[  ] ID: 3, Description: task 3'\nEnter new status 'done' // 'undone'\n✅ Task (ID: 3) status is has changed\n",
			expectedErr:    nil,
		},
		{
			name:           "Mark specific task undone in multiple tasks",
			input:          "4\nundone\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: true}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedPrompt: "Enter task ID to change status:\nCurrent task: '[✓ ] ID: 4, Description: task 4'\nEnter new status 'done' // 'undone'\n✅ Task (ID: 4) status is has changed\n",
			expectedErr:    nil,
		},
		{
			name:           "Mark already completed task",
			input:          "1\ndone\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: true}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: true}},
			expectedPrompt: "Enter task ID to change status:\nCurrent task: '[✓ ] ID: 1, Description: task 1'\nEnter new status 'done' // 'undone'\n✅ Task (ID: 1) status is has changed\n",
			expectedErr:    nil,
		},
		{
			name:           "Mark incompleted task undone",
			input:          "1\nundone\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedPrompt: "Enter task ID to change status:\nCurrent task: '[  ] ID: 1, Description: task 1'\nEnter new status 'done' // 'undone'\n✅ Task (ID: 1) status is has changed\n",
			expectedErr:    nil,
		},
		// Invalid format
		{
			name:           "Mark task done in empty list",
			input:          "1\n",
			initialTasks:   []task.Task{},
			expectedTasks:  []task.Task{},
			expectedPrompt: "Enter task ID to change status:\n",
			expectedErr:    task.ErrTaskNotFound,
		},
		{
			name:           "Mark non-existent task done",
			input:          "8\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedPrompt: "Enter task ID to change status:\n",
			expectedErr:    task.ErrTaskNotFound,
		},
		{
			name:           "Mark non-existent task undone",
			input:          "8\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedPrompt: "Enter task ID to change status:\n",
			expectedErr:    task.ErrTaskNotFound,
		},
		{
			name:           "Mark task with negative ID",
			input:          "-1\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedPrompt: "Enter task ID to change status:\n",
			expectedErr:    validation.ErrInvalidTaskID,
		},
		{
			name:           "Mark task with empty input",
			input:          "\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedPrompt: "Enter task ID to change status:\n",
			expectedErr:    ErrEmptyInput,
		},
		{
			name:           "Invalid status input",
			input:          "1\ninvalid\n",
			initialTasks:   []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks:  []task.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedPrompt: "Enter task ID to change status:\nCurrent task: '[  ] ID: 1, Description: task 1'\nEnter new status 'done' // 'undone'\n",
			expectedErr:    ErrInvalidStatus,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeInput := strings.NewReader(tc.input)
			output := &bytes.Buffer{}
			taskManager := task.NewTaskManager(output)
			cli := NewCLI(
				NewConsoleInputReader(fakeInput),
				output,
				taskManager,
				nil,
			)
			cli.taskManager.SetTasks(tc.initialTasks)

			// ==== ACT ====
			err := cli.handleStatusCommand()

			// === ASSERT ===
			assert.Equal(t, tc.expectedTasks, cli.taskManager.GetTasks())
			assert.Equal(t, tc.expectedPrompt, output.String())
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
