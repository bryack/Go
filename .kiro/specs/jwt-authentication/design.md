# JWT Authentication Design Document

## Overview

This design document outlines the implementation of JWT-based authentication for the Task Manager API. The authentication system will be built using Go's standard library along with third-party packages for JWT handling and password hashing. The design follows a layered architecture with clear separation between authentication logic, middleware, storage, and HTTP handlers.

The system will secure all task-related endpoints while keeping health check, registration, and login endpoints public. Each user will have isolated access to their own tasks, enforced through JWT token validation and user ID association.

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        HTTP Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Auth Handler │  │ Task Handler │  │Public Handler│      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Middleware Layer                          │
│              ┌──────────────────────┐                        │
│              │  Auth Middleware     │                        │
│              │  (JWT Validation)    │                        │
│              └──────────────────────┘                        │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                            │
│  ┌──────────────┐  ┌──────────────┐                         │
│  │  Auth Service│  │  JWT Service │                         │
│  │  (User Mgmt) │  │  (Token Ops) │                         │
│  └──────────────┘  └──────────────┘                         │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Storage Layer                             │
│              ┌──────────────────────┐                        │
│              │  Database Storage    │                        │
│              │  (Users + Tasks)     │                        │
│              └──────────────────────┘                        │
└─────────────────────────────────────────────────────────────┘
```

### Component Interaction Flow

**Registration Flow:**
1. Client sends POST /auth/register with email and password
2. Auth Handler validates input and calls Auth Service
3. Auth Service hashes password and stores user in database
4. JWT Service generates token with user ID
5. Handler returns token to client

**Login Flow:**
1. Client sends POST /auth/login with credentials
2. Auth Handler calls Auth Service to validate credentials
3. Auth Service retrieves user, verifies password hash
4. JWT Service generates token with user ID
5. Handler returns token to client

**Protected Endpoint Flow:**
1. Client sends request with Authorization: Bearer <token>
2. Auth Middleware extracts and validates JWT token
3. Middleware adds user ID to request context
4. Handler retrieves user ID from context
5. Handler filters data by user ID before responding

## Components and Interfaces

### 1. Auth Package (`auth/`)

#### JWT Service (`auth/jwt.go`)

Handles all JWT token operations including generation, validation, and claims extraction.

```go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT token payload
type Claims struct {
    UserID int `json:"user_id"`
    jwt.RegisteredClaims
}

// JWTService handles token generation and validation
type JWTService struct {
    secretKey []byte
    expiration time.Duration
}

// NewJWTService creates a new JWT service with secret key
func NewJWTService(secretKey string, expiration time.Duration) *JWTService

// GenerateToken creates a new JWT token for the given user ID
func (j *JWTService) GenerateToken(userID int) (string, error)

// ValidateToken verifies token signature and expiration
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error)

// ExtractClaims parses token and returns claims without full validation
func (j *JWTService) ExtractClaims(tokenString string) (*Claims, error)
```

#### Auth Service (`auth/service.go`)

Manages user authentication operations including registration and login.

```go
package auth

import (
    "myproject/storage"
    "golang.org/x/crypto/bcrypt"
)

// Service handles authentication business logic
type Service struct {
    userStorage storage.UserStorage
    jwtService  *JWTService
}

// NewService creates a new authentication service
func NewService(userStorage storage.UserStorage, jwtService *JWTService) *Service

// Register creates a new user account and returns a JWT token
func (s *Service) Register(email, password string) (token string, err error)

// Login validates credentials and returns a JWT token
func (s *Service) Login(email, password string) (token string, err error)

// ValidatePassword checks if password meets requirements
func (s *Service) ValidatePassword(password string) error

// HashPassword creates bcrypt hash of password
func (s *Service) HashPassword(password string) (string, error)

// ComparePassword verifies password against hash
func (s *Service) ComparePassword(hashedPassword, password string) error
```

### 2. Middleware Package (`middleware/`)

#### Auth Middleware (`middleware/auth.go`)

HTTP middleware that validates JWT tokens and enforces authentication.

```go
package middleware

import (
    "net/http"
    "myproject/auth"
    "context"
)

// ContextKey type for context values
type ContextKey string

const UserIDKey ContextKey = "user_id"

