# Design Document

## Overview

This design transforms the CLI from a direct database client into an HTTP client that communicates with the existing task management server. The migration eliminates tight coupling between CLI and storage layers, allowing independent evolution of both components.

The design introduces three new packages within the CLI codebase:
- **client**: HTTP API client for server communication
- **auth**: Authentication and token management
- **config**: Configuration management for server URL and settings

The existing CLI command structure and user interface remain unchanged, ensuring a seamless user experience during migration.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────┐
│           CLI Application               │
│  ┌───────────────────────────────────┐  │
│  │     Command Handlers              │  │
│  │  (add, list, status, update...)   │  │
│  └───────────┬───────────────────────┘  │
│              │                           │
│  ┌───────────▼───────────────────────┐  │
│  │      API Client Layer             │  │
│  │  - HTTP requests                  │  │
│  │  - Token injection                │  │
│  │  - Error handling                 │  │
│  └───────────┬───────────────────────┘  │
│              │                           │
│  ┌───────────▼───────────────────────┐  │
│  │   Auth Manager                    │  │
│  │  - Token storage                  │  │
│  │  - Login/Register                 │  │
│  └───────────────────────────────────┘  │
└──────────────┬──────────────────────────┘
               │ HTTP/JSON
               │
┌──────────────▼──────────────────────────┐
│         HTTP Server                     │
│  ┌───────────────────────────────────┐  │
│  │   REST API Endpoints              │  │
│  │  /register, /login, /tasks        │  │
│  └───────────┬───────────────────────┘  │
│              │                           │
│  ┌───────────▼───────────────────────┐  │
│  │   JWT Middleware                  │  │
│  └───────────┬───────────────────────┘  │
│              │                           │
│  ┌───────────▼───────────────────────┐  │
│  │   Storage Layer                   │  │
│  │  (SQLite with user isolation)     │  │
│  └───────────────────────────────────┘  │
└─────────────────────────────────────────┘
```

### Component Interaction Flow

**Startup Flow:**
1. CLI reads configuration (server URL from env/config)
2. CLI attempts to load stored JWT token from `~/.task-cli/token`
3. If token exists and valid → proceed to command loop
4. If no token or invalid → prompt for login/register

**Command Execution Flow:**
1. User enters command (e.g., "add")
2. Command handler validates input
3. API Client constructs HTTP request with JWT token
4. Server validates token and processes request
5. API Client receives response
6. CLI displays formatted result to user

**Authentication Flow:**
1. User provides email/password
2. CLI sends POST to `/login` or `/register`
3. Server validates credentials and returns JWT token
4. CLI stores token in `~/.task-cli/token` with 0600 permissions
5. Token is used for all subsequent requests

## Components and Interfaces

### 1. API Client Package (`cmd/cli/client`)

**Purpose:** Handles all HTTP communication with the server.

**Core Interface:**
```go
type TaskClient interface {
    // Task operations
    GetTasks() ([]Task, error)
    GetTask(id int) (Task, error)
    CreateTask(description string) (Task, error)
    UpdateTask(id int, description *string, done *bool) (Task, error)
    DeleteTask(id int) error
    
    // Authentication
    Login(email, password string) (string, error)
    Register(email, password string) (string, error)
    
    // Configuration
    SetToken(token string)
    GetServerURL() string
}
```

**Implementation Structure:**
```go
type HTTPClient struct {
    baseURL    string
    httpClient *http.Client
    token      string
}

func NewHTTPClient(baseURL string) *HTTPClient {
    return &HTTPClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}
