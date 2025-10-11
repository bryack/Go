package validation

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrInvalidTaskID = errors.New("invalid task ID")
)

// ValidateTaskID converts a string input to a valid task ID.
// Returns the parsed ID if valid (positive integer), or an error if invalid.
func ValidateTaskID(input string) (int, error) {
	id, err := strconv.Atoi(input)
	if err != nil {
		return 0, ErrInvalidTaskID
	}
	if id <= 0 {
		return 0, ErrInvalidTaskID
	}
	return id, nil
}

func ValidateTaskDescription(input string) (string, error) {
	if len(input) == 0 {
		return "", errors.New("description is required")
	}

	input = strings.TrimSpace(input)
	if len(input) > 200 {
		return "", errors.New("description too long (max 200 characters)")
	}

	return input, nil
}

func ExtractTaskIDFromPath(path string) (int, error) {
	if !strings.HasPrefix(path, "/tasks/") {
		return 0, ErrInvalidTaskID
	}

	idStr := strings.TrimPrefix(path, "/tasks/")
	if idStr == "" || strings.Contains(idStr, "/") {
		return 0, ErrInvalidTaskID
	}

	id, err := ValidateTaskID(idStr)
	if err != nil {
		return 0, ErrInvalidTaskID
	}

	return id, nil
}
