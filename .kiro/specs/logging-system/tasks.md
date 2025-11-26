# Implementation Plan

- [x] 1. Create logger package foundation
  - Create `logger/` package directory structure
  - Implement configuration types (Config struct with level, format, output, service name, environment)
  - Implement logger factory function that creates slog.Logger based on configuration
  - Implement helper functions for creating default and test loggers
  - Add support for JSON and text format handlers
  - Add support for stdout, stderr, and file output destinations
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.1, 2.2, 2.5, 8.1, 8.2_

- [x] 2. Implement standard fields and helper functions
  - Define standard field name constants (request_id, user_id, method, path, status_code, duration_ms, error, operation, task_id, email, trace_id, span_id)
  - Implement MaskEmail function for privacy protection
  - Implement MaskToken function for security
  - Implement helper functions for common log attributes
  - Ensure field naming follows log aggregation tool conventions (lowercase with underscores)
  - _Requirements: 1.1, 6.5, 10.5, 11.5_

- [x] 3. Implement context helpers for request correlation
  - Create context.go with functions for storing/retrieving request IDs
  - Implement WithRequestID and GetRequestID functions
  - Implement WithLogger and FromContext functions for logger propagation
  - Implement WithTraceID and GetTraceID functions for distributed tracing
  - Add request ID generation using UUID or similar
  - _Requirements: 3.3, 10.3, 11.1, 11.2, 11.4_

- [x] 4. Extend configuration system with logging options
  - Add LogConfig struct to cmd/server/config/config.go
  - Add logging configuration fields (level, format, output, add_source, service_name, environment)
  - Add file rotation configuration fields (enable_rotation, max_size, max_age, max_backups)
  - Set default values for logging configuration (level: info, format: json, output: stdout)
  - Add validation for logging configuration
  - Bind logging configuration to environment variables and command-line flags
  - Update ShowConfig to display logging configuration
  - _Requirements: 2.1, 2.2, 2.5, 8.1, 8.2, 8.3, 8.4, 10.2_

- [x] 5. Implement HTTP logging middleware
  - Create middleware.go with LoggingMiddleware function
  - Generate unique request ID for each HTTP request
  - Implement response writer wrapper to capture status code and bytes written
  - Log request start with method, path, user agent, and request ID
  - Log request completion with status code, duration, and bytes written
  - Extract and include user ID from context when available
  - Implement panic recovery with stack trace logging
  - Add request ID to request context for downstream logging
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 11.1, 11.2_

- [ ]* 6. Implement test logger utilities
  - Create logger_test.go with TestBuffer type
  - Implement NewTest function that returns logger and buffer
  - Implement buffer methods for capturing log entries (Entries, Contains, Reset)
  - Create LogEntry type with level, message, and attributes
  - Add thread-safe access to buffer using mutex
  - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_

- [x] 7. Integrate logger into server main
  - Update cmd/server/main.go to load logging configuration
  - Create logger instance using logger.New with configuration
  - Replace logRequest middleware with logger.LoggingMiddleware
  - Log server startup events (version, environment, configuration)
  - Log database initialization events
  - Log authentication system initialization
  - Log server listening address and available endpoints
  - Pass logger to storage, auth service, and handlers during initialization
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 3.1, 3.2, 10.1, 10.2, 11.3_

- [x] 8. Integrate logger into storage layer
  - Add logger field to DatabaseStorage struct
  - Update NewDatabaseStorage to accept logger parameter
  - Log database connection open event at INFO level
  - Log migration start, progress, and completion at INFO level
  - Log CRUD operations (CreateTask, UpdateTask, DeleteTask, GetTaskByID, LoadTasks) at DEBUG level
  - Log database errors at ERROR level with operation context and user ID
  - Include user ID in all database operation logs
  - Log RowsAffected for update and delete operations
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 6.1, 6.2, 6.3, 6.4_

- [x] 9. Integrate logger into auth service and middleware
  - Add logger field to auth.Service struct
  - Update auth.NewService to accept logger parameter
  - Log user registration attempts at INFO level with masked email
  - Log registration success/failure with appropriate level (INFO/WARN)
  - Log login attempts at INFO level with masked email
  - Log login success/failure with appropriate level (INFO/WARN)
  - Add logger field to auth.AuthMiddleware struct
  - Update auth.NewAuthMiddleware to accept logger parameter
  - Log missing Authorization header at WARN level
  - Log invalid token format at WARN level
  - Log JWT validation failures at WARN level with reason
  - Log authorization failures at WARN level with user ID and requested resource
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 6.1, 6.2, 6.5_

