package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"torimemo/internal/db"
	"torimemo/internal/models"
	"torimemo/internal/services"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	userRepo    *db.UserRepository
	authService *services.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo *db.UserRepository) *AuthHandler {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-this-in-production"
	}

	return &AuthHandler{
		userRepo:    userRepo,
		authService: services.NewAuthService(jwtSecret),
	}
}

// RegisterRoutes registers auth-related routes
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/auth/register", h.register)
	mux.HandleFunc("/api/auth/login", h.login)
	mux.HandleFunc("/api/auth/logout", h.logout)
	mux.HandleFunc("/api/auth/refresh", h.refresh)
	mux.HandleFunc("/api/auth/me", h.getCurrentUser)
}

// register handles POST /api/auth/register
func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user already exists
	if _, err := h.userRepo.GetByUsername(req.Username); err == nil {
		h.writeError(w, "Username already exists", http.StatusConflict)
		return
	}

	if _, err := h.userRepo.GetByEmail(req.Email); err == nil {
		h.writeError(w, "Email already exists", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		h.writeError(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	req.Password = hashedPassword

	// Create user
	user, err := h.userRepo.Create(&req)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate token
	token, expiresAt, err := h.authService.GenerateToken(user)
	if err != nil {
		h.writeError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Create session
	tokenHash := h.authService.GenerateTokenHash(token)
	_, err = h.userRepo.CreateSession(user.ID, tokenHash, expiresAt)
	if err != nil {
		h.writeError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	}

	h.writeJSON(w, response)
}

// login handles POST /api/auth/login
func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.writeError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user by username
	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		h.writeError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Check password
	if !h.authService.CheckPassword(req.Password, user.PasswordHash) {
		h.writeError(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, expiresAt, err := h.authService.GenerateToken(user)
	if err != nil {
		h.writeError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Create session
	tokenHash := h.authService.GenerateTokenHash(token)
	_, err = h.userRepo.CreateSession(user.ID, tokenHash, expiresAt)
	if err != nil {
		h.writeError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Update last login
	h.userRepo.UpdateLastLogin(user.ID)

	response := models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      *user,
	}

	h.writeJSON(w, response)
}

// logout handles POST /api/auth/logout
func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeError(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		h.writeError(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	// Revoke session
	tokenHash := h.authService.GenerateTokenHash(token)
	err := h.userRepo.RevokeSession(tokenHash)
	if err != nil {
		h.writeError(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	}

	h.writeJSON(w, response)
}

// refresh handles POST /api/auth/refresh
func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeError(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		h.writeError(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	// Validate current token
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Check session is valid
	tokenHash := h.authService.GenerateTokenHash(token)
	_, err = h.userRepo.GetSession(tokenHash)
	if err != nil {
		h.writeError(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Generate new token
	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil {
		h.writeError(w, "User not found", http.StatusUnauthorized)
		return
	}

	newToken, expiresAt, err := h.authService.GenerateToken(user)
	if err != nil {
		h.writeError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Revoke old session and create new one
	h.userRepo.RevokeSession(tokenHash)
	newTokenHash := h.authService.GenerateTokenHash(newToken)
	_, err = h.userRepo.CreateSession(user.ID, newTokenHash, expiresAt)
	if err != nil {
		h.writeError(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		Token:     newToken,
		ExpiresAt: expiresAt,
		User:      *user,
	}

	h.writeJSON(w, response)
}

// getCurrentUser handles GET /api/auth/me
func (h *AuthHandler) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get token from header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeError(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		h.writeError(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	// Validate token
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Check session is valid
	tokenHash := h.authService.GenerateTokenHash(token)
	_, err = h.userRepo.GetSession(tokenHash)
	if err != nil {
		h.writeError(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Get user
	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil {
		h.writeError(w, "User not found", http.StatusNotFound)
		return
	}

	h.writeJSON(w, user)
}

// writeJSON writes a JSON response
func (h *AuthHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *AuthHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}

	json.NewEncoder(w).Encode(response)
}