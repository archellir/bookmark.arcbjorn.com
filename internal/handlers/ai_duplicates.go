package handlers

import (
	"encoding/json/v2"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"torimemo/internal/ai"
	"torimemo/internal/db"
	"torimemo/internal/middleware"
	"torimemo/internal/models"
)

// AIDuplicatesHandler handles AI-powered duplicate detection
type AIDuplicatesHandler struct {
	bookmarkRepo    *db.BookmarkRepository
	similarityEngine *ai.SimilarityEngine
}

// NewAIDuplicatesHandler creates a new AI duplicates handler
func NewAIDuplicatesHandler(bookmarkRepo *db.BookmarkRepository) *AIDuplicatesHandler {
	return &AIDuplicatesHandler{
		bookmarkRepo:    bookmarkRepo,
		similarityEngine: ai.NewSimilarityEngine(),
	}
}

// RegisterRoutes registers AI duplicate detection routes
func (h *AIDuplicatesHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/ai/duplicates/check", h.checkDuplicates)
	mux.HandleFunc("/api/ai/duplicates/find-all", h.findAllDuplicates)
	mux.HandleFunc("/api/ai/duplicates/merge", h.mergeDuplicates)
	mux.HandleFunc("/api/ai/duplicates/analyze", h.analyzeDuplicates)
}

