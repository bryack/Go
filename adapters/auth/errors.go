package auth

import "errors"

// Ошибки валидации (400 Bad Request)
var (
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong  = errors.New("password must be max 72 bytes")
)

// Ошибки конфликта (409 Conflict)
var (
	ErrEmailAlreadyExists = errors.New("email already registered")
)

// Ошибки авторизации (401 Unauthorized)
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Внутренние ошибки (500 Internal Server Error)
var (
	ErrHashingFailed         = errors.New("failed to hash password")
	ErrTokenGenerationFailed = errors.New("failed to generate token")
	ErrStorageFailure        = errors.New("storage operation failed")
)
