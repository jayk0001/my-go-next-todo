# Todo App Backend

## Description
This project is a backend API for a Todo application built in Go, designed as a learning exercise to explore Go language fundamentals, project structure, and testing practices. It includes user authentication, Todo CRUD operations, and GraphQL integration. The focus was on implementing best practices for directory organization, dependency injection, and comprehensive testing (unit, integration, E2E, and mocks). The codebase demonstrates how to build a scalable API with Gin for routing, pgx for Postgres interactions, gqlgen for GraphQL, and JWT for authentication.

## Features
- User authentication with register, login, and token refresh.
- Todo management with CRUD operations, filtering, pagination, and batch updates.
- GraphQL API for queries, mutations, and subscriptions (placeholders for real-time).
- Ownership checks (users can only access their own todos).
- Health checks and CORS middleware.
- Graceful server shutdown.

## Tech Stack
- **Language**: Go 1.23+
- **Web Framework**: Gin
- **Database**: Postgres (pgx driver)
- **GraphQL**: gqlgen
- **Authentication**: JWT (golang-jwt)
- **Testing**: Go testing package, testify for assertions, testcontainers for DB integration, sqlmock for DB mocks, httptest for E2E.
- **Other**: godotenv for env vars, joho/godotenv, jackc/pgx, DATA-DOG/go-sqlmock.

## Directory Structure
The project follows a standard Go layout with internal packages for encapsulation. Key directories:

- `/cmd/server`: Entry point (main.go) for server startup and shutdown.
- `/internal/auth`: Authentication logic (service, repo, JWT, password hashing, tests).
- `/internal/config`: Configuration loading from env.
- `/internal/database`: DB connection and migrations.
- `/internal/graphql`: GraphQL schema, resolvers, generated code.
- `/internal/middleware`: Gin middleware (auth, CORS).
- `/internal/server`: Server setup (Gin router, routes).
- `/internal/todo`: Todo domain (models, repo, service, tests).
- `/migrations`: SQL migration files (manual or tool-based).
- `/pkg`: Reusable packages (if needed; empty here).
- `/api`: OpenAPI specs (if added).
- `/frontend`: Frontend code (Next.js, but not implemented).

This structure separates concerns: cmd for executables, internal for private code, tests colocated with code.

## Setup/Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/jayk0001/todo-app-backend.git
   cd todo-app-backend
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set up environment variables (copy .env.example to .env):
   - DB connection (e.g., DATABASE_URL for Neon or local Postgres).
   - JWT_SECRET.
   - SERVER_PORT=8080.

4. Run migrations (manual or add tool like goose):
   - Start DB, then run relevant SQL from /migrations.

5. Build and run:
   ```bash
   go run cmd/server/main.go
   ```
   - API at http://localhost:8080/graphql/query.
   - Playground at http://localhost:8080/graphql (dev mode).

## Running the Project
- **Dev Mode**: `go run cmd/server/main.go` (Gin debug).
- **Prod Mode**: Set ENVIRONMENT="production" in .env, build with `go build cmd/server/main.go`.
- **Health Check**: GET /health.

## Testing
Tests cover unit, integration, resolver, and E2E levels, with mocks for isolation.

- **Run All Tests**:
  ```bash
  make test-unit  # Or go test ./internal/...
  ```

- **Unit Tests**: Package-specific (e.g., auth, todo). Use mocks for deps (e.g., MockTodoRepository).
- **Integration Tests**: Real DB with testcontainers (e.g., todo/integration_test.go for CRUD flow).
- **GraphQL Resolver Tests**: gqlgen client with mocked services (resolver/todo_resolver_test.go).
- **E2E Tests**: httptest server with test DB (server/server_test.go for full API flow with auth).
- **Mock DB Tests**: sqlmock for repo queries (todo/repository_test.go).
- **Coverage**: `go test -cover ./...`.

Learned: Table-driven tests, require/assert with testify, testcontainers for DB, httptest for server, sqlmock for DB mocks, context for auth in resolvers.

## Lessons Learned from Development
- **Directory Structure**: Internal for private code, cmd for executables, colocated tests.
- **Dependency Injection**: Services/repos injected via New funcs for testability.
- **Testing Practices**:
  - Unit: Isolate layers with mocks (gomock or manual).
  - Integration: Testcontainers for Postgres, real flows without mocks.
  - E2E: httptest.NewServer for full API, with auth/token simulation.
  - Resolver: gqlgen client with mocked services/context.
  - Mock DB: sqlmock for query expectations/errors.
- **GraphQL Integration**: gqlgen for schema/resolvers, codegen for types.
- **Auth & Security**: JWT claims as int, middleware for token validation/context user.
- **Error Handling**: Custom errors, wrapped fmt.Errorf for chains.
- **Commits**: Conventional Commits (feat/test/fix) with scopes (auth/todo/graphql).
- **Tools**: Testcontainers/sqlmock for reliable tests, godotenv for env.

## License
MIT License. See LICENSE file.