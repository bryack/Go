package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"myproject/cmd/cli/auth"
	"myproject/cmd/cli/client"
	"myproject/validation"
	"strings"
)

const (
	maxCommandInputSize     = 10
	maxTaskIDInputSize      = 10
	maxDescriptionInputSize = 200
	maxStatusInputSize      = 10
)

var (
	ErrMaxSizeExceeded      = errors.New("input too long")
	ErrEmptyInput           = errors.New("empty input")
	ErrInvalidTaskId        = errors.New("invalid ID format")
	ErrInvalidCommand       = errors.New("invalid command")
	ErrInvalidStatus        = errors.New("invalid status")
	ErrDescUnchanged        = errors.New("description unchanged")
	ErrInvalidConfirmChoice = errors.New("invalid confirm choice")
)

// InputReader defines an interface for reading user input with size validation.
// Implementations must trim whitespace and enforce maximum input length constraints.
type InputReader interface {
	ReadInput(maxSize int) (string, error)
}

// CLI manages the command-line interface for task operations.
// Coordinates user input, API client communication, and authentication with proper error handling.
type CLI struct {
	input       InputReader
	output      io.Writer
	client      client.TaskClient
	authManager auth.AuthManager
	config      *Config
}

// NewCLI creates a new CLI instance with the provided dependencies.
// Returns a configured CLI ready to process user commands and manage tasks via API.
func NewCLI(input InputReader, output io.Writer, cfg *Config, client client.TaskClient, authManager auth.AuthManager) *CLI {
	return &CLI{
		input:       input,
		output:      output,
		client:      client,
		authManager: authManager,
		config:      cfg,
	}
}

// ConsoleInputReader implements InputReader for reading from console input streams.
// Uses buffered reading to handle user input line-by-line with proper error handling.
type ConsoleInputReader struct {
	reader *bufio.Reader
}

// NewConsoleInputReader creates a new ConsoleInputReader from an io.Reader.
// Returns a reader configured with buffered input for efficient line reading.
func NewConsoleInputReader(reader io.Reader) *ConsoleInputReader {
	return &ConsoleInputReader{
		reader: bufio.NewReader(reader),
	}
}

// ReadInput reads a line from the input stream and validates its length.
// Returns trimmed input or errors for empty input, EOF, or size limit violations.
func (c *ConsoleInputReader) ReadInput(maxSize int) (string, error) {
	input, err := c.reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return "", io.EOF
		}
		return "", err
	}

	input = strings.TrimSpace(input)
	if len(input) > maxSize {
		return "", ErrMaxSizeExceeded
	}

	if len(input) == 0 {
		return "", ErrEmptyInput
	}

	return input, nil
}

// formatTask formats a task for display
func formatTask(t client.Task) string {
	status := "[ ]"
	if t.Done {
		status = "[✓]"
	}
	return fmt.Sprintf("%s %d: %s", status, t.ID, t.Description)
}

// promptForTaskID prompts the user for a task ID and validates the input.
// Returns the validated task ID or an error if input is invalid or exceeds size limits.
func (cli *CLI) promptForTaskID(prompt string) (id int, err error) {
	fmt.Fprint(cli.output, prompt)

	input, err := cli.input.ReadInput(maxTaskIDInputSize)
	if err != nil {
		return 0, err
	}

	return validation.ValidateTaskID(input)
}

// promptForTaskWithDisplay prompts for a task ID and displays the current task details.
// Returns the task ID, task object, and any errors from validation or task retrieval.
func (cli *CLI) promptForTaskWithDisplay(prompt string) (id int, t *client.Task, err error) {
	id, err = cli.promptForTaskID(prompt)
	if err != nil {
		return 0, nil, err
	}

	t, err = cli.client.GetTask(id)
	if err != nil {
		return 0, nil, err
	}

	fmt.Fprintf(cli.output, "Current task: '%s'\n", formatTask(*t))

	return id, t, nil
}

// handleAddCommand prompts for a task description and adds a new task via the API.
// Validates input length and description format before creating the task.
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

	task, err := cli.client.CreateTask(desc)
	if err != nil {
		return fmt.Errorf("adding task: creation failed: %w", err)
	}

	fmt.Fprintf(cli.output, "✅ Task added (ID: %d)\n", task.ID)
	return nil
}

// handleStatusCommand prompts for a task ID and new status, then updates the task via API.
// Accepts 'done' or 'undone' as valid status values with proper validation.
func (cli *CLI) handleStatusCommand() error {
	id, _, err := cli.promptForTaskWithDisplay("Enter task ID to change status:\n")
	if err != nil {
		return fmt.Errorf("updating status: task id validation failed: %w", err)
	}

	fmt.Fprint(cli.output, "Enter new status 'done' // 'undone'\n")
	str, err := cli.input.ReadInput(maxStatusInputSize)
	if err != nil {
		return fmt.Errorf("updating status: read status for task id %d failed: %w", id, err)
	}

	var done bool
	switch str {
	case "done":
		done = true
	case "undone":
		done = false
	default:
		return fmt.Errorf("updating status: invalid status: %q for task id %d: %w (must be 'done' or 'undone')", str, id, ErrInvalidStatus)
	}

	_, err = cli.client.UpdateTask(id, nil, &done)
	if err != nil {
		return fmt.Errorf("updating status for task id %d failed: %w", id, err)
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) status is has changed\n", id)
	return nil
}

