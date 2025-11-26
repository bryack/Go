# Design Document

## Overview

This design transforms TaskManager from a stateful cache into a stateless service layer that acts as a facade over the Storage interface. The refactoring eliminates dual state management, establishes the database as the single source of truth, and maintains separation between business logic (formatting, processing) and persistence concerns.

The key architectural shift is: **TaskManager becomes a coordinator that delegates persistence to Storage while providing domain-specific operations and formatting.**

## Architecture

### Current Architecture (Before Refactoring)

```
┌─────────────────────────────────────────────────┐
│                    CLI                          │
│  ┌──────────────────────────────────────────┐  │
│  │  handleAddCommand()                      │  │
│  │    1. tm.AddTask(desc)                   │  │
│  │    2. storage.CreateTask(task)           │  │
│  │    3. tm.AddTaskWithID(task)             │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
         │                           │
         ▼                           ▼
┌──────────────────┐        ┌──────────────────┐
│  TaskManager     │        │  Storage         │
│  ┌────────────┐  │        │  (Database)      │
│  │ tasks []   │  │        │                  │
│  │ mutex      │  │        │  tasks table     │
│  └────────────┘  │        └──────────────────┘
│                  │
│  State stored    │
│  in memory       │
└──────────────────┘

Problem: Two sources of truth, sync issues, SaveTasks() overwrites DB
```

### New Architecture (After Refactoring)

```
┌─────────────────────────────────────────────────┐
│                    CLI                          │
│  ┌──────────────────────────────────────────┐  │
│  │  handleAddCommand()                      │  │
│  │    1. tm.AddTask(desc)                   │  │
│  │       └─> storage.CreateTask()           │  │
│  └──────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────┐
│  TaskManager (Stateless Service Layer)          │
│  ┌────────────────────────────────────────────┐ │
│  │  Business Logic:                           │ │
│  │  - FormatTask()                            │ │
│  │  - PrintTasks()                            │ │
│  │  - ProcessTasks()                          │ │
│  │                                            │ │
│  │  CRUD Delegation:                          │ │
│  │  - AddTask() → storage.CreateTask()        │ │
│  │  - UpdateTaskStatus() → storage.Update()   │ │
│  │  - DeleteTask() → storage.DeleteTask()     │ │
│  └────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────┐
│  Storage Interface                               │
│  ┌────────────────────────────────────────────┐ │
│  │  LoadTasks()                               │ │
│  │  GetTaskByID(id)                           │ │
│  │  CreateTask(task)                          │ │
│  │  UpdateTask(task)                          │ │
│  │  DeleteTask(id)                            │ │
│  └────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────┘
         │
         ▼
┌──────────────────────────────────────────────────┐
│  DatabaseStorage (SQLite)                        │
│  - Single source of truth                        │
│  - Thread-safe via database transactions         │
└──────────────────────────────────────────────────┘

Benefits: Single source of truth, no sync, consistent architecture
```

## Components and Interfaces

### 1. Refactored TaskManager Struct

**Before:**
```go
type TaskManager struct {
    tasks  []Task
    mu     sync.Mutex
    writer io.Writer
}
```

**After:**
```go
type TaskManager struct {
    storage storage.Storage  // Dependency injection
    writer  io.Writer        // For output formatting
}
```

**Changes:**
- Remove `tasks []Task` - no more state storage
- Remove `mu sync.Mutex` - no more synchronization needed
- Add `storage storage.Storage` - delegate persistence

### 2. Constructor Signature

**Before:**
```go
func NewTaskManager(writer io.Writer) *TaskManager
```

**After:**
```go
func NewTaskManager(storage storage.Storage, writer io.Writer) *TaskManager
```

**Rationale:** TaskManager needs Storage to delegate operations. Writer remains for formatting output.

### 3. Method Categories and Refactoring Strategy

#### Category A: Formatting Methods (Keep, Modify to Fetch from Storage)

These methods provide business logic and should remain, but fetch data from Storage:

**FormatTask(task Task) string**
- Status: Keep as-is (pure function, no state)
- No changes needed

**PrintTasks() error**
- Current: Locks mutex, formats internal tasks
- New: Fetch from storage, format, print
```go
func (tm *TaskManager) PrintTasks() error {
    tasks, err := tm.storage.LoadTasks()
    if err != nil {
        return err
    }
    return printToWriter(tasks, tm.writer)
}
```

