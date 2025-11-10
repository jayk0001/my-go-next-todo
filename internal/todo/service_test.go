package todo

import (
	"context"
	"errors"
	"fmt"
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

// ============================================================================
// Tests - GetUserTodos
// ============================================================================

func TestServiceGetUserTodos(t *testing.T) {
	setup := newServiceTestSetup()

	// Create multiple todos
	for i := 1; i <= 3; i++ {
		input := CreateTodoInput{
			Title: *stringPtr(fmt.Sprintf("TODO %d", i)),
		}
		_, err := setup.repo.Create(setup.ctx, setup.userID, input)
		if err != nil {
			t.Fatalf("Setup failed: %v", err)
		}
	}

	// Get all todos
	filter := TodoFilter{}
	res, err := setup.service.GetUserTodos(setup.ctx, setup.userID, filter)

	if err != nil {
		t.Fatalf("GetUserTodos should succeed: %v", err)
	}

	if len(res.Todos) != 3 {
		t.Errorf("Expected 3 todos, got %d", len(res.Todos))
	}

	if res.Total != 3 {
		t.Errorf("Expected total 3 todos, got %d", res.Total)
	}
}

func TestServiceGetUserTodosWithCompletedFilter(t *testing.T) {
	// set up -> create todos with different completion status -> filter completed todos
	setup := newServiceTestSetup()

	// Create todos with different completion status
	todos := []struct {
		title     string
		completed bool
	}{
		{"todo1", true},
		{"todo2", false},
		{"todo3", true},
	}

	for _, todo := range todos {
		input := CreateTodoInput{Title: todo.title}
		created, _ := setup.service.CreateTodo(setup.ctx, setup.userID, input)
		if todo.completed {
			setup.service.ToggleTodoComplete(setup.ctx, created.ID, setup.userID)
		}
	}

	// Filter completed todos
	filter := TodoFilter{Completed: boolPtr(true)}
	res, err := setup.service.GetUserTodos(setup.ctx, setup.userID, filter)
	if err != nil {
		t.Fatalf("GetUserTodos should succeed: %v", err)
	}

	if len(res.Todos) != 2 {
		t.Errorf("Expected 2 completed todo, got %d", len(res.Todos))
	}
}

func TestServiceGetUserTodosWithPaginaton(t *testing.T) {
	setup := newServiceTestSetup()

	for i := 1; i <= 5; i++ {
		input := CreateTodoInput{
			Title: fmt.Sprintf("Todo %d", i),
		}
		_, err := setup.service.CreateTodo(setup.ctx, setup.userID, input)
		if err != nil {
			t.Fatalf("CreateTodo should succeed: %v", err)
		}
	}

	// Get first page (limit:2, Offset: 0)
	filter := TodoFilter{Limit: 2, Offset: 0}
	res, err := setup.service.GetUserTodos(setup.ctx, setup.userID, filter)
	if err != nil {
		t.Fatalf("GetUserTodos should succeed: %v", err)
	}

	if len(res.Todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(res.Todos))
	}

	if !res.HasMore {
		t.Error("Should have more todos")
	}

	if res.Total != 5 {
		t.Errorf("Expected 5 todos, got %d", res.Total)
	}
}

func TestServiceGetUserTodosNormalizerFilter(t *testing.T) {
	// setup -> create tests types/variants -> tests for loop
	setup := newServiceTestSetup()

	tests := []struct {
		name           string
		inputFilter    TodoFilter
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "default Limit",
			inputFilter:    TodoFilter{},
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name:           "exceed Max Limit",
			inputFilter:    TodoFilter{Limit: 200},
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name:           "negative Offset",
			inputFilter:    TodoFilter{Offset: -5},
			expectedLimit:  20,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := setup.service.GetUserTodos(setup.ctx, setup.userID, tt.inputFilter)
			if err != nil {
				t.Fatalf("GetUserTodos should succeed: %v", err)
			}

			if res.Limit != tt.expectedLimit {
				t.Errorf("Expected Limit %d, got %d", tt.expectedLimit, res.Limit)
			}

			if res.Offset != tt.expectedOffset {
				t.Errorf("Expected Offset %d, got %d", tt.expectedOffset, res.Offset)
			}
		})
	}
}

