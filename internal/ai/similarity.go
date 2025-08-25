package ai

import (
	"math"
	"net/url"
	"sort"
	"strings"
	"unicode"

	"torimemo/internal/models"
)

// SimilarityEngine handles content and URL similarity analysis
type SimilarityEngine struct {
	// Configurable thresholds
	URLSimilarityThreshold     float64
	TitleSimilarityThreshold   float64
	ContentSimilarityThreshold float64
	OverallSimilarityThreshold float64
}

// NewSimilarityEngine creates a new similarity analysis engine
func NewSimilarityEngine() *SimilarityEngine {
	return &SimilarityEngine{
		URLSimilarityThreshold:     0.85,
		TitleSimilarityThreshold:   0.8,
		ContentSimilarityThreshold: 0.75,
		OverallSimilarityThreshold: 0.7,
	}
}

// DuplicateMatch represents a potential duplicate bookmark
type DuplicateMatch struct {
	Bookmark         *models.Bookmark `json:"bookmark"`
	SimilarityScore  float64          `json:"similarity_score"`
	URLSimilarity    float64          `json:"url_similarity"`
	TitleSimilarity  float64          `json:"title_similarity"`
	ContentSimilarity float64         `json:"content_similarity"`
	MatchType        string           `json:"match_type"` // "exact", "near_duplicate", "similar"
	Confidence       float64          `json:"confidence"`
}

// FindSimilarBookmarks finds bookmarks similar to the given bookmark
func (se *SimilarityEngine) FindSimilarBookmarks(target *models.Bookmark, candidates []models.Bookmark) []DuplicateMatch {
	var matches []DuplicateMatch

	for _, candidate := range candidates {
		if candidate.ID == target.ID {
			continue // Skip self-comparison
		}

		match := se.calculateSimilarity(target, &candidate)
		if match.SimilarityScore >= se.OverallSimilarityThreshold {
			matches = append(matches, match)
		}
	}

	// Sort by similarity score (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].SimilarityScore > matches[j].SimilarityScore
	})

	return matches
}

// calculateSimilarity computes comprehensive similarity between two bookmarks
func (se *SimilarityEngine) calculateSimilarity(a, b *models.Bookmark) DuplicateMatch {
	match := DuplicateMatch{
		Bookmark: b,
	}

	// URL similarity
	match.URLSimilarity = se.calculateURLSimilarity(a.URL, b.URL)

	// Title similarity
	match.TitleSimilarity = se.calculateTextSimilarity(a.Title, b.Title)

	// Content similarity (description)
	aDesc := ""
	if a.Description != nil {
		aDesc = *a.Description
	}
	bDesc := ""
	if b.Description != nil {
		bDesc = *b.Description
	}
	match.ContentSimilarity = se.calculateTextSimilarity(aDesc, bDesc)

	// Calculate weighted overall similarity
	match.SimilarityScore = se.calculateOverallSimilarity(match)

	// Determine match type and confidence
	match.MatchType, match.Confidence = se.determineMatchType(match)

	return match
}

// calculateURLSimilarity analyzes URL similarity using multiple techniques
func (se *SimilarityEngine) calculateURLSimilarity(url1, url2 string) float64 {
	// Exact match
	if url1 == url2 {
		return 1.0
	}

	// Normalize URLs for comparison
	norm1 := se.NormalizeURL(url1)
	norm2 := se.NormalizeURL(url2)

	if norm1 == norm2 {
		return 0.95
	}

	// Parse URLs for detailed comparison
	parsed1, err1 := url.Parse(norm1)
	parsed2, err2 := url.Parse(norm2)

	if err1 != nil || err2 != nil {
		// Fall back to string similarity if URL parsing fails
		return se.calculateStringSimilarity(url1, url2)
	}

	// Domain similarity
	domainSim := se.calculateDomainSimilarity(parsed1.Hostname(), parsed2.Hostname())

	// Path similarity
	pathSim := se.calculateStringSimilarity(parsed1.Path, parsed2.Path)

	// Query parameter similarity
	querySim := se.calculateQuerySimilarity(parsed1.RawQuery, parsed2.RawQuery)

	// Weighted combination
	urlSimilarity := (domainSim*0.4 + pathSim*0.4 + querySim*0.2)

	return math.Min(1.0, urlSimilarity)
}

