package resolver

import "github.com/jayk0001/my-go-next-todo/internal/auth"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	AuthService *auth.AuthService
}

// NewResolver created a new resolver with dependencies
func NewResolver(authService *auth.AuthService) *Resolver {
	return &Resolver{
		AuthService: authService,
	}
}
