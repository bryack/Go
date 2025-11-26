# Implementation Plan

- [ ] 1. Add authentication dependencies and update go.mod
  - Add `github.com/golang-jwt/jwt/v5` package for JWT token handling
  - Add `golang.org/x/crypto` package for bcrypt password hashing
  - Run `go mod tidy` to update dependencies
  - _Requirements: 5.1, 5.3, 5.4_

- [ ] 2. Create database migration for users table
  - Add migration version 2 to create users table with id, email, password_hash, created_at columns
  - Create unique index on email column for fast lookups and constraint enforcement
  - Update `NewMigratorWithDefaults` to include the new users migration
  - _Requirements: 1.1, 5.1, 5.2_

- [ ] 3. Create database migration for task-user association
  - Add migration version 3 to add user_id column to tasks table
  - Create index on user_id column for efficient filtering
  - Update `NewMigratorWithDefaults` to include the task-user migration
  - _Requirements: 3.1, 3.2, 3.5_

- [ ] 4. Implement User storage model and interface
  - [ ] 4.1 Create User struct in storage package
    - Define User struct with ID, Email, PasswordHash, CreatedAt fields
    - Use JSON tags to prevent password_hash serialization
    - _Requirements: 5.1, 5.2_
  
  - [ ] 4.2 Define UserStorage interface
    - Define CreateUser, GetUserByEmail, GetUserByID, EmailExists methods
    - Add interface to storage package
    - _Requirements: 1.1, 2.1, 2.3_
  
  - [ ] 4.3 Implement UserStorage methods in DatabaseStorage
    - Implement CreateUser to insert new user with hashed password
    - Implement GetUserByEmail to retrieve user by email address
    - Implement GetUserByID to retrieve user by ID
    - Implement EmailExists to check for duplicate emails
    - Use parameterized queries to prevent SQL injection
    - _Requirements: 1.1, 1.2, 2.1, 2.3, 5.1_

- [ ] 5. Implement JWT service
  - [ ] 5.1 Create JWT Claims struct
    - Define Claims struct with UserID and embedded RegisteredClaims
    - _Requirements: 2.4, 4.4_
  
  - [ ] 5.2 Create JWTService struct and constructor
    - Define JWTService with secretKey and expiration fields
    - Implement NewJWTService to initialize from environment variables
    - Load JWT_SECRET_KEY from environment (required)
    - Load JWT_EXPIRATION_HOURS from environment with 24-hour default
    - _Requirements: 2.5, 5.3, 5.5_
  
  - [ ] 5.3 Implement token generation
    - Implement GenerateToken method to create JWT with user ID
    - Set expiration time based on configured duration
    - Sign token with secret key using HS256 algorithm
    - _Requirements: 2.1, 2.4, 2.5, 5.4_
  
  - [ ] 5.4 Implement token validation
    - Implement ValidateToken method to verify signature and expiration
    - Return Claims if valid, error if expired or invalid signature
    - Handle malformed tokens gracefully
    - _Requirements: 4.2, 4.3, 5.4_

- [ ] 6. Implement authentication service
  - [ ] 6.1 Create Service struct and constructor
    - Define Service with userStorage and jwtService dependencies
    - Implement NewService constructor
    - _Requirements: 1.1, 2.1_
  
  - [ ] 6.2 Implement password utilities
    - Implement HashPassword using bcrypt with cost factor 10
    - Implement ComparePassword to verify password against hash
    - Implement ValidatePassword to enforce minimum 8 character requirement
    - _Requirements: 1.4, 2.3, 5.1, 6.5_
  
  - [ ] 6.3 Implement user registration
    - Implement Register method to create new user accounts
    - Validate email format using standard email validation
    - Validate password meets requirements
    - Check for duplicate email using EmailExists
    - Hash password before storage
    - Generate and return JWT token upon successful registration
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_
  
  - [ ] 6.4 Implement user login
    - Implement Login method to authenticate users
    - Retrieve user by email
    - Compare provided password with stored hash
    - Return generic error for both invalid email and password
    - Generate and return JWT token upon successful login
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [ ] 7. Implement authentication middleware
  - [ ] 7.1 Create AuthMiddleware struct
    - Define AuthMiddleware with jwtService dependency
    - Implement NewAuthMiddleware constructor
    - Define ContextKey type for user ID storage
    - _Requirements: 4.1, 4.4_
  
  - [ ] 7.2 Implement token extraction
    - Implement ExtractToken to parse Authorization header
    - Validate Bearer scheme format
    - Return error for missing or malformed headers
    - _Requirements: 4.1, 4.4, 6.1_
  
  - [ ] 7.3 Implement authentication wrapper
    - Implement Authenticate method to wrap http.HandlerFunc
    - Extract token from request headers
    - Validate token using JWT service
    - Add user ID to request context on success
    - Return appropriate error responses for auth failures
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  
  - [ ] 7.4 Implement context helper
    - Implement GetUserIDFromContext to retrieve user ID from context
    - Return error if user ID not found in context
    - _Requirements: 3.4, 4.4_

