# Task Decomposition Standards

This document defines the requirements for creating detailed, step-by-step task decompositions when specifically requested by the user.

## Purpose

While `tasks.md` in specs provides high-level implementation tasks suitable for autonomous execution, sometimes a user needs a more granular breakdown to understand and implement a task manually. This document defines how to create such decompositions.

## When to Use

- ✅ **Use**: Only when the user explicitly requests a detailed breakdown of a specific task
- ✅ **Use**: When the user says they don't fully understand how to implement a task
- ✅ **Use**: When the user wants to learn the implementation approach
- ❌ **Don't use**: For regular spec creation (use standard `tasks.md` format)
- ❌ **Don't use**: When executing tasks autonomously

## Document Structure

### 1. Task Overview

**Rule**: Start with a clear summary of what the task accomplishes.

```markdown
# Task Decomposition: [Task Name]

## Overview
[2-3 sentences describing what this task achieves and why it's needed]
```

### 2. Implementation Approach

**Rule**: Explain the chosen approach and rationale in 3-5 sentences.

- ✅ **Required**: Brief explanation of the implementation strategy
- ✅ **Required**: Why this approach was chosen over alternatives
- ✅ **Required**: Key concepts or patterns being used
- ❌ **Avoid**: Lengthy theoretical explanations
- **Rationale**: User needs context to understand the "why" before the "how"

```markdown
## Implementation Approach

We're using [pattern/strategy] because [reason]. This approach [key benefit]. 
The implementation follows [principle/pattern] which [explanation].

**Key Concepts:**
- [Concept 1]: Brief explanation
- [Concept 2]: Brief explanation
```

### 3. Avoid Overengineering

**Rule**: Always check if the proposed implementation is the simplest solution that meets requirements.