// ============================================================================
// Tests - UpdateTodo
// ============================================================================

func TestServiceUpdateTodo(t *testing.T) {
	setup := newServiceTestSetup()

	// Create a todo
	input := CreateTodoInput{Title: "Original Title"}
	createdTodo, _ := setup.repo.Create(setup.ctx, setup.userID, input)

	// Update it
	updateInput := UpdateTodoInput{
		Title:     stringPtr("Updated Title"),
		Completed: boolPtr(true),
	}

	updatedTodo, err := setup.service.UpdateTodo(setup.ctx, createdTodo.ID, setup.userID, updateInput)
	if err != nil {
		t.Fatalf("UpdateTodo should succeed: %v", err)
	}

	if updatedTodo.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", updatedTodo.Title)
	}

	if !updatedTodo.Completed {
		t.Error("Todo should be completed")
	}
}

func TestServiceUpdateTodoInvalidID(t *testing.T) {
	setup := newServiceTestSetup()

	updateInput := UpdateTodoInput{
		Title: stringPtr("Updated"),
	}

	_, err := setup.service.UpdateTodo(setup.ctx, 0, setup.userID, updateInput)
	if err == nil {
		t.Fatal("UpdateTodo should fail with invalid ID")
	}

	if !errors.Is(err, ErrInvalidTodoInput) {
		t.Errorf("Expected ErrInvalidTodoInput, got: %v", err)
	}
}

func TestServiceUpdateTodoEmptyInput(t *testing.T) {
	setup := newServiceTestSetup()

	// Create a todo
	input := CreateTodoInput{Title: "Test"}
	createdTodo, _ := setup.repo.Create(setup.ctx, setup.userID, input)

	// Try to update with empty input
	updateInput := UpdateTodoInput{}

	_, err := setup.service.UpdateTodo(setup.ctx, createdTodo.ID, setup.userID, updateInput)
	if err == nil {
		t.Fatal("UpdateTodo should fail with empty input")
	}

	if !errors.Is(err, ErrInvalidTodoInput) {
		t.Errorf("Expected ErrInvalidTodoInput, got: %v", err)
	}
}

func TestServiceUpdateTodoTitleTooLong(t *testing.T) {
	setup := newServiceTestSetup()

	// Create a todo
	input := CreateTodoInput{Title: "Test"}
	createdTodo, _ := setup.repo.Create(setup.ctx, setup.userID, input)

	// Try to update with title too long
	longTitle := make([]byte, 501)
	for i := range longTitle {
		longTitle[i] = 'a'
	}

	updateInput := UpdateTodoInput{
		Title: stringPtr(string(longTitle)),
	}

	_, err := setup.service.UpdateTodo(setup.ctx, createdTodo.ID, setup.userID, updateInput)
	if err == nil {
		t.Fatal("UpdateTodo should fail with title too long")
	}

	if !errors.Is(err, ErrTodoTitleTooLong) {
		t.Errorf("Expected ErrTodoTitleTooLong, got: %v", err)
	}
}

func TestServiceUpdateTodoNotFound(t *testing.T) {
	setup := newServiceTestSetup()

	updateInput := UpdateTodoInput{
		Title: stringPtr("Updated"),
	}

	_, err := setup.service.UpdateTodo(setup.ctx, 999, setup.userID, updateInput)
	if err == nil {
		t.Fatal("UpdateTodo should fail for non-existent todo")
	}

	if !errors.Is(err, ErrTodoAccessDenied) {
		t.Errorf("Expected ErrTodoAccessDenied, got: %v", err)
	}
}

