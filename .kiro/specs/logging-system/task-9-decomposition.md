# Task Decomposition: Integrate Logger into Auth Service and Middleware

## Overview

Add structured logging to the authentication service and middleware to track security-critical events including user registration, login attempts, authentication failures, and JWT validation errors. This provides visibility into potential security incidents and helps with debugging authentication issues.

## Implementation Approach

We're adding a logger field to both `auth.Service` and `auth.AuthMiddleware` structs using dependency injection. The logger will be passed during initialization in `main.go` and used throughout the authentication flow to log security events.

**Complexity Check:**
- **Requirements need**: Log authentication events (registration, login, failures) with masked sensitive data for security monitoring
- **Simple approach**: Add logger field to structs, pass during initialization, add log statements at key points (15-20 min)
- **Complex approach**: Create separate audit logging system with event types, structured events, and separate storage (2+ hours)
- **Recommendation**: Start simple. Requirements only need visibility into auth events for monitoring and debugging. The simple approach meets all requirements. Complex audit system can be added later if compliance requires it.

**Key Concepts:**
- **Security Logging**: Track authentication events to detect suspicious activity (brute force, credential stuffing)
- **Data Masking**: Never log passwords; mask emails to protect user privacy while maintaining debuggability
- **Log Levels**: INFO for successful operations, WARN for expected failures (wrong password), ERROR for unexpected failures (storage errors)

## Prerequisites

**Existing Code:**
- `auth/service.go` - Authentication service with Register and Login methods
- `auth/middleware.go` - JWT authentication middleware with Authenticate method
- `logger/fields.go` - MaskEmail function for privacy protection
- `logger/logger.go` - Logger factory and configuration
- `cmd/server/main.go` - Server initialization where logger is created

**Dependencies:**
- `log/slog` package (already in use)
- Logger instance created in main.go (already exists)

**Knowledge Required:**
- Understanding of dependency injection pattern
- Go struct field addition and constructor updates
- slog structured logging API (slog.Info, slog.Warn, slog.Error with fields)
- Security logging best practices (never log passwords/tokens)

## Step-by-Step Instructions

### Step 1: Add logger field to auth.Service

**File**: `auth/service.go`

**What to do:**
Add a logger field to the Service struct and update the constructor to accept it.

**What to implement:**
- Add `logger *slog.Logger` field to the `Service` struct (after existing fields)
- Update `NewService` function signature to accept `logger *slog.Logger` as the third parameter
- Store the logger in the struct during initialization
- Required imports: `log/slog` (add to existing imports)

**Why:**
The Service needs access to a logger instance to log authentication events. Using dependency injection makes the service testable and follows the existing codebase pattern.

**Expected result:**
Code compiles. Service struct now has a logger field. NewService accepts logger parameter.

---

### Step 2: Add logging to Register method

**File**: `auth/service.go` (in the `Register` method)

**What to do:**
Add log statements at key points in the registration flow to track attempts and outcomes.

**What to implement:**
- Log registration attempt at INFO level at the start of the method
  - Include masked email using `logger.MaskEmail(email)`
  - Use field name `logger.FieldEmail` for consistency
  - Use field name `logger.FieldOperation` with value "user_registration"
- Log successful registration at INFO level after token generation
  - Include masked email and the new user ID
  - Use message "User registered successfully"
- Log registration failures at WARN level for expected errors (email exists, validation failures)
  - Include masked email and error message
  - Use field name `logger.FieldError` for the error
- Log registration failures at ERROR level for unexpected errors (storage failures, hashing failures)
  - Include masked email, error message, and operation context
  - Use message "Registration failed"

**Why:**
Registration logging helps detect abuse patterns (spam registrations, bot activity) and debug registration issues. We log at INFO for normal operations, WARN for expected failures (user errors), and ERROR for system failures.

**Expected result:**
Registration attempts and outcomes are logged with appropriate levels and masked email addresses.

---

### Step 3: Add logging to Login method

**File**: `auth/service.go` (in the `Login` method)

**What to do:**
Add log statements to track login attempts and authentication outcomes.

