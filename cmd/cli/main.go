package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"myproject/storage"
	"myproject/task"
	"myproject/validation"
	"os"
	"strings"
)

// Command represents a valid user command in the task manager CLI.
// Commands are case-insensitive and validated against a predefined set.
type Command string

const (
	maxInputSize           = 10
	CommandAdd     Command = "add"     // Add a new task
	CommandDone    Command = "done"    // Mark task as completed
	CommandList    Command = "list"    // Show all tasks
	CommandProcess Command = "process" // Process all tasks in parallel
	CommandLoad    Command = "load"    // Load tasks from file
	CommandClear   Command = "clear"   // Clear task description
	CommandHelp    Command = "help"    // Show available commands
	CommandExit    Command = "exit"    // Save and exit program
	CommandUpdate  Command = "update"  // Update task description
	CommandDelete  Command = "delete"
)

var (
	validCommands      = []Command{CommandAdd, CommandDone, CommandList, CommandProcess, CommandLoad, CommandClear, CommandHelp, CommandExit, CommandUpdate, CommandDelete}
	ErrMaxSizeExceeded = errors.New("input too long")
	ErrEmptyInput      = errors.New("empty input")
	ErrInvalidTaskId   = errors.New("invalid ID format")
	ErrInvalidCommand  = errors.New("invalid command")
)

