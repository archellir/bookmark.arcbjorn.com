package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"torimemo/internal/db"
	"torimemo/internal/models"
	"torimemo/internal/services"
)

type contextKey string

const (
	UserIDKey  contextKey = "user_id"
	UserKey    contextKey = "user"
	ClaimsKey  contextKey = "claims"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	authService *services.AuthService
	userRepo    *db.UserRepository
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(userRepo *db.UserRepository) *AuthMiddleware {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
	}

	return &AuthMiddleware{
		authService: services.NewAuthService(jwtSecret),
		userRepo:    userRepo,
	}
}

// RequireAuth middleware that requires valid JWT token
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.writeError(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		token := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		} else {
			m.writeError(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			m.writeError(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Check session is valid
		tokenHash := m.authService.GenerateTokenHash(token)
		_, err = m.userRepo.GetSession(tokenHash)
		if err != nil {
			m.writeError(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		// Get user
		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil {
			m.writeError(w, "User not found", http.StatusUnauthorized)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserKey, user)
		ctx = context.WithValue(ctx, ClaimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserKey).(*models.User)
		if !ok || !user.IsAdmin {
			m.writeError(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}))
}

// OptionalAuth middleware that extracts user info if token is present but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		token := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		} else {
			next.ServeHTTP(w, r)
			return
		}

		// Validate token
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Check session is valid
		tokenHash := m.authService.GenerateTokenHash(token)
		_, err = m.userRepo.GetSession(tokenHash)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Get user
		user, err := m.userRepo.GetByID(claims.UserID)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserKey, user)
		ctx = context.WithValue(ctx, ClaimsKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext gets user from request context
func GetUserFromContext(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(UserKey).(*models.User)
	return user, ok
}

// GetUserIDFromContext gets user ID from request context
func GetUserIDFromContext(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	return userID, ok
}

// GetUserIDStringFromContext returns user ID as string for compatibility with monitoring
func GetUserIDStringFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(int); ok {
		return fmt.Sprintf("%d", userID)
	}
	return ""
}

// GetClaimsFromContext gets JWT claims from request context
func GetClaimsFromContext(r *http.Request) (*services.Claims, bool) {
	claims, ok := r.Context().Value(ClaimsKey).(*services.Claims)
	return claims, ok
}

// writeError writes an error response
func (m *AuthMiddleware) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(`{"error":"` + message + `","status":` + string(rune(statusCode)) + `}`))
}