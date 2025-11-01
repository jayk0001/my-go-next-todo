package todo

import (
	"context"
	"time"
)

// Todo represents a todo item in the database
type Todo struct {
	ID          int       `db:"id" json:"id"`
	UserID      int       `db:"user_id" json:"user_id"`
	Title       string    `db:"title" json:"title"`
	Description *string   `db:"description" json:"description"`
	Completed   bool      `db:"completed" json:"completed"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// CreateTodoInput represents input for creating a new todo
type CreateTodoInput struct {
	Title       string  `json:"title" validate:"required,max=500"`
	Description *string `json:"description,omitempty"`
}

// UpdateTodoInput represents input for updating a todo
type UpdateTodoInput struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,max=500"`
	Description *string `json:"description,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

// TodoFilter represents filtering options for querying todos
type TodoFilter struct {
	Completed *bool   `json:"completed,omitempty"`
	Search    *string `json:"search,omitempty"`
	Limit     int     `json:"limit,omitempty"`
	Offset    int     `json:"offset,omitempty"`
}

// TodoListResponse represents a paginated list of todos
type TodoListResponse struct {
	Todos   []*Todo `json:"todos"`
	Total   int     `json:"total"`
	Limit   int     `json:"limit"`
	Offset  int     `json:"offset"`
	HasMore bool    `json:"has_more"`
}

// TodoStats represents statistics about user's todos
type TodoStats struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Pending   int `json:"pending"`
}

// TodoRepository interface defines the contract for todo data operations
type Repository interface {
	Create(ctx context.Context, userID int, input CreateTodoInput) (*Todo, error)
	GetByID(ctx context.Context, todoID, userID int) (*Todo, error)
	GetByUserID(ctx context.Context, userID int, filter TodoFilter) (*TodoListResponse, error)
	Update(ctx context.Context, todoID, userID int, input UpdateTodoInput) (*Todo, error)
	Delete(ctx context.Context, todoID, userID int) error
	ToggleComplete(ctx context.Context, todoID, userID int) (*Todo, error)
	CountByUserID(ctx context.Context, userID int) (int, error)
}
