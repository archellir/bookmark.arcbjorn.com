package handlers

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"strings"
	"time"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

type ImportHandler struct {
	repo *db.BookmarkRepository
}

func NewImportHandler(repo *db.BookmarkRepository) *ImportHandler {
	return &ImportHandler{repo: repo}
}

// ChromeBookmark represents Chrome/Chromium bookmark format
type ChromeBookmark struct {
	DateAdded    string             `json:"date_added"`
	DateModified string             `json:"date_modified"`
	GUID         string             `json:"guid"`
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	URL          string             `json:"url"`
	Children     []*ChromeBookmark  `json:"children"`
}

// NetscapeBookmark represents Netscape bookmark format (used by Firefox, Safari, etc.)
type NetscapeBookmark struct {
	Title       string
	URL         string
	Description string
	Tags        []string
	AddDate     int64
	IsFolder    bool
	Children    []*NetscapeBookmark
}

// ImportRequest represents the import request payload
type ImportRequest struct {
	Format string `json:"format"` // "chrome", "firefox", "safari", "json"
	Data   string `json:"data"`   // Base64 encoded file content or JSON string
}

// ImportResponse represents the import response
type ImportResponse struct {
	Success   bool   `json:"success"`
	Imported  int    `json:"imported"`
	Skipped   int    `json:"skipped"`
	Errors    int    `json:"errors"`
	Message   string `json:"message"`
}

func (h *ImportHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/bookmarks/import", h.importBookmarks)
}

func (h *ImportHandler) importBookmarks(w http.ResponseWriter, r *http.Request) {
	var req ImportRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var bookmarks []*models.Bookmark
	var err error

	switch strings.ToLower(req.Format) {
	case "chrome", "chromium", "edge":
		bookmarks, err = h.parseChrome(req.Data)
	case "firefox", "safari", "netscape", "html":
		bookmarks, err = h.parseNetscape(req.Data)
	case "json":
		bookmarks, err = h.parseJSON(req.Data)
	default:
		http.Error(w, "Unsupported format", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse bookmarks: %v", err), http.StatusBadRequest)
		return
	}

	// Import bookmarks
	imported, skipped, errors := h.importBookmarksBatch(bookmarks)

	response := ImportResponse{
		Success:  errors == 0,
		Imported: imported,
		Skipped:  skipped,
		Errors:   errors,
		Message:  fmt.Sprintf("Imported %d bookmarks, skipped %d duplicates, %d errors", imported, skipped, errors),
	}

	w.Header().Set("Content-Type", "application/json")
	json.MarshalWrite(w, response)
}

func (h *ImportHandler) parseChrome(data string) ([]*models.Bookmark, error) {
	var chromeData struct {
		Roots struct {
			BookmarkBar ChromeBookmark `json:"bookmark_bar"`
			Other       ChromeBookmark `json:"other"`
			Synced      ChromeBookmark `json:"synced"`
		} `json:"roots"`
	}

	if err := json.Unmarshal([]byte(data), &chromeData); err != nil {
		return nil, fmt.Errorf("invalid Chrome bookmarks format: %w", err)
	}

	var bookmarks []*models.Bookmark

	// Extract bookmarks from all sections
	h.extractChromeBookmarks(&chromeData.Roots.BookmarkBar, &bookmarks, "Bookmark Bar")
	h.extractChromeBookmarks(&chromeData.Roots.Other, &bookmarks, "Other Bookmarks")
	h.extractChromeBookmarks(&chromeData.Roots.Synced, &bookmarks, "Mobile Bookmarks")

	return bookmarks, nil
}

func (h *ImportHandler) extractChromeBookmarks(node *ChromeBookmark, bookmarks *[]*models.Bookmark, folderPath string) {
	if node.Type == "folder" {
		newPath := folderPath
		if node.Name != "" {
			newPath = folderPath + "/" + node.Name
		}
		
		for _, child := range node.Children {
			h.extractChromeBookmarks(child, bookmarks, newPath)
		}
	} else if node.Type == "url" && node.URL != "" {
		// Convert Chrome timestamp (microseconds since Windows epoch) to Unix timestamp
		var addedTime time.Time
		if node.DateAdded != "" {
			// Chrome timestamps are microseconds since January 1, 1601
			// Convert to Unix timestamp
			if timestamp := h.parseTime(node.DateAdded); timestamp != 0 {
				addedTime = time.Unix(timestamp, 0)
			} else {
				addedTime = time.Now()
			}
		} else {
			addedTime = time.Now()
		}

		tags := []string{}
		if folderPath != "Bookmark Bar" && folderPath != "Other Bookmarks" && folderPath != "Mobile Bookmarks" {
			// Use folder path as tags
			parts := strings.Split(folderPath, "/")
			for _, part := range parts {
				if part != "" && part != "Bookmark Bar" && part != "Other Bookmarks" && part != "Mobile Bookmarks" {
					tags = append(tags, part)
				}
			}
		}

		bookmark := &models.Bookmark{
			Title:       node.Name,
			URL:         node.URL,
			Description: nil, // Chrome doesn't have descriptions
			CreatedAt:   addedTime,
			UpdatedAt:   addedTime,
			IsFavorite:  false,
			TagNames:    tags, // Use TagNames instead of Tags
		}

		*bookmarks = append(*bookmarks, bookmark)
	}
}

func (h *ImportHandler) parseNetscape(data string) ([]*models.Bookmark, error) {
	// This is a simplified Netscape bookmark parser
	// In a real implementation, you'd want a proper HTML parser
	lines := strings.Split(data, "\n")
	var bookmarks []*models.Bookmark
	var currentTags []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.Contains(line, "<DT><H3") {
			// Folder start
			if start := strings.Index(line, ">"); start != -1 {
				if end := strings.Index(line[start:], "</H3>"); end != -1 {
					folderName := strings.TrimSpace(line[start+1 : start+end])
					if folderName != "" {
						currentTags = append(currentTags, folderName)
					}
				}
			}
		} else if strings.Contains(line, "</DL>") {
			// Folder end
			if len(currentTags) > 0 {
				currentTags = currentTags[:len(currentTags)-1]
			}
		} else if strings.Contains(line, "<DT><A HREF=") {
			// Bookmark
			bookmark := h.parseNetscapeBookmark(line, currentTags)
			if bookmark != nil {
				bookmarks = append(bookmarks, bookmark)
			}
		}
	}

	return bookmarks, nil
}

