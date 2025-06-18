package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

const (
	processDelay = 500 * time.Millisecond
	maxInputSize = 10
)

var (
	ErrMaxSizeExceeded = errors.New("input too long")
	ErrEmptyInput      = errors.New("empty input")
	ErrTaskNotFound    = errors.New("task not found")
	ErrConversionTask  = errors.New("tasks conversion error")
	ErrFailedWriteFile = errors.New("failed to write tasks.json")
	ErrFileNotFound    = errors.New("file not found, tasks not downloaded")
	ErrParseJson       = errors.New("error parsing JSON")
	lastId             int
)

// loadTasks загружает задачи из файла tasks.json
func loadTasks() ([]Task, error) {
	// Попытка прочитать весь файл tasks.json
	data, err := os.ReadFile("tasks.json")
	if err != nil {
		// Если файл не существует или другая ошибка, возвращаем пустой список
		return []Task{}, ErrFileNotFound
	}
	// Декодируем JSON из []byte в срез Task
	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return []Task{}, ErrParseJson
	}
	// Обновляем lastId
	for _, task := range tasks {
		if task.ID > lastId {
			lastId = task.ID
		}
	}
	fmt.Println("tasks loaded from tasks.json:", tasks)
	return tasks, nil
}

// saveTasks сохраняет задачи в файл tasks.json
func saveTasks(tasks []Task) error {
	// Преобразуем срез задач в JSON-формат ([]byte)
	data, err := json.Marshal(tasks)
	if err != nil {
		return ErrConversionTask
	}
	if err = os.WriteFile("tasks.json", data, 0644); err != nil {
		return ErrFailedWriteFile
	}
	return nil
}

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

// addTask добавляет новую задачу в список
func addTask(tasks *[]Task, input string) int {
	lastId++
	*tasks = append(*tasks, Task{ID: lastId, Description: input, Done: false})
	return lastId
}

// markTaskDone помечает задачу как выполненную
func makeTaskDone(tasks *[]Task, id int) error {
	for index := range *tasks {
		if (*tasks)[index].ID == id {
			(*tasks)[index].Done = true
			return nil
		}
	}
	return ErrTaskNotFound
}

// printTasks выводит список всех задач
func printTasks(tasks []Task) {
	if len(tasks) == 0 {
		fmt.Println("No tasks available")
		return
	}
	for _, task := range tasks {
		status := "  "
		if task.Done {
			status = "✓ "
		}
		fmt.Printf("[%s] ID: %d, Description: %s\n", status, task.ID, task.Description)
	}
	fmt.Println("================")
}

// clearTaskDescription очищает описание задачи
func clearDescription(tasks *[]Task, id int) error {
	for i := range *tasks {
		if (*tasks)[i].ID == id {
			(*tasks)[i].Description = ""
			return nil
		}
	}
	return ErrTaskNotFound
}

// processTask обрабатывает одну задачу (симуляция работы)
func processTask(task Task, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Processing task ID: %d\n", task.ID)
	time.Sleep(processDelay)
	fmt.Printf("Task ID: %d processed successfully\n", task.ID)
}

// processTasks запускает параллельную обработку всех задач
func processTasks(tasks []Task) {
	if len(tasks) == 0 {
		fmt.Println("No tasks to process")
		return
	}
	fmt.Println("Starting parallel task processing...")
	var wg sync.WaitGroup
	for _, task := range tasks {
		wg.Add(1)
		go processTask(task, &wg)
	}
	wg.Wait()
	fmt.Println("All tasks processed successfully!")
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
	case ErrTaskNotFound:
		fmt.Printf("%s: task not found\n", context)
	case ErrConversionTask:
		fmt.Printf("%s: failed to convert tasks\n", context)
	case ErrFailedWriteFile:
		fmt.Printf("%s: failed to write file\n", context)
	case ErrFileNotFound:
		fmt.Printf("%s: file not found\n", context)
	case ErrParseJson:
		fmt.Printf("%s: JSON parsinf error\n", context)
	default:
		fmt.Printf("%s: %v\n", context, err)
	}
}

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
	tasks := []Task{}
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
			if err := saveTasks(tasks); err != nil {
				handleError(err, "Save error")
			} else {
				fmt.Println("Tasks saved successfully!")
			}
			fmt.Println("👋 Bye!")
			return

		case "list":
			printTasks(tasks)

		case "add":
			fmt.Println("enter task description:")
			desc, err := readInput(50)
			if err != nil {
				handleError(err, "Description input error")
				continue
			}
			id := addTask(&tasks, desc)
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
			if err := makeTaskDone(&tasks, id); err != nil {
				handleError(err, "Mark done error")
				continue
			}
			fmt.Println("Task marked as done")

		case "load":
			tasks, err = loadTasks()
			if err != nil {
				handleError(err, "Load error")
			} else {
				fmt.Println("✅ Tasks loaded successfully!")
			}
			// continue

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

			if err = clearDescription(&tasks, id); err != nil {
				handleError(err, "Clear description error")
				continue
			}
			fmt.Println("✅ Task description cleared!")

		case "process":
			processTasks(tasks)

		case "help":
			showHelp()

		default:
			fmt.Printf("❌ Unknown command: '%s'\n", input)
			fmt.Println("Type 'help' to see available commands")
		}
	}
}
