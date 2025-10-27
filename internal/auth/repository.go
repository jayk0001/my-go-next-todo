package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExits         = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// User represents a user in the database
type User struct {
	ID           int        `db:"id" json:"id"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
}

// CreateUserInput represents input for creating a user
type CreateUserInput struct {
	Email    string
	Password string
}

// UserRepository handles user database operations
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository created a new user repository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// Create created a new user in the database
func (r *UserRepository) Create(ctx context.Context, input CreateUserInput) (*User, error) {
	// Check if email already exists
	exists, err := r.EmailExists(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrEmailExits
	}

	// Hash password
	PasswordService := NewPasswordService()
	hashedPassword, err := PasswordService.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Insert User
	query := `
		INSERT INTO users (email, password_hash, created_at, updated_at)
		VALUES($1, $2, NOW(), NOW())
		RETURNING id, email, password_hash, created_at, updated_at, last_login_at
	`

	var user User
	err = r.db.QueryRow(ctx, query, input.Email, hashedPassword).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`

	var user User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool

	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int) error {
	query := `
		UPDATE users
		SET last_login_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)

	return err
}

func (r *UserRepository) Authenticate(ctx context.Context, email, password string) (*User, error) {
	user, err := r.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	passwordService := NewPasswordService()
	if err := passwordService.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	if err := r.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail authentication
		// In production, might want to log this
	}

	return user, nil
}
