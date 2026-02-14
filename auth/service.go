package auth

import (
	"errors"
	"log/slog"
	"myproject/adapters/storage"
	"myproject/logger"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication operations including user registration and login.
type Service struct {
	userStorage storage.UserStorage
	jwtService  *JWTService
	logger      *slog.Logger
}

// NewService creates a new authentication service with the provided dependencies.
func NewService(userStorage storage.UserStorage, jwtService *JWTService, logger *slog.Logger) *Service {
	return &Service{
		userStorage: userStorage,
		jwtService:  jwtService,
		logger:      logger,
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
	service.logger.Info("Register",
		slog.String(logger.FieldOperation, "user_registration"),
		slog.String(logger.FieldEmail, logger.MaskEmail(email)),
	)

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		service.logger.Warn("Failed to validate email",
			slog.String(logger.FieldOperation, "user_registration"),
			slog.String(logger.FieldEmail, logger.MaskEmail(email)),
			slog.String(logger.FieldError, ErrInvalidEmail.Error()),
		)
		return "", ErrInvalidEmail
	}

	if err = ValidatePassword(password); err != nil {
		service.logger.Warn("Failed to validate password",
			slog.String(logger.FieldOperation, "user_registration"),
			slog.String(logger.FieldEmail, logger.MaskEmail(email)),
			slog.String(logger.FieldError, err.Error()),
		)
		return "", ErrInvalidCredentials
	}

	exists, err := service.userStorage.EmailExists(email)
	if err != nil {
		service.logger.Error("Failed to check email existence in database",
			slog.String(logger.FieldOperation, "user_registration"),
			slog.String(logger.FieldEmail, logger.MaskEmail(email)),
			slog.String(logger.FieldError, err.Error()),
		)
		return "", ErrStorageFailure
	}

	if exists {
		service.logger.Warn("Email exists",
			slog.String(logger.FieldOperation, "user_registration"),
			slog.String(logger.FieldEmail, logger.MaskEmail(email)),
			slog.String(logger.FieldError, ErrEmailAlreadyExists.Error()),
		)
		return "", ErrEmailAlreadyExists
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		service.logger.Error("Failed to hash password",
			slog.String(logger.FieldOperation, "user_registration"),
			slog.String(logger.FieldEmail, logger.MaskEmail(email)),
			slog.String(logger.FieldError, err.Error()),
		)
		return "", ErrHashingFailed
	}

	userID, err := service.userStorage.CreateUser(email, passwordHash)
	if err != nil {
		service.logger.Error("Failed to create user in database",
			slog.String(logger.FieldOperation, "user_registration"),
			slog.String(logger.FieldEmail, logger.MaskEmail(email)),
			slog.String(logger.FieldError, err.Error()),
		)
		return "", ErrStorageFailure
	}

	token, err = service.jwtService.GenerateToken(userID)
	if err != nil {
		return "", ErrTokenGenerationFailed
	}

	service.logger.Info("User registered successfully",
		slog.String(logger.FieldOperation, "user_registration"),
		slog.String(logger.FieldEmail, logger.MaskEmail(email)),
		slog.Int(logger.FieldUserID, userID),
	)

	return token, nil
}

// Login authenticates a user with email and password, returning a JWT token on success.
func (service *Service) Login(email, password string) (token string, err error) {
	service.logger.Info("Login attempt",
		slog.String(logger.FieldOperation, "user_login"),
		slog.String(logger.FieldEmail, logger.MaskEmail(email)),
	)

	user, err := service.userStorage.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			service.logger.Warn("Failed login",
				slog.String(logger.FieldOperation, "user_login"),
				slog.String(logger.FieldEmail, logger.MaskEmail(email)),
				slog.String(logger.FieldError, ErrInvalidCredentials.Error()),
			)
			return "", ErrInvalidCredentials
		}
		service.logger.Error("Failed to fetch user by email from database",
			slog.String(logger.FieldOperation, "user_login"),
			slog.String(logger.FieldEmail, logger.MaskEmail(email)),
			slog.String(logger.FieldError, err.Error()),
		)
		return "", ErrStorageFailure
	}

	if err = ComparePassword(user.PasswordHash, password); err != nil {
		service.logger.Warn("Failed login",
			slog.String(logger.FieldOperation, "user_login"),
			slog.String(logger.FieldEmail, logger.MaskEmail(email)),
			slog.String(logger.FieldError, ErrInvalidCredentials.Error()),
		)
		return "", ErrInvalidCredentials
	}

	token, err = service.jwtService.GenerateToken(user.ID)
	if err != nil {
		service.logger.Error("Failed to generate token",
			slog.String(logger.FieldOperation, "user_login"),
			slog.String(logger.FieldEmail, logger.MaskEmail(email)),
			slog.String(logger.FieldError, err.Error()),
		)
		return "", ErrTokenGenerationFailed
	}

	service.logger.Info("Login successful",
		slog.String(logger.FieldOperation, "user_login"),
		slog.String(logger.FieldEmail, logger.MaskEmail(email)),
		slog.Int(logger.FieldUserID, user.ID),
	)

	return token, nil
}
