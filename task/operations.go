package task

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

const processDelay = 500 * time.Millisecond

var ErrTaskNotFound = errors.New("task not found")

func generateMaxID(tasks *[]Task) int {
	maxID := 0

	for i := range *tasks {
		if (*tasks)[i].ID > maxID {
			maxID = (*tasks)[i].ID
		}
	}
	return maxID + 1
}

// AddTask добавляет новую задачу в список
func AddTask(tasks *[]Task, input string) int {
	newID := generateMaxID(tasks)
	*tasks = append(*tasks, Task{ID: newID, Description: input, Done: false})
	return newID
}

// MarkTaskDone помечает задачу как выполненную
func MarkTaskDone(tasks *[]Task, id int) error {
	for index := range *tasks {
		if (*tasks)[index].ID == id {
			(*tasks)[index].Done = true
			return nil
		}
	}
	return ErrTaskNotFound
}

// PrintTasks выводит список всех задач
func PrintTasks(tasks []Task) {
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
func ClearDescription(tasks *[]Task, id int) error {
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

// ProcessTasks запускает параллельную обработку всех задач
func ProcessTasks(tasks []Task) {
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
