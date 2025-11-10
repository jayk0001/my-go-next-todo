package resolver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/jayk0001/my-go-next-todo/internal/auth"
	"github.com/jayk0001/my-go-next-todo/internal/graphql/generated"
	"github.com/jayk0001/my-go-next-todo/internal/middleware" // Add this import
	"github.com/jayk0001/my-go-next-todo/internal/todo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockTodoService implements TodoServiceInterface for testing
type MockTodoService struct {
	CreateTodoFn         func(ctx context.Context, userID int, input todo.CreateTodoInput) (*todo.Todo, error)
	GetUserTodosFn       func(ctx context.Context, userID int, filter todo.TodoFilter) (*todo.TodoListResponse, error)
	GetTodoFn            func(ctx context.Context, todoID, userID int) (*todo.Todo, error)
	UpdateTodoFn         func(ctx context.Context, todoID, userID int, input todo.UpdateTodoInput) (*todo.Todo, error)
	DeleteTodoFn         func(ctx context.Context, todoID, userID int) error
	ToggleTodoCompleteFn func(ctx context.Context, todoID, userID int) (*todo.Todo, error)
	BatchUpdateTodosFn   func(ctx context.Context, userID int, todoIDs []int, input todo.UpdateTodoInput) ([]*todo.Todo, error)
	GetUserTodoStatsFn   func(ctx context.Context, userID int) (*todo.TodoStats, error)
}

// CreateTodo mock
func (m *MockTodoService) CreateTodo(ctx context.Context, userID int, input todo.CreateTodoInput) (*todo.Todo, error) {
	if m.CreateTodoFn != nil {
		return m.CreateTodoFn(ctx, userID, input)
	}
	return nil, errors.New("not implemented")
}

// GetUserTodos mock
func (m *MockTodoService) GetUserTodos(ctx context.Context, userID int, filter todo.TodoFilter) (*todo.TodoListResponse, error) {
	if m.GetUserTodosFn != nil {
		return m.GetUserTodosFn(ctx, userID, filter)
	}
	return nil, errors.New("not implemented")
}

// GetTodo mock
func (m *MockTodoService) GetTodo(ctx context.Context, todoID, userID int) (*todo.Todo, error) {
	if m.GetTodoFn != nil {
		return m.GetTodoFn(ctx, todoID, userID)
	}
	return nil, errors.New("not implemented")
}

// UpdateTodo mock
func (m *MockTodoService) UpdateTodo(ctx context.Context, todoID, userID int, input todo.UpdateTodoInput) (*todo.Todo, error) {
	if m.UpdateTodoFn != nil {
		return m.UpdateTodoFn(ctx, todoID, userID, input)
	}
	return nil, errors.New("not implemented")
}

// DeleteTodo mock
func (m *MockTodoService) DeleteTodo(ctx context.Context, todoID, userID int) error {
	if m.DeleteTodoFn != nil {
		return m.DeleteTodoFn(ctx, todoID, userID)
	}
	return errors.New("not implemented")
}

// ToggleTodoComplete mock
func (m *MockTodoService) ToggleTodoComplete(ctx context.Context, todoID, userID int) (*todo.Todo, error) {
	if m.ToggleTodoCompleteFn != nil {
		return m.ToggleTodoCompleteFn(ctx, todoID, userID)
	}
	return nil, errors.New("not implemented")
}

// BatchUpdateTodos mock
func (m *MockTodoService) BatchUpdateTodos(ctx context.Context, userID int, todoIDs []int, input todo.UpdateTodoInput) ([]*todo.Todo, error) {
	if m.BatchUpdateTodosFn != nil {
		return m.BatchUpdateTodosFn(ctx, userID, todoIDs, input)
	}
	return nil, errors.New("not implemented")
}

