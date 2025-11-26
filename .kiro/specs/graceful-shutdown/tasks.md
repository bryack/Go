# Implementation Plan

- [x] 1. Add shutdown timeout configuration
  - Add `ShutdownTimeout` field to `ServerConfig` struct in `cmd/server/config/config.go`
  - Set default value to 30 seconds in config loading logic
  - Add YAML binding for `shutdown_timeout` key
  - Add validation to ensure timeout is positive duration
  - Log the configured shutdown timeout value at server startup
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [x] 2. Refactor server initialization to use explicit http.Server
  - Replace `http.ListenAndServe()` call with explicit `http.Server` instance in `cmd/server/main.go`
  - Configure server with address and handler (DefaultServeMux)
  - Call `server.ListenAndServe()` in main goroutine
  - Ensure server starts successfully and all existing endpoints work
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 2.5. Configure HTTP server timeouts for reliable shutdown
  - Add `ReadTimeout: 15 * time.Second` to prevent slow-read attacks and ensure request completion
  - Add `WriteTimeout: 15 * time.Second` to prevent slow-write issues and ensure response completion
  - Add `IdleTimeout: 2 * time.Second` to close idle keep-alive connections quickly
  - Ensure IdleTimeout is shorter than shutdown timeout to allow graceful shutdown to complete
  - Verify server still handles requests correctly with timeouts configured
  - _Requirements: 1.2, 1.3, 1.4 (enables server.Shutdown() to complete within timeout)_

- [x] 3. Implement signal handler goroutine
  - Import `os/signal` and `syscall` packages
  - Create buffered channel for signal notifications
  - Register signal handler for SIGINT and SIGTERM using `signal.Notify()`
  - Launch goroutine that blocks on signal channel
  - Log shutdown initiation when signal is received with signal type
  - Handle second signal as force quit (immediate exit with code 1)
  - _Requirements: 1.1, 1.5, 3.1, 5.4_

- [x] 4. Implement graceful shutdown sequence
  - Create context with timeout using configured `ShutdownTimeout` value
  - Call `server.Shutdown(ctx)` to stop accepting new requests and wait for in-flight requests
  - Handle context deadline exceeded (timeout) case
  - Log shutdown completion with duration and status
  - Handle errors from `server.Shutdown()` and log appropriately
  - _Requirements: 1.2, 1.3, 1.4, 1.5, 3.2, 3.3_

- [x] 5. Implement database cleanup during shutdown
  - Call `storage.Close()` after server shutdown completes
  - Log database close operation (success or error)
  - Ensure database close is called even if server shutdown times out
  - Handle and log any errors from database close operation
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 3.4_

- [x] 6. Implement proper exit code handling
  - Exit with code 0 on successful graceful shutdown
  - Exit with code 1 if shutdown times out
  - Exit with code 1 if database close fails
  - Exit with code 1 if server shutdown returns error
  - Exit with code 1 on second signal (force quit)
  - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ]* 7. Add unit tests for shutdown behavior
  - Test signal handler captures SIGINT correctly
  - Test signal handler captures SIGTERM correctly
  - Test second signal forces immediate exit
  - Test shutdown timeout is respected
  - Test database close is called during shutdown
  - Test proper exit codes for different scenarios
  - _Requirements: All requirements_

- [ ]* 8. Add integration tests for graceful shutdown
  - Test server stops accepting new requests during shutdown
  - Test in-flight requests complete before shutdown
  - Test shutdown completes within timeout period
  - Test database connections are closed properly
  - _Requirements: 1.2, 1.3, 2.1, 2.2_
