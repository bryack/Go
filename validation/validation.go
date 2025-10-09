package validation

import (
	"errors"
	"strconv"
)

var ErrInvalidTaskID = errors.New("invalid task ID")

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
