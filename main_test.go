package main

import (
	"testing"
)

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