// handleClearCommand prompts for a task ID and clears its description via API.
// Validates the task exists before clearing the description field.
func (cli *CLI) handleClearCommand() error {
	id, _, err := cli.promptForTaskWithDisplay("Enter task ID you want to clear description\n")
	if err != nil {
		return fmt.Errorf("clearing task description: task id validation failed: %w", err)
	}

	emptyDesc := ""
	_, err = cli.client.UpdateTask(id, &emptyDesc, nil)
	if err != nil {
		return fmt.Errorf("clearing task description for task id %d failed: %w", id, err)
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) description cleared!\n", id)
	return nil
}

// handleUpdateCommand prompts for a task ID and new description, then updates the task via API.
// Validates that the new description differs from the current one before updating.
func (cli *CLI) handleUpdateCommand() error {
	id, t, err := cli.promptForTaskWithDisplay("Enter task ID to update:\n")
	if err != nil {
		return fmt.Errorf("updating task description: task id validation failed: %w", err)
	}

	fmt.Fprint(cli.output, "Enter new description:\n")
	desc, err := cli.input.ReadInput(maxDescriptionInputSize)
	if err != nil {
		return fmt.Errorf("updating task description for task id %d: read description '%s' failed: %w", id, desc, err)
	}

	desc, err = validation.ValidateTaskDescription(desc)
	if err != nil {
		return fmt.Errorf("updating task description for task id %d: validate description '%s' failed: %w", id, desc, err)
	}

	if desc == t.Description {
		return fmt.Errorf("updating task description for task id %d: %w", id, ErrDescUnchanged)
	}

	_, err = cli.client.UpdateTask(id, &desc, nil)
	if err != nil {
		return fmt.Errorf("updating task description for task id %d failed: %w", id, err)
	}

	fmt.Fprintf(cli.output, "✅ Task (ID: %d) updated\n", id)
	return nil
}

// handleDeleteCommand prompts for a task ID and confirmation, then deletes the task via API.
// Requires explicit 'y' confirmation to proceed with deletion, 'n' cancels the operation.
func (cli *CLI) handleDeleteCommand() error {
	id, _, err := cli.promptForTaskWithDisplay("Enter task ID to delete task:\n")
	if err != nil {
		return fmt.Errorf("deleting task: id validation failed: %w", err)
	}

	fmt.Fprintln(cli.output, "Enter y/N:")
	str, err := cli.input.ReadInput(10)
	if err != nil {
		return fmt.Errorf("deleting task id %d: read confirmation failed: %w", id, err)
	}
	str = strings.ToLower(str)

	switch str {
	case "y":
		if err = cli.client.DeleteTask(id); err != nil {
			return fmt.Errorf("deleting task id %d failed: %w", id, err)
		}
		fmt.Fprintf(cli.output, "✅ Task (ID: %d) deleted\n", id)
		return nil
	case "n":
		fmt.Fprintln(cli.output, "Deletion canceled")
		return nil
	default:
		return fmt.Errorf("deleting task id %d: %q: %w (must be 'y' or 'n')", id, str, ErrInvalidConfirmChoice)
	}
}

// showHelp displays the list of available commands and their descriptions.
// Outputs a formatted help menu to the configured output writer.
func (cli *CLI) showHelp() {
	fmt.Fprintln(cli.output, "\n=== Available Commands ===")
	fmt.Fprintln(cli.output, "add      - Add a new task")
	fmt.Fprintln(cli.output, "status   - Change task status")
	fmt.Fprintln(cli.output, "list     - Show all tasks")
	fmt.Fprintln(cli.output, "process  - Process all tasks in parallel")
	fmt.Fprintln(cli.output, "clear    - Clear task description")
	fmt.Fprintln(cli.output, "update   - Update task description")
	fmt.Fprintln(cli.output, "delete   - Delete task")
	fmt.Fprintln(cli.output, "login    - Login with existing account")
	fmt.Fprintln(cli.output, "register - Register new account")
	fmt.Fprintln(cli.output, "logout   - Logout and clear token")
	fmt.Fprintln(cli.output, "help     - Show this help")
	fmt.Fprintln(cli.output, "exit     - Save and exit")
	fmt.Fprintln(cli.output, "==========================")
}

// handleError formats and displays error messages with context information.
// Provides user-friendly error messages and handles EOF as input interruption.
// Handles NetworkError and APIError with specific formatting for better user experience.
func (cli *CLI) handleError(err error, context string) {
	if errors.Is(err, io.EOF) {
		fmt.Fprintf(cli.output, "%s: input interrupted by user\n", context)
		return
	}

	// Handle NetworkError - connection failures
	var netErr *client.NetworkError
	if errors.As(err, &netErr) {
		fmt.Fprintf(cli.output, "❌ %s: Cannot connect to server at %s\n", context, netErr.URL)
		fmt.Fprintln(cli.output, "   Please check that the server is running and the URL is correct")
		return
	}

	// Handle APIError - server error responses
	var apiErr *client.APIError
	if errors.As(err, &apiErr) {
		fmt.Fprintf(cli.output, "❌ %s: %s\n", context, apiErr.Message)
		return
	}

	// Handle all other errors with generic format
	fmt.Fprintf(cli.output, "%s: %v\n", context, err)
}

