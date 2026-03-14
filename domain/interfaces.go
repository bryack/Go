package domain

import "context"

type TaskService interface {
	CreateTask(ctx context.Context, description string, userID int) (Task, error)
	UpdateTask(ctx context.Context, taskID, userID int, description *string, done *bool) (Task, error)
	GetTasks(ctx context.Context, userID int) ([]Task, error)
}

// Storage defines the interface for task persistence operations.
type Storage interface {
	LoadTasks(ctx context.Context, userID int) ([]Task, error)
	GetTaskByID(ctx context.Context, id int, userID int) (task Task, err error)
	CreateTask(ctx context.Context, task Task, userID int) (int, error)
	UpdateTask(ctx context.Context, task Task, userID int) error
	DeleteTask(ctx context.Context, id int, userID int) error
	Close(ctx context.Context) error
}

// UserStorage defines the interface for user persistence operations.
type UserStorage interface {
	CreateUser(ctx context.Context, email string, passwordHash string) (int, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id int) (*User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

type AppStorage interface {
	Storage
	UserStorage
}

type AuthService interface {
	Register(ctx context.Context, email, password string) (token string, err error)
	Login(ctx context.Context, email, password string) (token string, err error)
}

type TokenGenerator interface {
	GenerateToken(userID int) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type Claims struct {
	UserID int `json:"user_id"`
}
