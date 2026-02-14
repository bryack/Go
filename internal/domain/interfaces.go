package domain

// Storage defines the interface for task persistence operations.
type Storage interface {
	LoadTasks(userID int) ([]Task, error)
	GetTaskByID(id int, userID int) (task Task, err error)
	CreateTask(task Task, userID int) (int, error)
	UpdateTask(task Task, userID int) error
	DeleteTask(id int, userID int) error
	Close() error
}
