package handlers

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

// FolderHandler handles folder-related HTTP requests
type FolderHandler struct {
	repo         *db.FolderRepository
	bookmarkRepo *db.BookmarkRepository
}

// NewFolderHandler creates a new folder handler
func NewFolderHandler(repo *db.FolderRepository, bookmarkRepo *db.BookmarkRepository) *FolderHandler {
	return &FolderHandler{
		repo:         repo,
		bookmarkRepo: bookmarkRepo,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *FolderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/folders")

	switch {
	case r.Method == "GET" && path == "":
		h.listFolders(w, r)
	case r.Method == "POST" && path == "":
		h.createFolder(w, r)
	case r.Method == "GET" && path == "/tree":
		h.getFolderTree(w, r)
	case r.Method == "GET" && strings.HasPrefix(path, "/"):
		h.getFolder(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == "PUT" && strings.HasPrefix(path, "/"):
		h.updateFolder(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == "DELETE" && strings.HasPrefix(path, "/"):
		h.deleteFolder(w, r, strings.TrimPrefix(path, "/"))
	default:
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// RegisterRoutes registers additional folder-related routes
func (h *FolderHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/folders/bookmark/", h.handleBookmarkFolders)
}

// listFolders handles GET /api/folders
func (h *FolderHandler) listFolders(w http.ResponseWriter, r *http.Request) {
	// Parse optional parent_id parameter
	var parentID *int
	if parentStr := r.URL.Query().Get("parent_id"); parentStr != "" {
		if pid, err := strconv.Atoi(parentStr); err == nil {
			parentID = &pid
		}
	}

	folders, err := h.repo.List(parentID)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to list folders: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"folders": folders,
		"count":   len(folders),
	}

	h.writeJSON(w, response)
}

// createFolder handles POST /api/folders
func (h *FolderHandler) createFolder(w http.ResponseWriter, r *http.Request) {
	var req models.CreateFolderRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		h.writeError(w, "Name is required", http.StatusBadRequest)
		return
	}

	folder, err := h.repo.Create(&req)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.writeError(w, "Folder with this name already exists in the parent folder", http.StatusConflict)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to create folder: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	h.writeJSON(w, folder)
}

// getFolderTree handles GET /api/folders/tree
func (h *FolderHandler) getFolderTree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.repo.GetTree()
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get folder tree: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tree": tree,
	}

	h.writeJSON(w, response)
}

// getFolder handles GET /api/folders/{id}
func (h *FolderHandler) getFolder(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	folder, err := h.repo.GetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Folder not found", http.StatusNotFound)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to get folder: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, folder)
}

// updateFolder handles PUT /api/folders/{id}
func (h *FolderHandler) updateFolder(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateFolderRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	folder, err := h.repo.Update(id, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Folder not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.writeError(w, "Folder with this name already exists in the parent folder", http.StatusConflict)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to update folder: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, folder)
}

// deleteFolder handles DELETE /api/folders/{id}
func (h *FolderHandler) deleteFolder(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.writeError(w, "Invalid folder ID", http.StatusBadRequest)
		return
	}

	err = h.repo.Delete(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, "Folder not found", http.StatusNotFound)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to delete folder: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleBookmarkFolders handles bookmark-folder relationships
func (h *FolderHandler) handleBookmarkFolders(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/folders/bookmark/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 {
		h.writeError(w, "Invalid path", http.StatusBadRequest)
		return
	}

	bookmarkID, err := strconv.Atoi(parts[0])
	if err != nil {
		h.writeError(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		// Get all folders for a bookmark
		folders, err := h.repo.GetBookmarkFolders(bookmarkID)
		if err != nil {
			h.writeError(w, fmt.Sprintf("Failed to get bookmark folders: %v", err), http.StatusInternalServerError)
			return
		}
		h.writeJSON(w, map[string]interface{}{
			"folders": folders,
			"count":   len(folders),
		})

	case "POST":
		// Add bookmark to folder
		if len(parts) < 2 {
			h.writeError(w, "Folder ID required", http.StatusBadRequest)
			return
		}

		folderID, err := strconv.Atoi(parts[1])
		if err != nil {
			h.writeError(w, "Invalid folder ID", http.StatusBadRequest)
			return
		}

		err = h.repo.AddBookmarkToFolder(bookmarkID, folderID)
		if err != nil {
			h.writeError(w, fmt.Sprintf("Failed to add bookmark to folder: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		h.writeJSON(w, map[string]string{"status": "success"})

	case "DELETE":
		// Remove bookmark from folder
		if len(parts) < 2 {
			h.writeError(w, "Folder ID required", http.StatusBadRequest)
			return
		}

		folderID, err := strconv.Atoi(parts[1])
		if err != nil {
			h.writeError(w, "Invalid folder ID", http.StatusBadRequest)
			return
		}

		err = h.repo.RemoveBookmarkFromFolder(bookmarkID, folderID)
		if err != nil {
			h.writeError(w, fmt.Sprintf("Failed to remove bookmark from folder: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// writeJSON writes a JSON response
func (h *FolderHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.MarshalWrite(w, data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *FolderHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}

	json.MarshalWrite(w, response)
}