package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims defines the structure of JWT claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secretKey   string
	expiryHours time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, expiryHours time.Duration) *JWTService {
	return &JWTService{
		secretKey:   secretKey,
		expiryHours: expiryHours,
	}
}

// GenerateToken creates a new JWT token for the user
func (j *JWTService) GenerateToken(userID, email string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(j.expiryHours)

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "todo-app",
			Subject:   userID,
		},
	}

	// Create token with explicit HMAC signing method
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	// Debug: Print signing method
	fmt.Printf("DEBUG: Using signing method: %v\n", token.Method.Alg())

	// Sign token with byte slice key
	tokenString, err := token.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken creates a refresh token with longer expiry
func (j *JWTService) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(24 * 7 * time.Hour) // 7 days

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    "todo-app-refresh",
		Subject:   userID,
	}

	// Create token with explicit HMAC signing method
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims

	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken validates and parses a JWT token
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken validates a refresh token
func (j *JWTService) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims.Subject, nil
	}

	return "", errors.New("invalid refresh token")
}
