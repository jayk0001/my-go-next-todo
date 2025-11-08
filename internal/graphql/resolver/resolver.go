package resolver

import (
	"github.com/jayk0001/my-go-next-todo/internal/auth"
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
