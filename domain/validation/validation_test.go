package validation

import (
	"errors"
	"testing"
)

func TestValidateTaskID(t *testing.T) {
	// ====Arrange====
	testCases := []struct {
		name        string
		input       string
		expectedID  int
		expectedErr error
	}{
		{
			name:        "Valid ID string",
			input:       "1",
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name:        "Maximum int value",
			input:       "9223372036854775807",
			expectedID:  9223372036854775807,
			expectedErr: nil,
		},
		{
			name:        "Invalid ID with letters",
			input:       "abc",
			expectedID:  0,
			expectedErr: ErrInvalidTaskID,
		},
		{
			name:        "ID with trailing letters",
			input:       "123abc",
			expectedID:  0,
			expectedErr: ErrInvalidTaskID,
		},
		{
			name:        "Empty ID",
			input:       "",
			expectedID:  0,
			expectedErr: ErrInvalidTaskID,
		},
		{
			name:        "Empty ID with spaces",
			input:       "    ",
			expectedID:  0,
			expectedErr: ErrInvalidTaskID,
		},
		{
			name:        "Zero ID",
			input:       "0",
			expectedID:  0,
			expectedErr: ErrInvalidTaskID,
		},
		{
			name:        "Negative ID",
			input:       "-1",
			expectedID:  0,
			expectedErr: ErrInvalidTaskID,
		},
		{
			name:        "Invalid ID with special characters",
			input:       "#@`[]$%^*",
			expectedID:  0,
			expectedErr: ErrInvalidTaskID,
		},
		{
			name:        "Invalid ID with Unicode characters",
			input:       "1️⃣",
			expectedID:  0,
			expectedErr: ErrInvalidTaskID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ====Act====
			id, err := ValidateTaskID(tc.input)

			// ====Assert====
			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected %v, got %v", tc.expectedErr, err)
			}

			if id != tc.expectedID {
				t.Errorf("Expected ID %d, got %d", tc.expectedID, id)
			}
		})
	}
}

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
			name:        "Valid email with plus",
			email:       "user+tag@example.com",
			expectError: false,
		},
		{
			name:        "Valid email with dots",
			email:       "first.last@example.com",
			expectError: false,
		},
		{
			name:        "Valid email with numbers",
			email:       "user123@example123.com",
			expectError: false,
		},
		{
			name:        "Invalid email - no @",
			email:       "userexample.com",
			expectError: true,
		},
		{
			name:        "Invalid email - no domain",
			email:       "user@",
			expectError: true,
		},
		{
			name:        "Invalid email - no TLD",
			email:       "user@example",
			expectError: true,
		},
		{
			name:        "Invalid email - empty",
			email:       "",
			expectError: true,
		},
		{
			name:        "Invalid email - spaces",
			email:       "user @example.com",
			expectError: true,
		},
		{
			name:        "Invalid email - TLD too short",
			email:       "user@example.c",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEmail(tc.email)

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
		expectedErr error
	}{
		{
			name:        "Valid password - 8 characters",
			password:    "12345678",
			expectedErr: nil,
		},
		{
			name:        "Valid password - 72 characters",
			password:    "123456789012345678901234567890123456789012345678901234567890123456789012",
			expectedErr: nil,
		},
		{
			name:        "Valid password - medium length",
			password:    "MySecurePassword123",
			expectedErr: nil,
		},
		{
			name:        "Invalid password - too short (7 chars)",
			password:    "1234567",
			expectedErr: ErrPasswordTooShort,
		},
		{
			name:        "Invalid password - too short (empty)",
			password:    "",
			expectedErr: ErrPasswordTooShort,
		},
		{
			name:        "Invalid password - too long (73 chars)",
			password:    "1234567890123456789012345678901234567890123456789012345678901234567890123",
			expectedErr: ErrPasswordTooLong,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.password)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}