**GetFormattedTasks() string**
- Current: Calls GetTasks() (internal state)
- New: Fetch from storage, format
```go
func (tm *TaskManager) GetFormattedTasks() string {
    tasks, err := tm.storage.LoadTasks()
    if err != nil {
        return fmt.Sprintf("Error loading tasks: %v", err)
    }
    return formatTasks(tasks)
}
```

#### Category B: Processing Methods (Keep, Modify to Fetch from Storage)

**ProcessTasks()**
- Current: Gets copy of internal tasks, processes in parallel
- New: Fetch from storage, process in parallel
```go
func (tm *TaskManager) ProcessTasks() {
    tasks, err := tm.storage.LoadTasks()
    if err != nil {
        fmt.Fprintf(tm.writer, "Error loading tasks: %v\n", err)
        return
    }
    
    if len(tasks) == 0 {
        fmt.Fprintln(tm.writer, "No tasks to process")
        return
    }
    
    fmt.Fprintln(tm.writer, "Starting parallel task processing...")
    var wg sync.WaitGroup
    for _, task := range tasks {
        wg.Add(1)
        go processTask(task, &wg)
    }
    wg.Wait()
    fmt.Fprintln(tm.writer, "All tasks processed successfully!")
}
```

#### Category C: CRUD Methods (Refactor to Delegate to Storage)

**AddTask(input string) Task** → **AddTask(input string) (int, error)**
- Current: Creates task object without ID
- New: Create task in storage, return ID
```go
func (tm *TaskManager) AddTask(input string) (int, error) {
    newTask := Task{
        Description: input,
        Done:        false,
    }
    id, err := tm.storage.CreateTask(newTask)
    if err != nil {
        return 0, fmt.Errorf("failed to create task: %w", err)
    }
    return id, nil
}
```

**UpdateTaskStatus(id int, done bool) error**
- Current: Finds in internal slice, updates
- New: Fetch from storage, update, save
```go
func (tm *TaskManager) UpdateTaskStatus(id int, done bool) error {
    task, err := tm.storage.GetTaskByID(id)
    if err != nil {
        return err
    }
    
    task.Done = done
    
    if err := tm.storage.UpdateTask(task); err != nil {
        return err
    }
    
    return nil
}
```

**UpdateTaskDescription(id int, description string) error**
- Current: Finds in internal slice, updates
- New: Fetch from storage, update, save
```go
func (tm *TaskManager) UpdateTaskDescription(id int, description string) error {
    task, err := tm.storage.GetTaskByID(id)
    if err != nil {
        return err
    }
    
    task.Description = description
    
    if err := tm.storage.UpdateTask(task); err != nil {
        return err
    }
    
    return nil
}
```

**ClearDescription(id int) error**
- Current: Finds in internal slice, clears
- New: Fetch from storage, clear, save
```go
func (tm *TaskManager) ClearDescription(id int) error {
    task, err := tm.storage.GetTaskByID(id)
    if err != nil {
        return err
    }
    
    task.Description = ""
    
    if err := tm.storage.UpdateTask(task); err != nil {
        return err
    }
    
    return nil
}
```

**DeleteTask(id int) error**
- Current: Finds in internal slice, removes
- New: Delegate to storage
```go
func (tm *TaskManager) DeleteTask(id int) error {
    return tm.storage.DeleteTask(id)
}
```

**GetTaskByID(id int) (Task, error)**
- Current: Searches internal slice
- New: Delegate to storage
```go
func (tm *TaskManager) GetTaskByID(id int) (Task, error) {
    return tm.storage.GetTaskByID(id)
}
```

#### Category D: State Management Methods (Remove)

These methods exist only to manage internal state and should be removed:

- **GetTasks() []Task** - Remove (no internal state to get)
- **SetTasks(newTask []Task)** - Remove (no internal state to set)
- **AddTaskWithID(t Task)** - Remove (AddTask now handles everything)

### 4. CLI Integration Changes

The CLI struct and methods need updates to work with the refactored TaskManager:

