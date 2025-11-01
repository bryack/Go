package main

import (
	"log"
	"myproject/storage"
	"myproject/task"
	"os"
	"strings"
)

// Command represents a valid user command in the task manager CLI.
// Commands are case-insensitive and validated against a predefined set.
type Command string

const (
	maxInputSize           = 10
	CommandAdd     Command = "add"     // Add a new task
	CommandStatus  Command = "status"  // Change task status
	CommandList    Command = "list"    // Show all tasks
	CommandProcess Command = "process" // Process all tasks in parallel
	CommandClear   Command = "clear"   // Clear task description
	CommandHelp    Command = "help"    // Show available commands
	CommandExit    Command = "exit"    // Save and exit program
	CommandUpdate  Command = "update"  // Update task description
	CommandDelete  Command = "delete"
)

var (
	validCommands = []Command{CommandAdd, CommandStatus, CommandList, CommandProcess, CommandClear, CommandHelp, CommandExit, CommandUpdate, CommandDelete}
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
	dbPath := storage.GetDatabasePath()
	s, err := storage.NewDatabaseStorage(dbPath)
	if err != nil {
		log.Fatal("Failed to initialize database storage:", err)
	}

	cli := NewCLI(
		NewConsoleInputReader(os.Stdin),
		os.Stdout,
		task.NewTaskManager(s, os.Stdout),
		s,
	)

	cli.RunLoop()
}
