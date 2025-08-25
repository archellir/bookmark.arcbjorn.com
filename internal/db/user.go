package db

import (
	"database/sql"
	"fmt"
	"time"

	"torimemo/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(req *models.CreateUserRequest) (*models.User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, is_admin)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, username, email, full_name, is_active, is_admin, created_at, updated_at, last_login_at
	`

	var user models.User
	err := r.db.QueryRow(query, req.Username, req.Email, req.Password, req.FullName, req.IsAdmin).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, is_active, is_admin, 
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE id = ? AND is_active = TRUE
	`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByUsername gets a user by username
func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE username = ? AND is_active = TRUE
	`

	var user models.User
	err := r.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users 
		WHERE email = ? AND is_active = TRUE
	`

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(id int, req *models.UpdateUserRequest) (*models.User, error) {
	setParts := []string{}
	args := []interface{}{}
	
	if req.Username != nil {
		setParts = append(setParts, "username = ?")
		args = append(args, *req.Username)
	}
	if req.Email != nil {
		setParts = append(setParts, "email = ?")
		args = append(args, *req.Email)
	}
	if req.Password != nil {
		setParts = append(setParts, "password_hash = ?")
		args = append(args, *req.Password)
	}
	if req.FullName != nil {
		setParts = append(setParts, "full_name = ?")
		args = append(args, *req.FullName)
	}
	if req.IsActive != nil {
		setParts = append(setParts, "is_active = ?")
		args = append(args, *req.IsActive)
	}
	if req.IsAdmin != nil {
		setParts = append(setParts, "is_admin = ?")
		args = append(args, *req.IsAdmin)
	}

	if len(setParts) == 0 {
		return r.GetByID(id)
	}

	// Add updated_at
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE users SET %s WHERE id = ?
	`, fmt.Sprintf("%s", setParts[0]))
	
	for i := 1; i < len(setParts); i++ {
		query = fmt.Sprintf("%s, %s", query, setParts[i])
	}

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return r.GetByID(id)
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(id int) error {
	query := `UPDATE users SET last_login_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

// Delete soft-deletes a user by setting is_active to false
func (r *UserRepository) Delete(id int) error {
	query := `UPDATE users SET is_active = FALSE WHERE id = ?`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrUserNotFound
	}

	return nil
}

// CreateSession creates a new user session
func (r *UserRepository) CreateSession(userID int, tokenHash string, expiresAt time.Time) (*models.UserSession, error) {
	query := `
		INSERT INTO user_sessions (user_id, token_hash, expires_at)
		VALUES (?, ?, ?)
		RETURNING id, user_id, token_hash, expires_at, created_at, is_revoked
	`

	var session models.UserSession
	err := r.db.QueryRow(query, userID, tokenHash, expiresAt).Scan(
		&session.ID, &session.UserID, &session.TokenHash,
		&session.ExpiresAt, &session.CreatedAt, &session.IsRevoked,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &session, nil
}

// GetSession gets a session by token hash
func (r *UserRepository) GetSession(tokenHash string) (*models.UserSession, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, created_at, is_revoked
		FROM user_sessions 
		WHERE token_hash = ? AND is_revoked = FALSE AND expires_at > CURRENT_TIMESTAMP
	`

	var session models.UserSession
	err := r.db.QueryRow(query, tokenHash).Scan(
		&session.ID, &session.UserID, &session.TokenHash,
		&session.ExpiresAt, &session.CreatedAt, &session.IsRevoked,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, models.ErrTokenInvalid
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// RevokeSession revokes a session by token hash
func (r *UserRepository) RevokeSession(tokenHash string) error {
	query := `UPDATE user_sessions SET is_revoked = TRUE WHERE token_hash = ?`
	_, err := r.db.Exec(query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *UserRepository) RevokeAllUserSessions(userID int) error {
	query := `UPDATE user_sessions SET is_revoked = TRUE WHERE user_id = ?`
	_, err := r.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}
	return nil
}

// CleanupExpiredSessions removes expired and revoked sessions
func (r *UserRepository) CleanupExpiredSessions() error {
	query := `DELETE FROM user_sessions WHERE expires_at < CURRENT_TIMESTAMP OR is_revoked = TRUE`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	return nil
}