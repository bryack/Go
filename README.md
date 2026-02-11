# Task Manager API

A Go-based task management application with JWT authentication, SQLite storage, and RESTful API.

## Features

- RESTful API for task management (CRUD operations)
- JWT-based authentication and authorization
- SQLite database for persistent storage
- Docker containerization for easy deployment
- CI/CD pipeline with GitHub Actions
- Structured logging with JSON output
- Graceful shutdown support

## Quick Start

### Prerequisites

- Go 1.24 or higher
- Docker (for containerized deployment)

### Local Development

1. Clone the repository:
```bash
git clone <repository-url>
cd <repository-name>
```

2. Set required environment variables:
```bash
export JWT_SECRET_KEY="your-secure-secret-key-here"
export TASK_DB_PATH="./tasks.db"  # Optional, defaults to ./tasks.db
```

3. Run the server:
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

### Docker Quick Start

```bash
# Build and run with Docker
docker build -t task-manager:latest .
docker run -d \
  --name task-manager \
  -p 8080:8080 \
  -e JWT_SECRET_KEY="your-secure-secret-key-here" \
  -v $(pwd)/data:/data \
  task-manager:latest
```

## API Examples

### Authentication

Register a new user:
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"securepassword"}'
```

Login:
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"securepassword"}'
```

### Task Operations

Create a task (requires JWT token):
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"description":"New task"}'
```

Get all tasks:
```bash
curl -X GET http://localhost:8080/tasks \
  -H "Authorization: Bearer <token>"
```

Update a task:
```bash
curl -X PUT http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"description":"Updated task","done":true}'
```

### Health Check
```bash
curl http://localhost:8080/health
```

## Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Static analysis
go vet ./...

# Tidy dependencies
go mod tidy
```

### Project Structure
```
.
├── auth/              # JWT authentication and middleware
├── cmd/server/        # HTTP server application
├── internal/handlers/ # HTTP request handlers
├── storage/           # Database storage layer
├── task/              # Task business logic
├── validation/        # Input validation utilities
└── logger/            # Structured logging package
```

## Environment Variables

### Required
- `JWT_SECRET_KEY` - Secret for JWT token signing (minimum 32 characters)

### Optional
- `TASK_DB_PATH` - Database file path (default: `./tasks.db`)
- `PORT` - HTTP server port (default: `8080`)
- `LOG_LEVEL` - Logging level (default: `info`)
- `LOG_FORMAT` - Log format: `json` or `text` (default: `json`)

## Documentation

- **[Deployment Guide](docs/DEPLOYMENT.md)** - Docker, production setup, troubleshooting
- **[API Reference](docs/API.md)** - Complete endpoint documentation
- **[Development Guide](docs/DEVELOPMENT.md)** - Contributing, testing, CI/CD
- **[Logging Configuration](docs/LOGGING.md)** - Structured logging setup and integration

For operational commands and workflows, see the main [AGENTS.md](AGENTS.md) file.

## License

[Add your license information here]

## Contributing

See [Development Guide](docs/DEVELOPMENT.md) for contribution guidelines and setup instructions.
