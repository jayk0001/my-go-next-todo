package todo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TodoRepository hanldes todo database operations
type TodoRepository struct {
	db *pgxpool.Pool
}

// NewTodoRepository created a new todo repository
func NewTodoRepository(db *pgxpool.Pool) *TodoRepository {
	return &TodoRepository{
		db: db,
	}
}

// Create creates a new todo in the database
func (r *TodoRepository) Create(ctx context.Context, userID int, input CreateTodoInput) (*Todo, error) {
	query := `
		INSERT INTO todos (user_id, title, description, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, user_id, title, description, created_at, updated_at
	`

	var todo Todo
	err := r.db.QueryRow(ctx, query, userID, input.Title, input.Description).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return &todo, nil
}

// GetByID retrieves a todo by ID for a specific user
func (r *TodoRepository) GetByID(ctx context.Context, todoId, userID int) (*Todo, error) {
	// Build query with filters
	query := `
		SELECT id, user_id, title, description, completed, created_at, updated_at
		FROM todos
		WHERE id = $1 and user_id = $2
	`

	var todo Todo
	err := r.db.QueryRow(ctx, query, todoId, userID).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return &todo, nil
}

// GetByUserID retrieves todos for a specific user with filtering
func (r *TodoRepository) GetByUserID(ctx context.Context, userID int, filter TodoFilter) (*TodoListResponse, error) {
	// Build query with filters
	query := `
		SELECT id, user_id, title, description, completed, updated_at, created_at
		FROM todos
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argIndex := 2

	// Add completed filter
	if filter.Completed != nil {
		query += fmt.Sprintf(" AND completed $%d", argIndex)
		args = append(args, *filter.Completed)
		argIndex++
	}

	// Add search filter
	if filter.Search != nil && *filter.Search != "" {
		query += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex)
		searchTerm := "%" + *filter.Search + "%"
		args = append(args, searchTerm)
		argIndex++
	}

	// Add ordering
	query += " ORDER BY created_at DESC"

	// Add pagination
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", argIndex)
		args = append(args, filter.Offset)
		argIndex++
	}

	// Execute query
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)
	}

	defer rows.Close()

	// Create todos array from query and append to todos array from rows
	var todos []*Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(
			&todo.ID,
			&todo.UserID,
			&todo.Title,
			&todo.Description,
			&todo.Completed,
			&todo.CreatedAt,
			&todo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, &todo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate todos: %w", err)
	}

	// Get total count of todos for paginations
	total, err := r.CountByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count todos: %w", err)
	}

	// Calculate pagination info
	limit := filter.Limit
	if limit == 0 {
		limit = len(todos)
	}
	offset := filter.Offset
	hasMore := offset+len(todos) < total

	return &TodoListResponse{
		Todos:   todos,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
		Total:   total,
	}, nil

}

// Update updates a todo for a specific user
func (r *TodoRepository) Update(ctx context.Context, todoID, userID int, input UpdateTodoInput) (*Todo, error) {
	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{todoID, userID}
	argIndex := 3

	if input.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = %d", argIndex))
		args = append(args, input.Title)
		argIndex++
	}

	if input.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = %d", argIndex))
		args = append(args, input.Description)
		argIndex++
	}

	if input.Completed != nil {
		setParts = append(setParts, fmt.Sprintf("completed = %d", argIndex))
		args = append(args, input.Completed)
		argIndex++
	}

	// No fields to update, just return current todo
	if len(setParts) == 0 {
		return r.GetByID(ctx, todoID, userID)
	}

	// Add updated_at
	query := fmt.Sprintf(`
		UPDATE todos
		SET %s
		WHERE id = $1 and user_id = $2
		RETURNING id, user_id, title, description, completed, created_at, updated_at
		`, strings.Join(setParts, ", "))

	var todo Todo
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update todo: %w", err)
	}

	return &todo, nil
}

// Delete deletes a todo for a specific user
func (r *TodoRepository) Delete(ctx context.Context, todoID, userID int) error {
	query := `
		DELETE FROM todos
		WHERE id = $1 and user_id = $2
	`

	result, err := r.db.Exec(ctx, query, todoID, userID)

	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrTodoNotFound
	}

	return nil
}

// ToggleComplete toggles the completed status of a todo
func (r *TodoRepository) ToggleComplete(ctx context.Context, todoID, userID int) (*Todo, error) {
	query := `
		UPDATE todos
		SET completed = NOT completed, updated_at = NOW()
		WHERE id = $1 and user_id = $2
		RETURNING id, user_id, title, description, completed, created_at, updated_at
	`

	var todo Todo
	err := r.db.QueryRow(ctx, query, todoID, userID).Scan(
		&todo.ID,
		&todo.UserID,
		&todo.Title,
		&todo.Description,
		&todo.Completed,
		&todo.CreatedAt,
		&todo.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTodoNotFound
		}
		return nil, fmt.Errorf("failed to toggle complete todo: %w", err)
	}

	return &todo, nil
}

// CountByUserID counts total todos for a user
func (r *TodoRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	query := `SELECT COUNT(*) FROM todos where user_id = $1`

	count := 0
	err := r.db.QueryRow(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get todo count for user: %w", err)
	}
	return count, nil
}

// Ensure TodoRepository implements Repository interface
var _ Repository = (*TodoRepository)(nil)
