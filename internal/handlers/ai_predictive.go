package handlers

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"torimemo/internal/ai"
	"torimemo/internal/db"
	"torimemo/internal/models"
)

// AIPredictiveHandler handles AI-powered predictive tag suggestions
type AIPredictiveHandler struct {
	bookmarkRepo   *db.BookmarkRepository
	learningRepo   *db.LearningRepository
	predictiveEngine *ai.PredictiveTagEngine
}

// NewAIPredictiveHandler creates a new predictive handler
func NewAIPredictiveHandler(bookmarkRepo *db.BookmarkRepository, learningRepo *db.LearningRepository) *AIPredictiveHandler {
	return &AIPredictiveHandler{
		bookmarkRepo:     bookmarkRepo,
		learningRepo:     learningRepo,
		predictiveEngine: ai.NewPredictiveTagEngine(),
	}
}

// RegisterRoutes registers predictive tag routes
func (h *AIPredictiveHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/ai/predict-tags", h.predictTags)
	mux.HandleFunc("/api/ai/predict/learn", h.learnFromFeedback)
	mux.HandleFunc("/api/ai/predict/analyze", h.analyzeUserPatterns)
	mux.HandleFunc("/api/ai/predict/reset", h.resetUserPatterns)
}

// predictTags handles POST /api/ai/predict-tags
func (h *AIPredictiveHandler) predictTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ai.PredictiveAnalysisRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req.Context.UserID = userID

	// Set defaults
	if req.MaxSuggestions == 0 {
		req.MaxSuggestions = 10
	}
	if req.MinConfidence == 0 {
		req.MinConfidence = 0.3
	}

	// Populate time if not provided
	if req.Context.Time.IsZero() {
		req.Context.Time = time.Now()
	}

	// First, train the predictive engine from user's historical data
	if err := h.trainFromUserHistory(userID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to train predictive engine: %v\n", err)
	}

	// Get predictions
	result, err := h.predictiveEngine.PredictTags(req)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Prediction failed: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, result)
}

// learnFromFeedback handles POST /api/ai/predict/learn
func (h *AIPredictiveHandler) learnFromFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Context      ai.PredictionContext `json:"context"`
		SelectedTags []string             `json:"selected_tags"`
		RejectedTags []string             `json:"rejected_tags"`
	}

	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req.Context.UserID = userID

	// Learn from feedback
	err := h.predictiveEngine.LearnFromUserFeedback(userID, req.Context, req.SelectedTags, req.RejectedTags)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Learning failed: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Learned from feedback",
		"learned_tags": len(req.SelectedTags),
		"rejected_tags": len(req.RejectedTags),
	}

	h.writeJSON(w, response)
}

// analyzeUserPatterns handles GET /api/ai/predict/analyze
func (h *AIPredictiveHandler) analyzeUserPatterns(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// First train from user history
	if err := h.trainFromUserHistory(userID); err != nil {
		h.writeError(w, fmt.Sprintf("Failed to analyze patterns: %v", err), http.StatusInternalServerError)
		return
	}

	// Get user pattern analysis
	analysis, err := h.analyzeUserTaggingPatterns(userID)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Pattern analysis failed: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, analysis)
}

// resetUserPatterns handles POST /api/ai/predict/reset
func (h *AIPredictiveHandler) resetUserPatterns(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Reset user patterns by creating a new engine instance
	// In a real implementation, you might want to clear from persistent storage
	h.predictiveEngine = ai.NewPredictiveTagEngine()

	response := map[string]interface{}{
		"success": true,
		"message": "User patterns reset successfully",
	}

	h.writeJSON(w, response)
}

// Helper methods

func (h *AIPredictiveHandler) getUserID(r *http.Request) int {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		return 0
	}
	
	id, err := strconv.Atoi(userID)
	if err != nil {
		return 0
	}
	
	return id
}

func (h *AIPredictiveHandler) trainFromUserHistory(userID int) error {
	// Get user's bookmarks with tags
	response, err := h.bookmarkRepo.List(1, 1000, "", "", false, userID) // Get recent bookmarks
	if err != nil {
		return fmt.Errorf("failed to get user bookmarks: %w", err)
	}

	// Process bookmarks to train predictive patterns
	for _, bookmark := range response.Bookmarks {
		if len(bookmark.Tags) == 0 {
			continue
		}

		// Extract tag names from Tag structs
		var tagNames []string
		for _, tag := range bookmark.Tags {
			tagNames = append(tagNames, tag.Name)
		}

		// Create training context
		context := ai.PredictionContext{
			UserID:       userID,
			URL:          bookmark.URL,
			Title:        bookmark.Title,
			Description:  h.getStringValue(bookmark.Description),
			ExistingTags: []string{}, // No existing tags for training
			Time:         bookmark.CreatedAt,
		}

		// Train the engine with this bookmark's tags
		err = h.predictiveEngine.LearnFromUserFeedback(userID, context, tagNames, []string{})
		if err != nil {
			// Log error but continue training
			fmt.Printf("Warning: Failed to train from bookmark %d: %v\n", bookmark.ID, err)
		}
	}

	return nil
}

