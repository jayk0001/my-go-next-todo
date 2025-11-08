# My Go Next TODO

A full-stack TODO application built with Go (Gin + GraphQL) backend and Next.js frontend.

## 🏗️ Architecture

- **Backend**: Go + Gin + GraphQL + PostgreSQL
- **Frontend**: Next.js + TypeScript + Apollo Client
- **Database**: PostgreSQL with PGX driver
- **Authentication**: JWT tokens

## 📁 Project Structure

```
my-go-next-todo/
├── cmd/
│   └── server/              # Application entrypoint
├── internal/                # Private application code
│   ├── auth/               # Authentication logic
│   ├── todo/               # TODO business logic
│   ├── database/           # Database connection management
│   ├── config/             # Configuration management
│   └── server/             # HTTP server setup
├── pkg/                    # Public library code
├── api/                    # GraphQL schemas
├── migrations/             # Database migrations
├── frontend/               # Next.js application
└── docs/                   # Documentation
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 13+
- Node.js 18+
- Air (for hot reloading)
- golang-migrate

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/my-go-next-todo.git
   cd my-go-next-todo
   ```

2. **Install Go dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start PostgreSQL** (Docker)
   ```bash
   docker run --name todo-postgres \
     -e POSTGRES_DB=todoapp \
     -e POSTGRES_USER=postgres \
     -e POSTGRES_PASSWORD=password \
     -p 5432:5432 \
     -d postgres:15-alpine
   ```

5. **Run database migrations**
   ```bash
   make db-up
   ```

6. **Start the development server**
   ```bash
   make dev
   ```

The API will be available at `http://localhost:8080`

### Available Commands

```bash
# Development
make dev              # Start development server with hot reload
make build           # Build production binary
make test            # Run tests
make test-cover      # Run tests with coverage

# Database
make db-up           # Run migrations
make db-down         # Rollback migrations
make db-reset        # Reset database
make db-create-migration NAME=migration_name

# Code Quality
make fmt             # Format code
make lint            # Run linter
make vet             # Run go vet

# Help
make help            # Show all available commands
```

## 🏥 Health Checks

- `GET /health` - Overall health status
- `GET /health/ready` - Readiness probe
- `GET /health/live` - Liveness probe

## 📝 API Documentation

### Authentication Endpoints
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh token

### TODO Endpoints
- `GET /api/v1/todos` - List user's todos
- `POST /api/v1/todos` - Create new todo
- `GET /api/v1/todos/:id` - Get specific todo
- `PUT /api/v1/todos/:id` - Update todo
- `DELETE /api/v1/todos/:id` - Delete todo

### GraphQL Endpoint
- `POST /graphql` - GraphQL endpoint
- `GET /graphql` - GraphQL playground (development only)

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-cover

# Run specific package tests
go test ./internal/auth -v

# Run benchmarks
make bench
```

## 🐳 Docker

```bash
# Build Docker image
make docker-build

# Run with Docker Compose
make docker-compose-up
```

## 📊 Database Schema

### Users Table
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### Todos Table
```sql
CREATE TABLE todos (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

## 🤝 Contributing

1. Create a feature branch: `git checkout -b feature/amazing-feature`
2. Make your changes and commit: `git commit -m 'Add amazing feature'`
3. Push to the branch: `git push origin feature/amazing-feature`
4. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔧 Development Notes

### Code Style
- Follow Go conventions and best practices
- Use `gofmt` for code formatting
- Run `golangci-lint` for static analysis
- Write tests for all public functions

### Database Migrations
- Always create both up and down migrations
- Test migrations locally before committing
- Use descriptive names for migration files

### Commit Messages
Follow [Conventional Commits](https://www.conventionalcommits.org/):
- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation changes
- `refactor:` code refactoring
- `test:` adding tests
- `chore:` maintenance tasks