func TestServiceUpdateTodoWrongUser(t *testing.T) {
	setup := newServiceTestSetup()

	// User 1 creates a todo
	input := CreateTodoInput{Title: "User 1 Todo"}
	createdTodo, _ := setup.repo.Create(setup.ctx, 1, input)

	// User 2 tries to update it
	updateInput := UpdateTodoInput{
		Title: stringPtr("Hacked"),
	}

	_, err := setup.service.UpdateTodo(setup.ctx, createdTodo.ID, 2, updateInput)
	if err == nil {
		t.Fatal("UpdateTodo should fail for different user")
	}

	if !errors.Is(err, ErrTodoAccessDenied) {
		t.Errorf("Expected ErrTodoAccessDenied, got: %v", err)
	}
}

// ============================================================================
// Tests - DeleteTodo
// ============================================================================

func TestServiceDeleteTodo(t *testing.T) {
	setup := newServiceTestSetup()

	// Create a todo
	input := CreateTodoInput{Title: "Test Todo"}
	createdTodo, _ := setup.repo.Create(setup.ctx, setup.userID, input)

	// Delete it
	err := setup.service.DeleteTodo(setup.ctx, createdTodo.ID, setup.userID)
	if err != nil {
		t.Fatalf("DeleteTodo should succeed: %v", err)
	}

	// Verify deletion
	_, err = setup.repo.GetByID(setup.ctx, createdTodo.ID, setup.userID)
	if err != ErrTodoNotFound {
		t.Error("Todo should be deleted")
	}
}

func TestServiceDeleteTodoInvalidID(t *testing.T) {
	setup := newServiceTestSetup()

	err := setup.service.DeleteTodo(setup.ctx, 0, setup.userID)
	if err == nil {
		t.Fatal("DeleteTodo should fail with invalid ID")
	}

	if !errors.Is(err, ErrInvalidTodoInput) {
		t.Errorf("Expected ErrInvalidTodoInput, got: %v", err)
	}
}

func TestServiceDeleteTodoNotFound(t *testing.T) {
	setup := newServiceTestSetup()

	err := setup.service.DeleteTodo(setup.ctx, 999, setup.userID)
	if err == nil {
		t.Fatal("DeleteTodo should fail for non-existent todo")
	}

	if !errors.Is(err, ErrTodoAccessDenied) {
		t.Errorf("Expected ErrTodoAccessDenied, got: %v", err)
	}
}

func TestServiceDeleteTodoWrongUser(t *testing.T) {
	setup := newServiceTestSetup()

	// User 1 creates a todo
	input := CreateTodoInput{Title: "User 1 Todo"}
	createdTodo, _ := setup.repo.Create(setup.ctx, 1, input)

	// User 2 tries to delete it
	err := setup.service.DeleteTodo(setup.ctx, createdTodo.ID, 2)
	if err == nil {
		t.Fatal("DeleteTodo should fail for different user")
	}

	if !errors.Is(err, ErrTodoAccessDenied) {
		t.Errorf("Expected ErrTodoAccessDenied, got: %v", err)
	}
}

// ============================================================================
// Tests - ToggleTodoComplete
// ============================================================================

func TestServiceToggleTodoComplete(t *testing.T) {
	setup := newServiceTestSetup()

	// Create a todo
	input := CreateTodoInput{Title: "Test Todo"}
	createdTodo, _ := setup.repo.Create(setup.ctx, setup.userID, input)

	// Toggle to completed
	toggledTodo, err := setup.service.ToggleTodoComplete(setup.ctx, createdTodo.ID, setup.userID)
	if err != nil {
		t.Fatalf("ToggleTodoComplete should succeed: %v", err)
	}

	if !toggledTodo.Completed {
		t.Error("Todo should be completed after toggle")
	}

	// Toggle back to incomplete
	toggledTodo, err = setup.service.ToggleTodoComplete(setup.ctx, createdTodo.ID, setup.userID)
	if err != nil {
		t.Fatalf("Second toggle should succeed: %v", err)
	}

	if toggledTodo.Completed {
		t.Error("Todo should not be completed after second toggle")
	}
}

