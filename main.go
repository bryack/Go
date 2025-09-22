package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"myproject/storage"
	"myproject/task"
	"os"
	"strconv"
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
)

var (
	validCommands      = []Command{CommandAdd, CommandDone, CommandList, CommandProcess, CommandLoad, CommandClear, CommandHelp, CommandExit}
	ErrMaxSizeExceeded = errors.New("input too long")
	ErrEmptyInput      = errors.New("empty input")
	ErrInvalidTaskId   = errors.New("invalid ID format")
	ErrInvalidCommand  = errors.New("invalid command")
)

// readInput читает пользовательский ввод с ограничением размера
func readInput(maxSize int) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return "", io.EOF
		}
		return "", err
	}
	input = strings.TrimSpace(strings.TrimRight(input, "\r\n"))
	if len(input) > maxSize {
		return "", ErrMaxSizeExceeded
	}
	if len(input) == 0 {
		return "", ErrEmptyInput
	}
	return input, nil
}

// validateTaskID converts a string input to a valid task ID.
// Returns the parsed ID if valid (positive integer), or an error if invalid.
func validateTaskID(input string) (int, error) {
	id, err := strconv.Atoi(input)
	if err != nil {
		return 0, ErrInvalidTaskId
	}
	if id <= 0 {
		return 0, ErrInvalidTaskId
	}
	return id, nil
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

// handleError обрабатывает различные типы ошибок
func handleError(err error, context string) {
	switch {
	case errors.Is(err, io.EOF):
		fmt.Printf("%s: input interrupted by user\n", context)
	case errors.Is(err, ErrMaxSizeExceeded):
		fmt.Printf("%s: input exceeds %d characters\n", context, maxInputSize)
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
		input, err := readInput(10)
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
			fmt.Println("enter task description:")
			desc, err := readInput(50)
			if err != nil {
				handleError(err, "Description input error")
				continue
			}
			id := tm.AddTask(desc)
			fmt.Printf("✅ Task added (ID: %d)\n", id)

		case CommandDone:
			fmt.Println("Enter task ID to mark as done:")
			input, err := readInput(10)
			if err != nil {
				handleError(err, "ID input error")
				continue
			}

			id, err := validateTaskID(input)
			if err != nil {
				handleError(err, "❌ ID conversion error")
				continue
			}
			if err := tm.MarkTaskDone(id); err != nil {
				handleError(err, "Mark done error")
				continue
			}
			fmt.Println("Task marked as done")

		case CommandList:
			if err := tm.PrintTasks(); err != nil {
				handleError(err, "Print tasks error")
			}

		case CommandProcess:
			tm.ProcessTasks()

		case CommandLoad:
			loadedTasks, err := s.LoadTasks()
			if err != nil {
				handleError(err, "Load error")
			} else {
				tm.SetTasks(loadedTasks)
				fmt.Println("✅ Tasks loaded successfully!")
			}

		case CommandClear:
			fmt.Println("enter task id you want to clear description")
			idSrt, err := readInput(maxInputSize)
			if err != nil {
				handleError(err, "ID input error")
				continue
			}

			id, err := validateTaskID(idSrt)
			if err != nil {
				handleError(err, "❌ ID conversion error")
				continue
			}

			if err = tm.ClearDescription(id); err != nil {
				handleError(err, "Clear description error")
				continue
			}
			fmt.Println("✅ Task description cleared!")

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
		}
	}
}
