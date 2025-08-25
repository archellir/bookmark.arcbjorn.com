package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

// TagHandler handles tag-related HTTP requests
type TagHandler struct {
	repo *db.TagRepository
}

// NewTagHandler creates a new tag handler
func NewTagHandler(repo *db.TagRepository) *TagHandler {
	return &TagHandler{repo: repo}
}

// ServeHTTP implements the http.Handler interface
func (h *TagHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/tags")
	
	switch {
	case r.Method == "GET" && path == "":
		h.listTags(w, r)
	case r.Method == "POST" && path == "":
		h.createTag(w, r)
	case r.Method == "GET" && strings.HasPrefix(path, "/"):
		h.getTag(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == "PUT" && strings.HasPrefix(path, "/"):
		h.updateTag(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == "DELETE" && strings.HasPrefix(path, "/"):
		h.deleteTag(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == "GET" && path == "/cloud":
		h.getTagCloud(w, r)
	case r.Method == "GET" && path == "/popular":
		h.getPopularTags(w, r)
	case r.Method == "DELETE" && path == "/cleanup":
		h.cleanupUnusedTags(w, r)
	default:
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// listTags handles GET /api/tags
func (h *TagHandler) listTags(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	
	tags, err := h.repo.List(search)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to list tags: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tags":  tags,
		"count": len(tags),
	}

	h.writeJSON(w, response)
}

// createTag handles POST /api/tags
func (h *TagHandler) createTag(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		h.writeError(w, "Tag name is required", http.StatusBadRequest)
		return
	}

	// Trim and validate name
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) < 1 || len(req.Name) > 50 {
		h.writeError(w, "Tag name must be between 1 and 50 characters", http.StatusBadRequest)
		return
	}

	// Create tag
	tag, err := h.repo.Create(&req)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.writeError(w, "Tag with this name already exists", http.StatusConflict)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to create tag: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	h.writeJSON(w, tag)
}

// getTag handles GET /api/tags/{id}
func (h *TagHandler) getTag(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	tag, err := h.repo.GetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Tag not found", http.StatusNotFound)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to get tag: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, tag)
}

// updateTag handles PUT /api/tags/{id}
func (h *TagHandler) updateTag(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateTagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate name if provided
	if req.Name != nil {
		*req.Name = strings.TrimSpace(*req.Name)
		if len(*req.Name) < 1 || len(*req.Name) > 50 {
			h.writeError(w, "Tag name must be between 1 and 50 characters", http.StatusBadRequest)
			return
		}
	}

	tag, err := h.repo.Update(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Tag not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.writeError(w, "Tag with this name already exists", http.StatusConflict)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to update tag: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, tag)
}

// deleteTag handles DELETE /api/tags/{id}
func (h *TagHandler) deleteTag(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid tag ID", http.StatusBadRequest)
		return
	}

	err = h.repo.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Tag not found", http.StatusNotFound)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to delete tag: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getTagCloud handles GET /api/tags/cloud
func (h *TagHandler) getTagCloud(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	cloud, err := h.repo.GetTagCloud(limit)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get tag cloud: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tags":  cloud,
		"count": len(cloud),
	}

	h.writeJSON(w, response)
}

// getPopularTags handles GET /api/tags/popular
func (h *TagHandler) getPopularTags(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	tags, err := h.repo.GetPopularTags(limit)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get popular tags: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tags":  tags,
		"count": len(tags),
	}

	h.writeJSON(w, response)
}

// cleanupUnusedTags handles DELETE /api/tags/cleanup
func (h *TagHandler) cleanupUnusedTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.repo.CleanupUnusedTags()
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to cleanup tags: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message":       "Unused tags cleaned up successfully",
		"deleted_count": count,
	}

	h.writeJSON(w, response)
}

// writeJSON writes a JSON response
func (h *TagHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *TagHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}
	
	json.NewEncoder(w).Encode(response)
}