**CLI Constructor - No Changes Needed**
```go
func NewCLI(input InputReader, output io.Writer, taskManager *task.TaskManager, storage *storage.DatabaseStorage) *CLI
```
- CLI still receives both TaskManager and Storage
- TaskManager internally uses Storage
- CLI can use Storage directly for operations like LoadTasks if needed

**handleAddCommand() - Simplified**

Before:
```go
func (cli *CLI) handleAddCommand() error {
    // ... validation ...
    newTask := cli.taskManager.AddTask(desc)
    id, err := cli.storage.CreateTask(newTask)
    if err != nil {
        return fmt.Errorf("adding task: creation failed: %w", err)
    }
    newTask.ID = id
    cli.taskManager.AddTaskWithID(newTask)
    fmt.Fprintf(cli.output, "✅ Task added (ID: %d)\n", id)
    return nil
}
```

After:
```go
func (cli *CLI) handleAddCommand() error {
    fmt.Fprintln(cli.output, "Enter task description:")
    
    desc, err := cli.input.ReadInput(maxDescriptionInputSize)
    if err != nil {
        return fmt.Errorf("adding task: input failed: %w", err)
    }
    
    desc, err = validation.ValidateTaskDescription(desc)
    if err != nil {
        return fmt.Errorf("adding task: validation failed: %w", err)
    }
    
    id, err := cli.taskManager.AddTask(desc)
    if err != nil {
        return fmt.Errorf("adding task: creation failed: %w", err)
    }
    
    fmt.Fprintf(cli.output, "✅ Task added (ID: %d)\n", id)
    return nil
}
```

**handleStatusCommand() - No Changes**
- Already uses `cli.taskManager.UpdateTaskStatus()`
- Will work with refactored method

**handleUpdateCommand() - No Changes**
- Already uses `cli.taskManager.UpdateTaskDescription()`
- Will work with refactored method

**handleClearCommand() - No Changes**
- Already uses `cli.taskManager.ClearDescription()`
- Will work with refactored method

**handleDeleteCommand() - No Changes**
- Already uses `cli.taskManager.DeleteTask()`
- Will work with refactored method

**CommandList Handler - Simplified**

Before:
```go
case CommandList:
    tasks, err := cli.storage.LoadTasks()
    if err != nil {
        cli.handleError(err, "Failed to load tasks")
        break
    }
    cli.taskManager.SetTasks(tasks)
    if err := cli.taskManager.PrintTasks(); err != nil {
        cli.handleError(err, "Print tasks error")
    }
```

After:
```go
case CommandList:
    if err := cli.taskManager.PrintTasks(); err != nil {
        cli.handleError(err, "Print tasks error")
    }
```

**CommandExit Handler - Remove SaveTasks**

Before:
```go
case CommandExit:
    if err := cli.storage.SaveTasks(cli.taskManager.GetTasks()); err != nil {
        cli.handleError(err, "Save error")
    } else {
        fmt.Fprintln(cli.output, "Tasks saved successfully!")
    }
    fmt.Fprintln(cli.output, "👋 Bye!")
    return
```

After:
```go
case CommandExit:
    fmt.Fprintln(cli.output, "👋 Bye!")
    return
```

**promptForTaskWithDisplay() - Use Storage Directly**

Before:
```go
func (cli *CLI) promptForTaskWithDisplay(prompt string) (id int, t task.Task, err error) {
    id, err = cli.promptForTaskID(prompt)
    if err != nil {
        return 0, t, err
    }
    
    t, err = cli.taskManager.GetTaskByID(id)
    if err != nil {
        return 0, t, err
    }
    
    fmt.Fprintf(cli.output, "Current task: '%s'\n", task.FormatTask(t))
    return id, t, nil
}
```

After (Option 1 - Use TaskManager):
```go
func (cli *CLI) promptForTaskWithDisplay(prompt string) (id int, t task.Task, err error) {
    id, err = cli.promptForTaskID(prompt)
    if err != nil {
        return 0, t, err
    }
    
    t, err = cli.taskManager.GetTaskByID(id)  // TaskManager delegates to Storage
    if err != nil {
        return 0, t, err
    }
    
    fmt.Fprintf(cli.output, "Current task: '%s'\n", task.FormatTask(t))
    return id, t, nil
}
```