```

**Key Methods:**

- `doRequest(method, path string, body, result interface{}) error`
  - Generic HTTP request handler
  - Adds Authorization header with Bearer token
  - Handles JSON encoding/decoding
  - Maps HTTP status codes to appropriate errors

- `handleErrorResponse(resp *http.Response) error`
  - Parses error responses from server
  - Returns user-friendly error messages
  - Detects 401 for re-authentication trigger

**Error Handling Strategy:**
- Network errors → "Cannot connect to server at {URL}"
- 401 Unauthorized → Clear token, trigger re-authentication
- 4xx errors → Display server error message
- 5xx errors → "Server error, please try again later"
- Timeout → "Request timed out, check your connection"

### 2. Auth Manager Package (`cmd/cli/auth`)

**Purpose:** Manages authentication state and token persistence.

**Core Interface:**
```go
type AuthManager interface {
    // Token management
    LoadToken() (string, error)
    SaveToken(token string) error
    ClearToken() error
    
    // Authentication state
    IsAuthenticated() bool
    RequireAuth() (string, error)
    
    // Interactive authentication
    PromptLogin() (string, error)
    PromptRegister() (string, error)
}
```

**Implementation Structure:**
```go
type FileAuthManager struct {
    tokenPath  string
    client     TaskClient
    input      InputReader
    output     io.Writer
}

func NewFileAuthManager(client TaskClient, input InputReader, output io.Writer) *FileAuthManager {
    homeDir, _ := os.UserHomeDir()
    tokenPath := filepath.Join(homeDir, ".task-cli", "token")
    
    return &FileAuthManager{
        tokenPath: tokenPath,
        client:    client,
        input:     input,
        output:    output,
    }
}
```

**Token Storage Details:**
- Location: `~/.task-cli/token`
- Format: Plain text JWT token (single line)
- Permissions: 0600 (read/write for owner only)
- Directory creation: Automatic with 0700 permissions

**Authentication Flow Methods:**

- `RequireAuth() (string, error)`
  - Attempts to load stored token
  - If not found or invalid, prompts for login
  - Returns valid token or error

- `PromptLogin() (string, error)`
  - Prompts user for email and password
  - Calls client.Login()
  - Saves token on success
  - Returns token or error

- `PromptRegister() (string, error)`
  - Prompts user for email and password
  - Validates password requirements locally
  - Calls client.Register()
  - Saves token on success
  - Returns token or error

### 3. Config Package (`cmd/cli/config`)

**Purpose:** Manages CLI configuration including server URL.

**Core Interface:**
```go
type Config struct {
    ServerURL string
}

func LoadConfig() (*Config, error)
func (c *Config) Validate() error
```

**Configuration Priority (highest to lowest):**
1. Command-line flag: `--server-url`
2. Environment variable: `TASK_SERVER_URL`
3. Config file: `~/.task-cli/config.json`
4. Default: `http://localhost:8080`

**Config File Format:**
```json
{
  "server_url": "http://localhost:8080"
}
```

### 4. Modified CLI Package (`cmd/cli`)

**Changes to Existing Structure:**

**CLI Struct Modification:**
```go
type CLI struct {
    input       InputReader
    output      io.Writer
    client      client.TaskClient      // Replaces taskManager and storage
    authManager auth.AuthManager
    config      *config.Config
}
```

**Initialization Changes:**
```go
func NewCLI(input InputReader, output io.Writer, cfg *config.Config) (*CLI, error) {
    // Create HTTP client
    httpClient := client.NewHTTPClient(cfg.ServerURL)
    
    // Create auth manager
    authMgr := auth.NewFileAuthManager(httpClient, input, output)
    
    // Attempt authentication
    token, err := authMgr.RequireAuth()
    if err != nil {
        return nil, fmt.Errorf("authentication failed: %w", err)
    }
    
    httpClient.SetToken(token)
    
    return &CLI{
        input:       input,
        output:      output,
        client:      httpClient,
        authManager: authMgr,
        config:      cfg,
    }, nil
}
```

**Command Handler Modifications:**

All command handlers will be updated to use the API client instead of direct storage access:

- `handleAddCommand()`: Calls `client.CreateTask(description)`
- `handleListCommand()`: Calls `client.GetTasks()`
- `handleStatusCommand()`: Calls `client.UpdateTask(id, nil, &done)`
- `handleUpdateCommand()`: Calls `client.UpdateTask(id, &description, nil)`
- `handleDeleteCommand()`: Calls `client.DeleteTask(id)`
- `handleClearCommand()`: Calls `client.UpdateTask(id, &emptyString, nil)`

