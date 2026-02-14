package testhelpers

import (
	"fmt"
	"myproject/adapters/storage"
)

type StubTaskStore struct {
	Tasks            map[int]string
	CreateCall       []int
	TasksTable       []storage.Task
	UpdateTaskCalled int
}

func (s *StubTaskStore) GetTaskByID(id int, userID int) (task storage.Task, err error) {
	t, ok := s.Tasks[id]
	if !ok {
		return storage.Task{}, fmt.Errorf("Task not found")
	}
	return storage.Task{ID: id, Description: t}, nil
}

func (s *StubTaskStore) CreateTask(task storage.Task, userID int) (int, error) {
	s.CreateCall = append(s.CreateCall, task.ID)
	return task.ID, nil
}

func (s *StubTaskStore) LoadTasks(userID int) ([]storage.Task, error) {
	return s.TasksTable, nil
}

func (s *StubTaskStore) UpdateTask(task storage.Task, userID int) error {
	s.UpdateTaskCalled++
	s.Tasks[task.ID] = task.Description
	return nil
}

func (s *StubTaskStore) DeleteTask(id int, userID int) error {
	delete(s.Tasks, id)
	return nil
}

func (s *StubTaskStore) Close() error {
	return nil
}