func (h *AIPredictiveHandler) analyzeUserTaggingPatterns(userID int) (map[string]interface{}, error) {
	// Get user's bookmarks for analysis
	response, err := h.bookmarkRepo.List(1, 1000, "", "", false, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarks: %w", err)
	}

	analysis := make(map[string]interface{})

	// Basic statistics
	totalBookmarks := len(response.Bookmarks)
	taggedBookmarks := 0
	allTags := make(map[string]int)
	domainTagging := make(map[string]map[string]int)

	for _, bookmark := range response.Bookmarks {
		if len(bookmark.Tags) > 0 {
			taggedBookmarks++
		}

		domain := h.extractDomain(bookmark.URL)
		if domainTagging[domain] == nil {
			domainTagging[domain] = make(map[string]int)
		}

		for _, tag := range bookmark.Tags {
			allTags[tag.Name]++
			domainTagging[domain][tag.Name]++
		}
	}

	analysis["total_bookmarks"] = totalBookmarks
	analysis["tagged_bookmarks"] = taggedBookmarks
	analysis["tagging_rate"] = float64(taggedBookmarks) / float64(totalBookmarks)
	analysis["unique_tags"] = len(allTags)

	// Top tags
	topTags := h.getTopTags(allTags, 10)
	analysis["top_tags"] = topTags

	// Domain analysis
	domainStats := make(map[string]interface{})
	for domain, tags := range domainTagging {
		if len(tags) > 0 {
			domainStats[domain] = map[string]interface{}{
				"unique_tags": len(tags),
				"top_tags":    h.getTopTags(tags, 5),
			}
		}
	}
	analysis["domain_patterns"] = domainStats

	// Time patterns
	timePatterns := h.analyzeTimePatterns(response.Bookmarks)
	analysis["time_patterns"] = timePatterns

	// Tagging consistency
	consistency := h.calculateTaggingConsistency(response.Bookmarks)
	analysis["consistency_score"] = consistency

	// Insights
	insights := h.generatePatternInsights(analysis)
	analysis["insights"] = insights

	return analysis, nil
}

func (h *AIPredictiveHandler) getTopTags(tagCount map[string]int, limit int) []map[string]interface{} {
	type tagFreq struct {
		tag   string
		count int
	}

	var tags []tagFreq
	for tag, count := range tagCount {
		tags = append(tags, tagFreq{tag, count})
	}

	// Sort by frequency
	for i := 0; i < len(tags)-1; i++ {
		for j := i + 1; j < len(tags); j++ {
			if tags[i].count < tags[j].count {
				tags[i], tags[j] = tags[j], tags[i]
			}
		}
	}

	// Convert to response format
	var result []map[string]interface{}
	actualLimit := limit
	if len(tags) < actualLimit {
		actualLimit = len(tags)
	}

	for i := 0; i < actualLimit; i++ {
		result = append(result, map[string]interface{}{
			"tag":   tags[i].tag,
			"count": tags[i].count,
		})
	}

	return result
}

func (h *AIPredictiveHandler) analyzeTimePatterns(bookmarks []models.Bookmark) map[string]interface{} {
	hourCounts := make(map[int]int)
	dayOfWeekCounts := make(map[time.Weekday]int)

	for _, bookmark := range bookmarks {
		if len(bookmark.Tags) > 0 {
			hour := bookmark.CreatedAt.Hour()
			dayOfWeek := bookmark.CreatedAt.Weekday()
			
			hourCounts[hour]++
			dayOfWeekCounts[dayOfWeek]++
		}
	}

	// Find peak hours and days
	peakHour := 0
	maxHourCount := 0
	for hour, count := range hourCounts {
		if count > maxHourCount {
			maxHourCount = count
			peakHour = hour
		}
	}

	peakDay := time.Sunday
	maxDayCount := 0
	for day, count := range dayOfWeekCounts {
		if count > maxDayCount {
			maxDayCount = count
			peakDay = day
		}
	}

	return map[string]interface{}{
		"peak_hour":     peakHour,
		"peak_day":      peakDay.String(),
		"hourly_distribution": hourCounts,
		"daily_distribution":  h.convertWeekdayMap(dayOfWeekCounts),
	}
}

func (h *AIPredictiveHandler) convertWeekdayMap(weekdayMap map[time.Weekday]int) map[string]int {
	result := make(map[string]int)
	for day, count := range weekdayMap {
		result[day.String()] = count
	}
	return result
}

func (h *AIPredictiveHandler) calculateTaggingConsistency(bookmarks []models.Bookmark) float64 {
	// Simple consistency metric: ratio of tagged vs untagged bookmarks
	tagged := 0
	total := len(bookmarks)

	for _, bookmark := range bookmarks {
		if len(bookmark.Tags) > 0 {
			tagged++
		}
	}

	if total == 0 {
		return 0.0
	}

	return float64(tagged) / float64(total)
}

func (h *AIPredictiveHandler) generatePatternInsights(analysis map[string]interface{}) []string {
	var insights []string

	taggingRate := analysis["tagging_rate"].(float64)
	if taggingRate > 0.8 {
		insights = append(insights, "You have excellent tagging consistency")
	} else if taggingRate < 0.3 {
		insights = append(insights, "Consider tagging more bookmarks for better predictions")
	}

	uniqueTags := analysis["unique_tags"].(int)
	if uniqueTags > 50 {
		insights = append(insights, "You have a rich tag vocabulary")
	} else if uniqueTags < 10 {
		insights = append(insights, "Your tag vocabulary could be expanded")
	}

	return insights
}

func (h *AIPredictiveHandler) extractDomain(url string) string {
	url = strings.ToLower(url)
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}
	
	if strings.HasPrefix(url, "www.") {
		url = url[4:]
	}
	
	return url
}

func (h *AIPredictiveHandler) getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func (h *AIPredictiveHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.MarshalWrite(w, data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func (h *AIPredictiveHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}
	
	json.MarshalWrite(w, response)
}