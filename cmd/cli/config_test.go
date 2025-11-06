package main

import (
	"os"
	"testing"
)

func TestLoadConfig_DefaultURL(t *testing.T) {
	// Clear environment variable
	os.Unsetenv("TASK_SERVER_URL")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	expectedURL := "http://localhost:8080"
	if config.ServerURL != expectedURL {
		t.Errorf("Expected ServerURL to be %s, got %s", expectedURL, config.ServerURL)
	}
}

func TestLoadConfig_CustomURL(t *testing.T) {
	// Set custom URL
	customURL := "https://api.example.com:9090"
	os.Setenv("TASK_SERVER_URL", customURL)
	defer os.Unsetenv("TASK_SERVER_URL")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config.ServerURL != customURL {
		t.Errorf("Expected ServerURL to be %s, got %s", customURL, config.ServerURL)
	}
}

func TestValidateURL_ValidURLs(t *testing.T) {
	validURLs := []string{
		"http://localhost:8080",
		"https://localhost:8080",
		"http://example.com",
		"https://api.example.com:9090",
		"http://192.168.1.1:3000",
	}

	for _, url := range validURLs {
		t.Run(url, func(t *testing.T) {
			err := validateURL(url)
			if err != nil {
				t.Errorf("Expected URL %s to be valid, got error: %v", url, err)
			}
		})
	}
}

func TestValidateURL_InvalidURLs(t *testing.T) {
	invalidURLs := []string{
		"",
		"ftp://localhost:8080",
		"localhost:8080",
		"http://",
		"https://",
		"ws://localhost:8080",
	}

	for _, url := range invalidURLs {
		t.Run(url, func(t *testing.T) {
			err := validateURL(url)
			if err == nil {
				t.Errorf("Expected URL %s to be invalid", url)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		config := &Config{
			ServerURL: "http://localhost:8080",
		}
		err := config.Validate()
		if err != nil {
			t.Errorf("Expected config to be valid, got error: %v", err)
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		config := &Config{
			ServerURL: "invalid-url",
		}
		err := config.Validate()
		if err == nil {
			t.Error("Expected config to be invalid")
		}
	})
}
