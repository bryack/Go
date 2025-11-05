package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Config holds the CLI configuration settings
type Config struct {
	ServerURL string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	// Read server URL from environment variable, default to localhost
	serverURL := os.Getenv("TASK_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	config := &Config{
		ServerURL: serverURL,
	}

	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate ensures the configuration is valid
func (c *Config) Validate() error {
	// Validate server URL format
	if err := validateURL(c.ServerURL); err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	return nil
}

// validateURL checks if the URL is a valid HTTP/HTTPS URL
func validateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	// Check if scheme is http or https
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got: %s", parsedURL.Scheme)
	}

	// Check if host is present
	if parsedURL.Host == "" {
		return fmt.Errorf("URL must include a host")
	}

	return nil
}
