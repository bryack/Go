# Requirements Document

## Introduction

This specification covers the completion of the Task API for a Go-based task management system. The system currently has basic task listing (GET /tasks) and task creation (POST /tasks) functionality. We need to implement the remaining CRUD operations to provide a complete REST API that follows Go best practices and maintains consistency with the existing codebase.

## Requirements

### Requirement 1: Single Task Retrieval

**User Story:** As an API client, I want to retrieve a specific task by its ID, so that I can get detailed information about an individual task.

#### Acceptance Criteria

1. WHEN a client sends a GET request to /tasks/{id} with a valid task ID THEN the system SHALL return the task with HTTP status 200 OK
2. WHEN a client sends a GET request to /tasks/{id} with a non-existent task ID THEN the system SHALL return HTTP status 404 Not Found with appropriate error message
3. WHEN a client sends a GET request to /tasks/{id} with an invalid ID format THEN the system SHALL return HTTP status 400 Bad Request with validation error message
4. WHEN a client sends a non-GET request to /tasks/{id} THEN the system SHALL return HTTP status 405 Method Not Allowed

### Requirement 2: Task Update

**User Story:** As an API client, I want to update an existing task's properties, so that I can modify task details without recreating the task.

#### Acceptance Criteria

1. WHEN a client sends a PUT request to /tasks/{id} with valid JSON payload THEN the system SHALL update the task and return the updated task with HTTP status 200 OK
2. WHEN a client sends a PUT request to /tasks/{id} with a non-existent task ID THEN the system SHALL return HTTP status 404 Not Found
3. WHEN a client sends a PUT request to /tasks/{id} with invalid JSON format THEN the system SHALL return HTTP status 400 Bad Request
4. WHEN a client sends a PUT request to /tasks/{id} without Content-Type application/json THEN the system SHALL return HTTP status 415 Unsupported Media Type
5. WHEN a client sends a PUT request to /tasks/{id} with invalid field values THEN the system SHALL return HTTP status 400 Bad Request with validation errors
6. WHEN a client sends a non-PUT request to /tasks/{id} THEN the system SHALL return HTTP status 405 Method Not Allowed

### Requirement 3: Task Deletion

**User Story:** As an API client, I want to delete a specific task by its ID, so that I can remove tasks that are no longer needed.

#### Acceptance Criteria

1. WHEN a client sends a DELETE request to /tasks/{id} with a valid task ID THEN the system SHALL remove the task and return HTTP status 204 No Content
2. WHEN a client sends a DELETE request to /tasks/{id} with a non-existent task ID THEN the system SHALL return HTTP status 404 Not Found
3. WHEN a client sends a DELETE request to /tasks/{id} with an invalid ID format THEN the system SHALL return HTTP status 400 Bad Request
4. WHEN a client sends a non-DELETE request to /tasks/{id} THEN the system SHALL return HTTP status 405 Method Not Allowed

### Requirement 4: URL Routing and Path Parameter Handling

**User Story:** As a developer, I want proper URL routing that can extract task IDs from the URL path, so that the API can handle individual task operations correctly.

#### Acceptance Criteria

1. WHEN the system receives a request to /tasks/{id} THEN it SHALL extract the ID parameter from the URL path
2. WHEN the extracted ID is not a valid integer THEN the system SHALL return HTTP status 400 Bad Request
3. WHEN the extracted ID is a valid integer THEN the system SHALL pass it to the appropriate handler function
4. WHEN the URL path does not match expected patterns THEN the system SHALL return HTTP status 404 Not Found

### Requirement 5: Error Handling Consistency

**User Story:** As an API client, I want consistent error response formats across all endpoints, so that I can handle errors predictably in my client code.

#### Acceptance Criteria

1. WHEN any endpoint returns an error THEN the response SHALL use the existing JSONError format
2. WHEN validation fails THEN the error message SHALL clearly indicate what validation rule was violated
3. WHEN a resource is not found THEN the error message SHALL indicate which resource was not found
4. WHEN an internal server error occurs THEN the system SHALL return HTTP status 500 with a generic error message

### Requirement 6: Request Validation

**User Story:** As a system administrator, I want robust input validation on all endpoints, so that the system remains stable and secure.

#### Acceptance Criteria

1. WHEN updating a task THEN the description field SHALL be validated for length (max 200 characters)
2. WHEN updating a task THEN required fields SHALL be validated for presence
3. WHEN updating a task THEN the done field SHALL accept only boolean values
4. WHEN any validation fails THEN the system SHALL return specific validation error messages