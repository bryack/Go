package task

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAddTask(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name          string
		input         string
		initialTasks  []Task
		expectedID    int
		expectedTasks []Task
	}{
		{
			name:          "Add task to empty list",
			input:         "task 1",
			initialTasks:  []Task{},
			expectedID:    1,
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
		},
		{
			name:          "Add task to non-empty list",
			input:         "task 2",
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedID:    2,
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}},
		},
		{
			name:          "Add empty description",
			input:         "",
			initialTasks:  []Task{},
			expectedID:    1,
			expectedTasks: []Task{{ID: 1, Description: "", Done: false}},
		},
		{
			name:          "Add long description",
			input:         "Это очень длинное описание задачи, которое содержит много текста и проверяет, может ли наша функция корректно работать с большими строками",
			initialTasks:  []Task{},
			expectedID:    1,
			expectedTasks: []Task{{ID: 1, Description: "Это очень длинное описание задачи, которое содержит много текста и проверяет, может ли наша функция корректно работать с большими строками", Done: false}},
		},
		{
			name:          "Add task with special characters",
			input:         "Купить молоко & хлеб в магазине \"Пятёрочка\"",
			initialTasks:  []Task{},
			expectedID:    1,
			expectedTasks: []Task{{ID: 1, Description: "Купить молоко & хлеб в магазине \"Пятёрочка\"", Done: false}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ==== ACT ====
			tm := NewTaskManager(&strings.Builder{})
			tm.SetTasks(tc.initialTasks)
			actualId := tm.AddTask(tc.input)

			// === ASSERT ===
			if actualId != tc.expectedID {
				t.Errorf("Expected ID '%d', got '%d'", tc.expectedID, actualId)
			}

			if len(tm.GetTasks()) != len(tc.expectedTasks) {
				t.Errorf("Expected task list length '%d', got '%d'", len(tc.expectedTasks), len(tm.GetTasks()))
			}

			// Check input against expectedTasks.Description
			if tc.input != tc.expectedTasks[len(tc.expectedTasks)-1].Description {
				t.Errorf("Expected input '%s' to match expected task description '%s'", tc.input, tc.expectedTasks[len(tc.expectedTasks)-1].Description)
			}

			if diff := cmp.Diff(tc.expectedTasks, tm.GetTasks()); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestConcurrentAddTask проверяет, что одновременные вызовы AddTask безопасны
func TestConcurrentAddTask(t *testing.T) {
	tm := NewTaskManager(&strings.Builder{})
	const numGoroutines = 100
	const tasksPerGoroutine = 10

	var wg sync.WaitGroup

	// Запускаем множество горутин, каждая добавляет задачи
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < tasksPerGoroutine; j++ {
				taskDesc := fmt.Sprintf("Task %d-%d", goroutineID, j)
				tm.AddTask(taskDesc)
			}
		}(i)
	}

	wg.Wait()

	// Проверяем результат
	tasks := tm.GetTasks()
	expectedTaskCount := numGoroutines * tasksPerGoroutine

	if len(tasks) != expectedTaskCount {
		t.Errorf("Expected %d tasks, got %d", expectedTaskCount, len(tasks))
	}

	// Проверяем, что все ID уникальны
	idMap := make(map[int]bool)
	for _, task := range tasks {
		if idMap[task.ID] {
			t.Errorf("Duplicate ID found: %d", task.ID)
		}
		idMap[task.ID] = true
	}
}

