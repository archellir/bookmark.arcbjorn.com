package handlers

import (
	"encoding/json/v2"
	"net/http"
	"strconv"
	"strings"

	"torimemo/internal/services"
)

type HealthHandler struct {
	healthChecker *services.HealthChecker
}

func NewHealthHandler(healthChecker *services.HealthChecker) *HealthHandler {
	return &HealthHandler{
		healthChecker: healthChecker,
	}
}

func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health/stats", h.getHealthStats)
	mux.HandleFunc("GET /api/health/broken", h.getBrokenBookmarks)
	mux.HandleFunc("GET /api/health/all", h.getAllHealth)
	mux.HandleFunc("GET /api/health/bookmark/", h.getBookmarkHealth)
	mux.HandleFunc("POST /api/health/check/", h.checkBookmarkNow)
}

// getHealthStats returns overall health statistics
func (h *HealthHandler) getHealthStats(w http.ResponseWriter, r *http.Request) {
	stats := h.healthChecker.GetHealthStats()
	
	w.Header().Set("Content-Type", "application/json")
	json.MarshalWrite(w, stats)
}

// getBrokenBookmarks returns all bookmarks with broken links
func (h *HealthHandler) getBrokenBookmarks(w http.ResponseWriter, r *http.Request) {
	broken := h.healthChecker.GetBrokenBookmarks()
	
	w.Header().Set("Content-Type", "application/json")
	json.MarshalWrite(w, map[string]interface{}{
		"broken_bookmarks": broken,
		"count":           len(broken),
	})
}

// getAllHealth returns health data for all bookmarks
func (h *HealthHandler) getAllHealth(w http.ResponseWriter, r *http.Request) {
	health := h.healthChecker.GetAllHealth()
	
	w.Header().Set("Content-Type", "application/json")
	json.MarshalWrite(w, health)
}

// getBookmarkHealth returns health data for a specific bookmark
func (h *HealthHandler) getBookmarkHealth(w http.ResponseWriter, r *http.Request) {
	// Extract bookmark ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/health/bookmark/")
	bookmarkID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}
	
	health := h.healthChecker.GetHealth(bookmarkID)
	if health == nil {
		http.Error(w, "Health data not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.MarshalWrite(w, health)
}

// checkBookmarkNow triggers an immediate health check for a specific bookmark
func (h *HealthHandler) checkBookmarkNow(w http.ResponseWriter, r *http.Request) {
	// Extract bookmark ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/health/check/")
	bookmarkID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}
	
	health := h.healthChecker.CheckBookmarkNow(bookmarkID)
	
	w.Header().Set("Content-Type", "application/json")
	json.MarshalWrite(w, health)
}