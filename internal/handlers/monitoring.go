package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"torimemo/internal/db"
	"torimemo/internal/logger"
	"torimemo/internal/middleware"
)

// MonitoringHandler handles monitoring-related requests
type MonitoringHandler struct {
	userRepo *db.UserRepository
	db       *db.DB
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(userRepo *db.UserRepository, database *db.DB) *MonitoringHandler {
	return &MonitoringHandler{
		userRepo: userRepo,
		db:       database,
	}
}

// RegisterRoutes registers monitoring-related routes
func (h *MonitoringHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/errors", h.handleErrors)
	mux.HandleFunc("/api/metrics", h.handleMetrics)
	mux.HandleFunc("/api/security-violations", h.handleSecurityViolations)
}

// ErrorInfo represents frontend error information
type ErrorInfo struct {
	Message   string `json:"message"`
	Stack     string `json:"stack"`
	Timestamp int64  `json:"timestamp"`
	UserAgent string `json:"userAgent"`
	URL       string `json:"url"`
	UserID    string `json:"userId,omitempty"`
}

// ErrorReport represents a batch of errors from frontend
type ErrorReport struct {
	Errors []ErrorInfo `json:"errors"`
}

// PerformanceMetric represents frontend performance data
type PerformanceMetric struct {
	Name      string  `json:"name"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"timestamp"`
	URL       string  `json:"url"`
	UserID    string  `json:"userId,omitempty"`
}

// MetricsReport represents performance metrics from frontend
type MetricsReport struct {
	Metrics []PerformanceMetric            `json:"metrics"`
	Summary map[string]MetricSummary       `json:"summary"`
}

// MetricSummary represents aggregated metric data
type MetricSummary struct {
	Avg   float64 `json:"avg"`
	Max   float64 `json:"max"`
	Min   float64 `json:"min"`
	Count int     `json:"count"`
}

// SecurityViolation represents a security violation report
type SecurityViolation struct {
	Type      string                 `json:"type"`
	Details   map[string]interface{} `json:"details"`
	Timestamp int64                  `json:"timestamp"`
	UserAgent string                 `json:"userAgent"`
	URL       string                 `json:"url"`
	UserID    string                 `json:"userId,omitempty"`
}

// handleErrors processes error reports from frontend
func (h *MonitoringHandler) handleErrors(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var report ErrorReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get user ID from token if available
	userID := h.getUserIDFromRequest(r)

	// Process each error
	for _, errorInfo := range report.Errors {
		// Log error for monitoring
		logger.Error("Frontend error reported", map[string]interface{}{
			"message":    errorInfo.Message,
			"stack":      errorInfo.Stack,
			"timestamp":  time.Unix(errorInfo.Timestamp/1000, 0),
			"user_agent": errorInfo.UserAgent,
			"url":        errorInfo.URL,
			"user_id":    userID,
		})

		// Store in database for analysis (simplified version)
		h.storeError(errorInfo, userID)
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Processed %d error reports", len(report.Errors)),
	}

	h.writeJSON(w, response)
}

// handleMetrics processes performance metrics from frontend
func (h *MonitoringHandler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var report MetricsReport
	if err := json.NewDecoder(r.Body).Decode(&report); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get user ID from token if available
	userID := h.getUserIDFromRequest(r)

	// Process metrics
	for _, metric := range report.Metrics {
		// Log important performance metrics
		if h.isImportantMetric(metric.Name) {
			logger.Info("Performance metric", map[string]interface{}{
				"name":      metric.Name,
				"value":     metric.Value,
				"timestamp": time.Unix(metric.Timestamp/1000, 0),
				"url":       metric.URL,
				"user_id":   userID,
			})
		}

		// Store in database for analysis
		h.storeMetric(metric, userID)
	}

	// Process summary data for aggregated insights
	h.processSummaryMetrics(report.Summary, userID)

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Processed %d performance metrics", len(report.Metrics)),
	}

	h.writeJSON(w, response)
}

// handleSecurityViolations processes security violation reports
func (h *MonitoringHandler) handleSecurityViolations(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var violation SecurityViolation
	if err := json.NewDecoder(r.Body).Decode(&violation); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get user ID from token if available
	userID := h.getUserIDFromRequest(r)

	// Log security violation with high priority
	logger.Warn("Security violation reported", map[string]interface{}{
		"type":       violation.Type,
		"details":    violation.Details,
		"timestamp":  time.Unix(violation.Timestamp/1000, 0),
		"user_agent": violation.UserAgent,
		"url":        violation.URL,
		"user_id":    userID,
	})

	// Store security violation for analysis
	h.storeSecurityViolation(violation, userID)

	response := map[string]interface{}{
		"success": true,
		"message": "Security violation report processed",
	}

	h.writeJSON(w, response)
}