- ✅ **Required**: Review the actual requirements before proposing complex solutions
- ✅ **Required**: Offer simpler alternatives when appropriate
- ✅ **Required**: Explain trade-offs between simple and complex approaches
- ✅ **Required**: Default to the simpler approach unless complexity is justified
- ❌ **Avoid**: Implementing features not required by the spec
- ❌ **Avoid**: Adding abstractions "for future flexibility" without current need
- ❌ **Avoid**: Complex patterns when simple code would suffice
- **Rationale**: YAGNI (You Aren't Gonna Need It) - simpler code is easier to understand, maintain, and debug

```markdown
## Implementation Approach

We're using [pattern/strategy] because [reason].

**Complexity Check:**
- **Requirements need**: [What the spec actually requires]
- **Simple approach**: [Minimal implementation that meets requirements]
- **Complex approach**: [More sophisticated implementation]
- **Recommendation**: [Which to use and why]

**Example:**
- **Requirements need**: Log HTTP status code and duration
- **Simple approach**: Log completion with duration (10 min)
- **Complex approach**: Response writer wrapper to capture status codes (30 min)
- **Recommendation**: Start simple. Requirements don't mandate accurate status codes. Add wrapper later only if needed for debugging.
```

**When to Question Complexity:**
- Wrappers or decorators that only add one field
- Interfaces with single implementations
- Abstractions for "future extensibility"
- Custom types when built-in types suffice
- Middleware that could be a simple function call

**Red Flags:**
- "We might need this later"
- "It's more flexible this way"
- "This is the proper pattern"
- "Everyone does it this way"

**Green Flags:**
- "The requirements explicitly need this"
- "This prevents a real, current problem"
- "This simplifies the code"
- "This is already a proven pattern in our codebase"

### 4. Logging Output Standards

**Rule**: Always use stderr for application logs, stdout for program output.

- ✅ **Required**: Configure logging to stderr (not stdout)
- ✅ **Required**: Use stdout only for actual program output/data
- ❌ **Avoid**: Mixing logs with program output on stdout
- **Rationale**: Follows Unix philosophy, prevents log buffering issues, container-friendly

**Unix Philosophy:**
- **stdout** = Program output (data, results, actual output)
- **stderr** = Diagnostic messages (logs, errors, warnings)

**Why stderr for logs:**
- Unbuffered by default (no lost logs on shutdown)
- Industry standard (Docker, Kubernetes expect logs on stderr)
- Separation of concerns (logs separate from data)
- Better for pipelines and redirection
- Prevents buffering issues with `os.Exit()`

**Configuration example:**
```yaml
logging:
  output: "stderr"  # Use stderr for logs
  level: "info"
  format: "json"
```

**When to use stdout:**
- CLI tools outputting data (e.g., `mytool list` outputs task list)
- API responses or structured data output
- Any actual program results (not diagnostic info)

### 5. Prerequisites

**Rule**: List what needs to exist before starting this task.

- ✅ **Required**: Files that must already exist
- ✅ **Required**: Dependencies that must be installed
- ✅ **Required**: Configuration that must be in place
- ✅ **Required**: Knowledge or understanding required

```markdown
## Prerequisites

**Existing Code:**
- `path/to/file.go` - [what it provides]
- `path/to/other.go` - [what it provides]

**Dependencies:**
- Package X installed (`go get ...`)

**Knowledge Required:**
- Understanding of [concept]
- Familiarity with [pattern]
```

### 5. Step-by-Step Instructions

**Rule**: Break down the task into concrete, actionable steps describing what to implement.

- ✅ **Required**: Numbered steps in logical order
- ✅ **Required**: Specific file paths and locations
- ✅ **Required**: Clear description of what needs to be implemented
- ✅ **Required**: Explanation of what each step does
- ✅ **Required**: Expected outcome after each step
- ❌ **Avoid**: Vague instructions like "update the handler"
- ❌ **Avoid**: Specific code examples or implementations (user wants to code themselves)
- ❌ **Avoid**: Showing exact function signatures or code structure
- **Rationale**: User wants to understand what to do and why, but implement the code themselves

```markdown
## Step-by-Step Instructions

### Step 1: [Action to take]

**File**: `path/to/file.go`

**What to do:**
[Clear description of what needs to be implemented, without showing the actual code]

**What to implement:**
- [Specific requirement 1]
- [Specific requirement 2]
- [Specific requirement 3]

**Why:**
[1-2 sentences explaining the purpose of this step]

**Expected result:**
[What should work/exist after this step]

---

### Step 2: [Next action]

[Same structure as Step 1]
```

### 6. Verification Steps

**Rule**: Provide concrete ways to verify the implementation works.

- ✅ **Required**: Commands to run for verification
- ✅ **Required**: Expected output or behavior
- ✅ **Required**: How to test the changes
- **Rationale**: User needs confidence their implementation is correct

```markdown
## Verification

### Compile Check
```bash
go build ./...
```
**Expected**: No compilation errors

### Run Tests
```bash
go test ./path/to/package -v
```
**Expected**: All tests pass

### Manual Testing
1. Run the application: `go run cmd/server/main.go`
2. Test the endpoint: `curl http://localhost:8080/endpoint`
3. **Expected response**: `{"status": "ok"}`
```

### 7. Common Pitfalls

**Rule**: Warn about common mistakes and how to avoid them.

- ✅ **Required**: List 1-2 common mistakes
- ✅ **Required**: How to recognize each mistake
- ✅ **Required**: How to fix each mistake

```markdown
## Common Pitfalls

### Pitfall 1: [Mistake description]
**Symptom**: [How you'll know this happened]
**Fix**: [How to correct it]

### Pitfall 2: [Another mistake]
**Symptom**: [Error message or behavior]
**Fix**: [Solution]
```

### 8. Learning Resources

**Rule**: Provide 1-2 relevant resources for deeper understanding.

- ✅ **Required**: Links to official documentation
- ✅ **Preferred**: Practical guides or tutorials
- ✅ **Optional**: Theoretical articles or papers
- ❌ **Avoid**: More than 3 resources (overwhelming)

```markdown
## Learning Resources

### Essential Reading
- [Resource Title](URL) - [1 sentence about what it covers]
- [Another Resource](URL) - [Why it's relevant]

### Additional Resources (Optional)
- [Deep Dive Article](URL) - [For advanced understanding]
```

## Formatting Rules

### Implementation Descriptions

**Rule**: Describe what to implement without showing the actual code.

- ✅ **Required**: Clear description of functionality to implement
- ✅ **Required**: List specific requirements and behaviors
- ✅ **Required**: Mention necessary imports or dependencies
- ✅ **Required**: Describe the logic flow without code
- ❌ **Avoid**: Actual code implementations
- ❌ **Avoid**: Specific function signatures or method bodies
- ❌ **Avoid**: Line-by-line code examples

```markdown
// ✅ Good: Description without code
**What to implement:**
- Create a handler function that accepts a Storage parameter and returns an http.HandlerFunc
- Extract the task ID from the URL parameters
- Call the storage layer to fetch the task by ID
- Handle errors by returning 404 if task not found
- Return the task as JSON on success
- Required imports: net/http, encoding/json

// ❌ Bad: Showing actual code
```go
func GetTaskHandler(s storage.Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        id := extractID(r)
        task, err := s.GetTaskByID(id)
        // ... etc
    }
}
```
```

### File Modifications

**Rule**: Clearly indicate whether to create new files or modify existing ones.

```markdown
// ✅ Good: Clear indication
**File**: `handlers/tasks.go` (create new file)
**File**: `main.go` (modify existing - add to imports section)

// ❌ Bad: Unclear
**File**: `handlers/tasks.go`
```

### Step Granularity

**Rule**: Each step should be completable in 5-10 minutes.

- ✅ **Good step**: "Add the GetTaskByID method to DatabaseStorage that queries by ID and userID"
- ❌ **Too broad**: "Implement the storage layer"
- ❌ **Too granular**: "Type the word 'func' on line 42"
- ❌ **Too detailed**: "Write this exact function with these parameters..."

## Example Decomposition

Here's a minimal example showing the structure (note: no actual code implementations shown):

```markdown
# Task Decomposition: Implement Task Deletion Endpoint

## Overview
Add a DELETE /tasks/:id endpoint that allows authenticated users to delete their own tasks. This completes the CRUD operations for the task management API.

## Implementation Approach

We're implementing a RESTful DELETE endpoint following the existing handler pattern. The approach uses:
1. HTTP method routing to handle DELETE requests
2. URL parameter extraction for task ID
3. Storage layer for database operations
4. JWT middleware for authentication

This follows REST conventions where DELETE is idempotent and returns 204 No Content on success.

**Key Concepts:**
- **Idempotency**: Multiple DELETE requests have the same effect as one
- **HTTP 204**: Success with no response body
- **User ownership**: Only the task owner can delete their task

## Prerequisites

**Existing Code:**
- `storage/database.go` - Database storage implementation
- `auth/middleware.go` - JWT authentication middleware
- `internal/handlers/handlers.go` - Existing handler patterns

**Dependencies:**
- `github.com/gorilla/mux` for routing (already installed)

**Knowledge Required:**
- Basic HTTP methods and status codes
- Go http.HandlerFunc pattern
- SQL DELETE statements

## Step-by-Step Instructions

### Step 1: Add DeleteTask method to Storage interface

**File**: `storage/database.go`

**What to do:**
Add a new method to the Storage interface and implement it in DatabaseStorage.

**What to implement:**
- Add a DeleteTask method to the Storage interface that accepts task ID and user ID
- Implement the method in DatabaseStorage to execute a SQL DELETE statement
- Include both id and user_id in the WHERE clause to ensure ownership
- Check RowsAffected to determine if the task existed
- Return ErrTaskNotFound if no rows were affected
- Use mapSQLiteError for other database errors

**Why:**
The storage layer needs a method to delete tasks. We include userID to ensure users can only delete their own tasks. Checking RowsAffected lets us return 404 if the task doesn't exist.

**Expected result:**
Code compiles. Storage interface now includes DeleteTask method.

---

### Step 2: Create DELETE handler

**File**: `internal/handlers/handlers.go`

**What to do:**
Add a new handler function for DELETE requests.

**What to implement:**
- Create a DeleteTaskHandler function that accepts Storage and returns http.HandlerFunc
- Extract the user ID from the request context using auth.GetUserIDFromContext
- Return 401 Unauthorized if user ID extraction fails
- Extract the task ID from URL parameters using mux.Vars
- Convert the task ID to an integer and return 400 Bad Request if invalid
- Call storage.DeleteTask with both task ID and user ID
- Handle ErrTaskNotFound by returning 404 Not Found
- Handle other errors by returning 500 Internal Server Error
- Return 204 No Content on successful deletion
- Required imports: net/http, strconv, errors

**Why:**
This handler follows the established pattern: authenticate, validate input, call storage, handle errors. We return 204 (No Content) on success per REST conventions for DELETE.

**Expected result:**
Handler function exists and compiles.

---

### Step 3: Register route in main.go

**File**: `cmd/server/main.go`

**What to do:**
Add the DELETE route to the router with authentication middleware.

**What to implement:**
- Register a new route for DELETE requests to /tasks/{id}
- Wrap the DeleteTaskHandler with authMiddleware.Authenticate
- Pass the storage instance to the handler
- Use the Methods("DELETE") constraint to only match DELETE requests

**Why:**
This registers the handler for DELETE /tasks/:id and applies authentication middleware to protect the endpoint.

**Expected result:**
Server compiles and starts successfully.

## Verification

### Compile Check
```bash
go build ./...
```
**Expected**: No compilation errors

### Run Tests
```bash
# Test the storage layer
go test ./storage -v -run TestDeleteTask

# Test the handler
go test ./internal/handlers -v -run TestDeleteTaskHandler
```
**Expected**: All tests pass

### Manual Testing
1. Start the server: `go run cmd/server/main.go`
2. Login to get a token: `curl -X POST http://localhost:8080/login -d '{"email":"test@example.com","password":"password"}'`
3. Create a task: `curl -X POST http://localhost:8080/tasks -H "Authorization: Bearer YOUR_TOKEN" -d '{"description":"Test task"}'`
4. Delete the task: `curl -X DELETE http://localhost:8080/tasks/1 -H "Authorization: Bearer YOUR_TOKEN" -v`
5. **Expected response**: HTTP 204 No Content
6. Try to get the deleted task: `curl http://localhost:8080/tasks/1 -H "Authorization: Bearer YOUR_TOKEN"`
7. **Expected response**: HTTP 404 Not Found

## Common Pitfalls

### Pitfall 1: Forgetting to check RowsAffected
**Symptom**: DELETE returns 204 even for non-existent tasks
**Fix**: Always check `result.RowsAffected()` and return ErrTaskNotFound if zero

### Pitfall 2: Not including userID in WHERE clause
**Symptom**: Users can delete other users' tasks
**Fix**: Always include `AND user_id = ?` in the DELETE query

### Pitfall 3: Returning 200 instead of 204
**Symptom**: Works but doesn't follow REST conventions
**Fix**: Use `w.WriteHeader(http.StatusNoContent)` for successful DELETE

## Learning Resources

### Essential Reading
- [HTTP DELETE Method](https://developer.mozilla.org/en-US/docs/Web/HTTP/Methods/DELETE) - Official HTTP specification for DELETE
- [Go http.HandlerFunc](https://pkg.go.dev/net/http#HandlerFunc) - Understanding Go HTTP handlers
- [SQL DELETE Statement](https://www.sqlite.org/lang_delete.html) - SQLite DELETE syntax

### Additional Resources (Optional)
- [REST API Design Best Practices](https://stackoverflow.blog/2020/03/02/best-practices-for-rest-api-design/) - Understanding REST conventions
- [Idempotency in REST APIs](https://restfulapi.net/idempotent-rest-apis/) - Why DELETE should be idempotent
```

## Quality Checklist

When creating a task decomposition, verify:

- [ ] Overview clearly states what the task accomplishes
- [ ] Implementation approach is explained in 3-5 sentences
- [ ] Complexity is justified - simpler alternatives are considered
- [ ] Implementation doesn't add features beyond requirements
- [ ] Prerequisites list all required existing code and knowledge
- [ ] Steps are numbered and in logical order
- [ ] Each step includes: file path, what to do, what to implement, why, expected result
- [ ] No actual code implementations shown (user wants to code themselves)
- [ ] Descriptions are clear enough to guide implementation without showing code
- [ ] Verification section includes compile, test, and manual testing steps
- [ ] Common pitfalls section lists 2-4 realistic mistakes
- [ ] Learning resources include 2-5 relevant links
- [ ] Each step is completable in 5-10 minutes
- [ ] File modifications clearly indicate create vs. modify
- [ ] No vague instructions or placeholders

## Integration with Existing Standards

Task decompositions complement but don't replace:

- **tasks.md in specs**: High-level tasks for autonomous execution
- **code-review-standards.md**: Code quality requirements still apply
- **test-writing-standards.md**: Tests should follow established patterns

When a user requests a decomposition:
1. Reference the relevant task from tasks.md
2. Break it down following this document's structure
3. Describe what to implement without showing actual code
4. Ensure implementation descriptions align with code-review-standards.md
5. Include test steps that follow test-writing-standards.md

## Response Format

When the user requests a task decomposition, respond with:

1. Acknowledge which task you're decomposing
2. Present the full decomposition following the structure above
3. Ask if they need clarification on any steps
4. Offer to create the decomposition as a separate file if they want to save it

**Example response:**
```
I'll create a detailed decomposition for "Task 2.3: Implement User model with validation" from the tasks.md file.

[Full decomposition following the structure above]

Does this breakdown make sense? I can clarify any steps or create this as a separate file (e.g., `.kiro/specs/feature-name/task-2.3-decomposition.md`) if you'd like to reference it later.
```
