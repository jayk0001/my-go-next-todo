package todo

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TodoService handles todo business logic
type TodoService struct {
	repo      Repository
	validator *ValidatorService
}

// NewTodoService creates a new todo service
func NewTodoService(db *pgxpool.Pool) *TodoService {
	return &TodoService{
		repo:      NewTodoRepository(db),
		validator: NewValidatorService(),
	}
}

// CreateTodo creates a new todo with validation
func (s *TodoService) CreateTodo(ctx context.Context, userID int, input CreateTodoInput) (*Todo, error) {
	// Validate Input
	if err := s.validator.ValidateTodoInput(ctx, input); err != nil {
		return nil, fmt.Errorf("failed to validate input: %w", err)
	}

	// Create todo via repository
	todo, err := s.repo.Create(ctx, userID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new todo: %w", err)
	}

	return todo, nil
}

// GetTodo retrieves a todo by ID with ownership check
func (s *TodoService) GetTodo(ctx context.Context, todoID, userID int) (*Todo, error) {
	if todoID <= 0 {
		return nil, ErrInvalidTodoInput
	}

	todo, err := s.repo.GetByID(ctx, todoID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}

	return todo, nil
}

// GetUserTodos retrieves todos by userID with pagination and filtering
func (s *TodoService) GetUserTodos(ctx context.Context, userID int, filter TodoFilter) (*TodoListResponse, error) {
	// Validate and normalize filter
	normalizeFilter := s.normalizeFilter(filter)

	todos, err := s.repo.GetByUserID(ctx, userID, normalizeFilter)

	if err != nil {
		return nil, fmt.Errorf("failed to get todos: %w", err)
	}

	return todos, nil
}

// UpdateTodo updates a todo with validation and ownership check
func (s *TodoService) UpdateTodo(ctx context.Context, todoID, userID int, input UpdateTodoInput) (*Todo, error) {
	// Validate todoID
	if todoID <= 0 {
		return nil, ErrInvalidTodoInput
	}

	// Validate input
	if err := s.validator.ValidateUpdateInput(ctx, input); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if todo exists and user owns it
	_, err := s.repo.GetByID(ctx, todoID, userID)
	if err != nil {
		if err == ErrTodoNotFound {
			return nil, ErrTodoAccessDenied
		}
		return nil, fmt.Errorf("failed to validate todo ownership: %w", err)
	}

	// Update todo
	todo, err := s.repo.Update(ctx, todoID, userID, input)
	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return todo, nil
}

// DeleteTodo deletes a todo with ownership check
func (s *TodoService) DeleteTodo(ctx context.Context, todoID, userID int) error {
	// Validate todoID
	if todoID <= 0 {
		return ErrInvalidTodoInput
	}

	// Check if todo exists and user owns it
	_, err := s.repo.GetByID(ctx, todoID, userID)
	if err != nil {
		if err == ErrTodoNotFound {
			return ErrTodoAccessDenied
		}
		return fmt.Errorf("failed to verify todo ownership: %w", err)
	}

	// Delete todo
	err = s.repo.Delete(ctx, todoID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	return nil
}

// ToggleTodoComplete toggles the completed status of a todo
func (s *TodoService) ToggleTodoComplete(ctx context.Context, todoID, userID int) (*Todo, error) {
	// Validate todoID
	if todoID <= 0 {
		return nil, ErrInvalidTodoInput
	}

	// check if todo exists and user owns it
	_, err := s.repo.GetByID(ctx, todoID, userID)
	if err != nil {
		if err == ErrTodoNotFound {
			return nil, ErrTodoAccessDenied
		}
		return nil, fmt.Errorf("failed to verity todo ownership: %w", err)
	}

	// Toggle completion
	todo, err := s.repo.ToggleComplete(ctx, todoID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle complete todo: %w", err)
	}

	return todo, nil
}

// GetUserTodoStats returns statistics about user's todos
func (s *TodoService) GetUserTodoStats(ctx context.Context, userID int) (*TodoStats, error) {
	// Get total count
	total, err := s.repo.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count todos: %w", err)
	}

	// Get completed count
	completedFilter := TodoFilter{Completed: &[]bool{true}[0]}
	completed, err := s.repo.GetByUserID(ctx, userID, completedFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count completed todos: %w", err)
	}

	// get pending count
	pendingFilter := TodoFilter{Completed: &[]bool{false}[0]}
	pending, err := s.repo.GetByUserID(ctx, userID, pendingFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending todos: %w", err)
	}

	return &TodoStats{
		Total:     total,
		Completed: len(completed.Todos),
		Pending:   len(pending.Todos),
	}, nil

}

// BatchUpdateTodos updates multiple todos at once (bonus feature)
func (s *TodoService) BatchUpdateTodos(ctx context.Context, userID int, todoIDs []int, input UpdateTodoInput) ([]*Todo, error) {
	// Validate todo ids
	if len(todoIDs) == 0 {
		return []*Todo{}, nil
	}

	// Validate Input
	if err := s.validator.ValidateUpdateInput(ctx, input); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var updatedTodos []*Todo
	var errors []string

	for _, todoID := range todoIDs {
		todo, err := s.UpdateTodo(ctx, todoID, userID, input)
		if err != nil {
			errors = append(errors, fmt.Sprintf("todo %d: %v", todoID, err))
			continue
		}
		updatedTodos = append(updatedTodos, todo)
	}

	if len(errors) > 0 {
		return updatedTodos, fmt.Errorf("batch update errors: %s", strings.Join(errors, "; "))
	}

	return updatedTodos, nil
}

// normalizeFilter validates and normalizes filter parameters
func (s *TodoService) normalizeFilter(filter TodoFilter) TodoFilter {
	normalized := filter

	// Set default pagination if not provided
	if normalized.Limit <= 0 {
		normalized.Limit = 20
	}

	// Maximum limit to prevent performance issues
	if normalized.Limit > 100 {
		normalized.Limit = 100
	}

	if normalized.Offset < 0 {
		normalized.Offset = 0
	}

	// Normalize search term
	if normalized.Search != nil {
		trimmed := strings.TrimSpace(*normalized.Search)
		if trimmed == "" {
			normalized.Search = nil
		} else {
			normalized.Search = &trimmed
		}
	}

	return normalized
}