- [x] 10. Update HTTP handlers with structured logging
  - Create handler struct to hold logger and storage dependencies
  - Convert handler functions to methods on handler struct
  - Update tasksHandler to log request processing steps at DEBUG level
  - Log validation errors at WARN level with request context
  - Log task not found errors at WARN level with task ID and user ID
  - Log internal errors at ERROR level with full context and request ID
  - Update taskHandler with similar logging patterns
  - Update RegisterHandler to log registration flow
  - Update LoginHandler to log authentication flow
  - Include request ID from context in all handler logs
  - _Requirements: 3.4, 3.5, 6.1, 6.2, 6.3, 6.4, 11.2_

- [ ]* 11. Add CLI logging support (OPTIONAL - CLI already has excellent error handling)
  - Update cmd/cli/main.go to create logger with text format for stderr
  - Set default log level to ERROR for quiet CLI operation
  - Add --debug flag to enable DEBUG level logging
  - Log HTTP client connection errors at ERROR level
  - Log authentication errors at ERROR level
  - Log configuration loading errors at ERROR level
  - Ensure user-facing messages go to stdout without log formatting
  - Ensure internal errors and debug info go to stderr with structured format
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_
  - _Note: CLI already has excellent error handling with NetworkError/APIError types and clean user messages. Only implement if debugging HTTP/auth issues becomes necessary._

- [ ] 12. Implement log file rotation support
  - Add gopkg.in/natefinch/lumberjack.v2 dependency for log rotation
  - Update logger factory to use lumberjack when file output is configured
  - Configure lumberjack with max size, max age, and max backups from config
  - Implement file creation and directory creation if needed
  - Implement append mode for existing log files
  - _Requirements: 8.3, 8.4, 8.5_

- [ ] 13. Write comprehensive tests for logger package
  - Write tests for logger factory (New, NewDefault, NewTest)
  - Write tests for configuration validation
  - Write tests for log level filtering (verify DEBUG logs only appear when level is DEBUG)
  - Write tests for format selection (JSON vs text)
  - Write tests for output destination selection
  - Write tests for MaskEmail and MaskToken functions
  - Write tests for context helper functions (WithRequestID, GetRequestID, etc.)
  - Write tests for standard field helpers
  - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_

- [ ] 14. Write tests for HTTP logging middleware
  - Write tests for request ID generation and propagation
  - Write tests for request start logging
  - Write tests for request completion logging with status code and duration
  - Write tests for user ID extraction and inclusion in logs
  - Write tests for panic recovery and stack trace logging
  - Write tests for response writer wrapper (status code capture)
  - Verify log entries contain all expected fields
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 12.1, 12.2, 12.3_

- [ ] 15. Write integration tests for storage logging
  - Write tests verifying database operations are logged at DEBUG level
  - Write tests verifying errors are logged at ERROR level with context
  - Write tests verifying user ID is included in all operation logs
  - Write tests verifying migration events are logged at INFO level
  - Use TestBuffer to capture and assert log entries
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 12.1, 12.3_

- [ ] 16. Write integration tests for auth logging
  - Write tests verifying registration attempts are logged with masked email
  - Write tests verifying login attempts are logged with outcome
  - Write tests verifying authentication failures are logged at WARN level
  - Write tests verifying JWT validation failures are logged
  - Write tests verifying authorization failures are logged with context
  - Use TestBuffer to capture and assert log entries
  - Verify sensitive data (passwords, tokens) is never logged
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 6.5, 12.1, 12.3_

- [x] 17. Update documentation and examples
  - Update README.md with logging configuration section
  - Add example config.yaml with logging configuration
  - Document environment variables for logging
  - Document command-line flags for logging
  - Add examples of log output in JSON and text formats
  - Document integration with ELK Stack (include Logstash config example)
  - Document integration with Grafana Loki (include Promtail config example)
  - Document integration with Datadog (include agent config example)
  - Add troubleshooting section for common logging issues
  - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_
