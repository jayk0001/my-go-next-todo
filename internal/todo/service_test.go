package todo

import (
	"context"
	"errors"
	"testing"
)

// ============================================================================
// Test Setup
// ============================================================================

type serviceTestSetup struct {
	service   *TodoService
	repo      *MockTodoRepository
	validator *ValidatorService
	ctx       context.Context
	userID    int
}

func newServiceTestSetup() *serviceTestSetup {
	repo := NewMockTodoRepository()
	validator := NewValidatorService()
	service := NewTodoService(repo, validator)

	return &serviceTestSetup{
		service:   service,
		repo:      repo,
		validator: validator,
		ctx:       context.Background(),
		userID:    1,
	}
}

// ============================================================================
// Tests - CreateTodo
// ============================================================================

func TestServiceCreateTodo(t *testing.T) {
	setup := newServiceTestSetup()

	input := CreateTodoInput{
		Title:       "Test Todo",
		Description: stringPtr("Test Description"),
	}

	todo, err := setup.service.CreateTodo(setup.ctx, setup.userID, input)
	if err != nil {
		t.Fatalf("CreateTodo should succeed: %v", err)
	}

	if todo.UserID != setup.userID {
		t.Errorf("Expected userID %d, got %d", setup.userID, todo.UserID)
	}

	if todo.Title != input.Title {
		t.Errorf("Expected title %s, got %s", input.Title, todo.Title)
	}
}

func TestServiceCreateTodoValidationError(t *testing.T) {
	setup := newServiceTestSetup()

	// Invalid input - empty title
	input := CreateTodoInput{
		Title: "",
	}

	_, err := setup.service.CreateTodo(setup.ctx, setup.userID, input)
	if err == nil {
		t.Fatal("CreateTodo should fail with empty title")
	}

	// Validator should return error
	if !errors.Is(err, ErrInvalidTodoInput) && !errors.Is(err, ErrTodoTitleRequired) {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

func TestServiceCreateTodoRepositoryError(t *testing.T) {
	setup := newServiceTestSetup()

	// Mock repository failure
	testErr := errors.New("database connection failed")
	setup.repo.SetShouldFail(true, testErr)

	input := CreateTodoInput{
		Title: "Test Todo",
	}

	_, err := setup.service.CreateTodo(setup.ctx, setup.userID, input)
	if err == nil {
		t.Fatal("CreateTodo should fail when repository fails")
	}

	// Check if original error wrapped
	if !errors.Is(err, testErr) {
		t.Errorf("Expected wrapped error containing %v, got: %v", testErr, err)
	}
}

// ============================================================================
// Tests - GetTodo
// ============================================================================

func TestServiceGetTodo(t *testing.T) {
	setup := newServiceTestSetup()

	// Create a todo first
	input := CreateTodoInput{Title: "Test Todo"}
	createdTodo, err := setup.repo.Create(setup.ctx, setup.userID, input)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Get the todo
	todo, err := setup.service.GetTodo(setup.ctx, createdTodo.ID, setup.userID)
	if err != nil {
		t.Fatalf("GetTodo should succeed: %v", err)
	}

	if todo.ID != createdTodo.ID {
		t.Errorf("Expected ID %d, got %d", createdTodo.ID, todo.ID)
	}
}

func TestServiceGetTodoInvalidID(t *testing.T) {
	setup := newServiceTestSetup()

	tests := []struct {
		name   string
		todoID int
	}{
		{"zero ID", 0},
		{"negative ID", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := setup.service.GetTodo(setup.ctx, tt.todoID, setup.userID)
			if err == nil {
				t.Fatal("GetTodo should fail with invalid ID")
			}

			if !errors.Is(err, ErrInvalidTodoInput) {
				t.Errorf("Expected ErrInvalidTodoInput, got: %v", err)
			}
		})
	}
}

func TestServiceGetTodoNotFound(t *testing.T) {
	setup := newServiceTestSetup()

	_, err := setup.service.GetTodo(setup.ctx, 999, setup.userID)
	if err == nil {
		t.Fatal("GetTodo should fail for non-existent todo")
	}

	// ErrTodoNotFound from Repository should return
	if !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("Expected ErrTodoNotFound, got: %v", err)
	}
}

func TestServiceGetTodoWrongUser(t *testing.T) {
	setup := newServiceTestSetup()

	// Create user 1's todo
	input := CreateTodoInput{Title: "User 1 Todo"}
	createdTodo, _ := setup.repo.Create(setup.ctx, 1, input)

	// Create user 2's todo
	_, err := setup.service.GetTodo(setup.ctx, createdTodo.ID, 2)
	if err == nil {
		t.Fatal("GetTodo should fail for different user")
	}

	if !errors.Is(err, ErrTodoNotFound) {
		t.Errorf("Expected ErrTodoNotFound, got: %v", err)
	}
}
