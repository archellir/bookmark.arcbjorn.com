package services

import (
	"fmt"
	"strings"
	"time"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

// DuplicateService handles bookmark duplicate detection
type DuplicateService struct {
	bookmarkRepo *db.BookmarkRepository
	urlService   *URLService
}

// NewDuplicateService creates a new duplicate detection service
func NewDuplicateService(bookmarkRepo *db.BookmarkRepository) *DuplicateService {
	return &DuplicateService{
		bookmarkRepo: bookmarkRepo,
		urlService:   NewURLService(),
	}
}

// DuplicateCheckResult contains information about potential duplicates
type DuplicateCheckResult struct {
	HasExactDuplicate    bool                    `json:"has_exact_duplicate"`
	ExactDuplicate       *models.Bookmark        `json:"exact_duplicate,omitempty"`
	HasSimilarBookmarks  bool                    `json:"has_similar_bookmarks"`
	SimilarBookmarks     []models.Bookmark       `json:"similar_bookmarks,omitempty"`
	URLAnalysis          *URLNormalizationResult `json:"url_analysis"`
	Confidence           float64                 `json:"confidence"` // 0-1, how confident we are about duplicates
	Recommendations      []string                `json:"recommendations"`
}

// CheckForDuplicates analyzes a URL for potential duplicates
func (s *DuplicateService) CheckForDuplicates(url, title string) (*DuplicateCheckResult, error) {
	result := &DuplicateCheckResult{
		SimilarBookmarks: make([]models.Bookmark, 0),
		Recommendations:  make([]string, 0),
	}

	// Analyze the URL
	urlAnalysis, err := s.urlService.NormalizeURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze URL: %w", err)
	}
	result.URLAnalysis = urlAnalysis

	// Check for exact duplicates first
	exactDuplicate, err := s.findExactDuplicate(url)
	if err == nil && exactDuplicate != nil {
		result.HasExactDuplicate = true
		result.ExactDuplicate = exactDuplicate
		result.Confidence = 1.0
		result.Recommendations = append(result.Recommendations, 
			"This URL already exists in your bookmarks")
		return result, nil
	}

	// Get all bookmarks to check for similarities
	allBookmarks, err := s.getAllBookmarkURLs()
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarks: %w", err)
	}

	// Find similar URLs
	similarURLs := s.urlService.FindSimilarURLs(url, allBookmarks)
	
	if len(similarURLs) > 0 {
		result.HasSimilarBookmarks = true
		
		// Get full bookmark details for similar URLs
		for _, similarURL := range similarURLs {
			bookmark, err := s.bookmarkRepo.GetByURL(similarURL)
			if err == nil {
				result.SimilarBookmarks = append(result.SimilarBookmarks, *bookmark)
			}
		}
		
		// Calculate confidence based on similarity
		result.Confidence = s.calculateSimilarityConfidence(url, title, result.SimilarBookmarks)
		
		// Generate recommendations
		result.Recommendations = s.generateRecommendations(urlAnalysis, result.SimilarBookmarks)
	}

	return result, nil
}

// FindAllDuplicates finds all duplicate bookmarks in the database
func (s *DuplicateService) FindAllDuplicates() ([]DuplicateGroup, error) {
	// Get all bookmarks
	response, err := s.bookmarkRepo.List(1, 10000, "", "", false) // Get all
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarks: %w", err)
	}

	bookmarks := response.Bookmarks
	duplicateGroups := make([]DuplicateGroup, 0)
	processed := make(map[int]bool)

	for _, bookmark := range bookmarks {
		if processed[bookmark.ID] {
			continue
		}

		// Find duplicates for this bookmark
		duplicates := make([]models.Bookmark, 0)
		
		for _, other := range bookmarks {
			if other.ID == bookmark.ID || processed[other.ID] {
				continue
			}

			if s.areDuplicates(bookmark, other) {
				duplicates = append(duplicates, other)
				processed[other.ID] = true
			}
		}

		if len(duplicates) > 0 {
			group := DuplicateGroup{
				Primary:    bookmark,
				Duplicates: duplicates,
				Confidence: s.calculateGroupConfidence(bookmark, duplicates),
				Reason:     s.getDuplicateReason(bookmark, duplicates[0]),
			}
			duplicateGroups = append(duplicateGroups, group)
			processed[bookmark.ID] = true
		}
	}

	return duplicateGroups, nil
}