// AICheckDuplicatesRequest represents an AI-powered duplicate check request
type AICheckDuplicatesRequest struct {
	URL         string  `json:"url"`
	Title       string  `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Threshold   float64 `json:"threshold,omitempty"`
}

// DuplicatesResponse represents the duplicate check response
type DuplicatesResponse struct {
	HasDuplicates bool                     `json:"has_duplicates"`
	Matches       []ai.DuplicateMatch     `json:"matches"`
	Total         int                     `json:"total"`
	ExactMatches  int                     `json:"exact_matches"`
	SimilarMatches int                    `json:"similar_matches"`
	Recommendations []string              `json:"recommendations"`
}

// FindAllDuplicatesResponse represents response for finding all duplicates
type FindAllDuplicatesResponse struct {
	DuplicateGroups []DuplicateGroup `json:"duplicate_groups"`
	TotalGroups     int              `json:"total_groups"`
	TotalBookmarks  int              `json:"total_bookmarks"`
	Statistics      DuplicateStats   `json:"statistics"`
}

// DuplicateGroup represents a group of similar bookmarks
type DuplicateGroup struct {
	ID          int                     `json:"id"`
	Primary     *models.Bookmark        `json:"primary"`
	Duplicates  []ai.DuplicateMatch     `json:"duplicates"`
	GroupScore  float64                 `json:"group_score"`
	GroupType   string                  `json:"group_type"`
	Confidence  float64                 `json:"confidence"`
}

// DuplicateStats represents overall duplicate statistics
type DuplicateStats struct {
	TotalBookmarks    int     `json:"total_bookmarks"`
	DuplicateCount    int     `json:"duplicate_count"`
	DuplicateRate     float64 `json:"duplicate_rate"`
	ExactDuplicates   int     `json:"exact_duplicates"`
	SimilarDuplicates int     `json:"similar_duplicates"`
	TopDomains        []DomainDuplicateInfo `json:"top_domains"`
}

// DomainDuplicateInfo represents duplicate information for a domain
type DomainDuplicateInfo struct {
	Domain          string  `json:"domain"`
	Count           int     `json:"count"`
	DuplicateRate   float64 `json:"duplicate_rate"`
}

// AIMergeDuplicatesRequest represents an AI-powered request to merge duplicates
type AIMergeDuplicatesRequest struct {
	PrimaryID   int   `json:"primary_id"`
	DuplicateIDs []int `json:"duplicate_ids"`
	MergeStrategy string `json:"merge_strategy"` // "keep_primary", "merge_tags", "merge_all"
}

// checkDuplicates handles POST /api/ai/duplicates/check
func (h *AIDuplicatesHandler) checkDuplicates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AICheckDuplicatesRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		h.writeError(w, "URL is required", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Create target bookmark for comparison
	target := &models.Bookmark{
		URL:   req.URL,
		Title: req.Title,
		Description: req.Description,
	}

	// Get all user bookmarks for comparison
	allBookmarks, err := h.bookmarkRepo.List(1, 10000, "", "", false, userID)
	if err != nil {
		h.writeError(w, "Failed to get bookmarks", http.StatusInternalServerError)
		return
	}

	// Find similar bookmarks
	matches := h.similarityEngine.FindSimilarBookmarks(target, allBookmarks.Bookmarks)

	// Generate recommendations
	recommendations := h.generateDuplicateRecommendations(matches)

	// Categorize matches
	exactMatches := 0
	similarMatches := 0
	for _, match := range matches {
		switch match.MatchType {
		case "exact", "near_duplicate":
			exactMatches++
		case "similar":
			similarMatches++
		}
	}

	response := DuplicatesResponse{
		HasDuplicates:   len(matches) > 0,
		Matches:         matches,
		Total:           len(matches),
		ExactMatches:    exactMatches,
		SimilarMatches:  similarMatches,
		Recommendations: recommendations,
	}

	h.writeJSON(w, response)
}

// findAllDuplicates handles GET /api/ai/duplicates/find-all
func (h *AIDuplicatesHandler) findAllDuplicates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Get all user bookmarks
	allBookmarks, err := h.bookmarkRepo.List(1, 10000, "", "", false, userID)
	if err != nil {
		h.writeError(w, "Failed to get bookmarks", http.StatusInternalServerError)
		return
	}

	bookmarks := allBookmarks.Bookmarks
	if len(bookmarks) == 0 {
		response := FindAllDuplicatesResponse{
			DuplicateGroups: []DuplicateGroup{},
			TotalGroups:     0,
			TotalBookmarks:  0,
			Statistics:      DuplicateStats{},
		}
		h.writeJSON(w, response)
		return
	}

	// Find duplicate groups
	duplicateGroups := h.findDuplicateGroups(bookmarks)

	// Calculate statistics
	stats := h.calculateDuplicateStatistics(bookmarks, duplicateGroups)

	response := FindAllDuplicatesResponse{
		DuplicateGroups: duplicateGroups,
		TotalGroups:     len(duplicateGroups),
		TotalBookmarks:  len(bookmarks),
		Statistics:      stats,
	}

	h.writeJSON(w, response)
}

// mergeDuplicates handles POST /api/ai/duplicates/merge
func (h *AIDuplicatesHandler) mergeDuplicates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AIMergeDuplicatesRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.PrimaryID == 0 || len(req.DuplicateIDs) == 0 {
		h.writeError(w, "Primary ID and duplicate IDs are required", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Get primary bookmark
	primary, err := h.bookmarkRepo.GetByID(req.PrimaryID, userID)
	if err != nil {
		h.writeError(w, "Primary bookmark not found", http.StatusNotFound)
		return
	}

	// Get duplicate bookmarks
	var duplicates []models.Bookmark
	for _, id := range req.DuplicateIDs {
		if duplicate, err := h.bookmarkRepo.GetByID(id, userID); err == nil {
			duplicates = append(duplicates, *duplicate)
		}
	}

	// Perform merge based on strategy
	mergedBookmark, err := h.performMerge(primary, duplicates, req.MergeStrategy, userID)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to merge duplicates: %v", err), http.StatusInternalServerError)
		return
	}

	// Delete duplicate bookmarks
	deletedCount := 0
	for _, id := range req.DuplicateIDs {
		if err := h.bookmarkRepo.Delete(id, userID); err == nil {
			deletedCount++
		}
	}

	response := map[string]interface{}{
		"success":         true,
		"merged_bookmark": mergedBookmark,
		"deleted_count":   deletedCount,
		"merge_strategy":  req.MergeStrategy,
	}

	h.writeJSON(w, response)
}

// analyzeDuplicates handles GET /api/ai/duplicates/analyze
func (h *AIDuplicatesHandler) analyzeDuplicates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get bookmark ID from query parameter
	bookmarkIDStr := r.URL.Query().Get("bookmark_id")
	if bookmarkIDStr == "" {
		h.writeError(w, "bookmark_id parameter is required", http.StatusBadRequest)
		return
	}

	bookmarkID, err := strconv.Atoi(bookmarkIDStr)
	if err != nil {
		h.writeError(w, "Invalid bookmark ID", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := h.getUserID(r)

	// Get the target bookmark
	target, err := h.bookmarkRepo.GetByID(bookmarkID, userID)
	if err != nil {
		h.writeError(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	// Get all user bookmarks for comparison
	allBookmarks, err := h.bookmarkRepo.List(1, 10000, "", "", false, userID)
	if err != nil {
		h.writeError(w, "Failed to get bookmarks", http.StatusInternalServerError)
		return
	}

	// Find similar bookmarks with detailed analysis
	matches := h.similarityEngine.FindSimilarBookmarks(target, allBookmarks.Bookmarks)

	// Add detailed analysis
	analysis := map[string]interface{}{
		"target_bookmark":  target,
		"potential_matches": matches,
		"analysis": map[string]interface{}{
			"url_normalized":    h.similarityEngine.NormalizeURL(target.URL),
			"has_exact_matches": h.hasExactMatches(matches),
			"similarity_breakdown": h.analyzeSimilarityBreakdown(matches),
			"recommendations": h.generateDetailedRecommendations(target, matches),
		},
	}

	h.writeJSON(w, analysis)
}

// Helper methods

func (h *AIDuplicatesHandler) findDuplicateGroups(bookmarks []models.Bookmark) []DuplicateGroup {
	var groups []DuplicateGroup
	processed := make(map[int]bool)
	groupID := 1

	for i, bookmark := range bookmarks {
		if processed[bookmark.ID] {
			continue
		}

		// Find similar bookmarks for this one
		matches := h.similarityEngine.FindSimilarBookmarks(&bookmark, bookmarks)
		
		// Filter for high-confidence duplicates
		var duplicates []ai.DuplicateMatch
		for _, match := range matches {
			if match.SimilarityScore >= 0.7 && !processed[match.Bookmark.ID] {
				duplicates = append(duplicates, match)
				processed[match.Bookmark.ID] = true
			}
		}

		if len(duplicates) > 0 {
			// Calculate group score
			groupScore := 0.0
			for _, dup := range duplicates {
				groupScore += dup.SimilarityScore
			}
			groupScore /= float64(len(duplicates))

			group := DuplicateGroup{
				ID:         groupID,
				Primary:    &bookmarks[i],
				Duplicates: duplicates,
				GroupScore: groupScore,
				GroupType:  h.determineGroupType(duplicates),
				Confidence: h.calculateGroupConfidence(duplicates),
			}

			groups = append(groups, group)
			processed[bookmark.ID] = true
			groupID++
		}
	}

	return groups
}

func (h *AIDuplicatesHandler) calculateDuplicateStatistics(bookmarks []models.Bookmark, groups []DuplicateGroup) DuplicateStats {
	duplicateCount := 0
	exactDuplicates := 0
	similarDuplicates := 0
	domainCounts := make(map[string]int)

	for _, group := range groups {
		duplicateCount += len(group.Duplicates)
		
		for _, duplicate := range group.Duplicates {
			switch duplicate.MatchType {
			case "exact", "near_duplicate":
				exactDuplicates++
			case "similar":
				similarDuplicates++
			}

			// Extract domain
			if domain := h.extractDomain(duplicate.Bookmark.URL); domain != "" {
				domainCounts[domain]++
			}
		}
	}

	// Calculate top domains
	var topDomains []DomainDuplicateInfo
	for domain, count := range domainCounts {
		topDomains = append(topDomains, DomainDuplicateInfo{
			Domain:        domain,
			Count:         count,
			DuplicateRate: float64(count) / float64(len(bookmarks)),
		})
	}

	// Sort by count
	for i := 0; i < len(topDomains)-1; i++ {
		for j := i + 1; j < len(topDomains); j++ {
			if topDomains[j].Count > topDomains[i].Count {
				topDomains[i], topDomains[j] = topDomains[j], topDomains[i]
			}
		}
	}

	// Keep top 10
	if len(topDomains) > 10 {
		topDomains = topDomains[:10]
	}

	duplicateRate := 0.0
	if len(bookmarks) > 0 {
		duplicateRate = float64(duplicateCount) / float64(len(bookmarks))
	}

	return DuplicateStats{
		TotalBookmarks:    len(bookmarks),
		DuplicateCount:    duplicateCount,
		DuplicateRate:     duplicateRate,
		ExactDuplicates:   exactDuplicates,
		SimilarDuplicates: similarDuplicates,
		TopDomains:        topDomains,
	}
}

func (h *AIDuplicatesHandler) performMerge(primary *models.Bookmark, duplicates []models.Bookmark, strategy string, userID int) (*models.Bookmark, error) {
	updateReq := &models.UpdateBookmarkRequest{}

	switch strategy {
	case "keep_primary":
		// No changes to primary
		return primary, nil
		
	case "merge_tags":
		// Collect all unique tags
		tagSet := make(map[string]bool)
		for _, tag := range primary.Tags {
			tagSet[tag.Name] = true
		}
		
		for _, dup := range duplicates {
			for _, tag := range dup.Tags {
				tagSet[tag.Name] = true
			}
		}
		
		var mergedTags []string
		for tag := range tagSet {
			mergedTags = append(mergedTags, tag)
		}
		
		updateReq.Tags = mergedTags
		
	case "merge_all":
		// Merge everything: use best title, merge descriptions, merge tags
		bestTitle := h.chooseBestTitle(primary, duplicates)
		if bestTitle != primary.Title {
			updateReq.Title = &bestTitle
		}
		
		mergedDescription := h.mergeDescriptions(primary, duplicates)
		if mergedDescription != "" && (primary.Description == nil || *primary.Description != mergedDescription) {
			updateReq.Description = &mergedDescription
		}
		
		// Merge tags (same as merge_tags)
		tagSet := make(map[string]bool)
		for _, tag := range primary.Tags {
			tagSet[tag.Name] = true
		}
		
		for _, dup := range duplicates {
			for _, tag := range dup.Tags {
				tagSet[tag.Name] = true
			}
		}
		
		var mergedTags []string
		for tag := range tagSet {
			mergedTags = append(mergedTags, tag)
		}
		
		updateReq.Tags = mergedTags
	}

	// Update primary bookmark
	return h.bookmarkRepo.Update(primary.ID, updateReq, userID)
}

func (h *AIDuplicatesHandler) generateDuplicateRecommendations(matches []ai.DuplicateMatch) []string {
	var recommendations []string

	if len(matches) == 0 {
		recommendations = append(recommendations, "No duplicates found. You can safely add this bookmark.")
		return recommendations
	}

	exactCount := 0
	similarCount := 0
	for _, match := range matches {
		switch match.MatchType {
		case "exact":
			exactCount++
		case "near_duplicate":
			exactCount++
		case "similar":
			similarCount++
		}
	}

	if exactCount > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Found %d exact duplicate(s). Consider not adding this bookmark.", exactCount))
	}

	if similarCount > 0 {
		recommendations = append(recommendations, fmt.Sprintf("Found %d similar bookmark(s). Review for potential duplicates.", similarCount))
	}

	if exactCount > 0 || similarCount > 2 {
		recommendations = append(recommendations, "Consider using the merge feature to consolidate similar bookmarks.")
	}

	return recommendations
}

func (h *AIDuplicatesHandler) determineGroupType(duplicates []ai.DuplicateMatch) string {
	exactCount := 0
	for _, dup := range duplicates {
		if dup.MatchType == "exact" || dup.MatchType == "near_duplicate" {
			exactCount++
		}
	}

	if exactCount == len(duplicates) {
		return "exact_duplicates"
	} else if exactCount > 0 {
		return "mixed_duplicates"
	}
	return "similar_content"
}

func (h *AIDuplicatesHandler) calculateGroupConfidence(duplicates []ai.DuplicateMatch) float64 {
	if len(duplicates) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, dup := range duplicates {
		totalConfidence += dup.Confidence
	}

	return totalConfidence / float64(len(duplicates))
}

func (h *AIDuplicatesHandler) hasExactMatches(matches []ai.DuplicateMatch) bool {
	for _, match := range matches {
		if match.MatchType == "exact" {
			return true
		}
	}
	return false
}

func (h *AIDuplicatesHandler) analyzeSimilarityBreakdown(matches []ai.DuplicateMatch) map[string]interface{} {
	if len(matches) == 0 {
		return map[string]interface{}{}
	}

	avgURL := 0.0
	avgTitle := 0.0
	avgContent := 0.0

	for _, match := range matches {
		avgURL += match.URLSimilarity
		avgTitle += match.TitleSimilarity
		avgContent += match.ContentSimilarity
	}

	count := float64(len(matches))
	return map[string]interface{}{
		"average_url_similarity":     avgURL / count,
		"average_title_similarity":   avgTitle / count,
		"average_content_similarity": avgContent / count,
		"total_matches":              len(matches),
	}
}

func (h *AIDuplicatesHandler) generateDetailedRecommendations(target *models.Bookmark, matches []ai.DuplicateMatch) []string {
	var recommendations []string
	
	for _, match := range matches {
		switch match.MatchType {
		case "exact":
			recommendations = append(recommendations, 
				fmt.Sprintf("Exact duplicate found: '%s' (%.1f%% match)", match.Bookmark.Title, match.SimilarityScore*100))
		case "near_duplicate":
			recommendations = append(recommendations, 
				fmt.Sprintf("Near duplicate found: '%s' (%.1f%% match)", match.Bookmark.Title, match.SimilarityScore*100))
		case "similar":
			recommendations = append(recommendations, 
				fmt.Sprintf("Similar content found: '%s' (%.1f%% match)", match.Bookmark.Title, match.SimilarityScore*100))
		}
	}

	return recommendations
}

func (h *AIDuplicatesHandler) chooseBestTitle(primary *models.Bookmark, duplicates []models.Bookmark) string {
	bestTitle := primary.Title
	maxLength := len(primary.Title)

	// Choose the longest, most descriptive title
	for _, dup := range duplicates {
		if len(dup.Title) > maxLength && dup.Title != dup.URL {
			bestTitle = dup.Title
			maxLength = len(dup.Title)
		}
	}

	return bestTitle
}

func (h *AIDuplicatesHandler) mergeDescriptions(primary *models.Bookmark, duplicates []models.Bookmark) string {
	descriptions := []string{}
	
	if primary.Description != nil && *primary.Description != "" {
		descriptions = append(descriptions, *primary.Description)
	}
	
	for _, dup := range duplicates {
		if dup.Description != nil && *dup.Description != "" {
			// Check if this description is already included
			found := false
			for _, existing := range descriptions {
				if existing == *dup.Description {
					found = true
					break
				}
			}
			if !found {
				descriptions = append(descriptions, *dup.Description)
			}
		}
	}

	return strings.Join(descriptions, " | ")
}

func (h *AIDuplicatesHandler) extractDomain(url string) string {
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

func (h *AIDuplicatesHandler) getUserID(r *http.Request) int {
	if userID, ok := middleware.GetUserIDFromContext(r); ok {
		return userID
	}
	return 1 // Default for backward compatibility
}

func (h *AIDuplicatesHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.MarshalWrite(w, data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func (h *AIDuplicatesHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}
	
	json.MarshalWrite(w, response)
}