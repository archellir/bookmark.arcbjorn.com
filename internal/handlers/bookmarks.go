package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

// BookmarkHandler handles bookmark-related HTTP requests
type BookmarkHandler struct {
	repo *db.BookmarkRepository
}

// NewBookmarkHandler creates a new bookmark handler
func NewBookmarkHandler(repo *db.BookmarkRepository) *BookmarkHandler {
	return &BookmarkHandler{repo: repo}
}

// ServeHTTP implements the http.Handler interface
func (h *BookmarkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/bookmarks")
	
	switch {
	case r.Method == "GET" && path == "":
		h.listBookmarks(w, r)
	case r.Method == "POST" && path == "":
		h.createBookmark(w, r)
	case r.Method == "GET" && strings.HasPrefix(path, "/"):
		h.getBookmark(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == "PUT" && strings.HasPrefix(path, "/"):
		h.updateBookmark(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == "DELETE" && strings.HasPrefix(path, "/"):
		h.deleteBookmark(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == "GET" && path == "/search":
		h.searchBookmarks(w, r)
	default:
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listBookmarks handles GET /api/bookmarks
func (h *BookmarkHandler) listBookmarks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	
	// Parse pagination parameters
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(query.Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse filters
	searchQuery := query.Get("search")
	tagFilter := query.Get("tag")
	favoritesOnly := query.Get("favorites") == "true"

	// Get bookmarks
	response, err := h.repo.List(page, limit, searchQuery, tagFilter, favoritesOnly)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to list bookmarks: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, response)
}

// createBookmark handles POST /api/bookmarks
func (h *BookmarkHandler) createBookmark(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.URL == "" {
		h.writeError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Validate URL format
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		h.writeError(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	// Set title if not provided
	if req.Title == "" {
		req.Title = req.URL
	}

	// Create bookmark
	bookmark, err := h.repo.Create(&req)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.writeError(w, "Bookmark with this URL already exists", http.StatusConflict)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to create bookmark: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	h.writeJSON(w, bookmark)
}

// getBookmark handles GET /api/bookmarks/{id}
func (h *BookmarkHandler) getBookmark(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	bookmark, err := h.repo.GetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Bookmark not found", http.StatusNotFound)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to get bookmark: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, bookmark)
}

// updateBookmark handles PUT /api/bookmarks/{id}
func (h *BookmarkHandler) updateBookmark(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateBookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate URL if provided
	if req.URL != nil {
		if _, err := url.ParseRequestURI(*req.URL); err != nil {
			h.writeError(w, "Invalid URL format", http.StatusBadRequest)
			return
		}
	}

	bookmark, err := h.repo.Update(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Bookmark not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.writeError(w, "Bookmark with this URL already exists", http.StatusConflict)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to update bookmark: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, bookmark)
}

// deleteBookmark handles DELETE /api/bookmarks/{id}
func (h *BookmarkHandler) deleteBookmark(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	err = h.repo.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Bookmark not found", http.StatusNotFound)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to delete bookmark: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// searchBookmarks handles GET /api/bookmarks/search
func (h *BookmarkHandler) searchBookmarks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.writeError(w, "Search query is required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	results, err := h.repo.Search(query, limit)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
	}

	h.writeJSON(w, response)
}

// writeJSON writes a JSON response
func (h *BookmarkHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *BookmarkHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}
	
	json.NewEncoder(w).Encode(response)
}