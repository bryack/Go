package auth

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	testCases := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "Valid email",
			email:       "user@example.com",
			expectError: false,
		},
		{
			name:        "Valid email with subdomain",
			email:       "user@mail.example.com",
			expectError: false,
		},
		{
			name:        "Invalid email - no @",
			email:       "userexample.com",
			expectError: true,
		},
		{
			name:        "Invalid email - empty",
			email:       "",
			expectError: true,
		},
		{
			name:        "Invalid email - no domain",
			email:       "user@",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateEmail(tc.email)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for email %q, but got none", tc.email)
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for email %q, but got: %v", tc.email, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	testCases := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "Valid password - 8 characters",
			password:    "12345678",
			expectError: false,
		},
		{
			name:        "Valid password - 72 characters",
			password:    "123456789012345678901234567890123456789012345678901234567890123456789012",
			expectError: false,
		},
		{
			name:        "Invalid password - too short",
			password:    "1234567",
			expectError: true,
		},
		{
			name:        "Invalid password - empty",
			password:    "",
			expectError: true,
		},
		{
			name:        "Invalid password - too long",
			password:    "1234567890123456789012345678901234567890123456789012345678901234567890123",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePassword(tc.password)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for password length %d, but got none", len(tc.password))
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for password length %d, but got: %v", len(tc.password), err)
			}
		})
	}
}