// readInput читает пользовательский ввод с ограничением размера
func readInput(reader io.Reader, maxSize int) (string, error) {
	bufReader := bufio.NewReader(reader)
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

// isValid checks if the command is in the list of supported commands.
// Returns true if the command is valid, false otherwise.
func (cmd Command) isValid() bool {
	for _, valid := range validCommands {
		if cmd == valid {
			return true
		}
	}
	return false
}

// validateCommand converts user input to a valid Command.
// Input is normalized to lowercase before validation.
// Returns the valid command or an error if the command is not recognized.
func validateCommand(input string) (Command, error) {
	inputToLower := strings.ToLower(input)
	cmd := Command(inputToLower)
	if cmd.isValid() {
		return cmd, nil
	}
	return "", ErrInvalidCommand
}

// suggestCommand attempts to find a command that matches the input prefix.
// Returns the first matching command, or empty string if no match is found.
// Used to provide helpful suggestions for typos or partial commands.
func suggestCommand(input string) Command {
	for _, cmd := range validCommands {
		if strings.HasPrefix(string(cmd), input) {
			return cmd
		}
	}
	return ""
}

// promptForTaskID prompts the user for a task ID with a custom message.
// Returns the validated task ID or an error if input is invalid.
func promptForTaskID(prompt string) (int, error) {
	fmt.Println(prompt)
	input, err := readInput(os.Stdin, maxInputSize)
	if err != nil {
		return 0, ErrInvalidTaskId
	}
	return validation.ValidateTaskID(input)
}

func handleAddCommand(tm *task.TaskManager) error {
	fmt.Println("enter task description:")
	desc, err := readInput(os.Stdin, 50)
	if err != nil {
		return err
	}
	id := tm.AddTask(desc)
	fmt.Printf("✅ Task added (ID: %d)\n", id)
	return nil
}

func handleDoneCommand(tm *task.TaskManager) error {
	prompt := "Enter task ID to mark as done:"
	id, err := promptForTaskID(prompt)
	if err != nil {
		return err
	}

	t, err := tm.GetTaskByID(id)
	if err != nil {
		return err
	}
	fmt.Printf("Task to mark done: '%s'\n", task.FormatTask(t))

	if err := tm.MarkTaskDone(id); err != nil {
		return err
	}

	fmt.Println("✅ Task marked as done")
	return nil
}

func handleClearCommand(tm *task.TaskManager) error {
	prompt := "Enter task id you want to clear description"
	id, err := promptForTaskID(prompt)
	if err != nil {
		return err
	}

	t, err := tm.GetTaskByID(id)
	if err != nil {
		return err
	}
	fmt.Printf("Task to clear: '%s'\n", task.FormatTask(t))

	if err = tm.ClearDescription(id); err != nil {
		return err
	}
	fmt.Println("✅ Task description cleared!")
	return nil
}

func handleUpdateCommand(tm *task.TaskManager) error {
	prompt := "Enter task ID to update"
	id, err := promptForTaskID(prompt)
	if err != nil {
		return err
	}

	t, err := tm.GetTaskByID(id)
	if err != nil {
		return err
	}
	fmt.Printf("Current task: '%s'\n", task.FormatTask(t))

	fmt.Println("Enter new description")
	description, err := readInput(os.Stdin, 50)
	if err != nil {
		return err
	}

	err = tm.UpdateTaskDescription(id, description)
	if err != nil {
		return err
	}
	fmt.Printf("✅ Task updated (ID: %d)\n", id)
	return nil
}

func handleLoadCommand(tm *task.TaskManager, s storage.Storage) error {
	loadedTasks, err := s.LoadTasks()
	if err != nil {
		return err
	}
	tm.SetTasks(loadedTasks)
	fmt.Println("✅ Tasks loaded successfully!")

	return nil
}

func handleDeleteCommand(tm *task.TaskManager) error {
	prompt := "Enter task ID to delete task:"
	id, err := promptForTaskID(prompt)
	if err != nil {
		return err
	}

	t, err := tm.GetTaskByID(id)
	if err != nil {
		return err
	}
	fmt.Printf("Task to delete: '%s'\n", task.FormatTask(t))
	if err := tm.DeleteTask(id); err != nil {
		return err
	}

	fmt.Println("Task deleted")
	return nil
}

// handleError обрабатывает различные типы ошибок
func handleError(err error, context string) {
	switch {
	case errors.Is(err, io.EOF):
		fmt.Printf("%s: input interrupted by user\n", context)
	// case errors.Is(err, ErrMaxSizeExceeded):
	// 	fmt.Printf("%s: input exceeds %d characters\n", context, maxInputSize)
	case errors.Is(err, ErrEmptyInput):
		fmt.Printf("%s: empty input provided\n", context)
	case errors.Is(err, task.ErrTaskNotFound):
		fmt.Printf("%s: task not found\n", context)
	case errors.Is(err, storage.ErrConversionTask):
		fmt.Printf("%s: failed to convert tasks\n", context)
	case errors.Is(err, storage.ErrFailedWriteFile):
		fmt.Printf("%s: failed to write file\n", context)
	case errors.Is(err, storage.ErrFileNotFound):
		fmt.Printf("%s: file not found\n", context)
	case errors.Is(err, storage.ErrParseJson):
		fmt.Printf("%s: JSON parsing error\n", context)
	case errors.Is(err, ErrInvalidTaskId):
		fmt.Printf("%s: ID must be a positive number (greater than 0)\n", context)
	case errors.Is(err, task.ErrPrintTask):
		fmt.Printf("%s: failed to print tasks\n", context)
	case errors.Is(err, ErrInvalidCommand):
		fmt.Printf("%s: failed to assume command\n", context)
	default:
		fmt.Printf("%s: %v\n", context, err)
	}
}

// showHelp показывает доступные команды
func showHelp() {
	fmt.Println("\n=== Available Commands ===")
	fmt.Println("add     - Add a new task")
	fmt.Println("done    - Mark task as completed")
	fmt.Println("list    - Show all tasks")
	fmt.Println("process - Process all tasks in parallel")
	fmt.Println("load    - Load tasks from file")
	fmt.Println("clear   - Clear task description")
	fmt.Println("update  - Update task description")
	fmt.Println("delete  - Delete task")
	fmt.Println("help    - Show this help")
	fmt.Println("exit    - Save and exit")
	fmt.Println("=========================")
}

func main() {
	tm := task.NewTaskManager(os.Stdout)
	var s storage.Storage = storage.JsonStorage{}
	fmt.Println("🚀 Task Manager Started!")
	showHelp()
	for {
		fmt.Print("\nEnter command: ")
		input, err := readInput(os.Stdin, maxInputSize)
		if err != nil {
			handleError(err, "Input error")
			continue
		}

		cmd, err := validateCommand(input)
		if err != nil {
			suggestion := suggestCommand(input)
			if suggestion != "" {
				fmt.Printf("❌ Unknown command: '%s', maybe you wanted: '%s'\n", input, suggestion)
			} else {
				handleError(err, "Command validate error")
				fmt.Println("Type 'help' to see available commands")
			}
			continue
		}

		switch Command(cmd) {
		case CommandAdd:
			if err := handleAddCommand(tm); err != nil {
				handleError(err, "Add command error")
			}

		case CommandDone:
			if err := handleDoneCommand(tm); err != nil {
				handleError(err, "Done command error")
			}

		case CommandList:
			if err := tm.PrintTasks(); err != nil {
				handleError(err, "Print tasks error")
			}

		case CommandProcess:
			tm.ProcessTasks()

		case CommandLoad:
			if err := handleLoadCommand(tm, s); err != nil {
				handleError(err, "Load command error")
			}

		case CommandClear:
			if err := handleClearCommand(tm); err != nil {
				handleError(err, "Clear command error")
			}

		case CommandDelete:
			if err := handleDeleteCommand(tm); err != nil {
				handleError(err, "Delete command error")
			}

		case CommandHelp:
			showHelp()

		case CommandExit:
			if err := s.SaveTasks(tm.GetTasks()); err != nil {
				handleError(err, "Save error")
			} else {
				fmt.Println("Tasks saved successfully!")
			}
			fmt.Println("👋 Bye!")
			return

		case CommandUpdate:
			if err := handleUpdateCommand(tm); err != nil {
				handleError(err, "Update command error")
			}
		}
	}
}
