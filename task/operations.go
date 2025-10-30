package task

import (
	"errors"
	"fmt"
	"io"
	"myproject/storage"
	"strings"
	"sync"
	"time"
)

// TaskManager provides business logic operations for task management.
// Delegates all persistence operations to the Storage interface.
type TaskManager struct {
	s      storage.Storage
	writer io.Writer
}

const processDelay = 500 * time.Millisecond

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrPrintTask    = errors.New("failed to print tasks")
)

// NewTaskManager создает новый экземпляр TaskManager с указанным writer для вывода.
func NewTaskManager(s storage.Storage, writer io.Writer) *TaskManager {
	return &TaskManager{
		s:      s,
		writer: writer,
	}
}

// AddTask creates a new task with the given description and persists it to storage.
// The ID is automatically assigned by the database and returned.
// Returns the assigned task ID and any error encountered during creation.
func (tm *TaskManager) AddTask(input string) (int, error) {
	task := storage.Task{
		Description: input,
		Done:        false,
	}

	return tm.s.CreateTask(task)
}

// UpdateTaskStatus updates the completion status of a task by ID.
// Fetches the task from storage, modifies the status, and saves it back.
// Returns an error if the task is not found or if the update fails.
func (tm *TaskManager) UpdateTaskStatus(id int, done bool) error {
	t, err := tm.s.GetTaskByID(id)
	if err != nil {
		return err
	}

	t.Done = done

	if err = tm.s.UpdateTask(t); err != nil {
		return err
	}

	return nil
}

// PrintTasks выводит все задачи в указанный writer.
// Возвращает ошибку при проблемах с записью.
func (tm *TaskManager) PrintTasks() error {
	tasks, err := tm.s.LoadTasks()
	if err != nil {
		return err
	}

	if _, err = tm.writer.Write([]byte(formatTasks(tasks))); err != nil {
		return ErrPrintTask
	}

	return nil
}

// FormatTask форматирует одну задачу в строку со статусом и описанием.
func FormatTask(task storage.Task) string {
	status := "  "
	if task.Done {
		status = "✓ "
	}
	return fmt.Sprintf("[%s] ID: %d, Description: %s", status, task.ID, task.Description)
}

// formatTasks форматирует список задач в многострочную строку.
// Возвращает сообщение "No tasks available" для пустого списка.
func formatTasks(tasks []storage.Task) string {
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

// GetFormattedTasks retrieves all tasks from storage and returns them as a formatted string.
// Returns an error if tasks cannot be loaded from storage.
func (tm *TaskManager) GetFormattedTasks() (string, error) {
	taskCopy, err := tm.s.LoadTasks()
	if err != nil {
		return "", err
	}
	return formatTasks(taskCopy), nil
}

// ClearDescription clears the description of a task by ID.
// Fetches the task from storage, clears the description, and saves it back.
// Returns an error if the task is not found or if the update fails.
func (tm *TaskManager) ClearDescription(id int) error {
	t, err := tm.s.GetTaskByID(id)
	if err != nil {
		return err
	}

	t.Description = ""

	if err := tm.s.UpdateTask(t); err != nil {
		return err
	}

	return nil
}

// UpdateTaskDescription updates the description of a task by ID.
// Fetches the task from storage, updates the description, and saves it back.
// Returns an error if the task is not found or if the update fails.
func (tm *TaskManager) UpdateTaskDescription(id int, description string) error {
	t, err := tm.s.GetTaskByID(id)
	if err != nil {
		return err
	}

	t.Description = description

	if err := tm.s.UpdateTask(t); err != nil {
		return err
	}

	return nil
}

// processTask симулирует обработку одной задачи с задержкой.
// Выполняется в отдельной горутине.
func processTask(task storage.Task, wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Printf("Processing task ID: %d\n", task.ID)
	time.Sleep(processDelay)
	fmt.Printf("Task ID: %d processed successfully\n", task.ID)
}

// ProcessTasks retrieves all tasks from storage and processes them in parallel.
// Each task is processed in a separate goroutine with simulated delay.
// Returns an error if tasks cannot be loaded from storage.
func (tm *TaskManager) ProcessTasks() error {
	tasksToProcess, err := tm.s.LoadTasks()
	if err != nil {
		return err
	}

	if len(tasksToProcess) == 0 {
		fmt.Fprintln(tm.writer, "No tasks to process")
		return nil
	}
	fmt.Println("Starting parallel task processing...")
	var wg sync.WaitGroup
	for _, task := range tasksToProcess {
		wg.Add(1)
		go processTask(task, &wg) // Передаем копию Task в горутину
	}
	wg.Wait()
	fmt.Println("All tasks processed successfully!")

	return nil
}