After (Option 2 - Use Storage Directly):
```go
func (cli *CLI) promptForTaskWithDisplay(prompt string) (id int, t task.Task, err error) {
    id, err = cli.promptForTaskID(prompt)
    if err != nil {
        return 0, t, err
    }
    
    t, err = cli.storage.GetTaskByID(id)  // Direct storage access
    if err != nil {
        return 0, t, err
    }
    
    fmt.Fprintf(cli.output, "Current task: '%s'\n", task.FormatTask(t))
    return id, t, nil
}
```

**Recommendation:** Use Option 1 (TaskManager) for consistency, even though it's just a pass-through.

### 5. Main Function Changes

**cmd/cli/main.go**

Before:
```go
func main() {
    dbPath := storage.GetDatabasePath()
    s, err := storage.NewDatabaseStorage(dbPath)
    if err != nil {
        log.Fatal("Failed to initialize database storage:", err)
    }
    
    cli := NewCLI(
        NewConsoleInputReader(os.Stdin),
        os.Stdout,
        task.NewTaskManager(os.Stdout),
        s,
    )
    
    cli.RunLoop()
}
```

After:
```go
func main() {
    dbPath := storage.GetDatabasePath()
    s, err := storage.NewDatabaseStorage(dbPath)
    if err != nil {
        log.Fatal("Failed to initialize database storage:", err)
    }
    
    cli := NewCLI(
        NewConsoleInputReader(os.Stdin),
        os.Stdout,
        task.NewTaskManager(s, os.Stdout),  // Pass storage to TaskManager
        s,
    )
    
    cli.RunLoop()
}
```

## Data Models

No changes to the Task struct:

```go
type Task struct {
    ID          int    `json:"id"`
    Description string `json:"description"`
    Done        bool   `json:"done"`
}
```

The Storage interface remains unchanged and already provides all needed operations:

```go
type Storage interface {
    LoadTasks() ([]task.Task, error)
    GetTaskByID(id int) (task task.Task, err error)
    CreateTask(task task.Task) (int, error)
    UpdateTask(task task.Task) error
    DeleteTask(id int) error
    SaveTasks(tasks []task.Task) error  // Still used by server, not CLI
}
```

## Error Handling

### Error Types

Keep existing error types in TaskManager:
```go
var (
    ErrTaskNotFound = errors.New("task not found")
    ErrPrintTask    = errors.New("failed to print tasks")
)
```

### Error Propagation Strategy

1. **Storage errors bubble up**: TaskManager methods return storage errors directly
2. **Wrap with context**: Add context when helpful for debugging
3. **Handle in CLI**: CLI layer displays user-friendly messages

Example error flow:
```
Storage.GetTaskByID(999) → ErrTaskNotFound
    ↓
TaskManager.UpdateTaskStatus(999, true) → ErrTaskNotFound
    ↓
CLI.handleStatusCommand() → "Status command error: task not found"
```

### Error Handling in New Methods

```go
func (tm *TaskManager) AddTask(input string) (int, error) {
    newTask := Task{Description: input, Done: false}
    id, err := tm.storage.CreateTask(newTask)
    if err != nil {
        return 0, fmt.Errorf("failed to create task: %w", err)
    }
    return id, nil
}
```

## Testing Strategy

### Unit Testing Approach

**Mock Storage Interface**

Create a mock implementation for testing:

```go
type MockStorage struct {
    tasks map[int]task.Task
    nextID int
}

func NewMockStorage() *MockStorage {
    return &MockStorage{
        tasks: make(map[int]task.Task),
        nextID: 1,
    }
}

func (m *MockStorage) CreateTask(t task.Task) (int, error) {
    id := m.nextID
    m.nextID++
    t.ID = id
    m.tasks[id] = t
    return id, nil
}

func (m *MockStorage) GetTaskByID(id int) (task.Task, error) {
    t, exists := m.tasks[id]
    if !exists {
        return task.Task{}, storage.ErrTaskNotFound
    }
    return t, nil
}

func (m *MockStorage) UpdateTask(t task.Task) error {
    if _, exists := m.tasks[t.ID]; !exists {
        return storage.ErrTaskNotFound
    }
    m.tasks[t.ID] = t
    return nil
}

func (m *MockStorage) DeleteTask(id int) error {
    if _, exists := m.tasks[id]; !exists {
        return storage.ErrTaskNotFound
    }
    delete(m.tasks, id)
    return nil
}

func (m *MockStorage) LoadTasks() ([]task.Task, error) {
    tasks := make([]task.Task, 0, len(m.tasks))
    for _, t := range m.tasks {
        tasks = append(tasks, t)
    }
    return tasks, nil
}

func (m *MockStorage) SaveTasks(tasks []task.Task) error {
    return nil // Not used in refactored code
}
```

