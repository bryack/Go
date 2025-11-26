# Design Document

## Overview

This design implements containerization and CI/CD deployment for the to-do list application using Docker and GitHub Actions. The solution uses a multi-stage Docker build to create optimized container images for both the CLI and server components, with GitHub Actions automating the build, test, and deployment pipeline.

The application is a Go-based task manager with two entry points:
- **Server**: HTTP API server running on port 8080
- **CLI**: Interactive command-line interface

Both components use SQLite database storage (modernc.org/sqlite) and share common business logic.

## Architecture

### Containerization Strategy

**Multi-Stage Docker Build:**
1. **Builder Stage**: Compiles Go binaries with all dependencies
2. **Runtime Stage**: Creates minimal final image with only necessary runtime files

**Benefits:**
- Reduced image size (no build tools in final image)
- Faster deployment and startup times
- Improved security (minimal attack surface)
- Layer caching for faster rebuilds

### CI/CD Pipeline Flow

```
Push to main → Run Tests → Build Docker Image → Tag Image → (Optional) Push to Registry
```

**Pipeline Stages:**
1. **Checkout**: Clone repository code
2. **Setup Go**: Install Go toolchain
3. **Run Tests**: Execute all unit tests
4. **Build Docker Image**: Create container image
5. **Tag Image**: Apply version tags (commit SHA, latest)

## Components and Interfaces

### Dockerfile Structure

**Builder Stage:**
- Base image: `golang:1.24-alpine` (matches go.mod version)
- Working directory: `/build`
- Dependencies: Copy go.mod/go.sum and download modules
- Build: Compile both CLI and server binaries with CGO enabled for SQLite
- CGO configuration: Required for modernc.org/sqlite

**Runtime Stage:**
- Base image: `alpine:latest` (minimal Linux distribution)
- Runtime dependencies: `ca-certificates` for HTTPS, `sqlite-libs` for database
- Binaries: Copy compiled CLI and server from builder
- Database: Create `/data` directory for SQLite file persistence
- Default command: Run server on port 8080
- Exposed port: 8080

### GitHub Actions Workflow

**Workflow File:** `.github/workflows/deploy.yml`

**Trigger Events:**
- Push to `main` branch
- Pull requests to `main` branch (optional)

**Jobs:**

1. **Test Job:**
   - Runs on: `ubuntu-latest`
   - Steps:
     - Checkout code
     - Setup Go 1.24
     - Cache Go modules
     - Run `go test ./...`
   - Fail fast: Stop pipeline if tests fail

2. **Build Job:**
   - Depends on: Test job success
   - Runs on: `ubuntu-latest`
   - Steps:
     - Checkout code
     - Setup Docker Buildx (for advanced build features)
     - Build Docker image
     - Tag with commit SHA and `latest`
     - (Optional) Login to container registry
     - (Optional) Push image to registry

### Environment Configuration

**Environment Variables:**
- `PORT`: Server listening port (default: 8080)
- `DB_PATH`: SQLite database file path (default: `/data/tasks.db`)

**Volume Mounts:**
- `/data`: Persistent storage for SQLite database

## Data Models

### Docker Image Metadata

```
Image Name: task-manager
Tags:
  - latest (most recent build)
  - <commit-sha> (specific version)
  - <version-tag> (semantic version, if tagged)

Labels:
  - org.opencontainers.image.source: GitHub repository URL
  - org.opencontainers.image.revision: Git commit SHA
  - org.opencontainers.image.created: Build timestamp
```

### Build Artifacts

```
Binaries:
  - /app/server (HTTP API server)
  - /app/cli (Interactive CLI tool)

Database:
  - /data/tasks.db (SQLite database file)
```

## Error Handling

### Build Failures

**Go Module Download Errors:**
- Cause: Network issues, invalid dependencies
- Handling: Docker build fails with clear error message
- Recovery: Retry build, check go.mod/go.sum integrity

**Compilation Errors:**
- Cause: Syntax errors, type errors, missing dependencies
- Handling: Build stage fails with Go compiler output
- Recovery: Fix code errors, ensure all imports are valid

**CGO Compilation Errors:**
- Cause: Missing C compiler, SQLite library issues
- Handling: Install build-base and sqlite-dev in builder stage
- Recovery: Ensure alpine packages are installed

### Runtime Failures

**Database Connection Errors:**
- Cause: Missing /data directory, permission issues
- Handling: Application logs error and exits
- Recovery: Ensure volume is mounted, check file permissions

**Port Binding Errors:**
- Cause: Port 8080 already in use
- Handling: Server fails to start with clear error
- Recovery: Use different port via PORT environment variable

### CI/CD Pipeline Failures

**Test Failures:**
- Cause: Failing unit tests
- Handling: Pipeline stops, no image is built
- Recovery: Fix failing tests before merging

**Docker Build Failures:**
- Cause: Dockerfile syntax errors, missing files
- Handling: GitHub Actions job fails with build logs
- Recovery: Fix Dockerfile, ensure all required files exist

**Registry Push Failures:**
- Cause: Authentication issues, network problems
- Handling: Build succeeds but push fails (logged)
- Recovery: Check registry credentials, retry push

## Testing Strategy

### Pre-Deployment Testing

**Unit Tests:**
- Run all existing Go tests via `go test ./...`
- Must pass before Docker build proceeds
- Coverage: task operations, validation, storage, handlers

**Integration Tests:**
- Test database storage operations
- Verify API endpoints functionality
- Ensure CLI commands work correctly

### Container Testing

**Build Verification:**
- Dockerfile builds successfully
- Image size is reasonable (<50MB for runtime)
- All required binaries are present

**Runtime Verification:**
- Container starts without errors
- Server responds to health check endpoint
- Database file is created in /data directory
- CLI tool executes basic commands

**Manual Testing Commands:**
```bash
# Build image
docker build -t task-manager .

# Run server
docker run -p 8080:8080 -v $(pwd)/data:/data task-manager

# Test health endpoint
curl http://localhost:8080/health

# Run CLI
docker run -it -v $(pwd)/data:/data task-manager /app/cli
```

### CI/CD Testing

**Pipeline Validation:**
- Workflow syntax is valid
- All jobs execute in correct order
- Secrets and environment variables are configured
- Build artifacts are properly tagged

**Deployment Verification:**
- Image is pushed to registry (if configured)
- Tags are applied correctly
- Image can be pulled and run successfully

## Implementation Notes

### CGO Requirements

The application uses `modernc.org/sqlite` which requires CGO for compilation:
- Set `CGO_ENABLED=1` during build
- Install gcc and musl-dev in builder stage
- Ensure sqlite-libs in runtime stage

### Database Persistence

SQLite database requires persistent storage:
- Use Docker volumes for /data directory
- Database file: /data/tasks.db
- Migrations run automatically on startup

### Port Configuration

Server listens on port 8080 by default:
- Exposed in Dockerfile
- Configurable via PORT environment variable
- Map to host port using `-p` flag

### Security Considerations

- Run container as non-root user (optional enhancement)
- Use specific Go version (not latest)
- Minimal runtime image (Alpine)
- No secrets in Dockerfile or code
- Use GitHub Secrets for registry credentials
