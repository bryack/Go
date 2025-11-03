package auth

import (
	"fmt"
	"myproject/storage"

	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication operations including user registration and login.
type Service struct {
	userStorage storage.UserStorage
	jwtService  *JWTService
}

// NewService creates a new authentication service with the provided dependencies.
func NewService(userStorage storage.UserStorage, jwtService *JWTService) *Service {
	return &Service{
		userStorage: userStorage,
		jwtService:  jwtService,
	}
}

// ValidatePassword checks if a password meets minimum security requirements.
func (service *Service) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 72 {
		return fmt.Errorf("password must be max 72 bytes")
	}
	return nil
}

// HashPassword creates a bcrypt hash of the provided password for secure storage.
func (service *Service) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", fmt.Errorf("hashing password failed: %w", err)
	}

	return string(hashedPassword), nil
}

// ComparePassword verifies if the provided password matches the stored hash.
func (service *Service) ComparePassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid credentials")
	}
	return nil
}
