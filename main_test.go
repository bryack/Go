package main

import (
	"testing"
)

// Тест для функции addTask
// TestAddTaskToEmptySlice проверяет добавление задачи в пустой срез
func TestAddTaskToEmptySlice(t *testing.T) {
	// ===Arrange===
	// Сбрасываем lastId для изолированности теста
	originalLastId := lastId
	defer func() {
		lastId = originalLastId
	}() // Восстанавливаем после теста
	lastId = 0

	// Создаем пустой срез задач
	var tasks []Task
	description := "Exploring testing"

	// ===Act===
	resultId := addTask(&tasks, description)

	// ===Assert===
	expectedId := 1
	if resultId != expectedId {
		t.Errorf("Expected ID %d, but got %d", expectedId, resultId)
	}

	// Проверяем, что в срезе теперь одна задача
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task in slice, but got %d", len(tasks))
	}

	// Проверяем свойства добавленной задачи
	if tasks[0].ID != resultId {
		t.Errorf("Expected task ID %d, but got %d", resultId, tasks[0].ID)
	}

	if tasks[0].Description != description {
		t.Errorf("Expected task description %s, but got %s", description, tasks[0].Description)
	}

	if tasks[0].Done != false {
		t.Errorf("Expected tasks to be not done, but it was marked as done")
	}
}

// Тест для функции addTask
// TestAddTaskToNonEmptySlice проверяет добавление задачи в непустой срез
func TestAddTaskToNonEmptySlice(t *testing.T) {
	// ===Arrange===
	originalLastId := lastId
	defer func() {
		lastId = originalLastId
	}()
	lastId = 1

	tasks := []Task{{ID: 1, Description: "First task", Done: false}}
	originalLength := len(tasks)
	description := "Second task"

	// ===Act===
	resultId := addTask(&tasks, description)

	// ===Assert===
	expectedId := 2
	if resultId != expectedId {
		t.Errorf("Expected ID %d, but got %d", expectedId, resultId)
	}

	// Проверяем, что количество задач увеличилось на 1
	if len(tasks) != originalLength+1 {
		t.Errorf("Expected %d tasks, but got %d", originalLength+1, len(tasks))
	}

	// Проверяем, что новая задача добавлена в конец
	newTask := tasks[len(tasks)-1]
	if newTask.ID != expectedId {
		t.Errorf("Expected new task ID %d, but got %d", expectedId, newTask.ID)
	}

	if newTask.Description != description {
		t.Errorf("Expected task descrition %s, but got %s", description, newTask.Description)
	}
}

// Тест для функции addTask
// TestAddTaskWithEmptyDescription проверяет добавление задачи с пустым описанием
// ВАЖНО: Этот тест проверяет поведение функции addTask на уровне unit-теста.
// В реальном приложении пустые описания фильтруются функцией readInput,
// но мы тестируем addTask изолированно от других компонентов.
// Это помогает понять, как функция ведет себя с любыми входными данными.
func TestAddTaskWithEmptyDescription(t *testing.T) {
	// ===Arrange===
	originalLastId := lastId
	defer func() {
		lastId = originalLastId
	}()

	lastId = 0

	var tasks []Task
	emptyDescription := ""

	// ===Act===
	resultId := addTask(&tasks, emptyDescription)

	// ===Assert===
	// Проверяем, что функция не падает с пустой строкой
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, but got %d", len(tasks))
	}

	// Проверяем, что пустое описание сохраняется как есть
	if tasks[0].Description != emptyDescription {
		t.Errorf("Expected task description %s, but got %s", emptyDescription, tasks[0].Description)
	}

	if resultId != 1 {
		t.Errorf("Expected ID 1, but got %d", resultId)
	}

	// Дополнительная проверка: задача должна быть не выполненной
	if tasks[0].Done != false {
		t.Errorf("Expected task not to be done, but it was marked as done")
	}
}

// Тест для функции addTask
// TestAddTaskWithLongDescription проверяет добавление задачи с длинным описанием
// ВАЖНО: Этот тест проверяет поведение функции addTask на уровне unit-теста.
// В реальном приложении слишком длинные описания фильтруются функцией readInput,
// но мы тестируем addTask изолированно от других компонентов.
// Это помогает понять, как функция ведет себя с любыми входными данными.
func TestAddTaskWithLongDescription(t *testing.T) {
	// ===Arrange===
	originalLastId := lastId
	defer func() {
		lastId = originalLastId
	}()
	lastId = 0

	var tasks []Task
	longDescription := "Это очень длинное описание задачи, которое содержит много текста и проверяет, может ли наша функция корректно работать с большими строками"

	// ===Act===
	resultId := addTask(&tasks, longDescription)

	// ===Assert===
	if len(tasks) != 1 {
		t.Errorf("Expected 1 task, but got %d", len(tasks))
	}
	if resultId != 1 {
		t.Errorf("Expected ID 1, but got %d", resultId)
	}

	if tasks[0].Description != longDescription {
		t.Errorf("Long description was not saved correctly. Expected task description %s, but got %s", longDescription, tasks[0].Description)
	}
}

// Тест для функции addTask
// TestAddMultipleTasks проверяет последовательное добавление нескольких задач
func TestAddMultipleTasks(t *testing.T) {
	// ===Arrange===
	originalLastId := lastId
	defer func() {
		lastId = originalLastId
	}()
	lastId = 0

	var tasks []Task
	description := []string{"Задача 1", "Задача 2", "Задача 3"}

	// Act & Assert - проверяем каждое добавление
	for i, desc := range description {
		expectedId := i + 1
		resultId := addTask(&tasks, desc)

		if resultId != expectedId {
			t.Errorf("Step %d: Expected ID %d, but got %d", i+1, expectedId, resultId)
		}

		if len(tasks) != i+1 {
			t.Errorf("Step %d: Expected %d tasks, but got %d", i+1, i+1, len(tasks))
		}
	}

	// Проверяем финальное состояние
	if len(tasks) != len(description) {
		t.Errorf("Expected %d total tasks, but got %d", len(description), len(tasks))
	}
}

// Тест для функции addTask
// TestLastIdIncrement проверяет, что lastId корректно увеличивается
func TestLastIdIncrement(t *testing.T) {
	// ===Arrange===
	originalLastId := lastId
	defer func() {
		lastId = originalLastId
	}()
	startId := 5
	lastId = startId
	var tasks []Task

	// ===Act===
	resultId := addTask(&tasks, "Test ID")

	// ===Arrange===
	expectedId := startId + 1
	if resultId != expectedId {
		t.Errorf("Expected ID %d, but got %d", expectedId, resultId)
	}

	if lastId != expectedId {
		t.Errorf("Expected lastId to be %d, but got %d", expectedId, lastId)
	}
}

// Тест для функции addTask
// TestAddTaskWithNilSlice проверяет поведение функции `addTask`, когда ей передаётся `nil` вместо указателя на срез задач
func TestAddTaskWithNilSlice(t *testing.T) {
	// ===Arrange===
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when passing nil pointer, but no panic occurred")
		}
	}()

	// Act
	addTask(nil, "This should panic")

	// Если мы дошли до этой строки, значит паники не было
	t.Errorf("Function should have panicked but didn't")
}
