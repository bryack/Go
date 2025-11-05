# Task Manager API

A Go-based task management application with JWT authentication, SQLite storage, and RESTful API.

## Features

- RESTful API for task management (CRUD operations)
- JWT-based authentication and authorization
- SQLite database for persistent storage
- Docker containerization for easy deployment
- CI/CD pipeline with GitHub Actions

## Quick Start

### Prerequisites

- Go 1.24 or higher
- Docker (for containerized deployment)
- Git

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

## Deployment

### Docker Deployment

#### Building the Docker Image

Build the Docker image locally:

```bash
docker build -t task-manager:latest .
```

The build process uses a multi-stage Dockerfile that:
- Compiles the Go application with CGO support for SQLite
- Creates a minimal Alpine-based runtime image
- Includes all necessary dependencies

#### Running the Server Container

Run the server with persistent database storage:

```bash
docker run -d \
  --name task-manager \
  -p 8080:8080 \
  -e JWT_SECRET_KEY="your-secure-secret-key-here" \
  -v $(pwd)/data:/data \
  task-manager:latest
```

**Important:** Always set a secure `JWT_SECRET_KEY` in production. The default value in the Dockerfile is for development only.

#### Running the CLI Container (Interactive)

> **Note:** The CLI is currently disabled and requires refactoring to work with the new storage interface. This section will be updated once the CLI is re-enabled.

```bash
# This will be available after CLI refactoring
docker run -it \
  --rm \
  -v $(pwd)/data:/data \
  task-manager:latest /app/cli
```

### Environment Variables

The application supports the following environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `JWT_SECRET_KEY` | Secret key for JWT token signing and verification | `CHANGE_ME_IN_PRODUCTION` | **Yes** |
| `TASK_DB_PATH` | Path to SQLite database file | `/data/tasks.db` | No |
| `PORT` | HTTP server listening port (future use) | `8080` | No |

**Security Warning:** Never use the default `JWT_SECRET_KEY` in production. Always override it with a strong, randomly generated secret.

### Volume Mounts

The application uses `/data` as the persistent storage directory for the SQLite database:

```bash
# Mount a local directory for database persistence
-v /path/on/host:/data

# Example with current directory
-v $(pwd)/data:/data
```

Without a volume mount, all data will be lost when the container is removed.

### Docker Compose (Optional)

For easier local development, you can use Docker Compose. Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  task-manager:
    build: .
    ports:
      - "8080:8080"
    environment:
      # ⚠️  Change this to a secure secret in production!
      JWT_SECRET_KEY: "your-secure-secret-key-here"
      TASK_DB_PATH: "/data/tasks.db"
      PORT: "8080"
    volumes:
      # Persist database across container restarts
      - ./data:/data
    restart: unless-stopped
```

Run with Docker Compose:

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

## API Endpoints

### Authentication

#### Register a New User
```bash
POST /register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}

Response: 201 Created
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "email": "user@example.com"
}
```

#### Login
```bash
POST /login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword"
}

Response: 200 OK
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "email": "user@example.com"
}
```

### Tasks (Requires Authentication)

All task endpoints require a valid JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

#### Get All Tasks
```bash
GET /tasks
Authorization: Bearer <token>

Response: 200 OK
[
  {
    "id": 1,
    "description": "Complete project",
    "done": false
  }
]
```

#### Create a Task
```bash
POST /tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "description": "New task description"
}

Response: 201 Created
{
  "id": 1,
  "description": "New task description",
  "done": false
}
```

#### Get a Specific Task
```bash
GET /tasks/{id}
Authorization: Bearer <token>

Response: 200 OK
{
  "id": 1,
  "description": "Complete project",
  "done": false
}
```

#### Update a Task
```bash
PUT /tasks/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "description": "Updated description",
  "done": true
}

Response: 200 OK
{
  "id": 1,
  "description": "Updated description",
  "done": true
}
```

#### Delete a Task
```bash
DELETE /tasks/{id}
Authorization: Bearer <token>

Response: 204 No Content
```

### Health Check
```bash
GET /health

