package main

import (
	"bufio"
	"fmt"
	"io"
	"myproject/storage"
	"myproject/task"
	"myproject/validation"
	"strings"
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

func (cli *CLI) handleAddCommand() error {
	fmt.Fprintln(cli.output, "Enter task description:")

	desc, err := cli.input.ReadInput(200)
	if err != nil {
		return err
	}

	id := cli.taskManager.AddTask(desc)
	fmt.Fprintf(cli.output, "✅ Task added (ID: %d)\n", id)
	return nil
}

func (cli *CLI) promptForTaskID(prompt string) (int, error) {
	fmt.Fprintln(cli.output, prompt)

	input, err := cli.input.ReadInput(10)
	if err != nil {
		return 0, err
	}

	return validation.ValidateTaskID(input)
}

func (cli *CLI) handleStatusCommand() error {
	id, err := cli.promptForTaskID("Enter task ID to change status:")
	if err != nil {
		return err
	}

	t, err := cli.taskManager.GetTaskByID(id)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.output, "Current task: '%s'\n", task.FormatTask(t))

	fmt.Fprintln(cli.output, "Enter new status 'done' // 'undone'")
	str, err := cli.input.ReadInput(10)
	if err != nil {
		return err
	}

	var done bool
	switch str {
	case "done":
		done = true
	case "undone":
		done = false
	default:
		return fmt.Errorf("invalid status: %q; must be 'done' or 'undone'", str)
	}

	if err := cli.taskManager.UpdateTaskStatus(id, done); err != nil {
		return err
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) status is has changed", id)
	return nil
}

func (cli *CLI) handleClearCommand() error {
	id, err := cli.promptForTaskID("Enter task ID you want to clear description")
	if err != nil {
		return err
	}

	t, err := cli.taskManager.GetTaskByID(id)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.output, "Task to clear: '%s'\n", task.FormatTask(t))

	if err = cli.taskManager.ClearDescription(id); err != nil {
		return err
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) description cleared!\n", id)
	return nil
}

func (cli *CLI) handleUpdateCommand() error {
	id, err := cli.promptForTaskID("Enter task ID to update")
	if err != nil {
		return err
	}

	t, err := cli.taskManager.GetTaskByID(id)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.output, "Current task: '%s'\n", task.FormatTask(t))

	fmt.Fprintln(cli.output, "Enter new description:")
	desc, err := cli.input.ReadInput(200)
	if err != nil {
		return err
	}

	if desc == t.Description {
		return fmt.Errorf("Task (ID: %d) Description unchanged", id)
	}

	if err = cli.taskManager.UpdateTaskDescription(id, desc); err != nil {
		return err
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) updated\n", id)
	return nil
}

func (cli *CLI) handleLoadCommand() error {
	loadedTasks, err := cli.storage.LoadTasks()
	if err != nil {
		return err
	}

	cli.taskManager.SetTasks(loadedTasks)
	fmt.Fprintf(cli.output, "✅ %d tasks loaded successfully!", len(loadedTasks))

	return nil
}

func (cli *CLI) handleDeleteCommand() error {
	id, err := cli.promptForTaskID("Enter task ID to delete task:")
	if err != nil {
		return err
	}

	t, err := cli.taskManager.GetTaskByID(id)
	if err != nil {
		return err
	}
	fmt.Fprintf(cli.output, "Task to delete: '%s'. Enter y/N:\n", task.FormatTask(t))
	str, err := cli.input.ReadInput(10)
	if err != nil {
		return err
	}
	str = strings.ToLower(str)

	switch str {
	case "y":
		if err = cli.taskManager.DeleteTask(id); err != nil {
			return err
		}
		fmt.Fprintf(cli.output, "✅ Task (ID: %d) deleted\n", id)
		return nil
	case "n":
		fmt.Fprintln(cli.output, "Deletion canceled")
		return nil
	default:
		return fmt.Errorf("invalid choice: %q; must be 'y' or 'n'", str)
	}
}
