package handlers

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"torimemo/internal/ai"
	"torimemo/internal/db"
	"torimemo/internal/middleware"
	"torimemo/internal/models"
	"torimemo/internal/search"
	"torimemo/internal/services"
)

// BookmarkHandler handles bookmark-related HTTP requests
type BookmarkHandler struct {
	repo             *db.BookmarkRepository
	learningRepo     *db.LearningRepository
	categorizer      *ai.Categorizer
	fuzzyMatcher     *search.FuzzyMatcher
	duplicateService *services.DuplicateService
}

// NewBookmarkHandler creates a new bookmark handler
func NewBookmarkHandler(repo *db.BookmarkRepository, learningRepo *db.LearningRepository) *BookmarkHandler {
	return &BookmarkHandler{
		repo:             repo,
		learningRepo:     learningRepo,
		categorizer:      ai.NewCategorizerWithLearning(learningRepo),
		fuzzyMatcher:     search.DefaultFuzzyMatcher(),
		duplicateService: services.NewDuplicateService(repo),
	}
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
	case r.Method == "GET" && path == "/search":
		h.searchBookmarks(w, r)
	case r.Method == "GET" && path == "/fuzzy-search":
		h.fuzzySearchBookmarks(w, r)
	case r.Method == "POST" && path == "/suggest-tags":
		h.suggestTags(w, r)
	case r.Method == "GET" && strings.HasPrefix(path, "/"):
		h.getBookmark(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == "PUT" && strings.HasPrefix(path, "/"):
		h.updateBookmark(w, r, strings.TrimPrefix(path, "/"))
	case r.Method == "DELETE" && strings.HasPrefix(path, "/"):
		h.deleteBookmark(w, r, strings.TrimPrefix(path, "/"))
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

	// Get user ID from context
	userID := h.getUserID(r)

	// Get bookmarks
	response, err := h.repo.List(page, limit, searchQuery, tagFilter, favoritesOnly, userID)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to list bookmarks: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, response)
}