Response: 200 OK
{
  "status": "healthy",
  "timestamp": "2025-11-05T10:30:00Z",
  "service": "task-manager-api"
}
```

## CI/CD Pipeline

### GitHub Actions Workflow

The project includes an automated CI/CD pipeline that runs on every push to the `main` branch.

#### Workflow Steps

1. **Test Job**
   - Checks out the code
   - Sets up Go 1.24
   - Caches Go modules for faster builds
   - Runs all tests with `go test ./...`
   - Fails the pipeline if any tests fail

2. **Build Job** (runs only if tests pass)
   - Checks out the code
   - Sets up Docker Buildx
   - Builds the Docker image
   - Tags the image with `latest` and the commit SHA

#### Configuring GitHub Secrets

If you want to push images to a container registry (Docker Hub, GitHub Container Registry, etc.), you'll need to configure the following secrets in your GitHub repository:

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Add the following secrets:

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `DOCKER_USERNAME` | Your Docker registry username | `myusername` |
| `DOCKER_PASSWORD` | Your Docker registry password or access token | `dckr_pat_...` |

Then update `.github/workflows/deploy.yml` to include registry login and push steps:

```yaml
- name: Login to Docker Hub
  uses: docker/login-action@v3
  with:
    username: ${{ secrets.DOCKER_USERNAME }}
    password: ${{ secrets.DOCKER_PASSWORD }}

- name: Build and push Docker image
  uses: docker/build-push-action@v5
  with:
    context: .
    push: true
    tags: |
      ${{ secrets.DOCKER_USERNAME }}/task-manager:latest
      ${{ secrets.DOCKER_USERNAME }}/task-manager:${{ github.sha }}
```

## Troubleshooting

### Common Deployment Issues

#### 1. Container Fails to Start - JWT_SECRET_KEY Error

**Error:**
```
JWT_SECRET_KEY environment variable is required
```

**Solution:**
Always provide the `JWT_SECRET_KEY` environment variable when running the container:
```bash
docker run -e JWT_SECRET_KEY="your-secret-key" task-manager:latest
```

#### 2. Database Permission Errors

**Error:**
```
Failed to initialize database storage: unable to open database file
```

**Solution:**
Ensure the `/data` directory has proper permissions and is mounted correctly:
```bash
# Create the data directory on the host
mkdir -p ./data
chmod 755 ./data

# Run with volume mount
docker run -v $(pwd)/data:/data task-manager:latest
```

#### 3. Port Already in Use

**Error:**
```
bind: address already in use
```

**Solution:**
Either stop the process using port 8080 or map to a different host port:
```bash
# Map to port 3000 on the host
docker run -p 3000:8080 task-manager:latest
```

#### 4. Data Loss After Container Restart

**Problem:**
Tasks disappear after restarting the container.

**Solution:**
Always use a volume mount to persist the database:
```bash
docker run -v $(pwd)/data:/data task-manager:latest
```

Without the volume mount, the database is stored inside the container and is lost when the container is removed.

#### 5. Authentication Token Expired

**Error:**
```
401 Unauthorized
```

**Solution:**
JWT tokens expire after 24 hours. Login again to get a new token:
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password"}'
```

#### 6. Docker Build Fails - CGO Errors

**Error:**
```
gcc: not found
```

**Solution:**
This shouldn't happen with the provided Dockerfile, but if you encounter CGO-related errors, ensure the builder stage includes:
```dockerfile
RUN apk add --no-cache gcc musl-dev sqlite-dev
ENV CGO_ENABLED=1
```

#### 7. GitHub Actions Build Fails

**Common causes:**
- Tests are failing (check test job logs)
- Dockerfile syntax errors (check build job logs)
- Missing dependencies in go.mod

**Solution:**
1. Run tests locally: `go test ./...`
2. Build Docker image locally: `docker build -t task-manager .`
3. Fix any errors before pushing to main

#### 8. Cannot Connect to Database in Container

**Error:**
```
Failed to initialize database storage
```

**Solution:**
Ensure the `TASK_DB_PATH` points to a writable location inside the container:
```bash
docker run \
  -e TASK_DB_PATH="/data/tasks.db" \
  -v $(pwd)/data:/data \
  task-manager:latest
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./task
```

### Project Structure

```
.
├── auth/              # JWT authentication and middleware
├── cmd/
│   ├── cli/          # CLI application (currently disabled)
│   └── server/       # HTTP server application
├── internal/
│   └── handlers/     # HTTP request handlers and helpers
├── storage/          # Database storage layer
├── task/             # Task business logic
├── validation/       # Input validation utilities
├── Dockerfile        # Multi-stage Docker build
└── .github/
    └── workflows/    # CI/CD pipeline configuration
```

## License

[Add your license information here]

## Contributing

[Add contribution guidelines here]
