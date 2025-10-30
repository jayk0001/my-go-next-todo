package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockDatabase simulates database operations for testing
type MockDatabase struct {
	users        map[int]*User
	usersByEmail map[string]*User
	nextID       int
	shouldFail   bool
	queryError   error
}

// NewMockDatabase creates a new mock database
func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		users:        make(map[int]*User),
		usersByEmail: make(map[string]*User),
		nextID:       1,
	}
}

// Control methods for testing
func (m *MockDatabase) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

func (m *MockDatabase) SetQueryError(err error) {
	m.queryError = err
}

func (m *MockDatabase) Clear() {
	m.users = make(map[int]*User)
	m.usersByEmail = make(map[string]*User)
	m.nextID = 1
	m.shouldFail = false
	m.queryError = nil
}

func (m *MockDatabase) AddUser(user *User) {
	m.users[user.ID] = user
	m.usersByEmail[user.Email] = user
	if user.ID >= m.nextID {
		m.nextID = user.ID + 1
	}
}

// MockUserRepository implements UserRepository methods with mock database
type MockUserRepositoryDB struct {
	mockDB *MockDatabase
}

// NewMockUserRepositoryDB creates a repository with mock database
func NewMockUserRepositoryDB() *MockUserRepositoryDB {
	return &MockUserRepositoryDB{
		mockDB: NewMockDatabase(),
	}
}

func (r *MockUserRepositoryDB) Create(ctx context.Context, input CreateUserInput) (*User, error) {
	if r.mockDB.shouldFail {
		return nil, r.mockDB.queryError
	}

	// Check if email exists
	if r.mockDB.usersByEmail[input.Email] != nil {
		return nil, ErrEmailExits
	}

	// Hash password
	passwordService := NewPasswordService()
	hashedPassword, err := passwordService.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &User{
		ID:           r.mockDB.nextID,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	r.mockDB.users[user.ID] = user
	r.mockDB.usersByEmail[user.Email] = user
	r.mockDB.nextID++

	return user, nil
}

func (r *MockUserRepositoryDB) GetByID(ctx context.Context, id int) (*User, error) {
	if r.mockDB.shouldFail {
		return nil, r.mockDB.queryError
	}

	user, exists := r.mockDB.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (r *MockUserRepositoryDB) GetByEmail(ctx context.Context, email string) (*User, error) {
	if r.mockDB.shouldFail {
		return nil, r.mockDB.queryError
	}

	user, exists := r.mockDB.usersByEmail[email]
	if !exists {
		return nil, ErrUserNotFound
	}

	return user, nil
}

func (r *MockUserRepositoryDB) EmailExists(ctx context.Context, email string) (bool, error) {
	if r.mockDB.shouldFail {
		return false, r.mockDB.queryError
	}

	_, exists := r.mockDB.usersByEmail[email]
	return exists, nil
}

func (r *MockUserRepositoryDB) UpdateLastLogin(ctx context.Context, userID int) error {
	if r.mockDB.shouldFail {
		return r.mockDB.queryError
	}

	user, exists := r.mockDB.users[userID]
	if !exists {
		return ErrUserNotFound
	}

	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now

	return nil
}

func (r *MockUserRepositoryDB) Authenticate(ctx context.Context, email, password string) (*User, error) {
	if r.mockDB.shouldFail {
		return nil, r.mockDB.queryError
	}

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
	_ = r.UpdateLastLogin(ctx, user.ID)

	return user, nil
}

// Repository test data
type RepositoryTestData struct {
	testEmail    string
	testPassword string
	testUserID   int
}

var (
	repoTestData = RepositoryTestData{
		testEmail:    "test@example.com",
		testPassword: "password123",
		testUserID:   1,
	}
)

// TestNewUserRepository tests repository creation
func TestNewUserRepository(t *testing.T) {
	repo := NewUserRepository(nil)

	if repo == nil {
		t.Fatal("Repository should not be nil")
	}

	if repo.db != nil {
		t.Error("Expected nil database for test")
	}
}

// TestCreateUser tests user creation
func TestCreateUser(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	input := CreateUserInput{
		Email:    repoTestData.testEmail,
		Password: repoTestData.testPassword,
	}

	user, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	if user == nil {
		t.Fatal("User should not be nil")
	}

	if user.Email != repoTestData.testEmail {
		t.Errorf("Expected email %s, got %s", repoTestData.testEmail, user.Email)
	}

	if user.ID <= 0 {
		t.Error("User ID should be positive")
	}

	if user.PasswordHash == repoTestData.testPassword {
		t.Error("Password should be hashed")
	}

	if user.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should be set")
	}
}

// TestCreateUserEmailExists tests creating user with existing email
func TestCreateUserEmailExists(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	input := CreateUserInput{
		Email:    repoTestData.testEmail,
		Password: repoTestData.testPassword,
	}

	// Create first user
	_, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("First create should succeed: %v", err)
	}

	// Try to create second user with same email
	_, err = repo.Create(ctx, input)
	if err == nil {
		t.Error("Second create should fail with email exists error")
	}

	if !errors.Is(err, ErrEmailExits) {
		t.Errorf("Expected ErrEmailExits, got: %v", err)
	}
}

