package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

// LearningHandler handles learning and feedback operations
type LearningHandler struct {
	bookmarkRepo *db.BookmarkRepository
	learningRepo *db.LearningRepository
}

// NewLearningHandler creates a new learning handler
func NewLearningHandler(bookmarkRepo *db.BookmarkRepository, learningRepo *db.LearningRepository) *LearningHandler {
	return &LearningHandler{
		bookmarkRepo: bookmarkRepo,
		learningRepo: learningRepo,
	}
}

// ServeHTTP implements the http.Handler interface
func (h *LearningHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/learning")
	
	switch {
	case r.Method == "POST" && path == "/feedback":
		h.submitFeedback(w, r)
	case r.Method == "GET" && path == "/patterns":
		h.getLearnedPatterns(w, r)
	case r.Method == "GET" && path == "/corrections":
		h.getTagCorrections(w, r)
	case r.Method == "POST" && path == "/retrain":
		h.retrainFromCorrections(w, r)
	default:
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// submitFeedback handles POST /api/learning/feedback
func (h *LearningHandler) submitFeedback(w http.ResponseWriter, r *http.Request) {
	var feedback struct {
		BookmarkID    int      `json:"bookmark_id"`
		SuggestedTags []string `json:"suggested_tags"`
		FinalTags     []string `json:"final_tags"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&feedback); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if feedback.BookmarkID == 0 {
		h.writeError(w, "Bookmark ID is required", http.StatusBadRequest)
		return
	}

	// Get bookmark to analyze URL patterns  
	userID := 1 // Default userID for learning operations
	bookmark, err := h.bookmarkRepo.GetByID(feedback.BookmarkID, userID)
	if err != nil {
		h.writeError(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	// Analyze what changed
	keptTags, addedTags, removedTags := h.analyzeFeedback(feedback.SuggestedTags, feedback.FinalTags)

	// Save tag correction
	correction := &models.TagCorrection{
		BookmarkID:     feedback.BookmarkID,
		OriginalTags:   feedback.SuggestedTags,
		FinalTags:      feedback.FinalTags,
		CorrectionType: h.determineCorrectionType(keptTags, addedTags, removedTags),
	}

	if err := h.learningRepo.SaveTagCorrection(correction); err != nil {
		h.writeError(w, "Failed to save correction", http.StatusInternalServerError)
		return
	}

	// Update domain profile based on feedback
	if len(feedback.FinalTags) > 0 {
		go func() {
			if parsedURL, err := parseURL(bookmark.URL); err == nil {
				domain := parsedURL.Hostname()
				h.learningRepo.UpdateDomainProfile(domain, feedback.FinalTags)
			}
		}()
	}

	response := map[string]interface{}{
		"message":      "Feedback recorded successfully",
		"kept_tags":    len(keptTags),
		"added_tags":   len(addedTags),
		"removed_tags": len(removedTags),
		"correction_type": correction.CorrectionType,
	}

	h.writeJSON(w, response)
}

// getLearnedPatterns handles GET /api/learning/patterns
func (h *LearningHandler) getLearnedPatterns(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// For now, return basic stats - could expand to show actual patterns
	response := map[string]interface{}{
		"message": "Learning patterns endpoint - future expansion",
		"note": "Patterns are stored and used internally for categorization improvement",
	}

	h.writeJSON(w, response)
}

// getTagCorrections handles GET /api/learning/corrections
func (h *LearningHandler) getTagCorrections(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	corrections, err := h.learningRepo.GetTagCorrections(limit)
	if err != nil {
		h.writeError(w, "Failed to get corrections", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"corrections": corrections,
		"count":       len(corrections),
	}

	h.writeJSON(w, response)
}

// retrainFromCorrections handles POST /api/learning/retrain
func (h *LearningHandler) retrainFromCorrections(w http.ResponseWriter, r *http.Request) {
	// Get recent corrections for analysis
	corrections, err := h.learningRepo.GetTagCorrections(100)
	if err != nil {
		h.writeError(w, "Failed to get corrections", http.StatusInternalServerError)
		return
	}

	// Analyze corrections to improve patterns
	improvements := h.analyzeCorrectionsForPatterns(corrections)
	
	response := map[string]interface{}{
		"message": "Retraining analysis completed",
		"corrections_analyzed": len(corrections),
		"improvements_found": len(improvements),
		"patterns_updated": 0, // Placeholder for future ML model updates
	}

	h.writeJSON(w, response)
}

// Helper methods

func (h *LearningHandler) analyzeFeedback(suggested, final []string) (kept, added, removed []string) {
	suggestedSet := make(map[string]bool)
	finalSet := make(map[string]bool)
	
	for _, tag := range suggested {
		suggestedSet[tag] = true
	}
	
	for _, tag := range final {
		finalSet[tag] = true
	}
	
	// Find kept tags (in both)
	for _, tag := range suggested {
		if finalSet[tag] {
			kept = append(kept, tag)
		}
	}
	
	// Find added tags (in final but not in suggested)
	for _, tag := range final {
		if !suggestedSet[tag] {
			added = append(added, tag)
		}
	}
	
	// Find removed tags (in suggested but not in final)
	for _, tag := range suggested {
		if !finalSet[tag] {
			removed = append(removed, tag)
		}
	}
	
	return kept, added, removed
}

func (h *LearningHandler) determineCorrectionType(kept, added, removed []string) string {
	if len(added) > 0 && len(removed) == 0 {
		return "added"
	} else if len(removed) > 0 && len(added) == 0 {
		return "removed"
	} else if len(added) > 0 && len(removed) > 0 {
		return "modified"
	} else if len(kept) > 0 {
		return "kept"
	}
	return "unknown"
}

func (h *LearningHandler) analyzeCorrectionsForPatterns(corrections []models.TagCorrection) []string {
	// Analyze patterns in corrections to identify:
	// 1. Commonly rejected AI suggestions
	// 2. Commonly added user tags
	// 3. Domain-specific preferences
	
	var improvements []string
	
	// Count tag frequencies
	rejectedTags := make(map[string]int)
	addedTags := make(map[string]int)
	
	for _, correction := range corrections {
		// Analyze what was consistently rejected
		for _, original := range correction.OriginalTags {
			found := false
			for _, final := range correction.FinalTags {
				if original == final {
					found = true
					break
				}
			}
			if !found {
				rejectedTags[original]++
			}
		}
		
		// Analyze what was consistently added
		for _, final := range correction.FinalTags {
			found := false
			for _, original := range correction.OriginalTags {
				if final == original {
					found = true
					break
				}
			}
			if !found {
				addedTags[final]++
			}
		}
	}
	
	// Generate improvement suggestions
	for tag, count := range rejectedTags {
		if count >= 3 { // If rejected 3+ times
			improvements = append(improvements, "Consider reducing confidence for tag: "+tag)
		}
	}
	
	for tag, count := range addedTags {
		if count >= 3 { // If added 3+ times  
			improvements = append(improvements, "Consider adding rule for commonly added tag: "+tag)
		}
	}
	
	return improvements
}

// writeJSON writes a JSON response
func (h *LearningHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *LearningHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}
	
	json.NewEncoder(w).Encode(response)
}

// parseURL is a helper to parse URLs safely
func parseURL(urlStr string) (*url.URL, error) {
	return url.Parse(urlStr)
}