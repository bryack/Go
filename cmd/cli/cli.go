package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"myproject/storage"
	"myproject/task"
	"myproject/validation"
	"strings"
)

const (
	maxCommandInputSize     = 10
	maxTaskIDInputSize      = 10
	maxDescriptionInputSize = 200
	maxStatusInputSize      = 10
)

var (
	ErrMaxSizeExceeded      = errors.New("input too long")
	ErrEmptyInput           = errors.New("empty input")
	ErrInvalidTaskId        = errors.New("invalid ID format")
	ErrInvalidCommand       = errors.New("invalid command")
	ErrInvalidStatus        = errors.New("invalid status")
	ErrDescUnchanged        = errors.New("description unchanged")
	ErrInvalidConfirmChoice = errors.New("invalid confirm choice")
)

// InputReader defines an interface for reading user input with size validation.
// Implementations must trim whitespace and enforce maximum input length constraints.
type InputReader interface {
	ReadInput(maxSize int) (string, error)
}

// CLI manages the command-line interface for task operations.
// Coordinates user input, task management, and database storage with proper error handling.
type CLI struct {
	input       InputReader
	output      io.Writer
	taskManager *task.TaskManager
	storage     *storage.DatabaseStorage
}

// NewCLI creates a new CLI instance with the provided dependencies.
// Returns a configured CLI ready to process user commands and manage tasks.
func NewCLI(input InputReader, output io.Writer, taskManager *task.TaskManager, storage *storage.DatabaseStorage) *CLI {
	return &CLI{
		input:       input,
		output:      output,
		taskManager: taskManager,
		storage:     storage,
	}
}

// ConsoleInputReader implements InputReader for reading from console input streams.
// Uses buffered reading to handle user input line-by-line with proper error handling.
type ConsoleInputReader struct {
	reader *bufio.Reader
}

// NewConsoleInputReader creates a new ConsoleInputReader from an io.Reader.
// Returns a reader configured with buffered input for efficient line reading.
func NewConsoleInputReader(reader io.Reader) *ConsoleInputReader {
	return &ConsoleInputReader{
		reader: bufio.NewReader(reader),
	}
}

// ReadInput reads a line from the input stream and validates its length.
// Returns trimmed input or errors for empty input, EOF, or size limit violations.
func (c *ConsoleInputReader) ReadInput(maxSize int) (string, error) {
	input, err := c.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return "", io.EOF
		}
		return "", err
	}

	input = strings.TrimSpace(input)
	if len(input) > maxSize {
		return "", ErrMaxSizeExceeded
	}

	if len(input) == 0 {
		return "", ErrEmptyInput
	}

	return input, nil
}

// promptForTaskID prompts the user for a task ID and validates the input.
// Returns the validated task ID or an error if input is invalid or exceeds size limits.
func (cli *CLI) promptForTaskID(prompt string) (id int, err error) {
	fmt.Fprint(cli.output, prompt)

	input, err := cli.input.ReadInput(maxTaskIDInputSize)
	if err != nil {
		return 0, err
	}

	return validation.ValidateTaskID(input)
}

// promptForTaskWithDisplay prompts for a task ID and displays the current task details.
// Returns the task ID, task object, and any errors from validation or task retrieval.
func (cli *CLI) promptForTaskWithDisplay(prompt string) (id int, t task.Task, err error) {
	id, err = cli.promptForTaskID(prompt)
	if err != nil {
		return 0, t, err
	}

	t, err = cli.taskManager.GetTaskByID(id)
	if err != nil {
		return 0, t, err
	}

	fmt.Fprintf(cli.output, "Current task: '%s'\n", task.FormatTask(t))

	return id, t, nil
}

// handleAddCommand prompts for a task description and adds a new task to the manager.
// Validates input length and description format before creating the task.
func (cli *CLI) handleAddCommand() error {
	fmt.Fprintln(cli.output, "Enter task description:")

	desc, err := cli.input.ReadInput(maxDescriptionInputSize)
	if err != nil {
		return fmt.Errorf("adding task: input failed: %w", err)
	}

	desc, err = validation.ValidateTaskDescription(desc)
	if err != nil {
		return fmt.Errorf("adding task: validation failed: %w", err)
	}

	newTask := cli.taskManager.AddTask(desc)
	id, err := cli.storage.CreateTask(newTask)
	if err != nil {
		return fmt.Errorf("adding task: creation failed: %w", err)
	}
	newTask.ID = id
	cli.taskManager.AddTaskWithID(newTask)

	fmt.Fprintf(cli.output, "✅ Task added (ID: %d)\n", id)
	return nil
}

