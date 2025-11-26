# Test Writing Standards for To-Do List Project

This document defines the test writing requirements and patterns that must be followed when generating tests for this Go project.

## 1. Test Structure

**Rule**: Use table-driven tests with `t.Run()` for all test functions.

- ✅ **Required**: Table-driven test structure with named test cases
- ✅ **Required**: Use `t.Run()` for subtests
- ✅ **Required**: Arrange-Act-Assert pattern with clear comments
- ✅ **Required**: Use test case values instead of creating separate test functions when possible
- ❌ **Avoid**: Creating multiple test functions when test cases can be added to a single table
- **Rationale**: Consistent, readable tests that are easy to maintain and extend

```go
// ✅ Good: Table-driven test with t.Run()
func TestValidateEmail(t *testing.T) {
    // ====Arrange====
    testCases := []struct {
        name        string
        email       string
        expectedErr error
    }{
        {
            name:        "Valid email",
            email:       "user@example.com",
            expectedErr: nil,
        },
        {
            name:        "Empty email",
            email:       "",
            expectedErr: ErrEmptyEmail,
        },
        {
            name:        "Missing @ symbol",
            email:       "userexample.com",
            expectedErr: ErrInvalidEmail,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ====Act====
            err := ValidateEmail(tc.email)

            // ====Assert====
            if !errors.Is(err, tc.expectedErr) {
                t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
            }
        })
    }
}

// ❌ Bad: Separate test functions for related scenarios
func TestValidateEmail_Valid(t *testing.T) {
    err := ValidateEmail("user@example.com")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
}

func TestValidateEmail_Empty(t *testing.T) {
    err := ValidateEmail("")
    if !errors.Is(err, ErrEmptyEmail) {
        t.Errorf("Expected ErrEmptyEmail, got %v", err)
    }
}

func TestValidateEmail_MissingAt(t *testing.T) {
    err := ValidateEmail("userexample.com")
    if !errors.Is(err, ErrInvalidEmail) {
        t.Errorf("Expected ErrInvalidEmail, got %v", err)
    }
}
// These should be combined into a single table-driven test
```

## 2. Error Handling in Tests

**Rule**: Use `errors.Is()` for error comparison, never direct equality.

- ✅ **Required**: Use `errors.Is(err, expectedErr)` for error assertions
- ✅ **Required**: Test both success and error paths
- ✅ **Required**: Test all custom error types defined in the code
- ❌ **Avoid**: Direct error comparison with `==`
- **Rationale**: Proper error wrapping support and accurate error checking

```go
// ✅ Good: Using errors.Is()
if !errors.Is(err, tc.expectedErr) {
    t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
}

// ❌ Bad: Direct comparison
if err != tc.expectedErr {
    t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
}
```

## 3. Database Testing

**Rule**: Use in-memory SQLite (`:memory:`) for all database tests.

- ✅ **Required**: Create database with `:memory:` for isolation
- ✅ **Required**: Use helper functions like `seedTask()` for test data setup
- ✅ **Required**: Close database connections with `defer`
- ✅ **Required**: Test with real database operations, not mocks
- ❌ **Avoid**: Temporary files or persistent databases
- **Rationale**: Fast, isolated tests without side effects

```go
// ✅ Good: In-memory database
func TestDatabaseStorage_CreateTask(t *testing.T) {
    // ====Arrange====
    s, err := storage.NewDatabaseStorage(":memory:")
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }
    defer s.Close()

    testUserID := 1
    
    testCases := []struct {
        name        string
        task        storage.Task
        expectedErr error
    }{
        {
            name:        "Valid task",
            task:        storage.Task{Description: "Test", Done: false},
            expectedErr: nil,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ====Act====
            id, err := s.CreateTask(tc.task, testUserID)

            // ====Assert====
            if !errors.Is(err, tc.expectedErr) {
                t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
            }
            if tc.expectedErr == nil && id == 0 {
                t.Error("Expected non-zero ID for successful creation")
            }
        })
    }
}
```

## 4. Helper Functions

**Rule**: Create helper functions for common test setup operations.

- ✅ **Required**: Helper functions for seeding test data
- ✅ **Required**: Use `t.Fatalf()` in helpers for setup failures
- ✅ **Required**: Clear, descriptive helper function names
- **Rationale**: DRY principle and clearer test code

