package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"torimemo/internal/db"
	"torimemo/internal/services"
)

// DuplicateHandler handles duplicate detection requests
type DuplicateHandler struct {
	duplicateService *services.DuplicateService
}

// NewDuplicateHandler creates a new duplicate handler
func NewDuplicateHandler(bookmarkRepo *db.BookmarkRepository) *DuplicateHandler {
	return &DuplicateHandler{
		duplicateService: services.NewDuplicateService(bookmarkRepo),
	}
}

// RegisterRoutes registers duplicate-related routes
func (h *DuplicateHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/duplicates/check", h.checkDuplicates)
	mux.HandleFunc("/api/duplicates/find-all", h.findAllDuplicates)
	mux.HandleFunc("/api/duplicates/merge", h.mergeDuplicates)
	mux.HandleFunc("/api/url/analyze", h.analyzeURL)
	mux.HandleFunc("/api/url/expand", h.expandURL)
	mux.HandleFunc("/api/url/shorten", h.shortenURL)
}

// CheckDuplicatesRequest represents a duplicate check request
type CheckDuplicatesRequest struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// MergeDuplicatesRequest represents a merge duplicates request
type MergeDuplicatesRequest struct {
	PrimaryID     int    `json:"primary_id"`
	DuplicateIDs  []int  `json:"duplicate_ids"`
	MergeTags     bool   `json:"merge_tags"`
	MergeMetadata bool   `json:"merge_metadata"`
}

// URLAnalyzeRequest represents a URL analysis request
type URLAnalyzeRequest struct {
	URL string `json:"url"`
}

// URLShortenRequest represents a URL shortener request
type URLShortenRequest struct {
	URL     string `json:"url"`
	BaseURL string `json:"base_url"`
}

// checkDuplicates handles POST /api/duplicates/check
func (h *DuplicateHandler) checkDuplicates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CheckDuplicatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.writeError(w, "URL is required", http.StatusBadRequest)
		return
	}

	result, err := h.duplicateService.CheckForDuplicates(req.URL, req.Title)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to check duplicates: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, result)
}

// findAllDuplicates handles GET /api/duplicates/find-all
func (h *DuplicateHandler) findAllDuplicates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	duplicateGroups, err := h.duplicateService.FindAllDuplicates()
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to find duplicates: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"duplicate_groups": duplicateGroups,
		"count":           len(duplicateGroups),
	}

	h.writeJSON(w, response)
}

// mergeDuplicates handles POST /api/duplicates/merge
func (h *DuplicateHandler) mergeDuplicates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MergeDuplicatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.PrimaryID == 0 || len(req.DuplicateIDs) == 0 {
		h.writeError(w, "Primary ID and duplicate IDs are required", http.StatusBadRequest)
		return
	}

	err := h.duplicateService.MergeDuplicates(req.PrimaryID, req.DuplicateIDs, req.MergeTags, req.MergeMetadata)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to merge duplicates: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully merged %d duplicates", len(req.DuplicateIDs)),
	}

	h.writeJSON(w, response)
}

// analyzeURL handles POST /api/url/analyze
func (h *DuplicateHandler) analyzeURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req URLAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.writeError(w, "URL is required", http.StatusBadRequest)
		return
	}

	urlService := services.NewURLService()
	result, err := urlService.NormalizeURL(req.URL)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to analyze URL: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, result)
}

// expandURL handles POST /api/url/expand
func (h *DuplicateHandler) expandURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req URLAnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.writeError(w, "URL is required", http.StatusBadRequest)
		return
	}

	urlService := services.NewURLService()
	result, err := urlService.NormalizeURL(req.URL)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to analyze URL: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"original_url": req.URL,
		"expanded_url": result.ExpandedURL,
		"is_short_url": result.IsShortURL,
		"normalized":   result.Normalized,
	}

	h.writeJSON(w, response)
}

// shortenURL handles POST /api/url/shorten
func (h *DuplicateHandler) shortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req URLShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.writeError(w, "URL is required", http.StatusBadRequest)
		return
	}

	if req.BaseURL == "" {
		// Use request host as base URL
		scheme := "https"
		if r.TLS == nil {
			scheme = "http"
		}
		req.BaseURL = fmt.Sprintf("%s://%s", scheme, r.Host)
	}

	urlService := services.NewURLService()
	shortURL, err := urlService.GenerateShortURL(req.URL, req.BaseURL)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to generate short URL: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"original_url": req.URL,
		"short_url":    shortURL,
		"base_url":     req.BaseURL,
	}

	h.writeJSON(w, response)
}

// writeJSON writes a JSON response
func (h *DuplicateHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *DuplicateHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}

	json.NewEncoder(w).Encode(response)
}