// handleStatusCommand prompts for a task ID and new status, then updates the task.
// Accepts 'done' or 'undone' as valid status values with proper validation.
func (cli *CLI) handleStatusCommand() error {
	id, _, err := cli.promptForTaskWithDisplay("Enter task ID to change status:\n")
	if err != nil {
		return fmt.Errorf("updating status: task id validation failed: %w", err)
	}

	fmt.Fprint(cli.output, "Enter new status 'done' // 'undone'\n")
	str, err := cli.input.ReadInput(maxStatusInputSize)
	if err != nil {
		return fmt.Errorf("updating status: read status for task id %d failed: %w", id, err)
	}

	var done bool
	switch str {
	case "done":
		done = true
	case "undone":
		done = false
	default:
		return fmt.Errorf("updating status: invalid status: %q for task id %d: %w (must be 'done' or 'undone')", str, id, ErrInvalidStatus)
	}

	if err := cli.taskManager.UpdateTaskStatus(id, done); err != nil {
		return fmt.Errorf("updating status for task id %d failed: %w", id, err)
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) status is has changed\n", id)
	return nil
}

// handleClearCommand prompts for a task ID and clears its description.
// Validates the task exists before clearing the description field.
func (cli *CLI) handleClearCommand() error {
	id, _, err := cli.promptForTaskWithDisplay("Enter task ID you want to clear description\n")
	if err != nil {
		return fmt.Errorf("clearing task description: task id validation failed: %w", err)
	}

	if err = cli.taskManager.ClearDescription(id); err != nil {
		return fmt.Errorf("clearing task description for task id %d failed: %w", id, err)
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) description cleared!\n", id)
	return nil
}

// handleUpdateCommand prompts for a task ID and new description, then updates the task.
// Validates that the new description differs from the current one before updating.
func (cli *CLI) handleUpdateCommand() error {
	id, t, err := cli.promptForTaskWithDisplay("Enter task ID to update:\n")
	if err != nil {
		return fmt.Errorf("updating task description: task id validation failed: %w", err)
	}

	fmt.Fprint(cli.output, "Enter new description:\n")
	desc, err := cli.input.ReadInput(maxDescriptionInputSize)
	if err != nil {
		return fmt.Errorf("updating task description for task id %d: read description '%s' failed: %w", id, desc, err)
	}

	desc, err = validation.ValidateTaskDescription(desc)
	if err != nil {
		return fmt.Errorf("updating task description for task id %d: validate description '%s' failed: %w", id, desc, err)
	}

	if desc == t.Description {
		return fmt.Errorf("updating task description for task id %d: %w", id, ErrDescUnchanged)
	}

	if err = cli.taskManager.UpdateTaskDescription(id, desc); err != nil {
		return fmt.Errorf("updating task description for task id %d failed: %w", id, err)
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) updated\n", id)
	return nil
}

// handleDeleteCommand prompts for a task ID and confirmation, then deletes the task.
// Requires explicit 'y' confirmation to proceed with deletion, 'n' cancels the operation.
func (cli *CLI) handleDeleteCommand() error {
	id, _, err := cli.promptForTaskWithDisplay("Enter task ID to delete task:\n")
	if err != nil {
		return fmt.Errorf("deleting task: id validation failed: %w", err)
	}

	fmt.Fprintln(cli.output, "Enter y/N:")
	str, err := cli.input.ReadInput(10)
	if err != nil {
		return fmt.Errorf("deleting task id %d: read confirmation failed: %w", id, err)
	}
	str = strings.ToLower(str)

	switch str {
	case "y":
		if err = cli.taskManager.DeleteTask(id); err != nil {
			return fmt.Errorf("deleting task id %d failed: %w", id, err)
		}
		fmt.Fprintf(cli.output, "✅ Task (ID: %d) deleted\n", id)
		return nil
	case "n":
		fmt.Fprintln(cli.output, "Deletion canceled")
		return nil
	default:
		return fmt.Errorf("deleting task id %d: %q: %w (must be 'y' or 'n')", id, str, ErrInvalidConfirmChoice)
	}
}