// AuthMiddleware wraps handlers with JWT authentication
type AuthMiddleware struct {
    jwtService *auth.JWTService
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware

// Authenticate wraps an http.HandlerFunc with JWT validation
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc

// ExtractToken retrieves JWT token from Authorization header
func (m *AuthMiddleware) ExtractToken(r *http.Request) (string, error)

// GetUserIDFromContext retrieves user ID from request context
func GetUserIDFromContext(ctx context.Context) (int, error)
```

### 3. Storage Layer Extensions

#### User Storage Interface (`storage/user.go`)

Extends the storage layer to handle user data persistence.

```go
package storage

// User represents a user account
type User struct {
    ID           int       `json:"id"`
    Email        string    `json:"email"`
    PasswordHash string    `json:"-"` // Never serialize password
    CreatedAt    time.Time `json:"created_at"`
}

// UserStorage defines user persistence operations
type UserStorage interface {
    CreateUser(email, passwordHash string) (int, error)
    GetUserByEmail(email string) (*User, error)
    GetUserByID(id int) (*User, error)
    EmailExists(email string) (bool, error)
}
```

#### Database Storage Extension (`storage/database.go`)

Implements UserStorage interface in the existing DatabaseStorage struct.

```go
// Add to existing DatabaseStorage struct
func (ds *DatabaseStorage) CreateUser(email, passwordHash string) (int, error)
func (ds *DatabaseStorage) GetUserByEmail(email string) (*User, error)
func (ds *DatabaseStorage) GetUserByID(id int) (*User, error)
func (ds *DatabaseStorage) EmailExists(email string) (bool, error)
```

#### Migration for Users Table (`storage/migrations.go`)

Add new migration to create users table.

```go
// Add to NewMigratorWithDefaults
usersMigration := Migration{
    Version: 2,
    Name:    "create_users_table",
    Up: `
        CREATE TABLE users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            email TEXT NOT NULL UNIQUE,
            password_hash TEXT NOT NULL,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        
        CREATE UNIQUE INDEX idx_users_email ON users(email);
    `,
    Down: `
        DROP INDEX IF EXISTS idx_users_email;
        DROP TABLE IF EXISTS users;
    `,
}
```

#### Migration for Task User Association (`storage/migrations.go`)

Add migration to associate tasks with users.

```go
taskUserMigration := Migration{
    Version: 3,
    Name:    "add_user_id_to_tasks",
    Up: `
        ALTER TABLE tasks ADD COLUMN user_id INTEGER;
        CREATE INDEX idx_tasks_user_id ON tasks(user_id);
    `,
    Down: `
        DROP INDEX IF EXISTS idx_tasks_user_id;
        -- SQLite doesn't support DROP COLUMN easily, would need table recreation
    `,
}
```

### 4. HTTP Handlers

#### Auth Handlers (`internal/handlers/auth.go`)

HTTP handlers for authentication endpoints.

```go
package handlers

import (
    "net/http"
    "myproject/auth"
)

// RegisterRequest represents registration payload
type RegisterRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// LoginRequest represents login payload
type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

// AuthResponse represents auth success response
type AuthResponse struct {
    Token string `json:"token"`
    Email string `json:"email"`
}

// RegisterHandler handles user registration
func RegisterHandler(authService *auth.Service) http.HandlerFunc

// LoginHandler handles user login
func LoginHandler(authService *auth.Service) http.HandlerFunc
```

#### Modified Task Handlers (`cmd/server/main.go`)

Update existing task handlers to filter by user ID from context.

```go
// Modify tasksHandler to filter by user ID
func tasksHandler(s storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, err := middleware.GetUserIDFromContext(r.Context())
        // Filter tasks by userID
    }
}

// Modify taskHandler to verify ownership
func taskHandler(s storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, err := middleware.GetUserIDFromContext(r.Context())
        // Verify task belongs to userID
    }
}
```

## Data Models

### User Table Schema

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_users_email ON users(email);
```

### Updated Tasks Table Schema

```sql
ALTER TABLE tasks ADD COLUMN user_id INTEGER;
CREATE INDEX idx_tasks_user_id ON tasks(user_id);
```

### JWT Token Structure

```json
{
  "user_id": 123,
  "exp": 1699999999,
  "iat": 1699913599,
  "iss": "task-manager-api"
}
```