```go
// ✅ Good: Helper function for seeding data
func seedTask(t *testing.T, s storage.Storage, task storage.Task, userID int) int {
    id, err := s.CreateTask(task, userID)
    if err != nil {
        t.Fatalf("Failed to seed task: %v", err)
    }
    return id
}

// Usage in tests
func TestGetTaskByID(t *testing.T) {
    s, _ := storage.NewDatabaseStorage(":memory:")
    defer s.Close()
    
    testUserID := 1
    taskID := seedTask(t, s, storage.Task{Description: "Test", Done: false}, testUserID)
    
    // Now test GetTaskByID with the seeded task
}
```

## 5. Test Real Implementations, Not Mocks

**Rule**: Test actual function logic with real implementations, mock only dependencies.

- ✅ **Required**: Test the function being tested with real implementation
- ✅ **Required**: Mock external dependencies (interfaces, external services)
- ✅ **Required**: Use real `strings.Reader` for input testing
- ❌ **Avoid**: Mocking the function you're testing
- **Rationale**: Tests should verify actual behavior, not mock behavior

```go
// ✅ Good: Testing real implementation with real input
func TestConsoleInputReader_ReadInput(t *testing.T) {
    testCases := []struct {
        name        string
        input       string
        maxLength   int
        expected    string
        expectedErr error
    }{
        {
            name:        "Valid input",
            input:       "test input\n",
            maxLength:   20,
            expected:    "test input",
            expectedErr: nil,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ====Arrange====
            reader := NewConsoleInputReader(strings.NewReader(tc.input))

            // ====Act====
            result, err := reader.ReadInput(tc.maxLength)

            // ====Assert====
            if !errors.Is(err, tc.expectedErr) {
                t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
            }
            if result != tc.expected {
                t.Errorf("Expected %q, got %q", tc.expected, result)
            }
        })
    }
}

// ❌ Bad: Mocking the function being tested
type MockInputReader struct {
    returnValue string
}
func (m *MockInputReader) ReadInput(maxLength int) (string, error) {
    return m.returnValue, nil
}
// This tests the mock, not the real ReadInput implementation!
```

## 6. Auth-Ready Testing

**Rule**: All Storage method tests must include `userID` parameter.

- ✅ **Required**: Pass `userID` to all Storage method calls
- ✅ **Required**: Test user ownership validation (wrong userID should fail)
- ✅ **Required**: Use consistent test user ID (e.g., `testUserID := 1`)
- **Rationale**: Multi-tenancy support and security

```go
// ✅ Good: Tests include userID
func TestDatabaseStorage_GetTaskByID(t *testing.T) {
    // ====Arrange====
    s, _ := storage.NewDatabaseStorage(":memory:")
    defer s.Close()
    
    testUserID := 1
    wrongUserID := 2
    
    taskID := seedTask(t, s, storage.Task{Description: "Test", Done: false}, testUserID)
    
    testCases := []struct {
        name        string
        taskID      int
        userID      int
        expectedErr error
    }{
        {
            name:        "Valid task ID with correct user",
            taskID:      taskID,
            userID:      testUserID,
            expectedErr: nil,
        },
        {
            name:        "Valid task ID with wrong user",
            taskID:      taskID,
            userID:      wrongUserID,
            expectedErr: storage.ErrTaskNotFound,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ====Act====
            task, err := s.GetTaskByID(tc.taskID, tc.userID)

            // ====Assert====
            if !errors.Is(err, tc.expectedErr) {
                t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
            }
        })
    }
}
```

## 7. Test Case Coverage

**Rule**: Generate comprehensive test cases covering success, errors, and edge cases.

- ✅ **Required**: At least one success case
- ✅ **Required**: Test all error conditions
- ✅ **Required**: Test boundary values (empty strings, zero values, nil)
- ✅ **Required**: Test edge cases specific to the function
- **Rationale**: Comprehensive coverage catches bugs early

### Standard Test Cases by Function Type

**Validation Functions:**
- Valid input - should succeed
- Empty/nil input - should return appropriate error
- Invalid format - should return validation error
- Boundary values - should handle correctly

**CRUD Operations:**
- Valid operation - should succeed
- Non-existent ID - should return ErrNotFound
- Wrong user ID - should return ErrNotFound (auth check)
- Database error - should return wrapped error
- Invalid input - should return validation error

## 7a. Avoid Test Case Redundancy

**Rule**: Always check for redundant test cases that test the same behavior.

- ✅ **Required**: Review test cases to identify redundancy before finalizing
- ✅ **Required**: Consolidate multiple test cases that verify the same validation rule or behavior
- ✅ **Required**: Keep only boundary cases and representative examples
- ❌ **Avoid**: Multiple test cases that produce the same error message or test the same code path
- ❌ **Avoid**: Testing every possible input value when one representative case is sufficient
- **Rationale**: Reduces maintenance burden, improves test readability, and speeds up test execution

