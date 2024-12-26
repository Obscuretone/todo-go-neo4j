package models

// Task represents a task with optional parent ID.
type Task struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Completed bool    `json:"completed"`
	ParentID  *string `json:"parent_id"`
}
