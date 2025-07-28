package task

import (
	"fmt"
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
		expectedId    int
		expectedTasks []Task
	}{
		{
			name:          "Add task to empty list",
			input:         "task 1",
			initialTasks:  []Task{},
			expectedId:    1,
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
		},
		{
			name:          "Add task to non-empty list",
			input:         "task 2",
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedId:    2,
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}},
		},
		{
			name:          "Add empty description",
			input:         "",
			initialTasks:  []Task{},
			expectedId:    1,
			expectedTasks: []Task{{ID: 1, Description: "", Done: false}},
		},
		{
			name:          "Add long description",
			input:         "Это очень длинное описание задачи, которое содержит много текста и проверяет, может ли наша функция корректно работать с большими строками",
			initialTasks:  []Task{},
			expectedId:    1,
			expectedTasks: []Task{{ID: 1, Description: "Это очень длинное описание задачи, которое содержит много текста и проверяет, может ли наша функция корректно работать с большими строками", Done: false}},
		},
		{
			name:          "Add task with special characters",
			input:         "Купить молоко & хлеб в магазине \"Пятёрочка\"",
			initialTasks:  []Task{},
			expectedId:    1,
			expectedTasks: []Task{{ID: 1, Description: "Купить молоко & хлеб в магазине \"Пятёрочка\"", Done: false}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ==== ACT ====
			tm := NewTaskManager()
			tm.SetTasks(tc.initialTasks)
			actualId := tm.AddTask(tc.input)

			// === ASSERT ===
			if actualId != tc.expectedId {
				t.Errorf("Expected ID '%d', got '%d'", tc.expectedId, actualId)
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
	tm := NewTaskManager()
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

func TestMarkTaskDone(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name          string
		taskId        int
		initialTasks  []Task
		expectedTasks []Task
		expectedErr   error
	}{
		{
			name:          "Mark task done in one-task list",
			taskId:        1,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Mark task done in empty list",
			taskId:        1,
			initialTasks:  []Task{},
			expectedTasks: []Task{},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Mark specific task in multiple tasks",
			taskId:        3,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Mark non-existence task done",
			taskId:        8,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Mark already completed task",
			taskId:        1,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: true}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Mark task with negative ID",
			taskId:        -1,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
		{
			name:          "Mark task with zero ID",
			taskId:        0,
			initialTasks:  []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedTasks: []Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   ErrTaskNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ==== ACT ====
			tm := NewTaskManager()
			tm.SetTasks(tc.initialTasks)
			actualErr := tm.MarkTaskDone(tc.taskId)

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
			tm := NewTaskManager()
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
