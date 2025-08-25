package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"torimemo/internal/ai"
	"torimemo/internal/db"
	"torimemo/internal/models"
)

// AIClusteringHandler handles AI-powered bookmark clustering
type AIClusteringHandler struct {
	bookmarkRepo *db.BookmarkRepository
	clusterer    *ai.BookmarkClusterer
}

// NewAIClusteringHandler creates a new clustering handler
func NewAIClusteringHandler(bookmarkRepo *db.BookmarkRepository) *AIClusteringHandler {
	return &AIClusteringHandler{
		bookmarkRepo: bookmarkRepo,
		clusterer:    ai.NewBookmarkClusterer(),
	}
}

// RegisterRoutes registers clustering routes
func (h *AIClusteringHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/ai/cluster", h.clusterBookmarksHandler)
	mux.HandleFunc("/api/ai/cluster/analyze", h.analyzeClusterPotential)
	mux.HandleFunc("/api/ai/cluster/preview", h.previewClustering)
	mux.HandleFunc("/api/ai/cluster/suggestions", h.getClusterSuggestions)
}

// clusterBookmarksHandler handles POST /api/ai/cluster
func (h *AIClusteringHandler) clusterBookmarksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ai.ClusteringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	req.UserID = userID

	// Apply defaults
	if req.MinClusterSize == 0 {
		req.MinClusterSize = 3
	}
	if req.MaxClusters == 0 {
		req.MaxClusters = 20
	}
	if req.SimilarityThreshold == 0 {
		req.SimilarityThreshold = 0.4
	}
	if req.ClusteringMethod == "" {
		req.ClusteringMethod = "hybrid"
	}

	// Get bookmarks to cluster
	var bookmarks []models.Bookmark
	var err error

	if len(req.BookmarkIDs) > 0 {
		// Cluster specific bookmarks
		bookmarks, err = h.getBookmarksByIDs(req.BookmarkIDs, userID)
	} else {
		// Cluster all user bookmarks
		response, err := h.bookmarkRepo.List(1, 10000, "", "", false, userID) // High limit to get all
		if err == nil {
			bookmarks = response.Bookmarks
		}
	}

	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get bookmarks: %v", err), http.StatusInternalServerError)
		return
	}

	if len(bookmarks) < req.MinClusterSize {
		h.writeError(w, "Insufficient bookmarks for clustering", http.StatusBadRequest)
		return
	}

	// Perform clustering
	result, err := h.clusterer.ClusterBookmarksWithConfig(bookmarks, ai.ClusteringConfig{
		Method:              req.ClusteringMethod,
		MinClusterSize:      req.MinClusterSize,
		MaxClusters:         req.MaxClusters,
		SimilarityThreshold: req.SimilarityThreshold,
	})

	if err != nil {
		h.writeError(w, fmt.Sprintf("Clustering failed: %v", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, result)
}

// analyzeClusterPotential handles GET /api/ai/cluster/analyze
func (h *AIClusteringHandler) analyzeClusterPotential(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all user bookmarks
	response, err := h.bookmarkRepo.List(1, 10000, "", "", false, userID)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get bookmarks: %v", err), http.StatusInternalServerError)
		return
	}
	bookmarks := response.Bookmarks

	analysis := h.analyzeBookmarksForClustering(bookmarks)
	
	analysisResponse := map[string]interface{}{
		"total_bookmarks":    len(bookmarks),
		"clustering_potential": analysis,
		"recommended_method": h.recommendClusteringMethod(analysis),
		"estimated_clusters": h.estimateClusterCount(len(bookmarks), analysis),
	}

	h.writeJSON(w, analysisResponse)
}

