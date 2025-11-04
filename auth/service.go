package auth

import (
	"errors"
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
		return ErrPasswordTooShort
	}

	if len(password) > 72 {
		return ErrPasswordTooLong
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
		return "", ErrHashingFailed
	}

	return string(hashedPassword), nil
}

// ComparePassword verifies if the provided password matches the stored hash.
func ComparePassword(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return ErrInvalidCredentials
	}
	return nil
}

// Register creates a new user account with the provided credentials and returns a JWT token.
func (service *Service) Register(email, password string) (token string, err error) {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return "", ErrInvalidEmail
	}

	if err = ValidatePassword(password); err != nil {
		return "", ErrInvalidCredentials
	}

	exists, err := service.userStorage.EmailExists(email)
	if err != nil {
		return "", ErrStorageFailure
	}

	if exists {
		return "", ErrEmailAlreadyExists
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return "", ErrHashingFailed
	}

	userID, err := service.userStorage.CreateUser(email, passwordHash)
	if err != nil {
		return "", ErrStorageFailure
	}

	token, err = service.jwtService.GenerateToken(userID)
	if err != nil {
		return "", ErrTokenGenerationFailed
	}

	return token, nil
}

// Login authenticates a user with email and password, returning a JWT token on success.
func (service *Service) Login(email, password string) (token string, err error) {
	user, err := service.userStorage.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", ErrStorageFailure
	}

	if err = ComparePassword(user.PasswordHash, password); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err = service.jwtService.GenerateToken(user.ID)
	if err != nil {
		return "", ErrTokenGenerationFailed
	}

	return token, nil
}
