package domain

// Task represents a single task with ID, description, and completion status.
type Task struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Done        bool   `json:"done"`
}
