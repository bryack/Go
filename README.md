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

### Graceful Shutdown

The server supports graceful shutdown to ensure in-flight requests complete before termination:

```bash
# Send SIGINT (Ctrl+C) or SIGTERM to initiate graceful shutdown
kill -SIGTERM <pid>

# Or press Ctrl+C in the terminal
```

**Shutdown Behavior**:
- Server stops accepting new connections immediately
- Existing requests are allowed to complete (default 30s timeout)
- Database connections are closed cleanly
- Press Ctrl+C again to force immediate shutdown

**Configuration**:
```yaml
server:
  shutdown_timeout: "30s"  # Maximum time to wait for requests to complete
```

**Exit Codes**:
- `0` - Clean shutdown (all requests completed)
- `1` - Error during shutdown or forced termination

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

## Logging Configuration

The application uses structured logging with support for JSON and text formats, multiple log levels, and integration with industry-standard log analysis tools.

### Configuration Options

Logging can be configured through:
1. Configuration file (`config.yaml`)
2. Environment variables (prefixed with `TASKMANAGER_`)
3. Command-line flags

**Configuration precedence**: flags > environment variables > config file > defaults

### Log Levels

- `debug` - Detailed information for debugging (includes all database queries)
- `info` - General informational messages (default)
- `warn` - Warning messages (authentication failures, validation errors)
- `error` - Error messages requiring attention

### Log Formats

- `json` - Structured JSON format for production and log aggregation tools
- `text` - Human-readable format for development

### Output Destinations

- `stderr` - Standard error (default, recommended for production)
- `stdout` - Standard output (use only for program output/data)
- File path - Write to a file (e.g., `/var/log/taskmanager/app.log`)

**Best Practice**: Use `stderr` for all application logs following Unix philosophy. This prevents log buffering issues, is container-friendly, and separates diagnostic messages from program output.

### Configuration File Example

Add logging configuration to your `config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  path: "./data/tasks.db"

jwt:
  secret: "your-secret-key-here-min-32-chars"
  expiration: "24h"

logging:
  level: "info"                    # debug, info, warn, error
  format: "json"                   # json, text
  output: "stderr"                 # stderr (recommended), stdout, or file path
  add_source: true                 # Include file:line for error logs
  service_name: "task-manager-api" # Service identifier
  environment: "production"        # development, staging, production
  
  # File rotation (only used when output is a file path)
  enable_rotation: true
  max_size: 100                    # Maximum size in MB before rotation
  max_age: 30                      # Maximum days to retain old logs
  max_backups: 5                   # Maximum number of old log files to keep
```

### Environment Variables

```bash
# Log level
export TASKMANAGER_LOGGING_LEVEL=debug

# Log format
export TASKMANAGER_LOGGING_FORMAT=json

# Output destination
export TASKMANAGER_LOGGING_OUTPUT=/var/log/taskmanager/app.log

# Service identification
export TASKMANAGER_LOGGING_SERVICE_NAME=task-manager-api
export TASKMANAGER_LOGGING_ENVIRONMENT=production

# Include source file and line in logs
export TASKMANAGER_LOGGING_ADD_SOURCE=true

# File rotation settings
export TASKMANAGER_LOGGING_ENABLE_ROTATION=true
export TASKMANAGER_LOGGING_MAX_SIZE=100
export TASKMANAGER_LOGGING_MAX_AGE=30
export TASKMANAGER_LOGGING_MAX_BACKUPS=5
```

### Command-Line Flags

```bash
# Start server with custom logging configuration
go run cmd/server/main.go \
  --log-level=debug \
  --log-format=text \
  --log-output=stdout \
  --log-add-source=true \
  --log-service-name=task-manager-api \
  --log-environment=development

# View current configuration
go run cmd/server/main.go --show-config
```

### Log Output Examples

#### JSON Format (Production)

```json
{
  "time": "2024-11-12T10:30:45.123Z",
  "level": "INFO",
  "msg": "HTTP request completed",
  "service": "task-manager-api",
  "environment": "production",
  "request_id": "req_1699785045123_a1b2c3d4e5f6g7h8",
  "user_id": 42,
  "method": "POST",
  "path": "/tasks",
  "status_code": 201,
  "duration_ms": 45
}
```

#### Text Format (Development)

```
time=2024-11-12T10:30:45.123Z level=INFO msg="HTTP request completed" service=task-manager-api environment=development request_id=req_1699785045123_a1b2c3d4e5f6g7h8 user_id=42 method=POST path=/tasks status_code=201 duration_ms=45
```

### Standard Log Fields

All log entries include these standard fields:

- `time` - ISO8601 timestamp
- `level` - Log level (DEBUG, INFO, WARN, ERROR)
- `msg` - Log message
- `service` - Service name from configuration
- `environment` - Environment from configuration

HTTP request logs include:

- `request_id` - Unique identifier for request correlation
- `user_id` - Authenticated user ID (when available)
- `method` - HTTP method (GET, POST, PUT, DELETE)
- `path` - Request path
- `status_code` - HTTP response status code
- `duration_ms` - Request duration in milliseconds

Database operation logs include:

- `operation` - Database operation type (SELECT, INSERT, UPDATE, DELETE)
- `task_id` - Task identifier (when applicable)
- `user_id` - User identifier

Authentication logs include:

- `email` - Masked email address (e.g., `u***r@example.com`)
- `operation` - Auth operation (register, login, validate)

### Integration with Log Analysis Tools

#### ELK Stack (Elasticsearch, Logstash, Kibana)

**Logstash Configuration** (`/etc/logstash/conf.d/taskmanager.conf`):

