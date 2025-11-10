package auth

import (
	"testing"
	"time"
)

// TestPasswordService tests password hashing and verification
func TestPasswordService(t *testing.T) {
	ps := NewPasswordService()

	password := "testpassword123"

	// Test hashing
	hashedPassword, err := ps.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hashedPassword == password {
		t.Error("Hashed password should not equal original password")
	}

	// Test verification - correct password
	err = ps.VerifyPassword(hashedPassword, password)
	if err != nil {
		t.Error("Should verify correct password")
	}

	// Test verification - wrong password
	err = ps.VerifyPassword(hashedPassword, "wrongpassword")
	if err == nil {
		t.Error("Should not verify wrong password")
	}
}

// TestJWTService tests JWT token generation and validation
func TestJWTService(t *testing.T) {
	// Use longer secret key (32+ characters recommended for HMAC)
	secretKey := "this-is-a-very-long-secret-key-for-testing-jwt-tokens-safely"
	expiry := 1 * time.Hour

	jwtService := NewJWTService(secretKey, expiry)

	userID := 123
	email := "test@example.com"

	// Test token generation
	token, err := jwtService.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Test token validation
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}

	// Test refresh token
	refreshToken, err := jwtService.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	extractedUserID, err := jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if extractedUserID != userID {
		t.Errorf("Expected Subject: %d, got %s", TestData.userID, claims.Subject)
	}
}

// TestValidatorService tests input validation
func TestValidatorService(t *testing.T) {
	vs := NewValidatorService()

	// Test valid inputs
	err := vs.ValidateRegisterInput("test@example.com", "password123")
	if err != nil {
		t.Errorf("Should validate correct input: %v", err)
	}

	// Test invalid email
	err = vs.ValidateRegisterInput("invalid-email", "password123")
	if err == nil {
		t.Error("Should reject invalid email")
	}

	// Test short password
	err = vs.ValidateRegisterInput("test@example.com", "123")
	if err == nil {
		t.Error("Should reject short password")
	}

	// Test empty inputs
	err = vs.ValidateRegisterInput("", "")
	if err == nil {
		t.Error("Should reject empty inputs")
	}
}
