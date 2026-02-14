package application

import (
	"fmt"
	infraErrors "myproject/infrastructure/errors"
	"myproject/internal/domain"
	"myproject/validation"
)

type Service struct {
	store domain.Storage
}

func NewService(store domain.Storage) *Service {
	return &Service{store: store}
}

func (s *Service) UpdateTask(taskID, userID int, description *string, done *bool) (domain.Task, error) {
	if description == nil && done == nil {
		return domain.Task{}, infraErrors.ErrEmptyFieldsToUpdate
	}

	task, err := s.store.GetTaskByID(taskID, userID)
	if err != nil {
		return domain.Task{}, fmt.Errorf("failed to find task with id %d: %w", taskID, err)
	}

	if description != nil {
		desc := string(*description)
		desc, err = validation.ValidateTaskDescription(desc)
		if err != nil {
			return domain.Task{}, fmt.Errorf("failed to validate description for task with id %d: %w", taskID, err)
		}
		task.Description = desc
	}

	if done != nil {
		task.Done = *done
	}

	if err := s.store.UpdateTask(task, userID); err != nil {
		return domain.Task{}, fmt.Errorf("failed to update task with id %d: %w", taskID, err)
	}
	return task, nil
}

func (s *Service) CreateTask(description string, userID int) (domain.Task, error) {
	desc, err := validation.ValidateTaskDescription(description)
	if err != nil {
		return domain.Task{}, fmt.Errorf("failed to validate description: %w", err)
	}

	newTask := domain.Task{Description: desc, Done: false}
	id, err := s.store.CreateTask(newTask, userID)
	if err != nil {
		return domain.Task{}, fmt.Errorf("failed to create task: %w", err)
	}
	newTask.ID = id
	return newTask, nil
}