// showHelp displays the list of available commands and their descriptions.
// Outputs a formatted help menu to the configured output writer.
func (cli *CLI) showHelp() {
	fmt.Fprintln(cli.output, "\n=== Available Commands ===")
	fmt.Fprintln(cli.output, "add     - Add a new task")
	fmt.Fprintln(cli.output, "status  - Change task status")
	fmt.Fprintln(cli.output, "list    - Show all tasks")
	fmt.Fprintln(cli.output, "process - Process all tasks in parallel")
	fmt.Fprintln(cli.output, "clear   - Clear task description")
	fmt.Fprintln(cli.output, "update  - Update task description")
	fmt.Fprintln(cli.output, "delete  - Delete task")
	fmt.Fprintln(cli.output, "help    - Show this help")
	fmt.Fprintln(cli.output, "exit    - Save and exit")
	fmt.Fprintln(cli.output, "=========================")
}

// handleError formats and displays error messages with context information.
// Provides user-friendly error messages and handles EOF as input interruption.
func (cli *CLI) handleError(err error, context string) {
	if errors.Is(err, io.EOF) {
		fmt.Fprintf(cli.output, "%s: input interrupted by user\n", context)
		return
	}

	fmt.Fprintf(cli.output, "%s: %v\n", context, err)
}

// RunLoop starts the main command processing loop for the CLI application.
// Continuously reads commands, executes handlers, and manages application lifecycle until exit.
func (cli *CLI) RunLoop() {
	fmt.Fprintln(cli.output, "🚀 Task Manager Started!")
	fmt.Fprintln(cli.output, "📁 Database storage initialized")
	cli.showHelp()
	for {
		fmt.Fprint(cli.output, "\nEnter command: ")
		input, err := cli.input.ReadInput(maxCommandInputSize)
		if err != nil {
			cli.handleError(err, "Input error")
			continue
		}

		cmd, err := validateCommand(input)
		if err != nil {
			suggestion := suggestCommand(input)
			if suggestion != "" {
				fmt.Fprintf(cli.output, "❌ Unknown command: '%s', maybe you wanted: '%s'\n", input, suggestion)
			} else {
				cli.handleError(err, "Command validate error")
				fmt.Fprintln(cli.output, "Type 'help' to see available commands")
			}
			continue
		}

		switch Command(cmd) {
		case CommandAdd:
			if err := cli.handleAddCommand(); err != nil {
				cli.handleError(err, "Add command error")
			}

		case CommandStatus:
			if err := cli.handleStatusCommand(); err != nil {
				cli.handleError(err, "Status command error")
			}

		case CommandList:
			tasks, err := cli.storage.LoadTasks()
			if err != nil {
				cli.handleError(err, "Failed to load tasks")
				break
			}

			cli.taskManager.SetTasks(tasks)

			if err := cli.taskManager.PrintTasks(); err != nil {
				cli.handleError(err, "Print tasks error")
			}

		case CommandProcess:
			cli.taskManager.ProcessTasks()

		case CommandClear:
			if err := cli.handleClearCommand(); err != nil {
				cli.handleError(err, "Clear command error")
			}

		case CommandDelete:
			if err := cli.handleDeleteCommand(); err != nil {
				cli.handleError(err, "Delete command error")
			}

		case CommandHelp:
			cli.showHelp()

		case CommandExit:
			if err := cli.storage.SaveTasks(cli.taskManager.GetTasks()); err != nil {
				cli.handleError(err, "Save error")
			} else {
				fmt.Fprintln(cli.output, "Tasks saved successfully!")
			}
			fmt.Fprintln(cli.output, "👋 Bye!")
			return

		case CommandUpdate:
			if err := cli.handleUpdateCommand(); err != nil {
				cli.handleError(err, "Update command error")
			}
		}
	}
}
