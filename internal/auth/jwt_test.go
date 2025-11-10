package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JTWTestData struct {
	secretKey   string
	expiryHours time.Duration
	userID      int
	email       string
}

var (
	TestData = JTWTestData{
		secretKey:   "test-secret-key-32-chars-long",
		expiryHours: 1 * time.Hour,
		email:       "test@user.com",
		userID:      123,
	}

	service = NewJWTService(TestData.secretKey, TestData.expiryHours)
)

// TestNewJWTService tests JWT service creation
func TestNewJWTService(t *testing.T) {
	// Defined service as global variable to re-use in other test cases
	// service = NewJWTService(TestData.secretKey, TestData.expiryHours)

	if service == nil {
		t.Fatal("JWT service should not be nil")
	}

	if service.secretKey != TestData.secretKey {
		t.Errorf("Expected secret key: %s, got %s", TestData.secretKey, service.secretKey)
	}

	if service.expiryHours != TestData.expiryHours {
		t.Errorf("Expected expiry hours: %s, got %s", TestData.expiryHours, service.expiryHours)
	}
}

// TestGenerateToken tests JWT token generation
func TestGenerateToken(t *testing.T) {

	token, err := service.GenerateToken(TestData.userID, TestData.email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Check if token has proper JWT structure(head.payload.signature)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("JWT token should have 3 parts, got %d", len(parts))
	}
}

// TestGenerateRefreshToken tests refresh token generation
func TestGenerateRefreshToken(t *testing.T) {

	token, err := service.GenerateRefreshToken(TestData.userID)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Check if refresh token has proper JWT structure
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Refresh token should have 3 aprts, got %d", len(parts))
	}
}

// TestValidationToken tests JWT token validation
func TestValidationToken(t *testing.T) {

	token, err := service.GenerateToken(TestData.userID, TestData.email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != TestData.userID {
		t.Errorf("Expected UserID: %d, got %d", TestData.userID, claims.UserID)
	}

	if claims.Email != TestData.email {
		t.Errorf("Expected UserID: %s, got %s", TestData.email, claims.Email)
	}

	if claims.Issuer != "todo-app" {
		t.Errorf("Expected Issuer: %s, got %s", "todo-app", claims.Issuer)
	}

	// if claims.Subject != TestData.userID {
	// 	t.Errorf("Expected Subject: %s, got %s", TestData.userID, claims.Subject)
	// }
}

// TestValidateInvalidToken tests validation of invalid tokens
func TestValidateInvalidToken(t *testing.T) {
	testCases := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"marfolmed token", "invalid token"},
		{"invalid signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMTIzIiwiZW1haWwiOiJ0ZXN0QGV4YW1wbGUuY29tIn0.invalid_signature"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.ValidateToken(tc.token)
			if err == nil {
				t.Errorf("Expected error for %s, but got none", tc.name)
			}
		})
	}
}

// TestValidateTokenWithWrongSecret tests token validation with wrong secret
func TestValidateTokenWithWrongSecret(t *testing.T) {
	// Generate token with one secret
	service1 := NewJWTService("secret-key-1", TestData.expiryHours)

	token, err := service1.GenerateToken(TestData.userID, TestData.email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to validate token with wrong secret
	service2 := NewJWTService("secret-key-2", TestData.expiryHours)
	_, err = service2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating token with wrong secret")
	}
}

// TestValidateExpiredToken tests expired token validation
func TestValidateExpiredToken(t *testing.T) {
	service := NewJWTService(TestData.secretKey, 1*time.Millisecond)

	token, err := service.GenerateToken(TestData.userID, TestData.email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	_, err = service.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for expired token")
	}

	// Check if error is specifically about expiration
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected expiration error, got: %v", err)
	}
}

// TestTokenClaims tests the structure of token claims
func TestTokenClaims(t *testing.T) {
	token, err := service.GenerateToken(TestData.userID, TestData.email)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Test all claim fields
	if claims.UserID != TestData.userID {
		t.Errorf("Expected UserID: %d, got %d", TestData.userID, claims.UserID)
	}

	if claims.Email != TestData.email {
		t.Errorf("Expected UserID: %s, got %s", TestData.email, claims.Email)
	}

	if claims.Issuer != "todo-app" {
		t.Errorf("Expected Issuer: %s, got %s", "todo-app", claims.Issuer)
	}

	// if claims.Subject != TestData.userID {
	// 	t.Errorf("Expected Subject: %s, got %s", TestData.userID, claims.Subject)
	// }

	// Check time-based claims
	now := time.Now()

	if claims.IssuedAt.Time.After(now) {
		t.Error("IssuedAt should not be in the future")
	}

	if claims.ExpiresAt.Time.Before(now) {
		t.Error("ExpiresAt should be in the future")
	}

	if claims.NotBefore.Time.After(now) {
		t.Error("NotBefore should be in the future")
	}
}

// TestRefreshTokenExpiry tests refresh token expiry (7 days)
func TestRefreshTokenExpiry(t *testing.T) {
	refreshToken, err := service.GenerateRefreshToken(TestData.userID)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	// Parse the refresh token to check expiry
	token, err := jwt.ParseWithClaims(refreshToken, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(service.secretKey), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse refresh token: %v", err)
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		expectedExpirty := time.Now().Add(7 * 24 * time.Hour)
		timeDiff := claims.ExpiresAt.Time.Sub(expectedExpirty)

		// Allow 1 minute tolerance for test execution time
		if timeDiff > time.Minute || timeDiff < -time.Minute {
			t.Errorf("Refresh token expiry time is incorrect. Expected around %v, got %v",
				expectedExpirty, claims.ExpiresAt.Time)
		}
	} else {
		t.Error("Failed to extract climas from refresh token")
	}
}
