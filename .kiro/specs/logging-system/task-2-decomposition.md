# Task Decomposition: Implement Standard Fields and Helper Functions

## Overview

This task establishes standardized field names and helper functions for the logging system. You'll create constants for common log fields (like request_id, user_id, method, etc.) that ensure consistency across the application and compatibility with log aggregation tools. You'll also implement security-focused helper functions to mask sensitive data like emails and tokens before logging them.

## Implementation Approach

We're creating a dedicated file for field definitions and helper functions that will be used throughout the application. This approach centralizes logging conventions and ensures that:

1. Field names are consistent across all log entries
2. Log aggregation tools (ELK Stack, Grafana Loki, Datadog) can parse logs correctly
3. Sensitive data is automatically masked before logging
4. Common logging patterns are simplified with helper functions

The field naming follows industry standards (lowercase with underscores) which is compatible with most log aggregation systems. The masking functions implement privacy best practices by hiding sensitive information while preserving enough context for debugging.

**Key Concepts:**
- **Field Name Conventions**: Lowercase with underscores (snake_case) for log aggregation tool compatibility
- **Data Masking**: Partial obfuscation of sensitive data (keeping some information for debugging)
- **Privacy by Design**: Never log sensitive data in plain text
- **Type Safety**: Using constants prevents typos and enables IDE autocomplete

## Prerequisites

**Existing Code:**
- `logger/logger.go` - Logger factory functions
- `logger/config.go` - Config struct

**Dependencies:**
- `strings` package (standard library) - for string manipulation in masking functions
- `log/slog` package (standard library) - for slog.Attr types

**Knowledge Required:**
- Understanding of Go constants
- Basic string manipulation in Go
- Understanding of privacy and security concepts
- Familiarity with log aggregation tools (ELK, Loki, Datadog)

## Step-by-Step Instructions

### Step 1: Create fields.go file with package declaration

**File**: `logger/fields.go` (create new file)

**What to do:**
Create a new file to hold field name constants and helper functions.

**What to implement:**
- Add package declaration: `package logger`
- Import required packages: `strings` and `log/slog`
- Add a package-level comment explaining the purpose of this file:
  - Standard field names for consistent logging
  - Masking functions for sensitive data
  - Helper functions for common attributes
  - Compatibility with log aggregation tools
- Keep the comment concise (2-4 sentences)

**Why:**
Separating field definitions into their own file keeps the code organized and makes it easy to find and update field names. Clear documentation helps other developers understand the purpose and usage of the package.

**Expected result:**
You have a `logger/fields.go` file with package declaration and imports. The file compiles without errors.

---

### Step 2: Define standard field name constants

**File**: `logger/fields.go` (modify existing)

**What to do:**
Define constants for all standard field names used throughout the application.

**What to implement:**
- Create a `const` block
- Define the following string constants (all lowercase with underscores):
  - `FieldRequestID` = "request_id"
  - `FieldUserID` = "user_id"
  - `FieldMethod` = "method"
  - `FieldPath` = "path"
  - `FieldStatusCode` = "status_code"
  - `FieldDuration` = "duration_ms"
  - `FieldError` = "error"
  - `FieldOperation` = "operation"
  - `FieldTaskID` = "task_id"
  - `FieldEmail` = "email"
  - `FieldTraceID` = "trace_id"
  - `FieldSpanID` = "span_id"
- Add comments for fields that need clarification (e.g., "// Always masked" for email)

**Why:**
These field names follow the snake_case convention required by log aggregation tools like Grafana Loki. Using constants prevents typos in log calls, makes refactoring easier (change in one place), and enables IDE autocomplete. The constants ensure consistency across the codebase.

**Expected result:**
You have a set of field name constants that can be used throughout the application. The file compiles.

---

### Step 3: Implement MaskEmail function signature

**File**: `logger/fields.go` (modify existing)

**What to do:**
Create the function that will mask email addresses for privacy.

**What to implement:**
- Create a function named `MaskEmail`
- Accept one parameter: `email string`
- Return one value: `string` (the masked email)
- Add function documentation explaining:
  - What it does (masks email for privacy)
  - The masking pattern (e.g., "user@example.com" → "u***r@example.com")
  - When to use it (always when logging emails)

**Why:**
Email addresses are personally identifiable information (PII). Logging them in plain text violates privacy best practices and regulations like GDPR. Masking preserves enough information for debugging (you can identify which user) while protecting privacy.

**Expected result:**
You have a function signature with documentation. It can return an empty string for now.

---

### Step 4: Implement email validation logic

**File**: `logger/fields.go` (modify existing)

**What to do:**
Add logic to validate that the input is a valid email format.

**What to implement:**
- Split the email string by "@" using `strings.Split(email, "@")`
- Check if the result has exactly 2 parts (username and domain)
- Store the username part (index 0) and domain part (index 1) in variables
- If not 2 parts, return "***" (invalid email format)

**Why:**
Before masking, we need to ensure the input is actually an email. Invalid inputs (empty strings, malformed emails) should be handled gracefully. Returning "***" indicates something went wrong and this is safe.

