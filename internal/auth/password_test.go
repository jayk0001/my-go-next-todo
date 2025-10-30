package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestNewPasswordService tests passwrod service creation

type PasswordTestData struct {
	service         *PasswordService
	validPassword   string
	invalidPassword string
}

var (
	passwordTestData = PasswordTestData{
		service:         NewPasswordService(),
		validPassword:   "validpassword123",
		invalidPassword: "short",
	}
)

// TestNewPasswordService tests password service creation
func TestNewPasswordService(t *testing.T) {

	if passwordTestData.service == nil {
		t.Fatal("Password service should not be nil")
	}

	if passwordTestData.service.cost != DefaultCost {
		t.Errorf("Expected cost: %d, got %d", DefaultCost, passwordTestData.service.cost)
	}
}

// TestHashPassword tests password hashing functionality
func TestHashPassword(t *testing.T) {

	hashedPassword, err := passwordTestData.service.HashPassword(passwordTestData.validPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %d", err)
	}

	if hashedPassword == "" {
		t.Errorf("Hashed password should not be empty")
	}

	if hashedPassword == passwordTestData.validPassword {
		t.Errorf("hashed password shold not equal original password")
	}

	if !strings.HasPrefix(hashedPassword, "$2a$") && !strings.HasPrefix(hashedPassword, "$2b$") {
		t.Error("Hashed password should have bcrypt format")
	}
}

// TestHashPasswordConsistency test that same password produces different hashes
func TestHashPasswordConsistenty(t *testing.T) {

	hash1, err := passwordTestData.service.HashPassword(passwordTestData.validPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %d", err)
	}

	hash2, err := passwordTestData.service.HashPassword(passwordTestData.validPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %d", err)
	}

	// Bcrypt should produce different hashes for same password (due to salt)
	if hash1 == hash2 {
		t.Error("Same password should produce different hash due to salt")
	}
}

// TestVerifyPassword tests password verification with correct password
func TestVerifyPassword(t *testing.T) {
	// hash it, err handling
	// verify, if err handle it
	hashedPassword, err := passwordTestData.service.HashPassword(passwordTestData.validPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	err = passwordTestData.service.VerifyPassword(hashedPassword, passwordTestData.validPassword)
	if err != nil {
		t.Errorf("Should verify password, but got %v", err)
	}
}

// TestVerifyPasswordWithWrongPassword tests verify password with incorrect password
func TestVerifyPasswordWithWrongPassword(t *testing.T) {

	hashedPassword, err := passwordTestData.service.HashPassword(passwordTestData.validPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Test wrong password verification
	err = passwordTestData.service.VerifyPassword(hashedPassword, passwordTestData.invalidPassword)
	if err == nil {
		t.Error("Should not verify wrong password")
	}

	// Check if error is specifically about mismatch
	if err != bcrypt.ErrMismatchedHashAndPassword {
		t.Errorf("Expected bcrypt.ErrMismatchedHashAndPassword, got %v", err)
	}
}

// TestVerifyPasswordEdgecases tests edge cases for password verification
func TestVerifyPasswordEdgecases(t *testing.T) {
	testCases := []struct {
		name           string
		hashedPassword string
		password       string
		shouldFail     bool
	}{
		{"empty hash", "", "password", true},
		{"empty password", "$2a$12$valid.hash.here", "", true},
		{"invalid hash format", "not-a-bcrypt-hash", "password", true},
		{"malformed hash", "$2a$12$invalidhash", "password", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := passwordTestData.service.VerifyPassword(tc.hashedPassword, tc.password)
			if tc.shouldFail && err == nil {
				t.Errorf("Expected error for case '%s', but got none", tc.name)
			}
			if !tc.shouldFail && err != nil {
				t.Errorf("Expected no error for case '%s', but got: %v", tc.name, err)
			}
		})
	}
}

// TestHashPasswordWithSpecialCharacters tests hashing with special characters
func TestHashPasswordWithSpecialCharacters(t *testing.T) {
	specialPasswords := []string{
		"password!@#$%^&*()",
		"pässwörd123",
		"パスワード123",
		"🔒secure🔑password",
		"password with spaces",
		"password\twith\ttabs",
		"password\nwith\nnewlines",
	}

	for _, password := range specialPasswords {
		t.Run("special char", func(t *testing.T) {
			hashedPassword, err := passwordTestData.service.HashPassword(password)
			if err != nil {
				t.Errorf("Failed to hash speical characters %s: %v", password, err)
			}

			err = passwordTestData.service.VerifyPassword(hashedPassword, password)
			if err != nil {
				t.Errorf("Failed to verify password with special characters %s : %v", password, err)
			}
		})
	}
}

// TestPasswordServiceCost tests that the service uses correct bcrypt cost
func TestPasswordServiceCost(t *testing.T) {
	hashedPassword, err := passwordTestData.service.HashPassword(passwordTestData.validPassword)
	if err != nil {
		t.Fatalf("Failed to hash password :%v", err)
	}

	cost, err := bcrypt.Cost([]byte(hashedPassword))
	if err != nil {
		t.Fatalf("Failed to extract cost from hash: %v", err)
	}

	if cost != DefaultCost {
		t.Errorf("Expected cost %d, got %d", DefaultCost, cost)
	}
}

// TestHashPasswordPerformance tests that hashing doesn't take too long
func TestHashPasswordPerformance(t *testing.T) {

	// Hashing should complete within reasonable time (bcrypt cost 12 is slow but not too slow)
	start := testing.Short()
	if start {
		t.Skip("Skipping performance test in short mode")
	}

	_, err := passwordTestData.service.HashPassword(passwordTestData.validPassword)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Note: With cost 12, hashing can take 100-500ms, which is expected
	// This test mainly ensures it doesn't hang indefinitely
}

// TestPasswordLengthLimits tests password length edge cases
func TestPasswordLengthLimits(t *testing.T) {

	// Test very long password (bcrypt has 72 byte limit)
	longPassword := strings.Repeat("a", 50)
	hashedPassword, err := passwordTestData.service.HashPassword(longPassword)
	if err != nil {
		t.Errorf("Failed to hash long password: %v", err)
	}

	// Should still verify correctly (bcrypt truncates at 72 bytes)
	err = passwordTestData.service.VerifyPassword(hashedPassword, longPassword)
	if err != nil {
		t.Errorf("Failed to verify long password: %v", err)
	}

	// Test extremely long password
	veryLongPassword := strings.Repeat("x", 1000)
	_, err = passwordTestData.service.HashPassword(veryLongPassword)
	if err == nil {
		t.Errorf("Expect to fail to hash very long password: %v", err)
	}
}

// TestPasswordServiceMethods tests that all methods work together
func TestPasswordServiceMethods(t *testing.T) {

	// Test the complete flow
	if !passwordTestData.service.IsValidPassword(passwordTestData.validPassword) {
		t.Error("Password should be valid")
	}

	hashedPassword, err := passwordTestData.service.HashPassword(passwordTestData.validPassword)
	if err != nil {
		t.Fatalf("Failed to hash valid password: %v", err)
	}

	err = passwordTestData.service.VerifyPassword(hashedPassword, passwordTestData.validPassword)
	if err != nil {
		t.Errorf("Failed to verify hashed password: %v", err)
	}

	// Test with invalid password
	if passwordTestData.service.IsValidPassword(passwordTestData.invalidPassword) {
		t.Error("Short password should be invalid")
	}
}
