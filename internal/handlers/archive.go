package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"torimemo/internal/db"
)

// ArchiveHandler handles content archiving requests
type ArchiveHandler struct {
	repo *db.BookmarkRepository
}

// NewArchiveHandler creates a new archive handler
func NewArchiveHandler(repo *db.BookmarkRepository) *ArchiveHandler {
	return &ArchiveHandler{
		repo: repo,
	}
}

// RegisterRoutes registers archive-related routes
func (h *ArchiveHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/archive/content", h.archiveContent)
}

// ArchiveRequest represents the request payload for archiving content
type ArchiveRequest struct {
	URL        string `json:"url"`
	BookmarkID int    `json:"bookmark_id"`
}

// ArchiveResponse represents the response with archived content
type ArchiveResponse struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	TextContent string `json:"text_content"`
	Screenshot  string `json:"screenshot,omitempty"`
	CachedAt    int64  `json:"cached_at"`
	Size        int    `json:"size"`
}

// archiveContent handles POST /api/archive/content
func (h *ArchiveHandler) archiveContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.writeError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Validate URL format
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		h.writeError(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	// Fetch content from the URL
	content, err := h.fetchWebContent(req.URL)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to fetch content: %v", err), http.StatusInternalServerError)
		return
	}

	// Extract text content (simple HTML stripping)
	textContent := h.extractTextContent(content)

	response := ArchiveResponse{
		Title:       h.extractTitle(content, parsedURL.Host),
		Content:     content,
		TextContent: textContent,
		CachedAt:    time.Now().Unix(),
		Size:        len(content),
	}

	h.writeJSON(w, response)
}

// fetchWebContent fetches the HTML content from a URL
func (h *ArchiveHandler) fetchWebContent(urlStr string) (string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request with user agent
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Torimemo/1.0; +https://torimemo.app)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read response body with size limit (10MB max)
	const maxSize = 10 * 1024 * 1024
	limitedReader := io.LimitReader(resp.Body, maxSize)
	
	content, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

// extractTitle attempts to extract the page title from HTML content
func (h *ArchiveHandler) extractTitle(content, fallback string) string {
	// Simple regex-like approach to find title tag
	lowerContent := strings.ToLower(content)
	
	start := strings.Index(lowerContent, "<title")
	if start == -1 {
		return fallback
	}

	// Find the closing > of the opening tag
	openEnd := strings.Index(lowerContent[start:], ">")
	if openEnd == -1 {
		return fallback
	}

	titleStart := start + openEnd + 1

	// Find the closing </title>
	end := strings.Index(lowerContent[titleStart:], "</title>")
	if end == -1 {
		return fallback
	}

	title := strings.TrimSpace(content[titleStart : titleStart+end])
	if title == "" {
		return fallback
	}

	// Basic HTML entity decoding for common entities
	title = strings.ReplaceAll(title, "&amp;", "&")
	title = strings.ReplaceAll(title, "&lt;", "<")
	title = strings.ReplaceAll(title, "&gt;", ">")
	title = strings.ReplaceAll(title, "&quot;", "\"")
	title = strings.ReplaceAll(title, "&#39;", "'")
	title = strings.ReplaceAll(title, "&apos;", "'")

	return title
}

// extractTextContent removes HTML tags and extracts plain text
func (h *ArchiveHandler) extractTextContent(content string) string {
	// Simple approach: remove script and style tags completely, then strip HTML
	result := content
	
	// Remove script tags and their content
	for {
		start := strings.Index(strings.ToLower(result), "<script")
		if start == -1 {
			break
		}
		end := strings.Index(strings.ToLower(result[start:]), "</script>")
		if end == -1 {
			result = result[:start]
			break
		}
		result = result[:start] + result[start+end+9:]
	}

	// Remove style tags and their content
	for {
		start := strings.Index(strings.ToLower(result), "<style")
		if start == -1 {
			break
		}
		end := strings.Index(strings.ToLower(result[start:]), "</style>")
		if end == -1 {
			result = result[:start]
			break
		}
		result = result[:start] + result[start+end+8:]
	}

	// Simple HTML tag removal
	var text strings.Builder
	inTag := false
	
	for _, char := range result {
		switch char {
		case '<':
			inTag = true
		case '>':
			inTag = false
			text.WriteRune(' ') // Replace tags with spaces
		default:
			if !inTag {
				text.WriteRune(char)
			}
		}
	}

	// Clean up whitespace
	textContent := strings.TrimSpace(text.String())
	
	// Replace multiple whitespace with single spaces
	words := strings.Fields(textContent)
	return strings.Join(words, " ")
}

// writeJSON writes a JSON response
func (h *ArchiveHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *ArchiveHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}
	
	json.NewEncoder(w).Encode(response)
}