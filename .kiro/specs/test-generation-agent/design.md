# Design Document

## Overview

The Test Generation Agent is a Kiro workflow that automates the creation of comprehensive unit tests for Go source files. The agent analyzes Go code, generates a testing plan for user approval, and then creates test files that strictly follow the project's established testing patterns. This design ensures consistency, reduces manual test writing effort, and maintains high code quality standards.

The agent operates as a conversational workflow within Kiro, leveraging code analysis tools and file manipulation capabilities to understand source code structure and generate appropriate tests.

## Architecture

### High-Level Flow

```
User Request → File Analysis → Testing Plan Generation → User Approval → Test Code Generation → Verification
```

### Components

1. **Request Handler**: Parses user input to extract target file path
2. **Code Analyzer**: Examines Go source files to identify testable functions and methods
3. **Plan Generator**: Creates structured testing plan based on code analysis
4. **Approval Manager**: Handles user feedback and plan iterations
5. **Test Generator**: Generates actual test code following project patterns
6. **Verifier**: Validates generated tests compile and run

## Components and Interfaces

### 1. Request Handler

**Responsibility**: Parse and validate user requests

**Input**:
- User message containing file path (e.g., "Generate tests for storage/database.go")

**Output**:
- Validated file path
- Error if file doesn't exist or isn't a Go file

**Logic**:
- Extract file path from user message
- Verify file exists and has `.go` extension
- Ensure file is not already a test file (`*_test.go`)

### 2. Code Analyzer

**Responsibility**: Parse Go source code and extract testable elements

**Input**:
- Go source file path

**Output**:
- List of exported functions with signatures
- List of exported methods with receiver types
- Package name and imports
- Identified error types and constants

**Analysis Strategy**:
```go
type FunctionInfo struct {
    Name       string
    Receiver   string  // Empty for functions, type name for methods
    Parameters []Parameter
    Returns    []ReturnType
    IsExported bool
}

type Parameter struct {
    Name string
    Type string
}

type ReturnType struct {
    Type      string
    IsError   bool
    IsPointer bool
}
```

**Implementation Approach**:
- Use `grepSearch` to find function/method declarations
- Parse function signatures to extract parameters and return types
- Identify error types and sentinel errors in the file
- Detect dependencies (Storage interfaces, database connections, etc.)

### 3. Plan Generator

**Responsibility**: Create structured testing plan document

**Input**:
- Code analysis results
- Project code review standards
- Example test file patterns

**Output**:
- Markdown document with testing plan

**Plan Structure**:
```markdown
# Testing Plan for [filename]

## Package: [package_name]

## Functions/Methods to Test

### 1. FunctionName(param1 type1, param2 type2) (returnType, error)

**Test Cases**:
- Valid input scenario 1
- Valid input scenario 2
- Invalid input: empty/nil values
- Invalid input: boundary conditions
- Error handling: specific error type
- Edge case: [specific to function]

**Test Setup Requirements**:
- In-memory database (if applicable)
- Mock dependencies (if applicable)
- Helper functions needed

### 2. [Next function...]

## Test File Structure

- Package declaration: [package_name]
- Required imports: testing, errors, [others]
- Helper functions: [list]
- Test functions: [list with Test prefix]
```

**Plan Generation Rules**:
- Include all exported functions and methods
- Propose 4-8 test cases per function (success, failure, edge cases)
- Identify required test infrastructure (mocks, in-memory DB, helpers)
- Reference code review standards for patterns to follow
- Specify Arrange-Act-Assert structure for each test

### 4. Approval Manager

**Responsibility**: Handle user feedback loop

**Input**:
- Generated testing plan
- User feedback/approval

**Output**:
- Approved plan (proceed to generation)
- Modified plan (iterate)
- Cancellation signal

**Workflow**:
1. Present plan to user with clear formatting
2. Ask: "Does this testing plan look good? Please review the proposed test cases and let me know if you'd like any changes."
3. Wait for user response
4. If approved → proceed to Test Generator
5. If changes requested → modify plan and re-present
6. If cancelled → exit workflow

### 5. Test Generator

**Responsibility**: Generate actual Go test code

**Input**:
- Approved testing plan
- Original source file
- Code review standards
- Example test patterns

**Output**:
- Complete `*_test.go` file

**Generation Strategy**:

