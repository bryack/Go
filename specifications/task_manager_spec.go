package specifications

import (
	"myproject/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TaskManager interface {
	Register(email, password string) error
	Login(email, password string) (token string, err error)
	CreateTask(token, description string) (taskID int, err error)
	GetTasks(token string) ([]domain.Task, error)
}

func TaskManagerSpecification(t testing.TB, tm TaskManager) {
	err := tm.Register("test@email.com", "password123")
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	token, err := tm.Login("test@email.com", "password123")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	taskID, err := tm.CreateTask(token, "task 1")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if taskID == 0 {
		t.Error("expected non-zero task ID")
	}

	tasks, err := tm.GetTasks(token)
	if err != nil {
		t.Fatalf("failed to get tasks: %v", err)
	}

	assert.Len(t, tasks, 1)
	assert.Equal(t, "task 1", tasks[0].Description)
}

func TaskManagerSpecification_Isolation(t testing.TB, tm TaskManager) {
	err := tm.Register("first_user@email.com", "password123")
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	firstToken, err := tm.Login("first_user@email.com", "password123")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	err = tm.Register("second_user@email.com", "password123")
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	secondToken, err := tm.Login("second_user@email.com", "password123")
	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	taskID, err := tm.CreateTask(firstToken, "task 1")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	if taskID == 0 {
		t.Error("expected non-zero task ID")
	}

	tasks, err := tm.GetTasks(firstToken)
	if err != nil {
		t.Fatalf("failed to get tasks: %v", err)
	}

	assert.Len(t, tasks, 1)
	assert.Equal(t, "task 1", tasks[0].Description)

	tasks, err = tm.GetTasks(secondToken)
	assert.NoError(t, err)
	assert.Len(t, tasks, 0)
}
