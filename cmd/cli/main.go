package main

import (
	"fmt"
	"log"
	"myproject/cmd/cli/auth"
	"myproject/cmd/cli/client"
	"os"
	"strings"
)

// Command represents a valid user command in the task manager CLI.
// Commands are case-insensitive and validated against a predefined set.
type Command string

const (
	maxInputSize            = 10
	CommandAdd      Command = "add"      // Add a new task
	CommandStatus   Command = "status"   // Change task status
	CommandList     Command = "list"     // Show all tasks
	CommandProcess  Command = "process"  // Process all tasks in parallel
	CommandClear    Command = "clear"    // Clear task description
	CommandHelp     Command = "help"     // Show available commands
	CommandExit     Command = "exit"     // Save and exit program
	CommandUpdate   Command = "update"   // Update task description
	CommandDelete   Command = "delete"   // Delete task
	CommandLogin    Command = "login"    // Login with existing account
	CommandRegister Command = "register" // Register new account
	CommandLogout   Command = "logout"   // Logout and clear token
)

var (
	validCommands = []Command{CommandAdd, CommandStatus, CommandList, CommandProcess, CommandClear, CommandHelp, CommandExit, CommandUpdate, CommandDelete, CommandLogin, CommandRegister, CommandLogout}
)

// isValid checks if the command is in the list of supported commands.
// Returns true if the command is valid, false otherwise.
func (cmd Command) isValid() bool {
	for _, valid := range validCommands {
		if cmd == valid {
			return true
		}
	}
	return false
}

// validateCommand converts user input to a valid Command.
// Input is normalized to lowercase before validation.
// Returns the valid command or an error if the command is not recognized.
func validateCommand(input string) (Command, error) {
	inputToLower := strings.ToLower(input)
	cmd := Command(inputToLower)
	if cmd.isValid() {
		return cmd, nil
	}
	return "", ErrInvalidCommand
}

// suggestCommand attempts to find a command that matches the input prefix.
// Returns the first matching command, or empty string if no match is found.
// Used to provide helpful suggestions for typos or partial commands.
func suggestCommand(input string) Command {
	for _, cmd := range validCommands {
		if strings.HasPrefix(string(cmd), input) {
			return cmd
		}
	}
	return ""
}

func main() {
	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Display startup banner and server URL
	fmt.Println("🚀 Task Manager CLI (Client Mode)")
	fmt.Printf("📡 Server: %s\n", cfg.ServerURL)

	// Create HTTP client with configured server URL
	httpClient := client.NewHTTPClient(cfg.ServerURL)

	// Create input reader
	inputReader := NewConsoleInputReader(os.Stdin)

	// Create auth manager
	authManager := auth.NewFileAuthManager(httpClient, inputReader, os.Stdout)

	// Perform initial authentication
	// This will show authentication prompt if no token exists
	// and provide options: 1) Login 2) Register 3) Exit
	token, err := authManager.RequireAuth()
	if err != nil {
		// User chose to exit or authentication failed
		fmt.Fprintf(os.Stdout, "❌ Authentication failed: %v\n", err)
		os.Exit(1)
	}

	// Set token in HTTP client
	httpClient.SetToken(token)

	// Create and run CLI with client and auth manager
	// Proceed to command loop after successful authentication
	cli := NewCLI(
		inputReader,
		os.Stdout,
		cfg,
		httpClient,
		authManager,
	)

	cli.RunLoop()
}
