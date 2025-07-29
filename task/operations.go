package task

import (
	"errors"
	"fmt"
	"strings"
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
//	[]Task: Независимый срез задач.
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

	fmt.Print(formateTasks(tm.tasks))
}

// formatTask форматирует одну задачу в строку
func formateTask(task Task) string {
	status := "  "
	if task.Done {
		status = "✓ "
	}
	return fmt.Sprintf("[%s] ID: %d, Description: %s", status, task.ID, task.Description)
}

// formatTasks форматирует список задач в строку
func formateTasks(tasks []Task) string {
	var builder strings.Builder

	if len(tasks) == 0 {
		fmt.Println("No tasks available")
		return builder.String()
	}

	for i, task := range tasks {
		strTask := formateTask(task)
		builder.WriteString(strTask)
		if i < len(tasks)-1 {
			builder.WriteString(",\n")
		}
	}
	return builder.String()
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
	// 1. Получаем независимую, потокобезопасную копию списка задач.
	//    Метод GetTasks() уже заботится о блокировке мьютекса и создании копии.
	tasksToProcess := tm.GetTasks() // <-- ИСПОЛЬЗУЕМ GetTasks() здесь!

	if len(tasksToProcess) == 0 {
		fmt.Println("No tasks to process")
		return
	}
	fmt.Println("Starting parallel task processing...")
	var wg sync.WaitGroup
	// 2. Итерируемся по ЭТОЙ КОПИИ, которая не будет изменяться другими горутинами.
	for _, task := range tasksToProcess {
		wg.Add(1)
		go processTask(task, &wg) // Передаем копию Task в горутину
	}
	wg.Wait()
	fmt.Println("All tasks processed successfully!")
}