**New Commands:**

- `login`: Manually trigger login flow
- `register`: Manually trigger registration flow
- `logout`: Clear stored token
- `whoami`: Display current authenticated user (if server supports it)

### 5. Migration Strategy

**Phase 1: Add New Components (Non-Breaking)**
- Create `client`, `auth`, and `config` packages
- Implement HTTP client with full API coverage
- Add authentication manager with token storage
- Keep existing CLI code unchanged

**Phase 2: Add Mode Detection**
- Add `--mode` flag: `legacy` or `client`
- Environment variable: `TASK_CLI_MODE`
- Default to `legacy` mode initially
- Display deprecation warning in legacy mode

**Phase 3: Update CLI to Support Both Modes**
- Modify CLI struct to support both storage and client
- Use interface abstraction for task operations
- Route commands based on mode selection

**Phase 4: Switch Default to Client Mode**
- Change default mode to `client`
- Update documentation
- Provide migration guide

**Phase 5: Remove Legacy Mode**
- Remove direct storage dependencies
- Remove legacy mode code
- Simplify CLI structure

## Data Models

### Task Model (Shared)

The Task struct remains unchanged for compatibility:

```go
type Task struct {
    ID          int    `json:"id"`
    Description string `json:"description"`
    Done        bool   `json:"done"`
}
```

### API Request/Response Models

**CreateTaskRequest:**
```go
type CreateTaskRequest struct {
    Description string `json:"description"`
}
```

**UpdateTaskRequest:**
```go
type UpdateTaskRequest struct {
    Description *string `json:"description,omitempty"`
    Done        *bool   `json:"done,omitempty"`
}
```

**AuthRequest:**
```go
type AuthRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}
```

**AuthResponse:**
```go
type AuthResponse struct {
    Token string `json:"token"`
    Email string `json:"email"`
}
```

**ErrorResponse:**
```go
type ErrorResponse struct {
    Error string `json:"error"`
}
```

## Error Handling

### Error Types

**Network Errors:**
```go
type NetworkError struct {
    URL string
    Err error
}

func (e *NetworkError) Error() string {
    return fmt.Sprintf("cannot connect to server at %s: %v", e.URL, e.Err)
}
```

**Authentication Errors:**
```go
type AuthError struct {
    Message string
}

func (e *AuthError) Error() string {
    return e.Message
}
```

**API Errors:**
```go
type APIError struct {
    StatusCode int
    Message    string
}

func (e *APIError) Error() string {
    return e.Message
}
```

### Error Handling Flow

1. **Network Failure:**
   - Display connection error with server URL
   - Suggest checking server status
   - Exit gracefully

2. **401 Unauthorized:**
   - Clear stored token
   - Prompt for re-authentication
   - Retry original request

3. **4xx Client Errors:**
   - Display server error message
   - Return to command prompt
   - Don't exit CLI

4. **5xx Server Errors:**
   - Display generic server error
   - Suggest retrying later
   - Return to command prompt

5. **Timeout:**
   - Display timeout message
   - Suggest checking connection
   - Return to command prompt

## Testing Strategy

### Unit Tests

**API Client Tests:**
- Mock HTTP server for testing requests
- Test token injection in headers
- Test error response parsing
- Test timeout handling
- Test all CRUD operations

**Auth Manager Tests:**
- Test token file creation with correct permissions
- Test token loading and validation
- Test authentication prompts
- Test token clearing
- Mock file system operations

**Config Tests:**
- Test configuration loading priority
- Test URL validation
- Test default values
- Test environment variable parsing

### Integration Tests

**End-to-End Flow Tests:**
- Start test HTTP server
- Test full registration flow
- Test full login flow
- Test task operations with authentication
- Test token expiration handling
- Test network failure scenarios

