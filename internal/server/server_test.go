package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jayk0001/my-go-next-todo/internal/config"
	"github.com/jayk0001/my-go-next-todo/internal/database"
	"github.com/jayk0001/my-go-next-todo/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestE2E_TodoOperations tests the full Todo CRUD flow with authentication
func TestE2E_TodoOperations(t *testing.T) {
	ctx := context.Background()

	// Start test Postgres container
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("pass"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { pgContainer.Terminate(ctx) })

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Mock config for test
	testCfg := &config.Config{
		App: config.AppConfig{
			Environment: "test",
		},
		Database: config.DatabaseConfig{
			DatabaseURL: connStr,
		},
		JWT: config.JWTConfig{
			Secret:      "test-secret",
			ExpiryHours: time.Hour * 1, // Fix: 1 hour
		},
		Server: config.ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
	}

	// Initialize DB
	db, err := database.New(database.Config{DatabaseURL: connStr})
	require.NoError(t, err)
	pool := db.Pool
	t.Cleanup(func() {
		_, err := pool.Exec(ctx, "TRUNCATE users, todos RESTART IDENTITY CASCADE")
		assert.NoError(t, err)
		db.Close() // Close last
	})

	// Run migrations
	require.NoError(t, database.RunMigrations(pool))

	// Create Server instance (now initializes all services internally)
	srv := server.New(testCfg, db)

	// Use the router for httptest
	ginRouter := srv.Router()

	// Start test HTTP server
	testServer := httptest.NewServer(ginRouter)
	t.Cleanup(testServer.Close)

	// HTTP client for requests
	httpClient := testServer.Client()

	// Step 0: Setup test user by registering via GraphQL
	registerMutation := `mutation { register(input: {email: "test@example.com", password: "password123"}) { token user { id email } } }`
	registerPayload := map[string]string{"query": registerMutation}
	registerBody, _ := json.Marshal(registerPayload)
	registerReq, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql/query", bytes.NewBuffer(registerBody))
	registerReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(registerReq)
	require.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	t.Logf("Register response body: %s", string(bodyBytes)) // Debug log
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var registerResult struct {
		Data struct {
			Register struct {
				Token string `json:"token"`
				User  struct {
					ID    string `json:"id"`
					Email string `json:"email"`
				} `json:"user"`
			} `json:"register"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.Unmarshal(bodyBytes, &registerResult)
	require.NoError(t, err)
	if len(registerResult.Errors) > 0 {
		t.Fatalf("GraphQL errors: %v", registerResult.Errors)
	}
	token := registerResult.Data.Register.Token
	require.NotEmpty(t, token)

	// Helper to make GraphQL request with token
	makeGQLRequest := func(query string) *http.Response {
		payload := map[string]string{"query": query}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql/query", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		return resp
	}

	// Step 1: Create Todo
	createQuery := `mutation { createTodo(input: {title: "E2E Todo", description: "Test Desc"}) { id title description } }`
	resp = makeGQLRequest(createQuery)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, _ = io.ReadAll(resp.Body)
	t.Logf("Create response body: %s", string(bodyBytes))
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset for decode

	var createResult struct {
		Data struct {
			CreateTodo struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"createTodo"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&createResult)
	require.NoError(t, err)
	if len(createResult.Errors) > 0 {
		t.Fatalf("GraphQL create errors: %v", createResult.Errors)
	}
	todoID := createResult.Data.CreateTodo.ID
	assert.Equal(t, "E2E Todo", createResult.Data.CreateTodo.Title)
	assert.Equal(t, "Test Desc", createResult.Data.CreateTodo.Description)

	// Step 2: Get Todo by ID
	getQuery := `query { todo(id: "` + todoID + `") { id title description completed } }`
	resp = makeGQLRequest(getQuery)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var getResult struct {
		Data struct {
			Todo struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Completed   bool   `json:"completed"`
			} `json:"todo"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&getResult)
	require.NoError(t, err)
	if len(getResult.Errors) > 0 {
		t.Fatalf("GraphQL get errors: %v", getResult.Errors)
	}
	assert.Equal(t, todoID, getResult.Data.Todo.ID)
	assert.False(t, getResult.Data.Todo.Completed)

	// Step 3: List Todos
	listQuery := `query { todos { todos { id title } total limit offset hasMore } }`
	resp = makeGQLRequest(listQuery)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	bodyBytes, _ = io.ReadAll(resp.Body)
	t.Logf("List response body: %s", string(bodyBytes))
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reset for decode

	var listResult struct {
		Data struct {
			Todos struct {
				Todos []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"todos"`
				Total int `json:"total"`
			} `json:"todos"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&listResult)
	require.NoError(t, err)
	if len(listResult.Errors) > 0 {
		t.Fatalf("GraphQL list errors: %v", listResult.Errors)
	}
	require.Len(t, listResult.Data.Todos.Todos, 1)
	assert.Equal(t, todoID, listResult.Data.Todos.Todos[0].ID)
	assert.Equal(t, 1, listResult.Data.Todos.Total)

	// Step 4: Update Todo
	updateQuery := `mutation { updateTodo(id: "` + todoID + `", input: {title: "Updated E2E Todo", completed: true}) { id title completed } }`
	resp = makeGQLRequest(updateQuery)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updateResult struct {
		Data struct {
			UpdateTodo struct {
				ID        string `json:"id"`
				Title     string `json:"title"`
				Completed bool   `json:"completed"`
			} `json:"updateTodo"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&updateResult)
	require.NoError(t, err)
	if len(updateResult.Errors) > 0 {
		t.Fatalf("GraphQL update errors: %v", updateResult.Errors)
	}
	assert.Equal(t, "Updated E2E Todo", updateResult.Data.UpdateTodo.Title)
	assert.True(t, updateResult.Data.UpdateTodo.Completed)

	// Step 5: Toggle Todo
	toggleQuery := `mutation { toggleTodo(id: "` + todoID + `") { id completed } }`
	resp = makeGQLRequest(toggleQuery)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var toggleResult struct {
		Data struct {
			ToggleTodo struct {
				ID        string `json:"id"`
				Completed bool   `json:"completed"`
			} `json:"toggleTodo"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&toggleResult)
	require.NoError(t, err)
	if len(toggleResult.Errors) > 0 {
		t.Fatalf("GraphQL toggle errors: %v", toggleResult.Errors)
	}
	assert.False(t, toggleResult.Data.ToggleTodo.Completed) // Toggled back to false

	// Step 6: Batch Update (create another todo first)
	createQuery2 := `mutation { createTodo(input: {title: "E2E Todo 2"}) { id } }`
	resp = makeGQLRequest(createQuery2)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var createResult2 struct {
		Data struct {
			CreateTodo struct{ ID string } `json:"createTodo"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&createResult2)
	require.NoError(t, err)
	if len(createResult2.Errors) > 0 {
		t.Fatalf("GraphQL create2 errors: %v", createResult2.Errors)
	}
	todoID2 := createResult2.Data.CreateTodo.ID

	batchQuery := `mutation { batchUpdateTodos(input: {todoIds: ["` + todoID + `", "` + todoID2 + `"], updates: {completed: true}}) { id completed } }`
	resp = makeGQLRequest(batchQuery)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var batchResult struct {
		Data struct {
			BatchUpdateTodos []struct {
				ID        string `json:"id"`
				Completed bool   `json:"completed"`
			} `json:"batchUpdateTodos"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&batchResult)
	require.NoError(t, err)
	if len(batchResult.Errors) > 0 {
		t.Fatalf("GraphQL batch errors: %v", batchResult.Errors)
	}
	assert.Len(t, batchResult.Data.BatchUpdateTodos, 2)
	assert.True(t, batchResult.Data.BatchUpdateTodos[0].Completed)
	assert.True(t, batchResult.Data.BatchUpdateTodos[1].Completed)

	// Step 7: Get Stats
	statsQuery := `query { todoStats { total completed pending } }`
	resp = makeGQLRequest(statsQuery)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var statsResult struct {
		Data struct {
			TodoStats struct {
				Total     int `json:"total"`
				Completed int `json:"completed"`
				Pending   int `json:"pending"`
			} `json:"todoStats"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&statsResult)
	require.NoError(t, err)
	if len(statsResult.Errors) > 0 {
		t.Fatalf("GraphQL stats errors: %v", statsResult.Errors)
	}
	assert.Equal(t, 2, statsResult.Data.TodoStats.Total)
	assert.Equal(t, 2, statsResult.Data.TodoStats.Completed)
	assert.Equal(t, 0, statsResult.Data.TodoStats.Pending)

	// Step 8: Delete Todo
	deleteQuery := `mutation { deleteTodo(id: "` + todoID + `") }`
	resp = makeGQLRequest(deleteQuery)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var deleteResult struct {
		Data struct {
			DeleteTodo bool `json:"deleteTodo"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&deleteResult)
	require.NoError(t, err)
	if len(deleteResult.Errors) > 0 {
		t.Fatalf("GraphQL delete errors: %v", deleteResult.Errors)
	}
	assert.True(t, deleteResult.Data.DeleteTodo)

	// Verify delete: Get should fail
	getAfterDelete := `query { todo(id: "` + todoID + `") { id } }`
	resp = makeGQLRequest(getAfterDelete)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // GraphQL returns 200 even on errors

	var getErrorResult struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&getErrorResult)
	require.NoError(t, err)
	assert.NotEmpty(t, getErrorResult.Errors)
	assert.Contains(t, getErrorResult.Errors[0].Message, "not found")

	// Unauthorized test example: List without token
	listNoAuthPayload := map[string]string{"query": listQuery}
	listNoAuthBody, _ := json.Marshal(listNoAuthPayload)
	listNoAuthReq, _ := http.NewRequest(http.MethodPost, testServer.URL+"/graphql/query", bytes.NewBuffer(listNoAuthBody))
	listNoAuthReq.Header.Set("Content-Type", "application/json")
	// No Authorization header

	resp, err = httpClient.Do(listNoAuthReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode) // 200 with error for no token
	var unauthResult struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	err = json.NewDecoder(resp.Body).Decode(&unauthResult)
	require.NoError(t, err)
	assert.NotEmpty(t, unauthResult.Errors)
	assert.Contains(t, unauthResult.Errors[0].Message, "user not authenticated")
}
