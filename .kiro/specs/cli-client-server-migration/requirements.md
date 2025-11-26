# Requirements Document

## Introduction

This specification defines the migration of the CLI application from direct database access to a client-server architecture. Currently, the CLI directly accesses the SQLite database using the same storage layer as the HTTP server. This tight coupling causes maintenance issues as the server evolves with multi-user features (JWT authentication, user isolation) that don't apply to the CLI context.

The migration will transform the CLI into an HTTP client that communicates with the existing server via REST API, eliminating direct database dependencies and establishing clear architectural boundaries between the two components.

## Glossary

- **CLI Application**: The command-line interface tool that allows users to manage tasks through terminal commands
- **HTTP Server**: The existing web server that provides REST API endpoints for task management with JWT authentication
- **API Client**: A new component within the CLI that handles HTTP communication with the server
- **Token Storage**: Local file-based storage mechanism for persisting JWT authentication tokens
- **Legacy CLI**: The current CLI implementation that directly accesses the database
- **Migration Path**: The transition strategy from direct database access to HTTP-based communication

## Requirements

### Requirement 1: CLI Authentication

**User Story:** As a CLI user, I want to authenticate once and have my session persist across commands, so that I don't need to provide credentials repeatedly.

#### Acceptance Criteria

1. WHEN the CLI Application starts without stored credentials, THE CLI Application SHALL prompt the user for email and password
2. WHEN the user provides valid credentials, THE CLI Application SHALL send an authentication request to the HTTP Server
3. WHEN the HTTP Server returns a JWT token, THE CLI Application SHALL store the token in Token Storage
4. WHEN the CLI Application starts with valid stored credentials, THE CLI Application SHALL load the token from Token Storage without prompting
5. WHEN the stored token is expired or invalid, THE CLI Application SHALL prompt the user to re-authenticate

### Requirement 2: HTTP Communication Layer

**User Story:** As a developer, I want the CLI to communicate with the server via HTTP, so that the CLI and server can evolve independently.

#### Acceptance Criteria

1. THE CLI Application SHALL implement an API Client component for HTTP communication
2. THE API Client SHALL include the JWT token in the Authorization header for all authenticated requests
3. WHEN the API Client receives a 401 Unauthorized response, THE API Client SHALL clear stored credentials and prompt for re-authentication
4. THE API Client SHALL handle network errors and provide user-friendly error messages
5. THE API Client SHALL support configurable server URL via environment variable or configuration file

### Requirement 3: Task Operations via API

**User Story:** As a CLI user, I want all task operations to work seamlessly through the API, so that my workflow remains unchanged.

#### Acceptance Criteria

1. WHEN the user executes the "add" command, THE CLI Application SHALL send a POST request to create a task via the HTTP Server
2. WHEN the user executes the "list" command, THE CLI Application SHALL send a GET request to retrieve tasks via the HTTP Server
3. WHEN the user executes the "status" command, THE CLI Application SHALL send a PATCH/PUT request to update task status via the HTTP Server
4. WHEN the user executes the "update" command, THE CLI Application SHALL send a PATCH/PUT request to update task description via the HTTP Server
5. WHEN the user executes the "delete" command, THE CLI Application SHALL send a DELETE request to remove a task via the HTTP Server
6. WHEN the user executes the "clear" command, THE CLI Application SHALL send a PATCH/PUT request to clear task description via the HTTP Server

### Requirement 4: Error Handling and User Experience

**User Story:** As a CLI user, I want clear error messages when something goes wrong, so that I can understand and resolve issues quickly.

#### Acceptance Criteria

1. WHEN the HTTP Server is unreachable, THE CLI Application SHALL display a message indicating connection failure with the server URL
2. WHEN the API Client receives a 4xx error response, THE CLI Application SHALL display the error message from the server response
3. WHEN the API Client receives a 5xx error response, THE CLI Application SHALL display a message indicating a server error
4. WHEN network timeout occurs, THE CLI Application SHALL display a timeout message with retry suggestion
5. THE CLI Application SHALL maintain the same user interface and command structure as the Legacy CLI

### Requirement 5: Configuration Management

**User Story:** As a CLI user, I want to configure the server URL, so that I can connect to different environments (local, staging, production).

#### Acceptance Criteria

1. THE CLI Application SHALL read server URL from environment variable TASK_SERVER_URL
2. WHERE environment variable is not set, THE CLI Application SHALL use default value "http://localhost:8080"
3. THE CLI Application SHALL validate that the server URL is a valid HTTP/HTTPS URL
4. THE CLI Application SHALL display the configured server URL during startup
5. THE CLI Application SHALL allow server URL override via command-line flag

### Requirement 6: Token Storage Security

**User Story:** As a CLI user, I want my authentication token stored securely, so that unauthorized users cannot access my tasks.

#### Acceptance Criteria

1. THE CLI Application SHALL store the JWT token in a file with restricted permissions (0600 on Unix systems)
2. THE Token Storage file SHALL be located in the user's home directory under .task-cli/token
3. WHEN storing the token, THE CLI Application SHALL create parent directories if they do not exist
4. WHEN reading the token, THE CLI Application SHALL verify file permissions and warn if permissions are too permissive
5. THE CLI Application SHALL provide a "logout" command that deletes the stored token

### Requirement 7: Backward Compatibility During Migration

**User Story:** As a developer, I want a clear migration path from the old CLI to the new CLI, so that users can transition smoothly.

#### Acceptance Criteria

1. THE CLI Application SHALL detect if it is running in legacy mode (direct database access) or client mode (HTTP API)
2. WHERE legacy mode is detected, THE CLI Application SHALL display a deprecation warning
3. THE CLI Application SHALL provide a command-line flag to force client mode
4. THE CLI Application SHALL maintain the same command interface in both modes
5. THE CLI Application SHALL provide migration documentation for users

### Requirement 8: Registration Support

**User Story:** As a new CLI user, I want to register an account from the CLI, so that I can start using the application without accessing a web interface.

#### Acceptance Criteria

1. THE CLI Application SHALL provide a "register" command for new user registration
2. WHEN the user executes the "register" command, THE CLI Application SHALL prompt for email and password
3. WHEN registration is successful, THE CLI Application SHALL automatically store the returned JWT token
4. WHEN registration fails due to existing email, THE CLI Application SHALL display an appropriate error message
5. THE CLI Application SHALL validate email format and password requirements before sending registration request
