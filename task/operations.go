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

type TaskManager struct {
	tasks []Task
	mu    sync.Mutex
}

const processDelay = 500 * time.Millisecond

var ErrTaskNotFound = errors.New("task not found")

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: make([]Task, 0),
	}
}

func (tm *TaskManager) generateMaxID() int {
	maxID := 0

	for _, t := range tm.tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID + 1
}

// GetTasks возвращает независимую копию текущего списка задач.
//
// Этот метод обеспечивает потокобезопасное чтение внутреннего списка задач,
// используя мьютекс для предотвращения состояний гонки. Возвращаемый срез
// является копией, что гарантирует, что внешние модификации не повлияют
// на внутреннее состояние TaskManager.
//
// Возвращает:
//
//	[]Task: Независимый срез задач.с
func (tm *TaskManager) GetTasks() []Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tasksCopy := make([]Task, len(tm.tasks))
	copy(tasksCopy, tm.tasks)
	return tasksCopy
}

// SetTasks устанавливает новый список задач, заменяя текущий.
//
// Этот метод обеспечивает потокобезопасное обновление внутреннего списка задач,
// используя мьютекс для предотвращения состояний гонки. Входящий срез
// newTasks копируется во внутреннее хранилище, что гарантирует, что
// последующие внешние модификации newTasks не повлияют на внутреннее
// состояние TaskManager.
//
// Параметры:
//
//	newTasks []Task: Новый список задач для установки.
func (tm *TaskManager) SetTasks(newTask []Task) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.tasks = make([]Task, len(newTask))
	copy(tm.tasks, newTask)
}

// AddTask добавляет новую задачу в список
func (tm *TaskManager) AddTask(input string) int {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	newID := tm.generateMaxID()
	tm.tasks = append(tm.tasks, Task{ID: newID, Description: input, Done: false})
	return newID
}

// MarkTaskDone помечает задачу как выполненную
func (tm *TaskManager) MarkTaskDone(id int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i := range tm.tasks {
		if tm.tasks[i].ID == id {
			tm.tasks[i].Done = true
			return nil
		}
	}
	return ErrTaskNotFound
}

// PrintTasks выводит список всех задач
func (tm *TaskManager) PrintTasks() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if len(tm.tasks) == 0 {
		fmt.Println("No tasks available")
		return
	}
	for _, task := range tm.tasks {
		status := "  "
		if task.Done {
			status = "✓ "
		}
		fmt.Printf("[%s] ID: %d, Description: %s\n", status, task.ID, task.Description)
	}
	fmt.Println("================")
}

// clearTaskDescription очищает описание задачи
func (tm *TaskManager) ClearDescription(id int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i := range tm.tasks {
		if tm.tasks[i].ID == id {
			tm.tasks[i].Description = ""
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
func (tm *TaskManager) ProcessTasks() {
	if len(tm.tasks) == 0 {
		fmt.Println("No tasks to process")
		return
	}
	fmt.Println("Starting parallel task processing...")
	var wg sync.WaitGroup
	for _, task := range tm.tasks {
		wg.Add(1)
		go processTask(task, &wg)
	}
	wg.Wait()
	fmt.Println("All tasks processed successfully!")
}
