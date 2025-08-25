package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"

	"torimemo/internal/models"
)

const (
	// JWT token expiration time
	TokenExpiry = 24 * time.Hour
	// Cost for bcrypt hashing
	BcryptCost = 12
)

// AuthService handles user authentication
type AuthService struct {
	jwtSecret []byte
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret string) *AuthService {
	return &AuthService{
		jwtSecret: []byte(jwtSecret),
	}
}

// Claims represents JWT token claims
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	return string(bytes), err
}

// CheckPassword checks if the provided password matches the hash
func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateToken generates a JWT token for a user
func (s *AuthService) GenerateToken(user *models.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(TokenExpiry)
	
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "torimemo",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, models.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, models.ErrTokenInvalid
	}

	// Check if token is expired
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, models.ErrTokenExpired
	}

	return claims, nil
}

// GenerateTokenHash generates a hash of the token for session storage
func (s *AuthService) GenerateTokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateSecureToken generates a cryptographically secure random token
func (s *AuthService) GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// RefreshToken generates a new token with extended expiry
func (s *AuthService) RefreshToken(claims *Claims) (string, time.Time, error) {
	// Create new claims with updated expiry
	newExpiresAt := time.Now().Add(TokenExpiry)
	
	newClaims := &Claims{
		UserID:   claims.UserID,
		Username: claims.Username,
		IsAdmin:  claims.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(newExpiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "torimemo",
			Subject:   fmt.Sprintf("%d", claims.UserID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, newExpiresAt, nil
}