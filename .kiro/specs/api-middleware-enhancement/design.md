# API Middleware Enhancement Design

## Overview

This design implements a comprehensive middleware system for the to-do list API, building upon the existing basic logging middleware. The solution follows Go's standard middleware pattern using function closures and the http.Handler interface, ensuring compatibility with the existing net/http server implementation.

The middleware system will be modular, allowing individual middleware components to be enabled/disabled and configured independently. Each middleware will follow the chain-of-responsibility pattern, where requests pass through multiple middleware layers before reaching the final handler.

## Architecture

### Middleware Chain Structure

```
Request → CORS → Rate Limiting → Request Logging → Timeout → Size Limiting → Error Recovery → Handler
```

### Core Components

1. **Middleware Interface**: A standardized way to define and chain middleware
2. **Configuration System**: Centralized configuration for all middleware components
3. **Enhanced Logging**: Structured logging with configurable levels
4. **Error Handling**: Standardized error response format and recovery mechanisms

### Package Structure

```
internal/
├── middleware/
│   ├── middleware.go      # Core middleware types and chaining
│   ├── cors.go           # CORS middleware implementation
│   ├── ratelimit.go      # Rate limiting middleware
│   ├── logging.go        # Enhanced request logging
│   ├── timeout.go        # Request timeout handling
│   ├── sizelimit.go      # Request size limiting
│   ├── recovery.go       # Panic recovery middleware
│   └── config.go         # Middleware configuration
└── errors/
    ├── errors.go         # Standardized error types
    └── responses.go      # Error response formatting
```

## Components and Interfaces

### Middleware Interface

```go
type Middleware func(http.Handler) http.Handler

type MiddlewareChain struct {
    middlewares []Middleware
}

func (mc *MiddlewareChain) Use(middleware Middleware) *MiddlewareChain
func (mc *MiddlewareChain) Handler(handler http.Handler) http.Handler
```

### Configuration Structure

```go
type Config struct {
    CORS        CORSConfig
    RateLimit   RateLimitConfig
    Logging     LoggingConfig
    Timeout     TimeoutConfig
    SizeLimit   SizeLimitConfig
    Recovery    RecoveryConfig
}

type CORSConfig struct {
    Enabled        bool
    AllowedOrigins []string
    AllowedMethods []string
    AllowedHeaders []string
    MaxAge         int
}

type RateLimitConfig struct {
    Enabled       bool
    RequestsPerIP int
    WindowSize    time.Duration
}
```

### Error Response Format

```go
type ErrorResponse struct {
    Error     string                 `json:"error"`
    Code      string                 `json:"code"`
    Message   string                 `json:"message"`
    Timestamp time.Time              `json:"timestamp"`
    Details   map[string]interface{} `json:"details,omitempty"`
}
```

## Data Models

### Rate Limiting Storage

The rate limiter will use an in-memory store with cleanup mechanisms:

```go
type RateLimiter struct {
    requests map[string][]time.Time
    mutex    sync.RWMutex
    limit    int
    window   time.Duration
}
```

### Request Context Enhancement

Middleware will add contextual information to requests:

```go
type RequestContext struct {
    RequestID   string
    StartTime   time.Time
    ClientIP    string
    UserAgent   string
}
```

## Error Handling

### Panic Recovery

A recovery middleware will catch panics and convert them to proper HTTP error responses:

- Log the panic with stack trace
- Return HTTP 500 with generic error message
- Ensure the server continues running

### Error Response Standardization

All middleware will use a centralized error response system:

- Consistent JSON error format across all endpoints
- Appropriate HTTP status codes
- Sanitized error messages for security
- Detailed logging for debugging

### Validation Error Enhancement

Extend existing validation to provide structured error responses:

```go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   string `json:"value,omitempty"`
}
```

## Testing Strategy

### Unit Testing Approach

1. **Individual Middleware Testing**: Each middleware component will have comprehensive unit tests
   - Test normal operation and edge cases
   - Mock HTTP requests and responses
   - Verify proper header setting and status codes

2. **Middleware Chain Testing**: Test the interaction between multiple middleware components
   - Verify proper request/response flow through the chain
   - Test middleware ordering and dependencies

3. **Configuration Testing**: Validate configuration parsing and validation
   - Test invalid configurations
   - Test default value handling

### Integration Testing

1. **End-to-End API Testing**: Test complete request flows through all middleware
   - Test CORS preflight and actual requests
   - Test rate limiting behavior under load
   - Test timeout scenarios

2. **Error Scenario Testing**: Verify proper error handling across all middleware
   - Test panic recovery
   - Test various error conditions
   - Verify error response format consistency

### Performance Testing

1. **Middleware Overhead**: Measure performance impact of middleware chain
2. **Rate Limiting Performance**: Test rate limiter under concurrent load
3. **Memory Usage**: Monitor memory consumption of rate limiting and logging

## Implementation Considerations

### Backward Compatibility

The new middleware system will be designed to replace the existing `logRequest` middleware without breaking changes to the main server code.

### Configuration Management

Middleware configuration will be loaded from environment variables with sensible defaults, allowing easy deployment configuration without code changes.

### Monitoring and Observability

Enhanced logging will provide structured output suitable for log aggregation systems, including:
- Request/response metrics
- Error rates and types
- Rate limiting statistics
- Performance metrics

### Security Considerations

- Rate limiting prevents DoS attacks
- Request size limiting prevents memory exhaustion
- CORS configuration prevents unauthorized cross-origin requests
- Error responses don't expose internal system details
- All user inputs are properly validated and sanitized