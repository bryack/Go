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
