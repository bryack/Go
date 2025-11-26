# Requirements Document

## Introduction

This feature refactors the TaskManager from a stateful in-memory cache to a stateless service layer that delegates all persistence to the Storage interface. The current implementation maintains duplicate state between TaskManager and the database, leading to synchronization issues and architectural inconsistencies between the CLI and server components. This refactoring will establish the database as the single source of truth while preserving TaskManager's unique business logic capabilities (formatting, parallel processing).

## Requirements

### Requirement 1: Remove State Management from TaskManager

**User Story:** As a developer, I want TaskManager to be stateless, so that there is only one source of truth for task data and no synchronization issues.

#### Acceptance Criteria

1. WHEN TaskManager is initialized THEN it SHALL NOT maintain an internal tasks slice
2. WHEN TaskManager is initialized THEN it SHALL NOT use a mutex for state synchronization
3. WHEN any TaskManager method is called THEN it SHALL delegate data retrieval to the Storage interface
4. IF TaskManager needs task data THEN it SHALL fetch it from Storage on-demand

### Requirement 2: Preserve Business Logic and Formatting Capabilities

**User Story:** As a developer, I want TaskManager to retain its formatting and processing capabilities, so that business logic remains separate from persistence concerns.

#### Acceptance Criteria

1. WHEN FormatTask is called THEN it SHALL format a single task with status and description
2. WHEN PrintTasks is called THEN it SHALL retrieve tasks from Storage and format them for display
3. WHEN GetFormattedTasks is called THEN it SHALL retrieve tasks from Storage and return formatted string
4. WHEN ProcessTasks is called THEN it SHALL retrieve tasks from Storage and process them in parallel
5. WHEN any formatting method is called THEN it SHALL NOT maintain or cache task state

### Requirement 3: Update CRUD Operations to Use Storage Directly

**User Story:** As a developer, I want all CRUD operations to work directly with Storage, so that data persistence is consistent and reliable.

#### Acceptance Criteria

1. WHEN UpdateTaskStatus is called THEN it SHALL fetch the task from Storage, update it, and save via Storage
2. WHEN UpdateTaskDescription is called THEN it SHALL fetch the task from Storage, update it, and save via Storage
3. WHEN ClearDescription is called THEN it SHALL fetch the task from Storage, clear description, and save via Storage
4. WHEN DeleteTask is called THEN it SHALL delegate deletion to Storage
5. WHEN GetTaskByID is called THEN it SHALL delegate retrieval to Storage
6. WHEN any CRUD operation completes THEN it SHALL NOT update any in-memory cache

### Requirement 4: Refactor AddTask Methods

**User Story:** As a developer, I want task creation to be simplified and consistent, so that there's no confusion about ID assignment and state management.

#### Acceptance Criteria

1. WHEN AddTask is called with a description THEN it SHALL create the task in Storage and return the assigned ID
2. WHEN AddTask completes THEN it SHALL NOT maintain the task in memory
3. WHEN AddTask is called THEN the AddTaskWithID method SHALL be removed as it's no longer needed
4. IF task creation fails THEN it SHALL return an appropriate error

### Requirement 5: Remove Obsolete State Management Methods

**User Story:** As a developer, I want to remove methods that manage in-memory state, so that the API is clean and doesn't expose unnecessary complexity.

#### Acceptance Criteria

1. WHEN refactoring is complete THEN GetTasks method SHALL be removed
2. WHEN refactoring is complete THEN SetTasks method SHALL be removed
3. WHEN refactoring is complete THEN AddTaskWithID method SHALL be removed
4. IF any code depends on these methods THEN it SHALL be refactored to use Storage directly

### Requirement 6: Update TaskManager Constructor

**User Story:** As a developer, I want TaskManager to accept Storage as a dependency, so that it can delegate persistence operations.

#### Acceptance Criteria

1. WHEN NewTaskManager is called THEN it SHALL accept Storage interface as a parameter
2. WHEN NewTaskManager is called THEN it SHALL accept io.Writer for output as a parameter
3. WHEN NewTaskManager is initialized THEN it SHALL store the Storage reference for method use
4. WHEN NewTaskManager is initialized THEN it SHALL NOT initialize a tasks slice or mutex

### Requirement 7: Update CLI to Work with Refactored TaskManager

**User Story:** As a CLI user, I want the CLI to work seamlessly with the refactored TaskManager, so that all commands function correctly without duplicate state management.

#### Acceptance Criteria

1. WHEN CLI initializes THEN it SHALL pass Storage to TaskManager constructor
2. WHEN handleAddCommand is called THEN it SHALL use TaskManager.AddTask without manual ID assignment
3. WHEN handleListCommand is called THEN it SHALL use TaskManager.PrintTasks which fetches from Storage
4. WHEN handleStatusCommand is called THEN it SHALL use TaskManager.UpdateTaskStatus which updates Storage
5. WHEN handleUpdateCommand is called THEN it SHALL use TaskManager.UpdateTaskDescription which updates Storage
6. WHEN handleDeleteCommand is called THEN it SHALL use TaskManager.DeleteTask which updates Storage
7. WHEN handleClearCommand is called THEN it SHALL use TaskManager.ClearDescription which updates Storage
8. WHEN exit command is called THEN it SHALL NOT call SaveTasks as all changes are already persisted
9. IF any CLI command fails THEN it SHALL display appropriate error messages

### Requirement 8: Update Tests for Stateless TaskManager

**User Story:** As a developer, I want tests to validate the stateless behavior of TaskManager, so that the refactoring is properly verified.

#### Acceptance Criteria

1. WHEN tests are updated THEN they SHALL mock the Storage interface
2. WHEN testing CRUD operations THEN tests SHALL verify Storage methods are called correctly
3. WHEN testing formatting methods THEN tests SHALL verify correct output without state assumptions
4. WHEN testing ProcessTasks THEN tests SHALL verify parallel processing works with Storage-fetched data
5. IF any test relies on GetTasks or SetTasks THEN it SHALL be refactored or removed

### Requirement 9: Maintain Backward Compatibility for Public API

**User Story:** As a developer using TaskManager, I want the public API to remain intuitive, so that the refactoring doesn't break expected behavior.

#### Acceptance Criteria

1. WHEN FormatTask is called THEN it SHALL maintain the same signature and behavior
2. WHEN ProcessTasks is called THEN it SHALL maintain the same parallel processing behavior
3. WHEN error handling occurs THEN it SHALL return the same error types (ErrTaskNotFound, etc.)
4. IF a method signature changes THEN it SHALL be documented in the refactoring plan

### Requirement 10: Ensure Thread Safety Through Storage Layer

**User Story:** As a developer, I want concurrent operations to be safe, so that the application doesn't have race conditions.

#### Acceptance Criteria

1. WHEN multiple goroutines call TaskManager methods THEN Storage SHALL handle concurrency safely
2. WHEN ProcessTasks runs in parallel THEN it SHALL fetch a snapshot of tasks from Storage
3. WHEN concurrent updates occur THEN Storage layer SHALL ensure data integrity
4. IF TaskManager needs synchronization THEN it SHALL rely on Storage's thread-safety mechanisms
