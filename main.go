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

const (
	maxInputSize = 10
)

var (
	ErrMaxSizeExceeded = errors.New("input too long")
	ErrEmptyInput      = errors.New("empty input")
	ErrInvalidIdFormat = errors.New("invalid ID format")
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
	case errors.Is(err, ErrInvalidIdFormat):
		fmt.Printf("%s: invalid ID format\n", context)
	case errors.Is(err, task.ErrPrintTask):
		fmt.Printf("%s: failed to print tasks\n", context)
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
		switch input {
		case "exit":
			if err := s.SaveTasks(tm.GetTasks()); err != nil {
				handleError(err, "Save error")
			} else {
				fmt.Println("Tasks saved successfully!")
			}
			fmt.Println("👋 Bye!")
			return

		case "list":
			if err := tm.PrintTasks(); err != nil {
				handleError(err, "Print tasks error")
			}

		case "add":
			fmt.Println("enter task description:")
			desc, err := readInput(50)
			if err != nil {
				handleError(err, "Description input error")
				continue
			}
			id := tm.AddTask(desc)
			fmt.Printf("✅ Task added (ID: %d)\n", id)

		case "done":
			fmt.Println("Enter task ID to mark as done:")
			input, err := readInput(10)
			if err != nil {
				handleError(err, "ID input error")
				continue
			}

			id, err := strconv.Atoi(input)
			if err != nil {
				handleError(fmt.Errorf("%w: %v", ErrInvalidIdFormat, err), "❌ ID conversion error")
				continue
			}
			if err := tm.MarkTaskDone(id); err != nil {
				handleError(err, "Mark done error")
				continue
			}
			fmt.Println("Task marked as done")

		case "load":
			loadedTasks, err := s.LoadTasks()
			if err != nil {
				handleError(err, "Load error")
			} else {
				tm.SetTasks(loadedTasks)
				fmt.Println("✅ Tasks loaded successfully!")
			}

		case "clear":
			fmt.Println("enter task id you want to clear description")
			idSrt, err := readInput(maxInputSize)
			if err != nil {
				handleError(err, "ID input error")
				continue
			}

			id, err := strconv.Atoi(idSrt)
			if err != nil {
				handleError(fmt.Errorf("%w: %v", ErrInvalidIdFormat, err), "❌ ID conversion error")
				continue
			}

			if err = tm.ClearDescription(id); err != nil {
				handleError(err, "Clear description error")
				continue
			}
			fmt.Println("✅ Task description cleared!")

		case "process":
			tm.ProcessTasks()

		case "help":
			showHelp()

		default:
			fmt.Printf("❌ Unknown command: '%s'\n", input)
			fmt.Println("Type 'help' to see available commands")
		}
	}
}