### Examples of Redundancy to Avoid

```go
// ❌ Bad: Redundant test cases for the same validation rule
testCases := []struct {
    name        string
    port        int
    expectedErr string
}{
    {
        name:        "Port is zero",
        port:        0,
        expectedErr: "port must be between 1 and 65535",
    },
    {
        name:        "Port is negative",
        port:        -1,
        expectedErr: "port must be between 1 and 65535",
    },
    {
        name:        "Port is too high",
        port:        99999,
        expectedErr: "port must be between 1 and 65535",
    },
}
// All three test the same validation rule - only need one

// ✅ Good: Single test case for the validation rule
testCases := []struct {
    name        string
    port        int
    expectedErr string
}{
    {
        name:        "Invalid port",
        port:        99999,
        expectedErr: "port must be between 1 and 65535",
    },
}
```

```go
// ❌ Bad: Testing every length when behavior is the same
testCases := []struct {
    name     string
    input    string
    expected string
}{
    {name: "Empty string", input: "", expected: "****"},
    {name: "1 character", input: "a", expected: "****"},
    {name: "2 characters", input: "ab", expected: "****"},
    {name: "3 characters", input: "abc", expected: "****"},
    {name: "4 characters", input: "abcd", expected: "****"},
    {name: "5 characters", input: "abcde", expected: "ab****de"},
    {name: "6 characters", input: "abcdef", expected: "ab****ef"},
}
// Strings ≤4 chars all return "****" - only need one case
// Strings >4 chars all use same masking logic - only need boundary case

// ✅ Good: Minimal test cases covering all behaviors
testCases := []struct {
    name     string
    input    string
    expected string
}{
    {
        name:     "Short string (4 chars or less)",
        input:    "abc",
        expected: "****",
    },
    {
        name:     "Boundary case (5 characters)",
        input:    "abcde",
        expected: "ab****de",
    },
    {
        name:     "Long string",
        input:    "this-is-a-long-secret-key",
        expected: "th****ey",
    },
}
```

### Guidelines for Identifying Redundancy

1. **Same error message**: If multiple test cases produce the same error message, they likely test the same validation rule
2. **Same code path**: If multiple test cases execute the same code path with different inputs, consolidate to one representative case
3. **Boundary testing**: Keep boundary cases (e.g., exactly 32 chars for a 32-char minimum) but remove redundant cases on either side
4. **Representative examples**: Choose one clear example per behavior rather than exhaustive variations

### When Multiple Cases Are Justified

Multiple test cases ARE appropriate when:
- Testing different error conditions (different validation rules)
- Testing different code paths (if/else branches)
- Testing boundary values that trigger different behaviors
- Testing different types (e.g., nil vs empty vs invalid)
- Testing user ownership (correct user vs wrong user)

**HTTP Handlers:**
- Valid request - should return 200 with expected data
- Missing authentication - should return 401
- Invalid input - should return 400
- Not found - should return 404
- Internal error - should return 500

## 8. Edge Case Handling

**Rule**: Handle special Go patterns correctly in tests.

### Variadic Functions

```go
// Function with variadic parameters
func Sum(numbers ...int) int

// ✅ Good: Test variadic with 0, 1, and multiple arguments
func TestSum(t *testing.T) {
    testCases := []struct {
        name     string
        numbers  []int  // Converted from ...int
        expected int
    }{
        {
            name:     "Zero arguments",
            numbers:  nil,
            expected: 0,
        },
        {
            name:     "Single argument",
            numbers:  []int{5},
            expected: 5,
        },
        {
            name:     "Multiple arguments",
            numbers:  []int{1, 2, 3},
            expected: 6,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ====Act====
            result := Sum(tc.numbers...)  // Expand with ...

            // ====Assert====
            if result != tc.expected {
                t.Errorf("Expected %d, got %d", tc.expected, result)
            }
        })
    }
}
```

### Multiple Return Values

```go
// Function with multiple return values
func GetTask(id int) (Task, bool, error)

// ✅ Good: Assert all return values
func TestGetTask(t *testing.T) {
    testCases := []struct {
        name        string
        id          int
        expected1   Task   // First return
        expected2   bool   // Second return
        expectedErr error  // Error return
    }{
        // Test cases...
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ====Act====
            result1, result2, err := GetTask(tc.id)

            // ====Assert====
            if !errors.Is(err, tc.expectedErr) {
                t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
            }
            if result1 != tc.expected1 {
                t.Errorf("Expected task %v, got %v", tc.expected1, result1)
            }
            if result2 != tc.expected2 {
                t.Errorf("Expected found %v, got %v", tc.expected2, result2)
            }
        })
    }
}
```

