# Task Manager API & CLI

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <h3 align="center">Go Task Manager</h3>

  <p align="center">
    A robust, enterprise-ready task management system featuring Clean Architecture, REST API, and a companion CLI tool.
    <br />
    <a href="https://github.com/bryack/Go/tree/main/docs"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/bryack/Go/issues">Report Bug</a>
    ·
    <a href="https://github.com/bryack/Go/issues">Request Feature</a>
  </p>
</div>

---

## About The Project

This project is a comprehensive task management system built with **Go 1.24**. It serves as a personal learning laboratory where I've implemented various modern software engineering patterns and technologies, even when not strictly required, to master their practical application.

The core focus of this project is to demonstrate:
*   **Clean Architecture:** Strict separation of concerns between Domain, Application, and Infrastructure layers.
*   **REST API:** Full-featured RESTful JSON API with JWT authentication.
*   **Test-Driven Development (TDD):** High test coverage with unit, integration, and behavioral tests.
*   **Developer Experience:** A fully-featured interactive CLI for seamless interaction.

### Built With

*   ![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
*   ![SQLite](https://img.shields.io/badge/SQLite-07405E?style=for-the-badge&logo=sqlite&logoColor=white)
*   ![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
*   ![JWT](https://img.shields.io/badge/JWT-black?style=for-the-badge&logo=JSON%20web%20tokens)
*   ![Viper](https://img.shields.io/badge/Viper-FF6B6B?style=for-the-badge&logo=go&logoColor=white)
*   ![testify](https://img.shields.io/badge/testify-00ADD8?style=for-the-badge&logo=go&logoColor=white)

---

## Features

- **REST API:** Clean RESTful JSON API with JWT authentication for web clients.
- **Interactive CLI:** A powerful terminal client with built-in authentication, task processing, and command suggestions.
- **JWT Authentication:** Secure access control for all task-related operations with bcrypt password hashing.
- **Persistent Storage:** SQLite backend using a CGO-free driver (`modernc.org/sqlite`) for maximum portability.
- **Advanced Logging:** Structured JSON/Text logging with automatic log rotation and compression via Lumberjack.
- **Containerized:** Ready-to-deploy Docker and Docker Compose configurations (including Loki for log aggregation).
- **Observability:** Prepared for integration with Loki and Grafana.
- **Configuration:** Flexible configuration via environment variables, YAML/JSON files, or command-line flags.

---

## Getting Started

### Prerequisites

*   Go 1.24+
*   Docker (optional, for containerized deployment)

### Installation

1.  Clone the repository:
    ```sh
    git clone https://github.com/bryack/Go.git
    cd Go
    ```
2.  Install dependencies:
    ```sh
    go mod tidy
    ```
3.  Configure environment:
    ```sh
    export TASKMANAGER_JWT_SECRET="your-32-character-secret-key"
    ```

---

## Usage

### Running the Server

Start the REST API server:
```sh
TASKMANAGER_JWT_SECRET="your-32-char-min-secret" go run ./cmd/server
```

**Alternative methods:**
```sh
# With configuration file
go run ./cmd/server --config=config.yaml

# Show current configuration
go run ./cmd/server --show-config

# With custom port
TASKMANAGER_SERVER_PORT=3000 go run ./cmd/server
```

### Using the CLI

The CLI provides an interactive experience. Run it and follow the prompts:
```sh
go run ./cmd/cli
```

**CLI Commands:**
| Command | Description |
|---------|-------------|
| `login` | Authenticate with email and password |
| `register` | Create a new account |
| `logout` | Logout and clear stored token |
| `add` | Create a new task |
| `list` | Show all tasks |
| `update` | Update task description or status |
| `delete` | Delete a task |
| `status` | Toggle task completion status |
| `process` | Process all tasks in parallel |
| `clear` | Clear task description |
| `help` | Show available commands |
| `exit` | Save and exit the application

**CLI Configuration:**
```bash
# Set custom server URL
export TASK_SERVER_URL="http://localhost:3000"
```

### REST API Examples

**Health Check:**
```bash
curl http://localhost:8080/health
```

**Register a User:**
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

**Login:**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

**Create a Task:**
```bash
curl -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{"description":"My first task"}'
```

**Get All Tasks:**
```bash
curl -H "Authorization: Bearer <your_token>" http://localhost:8080/tasks
```

**Get Single Task:**
```bash
curl -H "Authorization: Bearer <your_token>" http://localhost:8080/tasks/1
```

**Update a Task:**
```bash
curl -X PUT http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{"description":"Updated task","done":true}'
```

**Delete a Task:**
```bash
curl -X DELETE http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer <your_token>"
```

---

## Environment Variables

### Server Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TASKMANAGER_JWT_SECRET` | **Yes** | — | Secret key for JWT signing (min 32 chars) |
| `TASKMANAGER_DATABASE_PATH` | No | `./data/tasks.db` | Path to SQLite database file |
| `TASKMANAGER_SERVER_PORT` | No | `8080` | HTTP server listening port |
| `TASKMANAGER_SERVER_HOST` | No | `0.0.0.0` | HTTP server host address |
| `TASKMANAGER_JWT_EXPIRATION` | No | `24h` | JWT token expiration duration |

### Logging Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TASKMANAGER_LOG_LEVEL` | No | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `TASKMANAGER_LOG_FORMAT` | No | `json` | Log format: `json` or `text` |
| `TASKMANAGER_LOG_OUTPUT` | No | `stderr` | Log output: `stdout`, `stderr`, or file path |

### CLI Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `TASK_SERVER_URL` | No | `http://localhost:8080` | Server URL for CLI client |

---

## Configuration File

The server supports configuration via YAML or JSON files:

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  shutdown_timeout: "30s"

database:
  path: "./data/tasks.db"

jwt:
  secret: "CHANGE_ME_IN_PRODUCTION_MIN_32_CHARS"
  expiration: "24h"

logging:
  level: "info"
  format: "json"
  output: "stderr"
  service_name: "task-manager-api"
  environment: "production"
```

**Configuration precedence:**
1. Command-line flags (highest priority)
2. Environment variables (`TASKMANAGER_*`)
3. Configuration file (`config.yaml` or `config.json`)
4. Default values (lowest priority)

---

## Architecture

The project follows the principles of **Clean Architecture**:

| Layer | Package | Responsibility |
|-------|---------|----------------|
| **Domain** | `internal/domain` | Pure business models and interfaces (no external dependencies) |
| **Application** | `application` | Use cases and business logic orchestration |
| **Adapters** | `adapters/*` | External implementations (HTTP server, SQLite storage) |
| **Infrastructure** | `infrastructure`, `logger`, `auth`, `validation` | Cross-cutting concerns |

**Key directories:**
- `cmd/server` — HTTP server entry point
- `cmd/cli` — Interactive CLI client
- `internal/handlers` — HTTP request handlers
- `adapters/storage` — SQLite persistence layer
- `auth` — JWT authentication and password hashing

---

## Testing

I prioritize reliability through comprehensive testing:

```sh
# Run all tests
go test ./...

# Run specific test
go test -run TestFunctionName ./...

# Verbose output
go test -v ./...

# Coverage report
go test -cover ./...

# Race detector
go test -race ./...

# Integration tests (requires Docker)
export TESTCONTAINERS_RYUK_DISABLED=true
go test ./adapters/storage/... -tags=integration
```

**Note for Fedora 43+:** Disable testcontainers Ryuk:
```sh
export TESTCONTAINERS_RYUK_DISABLED=true
```

---

## Docker

### Basic Usage

```sh
# Build and run
docker-compose up -d --build

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### With Loki (Log Aggregation)

```sh
# Run with Loki for centralized logging
docker-compose -f docker-compose.yml -f docker-compose.loki.yml up -d
```

### Manual Docker Run

```sh
docker run -d \
  -p 8080:8080 \
  -e TASKMANAGER_JWT_SECRET="your-secure-secret-key" \
  -e TASKMANAGER_DATABASE_PATH="/data/tasks.db" \
  -v ./data:/data \
  --name task-manager \
  task-manager:latest
```

---

## Roadmap

- [x] JWT Authentication
- [x] REST API Implementation
- [x] SQLite Storage (CGO-free)
- [x] Interactive CLI Tool
- [x] Structured Logging with Rotation
- [x] Docker & Docker Compose Support
- [ ] Full gRPC Implementation
- [ ] Integration with Prometheus/Grafana
- [ ] Frontend Web Dashboard

---

## License

Distributed under the MIT License

---

## Contact

**Anna Nurgaleeva**
*   **LinkedIn:** [Anna Nurgaleeva](https://www.linkedin.com/in/anna-nurgaleeva-ba9a6338)
*   **Telegram:** [@bryacka](https://t.me/bryacka)


<!-- MARKDOWN LINKS & IMAGES -->
[contributors-shield]: https://img.shields.io/github/contributors/bryack/Go.svg?style=for-the-badge
[contributors-url]: https://github.com/bryack/Go/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/bryack/Go.svg?style=for-the-badge
[forks-url]: https://github.com/bryack/Go/network/members
[stars-shield]: https://img.shields.io/github/stars/bryack/Go.svg?style=for-the-badge
[stars-url]: https://github.com/bryack/Go/stargazers
[issues-shield]: https://img.shields.io/github/issues/bryack/Go.svg?style=for-the-badge
[issues-url]: https://github.com/bryack/Go/issues
[license-shield]: https://img.shields.io/github/license/bryack/Go.svg?style=for-the-badge
[license-url]: https://github.com/bryack/Go/blob/main/LICENSE
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://www.linkedin.com/in/anna-nurgaleeva-ba9a6338
