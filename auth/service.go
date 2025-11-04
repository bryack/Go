package auth

import (
	"errors"
	"fmt"
	"myproject/storage"
	"regexp"

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
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 72 {
		return fmt.Errorf("password must be max 72 bytes")
	}
	return nil
}

// HashPassword creates a bcrypt hash of the provided password for secure storage.
func HashPassword(password string) (string, error) {
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
func ComparePassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid credentials")
	}
	return nil
}

// Register creates a new user account with the provided credentials and returns a JWT token.
func (service *Service) Register(email, password string) (token string, err error) {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return "", fmt.Errorf("invalid email format")
	}

	if err = ValidatePassword(password); err != nil {
		return "", err
	}

	exists, err := service.userStorage.EmailExists(email)
	if err != nil {
		return "", fmt.Errorf("failed to check email availability: %w", err)
	}

	if exists {
		return "", fmt.Errorf("email %s already registered", email)
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return "", err
	}

	userID, err := service.userStorage.CreateUser(email, passwordHash)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	token, err = service.jwtService.GenerateToken(userID)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// Login authenticates a user with email and password, returning a JWT token on success.
func (service *Service) Login(email, password string) (token string, err error) {
	user, err := service.userStorage.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", fmt.Errorf("invalid credentials")
		}
		return "", err
	}

	if err = ComparePassword(user.PasswordHash, password); err != nil {
		return "", fmt.Errorf("invalid credentials")
	}

	token, err = service.jwtService.GenerateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}