```ruby
input {
  file {
    path => "/var/log/taskmanager/app.log"
    codec => "json"
    type => "taskmanager"
  }
}

filter {
  # Parse JSON logs
  json {
    source => "message"
  }
  
  # Rename timestamp field for Elasticsearch
  mutate {
    rename => { "time" => "@timestamp" }
  }
  
  # Add tags for filtering
  mutate {
    add_tag => ["taskmanager", "%{environment}"]
  }
}

output {
  elasticsearch {
    hosts => ["localhost:9200"]
    index => "taskmanager-%{+YYYY.MM.dd}"
  }
}
```

**Kibana Queries**:

```
# Find all errors
level:ERROR

# Find requests by user
user_id:42

# Find slow requests (>1000ms)
duration_ms:>1000

# Trace a specific request
request_id:"req_1699785045123_a1b2c3d4e5f6g7h8"
```

#### Grafana Loki

**Promtail Configuration** (`/etc/promtail/config.yml`):

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: taskmanager
    static_configs:
      - targets:
          - localhost
        labels:
          job: taskmanager
          __path__: /var/log/taskmanager/*.log
    
    pipeline_stages:
      # Parse JSON logs
      - json:
          expressions:
            level: level
            service: service
            environment: environment
            request_id: request_id
            user_id: user_id
      
      # Extract labels for filtering
      - labels:
          level:
          service:
          environment:
```

**Loki Queries** (LogQL):

```
# All logs from the service
{service="task-manager-api"}

# Filter by log level
{service="task-manager-api", level="ERROR"}

# Search for specific text
{service="task-manager-api"} |= "database"

# Trace a specific request
{service="task-manager-api"} | json | request_id="req_1699785045123_a1b2c3d4e5f6g7h8"

# Count errors per minute
sum(rate({service="task-manager-api", level="ERROR"}[1m]))
```

#### Datadog

**Datadog Agent Configuration** (`/etc/datadog-agent/conf.d/taskmanager.d/conf.yaml`):

```yaml
logs:
  - type: file
    path: /var/log/taskmanager/app.log
    service: task-manager-api
    source: go
    sourcecategory: sourcecode
    tags:
      - env:production
      - team:backend
```

**Features**:
- Automatic JSON parsing
- APM integration with trace IDs
- Log-based metrics and alerting
- Correlation with infrastructure metrics

### Docker Logging

When running in Docker, configure logging output:

```bash
# Log to stderr (recommended - captured by Docker)
docker run -e TASKMANAGER_LOGGING_OUTPUT=stderr task-manager:latest

# Log to stdout (alternative)
docker run -e TASKMANAGER_LOGGING_OUTPUT=stdout task-manager:latest

# Log to a file with volume mount
docker run \
  -e TASKMANAGER_LOGGING_OUTPUT=/var/log/app.log \
  -v $(pwd)/logs:/var/log \
  task-manager:latest
```

**Docker Compose with Logging**:

```yaml
version: '3.8'

services:
  task-manager:
    build: .
    ports:
      - "8080:8080"
    environment:
      JWT_SECRET_KEY: "your-secure-secret-key-here"
      TASKMANAGER_LOGGING_LEVEL: "info"
      TASKMANAGER_LOGGING_FORMAT: "json"
      TASKMANAGER_LOGGING_OUTPUT: "stderr"
      TASKMANAGER_LOGGING_ENVIRONMENT: "production"
    volumes:
      - ./data:/data
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### Troubleshooting Logging Issues

#### No Logs Appearing

**Problem**: Server starts but no logs are visible.

**Solutions**:
1. Check log level - DEBUG logs won't appear if level is INFO
2. Verify output destination is correct
3. Check file permissions if logging to a file
4. Ensure log directory exists and is writable

```bash
# Test with debug level and stdout
go run cmd/server/main.go --log-level=debug --log-output=stdout
```

#### Log File Not Created

**Problem**: Application fails to create log file.

**Solutions**:
1. Verify directory exists or can be created
2. Check write permissions on the directory
3. Ensure sufficient disk space

```bash
# Create log directory manually
mkdir -p /var/log/taskmanager
chmod 755 /var/log/taskmanager

# Test file creation
touch /var/log/taskmanager/test.log
```

#### Logs Missing Request IDs

**Problem**: Request correlation is not working.

**Solution**: Ensure the logging middleware is properly configured in `main.go`. Request IDs are automatically generated by the middleware.

#### Sensitive Data in Logs

**Problem**: Passwords or tokens appearing in logs.

**Solution**: The application automatically masks sensitive data:
- Emails are masked: `user@example.com` → `u***r@example.com`
- Tokens are masked: `eyJhbGc...` → `eyJh****...`
- Passwords are never logged

If you find sensitive data in logs, please report it as a security issue.

#### High Log Volume

**Problem**: Too many logs in production.

**Solutions**:
1. Increase log level to `warn` or `error`
2. Enable log rotation to manage disk space
3. Use log sampling for high-traffic endpoints

```yaml
logging:
  level: "warn"  # Only warnings and errors
  enable_rotation: true
  max_size: 100
  max_age: 7  # Keep logs for 7 days
```

#### JSON Parsing Errors in Log Tools

**Problem**: Log aggregation tool can't parse JSON logs.

**Solutions**:
1. Verify `format: "json"` is set in configuration
2. Check for multi-line log entries (stack traces)
3. Ensure proper character encoding (UTF-8)

```bash
# Validate JSON format
tail -f /var/log/taskmanager/app.log | jq .
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
├── logger/           # Structured logging package
│   ├── logger.go     # Logger factory and configuration
│   ├── config.go     # Configuration types and validation
│   ├── middleware.go # HTTP logging middleware
│   ├── context.go    # Context helpers for request correlation
│   └── fields.go     # Standard field names and masking
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
