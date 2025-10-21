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

var (
	ErrMaxSizeExceeded = errors.New("input too long")
	ErrEmptyInput      = errors.New("empty input")
	ErrInvalidTaskId   = errors.New("invalid ID format")
	ErrInvalidCommand  = errors.New("invalid command")
)

type InputReader interface {
	ReadInput(maxSize int) (string, error)
}

type CLI struct {
	input       InputReader
	output      io.Writer
	taskManager *task.TaskManager
	storage     *storage.DatabaseStorage
}

func NewCLI(input InputReader, output io.Writer, taskManager *task.TaskManager, storage *storage.DatabaseStorage) *CLI {
	return &CLI{
		input:       input,
		output:      output,
		taskManager: taskManager,
		storage:     storage,
	}
}

type ConsoleInputReader struct {
	reader io.Reader
}

func NewConsoleInputReader(reader io.Reader) *ConsoleInputReader {
	return &ConsoleInputReader{
		reader: reader,
	}
}

func (c *ConsoleInputReader) ReadInput(maxSize int) (string, error) {
	bufReader := bufio.NewReader(c.reader)
	input, err := bufReader.ReadString('\n')
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

func (cli *CLI) promptForTask(prompt string) (id int, t task.Task, err error) {
	fmt.Fprint(cli.output, prompt)

	input, err := cli.input.ReadInput(10)
	if err != nil {
		return 0, t, err
	}

	id, err = validation.ValidateTaskID(input)
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

func (cli *CLI) handleAddCommand() error {
	fmt.Fprintln(cli.output, "Enter task description:\n")

	desc, err := cli.input.ReadInput(200)
	if err != nil {
		return fmt.Errorf("adding task: input failed: %w", err)
	}

	desc, err = validation.ValidateTaskDescription(desc)
	if err != nil {
		return fmt.Errorf("adding task: validation failed: %w", err)
	}

	id := cli.taskManager.AddTask(desc)
	fmt.Fprintf(cli.output, "✅ Task added (ID: %d)\n", id)
	return nil
}

func (cli *CLI) handleStatusCommand() error {
	id, _, err := cli.promptForTask("Enter task ID to change status:\n")
	if err != nil {
		return fmt.Errorf("updating status: task id %s validation failed: %w", id, err)
	}

	fmt.Fprint(cli.output, "Enter new status 'done' // 'undone'\n")
	str, err := cli.input.ReadInput(10)
	if err != nil {
		return fmt.Errorf("updating status: read status '%q' for task id %s failed: %w", str, id, err)
	}

	var done bool
	switch str {
	case "done":
		done = true
	case "undone":
		done = false
	default:
		return fmt.Errorf("updating status: invalid status: '%q' for task id %d; must be 'done' or 'undone'", str, id)
	}

	if err := cli.taskManager.UpdateTaskStatus(id, done); err != nil {
		return fmt.Errorf("updating status for task id %d failed: %w", id, err)
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) status is has changed\n", id)
	return nil
}

func (cli *CLI) handleClearCommand() error {
	id, _, err := cli.promptForTask("Enter task ID you want to clear description\n")
	if err != nil {
		return fmt.Errorf("clearing task description: task id %s validation failed: %w", id, err)
	}

	if err = cli.taskManager.ClearDescription(id); err != nil {
		return fmt.Errorf("clearing task description for task id %d failed: %w", id, err)
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) description cleared!\n", id)
	return nil
}

func (cli *CLI) handleUpdateCommand() error {
	id, t, err := cli.promptForTask("Enter task ID to update:\n")
	if err != nil {
		return fmt.Errorf("updating task description: task id %s validation failed: %w", id, err)
	}

	fmt.Fprint(cli.output, "Enter new description:\n")
	desc, err := cli.input.ReadInput(200)
	if err != nil {
		return fmt.Errorf("updating task description for task id %d: read description '%s' failed: %w", id, desc, err)
	}

	desc, err = validation.ValidateTaskDescription(desc)
	if err != nil {
		return fmt.Errorf("updating task description for task id %d: validate description '%s' failed: %w", id, desc, err)
	}

	if desc == t.Description {
		return fmt.Errorf("updating task description for task id %d: description unchanged", id)
	}

	if err = cli.taskManager.UpdateTaskDescription(id, desc); err != nil {
		return fmt.Errorf("updating task description for task id %d failed: %w", id, err)
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) updated\n", id)
	return nil
}

func (cli *CLI) handleLoadCommand() error {
	loadedTasks, err := cli.storage.LoadTasks()
	if err != nil {
		return fmt.Errorf("loading tasks failed: %w", err)
	}

	cli.taskManager.SetTasks(loadedTasks)
	fmt.Fprintf(cli.output, "✅ %d tasks loaded successfully!\n", len(loadedTasks))

	return nil
}

func (cli *CLI) handleDeleteCommand() error {
	id, _, err := cli.promptForTask("Enter task ID to delete task:\n")
	if err != nil {
		return fmt.Errorf("deliting task id %s validation failed: %w", id, err)
	}

	fmt.Fprintln(cli.output, "Enter y/N:\n")
	str, err := cli.input.ReadInput(10)
	if err != nil {
		return fmt.Errorf("deliting task id %s: read confirmation failed: %w", id, err)
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
		fmt.Fprintln(cli.output, "Deletion canceled\n")
		return nil
	default:
		return fmt.Errorf("deleting task id %d: invalid choice: %q; must be 'y' or 'n'", str)
	}
}

func (cli *CLI) showHelp() {
	fmt.Fprintln(cli.output, "\n=== Available Commands ===")
	fmt.Fprintln(cli.output, "add     - Add a new task")
	fmt.Fprintln(cli.output, "status  - Change task status")
	fmt.Fprintln(cli.output, "list    - Show all tasks")
	fmt.Fprintln(cli.output, "process - Process all tasks in parallel")
	fmt.Fprintln(cli.output, "load    - Load tasks from file")
	fmt.Fprintln(cli.output, "clear   - Clear task description")
	fmt.Fprintln(cli.output, "update  - Update task description")
	fmt.Fprintln(cli.output, "delete  - Delete task")
	fmt.Fprintln(cli.output, "help    - Show this help")
	fmt.Fprintln(cli.output, "exit    - Save and exit")
	fmt.Fprintln(cli.output, "=========================")
}

func (cli *CLI) handleError(err error, context string) {
	switch {
	case errors.Is(err, io.EOF):
		fmt.Fprintf(cli.output, "%s: input interrupted by user", context)
	// case errors.Is(err, ErrMaxSizeExceeded):
	// 	fmt.Printf("%s: input exceeds %d characters\n", context, maxInputSize)
	case errors.Is(err, ErrEmptyInput):
		fmt.Fprintf(cli.output, "%s: empty input provided\n", context)
	case errors.Is(err, task.ErrTaskNotFound):
		fmt.Fprintf(cli.output, "%s: task not found\n", context)
	case errors.Is(err, storage.ErrConversionTask):
		fmt.Fprintf(cli.output, "%s: failed to convert tasks\n", context)
	case errors.Is(err, storage.ErrFailedWriteFile):
		fmt.Fprintf(cli.output, "%s: failed to write file\n", context)
	case errors.Is(err, storage.ErrFileNotFound):
		fmt.Fprintf(cli.output, "%s: file not found\n", context)
	case errors.Is(err, storage.ErrParseJson):
		fmt.Fprintf(cli.output, "%s: JSON parsing error\n", context)
	case errors.Is(err, ErrInvalidTaskId):
		fmt.Fprintf(cli.output, "%s: ID must be a positive number (greater than 0)\n", context)
	case errors.Is(err, task.ErrPrintTask):
		fmt.Fprintf(cli.output, "%s: failed to print tasks\n", context)
	case errors.Is(err, ErrInvalidCommand):
		fmt.Fprintf(cli.output, "%s: failed to assume command\n", context)
	default:
		fmt.Fprintf(cli.output, "%s: %v\n", context, err)
	}
}

func (cli *CLI) RunLoop() {
	fmt.Fprintln(cli.output, "🚀 Task Manager Started!")
	fmt.Fprintln(cli.output, "📁 Database storage initialized")
	cli.showHelp()
	for {
		fmt.Fprint(cli.output, "\nEnter command: ")
		input, err := cli.input.ReadInput(10)
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
			if err := cli.taskManager.PrintTasks(); err != nil {
				cli.handleError(err, "Print tasks error")
			}

		case CommandProcess:
			cli.taskManager.ProcessTasks()

		case CommandLoad:
			if err := cli.handleLoadCommand(); err != nil {
				cli.handleError(err, "Load command error")
			}

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