// previewClustering handles POST /api/ai/cluster/preview
func (h *AIClusteringHandler) previewClustering(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ai.ClusteringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get sample of bookmarks for preview (max 50)
	response, err := h.bookmarkRepo.List(1, 50, "", "", false, userID)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get bookmarks: %v", err), http.StatusInternalServerError)
		return
	}
	bookmarks := response.Bookmarks

	// Limit to 50 for preview
	if len(bookmarks) > 50 {
		bookmarks = bookmarks[:50]
	}

	if len(bookmarks) < 3 {
		h.writeError(w, "Insufficient bookmarks for preview", http.StatusBadRequest)
		return
	}

	// Apply defaults for preview
	if req.ClusteringMethod == "" {
		req.ClusteringMethod = "hybrid"
	}

	// Quick clustering with smaller parameters
	result, err := h.clusterer.ClusterBookmarksWithConfig(bookmarks, ai.ClusteringConfig{
		Method:              req.ClusteringMethod,
		MinClusterSize:      2, // Lower for preview
		MaxClusters:         10, // Less for preview
		SimilarityThreshold: 0.3, // Lower for preview
	})

	if err != nil {
		h.writeError(w, fmt.Sprintf("Preview clustering failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return simplified preview
	preview := map[string]interface{}{
		"preview":           true,
		"sample_size":       len(bookmarks),
		"total_available":   len(bookmarks), // This would be total if we had it
		"clusters":          result.Clusters,
		"estimated_quality": result.QualityScore,
		"method_used":       result.Method,
	}

	h.writeJSON(w, preview)
}

// getClusterSuggestions handles GET /api/ai/cluster/suggestions
func (h *AIClusteringHandler) getClusterSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := h.getUserID(r)
	if userID == 0 {
		h.writeError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	response, err := h.bookmarkRepo.List(1, 10000, "", "", false, userID)
	if err != nil {
		h.writeError(w, fmt.Sprintf("Failed to get bookmarks: %v", err), http.StatusInternalServerError)
		return
	}
	bookmarks := response.Bookmarks

	suggestions := h.generateClusterSuggestions(bookmarks)
	
	suggestionsResponse := map[string]interface{}{
		"suggestions":     suggestions,
		"total_bookmarks": len(bookmarks),
		"actionable":      len(suggestions) > 0,
	}

	h.writeJSON(w, suggestionsResponse)
}

// Helper methods

func (h *AIClusteringHandler) getUserID(r *http.Request) int {
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

func (h *AIClusteringHandler) getBookmarksByIDs(bookmarkIDs []int, userID int) ([]models.Bookmark, error) {
	var bookmarks []models.Bookmark
	
	for _, bookmarkID := range bookmarkIDs {
		bookmark, err := h.bookmarkRepo.GetByID(bookmarkID, userID)
		if err == nil {
			bookmarks = append(bookmarks, *bookmark)
		}
	}
	
	return bookmarks, nil
}

func (h *AIClusteringHandler) analyzeBookmarksForClustering(bookmarks []models.Bookmark) map[string]interface{} {
	analysis := make(map[string]interface{})
	
	// Domain diversity analysis
	domainCount := make(map[string]int)
	for _, bookmark := range bookmarks {
		domain := h.extractDomain(bookmark.URL)
		domainCount[domain]++
	}
	
	analysis["unique_domains"] = len(domainCount)
	analysis["domain_diversity"] = float64(len(domainCount)) / float64(len(bookmarks))
	
	// Tag analysis
	tagCount := make(map[string]int)
	taggedBookmarks := 0
	for _, bookmark := range bookmarks {
		if len(bookmark.Tags) > 0 {
			taggedBookmarks++
		}
		for _, tag := range bookmark.Tags {
			tagCount[tag.Name]++
		}
	}
	
	analysis["unique_tags"] = len(tagCount)
	analysis["tag_coverage"] = float64(taggedBookmarks) / float64(len(bookmarks))
	analysis["avg_tags_per_bookmark"] = float64(len(tagCount)) / float64(len(bookmarks))
	
	// Large domain clusters
	largeDomainClusters := 0
	for _, count := range domainCount {
		if count >= 5 {
			largeDomainClusters++
		}
	}
	analysis["large_domain_clusters"] = largeDomainClusters
	
	// Clustering potential score
	domainScore := math.Min(1.0, float64(len(domainCount))/10.0) // Good if 10+ unique domains
	tagScore := analysis["tag_coverage"].(float64)
	diversityScore := math.Min(1.0, analysis["domain_diversity"].(float64)*2.0)
	
	analysis["clustering_score"] = (domainScore + tagScore + diversityScore) / 3.0
	
	return analysis
}

func (h *AIClusteringHandler) recommendClusteringMethod(analysis map[string]interface{}) string {
	clusteringScore := analysis["clustering_score"].(float64)
	domainDiversity := analysis["domain_diversity"].(float64)
	tagCoverage := analysis["tag_coverage"].(float64)
	
	// High tag coverage and good diversity = semantic clustering
	if tagCoverage > 0.7 && domainDiversity > 0.3 {
		return "semantic"
	}
	
	// Low diversity but large collections = domain clustering  
	if domainDiversity < 0.2 && clusteringScore > 0.5 {
		return "domain"
	}
	
	// Default to hybrid for balanced collections
	return "hybrid"
}

func (h *AIClusteringHandler) estimateClusterCount(totalBookmarks int, analysis map[string]interface{}) int {
	domainCount := analysis["unique_domains"].(int)
	clusteringScore := analysis["clustering_score"].(float64)
	
	// Base estimate on domain count and total bookmarks
	baseEstimate := int(float64(totalBookmarks) / 8.0) // ~8 bookmarks per cluster
	domainBasedEstimate := domainCount / 2             // ~2 domains per cluster
	
	// Use the smaller estimate, modified by clustering score
	estimate := int(math.Min(float64(baseEstimate), float64(domainBasedEstimate)) * clusteringScore)
	
	// Bound the estimate
	if estimate < 2 {
		estimate = 2
	}
	if estimate > 25 {
		estimate = 25
	}
	
	return estimate
}

func (h *AIClusteringHandler) generateClusterSuggestions(bookmarks []models.Bookmark) []map[string]interface{} {
	var suggestions []map[string]interface{}
	
	// Analyze for obvious clusters
	domainCount := make(map[string][]models.Bookmark)
	for _, bookmark := range bookmarks {
		domain := h.extractDomain(bookmark.URL)
		domainCount[domain] = append(domainCount[domain], bookmark)
	}
	
	// Suggest large domain clusters
	for domain, domainBookmarks := range domainCount {
		if len(domainBookmarks) >= 5 {
			suggestions = append(suggestions, map[string]interface{}{
				"type":        "domain_cluster",
				"name":        domain + " Collection",
				"description": fmt.Sprintf("Group %d bookmarks from %s", len(domainBookmarks), domain),
				"bookmark_count": len(domainBookmarks),
				"confidence": 0.9,
				"preview_bookmarks": h.getBookmarkPreview(domainBookmarks, 3),
			})
		}
	}
	
	// Suggest tag-based clusters
	tagCount := make(map[string][]models.Bookmark)
	for _, bookmark := range bookmarks {
		for _, tag := range bookmark.Tags {
			tagCount[tag.Name] = append(tagCount[tag.Name], bookmark)
		}
	}
	
	for tag, tagBookmarks := range tagCount {
		if len(tagBookmarks) >= 4 {
			suggestions = append(suggestions, map[string]interface{}{
				"type":        "tag_cluster", 
				"name":        strings.Title(tag) + " Resources",
				"description": fmt.Sprintf("Group %d bookmarks tagged with '%s'", len(tagBookmarks), tag),
				"bookmark_count": len(tagBookmarks),
				"confidence": 0.8,
				"preview_bookmarks": h.getBookmarkPreview(tagBookmarks, 3),
			})
		}
	}
	
	return suggestions
}

func (h *AIClusteringHandler) getBookmarkPreview(bookmarks []models.Bookmark, count int) []map[string]string {
	var preview []map[string]string
	
	limit := count
	if len(bookmarks) < limit {
		limit = len(bookmarks)
	}
	
	for i := 0; i < limit; i++ {
		preview = append(preview, map[string]string{
			"title": bookmarks[i].Title,
			"url":   bookmarks[i].URL,
		})
	}
	
	return preview
}

func (h *AIClusteringHandler) extractDomain(url string) string {
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

func (h *AIClusteringHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

func (h *AIClusteringHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error":  message,
		"status": statusCode,
	}
	
	json.NewEncoder(w).Encode(response)
}