### Functions with No Parameters

```go
// Function with no parameters
func GetVersion() string

// ✅ Good: Simple test case
func TestGetVersion(t *testing.T) {
    testCases := []struct {
        name     string
        expected string
    }{
        {
            name:     "Returns version string",
            expected: "1.0.0",
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ====Act====
            result := GetVersion()

            // ====Assert====
            if result != tc.expected {
                t.Errorf("Expected %s, got %s", tc.expected, result)
            }
        })
    }
}
```

### Methods with Pointer Receivers

```go
// Method with pointer receiver
func (tm *TaskManager) AddTask(description string) error

// ✅ Good: Proper receiver setup
func TestTaskManager_AddTask(t *testing.T) {
    // ====Arrange====
    s, _ := storage.NewDatabaseStorage(":memory:")
    defer s.Close()
    
    tm := task.NewTaskManager(s, &strings.Builder{})
    
    testCases := []struct {
        name        string
        description string
        expectedErr error
    }{
        // Test cases...
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ====Act====
            err := tm.AddTask(tc.description)

            // ====Assert====
            if !errors.Is(err, tc.expectedErr) {
                t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
            }
        })
    }
}
```

## 9. Import Organization

**Rule**: Include all necessary imports for testing.

- ✅ **Required**: `testing` package (always)
- ✅ **Required**: `errors` package (for error comparison)
- ✅ **Required**: `strings` package (for string builders in tests)
- ✅ **Required**: `database/sql` and driver (for database tests)
- ✅ **Required**: Project packages being tested
- **Rationale**: Complete, compilable test files

```go
// ✅ Good: Complete imports
package storage

import (
    "database/sql"
    "errors"
    "strings"
    "testing"

    _ "github.com/mattn/go-sqlite3"
)
```

## 10. Test Naming

**Rule**: Use clear, descriptive test names.

- ✅ **Required**: `Test<FunctionName>` for functions
- ✅ **Required**: `Test<Type>_<MethodName>` for methods
- ✅ **Required**: Descriptive test case names in table
- **Rationale**: Easy to identify what's being tested

```go
// ✅ Good: Clear naming
func TestValidateEmail(t *testing.T) { }
func TestDatabaseStorage_CreateTask(t *testing.T) { }
func TestTaskManager_AddTask(t *testing.T) { }

// Test case names
{
    name: "Valid email with subdomain",
    name: "Empty description returns error",
    name: "Non-existent task ID returns ErrTaskNotFound",
}
```

## 11. Warnings for Complex Patterns

**Rule**: Document when tests need special attention.

Patterns that may require manual customization:
- ⚠️ **Channels**: May require goroutine testing
- ⚠️ **Function parameters**: May require mock implementations
- ⚠️ **Empty interfaces**: Should cover multiple types
- ⚠️ **Goroutines**: May require synchronization primitives

```go
// When generating tests for complex patterns, add TODO comments:

// TODO: This test uses channels - verify goroutine behavior
// TODO: This test uses function parameters - implement appropriate mock
// TODO: This test uses empty interface - add test cases for different types
```

## 12. Temporary Files and Directories

**Rule**: Tests must not create unwanted temporary files or directories in the source tree.

- ✅ **Required**: Use absolute paths (e.g., `/tmp/`) for any file paths in test data
- ✅ **Required**: Clean up any temporary files/directories created during tests
- ✅ **Required**: Verify no artifacts remain after running tests
- ❌ **Avoid**: Relative paths like `./data/` or `./test/` that create directories in source tree
- ❌ **Avoid**: Leaving temporary files or directories after test completion
- **Rationale**: Keep the source tree clean and avoid polluting the repository with test artifacts

```go
// ✅ Good: Using absolute paths for test data
func TestConfigLoading(t *testing.T) {
    testCases := []struct {
        name           string
        configPath     string
        expectedDBPath string
    }{
        {
            name:           "Custom database path",
            configPath:     "/tmp/config.yaml",
            expectedDBPath: "/tmp/data/tasks.db",  // Absolute path
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

// ✅ Good: Cleaning up temporary files
func TestFileOperations(t *testing.T) {
    tmpFile, err := os.CreateTemp("", "test-*.txt")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }
    defer os.Remove(tmpFile.Name())  // Clean up
    defer tmpFile.Close()
    
    // Test implementation
}

// ❌ Bad: Using relative paths that create directories
func TestConfigLoading(t *testing.T) {
    testCases := []struct {
        name           string
        configPath     string
        expectedDBPath string
    }{
        {
            name:           "Custom database path",
            configPath:     "./config.yaml",
            expectedDBPath: "./data/tasks.db",  // Creates ./data/ directory!
        },
        {
            name:           "Test database path",
            configPath:     "./config.yaml",
            expectedDBPath: "./test/tasks.db",  // Creates ./test/ directory!
        },
    }
}
```

