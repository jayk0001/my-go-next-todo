package auth

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims defines the structure of JWT claims
type JWTClaims struct {
	UserID int    `json:"user_id"`
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
func (j *JWTService) GenerateToken(userID int, email string) (string, error) {
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
			Subject:   fmt.Sprintf("%d", userID),
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
func (j *JWTService) GenerateRefreshToken(userID int) (string, error) {
	now := time.Now()
	expiresAt := now.Add(24 * 7 * time.Hour) // 7 days

	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    "todo-app-refresh",
		Subject:   fmt.Sprintf("%d", userID),
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
func (j *JWTService) ValidateRefreshToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			return 0, err
		}
		return userID, nil
	}

	return 0, errors.New("invalid refresh token")
}
