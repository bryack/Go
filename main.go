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
	switch err {
	case io.EOF:
		fmt.Printf("%s: input interrupted by user\n", context)
	case ErrMaxSizeExceeded:
		fmt.Printf("%s: input exceeds %d characters\n", context, maxInputSize)
	case ErrEmptyInput:
		fmt.Printf("%s: empty input provided\n", context)
	case task.ErrTaskNotFound:
		fmt.Printf("%s: task not found\n", context)
	case storage.ErrConversionTask:
		fmt.Printf("%s: failed to convert tasks\n", context)
	case storage.ErrFailedWriteFile:
		fmt.Printf("%s: failed to write file\n", context)
	case storage.ErrFileNotFound:
		fmt.Printf("%s: file not found\n", context)
	case storage.ErrParseJson:
		fmt.Printf("%s: JSON parsinf error\n", context)
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
	tasks := []task.Task{}
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
			if err := storage.SaveTasks(tasks); err != nil {
				handleError(err, "Save error")
			} else {
				fmt.Println("Tasks saved successfully!")
			}
			fmt.Println("👋 Bye!")
			return

		case "list":
			task.PrintTasks(tasks)

		case "add":
			fmt.Println("enter task description:")
			desc, err := readInput(50)
			if err != nil {
				handleError(err, "Description input error")
				continue
			}
			id := task.AddTask(&tasks, desc)
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
				fmt.Println("❌ Invalid ID format")
			}
			if err := task.MarkTaskDone(&tasks, id); err != nil {
				handleError(err, "Mark done error")
				continue
			}
			fmt.Println("Task marked as done")

		case "load":
			tasks, err = storage.LoadTasks()
			if err != nil {
				handleError(err, "Load error")
			} else {
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
				fmt.Println("❌ Invalid ID format")
				continue
			}

			if err = task.ClearDescription(&tasks, id); err != nil {
				handleError(err, "Clear description error")
				continue
			}
			fmt.Println("✅ Task description cleared!")

		case "process":
			task.ProcessTasks(tasks)

		case "help":
			showHelp()

		default:
			fmt.Printf("❌ Unknown command: '%s'\n", input)
			fmt.Println("Type 'help' to see available commands")
		}
	}
}
