package testhelpers

import (
	infraErrors "myproject/infrastructure/errors"
	"myproject/internal/domain"
)

type StubTaskStore struct {
	Tasks            map[int]string
	CreateCall       []int
	TasksTable       []domain.Task
	UpdateTaskCalled int
}

func (s *StubTaskStore) GetTaskByID(id int, userID int) (task domain.Task, err error) {
	t, ok := s.Tasks[id]
	if !ok {
		return domain.Task{}, infraErrors.ErrTaskNotFound
	}
	return domain.Task{ID: id, Description: t}, nil
}

func (s *StubTaskStore) CreateTask(task domain.Task, userID int) (int, error) {
	s.CreateCall = append(s.CreateCall, task.ID)
	return task.ID, nil
}

func (s *StubTaskStore) LoadTasks(userID int) ([]domain.Task, error) {
	return s.TasksTable, nil
}

func (s *StubTaskStore) UpdateTask(task domain.Task, userID int) error {
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
