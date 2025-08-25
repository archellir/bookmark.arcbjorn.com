package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
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
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	// Get all bookmarks
	bookmarks, err := h.bookmarkRepo.List(1, 10000, "", "", false, 1) // Large limit for export
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to fetch bookmarks: %v", err), http.StatusInternalServerError)
		return
	}

	if format == "html" {
		h.exportHTML(w, r, bookmarks.Bookmarks)
		return
	}

	// JSON export (default)
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

	// Set headers
	w.Header().Set("Content-Type", "application/json")
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
		existing, err := h.bookmarkRepo.GetByURL(bookmark.URL, 1)
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

		_, err = h.bookmarkRepo.Create(&createReq, 1)
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

// exportHTML exports bookmarks as HTML (browser-compatible format)
func (h *ImportExportHandler) exportHTML(w http.ResponseWriter, r *http.Request, bookmarks []models.Bookmark) {
	// Set headers
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=torimemo-export-%s.html", 
		time.Now().Format("2006-01-02")))

	// Write HTML header
	fmt.Fprintf(w, `<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>Torimemo Bookmarks</TITLE>
<H1>Torimemo Bookmarks</H1>

<DL><p>
`)

	// Group bookmarks by their first tag (or "Untagged")
	tagGroups := make(map[string][]models.Bookmark)
	for _, bookmark := range bookmarks {
		tagName := "Untagged"
		if len(bookmark.Tags) > 0 {
			tagName = bookmark.Tags[0].Name
		}
		tagGroups[tagName] = append(tagGroups[tagName], bookmark)
	}

	// Write bookmark folders
	for tagName, tagBookmarks := range tagGroups {
		fmt.Fprintf(w, "    <DT><H3 ADD_DATE=\"%d\" LAST_MODIFIED=\"%d\">%s</H3>\n", 
			time.Now().Unix(), time.Now().Unix(), html.EscapeString(tagName))
		fmt.Fprintf(w, "    <DL><p>\n")
		
		for _, bookmark := range tagBookmarks {
			addDate := bookmark.CreatedAt.Unix()
			description := ""
			if bookmark.Description != nil && *bookmark.Description != "" {
				description = *bookmark.Description
			}
			
			// Create tag list for description
			var tagList []string
			for _, tag := range bookmark.Tags {
				tagList = append(tagList, tag.Name)
			}
			if len(tagList) > 0 {
				if description != "" {
					description += " | "
				}
				description += "Tags: " + strings.Join(tagList, ", ")
			}
			
			fmt.Fprintf(w, "        <DT><A HREF=\"%s\" ADD_DATE=\"%d\">%s</A>\n", 
				html.EscapeString(bookmark.URL), addDate, html.EscapeString(bookmark.Title))
			
			if description != "" {
				fmt.Fprintf(w, "        <DD>%s\n", html.EscapeString(description))
			}
		}
		
		fmt.Fprintf(w, "    </DL><p>\n")
	}

	// Write HTML footer
	fmt.Fprintf(w, "</DL><p>\n")
}

func (h *ImportExportHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  message,
		"status": statusCode,
	})
}