**Migration Tests:**
- Test legacy mode operation
- Test client mode operation
- Test mode switching
- Test data consistency between modes

### Manual Testing Checklist

- [ ] Register new user from CLI
- [ ] Login with existing credentials
- [ ] Create task via CLI
- [ ] List tasks via CLI
- [ ] Update task status via CLI
- [ ] Update task description via CLI
- [ ] Delete task via CLI
- [ ] Logout and verify token cleared
- [ ] Test with server offline
- [ ] Test with invalid token
- [ ] Test with expired token
- [ ] Test server URL configuration
- [ ] Test migration from legacy to client mode

## Security Considerations

### Token Storage Security

1. **File Permissions:**
   - Token file created with 0600 permissions
   - Directory created with 0700 permissions
   - Verify permissions on load, warn if too permissive

2. **Token Validation:**
   - No local token validation (server-side only)
   - Clear token immediately on 401 response
   - Don't log token values

3. **Password Handling:**
   - Never store passwords locally
   - Use terminal input masking for password entry
   - Validate password requirements before sending

### Network Security

1. **HTTPS Support:**
   - Support both HTTP and HTTPS URLs
   - Warn when using HTTP in production
   - Validate SSL certificates

2. **Timeout Configuration:**
   - Default 30-second timeout
   - Prevent indefinite hangs
   - Configurable via environment variable

## Performance Considerations

### HTTP Client Optimization

1. **Connection Reuse:**
   - Use single http.Client instance
   - Enable keep-alive connections
   - Connection pooling handled by Go's http package

2. **Request Timeouts:**
   - 30-second default timeout
   - Separate timeout for authentication (60 seconds)
   - Configurable via environment variable

3. **Response Buffering:**
   - Stream large responses when possible
   - Limit response body size
   - Close response bodies properly

### Caching Strategy

**Phase 1 (MVP):** No caching - all requests go to server

**Future Enhancement:**
- Cache task list locally
- Invalidate on mutations
- Sync on startup
- Offline mode support

## Deployment and Migration

### User Migration Guide

**For Existing CLI Users:**

1. Update to new CLI version
2. Set `TASK_SERVER_URL` environment variable (if not localhost)
3. Run CLI - will prompt for login/register
4. Existing local database remains untouched
5. Server must be running for CLI to work

**Environment Variables:**
```bash
export TASK_SERVER_URL="http://localhost:8080"
export JWT_SECRET_KEY="your-secret-key"  # Server only
```

**First Run Experience:**
```
🚀 Task Manager CLI (Client Mode)
📡 Server: http://localhost:8080

No authentication token found.
Choose an option:
1. Login with existing account
2. Register new account
3. Exit

Enter choice (1-3):
```

### Backward Compatibility

**Legacy Mode Support:**
- Keep legacy mode available via `--legacy` flag
- Display deprecation warning
- Document migration path
- Plan removal timeline (e.g., 3 months)

**Data Migration:**
- No automatic data migration
- Users must manually recreate tasks or
- Provide separate migration tool to import from local DB

## Future Enhancements

### Phase 2 Features

1. **Offline Mode:**
   - Local cache of tasks
   - Queue mutations for sync
   - Conflict resolution

2. **Batch Operations:**
   - Bulk task creation
   - Bulk status updates
   - Import/export functionality

3. **Advanced Features:**
   - Task filtering and search
   - Task categories/tags
   - Due dates and reminders
   - Task sharing between users

4. **Configuration Enhancements:**
   - Multiple server profiles
   - Custom output formatting
   - Color scheme customization
   - Shell completion scripts

### Monitoring and Observability

1. **Logging:**
   - Optional debug logging
   - Request/response logging
   - Error logging to file

2. **Metrics:**
   - Request latency tracking
   - Error rate monitoring
   - Usage statistics (opt-in)

3. **Health Checks:**
   - Server connectivity check command
   - Version compatibility check
   - Configuration validation command