func TestServiceToggleTodoCompleteInvalidID(t *testing.T) {
	setup := newServiceTestSetup()

	_, err := setup.service.ToggleTodoComplete(setup.ctx, 0, setup.userID)
	if err == nil {
		t.Fatal("ToggleTodoComplete should fail with invalid ID")
	}

	if !errors.Is(err, ErrInvalidTodoInput) {
		t.Errorf("Expected ErrInvalidTodoInput, got: %v", err)
	}
}

func TestServiceToggleTodoCompleteNotFound(t *testing.T) {
	setup := newServiceTestSetup()

	_, err := setup.service.ToggleTodoComplete(setup.ctx, 999, setup.userID)
	if err == nil {
		t.Fatal("ToggleTodoComplete should fail for non-existent todo")
	}

	if !errors.Is(err, ErrTodoAccessDenied) {
		t.Errorf("Expected ErrTodoAccessDenied, got: %v", err)
	}
}

func TestServiceToggleTodoCompleteWrongUser(t *testing.T) {
	setup := newServiceTestSetup()

	// User 1 creates a todo
	input := CreateTodoInput{Title: "User 1 Todo"}
	createdTodo, _ := setup.repo.Create(setup.ctx, 1, input)

	// User 2 tries to toggle it
	_, err := setup.service.ToggleTodoComplete(setup.ctx, createdTodo.ID, 2)
	if err == nil {
		t.Fatal("ToggleTodoComplete should fail for different user")
	}

	if !errors.Is(err, ErrTodoAccessDenied) {
		t.Errorf("Expected ErrTodoAccessDenied, got: %v", err)
	}
}

// ============================================================================
// Tests - GetUserTodoStats
// ============================================================================

func TestServiceGetUserTodoStats(t *testing.T) {
	setup := newServiceTestSetup()

	// Create todos with different completion status
	todos := []struct {
		title     string
		completed bool
	}{
		{"Todo 1", false},
		{"Todo 2", true},
		{"Todo 3", true},
		{"Todo 4", false},
		{"Todo 5", false},
	}

	for _, todo := range todos {
		input := CreateTodoInput{Title: todo.title}
		created, _ := setup.repo.Create(setup.ctx, setup.userID, input)
		if todo.completed {
			setup.repo.ToggleComplete(setup.ctx, created.ID, setup.userID)
		}
	}

	// Get stats
	stats, err := setup.service.GetUserTodoStats(setup.ctx, setup.userID)
	if err != nil {
		t.Fatalf("GetUserTodoStats should succeed: %v", err)
	}

	if stats.Total != 5 {
		t.Errorf("Expected total 5, got %d", stats.Total)
	}

	if stats.Completed != 2 {
		t.Errorf("Expected completed 2, got %d", stats.Completed)
	}

	if stats.Pending != 3 {
		t.Errorf("Expected pending 3, got %d", stats.Pending)
	}
}

func TestServiceGetUserTodoStatsEmpty(t *testing.T) {
	setup := newServiceTestSetup()

	// Get stats for user with no todos
	stats, err := setup.service.GetUserTodoStats(setup.ctx, setup.userID)
	if err != nil {
		t.Fatalf("GetUserTodoStats should succeed: %v", err)
	}

	if stats.Total != 0 {
		t.Errorf("Expected total 0, got %d", stats.Total)
	}

	if stats.Completed != 0 {
		t.Errorf("Expected completed 0, got %d", stats.Completed)
	}

	if stats.Pending != 0 {
		t.Errorf("Expected pending 0, got %d", stats.Pending)
	}
}

// ============================================================================
// Tests - BatchUpdateTodos
// ============================================================================

