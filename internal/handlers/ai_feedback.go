package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"torimemo/internal/ai"
	"torimemo/internal/db"
	"torimemo/internal/middleware"
	"torimemo/internal/models"
)

// AIFeedbackHandler handles AI feedback and training operations
type AIFeedbackHandler struct {
	bookmarkRepo *db.BookmarkRepository
	learningRepo *db.LearningRepository
	categorizer  *ai.Categorizer
}

// NewAIFeedbackHandler creates a new AI feedback handler
func NewAIFeedbackHandler(bookmarkRepo *db.BookmarkRepository, learningRepo *db.LearningRepository) *AIFeedbackHandler {
	return &AIFeedbackHandler{
		bookmarkRepo: bookmarkRepo,
		learningRepo: learningRepo,
		categorizer:  ai.NewCategorizerWithLearning(learningRepo),
	}
}

// RegisterRoutes registers AI feedback routes
func (h *AIFeedbackHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/ai/suggest-tags", h.suggestTags)
	mux.HandleFunc("/api/ai/feedback", h.submitFeedback)
	mux.HandleFunc("/api/ai/recategorize", h.recategorizeBookmark)
	mux.HandleFunc("/api/ai/batch-recategorize", h.batchRecategorize)
	mux.HandleFunc("/api/ai/training-status", h.getTrainingStatus)
}