**What to implement:**
- Log login attempt at INFO level at the start of the method
  - Include masked email
  - Use field name `logger.FieldOperation` with value "user_login"
  - Use message "Login attempt"
- Log successful login at INFO level after token generation
  - Include masked email and user ID
  - Use message "Login successful"
- Log failed login at WARN level when credentials are invalid
  - Include masked email (do NOT include the reason - don't reveal if email exists)
  - Use message "Login failed"
  - Use field name `logger.FieldError` with value "invalid credentials"
- Log storage errors at ERROR level
  - Include masked email and error details
  - Use message "Login failed due to storage error"

**Why:**
Login logging is critical for security monitoring. Failed login attempts can indicate brute force attacks or credential stuffing. We use WARN for authentication failures (expected) and ERROR for system failures (unexpected). We never reveal whether the email exists to prevent user enumeration.

**Expected result:**
Login attempts and outcomes are logged with appropriate security considerations.

---

### Step 4: Add logger field to auth.AuthMiddleware

**File**: `auth/middleware.go`

**What to do:**
Add a logger field to the AuthMiddleware struct and update the constructor.

**What to implement:**
- Add `logger *slog.Logger` field to the `AuthMiddleware` struct (after jwtService field)
- Update `NewAuthMiddleware` function signature to accept `logger *slog.Logger` as the second parameter
- Store the logger in the struct during initialization
- Required imports: `log/slog` and `myproject/logger` (add to existing imports)

**Why:**
The middleware needs to log authentication failures at the HTTP layer (missing tokens, invalid tokens, expired tokens). This is separate from the service layer logging and provides visibility into API security.

**Expected result:**
Code compiles. AuthMiddleware struct has logger field. NewAuthMiddleware accepts logger parameter.

---

### Step 5: Add logging to Authenticate middleware

**File**: `auth/middleware.go` (in the `Authenticate` method)

**What to do:**
Add log statements for authentication failures at the middleware level.

**What to implement:**
- Log missing Authorization header at WARN level
  - Include request path and method from the request
  - Use field names `logger.FieldPath` and `logger.FieldMethod`
  - Include request ID from context using `logger.GetRequestID(r.Context())`
  - Use field name `logger.FieldRequestID`
  - Use message "Missing authorization header"
- Log invalid token format at WARN level (when ExtractToken fails)
  - Include request path, method, request ID, and error message
  - Use message "Invalid authorization header format"
- Log JWT validation failures at WARN level (when ValidateToken fails)
  - Include request path, method, request ID, and error reason
  - Use message "JWT validation failed"
  - Use field name `logger.FieldError` for the error
- Log successful authentication at DEBUG level (optional, for detailed tracing)
  - Include user ID, request path, method, and request ID
  - Use message "Authentication successful"

**Why:**
Middleware logging captures authentication failures at the HTTP layer before they reach handlers. This helps identify API abuse, token manipulation attempts, and expired token issues. We use WARN level because these are expected security events (not system errors).

**Expected result:**
Authentication failures are logged with request context for debugging and security monitoring.

---

### Step 6: Update main.go to pass logger to auth components

**File**: `cmd/server/main.go`

**What to do:**
Update the initialization of auth service and middleware to pass the logger instance.

**What to implement:**
- Locate the line where `authService` is created with `auth.NewService`
- Add `l` (the logger variable) as the third parameter to `auth.NewService`
- Locate the line where `authMiddleware` is created with `auth.NewAuthMiddleware`
- Add `l` (the logger variable) as the second parameter to `auth.NewAuthMiddleware`
- No new imports needed (logger is already created and available)

**Why:**
This wires up the dependency injection, providing the logger instance to both auth components. The logger is already created earlier in main.go, so we just need to pass it through.

**Expected result:**
Code compiles. Auth service and middleware now have access to the logger for structured logging.

---

## Verification

### Compile Check
```bash
go build ./...
```
**Expected**: No compilation errors

### Run Tests
```bash
# Test auth package
go test ./auth -v

# If tests fail due to missing logger parameter, update test code:
# - Add logger parameter to NewService and NewAuthMiddleware calls in tests
# - Use logger.NewDefault() for test logger
```
**Expected**: All tests pass (may need minor test updates)

### Manual Testing

1. Start the server:
```bash
go run cmd/server/main.go
```

2. Test registration with logging:
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```
**Expected logs**: 
- INFO: "Login attempt" with masked email
- INFO: "User registered successfully" with masked email and user ID

3. Test login with correct credentials:
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```
**Expected logs**:
- INFO: "Login attempt" with masked email
- INFO: "Login successful" with masked email and user ID

4. Test login with wrong password:
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"wrongpassword"}'
```
**Expected logs**:
- INFO: "Login attempt" with masked email
- WARN: "Login failed" with masked email

5. Test protected endpoint without token:
```bash
curl http://localhost:8080/tasks
```
**Expected logs**:
- WARN: "Missing authorization header" with request path and method

6. Test protected endpoint with invalid token:
```bash
curl http://localhost:8080/tasks \
  -H "Authorization: Bearer invalid_token_here"