// getUserIDFromRequest extracts user ID from JWT token if present
func (h *MonitoringHandler) getUserIDFromRequest(r *http.Request) string {
	// Try to get user from context (set by auth middleware)
	if userID := middleware.GetUserIDStringFromContext(r.Context()); userID != "" {
		return userID
	}

	// For now, return "anonymous" for unauthenticated users
	// This allows monitoring to work without breaking the current single-user setup
	return "anonymous"
}

// isImportantMetric determines if a metric should be logged
func (h *MonitoringHandler) isImportantMetric(name string) bool {
	importantMetrics := map[string]bool{
		"LCP":              true, // Largest Contentful Paint
		"FID":              true, // First Input Delay
		"CLS":              true, // Cumulative Layout Shift
		"FCP":              true, // First Contentful Paint
		"TTFB":             true, // Time to First Byte
		"AppInit":          true, // App initialization time
		"DOMContentLoaded": true, // DOM loaded time
		"LoadComplete":     true, // Full load time
	}
	return importantMetrics[name]
}

// storeError stores error information in database (simplified)
func (h *MonitoringHandler) storeError(errorInfo ErrorInfo, userID string) {
	// In a production system, you'd store this in a dedicated errors table
	// For now, we'll just ensure it's logged
	query := `
		INSERT OR IGNORE INTO error_logs (message, stack, timestamp, user_agent, url, user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
	`
	
	// Create error_logs table if it doesn't exist
	h.db.Exec(`
		CREATE TABLE IF NOT EXISTS error_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message TEXT NOT NULL,
			stack TEXT,
			timestamp INTEGER NOT NULL,
			user_agent TEXT,
			url TEXT,
			user_id TEXT,
			created_at DATETIME NOT NULL
		)
	`)

	h.db.Exec(query, errorInfo.Message, errorInfo.Stack, errorInfo.Timestamp, 
		errorInfo.UserAgent, errorInfo.URL, userID)
}

// storeMetric stores performance metric in database (simplified)
func (h *MonitoringHandler) storeMetric(metric PerformanceMetric, userID string) {
	// In a production system, you'd store this in a dedicated metrics table
	query := `
		INSERT OR IGNORE INTO performance_metrics (name, value, timestamp, url, user_id, created_at)
		VALUES (?, ?, ?, ?, ?, datetime('now'))
	`
	
	// Create performance_metrics table if it doesn't exist
	h.db.Exec(`
		CREATE TABLE IF NOT EXISTS performance_metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value REAL NOT NULL,
			timestamp INTEGER NOT NULL,
			url TEXT,
			user_id TEXT,
			created_at DATETIME NOT NULL
		)
	`)

	h.db.Exec(query, metric.Name, metric.Value, metric.Timestamp, metric.URL, userID)
}

// processSummaryMetrics processes aggregated metric summaries
func (h *MonitoringHandler) processSummaryMetrics(summary map[string]MetricSummary, userID string) {
	// Log summary for important metrics
	for name, data := range summary {
		if h.isImportantMetric(name) {
			logger.Info("Metric summary", map[string]interface{}{
				"name":     name,
				"avg":      data.Avg,
				"max":      data.Max,
				"min":      data.Min,
				"count":    data.Count,
				"user_id":  userID,
			})
		}
	}
}

// storeSecurityViolation stores security violation in database
func (h *MonitoringHandler) storeSecurityViolation(violation SecurityViolation, userID string) {
	query := `
		INSERT OR IGNORE INTO security_violations (type, details, timestamp, user_agent, url, user_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
	`
	
	// Create security_violations table if it doesn't exist
	h.db.Exec(`
		CREATE TABLE IF NOT EXISTS security_violations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			details TEXT,
			timestamp INTEGER NOT NULL,
			user_agent TEXT,
			url TEXT,
			user_id TEXT,
			created_at DATETIME NOT NULL
		)
	`)

	detailsJSON, _ := json.Marshal(violation.Details)
	h.db.Exec(query, violation.Type, string(detailsJSON), violation.Timestamp,
		violation.UserAgent, violation.URL, userID)
}

// writeJSON writes a JSON response
func (h *MonitoringHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *MonitoringHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}

	json.NewEncoder(w).Encode(response)
}