- [ ] 8. Create authentication HTTP handlers
  - [ ] 8.1 Define request/response structs
    - Create RegisterRequest struct with Email and Password fields
    - Create LoginRequest struct with Email and Password fields
    - Create AuthResponse struct with Token and Email fields
    - _Requirements: 1.1, 2.1_
  
  - [ ] 8.2 Implement registration handler
    - Create RegisterHandler function accepting auth service
    - Parse and validate JSON request body
    - Call auth service Register method
    - Return token in AuthResponse format on success
    - Return appropriate error responses for validation failures
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 6.1, 6.2, 6.3, 6.5_
  
  - [ ] 8.3 Implement login handler
    - Create LoginHandler function accepting auth service
    - Parse and validate JSON request body
    - Call auth service Login method
    - Return token in AuthResponse format on success
    - Return generic error for invalid credentials
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 6.1, 6.2, 6.4_

- [x] 9. Update task storage to support user isolation
  - [x] 9.1 Modify CreateTask to accept user ID
    - Update CreateTask signature to include userID parameter
    - Insert user_id when creating new tasks
    - _Requirements: 3.1_
  
  - [x] 9.2 Modify LoadTasks to filter by user ID
    - Update LoadTasks signature to include userID parameter
    - Add WHERE clause to filter tasks by user_id
    - _Requirements: 3.2, 3.5_
  
  - [x] 9.3 Modify GetTaskByID to verify ownership
    - Update GetTaskByID signature to include userID parameter
    - Add WHERE clause to verify task belongs to user
    - Return ErrTaskNotFound if task doesn't exist or belongs to different user
    - _Requirements: 3.3, 3.4_
  
  - [x] 9.4 Modify UpdateTask to verify ownership
    - Update UpdateTask signature to include userID parameter
    - Add WHERE clause to verify task belongs to user before updating
    - Return ErrTaskNotFound if task doesn't exist or belongs to different user
    - _Requirements: 3.3, 3.4, 3.5_
  
  - [x] 9.5 Modify DeleteTask to verify ownership
    - Update DeleteTask signature to include userID parameter
    - Add WHERE clause to verify task belongs to user before deleting
    - Return ErrTaskNotFound if task doesn't exist or belongs to different user
    - _Requirements: 3.3, 3.4, 3.5_

- [x] 10. Update task HTTP handlers to use authentication
  - [x] 10.1 Modify tasksHandler for user isolation
    - Extract user ID from request context using GetUserIDFromContext
    - Pass user ID to CreateTask when creating new tasks
    - Pass user ID to LoadTasks when retrieving tasks
    - Handle context errors appropriately
    - _Requirements: 3.1, 3.2, 3.4_
  
  - [x] 10.2 Modify taskHandler for user isolation
    - Extract user ID from request context using GetUserIDFromContext
    - Pass user ID to GetTaskByID, UpdateTask, and DeleteTask operations
    - Return 404 for tasks that don't belong to user
    - Handle context errors appropriately
    - _Requirements: 3.3, 3.4, 3.5_

- [x] 11. Wire up authentication in main server
  - [x] 11.1 Initialize JWT service
    - Create JWTService instance with secret from environment
    - Handle missing JWT_SECRET_KEY with fatal error
    - _Requirements: 5.3, 5.5_
  
  - [x] 11.2 Initialize auth service
    - Create auth Service instance with storage and JWT service
    - _Requirements: 1.1, 2.1_
  
  - [x] 11.3 Initialize auth middleware
    - Create AuthMiddleware instance with JWT service
    - _Requirements: 4.1_
  
  - [x] 11.4 Register authentication endpoints
    - Register POST /auth/register with RegisterHandler (no auth required)
    - Register POST /auth/login with LoginHandler (no auth required)
    - Keep /health endpoint public (no auth required)
    - _Requirements: 1.1, 2.1, 4.5_
  
  - [x] 11.5 Protect task endpoints with middleware
    - Wrap /tasks handler with auth middleware
    - Wrap /tasks/ handler with auth middleware
    - Keep root handler public
    - _Requirements: 4.1, 4.5_
  
  - [x] 11.6 Update server startup logging
    - Add authentication endpoints to startup message
    - Document which endpoints require authentication
    - _Requirements: 4.5_

- [ ]* 12. Write tests for authentication system
  - [ ]* 12.1 Write JWT service tests
    - Test token generation with valid user ID
    - Test token validation with valid token
    - Test token validation with expired token
    - Test token validation with invalid signature
    - _Requirements: 2.1, 2.5, 4.2, 4.3_
  
  - [ ]* 12.2 Write auth service tests
    - Test user registration with valid data
    - Test registration with duplicate email
    - Test login with valid credentials
    - Test login with invalid credentials
    - Test password hashing and comparison
    - _Requirements: 1.1, 1.2, 2.1, 2.2, 5.1_
  
  - [ ]* 12.3 Write middleware tests
    - Test authentication with valid token
    - Test authentication with missing token
    - Test authentication with invalid token
    - Test user ID context propagation
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  
  - [ ]* 12.4 Write user storage tests
    - Test user creation
    - Test user retrieval by email
    - Test email existence check
    - Test unique email constraint
    - _Requirements: 1.1, 1.2, 2.1_
  
  - [ ]* 12.5 Write integration tests
    - Test complete registration flow
    - Test complete login flow
    - Test protected endpoint access with valid token
    - Test protected endpoint access without token
    - Test task isolation between users
    - _Requirements: 1.5, 2.5, 3.2, 3.3, 4.1_