**Expected result:**
The function can detect valid vs invalid email formats and handle both cases.

---

### Step 5: Implement username masking logic

**File**: `logger/fields.go` (modify existing)

**What to do:**
Add logic to mask the username portion of the email.

**What to implement:**
- Check the length of the username
- If username length is 2 or less:
  - Return "***@" + domain (completely mask short usernames)
- If username length is greater than 2:
  - Keep the first character
  - Replace middle characters with "***"
  - Keep the last character
  - Combine: first + "***" + last
- For example:
  - "a@example.com" → "***@example.com"
  - "ab@example.com" → "***@example.com"
  - "user@example.com" → "u***r@example.com"
  - "verylongusername@example.com" → "v***e@example.com"

**Why:**
Short usernames (1-2 chars) would be easily guessable if we kept any characters, so we mask them completely. For longer usernames, keeping first and last characters provides enough context for debugging while protecting privacy. The domain is preserved (it's not PII).

**Expected result:**
Emails are masked appropriately based on username length.

---

### Step 6: Implement MaskToken function

**File**: `logger/fields.go` (modify existing)

**What to do:**
Create a function to mask authentication tokens and API keys.

**What to implement:**
- Create a function named `MaskToken`
- Accept one parameter: `token string`
- Return one value: `string` (the masked token)
- Add function documentation explaining its purpose
- Implement the masking logic:
  - If token length is 8 or less: return "****"
  - If token length is greater than 8:
    - Keep first 4 characters
    - Replace middle with "****"
    - Keep last 4 characters
    - Return: first4 + "****" + last4
- For example:
  - "abc" → "****"
  - "abc123def456ghi789" → "abc1****i789"
  - "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." → "eyJh****VCJ9"

**Why:**
Tokens (JWT, API keys, session tokens) are authentication credentials. Logging them in plain text is a critical security vulnerability. Anyone with access to logs could steal tokens and impersonate users. Masking tokens securely while keeping enough information to identify which token was used.

**Expected result:**
Tokens are masked securely. Short tokens are completely hidden, long tokens show first/last 4 characters only.

---

### Step 7: Add helper functions for common attributes

**File**: `logger/fields.go` (modify existing)

**What to do:**
Create convenience functions that return commonly-used slog attributes.

**What to implement:**
- Create a function `RequestIDAttr` that:
  - Accepts `requestID string`
  - Returns `slog.Attr`
  - Returns `slog.String(FieldRequestID, requestID)`
- Create a function `UserIDAttr` that:
  - Accepts `userID int`
  - Returns `slog.Attr`
  - Returns `slog.Int(FieldUserID, userID)`
- Create a function `ErrorAttr` that:
  - Accepts `err error`
  - Returns `slog.Attr`
  - Returns `slog.String(FieldError, err.Error())`
- Add similar functions for other common fields if desired

**Why:**
These helper functions reduce boilerplate and ensure consistent field names. Instead of writing `slog.String("request_id", id)` every time, you write `RequestIDAttr(id)`. This is shorter, type-safe, and uses the constant automatically.

**Expected result:**
You have helper functions that make logging common attributes easier and more consistent.

---

### Step 8: Add package documentation

**File**: `logger/fields.go` (modify existing)

**What to do:**
Add a comment block at the top of the file explaining its purpose.

**What to implement:**
- Add a comment before the package declaration that explains:
  - What this file defines
  - Standard field names for consistent logging
  - Masking functions for sensitive data
  - Helper functions for common attributes
  - Compatibility with log aggregation tools
- Mention the naming conventions (snake_case)
- Keep it concise (2-4 sentences)

**Why:**
Good documentation helps developers understand the package. It appears in godoc and IDE tooltips.

**Expected result:**
The file has clear documentation explaining its contents and purpose.

---

## Complete File Structure

Here's what the file structure should look like (without showing actual code):

```
logger/fields.go
├── Package comment
├── Package declaration
├── Imports (strings, log/slog)
├── Constants block
│   ├── FieldRequestID
│   ├── FieldUserID
│   ├── ... (all other field constants)
│   └── FieldSpanID
├── MaskEmail function
│   ├── Split by @
│   ├── Validate format
│   ├── Mask username
│   └── Return masked email
├── MaskToken function
│   ├── Check length
│   ├── Mask middle portion
│   └── Return masked token
└── Helper functions (optional)
    ├── RequestIDAttr
    ├── UserIDAttr
    └── ErrorAttr
```

## Verification

### Compile Check
```bash
go build ./logger
```
**Expected**: No compilation errors

### Test Field Constants
```bash
# Create a test file
cat > test_fields.go << 'EOF'
package main

import (
    "fmt"
    "yourproject/logger"
)

func main() {
    fmt.Println("Request ID field:", logger.FieldRequestID)
    fmt.Println("User ID field:", logger.FieldUserID)
    fmt.Println("Method field:", logger.FieldMethod)
}
EOF

go run test_fields.go
```
**Expected output**:
```
Request ID field: request_id
User ID field: user_id
Method field: method
```

### Test MaskEmail Function
```bash
# Create a test file
cat > test_mask.go << 'EOF'
package main

import (
    "fmt"
    "yourproject/logger"
)

func main() {
    // Test various email formats
    emails := []string{
        "user@example.com",
        "a@example.com",
        "ab@example.com",
        "verylongusername@example.com",
        "invalid-email",
        "",
    }
    
    for _, email := range emails {
        masked := logger.MaskEmail(email)
        fmt.Printf("%s → %s\n", email, masked)
    }
}
EOF

go run test_mask.go
```

**Expected output**:
```
user@example.com → u***r@example.com
a@example.com → ***@example.com
ab@example.com → ***@example.com
verylongusername@example.com → v***e@example.com
invalid-email → ***
(empty) → ***
```

### Test MaskToken Function
```bash
# Create a test file
cat > test_token.go << 'EOF'
package main

import (
    "fmt"
    "yourproject/logger"
)

func main() {
    tokens := []string{
        "short",
        "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0",
        "abc123def456ghi789",
        "",
    }
    
    for _, token := range tokens {
        masked := logger.MaskToken(token)
        fmt.Printf("%s → %s\n", token, masked)
    }
}
EOF

go run test_token.go
```

**Expected output**:
```
short → ****
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0 → eyJh****ODkwIn0
abc123def456ghi789 → abc1****i789
(empty) → ****
```

### Test with Actual Logger
```bash
# Test using constants with slog
cat > test_with_logger.go << 'EOF'
package main

import (
    "log/slog"
    "yourproject/logger"
)

func main() {
    log := logger.NewDefault()
    
    // Using constants
    log.Info("User action",
        slog.String(logger.FieldOperation, "create_task"),
        slog.String(logger.FieldUserID, "123"),
        slog.String(logger.FieldEmail, logger.MaskEmail("user@example.com")))
}
EOF

go run test_with_logger.go
```

**Expected**: Log output includes the fields with correct names and masked email

## Common Pitfalls

### Pitfall 1: Using PascalCase or camelCase for field names
**Symptom**: Log aggregation tools can't parse fields correctly
**Fix**: Always use lowercase with underscores: `request_id` not `requestId` or `RequestID`

### Pitfall 2: Not handling empty strings in masking functions
**Symptom**: Panic or unexpected behavior with empty inputs
**Fix**: Check for empty strings and return a safe value like "***"

### Pitfall 3: Masking too much or too little
**Symptom**: Either can't debug (too much masked) or privacy concerns (too little masked)
**Fix**: Follow the patterns - emails show first/last characters, tokens show first/last 4 characters

### Pitfall 4: Not validating email format
**Symptom**: Panic when splitting by "@" on invalid input
**Fix**: Check that split result has exactly 2 parts before accessing indices

### Pitfall 5: Forgetting to export constants
**Symptom**: Constants not accessible from other packages
**Fix**: Start constant names with uppercase letter: `FieldRequestID` not `fieldRequestID`

## Learning Resources

### Essential Reading
- [Go Constants](https://go.dev/tour/basics/15) - Understanding constant declarations in Go
- [Go Strings Package](https://pkg.go.dev/strings) - String manipulation functions
- [GDPR and Logging](https://gdpr.eu/what-is-personal-data/) - Understanding PII and privacy requirements

### Additional Resources
- [Log Field Naming Conventions](https://www.elastic.co/guide/en/ecs/current/ecs-field-reference.html) - Elastic Common Schema (ECS) used by ELK Stack
- [Grafana Loki Label Naming](https://grafana.com/docs/loki/latest/fundamentals/labels/) - Label naming requirements for Loki

## Real-World Context

### Why These Specific Field Names?

**request_id**: Standard in distributed systems for tracing requests across services. Used by OpenTelemetry, Datadog, and most APM tools.

**user_id**: Critical for security auditing and user-specific debugging. Always numeric IDs, never usernames or emails.

**duration_ms**: Milliseconds is the standard unit for request timing. Makes it easy to create latency percentile metrics (p50, p95, p99).

**status_code**: HTTP status codes are universal. Log aggregation tools have built-in dashboards for status code analysis.

**email** (masked): Needed for debugging user-specific issues, but must be masked for GDPR compliance.

### Industry Standards Alignment

These field names align with:
- **Elastic Common Schema (ECS)**: Used by ELK Stack
- **OpenTelemetry**: Standard for observability
- **Datadog**: APM and log management platform
- **Grafana Loki**: Label-based log aggregation

Using these conventions means your logs will work seamlessly with these tools without custom parsing.

## Testing Your Implementation

After implementing, verify:

1. **Constants are accessible**: Import logger package and access constants
2. **MaskEmail works correctly**: Test with various email formats
3. **MaskToken works correctly**: Test with various token lengths
4. **No compilation errors**: `go build ./logger` succeeds

## Next Steps

After completing this task, you'll have standardized field names and security functions. Task 3 will build on this by implementing context helpers that use these field names to store and retrieve request-scoped data like request IDs and trace IDs.

The combination of Tasks 1-2 gives you:
- Task 1: Logger factory and configuration
- Task 2: Standardized field names and masking functions
- Task 3: Will add context management for request correlation