**File Structure**:
```go
package [package_name]

import (
    "errors"
    "testing"
    // Additional imports based on requirements
)

// Helper functions (if needed)
func seedTask(t *testing.T, s storage.Storage, task storage.Task) {
    // Implementation
}

// Test functions
func TestFunctionName(t *testing.T) {
    // ====Arrange====
    testCases := []struct {
        name        string
        // Input fields
        // Expected output fields
        expectedErr error
    }{
        {
            name: "Valid case description",
            // Test case data
        },
        // More test cases...
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // ====Act====
            result, err := FunctionName(tc.input)

            // ====Assert====
            if !errors.Is(err, tc.expectedErr) {
                t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
            }
            // Additional assertions
        })
    }
}
```

**Pattern Matching Rules**:

1. **Table-Driven Tests**: Always use struct slices with `testCases`
2. **Arrange-Act-Assert Comments**: Include in every test function
3. **Error Comparison**: Use `errors.Is()` for error checking
4. **Test Names**: Descriptive strings in `name` field
5. **Database Tests**: Use `:memory:` for SQLite
6. **Helper Functions**: Create `seedTask`-style helpers for setup
7. **Dependency Injection**: Pass dependencies to constructors in setup

**Code Review Standards Integration**:
- Check if function uses Storage interface → add `userID` parameter in tests
- Check if function does DB operations → use in-memory SQLite
- Check if function returns errors → test error wrapping with `fmt.Errorf`
- Check if function is a method → test with proper receiver setup

### 6. Verifier

**Responsibility**: Validate generated tests

**Input**:
- Generated test file path

**Output**:
- Compilation status
- Test execution results
- Error messages if any

**Verification Steps**:
1. Run `getDiagnostics` on generated test file
2. If compilation errors exist:
   - Attempt automatic fixes (missing imports, syntax errors)
   - Report to user if unfixable
3. Run `go test` on the test file
4. Report results to user

**Error Handling**:
- Missing imports → Add automatically
- Type mismatches → Report to user with context
- Test failures → Report but don't auto-fix (user should review logic)

## Data Models

### Testing Plan Document

```markdown
# Testing Plan for [filename]

## Package: [package_name]

## Overview
[Brief description of what the file does]

## Functions/Methods to Test

### [Function/Method Signature]
**Purpose**: [What it does]
**Test Cases**:
1. [Case description]
2. [Case description]
...

**Setup Requirements**:
- [Requirement 1]
- [Requirement 2]

## Test Infrastructure

### Helper Functions
- `[helperName]`: [purpose]

### Mock Objects (if needed)
- `[mockName]`: [purpose]

### Test Data
- [Data setup description]

## Dependencies
- [List of imports needed]
- [External dependencies]
```

### Code Analysis Result

```go
type AnalysisResult struct {
    PackageName     string
    FileName        string
    Functions       []FunctionInfo
    Methods         []MethodInfo
    ErrorTypes      []string
    Constants       []string
    RequiresDB      bool
    RequiresStorage bool
    Imports         []string
}

type FunctionInfo struct {
    Name           string
    Signature      string
    Parameters     []Parameter
    Returns        []ReturnType
    HasErrorReturn bool
    IsExported     bool
}

type MethodInfo struct {
    FunctionInfo
    ReceiverType string
    ReceiverPtr  bool
}
```

## Error Handling

### User-Facing Errors

1. **Invalid File Path**
   - Message: "Could not find file: [path]. Please provide a valid Go source file."
   - Action: Prompt user to provide correct path

2. **File is Already a Test File**
   - Message: "The file [path] is already a test file. Please specify a source file (not *_test.go)."
   - Action: Prompt user for source file

3. **No Exportable Functions Found**
   - Message: "No exported functions or methods found in [file]. Test generation requires exported functions."
   - Action: Explain that only exported functions can be tested from external test files

4. **Compilation Errors in Generated Tests**
   - Message: "Generated tests have compilation errors: [errors]. Attempting to fix..."
   - Action: Try automatic fixes, report if manual intervention needed

5. **Test Execution Failures**
   - Message: "Generated tests compiled but [X] tests failed. Please review the test logic."
   - Action: Show test output, let user decide on fixes

### Internal Errors

- Code analysis failures → Log and report to user
- File write failures → Report with file permissions check
- Verification failures → Report with diagnostic information

## Testing Strategy

### Unit Testing the Agent Workflow

Since this is a Kiro agent workflow, testing involves:

