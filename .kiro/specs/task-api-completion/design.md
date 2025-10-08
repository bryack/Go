# Design Document

## Overview

This design extends the existing Task API to complete the CRUD operations by adding GET /tasks/{id}, PUT /tasks/{id}, and DELETE /tasks/{id} endpoints. The design maintains consistency with the current architecture while introducing URL path parameter extraction and individual task operations.

## Architecture

### Current Architecture Analysis
The existing system follows a clean layered architecture:
- **HTTP Layer**: `cmd/server/main.go` handles HTTP routing and request/response
- **Business Logic**: `task/operations.go` contains the TaskManager with thread-safe operations
- **Storage Layer**: `storage/json.go` provides persistent storage interface
- **Utilities**: `internal/handlers/helpers.go` provides consistent JSON response handling

### Proposed Extensions
The design will extend this architecture by:
1. Adding URL path parameter extraction capability
2. Implementing individual task handlers that reuse existing TaskManager methods
3. Maintaining the same error handling and response patterns

## Components and Interfaces

### 1. URL Router Enhancement

**Current State**: Simple pattern matching with `http.HandleFunc`
**Enhancement Needed**: Path parameter extraction for `/tasks/{id}` patterns

**Design Decision**: Implement a custom router function that:
- Parses URL paths to extract task IDs
- Validates ID format (integer)
- Routes to appropriate handler based on HTTP method
- Maintains compatibility with existing `/tasks` endpoint

```go
// Proposed function signature
func taskRouterHandler(tm *TaskManager) http.HandlerFunc
```

### 2. Individual Task Operations Handler

**Purpose**: Handle operations on specific tasks (GET, PUT, DELETE /tasks/{id})

**Design Approach**:
- Extract task ID from URL path
- Validate ID format and convert to integer
- Route to appropriate operation based on HTTP method
- Reuse existing TaskManager methods where possible

**Key Methods to Leverage**:
- `tm.GetTaskByID(id)` - Already implemented
- `tm.UpdateTaskDescription(id, description)` - Already implemented  
- New method needed: `tm.DeleteTask(id)` - To be added to TaskManager

### 3. Request/Response Models

**Update Task Request Model**:
```go
type UpdateTaskRequest struct {
    Description string `json:"description"`
    Done        bool   `json:"done"`
}
```

**Design Rationale**: 
- Allows updating both description and completion status
- Maintains consistency with existing Task struct
- Enables partial updates (validation will handle required fields)

### 4. TaskManager Extensions

**New Method Required**: `DeleteTask(id int) error`
- Thread-safe task removal
- Returns `ErrTaskNotFound` for consistency
- Maintains existing error handling patterns

**Potential Enhancement**: `UpdateTask(id int, updates UpdateTaskRequest) error`
- More comprehensive update method
- Handles both description and done status updates
- Alternative to separate update methods

## Data Models

### Existing Models (No Changes)
```go
type Task struct {
    ID          int    `json:"id"`
    Description string `json:"description"`
    Done        bool   `json:"done"`
}
```

### New Request Models
```go
type UpdateTaskRequest struct {
    Description string `json:"description"`
    Done        bool   `json:"done"`
}
```

### URL Path Structure
- `/tasks` - Collection operations (existing: GET, POST)
- `/tasks/{id}` - Individual task operations (new: GET, PUT, DELETE)
- ID parameter: Must be positive integer

## Error Handling

### HTTP Status Code Mapping
- **200 OK**: Successful GET, PUT operations
- **204 No Content**: Successful DELETE operations  
- **400 Bad Request**: Invalid ID format, validation errors
- **404 Not Found**: Task not found, invalid routes
- **405 Method Not Allowed**: Unsupported HTTP methods
- **415 Unsupported Media Type**: Missing/incorrect Content-Type

### Error Response Consistency
All errors will use the existing `handlers.JSONError` function to maintain consistent error response format:
```json
{
    "error": "descriptive error message"
}
```

### Validation Rules
1. **ID Validation**: Must be positive integer
2. **Description Validation**: Max 200 characters (consistent with POST)
3. **Content-Type Validation**: Must be `application/json` for PUT requests
4. **JSON Format Validation**: Must be valid JSON structure

## Testing Strategy

### Unit Testing Approach
1. **TaskManager Method Tests**: Test new DeleteTask method
2. **Handler Function Tests**: Test URL parsing and routing logic
3. **Integration Tests**: Test complete request/response cycles
4. **Error Handling Tests**: Verify all error conditions return correct status codes

### Test Cases by Endpoint

**GET /tasks/{id}**:
- Valid ID returns task
- Invalid ID format returns 400
- Non-existent ID returns 404
- Wrong HTTP method returns 405

**PUT /tasks/{id}**:
- Valid update returns updated task
- Invalid JSON returns 400
- Missing Content-Type returns 415
- Non-existent ID returns 404
- Validation failures return 400
- Wrong HTTP method returns 405

**DELETE /tasks/{id}**:
- Valid ID deletes task and returns 204
- Invalid ID format returns 400
- Non-existent ID returns 404
- Wrong HTTP method returns 405

### Testing Tools
- Go's built-in `testing` package
- `httptest` package for HTTP handler testing
- Table-driven tests for comprehensive coverage

## Implementation Considerations

### Thread Safety
- All operations will use existing TaskManager mutex protection
- No additional synchronization needed due to existing design

### Performance
- ID extraction and validation are O(1) operations
- Task lookup remains O(n) as in existing implementation
- No performance regression expected

### Backward Compatibility
- Existing `/tasks` endpoint behavior unchanged
- All existing functionality preserved
- New endpoints are additive only

### Code Organization
- New functionality added to existing files where appropriate
- Maintain current package structure
- Follow existing naming conventions and code style