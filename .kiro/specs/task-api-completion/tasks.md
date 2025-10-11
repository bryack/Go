# Implementation Plan

- [x] 1. Add DeleteTask method to TaskManager
  - Implement thread-safe task deletion in task/operations.go
  - Return ErrTaskNotFound for non-existent tasks
  - Follow existing method patterns and error handling
  - _Requirements: 3.1, 3.2_

- [x] 1.1 Write unit tests for DeleteTask method
  - Test successful deletion returns no error
  - Test deletion of non-existent task returns ErrTaskNotFound
  - Test thread safety with concurrent operations
  - _Requirements: 3.1, 3.2_

- [x] 2. Create URL path parameter extraction utility
  - Implement function to extract task ID from /tasks/{id} URL paths
  - Add integer validation for extracted ID parameter
  - Handle malformed URLs and invalid ID formats
  - _Requirements: 4.1, 4.2, 4.3_

- [ ]* 2.1 Write unit tests for URL parameter extraction
  - Test valid ID extraction from various URL formats
  - Test invalid ID format handling
  - Test malformed URL handling
  - _Requirements: 4.1, 4.2_

- [x] 3. Implement GET /tasks/{id} endpoint
  - Create handler function for individual task retrieval
  - Integrate with existing GetTaskByID method
  - Return appropriate HTTP status codes (200, 400, 404)
  - Use existing JSONSuccess and JSONError helpers
  - _Requirements: 1.1, 1.2, 1.3, 5.1, 5.3_

- [ ]* 3.1 Write tests for GET /tasks/{id} endpoint
  - Test successful task retrieval with valid ID
  - Test 404 response for non-existent task
  - Test 400 response for invalid ID format
  - Test 405 response for wrong HTTP method
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [x] 4. Create UpdateTaskRequest model and validation
  - Define struct for PUT request payload with Description and Done fields
  - Implement validation function for update requests
  - Add description length validation (max 200 characters)
  - Handle required field validation
  - _Requirements: 2.3, 2.5, 6.1, 6.2, 6.3_

- [x] 5. Implement PUT /tasks/{id} endpoint
  - Create handler function for task updates
  - Parse and validate JSON request body
  - Integrate with existing UpdateTaskDescription and MarkTaskDone methods
  - Return updated task with 200 status code
  - Handle all error cases (400, 404, 415)
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 5.1, 5.2_

- [ ]* 5.1 Write tests for PUT /tasks/{id} endpoint
  - Test successful task update with valid payload
  - Test 404 response for non-existent task
  - Test 400 response for invalid JSON and validation errors
  - Test 415 response for missing Content-Type header
  - Test 405 response for wrong HTTP method
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

- [x] 6. Implement DELETE /tasks/{id} endpoint
  - Create handler function for task deletion
  - Integrate with new DeleteTask method
  - Return 204 No Content status for successful deletion
  - Handle error cases (400, 404)
  - _Requirements: 3.1, 3.2, 3.3, 5.1, 5.3_

- [ ]* 6.1 Write tests for DELETE /tasks/{id} endpoint
  - Test successful task deletion returns 204
  - Test 404 response for non-existent task
  - Test 400 response for invalid ID format
  - Test 405 response for wrong HTTP method
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [ ] 7. Create unified task router handler
  - Implement router function that handles both /tasks and /tasks/{id} patterns
  - Route collection operations to existing tasksHandler
  - Route individual operations to new individual task handlers
  - Maintain method validation for all endpoints
  - _Requirements: 4.1, 4.4, 1.4, 2.6, 3.4_

- [ ] 8. Update main.go to use new routing system
  - Replace existing /tasks handler registration with new router
  - Ensure backward compatibility with existing endpoints
  - Update endpoint documentation in startup messages
  - Test that all existing functionality still works
  - _Requirements: 4.1, 4.4_

- [ ]* 8.1 Write integration tests for complete API
  - Test full CRUD workflow (create, read, update, delete)
  - Test error handling across all endpoints
  - Test method validation for all routes
  - Verify response format consistency
  - _Requirements: 5.1, 5.2, 5.3, 5.4_