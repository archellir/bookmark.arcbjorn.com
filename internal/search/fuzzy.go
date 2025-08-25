package search

import (
	"sort"
	"strings"
	"unicode"
)

// FuzzyResult represents a fuzzy search result with similarity score
type FuzzyResult struct {
	Text       string  `json:"text"`
	Similarity float64 `json:"similarity"`
	Distance   int     `json:"distance"`
}

// FuzzyMatcher handles fuzzy string matching
type FuzzyMatcher struct {
	maxDistance int
	minScore    float64
}

// NewFuzzyMatcher creates a new fuzzy matcher
func NewFuzzyMatcher(maxDistance int, minScore float64) *FuzzyMatcher {
	return &FuzzyMatcher{
		maxDistance: maxDistance,
		minScore:    minScore,
	}
}

// DefaultFuzzyMatcher returns a matcher with sensible defaults
func DefaultFuzzyMatcher() *FuzzyMatcher {
	return NewFuzzyMatcher(3, 0.6) // Max 3 character edits, min 60% similarity
}

// Search performs fuzzy search on a list of strings
func (fm *FuzzyMatcher) Search(query string, candidates []string) []FuzzyResult {
	if len(query) == 0 {
		return nil
	}
	
	query = strings.ToLower(strings.TrimSpace(query))
	var results []FuzzyResult
	
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		
		candidateLower := strings.ToLower(candidate)
		
		// Try exact match first
		if candidateLower == query {
			results = append(results, FuzzyResult{
				Text:       candidate,
				Similarity: 1.0,
				Distance:   0,
			})
			continue
		}
		
		// Try substring match
		if strings.Contains(candidateLower, query) {
			similarity := float64(len(query)) / float64(len(candidate))
			results = append(results, FuzzyResult{
				Text:       candidate,
				Similarity: similarity,
				Distance:   0,
			})
			continue
		}
		
		// Calculate edit distance
		distance := fm.levenshteinDistance(query, candidateLower)
		if distance > fm.maxDistance {
			continue
		}
		
		// Calculate similarity score
		maxLen := max(len(query), len(candidate))
		similarity := 1.0 - float64(distance)/float64(maxLen)
		
		if similarity >= fm.minScore {
			results = append(results, FuzzyResult{
				Text:       candidate,
				Similarity: similarity,
				Distance:   distance,
			})
		}
	}
	
	// Sort by similarity score (descending) then by distance (ascending)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Similarity != results[j].Similarity {
			return results[i].Similarity > results[j].Similarity
		}
		return results[i].Distance < results[j].Distance
	})
	
	return results
}

// FindBest finds the best fuzzy match for a query
func (fm *FuzzyMatcher) FindBest(query string, candidates []string) *FuzzyResult {
	results := fm.Search(query, candidates)
	if len(results) == 0 {
		return nil
	}
	return &results[0]
}

// levenshteinDistance calculates the edit distance between two strings
func (fm *FuzzyMatcher) levenshteinDistance(s1, s2 string) int {
	// Convert to runes to handle Unicode properly
	r1 := []rune(s1)
	r2 := []rune(s2)
	
	len1 := len(r1)
	len2 := len(r2)
	
	// Create a matrix
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}
	
	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}
	
	// Fill the matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
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

// SoundexMatch provides phonetic matching using a simplified Soundex algorithm
func (fm *FuzzyMatcher) SoundexMatch(query string, candidates []string) []FuzzyResult {
	querySoundex := soundex(query)
	var results []FuzzyResult
	
	for _, candidate := range candidates {
		candidateSoundex := soundex(candidate)
		
		if querySoundex == candidateSoundex && querySoundex != "" {
			results = append(results, FuzzyResult{
				Text:       candidate,
				Similarity: 0.8, // High but not perfect for phonetic matches
				Distance:   0,
			})
		}
	}
	
	return results
}

// soundex implements a simplified Soundex algorithm
func soundex(s string) string {
	if s == "" {
		return ""
	}
	
	s = strings.ToUpper(s)
	
	// Keep only letters
	var letters []rune
	for _, r := range s {
		if unicode.IsLetter(r) {
			letters = append(letters, r)
		}
	}
	
	if len(letters) == 0 {
		return ""
	}
	
	// Start with first letter
	result := string(letters[0])
	
	// Soundex mapping
	mapping := map[rune]string{
		'B': "1", 'F': "1", 'P': "1", 'V': "1",
		'C': "2", 'G': "2", 'J': "2", 'K': "2", 'Q': "2", 'S': "2", 'X': "2", 'Z': "2",
		'D': "3", 'T': "3",
		'L': "4",
		'M': "5", 'N': "5",
		'R': "6",
	}
	
	var prev string
	for i := 1; i < len(letters) && len(result) < 4; i++ {
		if code, exists := mapping[letters[i]]; exists {
			if code != prev { // Don't repeat consecutive codes
				result += code
				prev = code
			}
		} else {
			prev = ""
		}
	}
	
	// Pad with zeros to make it 4 characters
	for len(result) < 4 {
		result += "0"
	}
	
	return result
}

// TokenBasedMatch performs fuzzy matching on individual words/tokens
func (fm *FuzzyMatcher) TokenBasedMatch(query string, candidates []string) []FuzzyResult {
	queryTokens := tokenize(query)
	var results []FuzzyResult
	
	for _, candidate := range candidates {
		candidateTokens := tokenize(candidate)
		score := fm.calculateTokenScore(queryTokens, candidateTokens)
		
		if score >= fm.minScore {
			results = append(results, FuzzyResult{
				Text:       candidate,
				Similarity: score,
				Distance:   0,
			})
		}
	}
	
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})
	
	return results
}

// calculateTokenScore calculates similarity based on token matching
func (fm *FuzzyMatcher) calculateTokenScore(queryTokens, candidateTokens []string) float64 {
	if len(queryTokens) == 0 || len(candidateTokens) == 0 {
		return 0.0
	}
	
	matches := 0
	for _, qToken := range queryTokens {
		for _, cToken := range candidateTokens {
			// Check for exact match
			if strings.EqualFold(qToken, cToken) {
				matches++
				break
			}
			
			// Check for fuzzy match
			distance := fm.levenshteinDistance(strings.ToLower(qToken), strings.ToLower(cToken))
			maxLen := max(len(qToken), len(cToken))
			
			if distance <= 2 && maxLen > 3 { // Allow 1-2 errors for longer words
				similarity := 1.0 - float64(distance)/float64(maxLen)
				if similarity >= 0.7 {
					matches++
					break
				}
			}
		}
	}
	
	return float64(matches) / float64(len(queryTokens))
}

// tokenize splits text into searchable tokens
func tokenize(text string) []string {
	// Split on whitespace and punctuation
	var tokens []string
	var current strings.Builder
	
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}
	
	// Add final token if exists
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	
	// Filter out very short tokens
	var filtered []string
	for _, token := range tokens {
		if len(token) >= 2 {
			filtered = append(filtered, token)
		}
	}
	
	return filtered
}

// Helper functions
func min3(a, b, c int) int {
	return min(min(a, b), c)
}

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