// DuplicateGroup represents a group of duplicate bookmarks
type DuplicateGroup struct {
	Primary    models.Bookmark   `json:"primary"`
	Duplicates []models.Bookmark `json:"duplicates"`
	Confidence float64           `json:"confidence"`
	Reason     string            `json:"reason"`
}

// MergeDuplicates merges duplicate bookmarks by keeping the primary and removing duplicates
func (s *DuplicateService) MergeDuplicates(primaryID int, duplicateIDs []int, mergeTags bool, mergeMetadata bool) error {
	// Get primary bookmark
	primary, err := s.bookmarkRepo.GetByID(primaryID)
	if err != nil {
		return fmt.Errorf("failed to get primary bookmark: %w", err)
	}

	// Get duplicates
	var duplicates []*models.Bookmark
	for _, id := range duplicateIDs {
		duplicate, err := s.bookmarkRepo.GetByID(id)
		if err != nil {
			continue // Skip if not found
		}
		duplicates = append(duplicates, duplicate)
	}

	if len(duplicates) == 0 {
		return fmt.Errorf("no valid duplicates found")
	}

	// Prepare update data
	updateReq := &models.UpdateBookmarkRequest{}

	if mergeTags {
		// Merge tags from all duplicates
		allTags := make(map[string]bool)
		
		// Add primary tags
		for _, tag := range primary.Tags {
			allTags[tag.Name] = true
		}
		
		// Add duplicate tags
		for _, duplicate := range duplicates {
			for _, tag := range duplicate.Tags {
				allTags[tag.Name] = true
			}
		}
		
		// Convert to slice
		mergedTags := make([]string, 0, len(allTags))
		for tag := range allTags {
			mergedTags = append(mergedTags, tag)
		}
		updateReq.Tags = mergedTags
	}

	if mergeMetadata {
		// Use the most complete title and description
		bestTitle := primary.Title
		var bestDescription *string = primary.Description

		for _, duplicate := range duplicates {
			// Use longer, more descriptive title
			if len(duplicate.Title) > len(bestTitle) && 
			   !strings.Contains(strings.ToLower(duplicate.Title), "untitled") {
				bestTitle = duplicate.Title
			}
			
			// Use non-empty description if primary doesn't have one
			if (bestDescription == nil || *bestDescription == "") && 
			   duplicate.Description != nil && *duplicate.Description != "" {
				bestDescription = duplicate.Description
			}
		}

		updateReq.Title = &bestTitle
		updateReq.Description = bestDescription
	}

	// Update primary bookmark if we have changes
	if updateReq.Title != nil || updateReq.Description != nil || updateReq.Tags != nil {
		_, err = s.bookmarkRepo.Update(primaryID, updateReq)
		if err != nil {
			return fmt.Errorf("failed to update primary bookmark: %w", err)
		}
	}

	// Delete duplicates
	for _, id := range duplicateIDs {
		err = s.bookmarkRepo.Delete(id)
		if err != nil {
			// Log but don't fail the whole operation
			fmt.Printf("Warning: failed to delete duplicate %d: %v\n", id, err)
		}
	}

	return nil
}

// Helper methods

func (s *DuplicateService) findExactDuplicate(url string) (*models.Bookmark, error) {
	return s.bookmarkRepo.GetByURL(url)
}

func (s *DuplicateService) getAllBookmarkURLs() ([]string, error) {
	response, err := s.bookmarkRepo.List(1, 10000, "", "", false)
	if err != nil {
		return nil, err
	}

	urls := make([]string, len(response.Bookmarks))
	for i, bookmark := range response.Bookmarks {
		urls[i] = bookmark.URL
	}

	return urls, nil
}

func (s *DuplicateService) areDuplicates(bookmark1, bookmark2 models.Bookmark) bool {
	// Check URL similarity
	similar := s.urlService.FindSimilarURLs(bookmark1.URL, []string{bookmark2.URL})
	if len(similar) > 0 {
		return true
	}

	// Check title similarity (fuzzy match)
	if s.isSimilarTitle(bookmark1.Title, bookmark2.Title) {
		return true
	}

	return false
}

func (s *DuplicateService) isSimilarTitle(title1, title2 string) bool {
	// Simple title similarity check
	t1 := strings.ToLower(strings.TrimSpace(title1))
	t2 := strings.ToLower(strings.TrimSpace(title2))

	// Exact match
	if t1 == t2 {
		return true
	}

	// One is substring of the other (with reasonable length)
	if len(t1) > 10 && len(t2) > 10 {
		if strings.Contains(t1, t2) || strings.Contains(t2, t1) {
			return true
		}
	}

	return false
}