```
**Expected logs**:
- WARN: "JWT validation failed" with error details

7. Check log format:
```bash
# If using JSON format, logs should be valid JSON
# If using text format, logs should be human-readable with key=value pairs
```

### Verify Security Requirements

1. **Check that passwords are NEVER logged**:
```bash
# Search logs for password values - should find NOTHING
grep -i "password123" /path/to/logs  # Should return no results
```

2. **Check that emails are masked**:
```bash
# Logs should show "t***t@example.com" not "test@example.com"
grep "test@example.com" /path/to/logs  # Should return no results
grep "t\*\*\*t@example.com" /path/to/logs  # Should find masked emails
```

3. **Check that tokens are not logged in full**:
```bash
# JWT tokens should not appear in logs
# Only validation failures should be logged, not token contents
```

## Common Pitfalls

### Pitfall 1: Logging passwords or tokens
**Symptom**: Sensitive data appears in log files
**Fix**: 
- Never log the `password` parameter
- Never log JWT token strings (only log validation failures)
- Always use `logger.MaskEmail()` for email addresses
- Use `logger.MaskToken()` if you need to log token references

### Pitfall 2: Revealing user existence in logs
**Symptom**: Logs show "user not found" vs "invalid password"
**Fix**: 
- Always log generic "invalid credentials" for login failures
- Don't differentiate between "email doesn't exist" and "wrong password" in logs
- This prevents user enumeration attacks

### Pitfall 3: Wrong log levels
**Symptom**: Too many ERROR logs for normal authentication failures
**Fix**:
- Use INFO for successful operations (registration, login)
- Use WARN for expected failures (wrong password, missing token, invalid token)
- Use ERROR only for unexpected system failures (storage errors, hashing failures)
- Use DEBUG for detailed tracing (optional, only when debugging)

### Pitfall 4: Missing request context
**Symptom**: Can't correlate logs with specific HTTP requests
**Fix**:
- Always include request ID using `logger.GetRequestID(r.Context())`
- Include request path and method for middleware logs
- Use consistent field names (`logger.FieldRequestID`, `logger.FieldPath`, `logger.FieldMethod`)

### Pitfall 5: Forgetting to update tests
**Symptom**: Tests fail with "not enough arguments" errors
**Fix**:
- Update all test files that create `auth.Service` or `auth.AuthMiddleware`
- Add logger parameter: `logger.NewDefault()` for simple tests
- For tests that verify logging, use `logger.NewTest()` to capture log output

## Learning Resources

### Essential Reading
- [Go slog Package Documentation](https://pkg.go.dev/log/slog) - Official documentation for structured logging
- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html) - Security best practices for logging authentication events
- [Go Dependency Injection Patterns](https://blog.drewolson.org/dependency-injection-in-go) - Understanding constructor injection pattern

### Additional Resources (Optional)
- [Security Logging and Monitoring Guide](https://owasp.org/www-project-proactive-controls/v3/en/c9-security-logging) - OWASP guide on security event logging
- [Structured Logging Best Practices](https://www.honeycomb.io/blog/structured-logging-and-your-team) - Why structured logging matters for observability
