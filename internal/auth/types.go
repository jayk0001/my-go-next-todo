package auth

import (
	"strconv"
	"time"

	"github.com/jayk0001/my-go-next-todo/internal/graphql/model"
)

// ToGraphQLUser converts auth.User to GraphQL User model
func (u *User) ToGraphQLUser() *model.User {
	user := &model.User{
		ID:        strconv.Itoa(u.ID),
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}

	if u.LastLoginAt != nil {
		lastLogin := u.LastLoginAt.Format(time.RFC3339)
		user.LastLoginAt = &lastLogin
	}

	return user
}

// ToGraphQLAuthPayload converts AuthResult to GraphQL AuthPayload
func (ar *AuthResult) ToGraphQLAuthPayload() *model.AuthPayload {
	return &model.AuthPayload{
		Token:        ar.Token,
		RefreshToken: ar.RefreshToken,
		User:         ar.User.ToGraphQLUser(),
		ExpiresAt:    ar.ExpiresAt.Format(time.RFC3339),
	}
}