// TestGetByID tests user retrieval by ID
func TestGetByID(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	// Create user first
	input := CreateUserInput{
		Email:    repoTestData.testEmail,
		Password: repoTestData.testPassword,
	}

	createdUser, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Get user by ID
	user, err := repo.GetByID(ctx, createdUser.ID)
	if err != nil {
		t.Fatalf("GetByID should succeed: %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("Expected ID %d, got %d", createdUser.ID, user.ID)
	}

	if user.Email != repoTestData.testEmail {
		t.Errorf("Expected email %s, got %s", repoTestData.testEmail, user.Email)
	}
}

// TestGetByIDNotFound tests getting non-existent user
func TestGetByIDNotFound(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)
	if err == nil {
		t.Error("GetByID should fail for non-existent user")
	}

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestGetByEmail tests user retrieval by email
func TestGetByEmail(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	// Create user first
	input := CreateUserInput{
		Email:    repoTestData.testEmail,
		Password: repoTestData.testPassword,
	}

	createdUser, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Get user by email
	user, err := repo.GetByEmail(ctx, repoTestData.testEmail)
	if err != nil {
		t.Fatalf("GetByEmail should succeed: %v", err)
	}

	if user.ID != createdUser.ID {
		t.Errorf("Expected ID %d, got %d", createdUser.ID, user.ID)
	}

	if user.Email != repoTestData.testEmail {
		t.Errorf("Expected email %s, got %s", repoTestData.testEmail, user.Email)
	}
}

// TestGetByEmailNotFound tests getting user by non-existent email
func TestGetByEmailNotFound(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	if err == nil {
		t.Error("GetByEmail should fail for non-existent email")
	}

	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got: %v", err)
	}
}

// TestEmailExists tests email existence check
func TestEmailExists(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	// Check non-existent email
	exists, err := repo.EmailExists(ctx, repoTestData.testEmail)
	if err != nil {
		t.Fatalf("EmailExists should not error: %v", err)
	}

	if exists {
		t.Error("Email should not exist initially")
	}

	// Create user
	input := CreateUserInput{
		Email:    repoTestData.testEmail,
		Password: repoTestData.testPassword,
	}

	_, err = repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Check existing email
	exists, err = repo.EmailExists(ctx, repoTestData.testEmail)
	if err != nil {
		t.Fatalf("EmailExists should not error: %v", err)
	}

	if !exists {
		t.Error("Email should exist after creation")
	}
}

// TestUpdateLastLogin tests last login update
func TestUpdateLastLogin(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	// Create user first
	input := CreateUserInput{
		Email:    repoTestData.testEmail,
		Password: repoTestData.testPassword,
	}

	user, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Initially LastLoginAt should be nil
	if user.LastLoginAt != nil {
		t.Error("LastLoginAt should be nil initially")
	}

	// Update last login
	err = repo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		t.Fatalf("UpdateLastLogin should succeed: %v", err)
	}

	// Get updated user
	updatedUser, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("GetByID should succeed: %v", err)
	}

	if updatedUser.LastLoginAt == nil {
		t.Error("LastLoginAt should be set after update")
	}

	if updatedUser.UpdatedAt.Before(user.UpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}
}

// TestAuthenticate tests user authentication
func TestAuthenticate(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	// Create user first
	input := CreateUserInput{
		Email:    repoTestData.testEmail,
		Password: repoTestData.testPassword,
	}

	_, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Test successful authentication
	user, err := repo.Authenticate(ctx, repoTestData.testEmail, repoTestData.testPassword)
	if err != nil {
		t.Fatalf("Authenticate should succeed: %v", err)
	}

	if user.Email != repoTestData.testEmail {
		t.Errorf("Expected email %s, got %s", repoTestData.testEmail, user.Email)
	}

	// LastLoginAt should be updated
	if user.LastLoginAt == nil {
		t.Error("LastLoginAt should be set after authentication")
	}
}

// TestAuthenticateInvalidCredentials tests authentication failures
func TestAuthenticateInvalidCredentials(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	// Create user first
	input := CreateUserInput{
		Email:    repoTestData.testEmail,
		Password: repoTestData.testPassword,
	}

	_, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	testCases := []struct {
		name     string
		email    string
		password string
	}{
		{"wrong email", "wrong@example.com", repoTestData.testPassword},
		{"wrong password", repoTestData.testEmail, "wrongpassword"},
		{"both wrong", "wrong@example.com", "wrongpassword"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := repo.Authenticate(ctx, tc.email, tc.password)
			if err == nil {
				t.Errorf("Authenticate should fail for %s", tc.name)
			}

			if !errors.Is(err, ErrInvalidCredentials) {
				t.Errorf("Expected ErrInvalidCredentials for %s, got: %v", tc.name, err)
			}
		})
	}
}

// TestRepositoryErrorHandling tests database error scenarios
func TestRepositoryErrorHandling(t *testing.T) {
	repo := NewMockUserRepositoryDB()
	ctx := context.Background()

	// Set mock to fail
	repo.mockDB.SetShouldFail(true)
	repo.mockDB.SetQueryError(errors.New("database connection failed"))

	input := CreateUserInput{
		Email:    repoTestData.testEmail,
		Password: repoTestData.testPassword,
	}

	// Test Create failure
	_, err := repo.Create(ctx, input)
	if err == nil {
		t.Error("Create should fail when database fails")
	}

	// Test GetByID failure
	_, err = repo.GetByID(ctx, 1)
	if err == nil {
		t.Error("GetByID should fail when database fails")
	}

	// Test GetByEmail failure
	_, err = repo.GetByEmail(ctx, repoTestData.testEmail)
	if err == nil {
		t.Error("GetByEmail should fail when database fails")
	}

	// Test EmailExists failure
	_, err = repo.EmailExists(ctx, repoTestData.testEmail)
	if err == nil {
		t.Error("EmailExists should fail when database fails")
	}
}
