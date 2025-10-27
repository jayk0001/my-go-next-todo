package auth

import (
	"testing"
	"time"
)

type MockUser struct {
	ID       string
	Email    string
	Password string
}

var (
	CorrectTestUser = MockUser{
		ID:       "123",
		Email:    "user@test.com",
		Password: "Password1234",
	}

	WrongTestUser = MockUser{
		ID:       "123",
		Email:    "usertestcom",
		Password: "pass12",
	}

	EmptyTestUser = MockUser{
		ID:       "",
		Email:    "",
		Password: "",
	}
)

// TestPasswordService tests password hasing and verification
func TestPasswordService(t *testing.T) {
	ps := NewPasswordService()

	// Testing hashing
	hashedPassword, err := ps.HashPassword(CorrectTestUser.Password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hashedPassword == CorrectTestUser.Password {
		t.Error("Hashed password should not equal original password")
	}

	// Testing verification - correct password
	err = ps.VerifyPassword(hashedPassword, CorrectTestUser.Password)
	if err != nil {
		t.Error("Should verify correct password")
	}

	err = ps.VerifyPassword(hashedPassword, "wrongpassword")
	if err == nil {
		t.Error("Should verify wrong password")
	}

}

// TestJWTService tests jwt token generation and validation
func TestJWTService(t *testing.T) {
	// Use longer secret key (32 + characters recommended for HMAC)
	secretKey := "this-is-a-very-long-secret-key-for-testing-jwt-tokens-safely"
	expiry := 1 * time.Hour
	js := NewJWTService(secretKey, expiry)

	userID := "123"
	email := "user@email.com"

	// Test token generation
	token, err := js.GenerateToken(userID, email)
	if err != nil {
		t.Fatalf("Failed to generate jwt token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Test token validation
	claims, err := js.ValidateToken(token)
	if err != nil {
		t.Errorf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID: %s, got %s", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email: %s, got %s", email, claims.Email)
	}

	// Test refresh token
	refreshToken, err := js.GenerateRefreshToken(userID)
	if err != nil {
		t.Errorf("Faliled to refresh token: %v", err)
	}

	extractedUserID, err := js.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Errorf("Failed to validate refresh token: %v", err)
	}

	if extractedUserID != userID {
		t.Errorf("Expected UserID: %s, got %s", userID, extractedUserID)
	}

}

// TestValidatorService tests input validation
func TestValidatorService(t *testing.T) {
	vs := NewValidatorService()

	// Test valid inputs
	err := vs.ValidateRegisterInput(CorrectTestUser.Email, CorrectTestUser.Password)

	if err != nil {
		t.Errorf("Failed to validate user input: %v", err)
	}

	// Test invalid email
	err = vs.ValidateRegisterInput(WrongTestUser.Email, CorrectTestUser.Password)
	if err == nil {
		t.Error("Should reject invalid email")
	}

	// Test short password
	err = vs.ValidateRegisterInput(CorrectTestUser.Email, WrongTestUser.Password)
	if err == nil {
		t.Error("Should reject short(<8) password")
	}

	// Test empty inputs - empty password
	err = vs.ValidateRegisterInput(CorrectTestUser.Email, EmptyTestUser.Password)
	if err == nil {
		t.Error("Should reject empty password")
	}

	// Test empty inputs - empty email
	err = vs.ValidateRegisterInput(EmptyTestUser.Email, CorrectTestUser.Password)
	if err == nil {
		t.Error("Should reject empty email")
	}

}

func TestGraphQLConversion(t *testing.T) {
	// Create Mock User
	user := &User{
		ID:        123,
		Email:     "test@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// GraphQL conversion test
	gqlUser := user.ToGraphQLUser()
	if gqlUser.Email != user.Email {
		t.Errorf("Email conversion failed")
	}

	// AuthResult conversion test
	authResult := &AuthResult{
		User:         user,
		Token:        "test-token",
		RefreshToken: "test-refresh",
		ExpiresAt:    time.Now(),
	}

	gqlPayload := authResult.ToGraphQLAuthPayload()
	if gqlPayload.Token != authResult.Token {
		t.Errorf("Token conversion failed")
	}
}
