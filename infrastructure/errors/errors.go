package errors

import "errors"

var ErrEmptyFieldsToUpdate = errors.New("at least one field must be provided for update")
var (
	ErrTaskNotFound = errors.New("task not found")
)

var (
	ErrDescriptionRequired = errors.New("description is required")
	ErrDescriptionTooLong  = errors.New("description too long (max 200 characters)")
)