// TestUpdateTaskStatus tests UpdateTaskStatus with various scenarios:
// completion/incompletion, valid/invalid IDs, and edge cases.
func TestUpdateTaskStatus(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name          string
		taskId        int
		done          bool
		initialTasks  []Task
		expectedTasks []Task
		expectedErr   error
	}{
		{
			name:          "Mark task done in one-task list",
			taskId:        1,
			done:          true,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Mark task undone in one-task list",
			taskId:        1,
			done:          false,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Mark task done in empty list",
			taskId:        1,
			done:          true,
			initialTasks:  []Task{},
			expectedTasks: []Task{},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Mark specific task done in multiple tasks",
			taskId:        3,
			done:          true,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Mark specific task undone in multiple tasks",
			taskId:        4,
			done:          false,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Mark non-existence task done",
			taskId:        8,
			done:          true,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Mark non-existence task undone",
			taskId:        8,
			done:          true,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Mark already completed task",
			taskId:        1,
			done:          true,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Mark incompleted task undone",
			taskId:        1,
			done:          false,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Mark task done with negative ID",
			taskId:        -1,
			done:          true,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Mark task done with zero ID",
			taskId:        0,
			done:          true,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ==== ACT ====
			tm := NewTaskManager(&strings.Builder{})
			tm.SetTasks(tc.initialTasks)
			actualErr := tm.UpdateTaskStatus(tc.taskId, tc.done)

			// === ASSERT ===
			if tc.expectedErr != actualErr {
				t.Errorf("Expected error: '%v', got '%v'", tc.expectedErr, actualErr)
			}

			if diff := cmp.Diff(tc.expectedTasks, tm.GetTasks()); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestClearTaskDescription(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name          string
		taskId        int
		initialTasks  []Task
		expectedTasks []Task
		expectedErr   error
	}{
		{
			name:          "Clear task description in one-task list",
			taskId:        1,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Clear task description in empty list",
			taskId:        1,
			initialTasks:  []Task{},
			expectedTasks: []Task{},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Clear specific task description in multiple tasks",
			taskId:        3,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Clear non-existence task description",
			taskId:        8,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Clear task description with negative ID",
			taskId:        -1,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Clear task description with zero ID",
			taskId:        0,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := NewTaskManager(&strings.Builder{})
			tm.SetTasks(tc.initialTasks)

			// ==== ACT ====
			actualErr := tm.ClearDescription(tc.taskId)

			// === ASSERT ===
			if tc.expectedErr != actualErr {
				t.Errorf("Expected error: '%v', got '%v'", tc.expectedErr, actualErr)
			}

			if diff := cmp.Diff(tc.expectedTasks, tm.GetTasks()); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFormatTask(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name         string
		task         Task
		expectedTask string
	}{
		{
			name:         "Incomplete task",
			task:         Task{ID: 1, Description: "task 1", Done: false},
			expectedTask: "[  ] ID: 1, Description: task 1",
		},
		{
			name:         "Complete task",
			task:         Task{ID: 1, Description: "task 1", Done: true},
			expectedTask: "[✓ ] ID: 1, Description: task 1",
		},
		{
			name:         "Task with empty description",
			task:         Task{ID: 1, Description: "", Done: true},
			expectedTask: "[✓ ] ID: 1, Description: ",
		},
		{
			name:         "Task with zero ID",
			task:         Task{ID: 0, Description: "task 1", Done: true},
			expectedTask: "[✓ ] ID: 0, Description: task 1",
		},
		{
			name:         "Task with negative ID",
			task:         Task{ID: -1, Description: "task 1", Done: true},
			expectedTask: "[✓ ] ID: -1, Description: task 1",
		},
		{
			name:         "Task with description with special characters",
			task:         Task{ID: 1, Description: "#@`[]$%^*", Done: true},
			expectedTask: "[✓ ] ID: 1, Description: #@`[]$%^*",
		},
		{
			name:         "Task with description with spaces only",
			task:         Task{ID: 1, Description: "      ", Done: true},
			expectedTask: "[✓ ] ID: 1, Description:       ",
		},
		{
			name:         "Task with very long description",
			task:         Task{ID: 1, Description: "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.", Done: true},
			expectedTask: "[✓ ] ID: 1, Description: Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.",
		},
		{
			name:         "Task with very large ID",
			task:         Task{ID: 1111111111111111111, Description: "task 1", Done: true},
			expectedTask: "[✓ ] ID: 1111111111111111111, Description: task 1",
		},
		{
			name:         "Task with Unicode characters",
			task:         Task{ID: 1, Description: "Buy 🍞 and 🥛", Done: true},
			expectedTask: "[✓ ] ID: 1, Description: Buy 🍞 and 🥛",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			fTask := FormatTask(tc.task)

			// ====Assert====
			if diff := cmp.Diff(tc.expectedTask, fTask); diff != "" {
				t.Errorf("Task mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUpdateTaskDescription(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name          string
		id            int
		description   string
		initialTasks  []Task
		expectedTasks []Task
		expectedErr   error
	}{
		{
			name:          "Existent task in one-task list",
			id:            1,
			description:   "new task 1",
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "new task 1", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Completed task in multiple tasks list",
			id:            3,
			description:   "new task 3",
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "new task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Non-existent task in multiple tasks list",
			id:            5,
			description:   "new task 5",
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Task in empty task list",
			id:            1,
			description:   "new task 1",
			initialTasks:  []Task{},
			expectedTasks: []Task{},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Empty description",
			id:            1,
			description:   "",
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Zero ID",
			id:            0,
			description:   "new task 0",
			initialTasks:  []Task{{ID: 0, Description: "task 0", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTasks: []Task{{ID: 0, Description: "new task 0", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Negative ID",
			id:            -1,
			description:   "new task -1",
			initialTasks:  []Task{{ID: -1, Description: "task -1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTasks: []Task{{ID: -1, Description: "new task -1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Very large ID",
			id:            1111111111111111111,
			description:   "new task 1111111111111111111",
			initialTasks:  []Task{{ID: 1111111111111111111, Description: "task 1111111111111111111", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTasks: []Task{{ID: 1111111111111111111, Description: "new task 1111111111111111111", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "New task description with special characters",
			id:            1,
			description:   "#@`[]$%^*",
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "#@`[]$%^*", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "New task description with Unicode characters",
			id:            2,
			description:   "Buy 🍞 and 🥛",
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "Buy 🍞 and 🥛", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Very long description update",
			id:            3,
			description:   "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.",
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.", Done: true}},
			expectedErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := NewTaskManager(&strings.Builder{})
			tm.SetTasks(tc.initialTasks)

			// ====Act====
			err := tm.UpdateTaskDescription(tc.id, tc.description)

			// ====Assert====
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected %v, got %v", tc.expectedErr, err)
			}

			if len(tm.GetTasks()) != len(tc.expectedTasks) {
				t.Errorf("Expected task list length '%d', got '%d'", len(tc.expectedTasks), len(tm.GetTasks()))
			}

			if diff := cmp.Diff(tm.GetTasks(), tc.expectedTasks); diff != "" {
				t.Errorf("Struct mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetTaskByID(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name         string
		id           int
		initialTasks []Task
		expectedTask Task
		expectedErr  error
	}{
		{
			name:         "Existent task in one-task list",
			id:           1,
			initialTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTask: Task{ID: 1, Description: "task 1"},
			expectedErr:  nil,
		},
		{
			name:         "Completed task in multiple tasks list",
			id:           3,
			initialTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTask: Task{ID: 3, Description: "task 3", Done: true},
			expectedErr:  nil,
		},
		{
			name:         "Non-existent task in multiple tasks list",
			id:           4,
			initialTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedTask: Task{},
			expectedErr:  ErrTaskNotFound,
		},
		{
			name:         "Non-existent task in empty tasks list",
			id:           1,
			initialTasks: []Task{},
			expectedTask: Task{},
			expectedErr:  ErrTaskNotFound,
		},
		{
			name:         "Zero ID",
			id:           0,
			initialTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTask: Task{},
			expectedErr:  ErrTaskNotFound,
		},
		{
			name:         "Negative ID",
			id:           -1,
			initialTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTask: Task{},
			expectedErr:  ErrTaskNotFound,
		},
		{
			name:         "Very large ID",
			id:           1111111111111111111,
			initialTasks: []Task{{ID: 1111111111111111111, Description: "task 1111111111111111111", Done: false}},
			expectedTask: Task{ID: 1111111111111111111, Description: "task 1111111111111111111"},
			expectedErr:  nil,
		},
		{
			name:         "Existent task with very long description",
			id:           3,
			initialTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.", Done: true}},
			expectedTask: Task{ID: 3, Description: "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.", Done: true},
			expectedErr:  nil,
		},
		{
			name:         "Existent task with empty description",
			id:           3,
			initialTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "", Done: true}},
			expectedTask: Task{ID: 3, Done: true},
			expectedErr:  nil,
		},
		{
			name:         "Existent task with Unicode and special characters",
			id:           3,
			initialTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "Buy 🍞 and 🥛#@`[]$%^*", Done: true}},
			expectedTask: Task{ID: 3, Description: "Buy 🍞 and 🥛#@`[]$%^*", Done: true},
			expectedErr:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := NewTaskManager(&strings.Builder{})
			tm.SetTasks(tc.initialTasks)

			// ====Act====
			strTask, err := tm.GetTaskByID(tc.id)

			// ====Assert====
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected '%v', got '%v'", tc.expectedErr, err)
			}

			if diff := cmp.Diff(tc.expectedTask, strTask); diff != "" {
				t.Errorf("String mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name          string
		taskId        int
		initialTasks  []Task
		expectedTasks []Task
		expectedErr   error
	}{
		{
			name:          "Delete first task",
			taskId:        1,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Delete last task",
			taskId:        4,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Delete non-existense task",
			taskId:        7,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Delete task in empty list",
			taskId:        1,
			initialTasks:  []Task{},
			expectedTasks: []Task{},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Delete task with negative ID",
			taskId:        -1,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Delete task with zero ID",
			taskId:        0,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Delete task with max int value ID",
			taskId:        9223372036854775807,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 9223372036854775807, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tm := NewTaskManager(&strings.Builder{})
			tm.SetTasks(tc.initialTasks)

			// ==== ACT ====
			err := tm.DeleteTask(tc.taskId)

			// ==== ASSERT ====
			if err != tc.expectedErr {
				t.Errorf("Expected %v, got %v", tc.expectedErr, err)
			}

			if diff := cmp.Diff(tc.expectedTasks, tm.GetTasks()); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
