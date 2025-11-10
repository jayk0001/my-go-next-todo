package todo

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockTodoRepository implements Repository Interface for testing
type MockTodoRepository struct {
	todos        map[int]*Todo
	todosByUser  map[int][]*Todo
	nextID       int
	shouldFail   bool
	failureError error
}

// NewMockTodoRepository creates a new mock repository
func NewMockTodoRepository() *MockTodoRepository {
	return &MockTodoRepository{
		todos:       make(map[int]*Todo),
		todosByUser: make(map[int][]*Todo),
		nextID:      1,
	}
}

// Create Implements Repository interface
func (m *MockTodoRepository) Create(ctx context.Context, userID int, input CreateTodoInput) (*Todo, error) {
	if m.shouldFail {
		return nil, m.failureError
	}

	todo := &Todo{
		ID:          m.nextID,
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		Completed:   false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	m.todos[todo.ID] = todo
	m.todosByUser[userID] = append(m.todosByUser[userID], todo)
	m.nextID++ // ✅ ID 자동 증가

	return todo, nil
}

// GetByID implements Repository interface
func (m *MockTodoRepository) GetByID(ctx context.Context, todoID, userID int) (*Todo, error) {
	if m.shouldFail {
		return nil, m.failureError
	}

	todo, exists := m.todos[todoID]
	if !exists || todo.UserID != userID {
		return nil, ErrTodoNotFound
	}

	return todo, nil
}

// GetByUserID implements Repository interface
func (m *MockTodoRepository) GetByUserID(ctx context.Context, userID int, filter TodoFilter) (*TodoListResponse, error) {
	if m.shouldFail {
		return nil, m.failureError
	}

	userTodos := m.todosByUser[userID]
	var filteredTodos []*Todo

	// Apply filters
	for _, todo := range userTodos {
		if filter.Completed != nil && todo.Completed != *filter.Completed {
			continue
		}
		if filter.Search != nil && *filter.Search != "" {
			if !containString(todo.Title, *filter.Search) {
				continue
			}
		}
		filteredTodos = append(filteredTodos, todo)
	}

	// Apply pagination
	total := len(filteredTodos)
	offset := filter.Offset
	limit := filter.Limit

	if limit == 0 {
		limit = total
	}

	end := offset + limit
	if end > total {
		end = total
	}

	if offset > total {
		offset = total
	}

	paginatedTodos := filteredTodos[offset:end]
	hasMore := end < total

	return &TodoListResponse{
		Todos:   paginatedTodos,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}, nil
}

// Update implements Repository interface
func (m *MockTodoRepository) Update(ctx context.Context, todoID, userID int, input UpdateTodoInput) (*Todo, error) {
	if m.shouldFail {
		return nil, m.failureError
	}

	todo, exists := m.todos[todoID]
	if !exists || todo.UserID != userID {
		return nil, ErrTodoNotFound
	}

	// Update fields
	if input.Title != nil {
		todo.Title = *input.Title
	}

	if input.Description != nil {
		todo.Description = input.Description
	}

	if input.Completed != nil {
		todo.Completed = *input.Completed
	}

	todo.UpdatedAt = time.Now()

	return todo, nil
}

// Delete implements Repository interface
func (m *MockTodoRepository) Delete(ctx context.Context, todoID, userID int) error {
	if m.shouldFail {
		return m.failureError
	}

	todo, exists := m.todos[todoID]
	if !exists || todo.UserID != userID {
		return ErrTodoNotFound
	}

	delete(m.todos, todoID)

	// Remove from user todos
	userTodos := m.todosByUser[userID]
	for i, t := range userTodos {
		if t.ID == todoID {
			m.todosByUser[userID] = append(userTodos[:i], userTodos[i+1:]...)
			break
		}
	}

	return nil
}

// ToggleComplete implements Repository interface
func (m *MockTodoRepository) ToggleComplete(ctx context.Context, todoID, userID int) (*Todo, error) {
	if m.shouldFail {
		return nil, m.failureError
	}

	todo, exists := m.todos[todoID]
	if !exists || todo.UserID != userID {
		return nil, ErrTodoNotFound
	}

	todo.Completed = !todo.Completed
	todo.UpdatedAt = time.Now()

	return todo, nil
}

// CountByUserID implements Repository interface
func (m *MockTodoRepository) CountByUserID(ctx context.Context, userID int) (int, error) {
	if m.shouldFail {
		return 0, m.failureError
	}

	return len(m.todosByUser[userID]), nil
}

// Mock control methods
func (m *MockTodoRepository) SetShouldFail(shouldFail bool, err error) {
	m.shouldFail = shouldFail
	m.failureError = err
}

func (m *MockTodoRepository) Clear() {
	m.todos = make(map[int]*Todo)
	m.todosByUser = make(map[int][]*Todo)
	m.nextID = 1
	m.shouldFail = false
	m.failureError = nil
}

// ============================================================================
// Helper functions for pointer creation
// ============================================================================

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

// ============================================================================
// Test data
// ============================================================================

type RepositoryTestData struct {
	testUserID  int
	testUserID2 int
	testTitle   string
	testTitle2  string
}

var repoTestData = RepositoryTestData{
	testUserID:  1,
	testUserID2: 2,
	testTitle:   "Test Todo",
	testTitle2:  "Updated Todo",
}

// ============================================================================
// Tests - Create
// ============================================================================

func TestCreateTodo(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	input := CreateTodoInput{
		Title:       repoTestData.testTitle,
		Description: stringPtr("Test Description"),
	}

	todo, err := repo.Create(ctx, repoTestData.testUserID, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	if todo.ID <= 0 {
		t.Error("Todo ID should be positive")
	}

	if todo.UserID != repoTestData.testUserID {
		t.Errorf("Expected user ID %d, got %d", repoTestData.testUserID, todo.UserID)
	}

	if todo.Title != repoTestData.testTitle {
		t.Errorf("Expected title %s, got %s", repoTestData.testTitle, todo.Title)
	}

	if todo.Completed {
		t.Error("New todo should not be completed")
	}
}

func TestCreateTodoWithNextID(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create first todo
	input1 := CreateTodoInput{Title: "First Todo"}
	todo1, err := repo.Create(ctx, 1, input1)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Create second todo
	input2 := CreateTodoInput{Title: "Second Todo"}
	todo2, err := repo.Create(ctx, 1, input2)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Verify ID increment
	if todo2.ID != todo1.ID+1 {
		t.Errorf("Expected ID %d, got %d", todo1.ID+1, todo2.ID)
	}
}

// ============================================================================
// Tests - Read
// ============================================================================

func TestGetByID(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create todo first
	input := CreateTodoInput{Title: repoTestData.testTitle}
	createdTodo, err := repo.Create(ctx, repoTestData.testUserID, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Get todo
	todo, err := repo.GetByID(ctx, createdTodo.ID, repoTestData.testUserID)
	if err != nil {
		t.Fatalf("GetByID should succeed: %v", err)
	}

	if todo.ID != createdTodo.ID {
		t.Errorf("Expected ID %d, got %d", createdTodo.ID, todo.ID)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999, repoTestData.testUserID)
	if err != ErrTodoNotFound {
		t.Errorf("Expected ErrTodoNotFound, got %v", err)
	}
}

func TestGetByIDWrongUser(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create todo for user 1
	input := CreateTodoInput{Title: repoTestData.testTitle}
	createdTodo, err := repo.Create(ctx, repoTestData.testUserID, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Try to get with user 2
	_, err = repo.GetByID(ctx, createdTodo.ID, repoTestData.testUserID2)
	if err != ErrTodoNotFound {
		t.Errorf("Expected ErrTodoNotFound for wrong user, got: %v", err)
	}
}

// ============================================================================
// Tests - Update
// ============================================================================

func TestUpdateTodo(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create todo first
	input := CreateTodoInput{Title: repoTestData.testTitle}
	createdTodo, err := repo.Create(ctx, repoTestData.testUserID, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Update todo
	updateInput := UpdateTodoInput{
		Title:     stringPtr(repoTestData.testTitle2),
		Completed: boolPtr(true),
	}

	updatedTodo, err := repo.Update(ctx, createdTodo.ID, repoTestData.testUserID, updateInput)
	if err != nil {
		t.Fatalf("Update should succeed: %v", err)
	}

	if updatedTodo.Title != repoTestData.testTitle2 {
		t.Errorf("Expected title %s, got %s", repoTestData.testTitle2, updatedTodo.Title)
	}

	if !updatedTodo.Completed {
		t.Error("Todo should be completed")
	}
}

func TestUpdateTodoNotFound(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	updateInput := UpdateTodoInput{
		Title: stringPtr("Updated"),
	}

	_, err := repo.Update(ctx, 999, 1, updateInput)
	if err != ErrTodoNotFound {
		t.Errorf("Expected ErrTodoNotFound, got %v", err)
	}
}

// ============================================================================
// Tests - Delete
// ============================================================================

func TestDeleteTodo(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create todo first
	input := CreateTodoInput{Title: repoTestData.testTitle}
	createdTodo, err := repo.Create(ctx, repoTestData.testUserID, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Delete todo
	err = repo.Delete(ctx, createdTodo.ID, repoTestData.testUserID)
	if err != nil {
		t.Fatalf("Delete should succeed: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(ctx, createdTodo.ID, repoTestData.testUserID)
	if err != ErrTodoNotFound {
		t.Error("Todo should be deleted")
	}
}

func TestDeleteNotFound(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	err := repo.Delete(ctx, 999, 1)
	if err != ErrTodoNotFound {
		t.Errorf("Expected ErrTodoNotFound, got %v", err)
	}
}

// ============================================================================
// Tests - Toggle
// ============================================================================

func TestToggleComplete(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create todo first
	input := CreateTodoInput{Title: repoTestData.testTitle}
	createdTodo, err := repo.Create(ctx, repoTestData.testUserID, input)
	if err != nil {
		t.Fatalf("Create should succeed: %v", err)
	}

	// Toggle completion (false -> true)
	toggledTodo, err := repo.ToggleComplete(ctx, createdTodo.ID, repoTestData.testUserID)
	if err != nil {
		t.Fatalf("Toggle should succeed: %v", err)
	}

	if !toggledTodo.Completed {
		t.Error("Todo should be completed after toggle")
	}

	// Toggle again (true -> false)
	toggledTodo, err = repo.ToggleComplete(ctx, createdTodo.ID, repoTestData.testUserID)
	if err != nil {
		t.Fatalf("Second toggle should succeed: %v", err)
	}

	if toggledTodo.Completed {
		t.Error("Todo should not be completed after second toggle")
	}
}

// ============================================================================
// Tests - Filtering & Pagination
// ============================================================================

func TestGetByUserIDWithFilters(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create multiple todos
	todos := []struct {
		title     string
		completed bool
	}{
		{"Todo 1", false},
		{"Todo 2", true},
		{"Important task", false},
	}

	for _, todo := range todos {
		input := CreateTodoInput{Title: todo.title}
		createdTodo, err := repo.Create(ctx, repoTestData.testUserID, input)
		if err != nil {
			t.Fatalf("Create should succeed: %v", err)
		}

		if todo.completed {
			_, err = repo.ToggleComplete(ctx, createdTodo.ID, repoTestData.testUserID)
			if err != nil {
				t.Fatalf("Toggle should succeed: %v", err)
			}
		}
	}

	// Test completed filter
	completedFilter := TodoFilter{Completed: boolPtr(true)}
	result, err := repo.GetByUserID(ctx, repoTestData.testUserID, completedFilter)
	if err != nil {
		t.Fatalf("GetByUserID should succeed: %v", err)
	}

	if len(result.Todos) != 1 {
		t.Errorf("Expected 1 completed todo, got %d", len(result.Todos))
	}

	// Test search filter
	searchFilter := TodoFilter{Search: stringPtr("Important")}
	result, err = repo.GetByUserID(ctx, repoTestData.testUserID, searchFilter)
	if err != nil {
		t.Fatalf("GetByUserID should succeed: %v", err)
	}

	if len(result.Todos) != 1 {
		t.Errorf("Expected 1 searched todo, got %d", len(result.Todos))
	}
}

func TestGetByUserIDPagination(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Create 5 todos
	for i := 1; i <= 5; i++ {
		input := CreateTodoInput{
			Title: fmt.Sprintf("Todo %d", i),
		}
		_, err := repo.Create(ctx, 1, input)
		if err != nil {
			t.Fatalf("Create should succeed: %v", err)
		}
	}

	// First page (limit=2, offset=0)
	filter := TodoFilter{Limit: 2, Offset: 0}
	result, err := repo.GetByUserID(ctx, 1, filter)
	if err != nil {
		t.Fatalf("GetByUserID should succeed: %v", err)
	}

	if len(result.Todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(result.Todos))
	}

	if !result.HasMore {
		t.Error("Should have more todos")
	}

	if result.Total != 5 {
		t.Errorf("Expected total 5, got %d", result.Total)
	}

	// Second page
	filter = TodoFilter{Limit: 2, Offset: 2}
	result, err = repo.GetByUserID(ctx, 1, filter)
	if err != nil {
		t.Fatalf("GetByUserID should succeed: %v", err)
	}

	if len(result.Todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(result.Todos))
	}

	// Last page
	filter = TodoFilter{Limit: 2, Offset: 4}
	result, err = repo.GetByUserID(ctx, 1, filter)
	if err != nil {
		t.Fatalf("GetByUserID should succeed: %v", err)
	}

	if len(result.Todos) != 1 {
		t.Errorf("Expected 1 todo, got %d", len(result.Todos))
	}

	if result.HasMore {
		t.Error("Should not have more todos")
	}
}

// ============================================================================
// Tests - Mock Control
// ============================================================================

func TestMockFailure(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	testErr := fmt.Errorf("mock database error")
	repo.SetShouldFail(true, testErr)

	// Test Create failure
	input := CreateTodoInput{Title: "Test"}
	_, err := repo.Create(ctx, 1, input)
	if err != testErr {
		t.Errorf("Expected mock error, got %v", err)
	}

	// Test GetByID failure
	_, err = repo.GetByID(ctx, 1, 1)
	if err != testErr {
		t.Errorf("Expected mock error, got %v", err)
	}

	// Test GetByUserID failure
	_, err = repo.GetByUserID(ctx, 1, TodoFilter{})
	if err != testErr {
		t.Errorf("Expected mock error, got %v", err)
	}
}

func TestCountByUserID(t *testing.T) {
	repo := NewMockTodoRepository()
	ctx := context.Background()

	// Initially should be 0
	count, err := repo.CountByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("CountByUserID should succeed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Create 3 todos
	for i := 1; i <= 3; i++ {
		input := CreateTodoInput{Title: fmt.Sprintf("Todo %d", i)}
		_, err := repo.Create(ctx, 1, input)
		if err != nil {
			t.Fatalf("Create should succeed: %v", err)
		}
	}

	// Count should be 3
	count, err = repo.CountByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("CountByUserID should succeed: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

// ============================================================================
// Helper function for string search
// ============================================================================

func containString(str, substr string) bool {
	if len(substr) == 0 {
		return true
	}

	if len(str) < len(substr) {
		return false
	}

	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
