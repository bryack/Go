package application

import (
	"fmt"
	"myproject/adapters/storage"
	"myproject/validation"
)

type Service struct {
	store storage.Storage
}

func NewService(store storage.Storage) *Service {
	return &Service{store: store}
}

func (s *Service) UpdateTask(taskID, userID int, description *string, done *bool) (storage.Task, error) {
	if description == nil && done == nil {
		return storage.Task{}, fmt.Errorf("at least one field must be provided for update")
	}

	task, err := s.store.GetTaskByID(taskID, userID)
	if err != nil {
		return storage.Task{}, fmt.Errorf("failed to find task with id %d: %w", taskID, err)
	}

	if description != nil {
		desc := string(*description)
		desc, err = validation.ValidateTaskDescription(desc)
		if err != nil {
			return storage.Task{}, fmt.Errorf("failed to validate description for task with id %d: %w", taskID, err)
		}
		task.Description = desc
	}

	if done != nil {
		task.Done = *done
	}

	if err := s.store.UpdateTask(task, userID); err != nil {
		return storage.Task{}, fmt.Errorf("failed to update task with id %d: %w", taskID, err)
	}
	return task, nil
}