func (h *ImportHandler) parseNetscapeBookmark(line string, tags []string) *models.Bookmark {
	// Extract URL
	hrefStart := strings.Index(line, `HREF="`)
	if hrefStart == -1 {
		return nil
	}
	hrefStart += 6
	
	hrefEnd := strings.Index(line[hrefStart:], `"`)
	if hrefEnd == -1 {
		return nil
	}
	
	url := line[hrefStart : hrefStart+hrefEnd]
	
	// Extract title
	titleStart := strings.Index(line, ">")
	if titleStart == -1 {
		return nil
	}
	titleStart += 1
	
	titleEnd := strings.Index(line[titleStart:], "</A>")
	if titleEnd == -1 {
		return nil
	}
	
	title := strings.TrimSpace(line[titleStart : titleStart+titleEnd])
	
	// Extract add date if present
	var addedTime time.Time
	if addDateStart := strings.Index(line, `ADD_DATE="`); addDateStart != -1 {
		addDateStart += 10
		if addDateEnd := strings.Index(line[addDateStart:], `"`); addDateEnd != -1 {
			if timestamp := h.parseTime(line[addDateStart : addDateStart+addDateEnd]); timestamp != 0 {
				addedTime = time.Unix(timestamp, 0)
			} else {
				addedTime = time.Now()
			}
		}
	} else {
		addedTime = time.Now()
	}

	return &models.Bookmark{
		Title:       title,
		URL:         url,
		Description: nil,
		CreatedAt:   addedTime,
		UpdatedAt:   addedTime,
		IsFavorite:  false,
		TagNames:    append([]string{}, tags...), // Copy tags
	}
}

func (h *ImportHandler) parseJSON(data string) ([]*models.Bookmark, error) {
	var bookmarks []*models.Bookmark
	if err := json.Unmarshal([]byte(data), &bookmarks); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}
	
	// Set timestamps if missing
	now := time.Now()
	for _, bookmark := range bookmarks {
		if bookmark.CreatedAt.IsZero() {
			bookmark.CreatedAt = now
		}
		if bookmark.UpdatedAt.IsZero() {
			bookmark.UpdatedAt = now
		}
	}
	
	return bookmarks, nil
}

func (h *ImportHandler) parseTime(timeStr string) int64 {
	// Try parsing as Unix timestamp first
	if timestamp, err := time.Parse("1136239445", timeStr); err == nil {
		return timestamp.Unix()
	}
	
	// Try parsing Chrome timestamp (microseconds since Windows epoch)
	if len(timeStr) > 10 {
		// Chrome uses microseconds since January 1, 1601
		// Convert to Unix timestamp (seconds since January 1, 1970)
		if chromeTime, err := time.Parse("1136239445000000", timeStr); err == nil {
			// Windows epoch is January 1, 1601, Unix epoch is January 1, 1970
			// Difference is 11644473600 seconds
			return chromeTime.Unix() - 11644473600
		}
	}
	
	return 0
}

func (h *ImportHandler) importBookmarksBatch(bookmarks []*models.Bookmark) (imported, skipped, errors int) {
	userID := 1 // Default userID for import operations
	for _, bookmark := range bookmarks {
		// Check for duplicates by URL
		existing, err := h.repo.GetByURL(bookmark.URL, userID)
		if err == nil && existing != nil {
			skipped++
			continue
		}

		// Validate required fields
		if bookmark.Title == "" || bookmark.URL == "" {
			errors++
			continue
		}

		// Convert to CreateBookmarkRequest
		req := &models.CreateBookmarkRequest{
			Title:       bookmark.Title,
			URL:         bookmark.URL,
			Description: bookmark.Description,
			Tags:        bookmark.TagNames, // Use the temporary TagNames field
		}

		// Create bookmark
		if _, err := h.repo.Create(req, userID); err != nil {
			errors++
			continue
		}

		imported++
	}

	return imported, skipped, errors
}