// createBookmark handles POST /api/bookmarks
func (h *BookmarkHandler) createBookmark(w http.ResponseWriter, r *http.Request) {
	var req models.CreateBookmarkRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
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

	// Use AI to enhance bookmark data (tags, title, description, favicon)
	tempBookmark := &models.Bookmark{
		Title:       req.Title,
		URL:         req.URL,
		Description: req.Description,
	}

	suggestions, err := h.categorizer.CategorizeBookmark(tempBookmark)
	var faviconURL string
	if err == nil {
		// Use AI suggested tags if none provided
		if len(req.Tags) == 0 && len(suggestions.Tags) > 0 {
			req.Tags = suggestions.Tags
		}
		
		// Use AI suggested title if none provided or is just the URL
		if (req.Title == "" || req.Title == req.URL) && suggestions.Title != "" {
			req.Title = suggestions.Title
		}
		
		// Use AI suggested description if none provided
		if req.Description == nil && suggestions.Description != "" {
			req.Description = &suggestions.Description
		}
		
		// Store favicon URL for later use
		if suggestions.FaviconURL != "" {
			faviconURL = suggestions.FaviconURL
		}
	}

	// Set favicon URL if fetched and not provided
	if faviconURL != "" && req.FaviconURL == nil {
		req.FaviconURL = &faviconURL
	}

	// Fallback: set title if still empty
	if req.Title == "" {
		req.Title = req.URL
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Create bookmark
	bookmark, err := h.repo.Create(&req, userID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			h.writeError(w, "Bookmark with this URL already exists", http.StatusConflict)
			return
		}
		h.writeError(w, fmt.Sprintf("Failed to create bookmark: %v", err), http.StatusInternalServerError)
		return
	}

	// Save AI suggestions to learning system (async, don't fail if it errors)
	if suggestions != nil {
		go func() {
			// Convert TagSuggestion to LearnedPattern for storage
			domain := ""
			if parsedURL, err := url.Parse(suggestions.URL); err == nil {
				domain = parsedURL.Hostname()
			}
			
			pattern := &models.LearnedPattern{
				URLPattern:     suggestions.URL,
				Domain:         domain,
				ConfirmedTags:  suggestions.Tags,
				Confidence:     suggestions.Confidence,
				SampleCount:    1,
			}
			h.learningRepo.SavePattern(pattern)
		}()
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

	// Get user ID from context
	userID := h.getUserID(r)

	bookmark, err := h.repo.GetByID(id, userID)
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
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
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

	// Get original bookmark for learning feedback
	var originalTags []string
	// Get user ID from context
	userID := h.getUserID(r)

	if len(req.Tags) > 0 {
		originalBookmark, err := h.repo.GetByID(id, userID)
		if err == nil && len(originalBookmark.Tags) > 0 {
			for _, tag := range originalBookmark.Tags {
				originalTags = append(originalTags, tag.Name)
			}
		}
	}

	bookmark, err := h.repo.Update(id, &req, userID)
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

	// Record learning feedback if tags were changed
	if len(originalTags) > 0 && len(req.Tags) > 0 {
		go func() {
			correction := &models.TagCorrection{
				BookmarkID:   id,
				OriginalTags: originalTags,
				FinalTags:    req.Tags,
				CorrectionType: h.determineCorrectionType(originalTags, req.Tags),
			}
			h.learningRepo.SaveTagCorrection(correction)
		}()
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

	// Get user ID from context
	userID := h.getUserID(r)

	err = h.repo.Delete(id, userID)
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

	// Get user ID from context
	userID := h.getUserID(r)

	results, err := h.repo.Search(query, limit, userID)
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

// fuzzySearchBookmarks handles GET /api/bookmarks/fuzzy-search
func (h *BookmarkHandler) fuzzySearchBookmarks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.writeError(w, "Search query is required", http.StatusBadRequest)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// First try exact FTS search
	exactResults, err := h.repo.Search(query, limit, userID)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// If we have good exact results, return them
	if len(exactResults) >= limit/2 {
		response := map[string]interface{}{
			"query":       query,
			"results":     exactResults,
			"count":       len(exactResults),
			"search_type": "exact",
		}
		h.writeJSON(w, response)
		return
	}

	// Get all bookmarks for fuzzy matching
	allBookmarks, err := h.repo.List(1, 1000, "", "", false, userID) // Get more bookmarks for fuzzy matching
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get bookmarks for fuzzy search: %v", err), http.StatusInternalServerError)
		return
	}

	// Prepare candidates for fuzzy matching
	var titleCandidates []string
	var descriptionCandidates []string
	var tagCandidates []string
	bookmarkIndex := make(map[string]models.Bookmark)
	
	for _, bookmark := range allBookmarks.Bookmarks {
		// Index by title
		titleCandidates = append(titleCandidates, bookmark.Title)
		bookmarkIndex[bookmark.Title] = bookmark
		
		// Index by description
		if bookmark.Description != nil && *bookmark.Description != "" {
			descriptionCandidates = append(descriptionCandidates, *bookmark.Description)
			bookmarkIndex[*bookmark.Description] = bookmark
		}
		
		// Index by tags
		for _, tag := range bookmark.Tags {
			tagKey := "tag:" + tag.Name
			tagCandidates = append(tagCandidates, tagKey)
			bookmarkIndex[tagKey] = bookmark
		}
	}

	// Perform fuzzy search on titles
	titleMatches := h.fuzzyMatcher.Search(query, titleCandidates)
	
	// Perform fuzzy search on descriptions
	descMatches := h.fuzzyMatcher.Search(query, descriptionCandidates)
	
	// Perform fuzzy search on tags
	tagMatches := h.fuzzyMatcher.Search(query, tagCandidates)

	// Combine and deduplicate results
	seenBookmarks := make(map[int]bool)
	var fuzzyResults []models.SearchResult
	
	// Process title matches first (highest priority)
	for _, match := range titleMatches {
		if bookmark, exists := bookmarkIndex[match.Text]; exists {
			if !seenBookmarks[bookmark.ID] {
				fuzzyResults = append(fuzzyResults, models.SearchResult{
					Bookmark: bookmark,
					Rank:     match.Similarity,
					Snippet:  h.createSnippet(bookmark.Title, query),
				})
				seenBookmarks[bookmark.ID] = true
			}
		}
	}
	
	// Process description matches
	for _, match := range descMatches {
		if bookmark, exists := bookmarkIndex[match.Text]; exists {
			if !seenBookmarks[bookmark.ID] && len(fuzzyResults) < limit {
				fuzzyResults = append(fuzzyResults, models.SearchResult{
					Bookmark: bookmark,
					Rank:     match.Similarity * 0.8, // Lower priority for description matches
					Snippet:  h.createSnippet(match.Text, query),
				})
				seenBookmarks[bookmark.ID] = true
			}
		}
	}
	
	// Process tag matches
	for _, match := range tagMatches {
		if bookmark, exists := bookmarkIndex[match.Text]; exists {
			if !seenBookmarks[bookmark.ID] && len(fuzzyResults) < limit {
				fuzzyResults = append(fuzzyResults, models.SearchResult{
					Bookmark: bookmark,
					Rank:     match.Similarity * 0.6, // Lower priority for tag matches
					Snippet:  strings.TrimPrefix(match.Text, "tag:"),
				})
				seenBookmarks[bookmark.ID] = true
			}
		}
	}
	
	// Limit results
	if len(fuzzyResults) > limit {
		fuzzyResults = fuzzyResults[:limit]
	}

	response := map[string]interface{}{
		"query":        query,
		"results":      fuzzyResults,
		"count":        len(fuzzyResults),
		"search_type":  "fuzzy",
		"exact_count":  len(exactResults),
	}

	h.writeJSON(w, response)
}

// createSnippet creates a search snippet with highlighted terms
func (h *BookmarkHandler) createSnippet(text, query string) string {
	if len(text) <= 100 {
		return text
	}
	
	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)
	
	// Find the query in the text
	index := strings.Index(lowerText, lowerQuery)
	if index == -1 {
		// If not found, return first 100 chars
		return text[:100] + "..."
	}
	
	// Create snippet around the match
	start := max(0, index-30)
	end := min(len(text), index+len(query)+30)
	
	snippet := text[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(text) {
		snippet = snippet + "..."
	}
	
	return snippet
}

// suggestTags handles POST /api/bookmarks/suggest-tags
func (h *BookmarkHandler) suggestTags(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL         string `json:"url"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.writeError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Create temporary bookmark for categorization
	var description *string
	if req.Description != "" {
		description = &req.Description
	}
	tempBookmark := &models.Bookmark{
		URL:         req.URL,
		Title:       req.Title,
		Description: description,
	}

	// Get AI suggestions
	suggestions, err := h.categorizer.CategorizeBookmark(tempBookmark)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"tags":       suggestions.Tags,
		"category":   suggestions.Category,
		"confidence": suggestions.Confidence,
		"source":     suggestions.Source,
	}

	h.writeJSON(w, response)
}

// getUserID extracts user ID from request context, defaults to 1 for backward compatibility
func (h *BookmarkHandler) getUserID(r *http.Request) int {
	if userID, ok := middleware.GetUserIDFromContext(r); ok {
		return userID
	}
	// Default to user ID 1 for backward compatibility with single-user setup
	return 1
}

// writeJSON writes a JSON response
func (h *BookmarkHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.MarshalWrite(w, data); err != nil {
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
	
	json.MarshalWrite(w, response)
}

// determineCorrectionType determines the type of tag correction
func (h *BookmarkHandler) determineCorrectionType(originalTags, finalTags []string) string {
	originalSet := make(map[string]bool)
	finalSet := make(map[string]bool)
	
	for _, tag := range originalTags {
		originalSet[tag] = true
	}
	
	for _, tag := range finalTags {
		finalSet[tag] = true
	}
	
	var added, removed int
	
	// Count added tags
	for _, tag := range finalTags {
		if !originalSet[tag] {
			added++
		}
	}
	
	// Count removed tags
	for _, tag := range originalTags {
		if !finalSet[tag] {
			removed++
		}
	}
	
	if added > 0 && removed == 0 {
		return "added"
	} else if removed > 0 && added == 0 {
		return "removed"
	} else if added > 0 && removed > 0 {
		return "modified"
	}
	
	return "kept"
}

// Helper functions for fuzzy search
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}