// GetUserTodoStats mock
func (m *MockTodoService) GetUserTodoStats(ctx context.Context, userID int) (*todo.TodoStats, error) {
	if m.GetUserTodoStatsFn != nil {
		return m.GetUserTodoStatsFn(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

// newTestClient creates a gqlgen test client with the mock service
func newTestClient(mockTodoSvc todo.TodoServiceInterface) *client.Client {
	resolver := NewResolver(nil, mockTodoSvc) // AuthService nil as not used in tests
	cfg := generated.Config{Resolvers: resolver}
	schema := generated.NewExecutableSchema(cfg)
	server := handler.NewDefaultServer(schema)
	return client.New(server)
}

// withAuthUserModifier returns a client.Option to add a mock user to the context
func withAuthUserModifier(userID int) client.Option {
	return func(bd *client.Request) {
		mockUser := &auth.User{
			ID: userID,
		}
		ctx := context.WithValue(bd.HTTP.Context(), middleware.UserContextKey, mockUser)
		bd.HTTP = bd.HTTP.WithContext(ctx)
	}
}

func TestQuery_Todos(t *testing.T) {
	mockSvc := &MockTodoService{
		GetUserTodosFn: func(ctx context.Context, userID int, filter todo.TodoFilter) (*todo.TodoListResponse, error) {
			assert.Equal(t, 1, userID)
			return &todo.TodoListResponse{
				Todos: []*todo.Todo{
					{ID: 1, Title: "Test Todo", Completed: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				},
				Total:   1,
				Limit:   10,
				Offset:  0,
				HasMore: false,
			}, nil
		},
	}

	c := newTestClient(mockSvc)

	var resp struct {
		Todos struct {
			Todos []struct {
				ID    string
				Title string
			}
			Total   int
			Limit   int
			Offset  int
			HasMore bool
		}
	}

	err := c.Post(
		`query { todos { todos { id title } total limit offset hasMore } }`,
		&resp,
		withAuthUserModifier(1),
	)
	require.NoError(t, err)
	assert.Len(t, resp.Todos.Todos, 1)
	assert.Equal(t, "1", resp.Todos.Todos[0].ID)
	assert.Equal(t, "Test Todo", resp.Todos.Todos[0].Title)
	assert.Equal(t, 1, resp.Todos.Total)
}

func TestQuery_Todos_WithFilter(t *testing.T) {
	mockSvc := &MockTodoService{
		GetUserTodosFn: func(ctx context.Context, userID int, filter todo.TodoFilter) (*todo.TodoListResponse, error) {
			assert.Equal(t, 1, userID)
			assert.True(t, *filter.Completed)
			return &todo.TodoListResponse{
				Todos: []*todo.Todo{
					{ID: 1, Title: "Completed Todo", Completed: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				},
				Total:   1,
				Limit:   10,
				Offset:  0,
				HasMore: false,
			}, nil
		},
	}

	c := newTestClient(mockSvc)

	var resp struct {
		Todos struct {
			Todos []struct {
				ID        string
				Title     string
				Completed bool
			}
		}
	}

	err := c.Post(
		`query { todos(filter: {completed: true}) { todos { id title completed } } }`,
		&resp,
		withAuthUserModifier(1),
	)
	require.NoError(t, err)
	assert.Len(t, resp.Todos.Todos, 1)
	assert.True(t, resp.Todos.Todos[0].Completed)
}

func TestQuery_Todos_Unauthorized(t *testing.T) {
	mockSvc := &MockTodoService{}

	c := newTestClient(mockSvc)

	var resp struct{ Todos interface{} }
	err := c.Post(
		`query { todos { todos { id } } }`,
		&resp,
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not authenticated")
}

func TestQuery_Todo(t *testing.T) {
	mockSvc := &MockTodoService{
		GetTodoFn: func(ctx context.Context, todoID, userID int) (*todo.Todo, error) {
			assert.Equal(t, 1, userID)
			assert.Equal(t, 1, todoID)
			return &todo.Todo{ID: 1, Title: "Single Todo", Completed: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
	}

	c := newTestClient(mockSvc)

	var resp struct {
		Todo struct {
			ID    string
			Title string
		}
	}

	err := c.Post(
		`query { todo(id: "1") { id title } }`,
		&resp,
		withAuthUserModifier(1),
	)
	require.NoError(t, err)
	assert.Equal(t, "1", resp.Todo.ID)
	assert.Equal(t, "Single Todo", resp.Todo.Title)
}

func TestQuery_Todo_NotFound(t *testing.T) {
	mockSvc := &MockTodoService{
		GetTodoFn: func(ctx context.Context, todoID, userID int) (*todo.Todo, error) {
			return nil, todo.ErrTodoNotFound
		},
	}

	c := newTestClient(mockSvc)

	var resp struct{ Todo interface{} }
	err := c.Post(
		`query { todo(id: "999") { id } }`,
		&resp,
		withAuthUserModifier(1),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "todo not found")
}

func TestMutation_CreateTodo(t *testing.T) {
	mockSvc := &MockTodoService{
		CreateTodoFn: func(ctx context.Context, userID int, input todo.CreateTodoInput) (*todo.Todo, error) {
			assert.Equal(t, 1, userID)
			assert.Equal(t, "New Todo", input.Title)
			assert.Equal(t, stringPtr("Desc"), input.Description)
			return &todo.Todo{ID: 1, Title: "New Todo", Description: stringPtr("Desc"), Completed: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
	}

	c := newTestClient(mockSvc)

	var resp struct {
		CreateTodo struct {
			ID          string
			Title       string
			Description string
		}
	}

	err := c.Post(
		`mutation { createTodo(input: {title: "New Todo", description: "Desc"}) { id title description } }`,
		&resp,
		withAuthUserModifier(1),
	)
	require.NoError(t, err)
	assert.Equal(t, "1", resp.CreateTodo.ID)
	assert.Equal(t, "New Todo", resp.CreateTodo.Title)
	assert.Equal(t, "Desc", resp.CreateTodo.Description)
}

func TestMutation_CreateTodo_InvalidInput(t *testing.T) {
	mockSvc := &MockTodoService{
		CreateTodoFn: func(ctx context.Context, userID int, input todo.CreateTodoInput) (*todo.Todo, error) {
			return nil, todo.ErrTodoTitleRequired
		},
	}

	c := newTestClient(mockSvc)

	var resp struct{ CreateTodo interface{} }
	err := c.Post(
		`mutation { createTodo(input: {title: ""}) { id } }`,
		&resp,
		withAuthUserModifier(1),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "todo title is required")
}

func TestMutation_UpdateTodo(t *testing.T) {
	mockSvc := &MockTodoService{
		UpdateTodoFn: func(ctx context.Context, todoID, userID int, input todo.UpdateTodoInput) (*todo.Todo, error) {
			assert.Equal(t, 1, userID)
			assert.Equal(t, 1, todoID)
			assert.Equal(t, stringPtr("Updated Title"), input.Title)
			return &todo.Todo{ID: 1, Title: "Updated Title", Completed: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
	}

	c := newTestClient(mockSvc)

	var resp struct {
		UpdateTodo struct {
			ID        string
			Title     string
			Completed bool
		}
	}

	err := c.Post(
		`mutation { updateTodo(id: "1", input: {title: "Updated Title", completed: true}) { id title completed } }`,
		&resp,
		withAuthUserModifier(1),
	)
	require.NoError(t, err)
	assert.Equal(t, "1", resp.UpdateTodo.ID)
	assert.Equal(t, "Updated Title", resp.UpdateTodo.Title)
	assert.True(t, resp.UpdateTodo.Completed)
}

func TestMutation_DeleteTodo(t *testing.T) {
	mockSvc := &MockTodoService{
		DeleteTodoFn: func(ctx context.Context, todoID, userID int) error {
			assert.Equal(t, 1, userID)
			assert.Equal(t, 1, todoID)
			return nil
		},
	}

	c := newTestClient(mockSvc)

	var resp struct {
		DeleteTodo bool
	}

	err := c.Post(
		`mutation { deleteTodo(id: "1") }`,
		&resp,
		withAuthUserModifier(1),
	)
	require.NoError(t, err)
	assert.True(t, resp.DeleteTodo)
}

func TestMutation_DeleteTodo_NotFound(t *testing.T) {
	mockSvc := &MockTodoService{
		DeleteTodoFn: func(ctx context.Context, todoID, userID int) error {
			return todo.ErrTodoNotFound
		},
	}

	c := newTestClient(mockSvc)

	var resp struct{ DeleteTodo bool }
	err := c.Post(
		`mutation { deleteTodo(id: "999") }`,
		&resp,
		withAuthUserModifier(1),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "todo not found")
}

func TestMutation_ToggleTodo(t *testing.T) {
	mockSvc := &MockTodoService{
		ToggleTodoCompleteFn: func(ctx context.Context, todoID, userID int) (*todo.Todo, error) {
			assert.Equal(t, 1, userID)
			assert.Equal(t, 1, todoID)
			return &todo.Todo{ID: 1, Title: "Toggled", Completed: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
		},
	}

	c := newTestClient(mockSvc)

	var resp struct {
		ToggleTodo struct {
			ID        string
			Completed bool
		}
	}

	err := c.Post(
		`mutation { toggleTodo(id: "1") { id completed } }`,
		&resp,
		withAuthUserModifier(1),
	)
	require.NoError(t, err)
	assert.True(t, resp.ToggleTodo.Completed)
}

func TestMutation_BatchUpdateTodos(t *testing.T) {
	mockSvc := &MockTodoService{
		BatchUpdateTodosFn: func(ctx context.Context, userID int, todoIDs []int, input todo.UpdateTodoInput) ([]*todo.Todo, error) {
			assert.Equal(t, 1, userID)
			assert.Equal(t, []int{1, 2}, todoIDs)
			assert.True(t, *input.Completed)
			return []*todo.Todo{
				{ID: 1, Completed: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
				{ID: 2, Completed: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
			}, nil
		},
	}

	c := newTestClient(mockSvc)

	var resp struct {
		BatchUpdateTodos []struct {
			ID        string
			Completed bool
		}
	}

	err := c.Post(
		`mutation { batchUpdateTodos(input: {todoIds: ["1", "2"], updates: {completed: true}}) { id completed } }`,
		&resp,
		withAuthUserModifier(1),
	)
	require.NoError(t, err)
	assert.Len(t, resp.BatchUpdateTodos, 2)
	assert.True(t, resp.BatchUpdateTodos[0].Completed)
	assert.True(t, resp.BatchUpdateTodos[1].Completed)
}

func TestMutation_BatchUpdateTodos_PartialFailure(t *testing.T) {
	mockSvc := &MockTodoService{
		BatchUpdateTodosFn: func(ctx context.Context, userID int, todoIDs []int, input todo.UpdateTodoInput) ([]*todo.Todo, error) {
			return []*todo.Todo{{ID: 1, Completed: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}}, errors.New("partial error")
		},
	}

	c := newTestClient(mockSvc)

	var resp struct{ BatchUpdateTodos []interface{} }
	err := c.Post(
		`mutation { batchUpdateTodos(input: {todoIds: ["1", "2"], updates: {completed: true}}) { id } }`,
		&resp,
		withAuthUserModifier(1),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "partial error")
}

func TestQuery_TodoStats(t *testing.T) {
	mockSvc := &MockTodoService{
		GetUserTodoStatsFn: func(ctx context.Context, userID int) (*todo.TodoStats, error) {
			assert.Equal(t, 1, userID)
			return &todo.TodoStats{Total: 5, Completed: 2, Pending: 3}, nil
		},
	}

	c := newTestClient(mockSvc)

	var resp struct {
		TodoStats struct {
			Total     int
			Completed int
			Pending   int
		}
	}

	err := c.Post(
		`query { todoStats { total completed pending } }`,
		&resp,
		withAuthUserModifier(1),
	)
	require.NoError(t, err)
	assert.Equal(t, 5, resp.TodoStats.Total)
	assert.Equal(t, 2, resp.TodoStats.Completed)
	assert.Equal(t, 3, resp.TodoStats.Pending)
}

func TestSubscription_TodoChanged(t *testing.T) {
	mockSvc := &MockTodoService{}

	c := newTestClient(mockSvc)

	sub := c.Websocket(
		`subscription { todoChanged { id } }`,
		withAuthUserModifier(1),
	)
	defer sub.Close()

	var resp struct{ TodoChanged struct{ ID string } }
	err := sub.Next(&resp)
	assert.Error(t, err) // Expect error for closed channel
}

func TestSubscription_TodoStatsChanged(t *testing.T) {
	mockSvc := &MockTodoService{}

	c := newTestClient(mockSvc)

	sub := c.Websocket(
		`subscription { todoStatsChanged { total } }`,
		withAuthUserModifier(1),
	)
	defer sub.Close()

	var resp struct{ TodoStatsChanged struct{ Total int } }
	err := sub.Next(&resp)
	assert.Error(t, err)
}

// Helper for string pointers
func stringPtr(s string) *string {
	return &s
}