// TagSuggestionRequest represents a request for AI tag suggestions
type TagSuggestionRequest struct {
	URL         string `json:"url"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// TagSuggestionResponse represents AI tag suggestions with metadata
type TagSuggestionResponse struct {
	*ai.TagSuggestion
	ExistingBookmark *models.Bookmark `json:"existing_bookmark,omitempty"`
	IsDuplicate      bool             `json:"is_duplicate"`
}

// FeedbackRequest represents user feedback on AI suggestions
type FeedbackRequest struct {
	BookmarkID      int      `json:"bookmark_id"`
	OriginalTags    []string `json:"original_tags"`
	AcceptedTags    []string `json:"accepted_tags"`
	RejectedTags    []string `json:"rejected_tags"`
	UserTags        []string `json:"user_tags"`
	FeedbackType    string   `json:"feedback_type"` // "accept", "reject", "modify"
	ConfidenceRating int     `json:"confidence_rating,omitempty"` // 1-5 user rating
}

// RecategorizeRequest represents a request to recategorize bookmarks
type RecategorizeRequest struct {
	BookmarkIDs []int  `json:"bookmark_ids,omitempty"`
	UserID      *int   `json:"user_id,omitempty"`
	Domain      string `json:"domain,omitempty"`
	ForceUpdate bool   `json:"force_update"`
}

// suggestTags handles POST /api/ai/suggest-tags
func (h *AIFeedbackHandler) suggestTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TagSuggestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.writeError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Check if bookmark already exists
	existing, _ := h.bookmarkRepo.GetByURL(req.URL, userID)

	// Create temporary bookmark for AI analysis
	tempBookmark := &models.Bookmark{
		URL:   req.URL,
		Title: req.Title,
	}
	if req.Description != "" {
		tempBookmark.Description = &req.Description
	}

	// Get AI suggestions
	suggestions, err := h.categorizer.CategorizeBookmark(tempBookmark)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get AI suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	response := &TagSuggestionResponse{
		TagSuggestion:    suggestions,
		ExistingBookmark: existing,
		IsDuplicate:      existing != nil,
	}

	h.writeJSON(w, response)
}

// submitFeedback handles POST /api/ai/feedback
func (h *AIFeedbackHandler) submitFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.BookmarkID == 0 {
		h.writeError(w, "Bookmark ID is required", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Get the bookmark to ensure it exists and belongs to user
	bookmark, err := h.bookmarkRepo.GetByID(req.BookmarkID, userID)
	if err != nil {
		h.writeError(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	// Process the feedback based on type
	switch req.FeedbackType {
	case "accept":
		err = h.processFeedbackAccept(bookmark, req.AcceptedTags, req.ConfidenceRating)
	case "reject":
		err = h.processFeedbackReject(bookmark, req.RejectedTags, req.ConfidenceRating)
	case "modify":
		err = h.processFeedbackModify(bookmark, req.OriginalTags, req.UserTags, req.ConfidenceRating)
	default:
		h.writeError(w, "Invalid feedback type", http.StatusBadRequest)
		return
	}

	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to process feedback: %v", err), http.StatusInternalServerError)
		return
	}

	// Update domain profile based on feedback
	h.updateDomainProfile(bookmark.URL, req.UserTags)

	response := map[string]interface{}{
		"success": true,
		"message": "Feedback recorded successfully",
	}

	h.writeJSON(w, response)
}

// recategorizeBookmark handles POST /api/ai/recategorize
func (h *AIFeedbackHandler) recategorizeBookmark(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get bookmark ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/ai/recategorize/")
	if path == "" {
		h.writeError(w, "Bookmark ID required", http.StatusBadRequest)
		return
	}

	bookmarkID, err := strconv.Atoi(path)
	if err != nil {
		h.writeError(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Get the bookmark
	bookmark, err := h.bookmarkRepo.GetByID(bookmarkID, userID)
	if err != nil {
		h.writeError(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	// Get new AI suggestions
	suggestions, err := h.categorizer.CategorizeBookmark(bookmark)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get AI suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	// Update bookmark with new suggestions
	updateReq := &models.UpdateBookmarkRequest{
		Tags: suggestions.Tags,
	}

	// Update title and description if AI provided better ones
	if suggestions.Title != "" && suggestions.Title != bookmark.Title {
		updateReq.Title = &suggestions.Title
	}
	if suggestions.Description != "" && (bookmark.Description == nil || *bookmark.Description != suggestions.Description) {
		updateReq.Description = &suggestions.Description
	}

	updatedBookmark, err := h.bookmarkRepo.Update(bookmarkID, updateReq, userID)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to update bookmark: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"bookmark":    updatedBookmark,
		"suggestions": suggestions,
		"updated":     true,
	}

	h.writeJSON(w, response)
}

// batchRecategorize handles POST /api/ai/batch-recategorize
func (h *AIFeedbackHandler) batchRecategorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RecategorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := h.getUserID(r)

	var bookmarks []models.Bookmark

	// Determine which bookmarks to recategorize
	if len(req.BookmarkIDs) > 0 {
		// Specific bookmark IDs
		for _, id := range req.BookmarkIDs {
			if bookmark, err := h.bookmarkRepo.GetByID(id, userID); err == nil {
				bookmarks = append(bookmarks, *bookmark)
			}
		}
	} else if req.Domain != "" {
		// All bookmarks from specific domain
		allBookmarks, err := h.bookmarkRepo.List(1, 10000, "", "", false, userID)
		if err != nil {
			h.writeError(w, "Failed to get bookmarks", http.StatusInternalServerError)
			return
		}
		for _, bookmark := range allBookmarks.Bookmarks {
			if strings.Contains(bookmark.URL, req.Domain) {
				bookmarks = append(bookmarks, bookmark)
			}
		}
	} else {
		// All user bookmarks
		allBookmarks, err := h.bookmarkRepo.List(1, 10000, "", "", false, userID)
		if err != nil {
			h.writeError(w, "Failed to get bookmarks", http.StatusInternalServerError)
			return
		}
		bookmarks = allBookmarks.Bookmarks
	}

	// Process bookmarks in batches
	updated := 0
	failed := 0

	for _, bookmark := range bookmarks {
		// Skip if not forced update and bookmark has tags
		if !req.ForceUpdate && len(bookmark.Tags) > 0 {
			continue
		}

		suggestions, err := h.categorizer.CategorizeBookmark(&bookmark)
		if err != nil {
			failed++
			continue
		}

		// Only update if AI has good suggestions
		if len(suggestions.Tags) > 0 && suggestions.Confidence > 0.5 {
			updateReq := &models.UpdateBookmarkRequest{
				Tags: suggestions.Tags,
			}

			if _, updateErr := h.bookmarkRepo.Update(bookmark.ID, updateReq, userID); updateErr != nil {
				failed++
			} else {
				updated++
			}
		}
	}

	response := map[string]interface{}{
		"total_processed": len(bookmarks),
		"updated":         updated,
		"failed":          failed,
		"success":         true,
	}

	h.writeJSON(w, response)
}

// getTrainingStatus handles GET /api/ai/training-status
func (h *AIFeedbackHandler) getTrainingStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get learning statistics
	corrections, err := h.learningRepo.GetTagCorrections(100)
	if err != nil {
		h.writeError(w, "Failed to get corrections", http.StatusInternalServerError)
		return
	}

	// Calculate training metrics
	totalCorrections := len(corrections)
	recentCorrections := 0
	weekAgo := time.Now().AddDate(0, 0, -7)

	for _, correction := range corrections {
		if correction.CreatedAt.After(weekAgo) {
			recentCorrections++
		}
	}

	response := map[string]interface{}{
		"total_corrections":  totalCorrections,
		"recent_corrections": recentCorrections,
		"learning_active":    totalCorrections > 10,
		"confidence_boost":   float64(totalCorrections) * 0.01, // Simple metric
		"last_updated":       time.Now().Format(time.RFC3339),
	}

	h.writeJSON(w, response)
}

// Helper methods

func (h *AIFeedbackHandler) processFeedbackAccept(bookmark *models.Bookmark, acceptedTags []string, rating int) error {
	// Save as learned pattern with high confidence
	return h.saveLearnedPattern(bookmark, acceptedTags, 0.8+float64(rating)*0.04)
}

func (h *AIFeedbackHandler) processFeedbackReject(bookmark *models.Bookmark, rejectedTags []string, rating int) error {
	// Save tag correction showing what was rejected
	correction := &models.TagCorrection{
		BookmarkID:     bookmark.ID,
		OriginalTags:   rejectedTags,
		FinalTags:      []string{}, // User rejected all
		CorrectionType: "rejected",
		CreatedAt:      time.Now(),
	}
	return h.learningRepo.SaveTagCorrection(correction)
}

func (h *AIFeedbackHandler) processFeedbackModify(bookmark *models.Bookmark, originalTags, userTags []string, rating int) error {
	// Save tag correction showing the modification
	correction := &models.TagCorrection{
		BookmarkID:     bookmark.ID,
		OriginalTags:   originalTags,
		FinalTags:      userTags,
		CorrectionType: "modified",
		CreatedAt:      time.Now(),
	}
	
	if err := h.learningRepo.SaveTagCorrection(correction); err != nil {
		return err
	}

	// Save as learned pattern with moderate confidence
	return h.saveLearnedPattern(bookmark, userTags, 0.6+float64(rating)*0.06)
}

func (h *AIFeedbackHandler) saveLearnedPattern(bookmark *models.Bookmark, tags []string, confidence float64) error {
	pattern := &models.LearnedPattern{
		URLPattern:    bookmark.URL,
		Domain:        h.extractDomain(bookmark.URL),
		ConfirmedTags: tags,
		Confidence:    confidence,
		SampleCount:   1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	return h.learningRepo.SavePattern(pattern)
}

func (h *AIFeedbackHandler) updateDomainProfile(url string, tags []string) {
	domain := h.extractDomain(url)
	if domain != "" {
		go func() {
			h.learningRepo.UpdateDomainProfile(domain, tags)
		}()
	}
}

func (h *AIFeedbackHandler) extractDomain(url string) string {
	if parsedURL, err := h.parseURL(url); err == nil {
		return parsedURL.Hostname()
	}
	return ""
}

func (h *AIFeedbackHandler) parseURL(urlStr string) (*url.URL, error) {
	return url.Parse(urlStr)
}

func (h *AIFeedbackHandler) getUserID(r *http.Request) int {
	if userID, ok := middleware.GetUserIDFromContext(r); ok {
		return userID
	}
	return 1 // Default for backward compatibility
}

func (h *AIFeedbackHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func (h *AIFeedbackHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}
	
	json.NewEncoder(w).Encode(response)
}