package specifications

import "testing"

type TaskManager interface {
	Register(email, password string) error
	Login(email, password string) (token string, err error)
	CreateTask(token, description string) (taskID int, err error)
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
}
