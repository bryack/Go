package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT token payload containing user identification and standard claims.
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token generation and validation operations.
type JWTService struct {
	secretKey  []byte
	expiration time.Duration
}

// NewJWTService creates a new JWT service with the provided secret key and token expiration duration.
func NewJWTService(secret string, expiration time.Duration) *JWTService {
	secretKey := []byte(secret)
	return &JWTService{
		secretKey:  secretKey,
		expiration: expiration,
	}
}

// GenerateToken creates a signed JWT token for the specified user ID with configured expiration.
func (j *JWTService) GenerateToken(userID int) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken verifies the token signature and expiration, returning the extracted claims.
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method, got %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims type")
	}

	return claims, nil
}
