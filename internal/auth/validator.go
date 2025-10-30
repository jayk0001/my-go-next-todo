package auth

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// EmailRegex is a basic email validation regex
	EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z0-9]{2,}$`)
)

// ValidationErrors contains validation error messages
type ValidationErrors struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	var errors []string
	if ve.Email != "" {
		errors = append(errors, "email: "+ve.Email)
	}
	if ve.Password != "" {
		errors = append(errors, "password: "+ve.Password)
	}
	return strings.Join(errors, ", ")
}

// HasErrors returns treu if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return ve.Email != "" || ve.Password != ""
}

// ValidatorService handles input validation
type ValidatorService struct{}

// NewValidatorService created a new validator service
func NewValidatorService() *ValidatorService {
	return &ValidatorService{}
}

// ValidateRegisterInput validates user registration input
func (v *ValidatorService) ValidateRegisterInput(email, password string) error {
	ve := ValidationErrors{}

	// Validate email
	if email == "" {
		ve.Email = "email is required"
	} else if !v.IsValidEmail(email) {
		ve.Email = "email format is invalid"
	}

	// Validate password
	if password == "" {
		ve.Password = "password is required"
	} else if !v.IsValidPassword(password) {
		ve.Password = "password must be at least 8 character long"
	}

	if ve.HasErrors() {
		return ve
	}

	return nil
}

// ValidateLoginInput validates user login input
func (v *ValidatorService) ValidateLoginInput(email, password string) error {
	if email == "" {
		return errors.New("email is required")
	}
	if password == "" {
		return errors.New("password is required")
	}
	return nil
}

// IsValidEmail checks if email format is valid
func (v *ValidatorService) IsValidEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	// Basic regex check
	if !EmailRegex.MatchString(email) {
		return false
	}

	// Additional checks for edge cases
	if strings.Contains(email, "..") {
		return false
	}

	if strings.HasPrefix(email, ".") || strings.HasSuffix(email, ".") {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) < 2 {
		return false
	}

	localPart := parts[0]
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return false
	}

	return true
}

// IsValidPassword checks if password meets requirements
func (v *ValidatorService) IsValidPassword(password string) bool {
	return len(password) >= 8
}