### Verification Steps

After writing tests, always verify:

1. Run the tests: `go test ./...`
2. Check for new directories: `ls -la` in the test package directory
3. If unwanted directories exist, update test data to use `/tmp/` paths
4. Clean up any artifacts: `rm -rf unwanted_directory`
5. Re-run tests to confirm no artifacts are created

### Common Culprits

Watch out for these patterns that may create unwanted directories:

- Database file paths in configuration tests
- File paths in test case data structures
- Paths used in temporary file creation
- Working directory assumptions in tests

## 13. Test Verification

**Rule**: Always verify generated tests compile and run.

- ✅ **Required**: Check compilation with `go build`
- ✅ **Required**: Run tests with `go test`
- ✅ **Required**: Fix any compilation errors
- ✅ **Required**: Review and fix any failing tests
- ✅ **Required**: Verify no temporary files or directories were created
- **Rationale**: Ensure tests are functional, not just syntactically correct

## Test Generation Workflow

When generating tests, follow this workflow:

1. **Analyze** the source file to understand:
   - Package name and imports
   - Exported functions and methods
   - Dependencies (Storage, DB, HTTP, etc.)
   - Error types and constants
   - Special patterns (variadic, context, etc.)

2. **Generate test cases** covering:
   - Success paths
   - Error conditions
   - Edge cases
   - Boundary values
   - User ownership (for Storage methods)

3. **Create test structure**:
   - Helper functions for setup
   - Table-driven tests with t.Run()
   - Arrange-Act-Assert pattern
   - Proper error handling

4. **Verify tests**:
   - Check compilation
   - Run tests
   - Fix any issues
   - Ensure all tests pass

## Examples by Pattern

### Simple Validation Function
```go
func ValidateEmail(email string) error
```
**Test cases**: Valid email, empty email, invalid format, unicode characters

### Database CRUD Method
```go
func (ds *DatabaseStorage) GetTaskByID(id int, userID int) (Task, error)
```
**Test cases**: Valid ID, non-existent ID, wrong user ID, database error

### HTTP Handler
```go
func GetTaskHandler(s storage.Storage) http.HandlerFunc
```
**Test cases**: Valid request, missing auth, invalid ID, not found, internal error

### Variadic Function
```go
func Sum(numbers ...int) int
```
**Test cases**: Zero args, single arg, multiple args, negative numbers

## Review Checklist

When reviewing generated tests, verify:

- [ ] Table-driven structure with t.Run()
- [ ] Arrange-Act-Assert pattern with comments
- [ ] Error comparison using errors.Is()
- [ ] In-memory database for storage tests
- [ ] Helper functions for test data setup
- [ ] All test cases include userID for Storage methods
- [ ] Comprehensive coverage (success, errors, edge cases)
- [ ] Clear, descriptive test case names
- [ ] All necessary imports included
- [ ] Tests compile without errors
- [ ] Tests run and pass
- [ ] Real implementations tested, not mocks
- [ ] Special patterns handled correctly (variadic, multiple returns, etc.)
- [ ] No temporary files or directories created in source tree
- [ ] All file paths in test data use absolute paths (e.g., `/tmp/`)

## Common Mistakes to Avoid

❌ **Don't** use direct error comparison (`err == expectedErr`)
❌ **Don't** mock the function being tested
❌ **Don't** use persistent databases or temp files
❌ **Don't** forget userID parameter in Storage tests
❌ **Don't** skip edge cases and error paths
❌ **Don't** use unclear test case names
❌ **Don't** forget to close database connections
❌ **Don't** generate tests for unexported functions
❌ **Don't** use relative paths in test data that create directories in source tree
❌ **Don't** leave temporary files or directories after tests complete

## Integration with Code Review Standards

Generated tests must follow the same standards as production code:

- Stateless services (in-memory databases)
- Granular CRUD operations
- Dependency injection
- Proper error handling
- Auth-ready (userID parameters)

Tests are part of the codebase and must meet the same quality standards.
