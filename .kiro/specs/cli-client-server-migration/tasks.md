# Implementation Plan

- [x] 1. Create configuration management
  - Create `cmd/cli/config.go` with Config struct and LoadConfig function
  - Implement environment variable reading for TASK_SERVER_URL with default "http://localhost:8080"
  - Implement URL validation to ensure valid HTTP/HTTPS format
  - _Requirements: 5.1, 5.2, 5.3_

- [x] 2. Create HTTP API client package
  - Create `cmd/cli/client/client.go` with TaskClient interface and HTTPClient implementation
  - Implement NewHTTPClient constructor with configurable base URL and 30-second timeout
  - _Requirements: 2.1, 2.5_

- [x] 2.1 Implement core HTTP request handling
  - Implement doRequest method for generic HTTP requests with JSON encoding/decoding
  - Add Authorization header injection with Bearer token format
  - Implement handleErrorResponse for parsing server error responses
  - _Requirements: 2.1, 2.2, 2.4_

- [x] 2.2 Implement authentication API methods
  - Implement Login method that sends POST to /login endpoint
  - Implement Register method that sends POST to /register endpoint
  - Parse AuthResponse and extract JWT token from response
  - _Requirements: 1.2, 8.2, 8.3_

- [x] 2.3 Implement task operation API methods
  - Implement GetTasks method that sends GET to /tasks endpoint
  - Implement GetTask method that sends GET to /tasks/{id} endpoint
  - Implement CreateTask method that sends POST to /tasks endpoint
  - Implement UpdateTask method that sends PUT to /tasks/{id} endpoint with optional fields
  - Implement DeleteTask method that sends DELETE to /tasks/{id} endpoint
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6_

- [x] 2.4 Implement error handling for network and HTTP errors
  - Create NetworkError type for connection failures with server URL
  - Create APIError type for HTTP error responses with status code and message
  - Map network errors to user-friendly messages
  - Map 4xx errors to display server error message
  - Map 5xx errors to generic server error message
  - Handle timeout errors with retry suggestion
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 3. Create authentication manager package
  - Create `cmd/cli/auth/auth.go` with AuthManager interface and FileAuthManager implementation
  - Implement NewFileAuthManager constructor with token path in ~/.task-cli/token
  - _Requirements: 1.1, 6.2_

- [x] 3.1 Implement token storage operations
  - Implement SaveToken method that writes token to file with 0600 permissions
  - Create parent directories with 0700 permissions if they don't exist
  - Implement LoadToken method that reads token from file
  - Verify file permissions on load and warn if too permissive
  - Implement ClearToken method that deletes token file
  - _Requirements: 1.3, 1.4, 6.1, 6.3, 6.4, 6.5_

- [x] 3.2 Implement interactive authentication flows
  - Implement RequireAuth method that loads token or prompts for authentication
  - Implement PromptLogin method that prompts for email/password and calls client.Login
  - Implement PromptRegister method that prompts for email/password and calls client.Register
  - Save token automatically after successful login or registration
  - Handle 401 responses by clearing token and re-prompting for authentication
  - _Requirements: 1.1, 1.2, 1.5, 2.3, 8.1, 8.2, 8.3_

- [x] 3.3 Add password input masking
  - Implement secure password input that masks characters during entry
  - Use golang.org/x/term package for terminal password reading
  - _Requirements: 1.1, 8.2_

- [x] 4. Update CLI main entry point
  - Modify `cmd/cli/main.go` to load configuration using config.LoadConfig
  - Create HTTP client with configured server URL
  - Create auth manager and perform initial authentication
  - Display server URL during startup
  - Pass client and auth manager to CLI constructor
  - _Requirements: 1.1, 1.4, 5.1, 5.2, 5.4_

- [x] 5. Update CLI struct and initialization
  - Modify CLI struct in `cmd/cli/cli.go` to replace taskManager and storage with client and authManager
  - Update NewCLI constructor to accept config, client, and authManager
  - Remove database storage initialization from CLI
  - _Requirements: 2.1, 4.5_

- [x] 6. Update command handlers to use API client
  - Update handleAddCommand to call client.CreateTask instead of taskManager.AddTask
  - Update handleListCommand to call client.GetTasks and format output
  - Update handleStatusCommand to call client.UpdateTask with done parameter
  - Update handleUpdateCommand to call client.UpdateTask with description parameter
  - Update handleDeleteCommand to call client.DeleteTask
  - Update handleClearCommand to call client.UpdateTask with empty description
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 4.5_

- [x] 7. Add new authentication commands
  - Add "login" command that calls authManager.PromptLogin
  - Add "register" command that calls authManager.PromptRegister
  - Add "logout" command that calls authManager.ClearToken and displays confirmation
  - Update command validation to recognize new commands
  - Update help text to include new commands
  - _Requirements: 6.5, 8.1_

- [x] 8. Implement 401 re-authentication handling
  - Detect 401 Unauthorized responses in API client error handling
  - Clear stored token when 401 is received
  - Return specific error type that triggers re-authentication prompt
  - Update CLI command loop to catch auth errors and prompt for re-authentication
  - _Requirements: 1.5, 2.3_

- [x] 9. Add error handling to CLI command handlers
  - Update all command handlers to display user-friendly error messages from API client
  - Handle NetworkError with connection failure message including server URL
  - Handle APIError with server error message
  - Maintain existing error display format for consistency
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_

- [x] 10. Update CLI startup flow
  - Display "Task Manager CLI (Client Mode)" banner
  - Display configured server URL
  - Show authentication prompt if no token exists
  - Provide options: 1) Login 2) Register 3) Exit
  - Handle user choice and proceed to command loop after successful authentication
  - _Requirements: 1.1, 5.4_

- [x] 11. Remove direct database dependencies from CLI
  - Remove storage package imports from cmd/cli files
  - Remove task package imports from cmd/cli files
  - Remove database initialization code
  - Update build to exclude storage dependencies for CLI binary
  - _Requirements: 2.1_

- [x] 12. Add input validation before API calls
  - Validate email format in register command before sending request
  - Validate password requirements (8-72 chars) in register command before sending request
  - Keep existing task description validation in command handlers
  - Display validation errors without making API call
  - _Requirements: 8.4, 8.5_