## Error Handling

### Authentication Errors

| Error Scenario | HTTP Status | Error Message |
|---------------|-------------|---------------|
| Missing Authorization header | 401 | "Authorization header required" |
| Invalid token format | 401 | "Invalid token format" |
| Expired token | 401 | "Token has expired" |
| Invalid signature | 401 | "Invalid token signature" |
| User not found | 401 | "Invalid credentials" |
| Wrong password | 401 | "Invalid credentials" |
| Email already exists | 400 | "Email already registered" |
| Invalid email format | 400 | "Invalid email format" |
| Password too short | 400 | "Password must be at least 8 characters" |
| Task not found or unauthorized | 404 | "Task not found" |

### Error Response Format

All errors follow the existing JSON error format:

```json
{
  "error": "Error message here"
}
```

## Testing Strategy

### Unit Tests

1. **JWT Service Tests** (`auth/jwt_test.go`)
   - Test token generation with valid user ID
   - Test token validation with valid token
   - Test token validation with expired token
   - Test token validation with invalid signature
   - Test claims extraction

2. **Auth Service Tests** (`auth/service_test.go`)
   - Test user registration with valid data
   - Test registration with duplicate email
   - Test login with valid credentials
   - Test login with invalid credentials
   - Test password hashing and comparison
   - Test password validation rules

3. **Middleware Tests** (`middleware/auth_test.go`)
   - Test authentication with valid token
   - Test authentication with missing token
   - Test authentication with invalid token
   - Test token extraction from header
   - Test user ID context propagation

4. **Storage Tests** (`storage/user_test.go`)
   - Test user creation
   - Test user retrieval by email
   - Test user retrieval by ID
   - Test email existence check
   - Test unique email constraint

### Integration Tests

1. **Registration Flow Test**
   - Register new user → verify token returned
   - Register duplicate email → verify error

2. **Login Flow Test**
   - Login with valid credentials → verify token
   - Login with invalid credentials → verify error

3. **Protected Endpoint Test**
   - Access task endpoint with valid token → success
   - Access task endpoint without token → 401 error
   - Access task endpoint with expired token → 401 error

4. **Task Isolation Test**
   - User A creates task → User B cannot access it
   - User A lists tasks → only sees their own tasks

### Security Testing

1. Test bcrypt password hashing (verify not plaintext)
2. Test JWT secret key from environment variable
3. Test token expiration enforcement
4. Test SQL injection prevention in user queries
5. Test that password hashes are never returned in API responses

## Dependencies

### Required Go Packages

```go
require (
    github.com/golang-jwt/jwt/v5 v5.2.0
    golang.org/x/crypto v0.17.0
)
```

- `github.com/golang-jwt/jwt/v5`: Industry-standard JWT implementation for Go
- `golang.org/x/crypto/bcrypt`: Secure password hashing using bcrypt algorithm

### Environment Variables

```bash
JWT_SECRET_KEY=<random-secret-key>  # Required for JWT signing
JWT_EXPIRATION_HOURS=24             # Optional, defaults to 24
TASK_DB_PATH=./tasks.db             # Existing variable
```

## Security Considerations

1. **Password Storage**: All passwords hashed with bcrypt (cost factor 10)
2. **JWT Secret**: Stored in environment variable, never in code
3. **Token Expiration**: 24-hour expiration enforced
4. **HTTPS**: Recommend HTTPS in production (outside scope of this implementation)
5. **Rate Limiting**: Consider adding rate limiting to auth endpoints (future enhancement)
6. **Password Requirements**: Minimum 8 characters enforced
7. **Error Messages**: Generic "Invalid credentials" to prevent user enumeration
8. **SQL Injection**: Parameterized queries used throughout

## Implementation Notes

1. The existing `storage.Storage` interface will be extended to include `UserStorage` methods
2. The `DatabaseStorage` struct will implement both task and user storage
3. All existing task endpoints will be wrapped with auth middleware
4. Public endpoints (health, register, login) will not use auth middleware
5. The migration system will handle schema updates automatically
6. User ID will be stored in request context after authentication
7. Task queries will be filtered by user ID from context
8. Attempting to access another user's task will return 404 (not 403) to prevent information disclosure
