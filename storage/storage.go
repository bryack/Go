package storage

import "myproject/task"

// Storage defines the interface for task persistence operations.
type Storage interface {
	LoadTasks() ([]task.Task, error)
	SaveTasks(tasks []task.Task) error
}