### Test Cases to Update

**Test: AddTask**
```go
func TestAddTask(t *testing.T) {
    mockStorage := NewMockStorage()
    tm := task.NewTaskManager(mockStorage, &strings.Builder{})
    
    id, err := tm.AddTask("Test task")
    
    assert.NoError(t, err)
    assert.Equal(t, 1, id)
    
    // Verify task was created in storage
    savedTask, err := mockStorage.GetTaskByID(id)
    assert.NoError(t, err)
    assert.Equal(t, "Test task", savedTask.Description)
    assert.False(t, savedTask.Done)
}
```

**Test: UpdateTaskStatus**
```go
func TestUpdateTaskStatus(t *testing.T) {
    mockStorage := NewMockStorage()
    tm := task.NewTaskManager(mockStorage, &strings.Builder{})
    
    // Setup: Create a task
    id, _ := mockStorage.CreateTask(task.Task{Description: "Test", Done: false})
    
    // Act: Update status
    err := tm.UpdateTaskStatus(id, true)
    
    // Assert
    assert.NoError(t, err)
    updatedTask, _ := mockStorage.GetTaskByID(id)
    assert.True(t, updatedTask.Done)
}
```

**Test: PrintTasks**
```go
func TestPrintTasks(t *testing.T) {
    mockStorage := NewMockStorage()
    var output strings.Builder
    tm := task.NewTaskManager(mockStorage, &output)
    
    // Setup: Add tasks to storage
    mockStorage.CreateTask(task.Task{Description: "Task 1", Done: false})
    mockStorage.CreateTask(task.Task{Description: "Task 2", Done: true})
    
    // Act
    err := tm.PrintTasks()
    
    // Assert
    assert.NoError(t, err)
    assert.Contains(t, output.String(), "Task 1")
    assert.Contains(t, output.String(), "Task 2")
}
```

**Test: ProcessTasks**
```go
func TestProcessTasks(t *testing.T) {
    mockStorage := NewMockStorage()
    var output strings.Builder
    tm := task.NewTaskManager(mockStorage, &output)
    
    // Setup: Add tasks
    mockStorage.CreateTask(task.Task{Description: "Task 1", Done: false})
    mockStorage.CreateTask(task.Task{Description: "Task 2", Done: false})
    
    // Act
    tm.ProcessTasks()
    
    // Assert: Check output contains processing messages
    assert.Contains(t, output.String(), "Starting parallel task processing")
    assert.Contains(t, output.String(), "All tasks processed successfully")
}
```

### Tests to Remove

- Any test using `GetTasks()` or `SetTasks()` directly
- Tests that verify internal state management
- Concurrency tests for TaskManager mutex (Storage handles this now)

### Integration Testing

Test the full stack: CLI → TaskManager → DatabaseStorage

```go
func TestCLIIntegration(t *testing.T) {
    // Setup: Create temp database
    tmpDB := createTempDatabase(t)
    defer os.Remove(tmpDB)
    
    storage, _ := storage.NewDatabaseStorage(tmpDB)
    tm := task.NewTaskManager(storage, os.Stdout)
    
    // Test: Add task via TaskManager
    id, err := tm.AddTask("Integration test task")
    assert.NoError(t, err)
    
    // Verify: Task exists in database
    task, err := storage.GetTaskByID(id)
    assert.NoError(t, err)
    assert.Equal(t, "Integration test task", task.Description)
}
```

## Thread Safety

### Current Approach (Mutex in TaskManager)
- TaskManager uses `sync.Mutex` to protect internal slice
- Each method locks/unlocks mutex

### New Approach (Delegate to Storage)
- TaskManager has no state, no mutex needed
- DatabaseStorage handles concurrency via SQLite's built-in locking
- SQLite provides serialized access to database file
- Multiple goroutines can safely call TaskManager methods

