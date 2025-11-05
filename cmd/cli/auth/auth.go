package auth

import (
	"bufio"
	"fmt"
	"io"
	"myproject/cmd/cli/client"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// AuthManager defines the interface for managing authentication state and token persistence
type AuthManager interface {
	// Token management
	LoadToken() (string, error)
	SaveToken(token string) error
	ClearToken() error

	// Authentication state
	IsAuthenticated() bool
	RequireAuth() (string, error)

	// Interactive authentication
	PromptLogin() (string, error)
	PromptRegister() (string, error)

	// Re-authentication handling
	HandleAuthError() (string, error)
}

// InputReader defines an interface for reading user input
type InputReader interface {
	ReadInput(maxSize int) (string, error)
}

// FileAuthManager implements AuthManager using file-based token storage
type FileAuthManager struct {
	tokenPath string
	client    client.TaskClient
	input     InputReader
	output    io.Writer
}

// NewFileAuthManager creates a new FileAuthManager with token storage in ~/.task-cli/token
func NewFileAuthManager(client client.TaskClient, input InputReader, output io.Writer) *FileAuthManager {
	homeDir, _ := os.UserHomeDir()
	tokenPath := filepath.Join(homeDir, ".task-cli", "token")

	return &FileAuthManager{
		tokenPath: tokenPath,
		client:    client,
		input:     input,
		output:    output,
	}
}

// SaveToken writes the token to file with 0600 permissions
// Creates parent directories with 0700 permissions if they don't exist
func (m *FileAuthManager) SaveToken(token string) error {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(m.tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token directory: %w", err)
	}

	// Write token to file with restricted permissions
	if err := os.WriteFile(m.tokenPath, []byte(token), 0600); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

// LoadToken reads the token from file
// Verifies file permissions on load and warns if too permissive
func (m *FileAuthManager) LoadToken() (string, error) {
	// Check if file exists
	info, err := os.Stat(m.tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no token found")
		}
		return "", fmt.Errorf("failed to stat token file: %w", err)
	}

	// Verify file permissions (warn if too permissive)
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		fmt.Fprintf(m.output, "token file has insecure permissions (%o), run: chmod 600 %s", mode, m.tokenPath)
	}

	// Read token from file
	data, err := os.ReadFile(m.tokenPath)
	if err != nil {
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("token file is empty")
	}

	return token, nil
}

