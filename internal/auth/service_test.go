package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// MockUserRepository embeds UserRepository and overrides methods for testing
type MockUserRepository struct {
	users           map[int]*User
	usersByEmail    map[string]*User
	nextID          int
	shouldFailOnDB  bool
	emailExists     bool
	forceEmailError bool
}

// NewMockUserRepository creates a new mock user repository
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:        make(map[int]*User),
		usersByEmail: make(map[string]*User),
		nextID:       1,
	}
}

// Override UserRepository methods for testing
func (m *MockUserRepository) Create(ctx context.Context, input CreateUserInput) (*User, error) {
	if m.shouldFailOnDB {
		return nil, errors.New("database error")
	}

	if m.emailExists || m.usersByEmail[input.Email] != nil {
		return nil, ErrEmailExits
	}

	// Hash password using real password service for testing
	passwordService := NewPasswordService()
	hashedPassword, err := passwordService.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &User{
		ID:           m.nextID,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	m.nextID++

	return user, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int) (*User, error) {
	if m.shouldFailOnDB {
		return nil, errors.New("database error")
	}

	user, exists := m.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	if m.shouldFailOnDB {
		return nil, errors.New("database error")
	}

	user, exists := m.usersByEmail[email]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	if m.forceEmailError {
		return false, errors.New("database error")
	}

	if m.emailExists {
		return true, nil
	}

	_, exists := m.usersByEmail[email]
	return exists, nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID int) error {
	if m.shouldFailOnDB {
		return errors.New("database error")
	}

	user, exists := m.users[userID]
	if !exists {
		return ErrUserNotFound
	}

	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now

	return nil
}

func (m *MockUserRepository) Authenticate(ctx context.Context, email, password string) (*User, error) {
	if m.shouldFailOnDB {
		return nil, errors.New("database error")
	}

	user, err := m.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password using real password service
	passwordService := NewPasswordService()
	if err := passwordService.VerifyPassword(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	_ = m.UpdateLastLogin(ctx, user.ID)

	return user, nil
}

// Mock control methods
func (m *MockUserRepository) SetShouldFailOnDB(shouldFail bool) {
	m.shouldFailOnDB = shouldFail
}

func (m *MockUserRepository) SetEmailExists(exists bool) {
	m.emailExists = exists
}

func (m *MockUserRepository) SetForceEmailError(forceError bool) {
	m.forceEmailError = forceError
}

func (m *MockUserRepository) AddUser(user *User) {
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	if user.ID >= m.nextID {
		m.nextID = user.ID + 1
	}
}

func (m *MockUserRepository) Clear() {
	m.users = make(map[int]*User)
	m.usersByEmail = make(map[string]*User)
	m.nextID = 1
	m.shouldFailOnDB = false
	m.emailExists = false
	m.forceEmailError = false
}

// Test data
type ServiceTestData struct {
	jwtSecret    string
	tokenExpiry  time.Duration
	testEmail    string
	testPassword string
	testUserID   int
}

var (
	serviceTestData = ServiceTestData{
		jwtSecret:    "test-secret-key-32-chars-long-for-testing",
		tokenExpiry:  1 * time.Hour,
		testEmail:    "test@example.com",
		testPassword: "password123",
		testUserID:   1,
	}
)

// createTestAuthServiceWithMock creates an auth service with mock repository
func createTestAuthServiceWithMock() (*AuthService, *MockUserRepository) {
	mockRepo := NewMockUserRepository()
	authService := &AuthService{
		userRepo:         mockRepo,
		jwtService:       NewJWTService(serviceTestData.jwtSecret, serviceTestData.tokenExpiry),
		passwordService:  NewPasswordService(),
		validatorService: NewValidatorService(),
	}

	return authService, mockRepo
}

// TestNewAuthService tests auth service creation
func TestNewAuthService(t *testing.T) {
	authService := NewAuthService(nil, serviceTestData.jwtSecret, serviceTestData.tokenExpiry)

	if authService == nil {
		t.Fatal("Auth service should not be nil")
	}

	if authService.jwtService == nil {
		t.Error("JWT service should be initialized")
	}

	if authService.passwordService == nil {
		t.Error("Password service should be initialized")
	}

	if authService.validatorService == nil {
		t.Error("Validator service should be initialized")
	}

	if authService.userRepo == nil {
		t.Error("User repository should be initialized")
	}
}

// TestRegisterSuccess tests successful user registration
func TestRegisterSuccess(t *testing.T) {
	authService, mockRepo := createTestAuthServiceWithMock()
	ctx := context.Background()

	// Test successful registration
	result, err := authService.Register(ctx, serviceTestData.testEmail, serviceTestData.testPassword)
	if err != nil {
		t.Fatalf("Registration should succeed: %v", err)
	}

	// Verify AuthResult structure
	if result == nil {
		t.Fatal("AuthResult should not be nil")
	}

	if result.User == nil {
		t.Fatal("User should not be nil")
	}

	if result.User.Email != serviceTestData.testEmail {
		t.Errorf("Expected email %s, got %s", serviceTestData.testEmail, result.User.Email)
	}

	if result.Token == "" {
		t.Error("Token should not be empty")
	}

	if result.RefreshToken == "" {
		t.Error("Refresh token should not be empty")
	}

	if result.ExpiresAt.Before(time.Now()) {
		t.Error("ExpiresAt should be in the future")
	}

	// Verify user was created in repository
	user, err := mockRepo.GetByEmail(ctx, serviceTestData.testEmail)
	if err != nil {
		t.Fatalf("User should exist in repository: %v", err)
	}

	if user.Email != serviceTestData.testEmail {
		t.Errorf("Expected email %s, got %s", serviceTestData.testEmail, user.Email)
	}
}

// TestRegisterEmailExists tests registration with existing email
func TestRegisterEmailExists(t *testing.T) {
	authService, _ := createTestAuthServiceWithMock()
	ctx := context.Background()

	// Create user first
	_, err := authService.Register(ctx, serviceTestData.testEmail, serviceTestData.testPassword)
	if err != nil {
		t.Fatalf("First registration should succeed: %v", err)
	}

	// Try to register again with same email
	_, err = authService.Register(ctx, serviceTestData.testEmail, serviceTestData.testPassword)
	if err == nil {
		t.Error("Second registration should fail with email exists error")
	}

	if !errors.Is(err, ErrEmailExits) && !strings.Contains(err.Error(), "email already exists") {
		t.Errorf("Expected email exists error, got: %v", err)
	}
}

// TestLoginSuccess tests successful user login
func TestLoginSuccess(t *testing.T) {
	authService, _ := createTestAuthServiceWithMock()
	ctx := context.Background()

	// Register user first
	_, err := authService.Register(ctx, serviceTestData.testEmail, serviceTestData.testPassword)
	if err != nil {
		t.Fatalf("Registration should succeed: %v", err)
	}

	// Test login
	result, err := authService.Login(ctx, serviceTestData.testEmail, serviceTestData.testPassword)
	if err != nil {
		t.Fatalf("Login should succeed: %v", err)
	}

	// Verify AuthResult
	if result == nil {
		t.Fatal("AuthResult should not be nil")
	}

	if result.User.Email != serviceTestData.testEmail {
		t.Errorf("Expected email %s, got %s", serviceTestData.testEmail, result.User.Email)
	}

	if result.Token == "" {
		t.Error("Token should not be empty")
	}
}

// TestGetUserFromTokenSuccess tests successful user extraction from token
func TestGetUserFromTokenSuccess(t *testing.T) {
	authService, _ := createTestAuthServiceWithMock()
	ctx := context.Background()

	// Register user and get token
	result, err := authService.Register(ctx, serviceTestData.testEmail, serviceTestData.testPassword)
	if err != nil {
		t.Fatalf("Registration should succeed: %v", err)
	}

	// Get user from token
	user, err := authService.GetUserFromToken(ctx, result.Token)
	if err != nil {
		t.Fatalf("GetUserFromToken should succeed: %v", err)
	}

	if user.Email != serviceTestData.testEmail {
		t.Errorf("Expected email %s, got %s", serviceTestData.testEmail, user.Email)
	}

	if user.ID != result.User.ID {
		t.Errorf("Expected user ID %d, got %d", result.User.ID, user.ID)
	}
}
