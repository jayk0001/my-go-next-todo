package todo

import "errors"

var (
	// ErrTodoNotFound is returned when a todo is not found
	ErrTodoNotFound = errors.New("todo not found")

	// ErrTodoAccessDenied is returned when user tries to access todo they don't own
	ErrTodoAccessDenied = errors.New("access denied: todo does not belong to user")

	// ErrInvalidTodoInput is return when input validation fails
	ErrInvalidTodoInput = errors.New("invalid todo input")

	// ErrTodoTitleRequired is return when title is empty
	ErrTodoTitleRequired = errors.New("todo title is required")

	// ErrTodoTitleTooLong is returned when title exceeds max length
	ErrTodoTitleTooLong = errors.New("todo title too long (max 500 characters)")

	// ErrTodoDescriptionTooLoing is return when description exceeds max lengths
	ErrTodoDescriptionTooLong = errors.New("todo description too long (max 2000 characters)")
)
