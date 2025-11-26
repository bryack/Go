# Implementation Plan

- [ ] 1. Create MockStorage for testing
  - Create `storage/mock_storage.go` file with MockStorage struct
  - Implement all Storage interface methods (CreateTask, GetTaskByID, UpdateTask, DeleteTask, LoadTasks, SaveTasks)
  - Use in-memory map for task storage with auto-incrementing ID
  - _Requirements: 8.1, 8.2_

- [ ] 2. Refactor TaskManager struct and constructor
  - [ ] 2.1 Update TaskManager struct definition
    - Remove `tasks []Task` field from struct
    - Remove `mu sync.Mutex` field from struct
    - Add `storage storage.Storage` field to struct
    - Keep `writer io.Writer` field
    - _Requirements: 1.1, 1.2, 6.3_

  - [ ] 2.2 Update NewTaskManager constructor
    - Change signature to accept `storage storage.Storage` as first parameter
    - Keep `writer io.Writer` as second parameter
    - Initialize TaskManager with storage and writer
    - Remove initialization of tasks slice and mutex
    - _Requirements: 6.1, 6.2, 6.4_

- [ ] 3. Refactor CRUD methods to delegate to Storage
  - [ ] 3.1 Refactor AddTask method
    - Change signature from `AddTask(input string) Task` to `AddTask(input string) (int, error)`
    - Create Task struct with description and done=false
    - Call `storage.CreateTask()` and return ID and error
    - Remove any internal state management
    - _Requirements: 3.1, 3.2, 4.1, 4.2, 4.3, 4.4_

  - [ ] 3.2 Refactor UpdateTaskStatus method
    - Fetch task using `storage.GetTaskByID(id)`
    - Update task.Done field
    - Save using `storage.UpdateTask(task)`
    - Return error if any operation fails
    - Remove mutex locks and internal state access
    - _Requirements: 3.1, 3.6_

  - [ ] 3.3 Refactor UpdateTaskDescription method
    - Fetch task using `storage.GetTaskByID(id)`
    - Update task.Description field
    - Save using `storage.UpdateTask(task)`
    - Return error if any operation fails
    - Remove mutex locks and internal state access
    - _Requirements: 3.1, 3.6_

  - [ ] 3.4 Refactor ClearDescription method
    - Fetch task using `storage.GetTaskByID(id)`
    - Set task.Description to empty string
    - Save using `storage.UpdateTask(task)`
    - Return error if any operation fails
    - Remove mutex locks and internal state access
    - _Requirements: 3.1, 3.6_

  - [ ] 3.5 Refactor DeleteTask method
    - Change implementation to call `storage.DeleteTask(id)` directly
    - Return error from storage operation
    - Remove mutex locks and internal state management
    - _Requirements: 3.5, 3.6_

  - [ ] 3.6 Refactor GetTaskByID method
    - Change implementation to call `storage.GetTaskByID(id)` directly
    - Return task and error from storage operation
    - Remove mutex locks and internal state access
    - _Requirements: 3.5, 3.6_

- [ ] 4. Refactor formatting and display methods
  - [ ] 4.1 Refactor PrintTasks method
    - Call `storage.LoadTasks()` to fetch tasks
    - Handle error from LoadTasks
    - Pass tasks to `printToWriter()` function
    - Remove mutex locks and internal state access
    - _Requirements: 2.2, 2.5_

  - [ ] 4.2 Refactor GetFormattedTasks method
    - Call `storage.LoadTasks()` to fetch tasks
    - Handle error by returning error message string
    - Pass tasks to `formatTasks()` function
    - Remove call to GetTasks() method
    - _Requirements: 2.3, 2.5_

  - [ ] 4.3 Keep FormatTask function unchanged
    - Verify FormatTask is a pure function with no state dependencies
    - No changes needed
    - _Requirements: 2.1_

- [ ] 5. Refactor ProcessTasks method
  - Call `storage.LoadTasks()` to fetch tasks snapshot
  - Handle error from LoadTasks by printing error message
  - Check if tasks slice is empty and print appropriate message
  - Process tasks in parallel using existing goroutine logic
  - Remove call to GetTasks() method
  - _Requirements: 2.4, 2.5, 10.2_

- [ ] 6. Remove obsolete state management methods
  - Remove GetTasks() method entirely
  - Remove SetTasks() method entirely
  - Remove AddTaskWithID() method entirely
  - _Requirements: 5.1, 5.2, 5.3_

- [ ] 7. Update CLI handleAddCommand
  - Remove call to `cli.storage.CreateTask()`
  - Remove call to `cli.taskManager.AddTaskWithID()`
  - Change to call `cli.taskManager.AddTask(desc)` which returns (id, error)
  - Handle error from AddTask
  - Print success message with returned ID
  - _Requirements: 7.2, 7.9_