func TestServiceBatchUpdateTodos(t *testing.T) {
	setup := newServiceTestSetup()

	// Create multiple todos
	var todoIDs []int
	for i := 1; i <= 3; i++ {
		input := CreateTodoInput{
			Title: fmt.Sprintf("Todo %d", i),
		}
		created, _ := setup.repo.Create(setup.ctx, setup.userID, input)
		todoIDs = append(todoIDs, created.ID)
	}

	// Batch update
	updateInput := UpdateTodoInput{
		Completed: boolPtr(true),
	}

	updatedTodos, err := setup.service.BatchUpdateTodos(setup.ctx, setup.userID, todoIDs, updateInput)
	if err != nil {
		t.Fatalf("BatchUpdateTodos should succeed: %v", err)
	}

	if len(updatedTodos) != 3 {
		t.Errorf("Expected 3 updated todos, got %d", len(updatedTodos))
	}

	// Verify all are completed
	for _, todo := range updatedTodos {
		if !todo.Completed {
			t.Errorf("Todo %d should be completed", todo.ID)
		}
	}
}

func TestServiceBatchUpdateTodosEmptyList(t *testing.T) {
	setup := newServiceTestSetup()

	updateInput := UpdateTodoInput{
		Title: stringPtr("Updated"),
	}

	updatedTodos, err := setup.service.BatchUpdateTodos(setup.ctx, setup.userID, []int{}, updateInput)
	if err != nil {
		t.Fatalf("BatchUpdateTodos should succeed with empty list: %v", err)
	}

	if len(updatedTodos) != 0 {
		t.Errorf("Expected 0 updated todos, got %d", len(updatedTodos))
	}
}

func TestServiceBatchUpdateTodosInvalidInput(t *testing.T) {
	setup := newServiceTestSetup()

	// Create a todo
	input := CreateTodoInput{Title: "Test"}
	created, _ := setup.repo.Create(setup.ctx, setup.userID, input)

	// Try batch update with invalid input
	updateInput := UpdateTodoInput{} // Empty

	_, err := setup.service.BatchUpdateTodos(setup.ctx, setup.userID, []int{created.ID}, updateInput)
	if err == nil {
		t.Fatal("BatchUpdateTodos should fail with invalid input")
	}

	if !errors.Is(err, ErrInvalidTodoInput) {
		t.Errorf("Expected ErrInvalidTodoInput, got: %v", err)
	}
}

func TestServiceBatchUpdateTodosPartialFailure(t *testing.T) {
	setup := newServiceTestSetup()

	// Create 2 todos for user 1
	input := CreateTodoInput{Title: "User 1 Todo"}
	created1, _ := setup.repo.Create(setup.ctx, 1, input)
	created2, _ := setup.repo.Create(setup.ctx, 1, input)

	// Create 1 todo for user 2
	created3, _ := setup.repo.Create(setup.ctx, 2, input)

	// User 1 tries to batch update all 3
	// Expected: first 2 succeed, last one fails
	updateInput := UpdateTodoInput{
		Completed: boolPtr(true),
	}

	updatedTodos, err := setup.service.BatchUpdateTodos(
		setup.ctx,
		1,
		[]int{created1.ID, created2.ID, created3.ID},
		updateInput,
	)

	// Should return partial results with error
	if err == nil {
		t.Fatal("BatchUpdateTodos should fail partially")
	}

	// Two should succeed (user 1's todos)
	if len(updatedTodos) != 2 {
		t.Errorf("Expected 2 updated todos (partial success), got %d", len(updatedTodos))
	}

	// Verify all updated todos belong to user 1
	for _, todo := range updatedTodos {
		if todo.UserID != 1 {
			t.Errorf("Updated todo should belong to user 1, got user %d", todo.UserID)
		}
		if !todo.Completed {
			t.Error("Updated todo should be completed")
		}
	}
}
