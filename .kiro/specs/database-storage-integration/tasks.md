# Implementation Plan

- [x] 1. Set up database foundation and dependencies
  - Add SQLite driver dependency (`modernc.io/sqlite`) to go.mod
  - Create storage/errors.go with database-specific error types
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [ ] 2. Implement database connection management
  - [ ] 2.1 Create connection.go with ConnectionManager and ConnectionConfig structs
    - Write database connection establishment with proper configuration
    - Implement connection pool settings (MaxOpenConns, MaxIdleConns, timeouts)
    - _Requirements: 2.1, 2.2, 2.4, 2.5_

  - [ ] 2.2 Add connection retry logic with exponential backoff
    - Implement retry mechanism for connection failures
    - Add timeout handling and graceful error reporting
    - _Requirements: 2.3, 4.1_

- [ ] 3. Create database schema and migration system
  - [x] 3.1 Design and implement migrations.go with Migration struct and Migrator
    - Write Migration struct with Version, Name, Up, Down fields
    - Create Migrator with database reference and migration slice
    - _Requirements: 3.1, 3.2, 3.4_

  - [ ] 3.2 Implement initial schema migration (version 1)
    - Write SQL for tasks table creation with proper indexes
    - Create schema_migrations table for tracking applied migrations
    - _Requirements: 1.2, 3.4, 5.4_

  - [ ] 3.3 Add migration execution logic with rollback support
    - Implement migration application with transaction safety
    - Add rollback functionality for failed migrations
    - _Requirements: 3.2, 3.3_

- [ ] 4. Implement DatabaseStorage struct and core operations
  - [ ] 4.1 Create database.go with DatabaseStorage struct implementing Storage interface
    - Write NewDatabaseStorage constructor with database initialization
    - Implement automatic database and schema creation on startup
    - _Requirements: 1.1, 1.2, 2.1_

  - [ ] 4.2 Implement LoadTasks() method with proper SQL queries
    - Write SELECT query to retrieve all tasks from database
    - Map database rows to task.Task structs maintaining existing format
    - _Requirements: 5.1, 5.4_

  - [ ] 4.3 Implement SaveTasks() method with transaction handling
    - Write transaction-based replacement of all tasks (DELETE + INSERT)
    - Handle batch operations efficiently with prepared statements
    - _Requirements: 5.2, 5.5_

  - [ ] 4.4 Add Close() method for proper resource cleanup
    - Implement database connection closing and resource cleanup
    - _Requirements: 2.4_

- [ ] 5. Integrate database storage with existing application
  - [ ] 5.1 Update main.go files to use DatabaseStorage instead of JsonStorage
    - Modify cmd/server/main.go to initialize DatabaseStorage
    - Update cmd/cli/main.go to use new database storage
    - _Requirements: 1.4, 5.3_

  - [ ] 5.2 Add configuration support for database path
    - Implement environment variable support for database file location
    - Add default database path handling
    - _Requirements: 1.1_

  - [ ] 5.3 Implement data migration from JSON to database
    - Create migration utility to convert existing tasks.json to database
    - Add automatic detection and migration of existing JSON data
    - _Requirements: 5.5_

- [ ] 6. Add comprehensive error handling and logging
  - [ ] 6.1 Implement database-specific error classification and handling
    - Map SQLite errors to custom error types
    - Add specific handling for constraint violations, lock timeouts, disk full
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [ ] 6.2 Add retry logic for database lock and busy scenarios
    - Implement retry mechanism for database busy/locked conditions
    - Add appropriate timeouts and backoff strategies
    - _Requirements: 4.3_

- [ ]* 7. Create comprehensive test suite
  - [ ]* 7.1 Write unit tests for DatabaseStorage operations
    - Test LoadTasks and SaveTasks with various data scenarios
    - Test error conditions and edge cases
    - _Requirements: 5.1, 5.2, 4.1_

  - [ ]* 7.2 Write integration tests for migration system
    - Test schema creation and migration application
    - Test rollback scenarios and error handling
    - _Requirements: 3.1, 3.2, 3.3_

  - [ ]* 7.3 Write concurrency tests for database operations
    - Test multiple goroutines accessing database simultaneously
    - Verify connection pool behavior under load
    - _Requirements: 1.3, 2.5_

  - [ ]* 7.4 Write performance comparison tests
    - Compare database storage performance with JSON storage
    - Test with various data sizes and concurrent access patterns
    - _Requirements: 1.3_