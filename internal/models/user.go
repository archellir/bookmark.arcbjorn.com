package models

import (
	"time"
)

// User represents a user account
type User struct {
	ID          int        `json:"id" db:"id"`
	Username    string     `json:"username" db:"username"`
	Email       string     `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Never expose password hash
	FullName    *string    `json:"full_name,omitempty" db:"full_name"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	IsAdmin     bool       `json:"is_admin" db:"is_admin"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
}

// UserSession represents a user session for JWT token management
type UserSession struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	TokenHash string    `json:"-" db:"token_hash"` // Never expose token hash
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IsRevoked bool      `json:"is_revoked" db:"is_revoked"`
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Username string  `json:"username"`
	Email    string  `json:"email"`
	Password string  `json:"password"`
	FullName *string `json:"full_name,omitempty"`
	IsAdmin  bool    `json:"is_admin,omitempty"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
	FullName *string `json:"full_name,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
	IsAdmin  *bool   `json:"is_admin,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}

// ValidateCreateUser validates user creation request
func (r *CreateUserRequest) Validate() error {
	if r.Username == "" {
		return ErrInvalidUsername
	}
	if r.Email == "" {
		return ErrInvalidEmail
	}
	if r.Password == "" {
		return ErrInvalidPassword
	}
	if len(r.Password) < 6 {
		return ErrPasswordTooShort
	}
	return nil
}

// ValidateLogin validates login request
func (r *LoginRequest) Validate() error {
	if r.Username == "" {
		return ErrInvalidUsername
	}
	if r.Password == "" {
		return ErrInvalidPassword
	}
	return nil
}