// ClearToken deletes the token file
func (m *FileAuthManager) ClearToken() error {
	if err := os.Remove(m.tokenPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already deleted
		}
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

// IsAuthenticated checks if a valid token is stored
func (m *FileAuthManager) IsAuthenticated() bool {
	token, err := m.LoadToken()
	return err == nil && token != ""
}

// RequireAuth loads token or prompts for authentication
// Returns a valid token or error
func (m *FileAuthManager) RequireAuth() (string, error) {
	// Try to load existing token
	token, err := m.LoadToken()
	if err == nil && token != "" {
		return token, nil
	}

	// No token found, prompt for authentication
	fmt.Fprintln(m.output, "\nNo authentication token found.")
	fmt.Fprintln(m.output, "Choose an option:")
	fmt.Fprintln(m.output, "1. Login with existing account")
	fmt.Fprintln(m.output, "2. Register new account")
	fmt.Fprintln(m.output, "3. Exit")
	fmt.Fprint(m.output, "\nEnter choice (1-3): ")

	choice, err := m.input.ReadInput(10)
	if err != nil {
		return "", fmt.Errorf("failed to read choice: %w", err)
	}

	switch choice {
	case "1":
		return m.PromptLogin()
	case "2":
		return m.PromptRegister()
	case "3":
		return "", fmt.Errorf("authentication cancelled by user")
	default:
		return "", fmt.Errorf("invalid choice: %s", choice)
	}
}

// PromptLogin prompts for email/password and calls client.Login
// Saves token automatically after successful login
func (m *FileAuthManager) PromptLogin() (string, error) {
	fmt.Fprintln(m.output, "\n=== Login ===")

	// Prompt for email
	fmt.Fprint(m.output, "Email: ")
	email, err := m.input.ReadInput(100)
	if err != nil {
		return "", fmt.Errorf("failed to read email: %w", err)
	}

	// Prompt for password (masked)
	password, err := m.readPassword("Password: ")
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	// Call client.Login
	token, err := m.client.Login(email, password)
	if err != nil {
		// Check if it's a 401 error
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 401 {
			return "", fmt.Errorf("login failed: invalid credentials")
		}
		return "", fmt.Errorf("login failed: %w", err)
	}

	// Save token
	if err := m.SaveToken(token); err != nil {
		return "", fmt.Errorf("login successful but failed to save token: %w", err)
	}

	fmt.Fprintln(m.output, "✅ Login successful!")
	return token, nil
}

// PromptRegister prompts for email/password and calls client.Register
// Saves token automatically after successful registration
func (m *FileAuthManager) PromptRegister() (string, error) {
	fmt.Fprintln(m.output, "\n=== Register ===")

	// Prompt for email
	fmt.Fprint(m.output, "Email: ")
	email, err := m.input.ReadInput(100)
	if err != nil {
		return "", fmt.Errorf("failed to read email: %w", err)
	}

	// Validate email format before making API call
	if err := validateEmail(email); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Prompt for password (masked)
	password, err := m.readPassword("Password (8-72 characters): ")
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	// Validate password requirements before making API call
	if err := validatePassword(password); err != nil {
		return "", fmt.Errorf("validation failed: %w", err)
	}

	// Prompt for password confirmation (masked)
	confirmPassword, err := m.readPassword("Confirm password: ")
	if err != nil {
		return "", fmt.Errorf("failed to read password confirmation: %w", err)
	}

	// Verify passwords match
	if password != confirmPassword {
		return "", fmt.Errorf("passwords do not match")
	}

	// Call client.Register
	token, err := m.client.Register(email, password)
	if err != nil {
		// Check if it's a conflict error (user already exists)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 409 {
			return "", fmt.Errorf("registration failed: email already registered")
		}
		return "", fmt.Errorf("registration failed: %w", err)
	}

	// Save token
	if err := m.SaveToken(token); err != nil {
		return "", fmt.Errorf("registration successful but failed to save token: %w", err)
	}

	fmt.Fprintln(m.output, "✅ Registration successful!")
	return token, nil
}

// HandleAuthError handles 401 authentication errors by clearing the token and prompting for re-authentication
// Returns a new valid token or error
func (m *FileAuthManager) HandleAuthError() (string, error) {
	// Clear the invalid token
	if err := m.ClearToken(); err != nil {
		fmt.Fprintf(m.output, "⚠️  Warning: failed to clear invalid token: %v\n", err)
	}

	fmt.Fprintln(m.output, "\n🔒 Your session has expired or is invalid.")
	fmt.Fprintln(m.output, "Please authenticate again.")
	fmt.Fprintln(m.output, "\nChoose an option:")
	fmt.Fprintln(m.output, "1. Login")
	fmt.Fprintln(m.output, "2. Register")
	fmt.Fprintln(m.output, "3. Exit")
	fmt.Fprint(m.output, "\nEnter choice (1-3): ")

	choice, err := m.input.ReadInput(10)
	if err != nil {
		return "", fmt.Errorf("failed to read choice: %w", err)
	}

	switch choice {
	case "1":
		return m.PromptLogin()
	case "2":
		return m.PromptRegister()
	case "3":
		return "", fmt.Errorf("re-authentication cancelled by user")
	default:
		return "", fmt.Errorf("invalid choice: %s", choice)
	}
}

// readPassword reads password input with character masking
// Uses golang.org/x/term package for secure terminal password reading
func (m *FileAuthManager) readPassword(prompt string) (string, error) {
	fmt.Fprint(m.output, prompt)

	// Check if stdin is a terminal
	fd := int(syscall.Stdin)
	if !term.IsTerminal(fd) {
		// Not a terminal, fall back to regular input reading
		reader := bufio.NewReader(os.Stdin)
		password, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(password), nil
	}

	// Read password with masking
	passwordBytes, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}

	// Print newline after password input (since ReadPassword doesn't echo)
	fmt.Fprintln(m.output)

	return string(passwordBytes), nil
}

// validateEmail checks if an email address has a valid format
func validateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("invalid email format")
	}

	// Use the same regex pattern as the server-side validation
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}
	if !matched {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// validatePassword checks if a password meets minimum security requirements
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 72 {
		return fmt.Errorf("password must be max 72 characters")
	}

	return nil
}
