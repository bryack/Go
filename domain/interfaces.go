package domain

type TaskService interface {
	CreateTask(description string, userID int) (Task, error)
	UpdateTask(taskID, userID int, description *string, done *bool) (Task, error)
	GetTasks(userID int) ([]Task, error)
}

// Storage defines the interface for task persistence operations.
type Storage interface {
	LoadTasks(userID int) ([]Task, error)
	GetTaskByID(id int, userID int) (task Task, err error)
	CreateTask(task Task, userID int) (int, error)
	UpdateTask(task Task, userID int) error
	DeleteTask(id int, userID int) error
	Close() error
}

// UserStorage defines the interface for user persistence operations.
type UserStorage interface {
	CreateUser(email string, passwordHash string) (int, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	EmailExists(email string) (bool, error)
}

type AppStorage interface {
	Storage
	UserStorage
}

type AuthService interface {
	Register(email, password string) (token string, err error)
	Login(email, password string) (token string, err error)
}

type TokenGenerator interface {
	GenerateToken(userID int) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type Claims struct {
	UserID int `json:"user_id"`
}
