package task

import (
	"errors"
	"myproject/storage"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func seedTask(t *testing.T, s storage.Storage, task storage.Task) {
	_, err := s.CreateTask(storage.Task{
		Description: task.Description,
		Done:        task.Done,
	},
	)

	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
}

func TestAddTask(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name          string
		input         string
		initialTasks  []storage.Task
		expectedTasks []storage.Task
		expectedErr   error
	}{
		// Valid inputs
		{
			name:          "Add task to empty list",
			input:         "task 1",
			initialTasks:  []storage.Task{},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Add task (description with spaces) to non-empty list",
			input:         " task 2 ",
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: " task 2 ", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Add long description",
			input:         "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec qu",
			initialTasks:  []storage.Task{},
			expectedTasks: []storage.Task{{ID: 1, Description: "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec qu", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Add task with Unicode and special characters",
			input:         "Купить 🥛 & 🍞 в магазине \"Пятёрочка\"",
			initialTasks:  []storage.Task{},
			expectedTasks: []storage.Task{{ID: 1, Description: "Купить 🥛 & 🍞 в магазине \"Пятёрочка\"", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Add empty description",
			input:         "",
			initialTasks:  []storage.Task{},
			expectedTasks: []storage.Task{{ID: 1, Description: "", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Add description with spaces only",
			input:         "         ",
			initialTasks:  []storage.Task{},
			expectedTasks: []storage.Task{{ID: 1, Description: "         ", Done: false}},
			expectedErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := storage.NewDatabaseStorage(":memory:")
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			tm := NewTaskManager(s, &strings.Builder{})
			for _, it := range tc.initialTasks {
				seedTask(t, s, it)
			}

			// ==== ACT ====
			id, err := tm.AddTask(tc.input)

			// === ASSERT ===
			if err != tc.expectedErr {
				t.Errorf("Expected error: %v, got: %v", tc.expectedErr, err)
			}

			if tc.expectedErr == nil && id != tc.expectedTasks[len(tc.expectedTasks)-1].ID {
				t.Errorf("Expected ID: %d, got: %d", tc.expectedTasks[len(tc.expectedTasks)-1].ID, id)
			}

			lt, err := s.LoadTasks()
			if err != nil {
				t.Fatalf("Failed to load tasks: %v", err)
			}
			if diff := cmp.Diff(tc.expectedTasks, lt); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// // TestConcurrentAddTask проверяет, что одновременные вызовы AddTask безопасны
// func TestConcurrentAddTask(t *testing.T) {
// 	tm := NewTaskManager(&strings.Builder{})
// 	const numGoroutines = 100
// 	const tasksPerGoroutine = 10

// 	var wg sync.WaitGroup

// 	// Запускаем множество горутин, каждая добавляет задачи
// 	for i := 0; i < numGoroutines; i++ {
// 		wg.Add(1)
// 		go func(goroutineID int) {
// 			defer wg.Done()
// 			for j := 0; j < tasksPerGoroutine; j++ {
// 				taskDesc := fmt.Sprintf("Task %d-%d", goroutineID, j)
// 				tm.AddTask(taskDesc)
// 			}
// 		}(i)
// 	}

// 	wg.Wait()

// 	// Проверяем результат
// 	tasks := tm.GetTasks()
// 	expectedTaskCount := numGoroutines * tasksPerGoroutine

// 	if len(tasks) != expectedTaskCount {
// 		t.Errorf("Expected %d tasks, got %d", expectedTaskCount, len(tasks))
// 	}

// 	// Проверяем, что все ID уникальны
// 	idMap := make(map[int]bool)
// 	for _, task := range tasks {
// 		if idMap[task.ID] {
// 			t.Errorf("Duplicate ID found: %d", task.ID)
// 		}
// 		idMap[task.ID] = true
// 	}
// }

// TestUpdateTaskStatus tests UpdateTaskStatus with various scenarios:
// completion/incompletion, valid/invalid IDs, and edge cases.
func TestUpdateTaskStatus(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name          string
		taskId        int
		done          bool
		initialTasks  []storage.Task
		expectedTasks []storage.Task
		expectedErr   error
	}{
		{
			name:          "Mark task done in one-task list",
			taskId:        1,
			done:          true,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Mark task undone in one-task list",
			taskId:        1,
			done:          false,
			initialTasks:  []storage.Task{{Description: "task 1", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Mark task done in empty list",
			taskId:        1,
			done:          true,
			initialTasks:  []storage.Task{},
			expectedTasks: []storage.Task{},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Mark specific task done in multiple tasks",
			taskId:        3,
			done:          true,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: false}, {Description: "task 4", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Mark specific task undone in multiple tasks",
			taskId:        4,
			done:          false,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: false}, {Description: "task 4", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Mark non-existence task done",
			taskId:        8,
			done:          true,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: false}, {Description: "task 4", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Mark non-existence task undone",
			taskId:        8,
			done:          true,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: false}, {Description: "task 4", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Mark already completed task",
			taskId:        1,
			done:          true,
			initialTasks:  []storage.Task{{Description: "task 1", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Mark incompleted task undone",
			taskId:        1,
			done:          false,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Mark task done with negative ID",
			taskId:        -1,
			done:          true,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Mark task done with zero ID",
			taskId:        0,
			done:          true,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   storage.ErrTaskNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := storage.NewDatabaseStorage(":memory:")
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			tm := NewTaskManager(s, &strings.Builder{})
			for _, it := range tc.initialTasks {
				seedTask(t, s, it)
			}

			// ==== ACT ====
			actualErr := tm.UpdateTaskStatus(tc.taskId, tc.done)

			// === ASSERT ===
			if !errors.Is(actualErr, tc.expectedErr) {
				t.Errorf("Expected error: '%v', got '%v'", tc.expectedErr, actualErr)
			}

			lt, err := s.LoadTasks()
			if err != nil {
				t.Fatalf("Failed to load tasks: %v", err)
			}

			if diff := cmp.Diff(tc.expectedTasks, lt); diff != "" {
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
		initialTasks  []storage.Task
		expectedTasks []storage.Task
		expectedErr   error
	}{
		{
			name:          "Clear task description in one-task list",
			taskId:        1,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Clear task description in empty list",
			taskId:        1,
			initialTasks:  []storage.Task{},
			expectedTasks: []storage.Task{},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Clear specific task description in multiple tasks",
			taskId:        3,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: false}, {Description: "task 4", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Clear non-existence task description",
			taskId:        8,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: false}, {Description: "task 4", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: false}, {ID: 4, Description: "task 4", Done: false}},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Clear task description with negative ID",
			taskId:        -1,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Clear task description with zero ID",
			taskId:        0,
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}},
			expectedErr:   storage.ErrTaskNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := storage.NewDatabaseStorage(":memory:")
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			tm := NewTaskManager(s, &strings.Builder{})
			for _, it := range tc.initialTasks {
				seedTask(t, s, it)
			}

			// ==== ACT ====
			actualErr := tm.ClearDescription(tc.taskId)

			// === ASSERT ===
			if !errors.Is(actualErr, tc.expectedErr) {
				t.Errorf("Expected error: '%v', got '%v'", tc.expectedErr, actualErr)
			}

			lt, err := s.LoadTasks()
			if err != nil {
				t.Fatalf("Failed to load tasks: %v", err)
			}

			if diff := cmp.Diff(tc.expectedTasks, lt); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFormatTask(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name         string
		task         storage.Task
		expectedTask string
	}{
		{
			name:         "Incomplete task",
			task:         storage.Task{ID: 1, Description: "task 1", Done: false},
			expectedTask: "[  ] ID: 1, Description: task 1",
		},
		{
			name:         "Complete task",
			task:         storage.Task{ID: 1, Description: "task 1", Done: true},
			expectedTask: "[✓ ] ID: 1, Description: task 1",
		},
		{
			name:         "Task with empty description",
			task:         storage.Task{ID: 1, Description: "", Done: true},
			expectedTask: "[✓ ] ID: 1, Description: ",
		},
		{
			name:         "Task with zero ID",
			task:         storage.Task{ID: 0, Description: "task 1", Done: true},
			expectedTask: "[✓ ] ID: 0, Description: task 1",
		},
		{
			name:         "Task with negative ID",
			task:         storage.Task{ID: -1, Description: "task 1", Done: true},
			expectedTask: "[✓ ] ID: -1, Description: task 1",
		},
		{
			name:         "Task with description with special characters",
			task:         storage.Task{ID: 1, Description: "#@`[]$%^*", Done: true},
			expectedTask: "[✓ ] ID: 1, Description: #@`[]$%^*",
		},
		{
			name:         "Task with description with spaces only",
			task:         storage.Task{ID: 1, Description: "      ", Done: true},
			expectedTask: "[✓ ] ID: 1, Description:       ",
		},
		{
			name:         "Task with very long description",
			task:         storage.Task{ID: 1, Description: "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.", Done: true},
			expectedTask: "[✓ ] ID: 1, Description: Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.",
		},
		{
			name:         "Task with very large ID",
			task:         storage.Task{ID: 1111111111111111111, Description: "task 1", Done: true},
			expectedTask: "[✓ ] ID: 1111111111111111111, Description: task 1",
		},
		{
			name:         "Task with Unicode characters",
			task:         storage.Task{ID: 1, Description: "Buy 🍞 and 🥛", Done: true},
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
		initialTasks  []storage.Task
		expectedTasks []storage.Task
		expectedErr   error
	}{
		{
			name:          "Existent task in one-task list",
			id:            1,
			description:   "new task 1",
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}},
			expectedTasks: []storage.Task{{ID: 1, Description: "new task 1", Done: false}},
			expectedErr:   nil,
		},
		{
			name:          "Completed task in multiple tasks list",
			id:            3,
			description:   "new task 3",
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "new task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Non-existent task in multiple tasks list",
			id:            5,
			description:   "new task 5",
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Task in empty task list",
			id:            1,
			description:   "new task 1",
			initialTasks:  []storage.Task{},
			expectedTasks: []storage.Task{},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Empty description",
			id:            1,
			description:   "",
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Zero ID",
			id:            0,
			description:   "new task 0",
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Negative ID",
			id:            -1,
			description:   "new task -1",
			initialTasks:  []storage.Task{{ID: 1, Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "Very large ID",
			id:            1111111111111111111,
			description:   "new task 1111111111111111111",
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   storage.ErrTaskNotFound,
		},
		{
			name:          "New task description with special characters",
			id:            1,
			description:   "#@`[]$%^*",
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "#@`[]$%^*", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "New task description with Unicode characters",
			id:            2,
			description:   "Buy 🍞 and 🥛",
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "Buy 🍞 and 🥛", Done: false}, {ID: 3, Description: "task 3", Done: true}},
			expectedErr:   nil,
		},
		{
			name:          "Very long description update",
			id:            3,
			description:   "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.",
			initialTasks:  []storage.Task{{Description: "task 1", Done: false}, {Description: "task 2", Done: false}, {Description: "task 3", Done: true}},
			expectedTasks: []storage.Task{{ID: 1, Description: "task 1", Done: false}, {ID: 2, Description: "task 2", Done: false}, {ID: 3, Description: "Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.", Done: true}},
			expectedErr:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := storage.NewDatabaseStorage(":memory:")
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			tm := NewTaskManager(s, &strings.Builder{})
			for _, it := range tc.initialTasks {
				seedTask(t, s, it)
			}

			// ==== ACT ====
			actualErr := tm.UpdateTaskDescription(tc.id, tc.description)

			// === ASSERT ===
			if !errors.Is(actualErr, tc.expectedErr) {
				t.Errorf("Expected error: '%v', got '%v'", tc.expectedErr, actualErr)
			}

			lt, err := s.LoadTasks()
			if err != nil {
				t.Fatalf("Failed to load tasks: %v", err)
			}

			if diff := cmp.Diff(tc.expectedTasks, lt); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
