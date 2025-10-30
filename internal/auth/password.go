package auth

import "golang.org/x/crypto/bcrypt"

const (
	// DefaultCost is the default bcrypt cost for password hashing
	DefaultCost = 12
)

// PasswordService handles password hashing and verification
type PasswordService struct {
	cost int
}

// NewPasswordService created a new password service
func NewPasswordService() *PasswordService {
	return &PasswordService{
		cost: DefaultCost,
	}
}

// HashPassword hashes a plain text password using bcrypt
func (p *PasswordService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a plain text password against a hashed password
func (p *PasswordService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// IsValidPassword checks if password meets minimum requirements
func (p *PasswordService) IsValidPassword(password string) bool {
	// Basic password validation
	// In production, might want more sophisticated rules
	return len(password) >= 8
}
