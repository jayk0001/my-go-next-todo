package todo

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jayk0001/my-go-next-todo/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestTodoCRUD_Integration(t *testing.T) {
	ctx := context.Background()

	// Start test DB container
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

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	// Run migrations (assume a func RunMigrations(pool))
	require.NoError(t, database.RunMigrations(pool)) // Implement this to create todos table

	// Create a test user (since todos has foreign key to users)
	_, err = pool.Exec(ctx, `
		INSERT INTO users (email, password_hash) VALUES ('test@example.com', 'test_hash')
	`)
	require.NoError(t, err)

	repo := NewTodoRepository(pool)
	service := NewTodoService(repo, NewValidatorService())

	// Test Create
	createInput := CreateTodoInput{Title: "Test", Description: stringPtr("Desc")}
	created, err := service.CreateTodo(ctx, 1, createInput) // userID=1 from insert above
	require.NoError(t, err)                                 // Use require to halt on error
	assert.Equal(t, "Test", created.Title)

	// Test Get
	got, err := service.GetTodo(ctx, created.ID, 1)
	assert.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	// Test Update
	updateInput := UpdateTodoInput{Completed: boolPtr(true)}
	updated, err := service.UpdateTodo(ctx, created.ID, 1, updateInput)
	assert.NoError(t, err)
	assert.True(t, updated.Completed)

	// Test Delete
	err = service.DeleteTodo(ctx, created.ID, 1)
	assert.NoError(t, err)

	// Verify deleted
	_, err = service.GetTodo(ctx, created.ID, 1)
	assert.ErrorIs(t, err, ErrTodoNotFound)
}