1. **Manual Testing**:
   - Test with various Go files (simple functions, methods, database operations)
   - Verify plan generation quality
   - Verify generated test code quality
   - Test approval workflow iterations

2. **Example Test Cases**:
   - Simple validation functions (like `validation/validation.go`)
   - Database operations (like `storage/database.go`)
   - HTTP handlers (like `internal/handlers/handlers.go`)
   - CLI operations (like `cmd/cli/cli.go`)

3. **Quality Checks**:
   - Generated tests must compile
   - Generated tests must follow Arrange-Act-Assert pattern
   - Generated tests must use table-driven approach
   - Generated tests must include error checking with `errors.Is()`
   - Generated tests must use in-memory DB when applicable

### Edge Cases to Handle

1. **Functions with no parameters**: Generate simple test cases
2. **Functions with many parameters**: Use struct initialization
3. **Variadic functions**: Test with 0, 1, and multiple arguments
4. **Functions returning multiple values**: Assert all return values
5. **Methods on pointer receivers**: Ensure proper receiver setup
6. **Functions using interfaces**: Create appropriate mocks
7. **Functions with complex dependencies**: Identify and document in plan

## Implementation Notes

### File Naming Convention

- Source file: `example.go`
- Generated test file: `example_test.go`
- Location: Same directory as source file

### Import Management

Common imports for generated tests:
```go
import (
    "errors"
    "testing"
    "strings"  // For string builders
    "github.com/google/go-cmp/cmp"  // For deep comparisons
)
```

Additional imports based on analysis:
- Database tests: `"myproject/storage"`
- HTTP tests: `"net/http"`, `"net/http/httptest"`
- Mock tests: `"github.com/stretchr/testify/assert"`

### Pattern Recognition

The agent should recognize these patterns in source code:

1. **Storage Interface Usage**: Add `userID` parameter
2. **Database Operations**: Use `:memory:` SQLite
3. **Error Returns**: Test with `errors.Is()`
4. **Validation Functions**: Test valid and invalid inputs
5. **HTTP Handlers**: Use `httptest.ResponseRecorder`

### Customization Points

Users can customize the testing plan before generation:
- Add/remove test cases
- Modify test case descriptions
- Change setup requirements
- Add specific edge cases
- Request different assertion styles

## Workflow Integration

### Kiro Agent Execution

The agent operates as a conversational workflow:

1. User: "Generate tests for storage/database.go"
2. Agent: Analyzes file, generates plan
3. Agent: Presents plan with approval prompt
4. User: Reviews and approves/modifies
5. Agent: Generates test code
6. Agent: Verifies compilation
7. Agent: Reports completion with file location

### Context Requirements

The agent needs access to:
- Target source file
- Code review standards (`.kiro/steering/code-review-standards.md`)
- Example test files for pattern reference
- Project structure for import paths

### Output Artifacts

1. **Testing Plan** (temporary, shown to user)
2. **Test File** (`*_test.go` in source directory)
3. **Verification Report** (compilation and test results)

## Design Decisions and Rationales

### Why Two-Phase Approach (Plan → Generate)?

- **User Control**: Allows review before code generation
- **Customization**: Users can modify test cases before generation
- **Transparency**: Clear visibility into what will be tested
- **Iteration**: Easy to refine without regenerating code

### Why Table-Driven Tests?

- **Project Standard**: Matches existing test patterns
- **Maintainability**: Easy to add new test cases
- **Readability**: Clear structure with named test cases
- **Efficiency**: Reduces code duplication

### Why In-Memory SQLite for DB Tests?

- **Speed**: Faster than file-based databases
- **Isolation**: Each test gets fresh database
- **No Cleanup**: Automatic cleanup when test ends
- **Project Standard**: Matches code review standards

### Why Arrange-Act-Assert Comments?

- **Project Standard**: Existing tests use this pattern
- **Clarity**: Makes test structure explicit
- **Consistency**: All tests follow same structure
- **Readability**: Easy to understand test flow

## Future Enhancements

Potential improvements for future iterations:

1. **Benchmark Test Generation**: Generate performance tests
2. **Coverage Analysis**: Identify untested code paths
3. **Mock Generation**: Auto-generate mock implementations
4. **Integration Test Support**: Generate end-to-end tests
5. **Test Data Generation**: Smart test data based on types
6. **Mutation Testing**: Generate tests that catch common bugs
7. **Batch Processing**: Generate tests for multiple files
8. **CI/CD Integration**: Auto-generate tests on file changes
