package auth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo         *UserRepository
	jwtService       *JWTService
	passwordService  *PasswordService
	validatorService *ValidatorService
}

// AuthResult represents the result of authentication operations
type AuthResult struct {
	User         *User
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
}

// NewAuthService created a new authentication service
func NewAuthService(db *pgxpool.Pool, jwtSecret string, tokenExpiry time.Duration) *AuthService {
	return &AuthService{
		userRepo:         NewUserRepository(db),
		jwtService:       NewJWTService(jwtSecret, tokenExpiry),
		passwordService:  NewPasswordService(),
		validatorService: NewValidatorService(),
	}
}

// Register created a new user account
func (s *AuthService) Register(ctx context.Context, email, password string) (*AuthResult, error) {
	// Validate input
	if err := s.validatorService.ValidateRegisterInput(email, password); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create user
	user, err := s.userRepo.Create(ctx, CreateUserInput{
		Email:    email,
		Password: password,
	})

	if err != nil {
		if err == ErrEmailExits {
			return nil, fmt.Errorf("registration failed %w", err)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	UserIDStr := strconv.Itoa(user.ID)
	token, err := s.jwtService.GenerateToken(UserIDStr, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(UserIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &AuthResult{
		User:         user,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * 7 * time.Hour), // Same as JWT Expiry
	}, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	// Validate input
	if err := s.validatorService.ValidateLoginInput(email, password); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Authenticate user
	user, err := s.userRepo.Authenticate(ctx, email, password)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Generate tokens
	UserIDStr := strconv.Itoa(user.ID)
	token, err := s.jwtService.GenerateToken(UserIDStr, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(UserIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &AuthResult{
		User:         user,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(24 * 7 * time.Hour),
	}, nil
}

// RefreshToken generates new tokens from refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Validate refresh token
	userIDStr, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	fmt.Println(userIDStr, "hmm")
	// Get user
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found %w", err)
	}

	// Generate new tokens
	newToken, err := s.jwtService.GenerateToken(userIDStr, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	newRefreshToken, err := s.jwtService.GenerateRefreshToken(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &AuthResult{
		User:         user,
		Token:        newToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(24 * 7 * time.Hour),
	}, nil
}

// GetUserFromToken extracts user from JWT token
func (s *AuthService) GetUserFromToken(ctx context.Context, token string) (*User, error) {
	// Validate token
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get user
	userId, err := strconv.Atoi(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invaild user ID in token: %w", err)
	}

	user, err := s.userRepo.GetByID(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil

}

// GetUserByID retrieves user by ID
func (s *AuthService) GetUserByID(ctx context.Context, userID int) (*User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}
