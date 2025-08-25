package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

// ImportExportHandler handles bookmark import/export operations
type ImportExportHandler struct {
	bookmarkRepo *db.BookmarkRepository
	tagRepo      *db.TagRepository
}

// NewImportExportHandler creates a new import/export handler
func NewImportExportHandler(bookmarkRepo *db.BookmarkRepository, tagRepo *db.TagRepository) *ImportExportHandler {
	return &ImportExportHandler{
		bookmarkRepo: bookmarkRepo,
		tagRepo:      tagRepo,
	}
}

// ExportData represents the complete export format
type ExportData struct {
	Version   string             `json:"version"`
	ExportedAt string            `json:"exported_at"`
	Bookmarks []models.Bookmark  `json:"bookmarks"`
	Tags      []models.Tag       `json:"tags"`
}

// ServeHTTP implements the http.Handler interface
func (h *ImportExportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		h.exportBookmarks(w, r)
	case "POST":
		h.importBookmarks(w, r)
	default:
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// exportBookmarks handles GET /api/export
func (h *ImportExportHandler) exportBookmarks(w http.ResponseWriter, r *http.Request) {
	// Get all bookmarks
	bookmarks, err := h.bookmarkRepo.List(1, 10000, "", "", false) // Large limit for export
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to fetch bookmarks: %v", err), http.StatusInternalServerError)
		return
	}

	// Get all tags
	tagList, err := h.tagRepo.List("")
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to fetch tags: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert TagWithCount to Tag
	tags := make([]models.Tag, len(tagList))
	for i, tag := range tagList {
		tags[i] = models.Tag{
			ID:        tag.ID,
			Name:      tag.Name,
			Color:     tag.Color,
			CreatedAt: tag.CreatedAt,
		}
	}

	// Create export data
	exportData := ExportData{
		Version:    "1.0",
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Bookmarks:  bookmarks.Bookmarks,
		Tags:       tags,
	}

	// Set content disposition for download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=torimemo-export-%s.json", 
		time.Now().Format("2006-01-02")))

	// Encode and send
	if err := json.NewEncoder(w).Encode(exportData); err != nil {
		h.writeError(w, fmt.Sprintf("Failed to encode export: %v", err), http.StatusInternalServerError)
		return
	}
}

// importBookmarks handles POST /api/import
func (h *ImportExportHandler) importBookmarks(w http.ResponseWriter, r *http.Request) {
	var importData ExportData
	
	if err := json.NewDecoder(r.Body).Decode(&importData); err != nil {
		h.writeError(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	imported := 0
	skipped := 0
	
	// Import bookmarks
	for _, bookmark := range importData.Bookmarks {
		// Check if bookmark already exists by URL
		existing, err := h.bookmarkRepo.GetByURL(bookmark.URL)
		if err == nil && existing != nil {
			skipped++
			continue
		}

		// Create new bookmark (tags will be created automatically if they don't exist)
		tagNames := make([]string, len(bookmark.Tags))
		for i, tag := range bookmark.Tags {
			tagNames[i] = tag.Name
		}

		createReq := models.CreateBookmarkRequest{
			Title:       bookmark.Title,
			URL:         bookmark.URL,
			Description: bookmark.Description,
			Tags:        tagNames,
		}

		_, err = h.bookmarkRepo.Create(&createReq)
		if err != nil {
			// Log error but continue with other bookmarks
			continue
		}
		
		imported++
	}

	// Return import summary
	response := map[string]interface{}{
		"imported": imported,
		"skipped":  skipped,
		"total":    len(importData.Bookmarks),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ImportExportHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  message,
		"status": statusCode,
	})
}