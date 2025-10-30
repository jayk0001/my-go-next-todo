package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jayk0001/my-go-next-todo/internal/auth"
)

type contextKey string

const (
	userContextKey contextKey = "github.com/jayk0001/my-go-next-todo/middleware.user"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(authService *auth.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from "Bearer <token>" format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := tokenParts[1]
		if len(token) > 20 {
			fmt.Printf("AuthMiddleware: Extracted token: %s...\n", token[:20])
		} else {
			fmt.Printf("AuthMiddleware: Extracted token: %s\n", token)
		}

		// Validate token and get user
		user, err := authService.GetUserFromToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Add user to context
		ctx := context.WithValue(c.Request.Context(), userContextKey, user)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// CORS middleware for development
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-type, Content-length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// GetUserFromContext helper function to extract user from context
func GetUserFromContext(ctx context.Context) (*auth.User, bool) {
	user, ok := ctx.Value(userContextKey).(*auth.User)
	return user, ok
}
