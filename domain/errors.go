package domain

import "errors"

var ErrEmptyFieldsToUpdate = errors.New("at least one field must be provided for update")
var (
	ErrTaskNotFound = errors.New("task not found")
)

var (
	ErrDescriptionRequired = errors.New("description is required")
	ErrDescriptionTooLong  = errors.New("description too long (max 200 characters)")
)

// Authentication errors
var (
	// Ошибки валидации (400 Bad Request)
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must be max 72 bytes")

	// Ошибки конфликта (409 Conflict)
	ErrEmailAlreadyExists = errors.New("email already registered")

	// Ошибки авторизации (401 Unauthorized)
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Internal errors
var (
	ErrHashingFailed         = errors.New("failed to hash password")
	ErrTokenGenerationFailed = errors.New("failed to generate token")
	ErrStorageFailure        = errors.New("storage operation failed")
	ErrUserNotFound          = errors.New("user not found")
)