// calculateTextSimilarity computes semantic similarity between text content
func (se *SimilarityEngine) calculateTextSimilarity(text1, text2 string) float64 {
	if text1 == "" && text2 == "" {
		return 1.0
	}
	if text1 == "" || text2 == "" {
		return 0.0
	}

	// Exact match
	if strings.EqualFold(text1, text2) {
		return 1.0
	}

	// Normalize texts
	norm1 := se.normalizeText(text1)
	norm2 := se.normalizeText(text2)

	// Jaccard similarity on words
	jaccardSim := se.calculateJaccardSimilarity(norm1, norm2)

	// Levenshtein-based similarity
	levenSim := se.calculateLevenshteinSimilarity(norm1, norm2)

	// N-gram similarity
	ngramSim := se.calculateNGramSimilarity(norm1, norm2, 3)

	// Weighted combination
	textSim := (jaccardSim*0.4 + levenSim*0.3 + ngramSim*0.3)

	return math.Min(1.0, textSim)
}

// calculateOverallSimilarity computes weighted overall similarity
func (se *SimilarityEngine) calculateOverallSimilarity(match DuplicateMatch) float64 {
	// Weights: URL is most important, then title, then content
	urlWeight := 0.5
	titleWeight := 0.3
	contentWeight := 0.2

	// Adjust weights based on content availability
	if match.ContentSimilarity == 0 {
		// No content to compare, redistribute weight
		urlWeight = 0.6
		titleWeight = 0.4
		contentWeight = 0.0
	}

	overall := (match.URLSimilarity*urlWeight +
		match.TitleSimilarity*titleWeight +
		match.ContentSimilarity*contentWeight)

	return math.Min(1.0, overall)
}

// determineMatchType categorizes the type of match based on similarity scores
func (se *SimilarityEngine) determineMatchType(match DuplicateMatch) (string, float64) {
	if match.URLSimilarity >= 0.95 && match.TitleSimilarity >= 0.9 {
		return "exact", 0.95
	} else if match.SimilarityScore >= 0.85 {
		return "near_duplicate", 0.8
	} else if match.SimilarityScore >= 0.7 {
		return "similar", 0.6
	}
	return "related", 0.3
}

// Helper methods for URL normalization and comparison

func (se *SimilarityEngine) NormalizeURL(rawURL string) string {
	// Remove common tracking parameters
	trackingParams := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "gclid", "ref", "source", "_ga", "_gl"}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(rawURL))
	}

	// Normalize scheme
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)

	// Normalize host
	parsed.Host = strings.ToLower(parsed.Host)
	// Remove www prefix for comparison
	if strings.HasPrefix(parsed.Host, "www.") {
		parsed.Host = parsed.Host[4:]
	}

	// Remove trailing slash from path
	parsed.Path = strings.TrimRight(parsed.Path, "/")

	// Filter out tracking parameters
	query := parsed.Query()
	for _, param := range trackingParams {
		query.Del(param)
	}
	parsed.RawQuery = query.Encode()

	// Remove fragment
	parsed.Fragment = ""

	return parsed.String()
}

func (se *SimilarityEngine) calculateDomainSimilarity(domain1, domain2 string) float64 {
	if domain1 == domain2 {
		return 1.0
	}

	// Remove www prefix for comparison
	d1 := strings.TrimPrefix(strings.ToLower(domain1), "www.")
	d2 := strings.TrimPrefix(strings.ToLower(domain2), "www.")

	if d1 == d2 {
		return 0.95
	}

	// Check if one is subdomain of another
	if strings.Contains(d1, d2) || strings.Contains(d2, d1) {
		return 0.8
	}

	// Calculate string similarity
	return se.calculateStringSimilarity(d1, d2)
}

