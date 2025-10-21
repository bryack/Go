package task

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Task represents a single task with unique ID, description, and completion status.
// All fields are JSON-serializable for API responses.
type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}

// TaskManager provides thread-safe operations for managing a collection of tasks.
// Uses mutex synchronization to prevent race conditions in concurrent access.
type TaskManager struct {
	tasks  []Task
	mu     sync.Mutex
	writer io.Writer
}

const processDelay = 500 * time.Millisecond

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrPrintTask    = errors.New("failed to print tasks")
)

// NewTaskManager создает новый экземпляр TaskManager с указанным writer для вывода.
func NewTaskManager(writer io.Writer) *TaskManager {
	return &TaskManager{
		tasks:  make([]Task, 0),
		writer: writer,
	}
}

// generateMaxID возвращает следующий доступный ID для новой задачи.
func (tm *TaskManager) generateMaxID() int {
	maxID := 0

	for _, t := range tm.tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID + 1
}

// GetTasks returns an independent copy of all tasks in the manager.
// This method is thread-safe and prevents external modifications to internal state.
func (tm *TaskManager) GetTasks() []Task {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tasksCopy := make([]Task, len(tm.tasks))
	copy(tasksCopy, tm.tasks)
	return tasksCopy
}

// SetTasks replaces the current task collection with the provided tasks.
// Creates an independent copy to prevent external modifications after assignment.
func (tm *TaskManager) SetTasks(newTask []Task) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.tasks = make([]Task, len(newTask))
	copy(tm.tasks, newTask)
}

// AddTask добавляет новую задачу с указанным описанием и возвращает ее ID.
func (tm *TaskManager) AddTask(input string) int {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	newID := tm.generateMaxID()
	tm.tasks = append(tm.tasks, Task{ID: newID, Description: input, Done: false})
	return newID
}

// UpdateTaskStatus sets the completion status of a task by ID.
// Returns ErrTaskNotFound if the task doesn't exist.
// This operation is thread-safe.
func (tm *TaskManager) UpdateTaskStatus(id int, done bool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i := range tm.tasks {
		if tm.tasks[i].ID == id {
			tm.tasks[i].Done = done
			return nil
		}
	}
	return ErrTaskNotFound
}

// PrintTasks выводит все задачи в указанный writer.
// Возвращает ошибку при проблемах с записью.
func (tm *TaskManager) PrintTasks() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	return printToWriter(tm.tasks, tm.writer)
}

// FormatTask форматирует одну задачу в строку со статусом и описанием.
func FormatTask(task Task) string {
	status := "  "
	if task.Done {
		status = "✓ "
	}
	return fmt.Sprintf("[%s] ID: %d, Description: %s", status, task.ID, task.Description)
}

// formatTasks форматирует список задач в многострочную строку.
// Возвращает сообщение "No tasks available" для пустого списка.
func formatTasks(tasks []Task) string {
	var builder strings.Builder

	if len(tasks) == 0 {
		builder.WriteString("No tasks available")
		return builder.String()
	}

	for i, task := range tasks {
		strTask := FormatTask(task)
		builder.WriteString(strTask)
		if i < len(tasks)-1 {
			builder.WriteString("\n")
		}
	}
	return builder.String()
}

// GetFormattedTasks возвращает отформатированную строку всех задач.
// Использует потокобезопасную копию списка задач.
func (tm *TaskManager) GetFormattedTasks() string {
	taskCopy := tm.GetTasks()
	return formatTasks(taskCopy)
}

// printToWriter записывает отформатированный список задач в указанный writer.
// Абстракция формата вывода: Функция принимает io.Writer, чтобы записывать
// отформатированные задачи (formateTasks) в любой получатель (консоль, файл, буфер).
func printToWriter(tasks []Task, writer io.Writer) error {
	_, err := writer.Write([]byte(formatTasks(tasks)))
	if err != nil {
		return ErrPrintTask
	}
	return nil
}

// ClearDescription очищает описание задачи с указанным ID.
// Возвращает ErrTaskNotFound, если задача не найдена.
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

// UpdateTaskDescription updates the description of a task with the specified ID.
// Returns ErrTaskNotFound if the task is not found.
func (tm *TaskManager) UpdateTaskDescription(id int, description string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i := range tm.tasks {
		if tm.tasks[i].ID == id {
			tm.tasks[i].Description = description
			return nil
		}
	}
	return ErrTaskNotFound
}

// GetTaskByID retrieves a task by its ID and returns the Task struct.
// Returns ErrTaskNotFound if no task with the specified ID exists.
func (tm *TaskManager) GetTaskByID(id int) (Task, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, task := range tm.tasks {
		if tm.tasks[i].ID == id {
			return task, nil
		}
	}
	return Task{}, ErrTaskNotFound
}

// DeleteTask removes a task with the specified ID from the collection.
// Returns ErrTaskNotFound if no task exists with the given ID.
func (tm *TaskManager) DeleteTask(id int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i := range tm.tasks {
		if tm.tasks[i].ID == id {
			tm.tasks = append(tm.tasks[:i], tm.tasks[i+1:]...)
			return nil
		}
	}
	return ErrTaskNotFound
}

// processTask симулирует обработку одной задачи с задержкой.
// Выполняется в отдельной горутине.
func processTask(task Task, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Processing task ID: %d\n", task.ID)
	time.Sleep(processDelay)
	fmt.Printf("Task ID: %d processed successfully\n", task.ID)
}

// ProcessTasks запускает параллельную обработку всех задач.
// Каждая задача обрабатывается в отдельной горутине.
func (tm *TaskManager) ProcessTasks() {
	// 1. Получаем независимую, потокобезопасную копию списка задач.
	//    Метод GetTasks() уже заботится о блокировке мьютекса и создании копии.
	tasksToProcess := tm.GetTasks()

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