func (s *DuplicateService) calculateSimilarityConfidence(url, title string, similar []models.Bookmark) float64 {
	if len(similar) == 0 {
		return 0.0
	}

	maxConfidence := 0.0

	for _, bookmark := range similar {
		confidence := 0.0

		// URL similarity weight (70%)
		urlSimilarity := s.calculateURLSimilarity(url, bookmark.URL)
		confidence += urlSimilarity * 0.7

		// Title similarity weight (30%)
		titleSimilarity := s.calculateTitleSimilarity(title, bookmark.Title)
		confidence += titleSimilarity * 0.3

		if confidence > maxConfidence {
			maxConfidence = confidence
		}
	}

	return maxConfidence
}

func (s *DuplicateService) calculateURLSimilarity(url1, url2 string) float64 {
	result1, err1 := s.urlService.NormalizeURL(url1)
	result2, err2 := s.urlService.NormalizeURL(url2)

	if err1 != nil || err2 != nil {
		return 0.0
	}

	// Exact normalized match
	if result1.Normalized == result2.Normalized {
		return 1.0
	}

	// Check variations
	for _, var1 := range result1.Variations {
		for _, var2 := range result2.Variations {
			if var1 == var2 {
				return 0.9
			}
		}
	}

	return 0.0
}

func (s *DuplicateService) calculateTitleSimilarity(title1, title2 string) float64 {
	t1 := strings.ToLower(strings.TrimSpace(title1))
	t2 := strings.ToLower(strings.TrimSpace(title2))

	if t1 == t2 {
		return 1.0
	}

	if strings.Contains(t1, t2) || strings.Contains(t2, t1) {
		return 0.8
	}

	return 0.0
}

func (s *DuplicateService) calculateGroupConfidence(primary models.Bookmark, duplicates []models.Bookmark) float64 {
	totalConfidence := 0.0

	for _, duplicate := range duplicates {
		confidence := s.calculateSimilarityConfidence(primary.URL, primary.Title, []models.Bookmark{duplicate})
		totalConfidence += confidence
	}

	return totalConfidence / float64(len(duplicates))
}

func (s *DuplicateService) getDuplicateReason(bookmark1, bookmark2 models.Bookmark) string {
	// Check URL similarity first
	similar := s.urlService.FindSimilarURLs(bookmark1.URL, []string{bookmark2.URL})
	if len(similar) > 0 {
		urlResult1, _ := s.urlService.NormalizeURL(bookmark1.URL)
		urlResult2, _ := s.urlService.NormalizeURL(bookmark2.URL)
		
		if urlResult1 != nil && urlResult2 != nil {
			if urlResult1.Normalized == urlResult2.Normalized {
				return "Identical URLs (after normalization)"
			}
			
			if urlResult1.IsShortURL || urlResult2.IsShortURL {
				return "Short URL pointing to same destination"
			}
			
			return "Similar URLs (different protocols/www/trailing slash)"
		}
	}

	// Check title similarity
	if s.isSimilarTitle(bookmark1.Title, bookmark2.Title) {
		return "Similar titles"
	}

	return "Potential duplicate"
}

func (s *DuplicateService) generateRecommendations(urlAnalysis *URLNormalizationResult, similar []models.Bookmark) []string {
	recommendations := make([]string, 0)

	if urlAnalysis.IsShortURL && urlAnalysis.ExpandedURL != "" {
		recommendations = append(recommendations, 
			fmt.Sprintf("This short URL expands to: %s", urlAnalysis.ExpandedURL))
	}

	if len(similar) > 0 {
		recommendations = append(recommendations, 
			fmt.Sprintf("Found %d similar bookmark(s) that might be duplicates", len(similar)))
		
		for _, bookmark := range similar {
			age := time.Since(bookmark.CreatedAt)
			recommendations = append(recommendations, 
				fmt.Sprintf("Similar: \"%s\" (created %s ago)", 
					bookmark.Title, s.formatDuration(age)))
		}
	}

	return recommendations
}

func (s *DuplicateService) formatDuration(d time.Duration) string {
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	} else {
		return fmt.Sprintf("%d days", int(d.Hours()/24))
	}
}