func (se *SimilarityEngine) calculateQuerySimilarity(query1, query2 string) float64 {
	if query1 == query2 {
		return 1.0
	}

	if query1 == "" && query2 == "" {
		return 1.0
	}

	if query1 == "" || query2 == "" {
		return 0.5 // Partial match if one has no query
	}

	// Parse query parameters
	params1, _ := url.ParseQuery(query1)
	params2, _ := url.ParseQuery(query2)

	if len(params1) == 0 && len(params2) == 0 {
		return 1.0
	}

	// Calculate parameter overlap
	common := 0
	total := len(params1) + len(params2)

	for key, values1 := range params1 {
		if values2, exists := params2[key]; exists {
			if len(values1) > 0 && len(values2) > 0 && values1[0] == values2[0] {
				common += 2 // Both key and value match
			} else {
				common += 1 // Only key matches
			}
		}
	}

	if total == 0 {
		return 1.0
	}

	return float64(common) / float64(total)
}

// Text processing and similarity methods

func (se *SimilarityEngine) normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove punctuation and extra whitespace
	var result strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			result.WriteRune(r)
		} else {
			result.WriteRune(' ')
		}
	}

	// Normalize whitespace
	words := strings.Fields(result.String())
	return strings.Join(words, " ")
}

func (se *SimilarityEngine) calculateJaccardSimilarity(text1, text2 string) float64 {
	words1 := strings.Fields(text1)
	words2 := strings.Fields(text2)

	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range words1 {
		set1[word] = true
	}

	for _, word := range words2 {
		set2[word] = true
	}

	// Calculate intersection
	intersection := 0
	for word := range set1 {
		if set2[word] {
			intersection++
		}
	}

	// Calculate union
	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 1.0
	}

	return float64(intersection) / float64(union)
}

func (se *SimilarityEngine) calculateLevenshteinSimilarity(s1, s2 string) float64 {
	distance := se.levenshteinDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))

	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - (float64(distance) / maxLen)
}

func (se *SimilarityEngine) levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	len1, len2 := len(r1), len(r2)

	if len1 == 0 {
		return len2
	}
	if len2 == 0 {
		return len1
	}

	// Create matrix
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
		matrix[i][0] = i
	}

	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 1
			if r1[i-1] == r2[j-1] {
				cost = 0
			}

			matrix[i][j] = min3(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len1][len2]
}

func (se *SimilarityEngine) calculateNGramSimilarity(text1, text2 string, n int) float64 {
	ngrams1 := se.generateNGrams(text1, n)
	ngrams2 := se.generateNGrams(text2, n)

	if len(ngrams1) == 0 && len(ngrams2) == 0 {
		return 1.0
	}

	if len(ngrams1) == 0 || len(ngrams2) == 0 {
		return 0.0
	}

	// Count common n-grams
	common := 0
	for ngram := range ngrams1 {
		if ngrams2[ngram] {
			common++
		}
	}

	total := len(ngrams1) + len(ngrams2) - common
	if total == 0 {
		return 1.0
	}

	return float64(common) / float64(total)
}

func (se *SimilarityEngine) generateNGrams(text string, n int) map[string]bool {
	ngrams := make(map[string]bool)
	runes := []rune(text)

	if len(runes) < n {
		ngrams[text] = true
		return ngrams
	}

	for i := 0; i <= len(runes)-n; i++ {
		ngram := string(runes[i : i+n])
		ngrams[ngram] = true
	}

	return ngrams
}

func (se *SimilarityEngine) calculateStringSimilarity(s1, s2 string) float64 {
	return se.calculateLevenshteinSimilarity(s1, s2)
}

// Helper function
func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}