- [ ] 8. Update CLI CommandList handler
  - Remove call to `cli.storage.LoadTasks()`
  - Remove call to `cli.taskManager.SetTasks()`
  - Change to call `cli.taskManager.PrintTasks()` directly
  - Handle error from PrintTasks
  - _Requirements: 7.3, 7.9_

- [ ] 9. Update CLI CommandExit handler
  - Remove call to `cli.storage.SaveTasks()`
  - Remove "Tasks saved successfully!" message
  - Keep only "👋 Bye!" message and return
  - _Requirements: 7.8_

- [ ] 10. Update CLI main.go initialization
  - Update TaskManager initialization to pass storage as first parameter
  - Change from `task.NewTaskManager(os.Stdout)` to `task.NewTaskManager(s, os.Stdout)`
  - Keep CLI initialization unchanged (still receives both TaskManager and Storage)
  - _Requirements: 7.1_

- [ ] 11. Update TaskManager tests with MockStorage
  - [ ] 11.1 Update TestAddTask
    - Create MockStorage instance
    - Pass MockStorage to NewTaskManager
    - Update test to expect (id, error) return from AddTask
    - Verify task was created in MockStorage
    - _Requirements: 8.1, 8.2_

  - [ ] 11.2 Update TestUpdateTaskStatus
    - Create MockStorage instance and pre-populate with test task
    - Pass MockStorage to NewTaskManager
    - Call UpdateTaskStatus and verify task updated in MockStorage
    - Remove any SetTasks calls
    - _Requirements: 8.1, 8.2_

  - [ ] 11.3 Update TestUpdateTaskDescription
    - Create MockStorage instance and pre-populate with test task
    - Pass MockStorage to NewTaskManager
    - Call UpdateTaskDescription and verify task updated in MockStorage
    - Remove any SetTasks calls
    - _Requirements: 8.1, 8.2_

  - [ ] 11.4 Update TestClearDescription
    - Create MockStorage instance and pre-populate with test task
    - Pass MockStorage to NewTaskManager
    - Call ClearDescription and verify description cleared in MockStorage
    - Remove any SetTasks calls
    - _Requirements: 8.1, 8.2_

  - [ ] 11.5 Update TestDeleteTask
    - Create MockStorage instance and pre-populate with test task
    - Pass MockStorage to NewTaskManager
    - Call DeleteTask and verify task removed from MockStorage
    - Remove any SetTasks calls
    - _Requirements: 8.1, 8.2_

  - [ ] 11.6 Update TestGetTaskByID
    - Create MockStorage instance and pre-populate with test task
    - Pass MockStorage to NewTaskManager
    - Call GetTaskByID and verify correct task returned
    - Remove any SetTasks calls
    - _Requirements: 8.1, 8.2_

  - [ ] 11.7 Update TestPrintTasks
    - Create MockStorage instance and pre-populate with test tasks
    - Pass MockStorage to NewTaskManager
    - Call PrintTasks and verify output contains task information
    - Remove any SetTasks calls
    - _Requirements: 8.1, 8.3_

  - [ ] 11.8 Update TestGetFormattedTasks
    - Create MockStorage instance and pre-populate with test tasks
    - Pass MockStorage to NewTaskManager
    - Call GetFormattedTasks and verify formatted output
    - Remove any SetTasks calls
    - _Requirements: 8.1, 8.3_

  - [ ] 11.9 Update TestProcessTasks
    - Create MockStorage instance and pre-populate with test tasks
    - Pass MockStorage to NewTaskManager
    - Call ProcessTasks and verify parallel processing output
    - Remove any SetTasks calls
    - _Requirements: 8.1, 8.4_

  - [ ] 11.10 Remove tests for deleted methods
    - Remove any tests for GetTasks() method
    - Remove any tests for SetTasks() method
    - Remove any tests for AddTaskWithID() method
    - _Requirements: 8.5_

- [ ] 12. Verify and test the refactoring
  - [ ] 12.1 Run all TaskManager unit tests
    - Execute `go test ./task/...` and verify all tests pass
    - Fix any failing tests
    - _Requirements: 8.1, 8.2, 8.3, 8.4_

  - [ ] 12.2 Run all CLI tests
    - Execute `go test ./cmd/cli/...` and verify all tests pass
    - Fix any failing tests
    - _Requirements: 7.9_

  - [ ] 12.3 Build and manually test CLI
    - Build CLI with `go build -o cli ./cmd/cli`
    - Test add command and verify task is created
    - Test list command and verify tasks are displayed
    - Test update, status, clear, delete commands
    - Test exit command (verify no save message)
    - _Requirements: 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 7.8_

  - [ ] 12.4 Verify server still works
    - Build server with `go build -o server ./cmd/server`
    - Start server and test API endpoints
    - Verify server handlers work correctly (they don't use TaskManager)
    - _Requirements: 9.1, 9.2, 9.3_
