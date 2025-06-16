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
	"time"
)

type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

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
	for _, task := range tasks {
		if task.ID > lastId {
			lastId = task.ID
		}
	}
	fmt.Println("tasks downloaded from tasks.json:", tasks)
	return tasks, nil
}
func saveTasks(tasks []Task) error {
	// Преобразуем срез задач в JSON-формат ([]byte)
	data, err := json.Marshal(tasks)
	if err != nil {
		return ErrConversionTask
	}
	err = os.WriteFile("tasks.json", data, 0644)
	if err != nil {
		return ErrFailedWriteFile
	}
	return nil
}
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
func addTask(tasks *[]Task, input string) int {
	lastId++
	*tasks = append(*tasks, Task{ID: lastId, Description: input, Done: false})
	return lastId
}
func makeTaskDone(tasks *[]Task, id int) error {
	for index := range *tasks {
		if (*tasks)[index].ID == id {
			(*tasks)[index].Done = true
			return nil
		}
	}
	return ErrTaskNotFound
}
func printTasks(tasks []Task) {
	if len(tasks) == 0 {
		fmt.Println("No tasks available")
		return
	}
	for _, task := range tasks {
		go processTask(task)
	}
	time.Sleep(1 * time.Second) // Временная заглушка
	for _, task := range tasks {
		status := "  "
		if task.Done {
			status = "✓ "
		}
		fmt.Printf("[%s] ID: %d, Description: %s\n", status, task.ID, task.Description)
	}
}
func clearDescription(tasks *[]Task, id int) error {
	for i, task := range *tasks {
		if task.ID == id {
			(*tasks)[i].Description = ""
			return nil
		}
	}
	return ErrTaskNotFound
}
func processTask(task Task) {
	time.Sleep(200 * time.Millisecond)
	fmt.Printf("Processing task ID: %d\n", task.ID)
}
func main() {
	tasks := []Task{}
	fmt.Println("Task Manager. Commands: add, done, list, load, clear, exit")
	for {
		input, err := readInput(10)
		if err != nil {
			switch err {
			case io.EOF:
				fmt.Println("input interrupted by user:", err)
			case ErrMaxSizeExceeded:
				fmt.Println("input exceeds 10 bytes:", err)
			case ErrEmptyInput:
				fmt.Println("you should enter smth:", err)
			default:
				fmt.Printf("%v\n", err)
			}
			continue
		}
		switch {
		case input == "exit":
			errSave := saveTasks(tasks)
			if errSave == ErrConversionTask {
				fmt.Println("tasks could not converse:", errSave)
				return
			}
			if errSave == ErrFailedWriteFile {
				fmt.Println("file could not write:", errSave)
			}
			fmt.Println("Bye")
			return
		case input == "list":
			printTasks(tasks)
		case input == "add":
			fmt.Println("enter task description:")
			desc, err := readInput(10)
			if err != nil {
				fmt.Println("error reading description:", err)
				continue
			}
			id := addTask(&tasks, desc)
			fmt.Printf("Task added (ID: %d)\n", id)
		case input == "done":
			fmt.Println("enter task id for done:")
			input, err := readInput(10)
			if err != nil {
				fmt.Println("error reading ID", err)
				continue
			}
			inputInt, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("invalid ID:", err)
			}
			if err := makeTaskDone(&tasks, inputInt); err != nil {
				fmt.Println("error:", err)
				continue
			}
			fmt.Println("Task marked as done")
		case input == "load":
			tasks, err = loadTasks()
			if err != nil {
				switch err {
				case ErrFileNotFound:
					fmt.Println("load error:", err)
				case ErrParseJson:
					fmt.Println("parse error:", err)
				default:
					fmt.Println("Unknown load file error")
				}
			}
			fmt.Println("tasks downloaded from tasks.json")
			continue
		case input == "clear":
			fmt.Println("enter task id you want to clear description")
			input, err := readInput(10)
			if err != nil {
				fmt.Println("error reading ID", err)
				continue
			}
			inputInt, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("invalid ID:", err)
				continue
			}
			err = clearDescription(&tasks, inputInt)
			if err == ErrTaskNotFound {
				fmt.Println("error:", err)
				continue
			}
			fmt.Println("Task dexcription cleared")
		default:
			fmt.Println("Unknown command. Available: add, done, list, load, clear, exit")
		}
	}
}
