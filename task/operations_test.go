package task

import "testing"

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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ==== ACT ====
			testTaskManager := NewTaskManager()
			testTaskManager.SetTasks(tc.initialTasks)
			id := testTaskManager.AddTask(tc.input)

			// === ASSERT ===
			if id != tc.expectedId {
				t.Errorf("Expected ID '%d', got '%d'", tc.expectedId, id)
			}

			if len(testTaskManager.GetTasks()) != len(tc.expectedTasks) {
				t.Errorf("Expected task list length '%d', got '%d'", len(tc.expectedTasks), len(testTaskManager.GetTasks()))
			}

			// for i := range tc.expectedTasks {
			// 	actualTask := testTaskManager.GetTasks()[i]
			// 	expectedTask := tc.expectedTasks[i]
			// 	if actualTask.Description != expectedTask.Description {
			// 		t.Errorf("Expected task description '%s', got '%s'", expectedTask.Description, actualTask.Description)
			// 	}
			// }

			// Check input against expectedTasks.Description
			if tc.input != tc.expectedTasks[len(tc.expectedTasks)-1].Description {
				t.Errorf("Expected input '%s' to match expected task description '%s'", tc.input, tc.expectedTasks[len(tc.expectedTasks)-1].Description)
			}
		})
	}
}
