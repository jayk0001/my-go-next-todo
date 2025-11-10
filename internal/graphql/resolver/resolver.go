package resolver

import (
	"context"
	"errors"

	"github.com/jayk0001/my-go-next-todo/internal/auth"
	"github.com/jayk0001/my-go-next-todo/internal/middleware"
	"github.com/jayk0001/my-go-next-todo/internal/todo"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	AuthService *auth.AuthService
	TodoService todo.TodoServiceInterface
}

// NewResolver created a new resolver with dependencies
func NewResolver(authService *auth.AuthService, todoService todo.TodoServiceInterface) *Resolver {
	return &Resolver{
		AuthService: authService,
		TodoService: todoService,
	}
}

// getUserIDFromContext Helper to extract user ID from context
func getUserIDFromContext(ctx context.Context) (int, error) {
	user, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		return 0, errors.New("user not authenticated")
	}
	return user.ID, nil
}