// handleAuthError detects authentication errors and triggers re-authentication flow
// Returns true if re-authentication was successful, false otherwise
func (cli *CLI) handleAuthError(err error) bool {
	if !client.IsAuthError(err) {
		return false
	}

	// Trigger re-authentication
	token, authErr := cli.authManager.HandleAuthError()
	if authErr != nil {
		fmt.Fprintf(cli.output, "❌ Re-authentication failed: %v\n", authErr)
		return false
	}

	// Update client with new token
	cli.client.SetToken(token)
	fmt.Fprintln(cli.output, "✅ Re-authentication successful! Please try your command again.")
	return true
}

// handleListCommand retrieves and displays all tasks from the API
func (cli *CLI) handleListCommand() error {
	tasks, err := cli.client.GetTasks()
	if err != nil {
		return fmt.Errorf("failed to retrieve tasks: %w", err)
	}

	if len(tasks) == 0 {
		fmt.Fprintln(cli.output, "No tasks found")
		return nil
	}

	fmt.Fprintln(cli.output, "\n=== Your Tasks ===")
	for _, task := range tasks {
		fmt.Fprintln(cli.output, formatTask(task))
	}
	fmt.Fprintln(cli.output, "==================")

	return nil
}

// handleLoginCommand prompts for credentials and authenticates the user
func (cli *CLI) handleLoginCommand() error {
	token, err := cli.authManager.PromptLogin()
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Update client with new token
	cli.client.SetToken(token)

	return nil
}

// handleRegisterCommand prompts for credentials and registers a new user
func (cli *CLI) handleRegisterCommand() error {
	token, err := cli.authManager.PromptRegister()
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	// Update client with new token
	cli.client.SetToken(token)

	return nil
}

// handleLogoutCommand clears the stored authentication token
func (cli *CLI) handleLogoutCommand() error {
	err := cli.authManager.ClearToken()
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}

	fmt.Fprintln(cli.output, "✅ Logged out successfully")
	fmt.Fprintln(cli.output, "👋 Bye!")
	return nil
}

// RunLoop starts the main command processing loop for the CLI application.
// Continuously reads commands, executes handlers, and manages application lifecycle until exit.
func (cli *CLI) RunLoop() {
	cli.showHelp()
	for {
		fmt.Fprint(cli.output, "\nEnter command: ")
		input, err := cli.input.ReadInput(maxCommandInputSize)
		if err != nil {
			cli.handleError(err, "Input error")
			continue
		}

		cmd, err := validateCommand(input)
		if err != nil {
			suggestion := suggestCommand(input)
			if suggestion != "" {
				fmt.Fprintf(cli.output, "❌ Unknown command: '%s', maybe you wanted: '%s'\n", input, suggestion)
			} else {
				cli.handleError(err, "Command validate error")
				fmt.Fprintln(cli.output, "Type 'help' to see available commands")
			}
			continue
		}

		switch Command(cmd) {
		case CommandAdd:
			if err := cli.handleAddCommand(); err != nil {
				if cli.handleAuthError(err) {
					continue
				}
				cli.handleError(err, "Add command error")
			}

		case CommandStatus:
			if err := cli.handleStatusCommand(); err != nil {
				if cli.handleAuthError(err) {
					continue
				}
				cli.handleError(err, "Status command error")
			}

		case CommandList:
			if err := cli.handleListCommand(); err != nil {
				if cli.handleAuthError(err) {
					continue
				}
				cli.handleError(err, "List command error")
			}

		case CommandProcess:
			fmt.Fprintln(cli.output, "⚠️  Process command not available in client mode")

		case CommandClear:
			if err := cli.handleClearCommand(); err != nil {
				if cli.handleAuthError(err) {
					continue
				}
				cli.handleError(err, "Clear command error")
			}

		case CommandDelete:
			if err := cli.handleDeleteCommand(); err != nil {
				if cli.handleAuthError(err) {
					continue
				}
				cli.handleError(err, "Delete command error")
			}

		case CommandHelp:
			cli.showHelp()

		case CommandExit:
			fmt.Fprintln(cli.output, "👋 Bye!")
			return

		case CommandUpdate:
			if err := cli.handleUpdateCommand(); err != nil {
				if cli.handleAuthError(err) {
					continue
				}
				cli.handleError(err, "Update command error")
			}

		case CommandLogin:
			if err := cli.handleLoginCommand(); err != nil {
				cli.handleError(err, "Login command error")
			}

		case CommandRegister:
			if err := cli.handleRegisterCommand(); err != nil {
				cli.handleError(err, "Register command error")
			}

		case CommandLogout:
			if err := cli.handleLogoutCommand(); err != nil {
				cli.handleError(err, "Logout command error")
			}
			return
		}
	}
}