### Concurrency Considerations

**ProcessTasks() Parallel Execution:**
```go
func (tm *TaskManager) ProcessTasks() {
    tasks, err := tm.storage.LoadTasks()  // Single snapshot
    if err != nil {
        return
    }
    
    var wg sync.WaitGroup
    for _, task := range tasks {
        wg.Add(1)
        go processTask(task, &wg)  // Process snapshot in parallel
    }
    wg.Wait()
}
```

- Fetches snapshot of tasks from storage
- Processes snapshot in parallel (read-only operations)
- No concurrent writes to shared state
- Safe without additional synchronization

## Migration Path

### Phase 1: Refactor TaskManager
1. Update TaskManager struct
2. Update constructor
3. Refactor all methods
4. Remove obsolete methods

### Phase 2: Update CLI
1. Update main.go to pass Storage to TaskManager
2. Simplify handleAddCommand
3. Simplify CommandList handler
4. Remove SaveTasks from CommandExit
5. Update any other affected handlers

### Phase 3: Update Tests
1. Create MockStorage
2. Update existing tests to use MockStorage
3. Remove tests for deleted methods
4. Add integration tests

### Phase 4: Verification
1. Run all tests
2. Manual testing of CLI commands
3. Verify server still works (doesn't use TaskManager)
4. Check for any remaining references to deleted methods

## Design Decisions and Rationales

### Decision 1: Keep TaskManager vs Remove It Entirely

**Chosen:** Keep TaskManager as a service layer

**Rationale:**
- Provides separation of concerns (business logic vs persistence)
- Formatting logic belongs in domain layer, not storage layer
- ProcessTasks is unique business logic worth preserving
- Easier to test business logic with mocked storage
- Maintains consistent API for CLI

**Alternative Considered:** Remove TaskManager, have CLI call Storage directly
- Would work, but mixes concerns
- Formatting logic would need to move somewhere
- Less testable

### Decision 2: AddTask Return Signature

**Chosen:** `AddTask(input string) (int, error)`

**Rationale:**
- Returns ID immediately after creation
- Matches the actual operation (create and get ID)
- Simplifies CLI code (no need for separate CreateTask call)
- Error handling is explicit

**Alternative Considered:** `AddTask(input string) (Task, error)`
- Would return full task object
- Requires extra database query to fetch created task
- Less efficient

### Decision 3: Error Handling Strategy

**Chosen:** Propagate storage errors directly, wrap with context when helpful

**Rationale:**
- Storage errors (like ErrTaskNotFound) are already meaningful
- Wrapping adds context without hiding original error
- CLI can handle errors appropriately

**Alternative Considered:** Define new error types in TaskManager
- Would add unnecessary abstraction
- Storage errors are already well-defined

### Decision 4: Keep GetTaskByID in TaskManager

**Chosen:** Keep as pass-through to Storage

**Rationale:**
- Maintains consistent API (all task operations through TaskManager)
- CLI doesn't need to know about Storage implementation
- Future: Could add caching or business logic here

**Alternative Considered:** Have CLI call Storage.GetTaskByID directly
- Would work, but breaks consistency
- CLI would need to know about both TaskManager and Storage APIs

### Decision 5: Remove SaveTasks from CLI Exit

**Chosen:** Remove entirely, rely on per-operation persistence

**Rationale:**
- Every operation already persists to database immediately
- SaveTasks() bulk overwrite is dangerous (could lose concurrent changes)
- Database is always up-to-date
- Simpler exit logic

**Alternative Considered:** Keep SaveTasks as safety net
- Unnecessary with immediate persistence
- Could cause data loss if database was updated elsewhere

## Summary

This refactoring transforms TaskManager from a stateful cache into a stateless service layer that:

1. **Eliminates dual state management** - Database is single source of truth
2. **Maintains separation of concerns** - Business logic separate from persistence
3. **Simplifies CLI code** - No more manual synchronization
4. **Improves consistency** - Same architecture as server handlers
5. **Preserves unique features** - Formatting and parallel processing remain
6. **Enhances testability** - Easy to mock Storage interface
7. **Ensures thread safety** - Delegates to Storage's concurrency handling

The refactoring is backward-compatible in behavior (CLI commands work the same) while significantly improving the internal architecture.
