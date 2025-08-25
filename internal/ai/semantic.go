package ai

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// SemanticAnalyzer provides advanced semantic analysis for tag suggestions
type SemanticAnalyzer struct {
	// Word embeddings approximated through co-occurrence patterns
	wordVectors map[string][]float64
	// Semantic clusters of related terms
	semanticClusters map[string][]string
	// Context patterns for better understanding
	contextPatterns map[string][]string
}

// NewSemanticAnalyzer creates a new semantic analysis engine
func NewSemanticAnalyzer() *SemanticAnalyzer {
	sa := &SemanticAnalyzer{
		wordVectors:      make(map[string][]float64),
		semanticClusters: getSemanticClusters(),
		contextPatterns:  getContextPatterns(),
	}
	sa.initializeWordVectors()
	return sa
}

// SemanticSuggestion represents a semantically-enhanced tag suggestion
type SemanticSuggestion struct {
	Tag             string  `json:"tag"`
	Confidence      float64 `json:"confidence"`
	SemanticScore   float64 `json:"semantic_score"`
	ContextRelevance float64 `json:"context_relevance"`
	Source          string  `json:"source"` // "semantic", "contextual", "cluster"
	RelatedTerms    []string `json:"related_terms,omitempty"`
}

// AnalyzeSemanticContent performs deep semantic analysis of content
func (sa *SemanticAnalyzer) AnalyzeSemanticContent(title, description, url string) []SemanticSuggestion {
	var suggestions []SemanticSuggestion

	// Combine all text for analysis
	fullText := strings.ToLower(title + " " + description)
	words := sa.extractWords(fullText)

	// Semantic vector analysis
	semanticSuggestions := sa.analyzeSemanticVectors(words, fullText)
	suggestions = append(suggestions, semanticSuggestions...)

	// Contextual pattern analysis
	contextualSuggestions := sa.analyzeContextualPatterns(fullText, url)
	suggestions = append(suggestions, contextualSuggestions...)

	// Cluster-based suggestions
	clusterSuggestions := sa.analyzeSemanticClusters(words)
	suggestions = append(suggestions, clusterSuggestions...)

	// Remove duplicates and sort by confidence
	suggestions = sa.deduplicateAndSort(suggestions)

	return suggestions
}

// EnhanceExistingSuggestions improves existing tag suggestions with semantic analysis
func (sa *SemanticAnalyzer) EnhanceExistingSuggestions(existing []TagSuggestion, content string) []TagSuggestion {
	enhanced := make([]TagSuggestion, len(existing))
	copy(enhanced, existing)

	words := sa.extractWords(strings.ToLower(content))

	for i, suggestion := range enhanced {
		// Calculate average semantic relevance for all tags in suggestion
		totalSemanticScore := 0.0
		validTags := 0
		
		for _, tag := range suggestion.Tags {
			semanticScore := sa.calculateSemanticRelevance(tag, words, content)
			totalSemanticScore += semanticScore
			validTags++
		}
		
		if validTags > 0 {
			avgSemanticScore := totalSemanticScore / float64(validTags)
			
			// Adjust confidence based on semantic analysis
			semanticBoost := avgSemanticScore * 0.3 // Max 30% boost
			enhanced[i].Confidence = math.Min(1.0, suggestion.Confidence+semanticBoost)
			
			// Add semantic source information
			if avgSemanticScore > 0.7 {
				enhanced[i].Source += "+semantic"
			}
		}
	}

	return enhanced
}

// analyzeSemanticVectors uses word vector approximations for analysis
func (sa *SemanticAnalyzer) analyzeSemanticVectors(words []string, fullText string) []SemanticSuggestion {
	var suggestions []SemanticSuggestion
	
	// Calculate document vector as average of word vectors
	docVector := sa.calculateDocumentVector(words)
	if docVector == nil {
		return suggestions
	}

	// Find semantically similar tags
	tagScores := make(map[string]float64)
	
	// Check against common tags with vectors
	commonTags := []string{
		"programming", "web-development", "design", "tutorial", "reference",
		"news", "blog", "documentation", "tool", "framework", "library",
		"article", "guide", "javascript", "python", "react", "vue", "angular",
		"database", "api", "security", "performance", "testing", "devops",
		"machine-learning", "ai", "data-science", "mobile", "android", "ios",
		"startup", "business", "marketing", "finance", "productivity", "health",
		"science", "technology", "education", "entertainment", "gaming",
	}

	for _, tag := range commonTags {
		if tagVector := sa.getTagVector(tag); tagVector != nil {
			similarity := sa.calculateCosineSimilarity(docVector, tagVector)
			if similarity > 0.3 { // Threshold for semantic relevance
				tagScores[tag] = similarity
			}
		}
	}

	// Convert to suggestions
	for tag, score := range tagScores {
		// Boost score based on direct word matches
		directMatch := sa.hasDirectWordMatch(tag, words)
		if directMatch {
			score *= 1.2
		}

		suggestions = append(suggestions, SemanticSuggestion{
			Tag:              tag,
			Confidence:       math.Min(0.9, score),
			SemanticScore:    score,
			ContextRelevance: sa.calculateContextRelevance(tag, fullText),
			Source:           "semantic",
		})
	}

	return suggestions
}

// analyzeContextualPatterns identifies patterns in context for better tagging
func (sa *SemanticAnalyzer) analyzeContextualPatterns(fullText, url string) []SemanticSuggestion {
	var suggestions []SemanticSuggestion

	// URL-based contextual analysis
	urlSuggestions := sa.analyzeURLContext(url)
	suggestions = append(suggestions, urlSuggestions...)

	// Text pattern analysis
	for pattern, tags := range sa.contextPatterns {
		if strings.Contains(fullText, pattern) {
			for _, tag := range tags {
				confidence := 0.6 + (float64(strings.Count(fullText, pattern)) * 0.1)
				confidence = math.Min(0.85, confidence)

				suggestions = append(suggestions, SemanticSuggestion{
					Tag:              tag,
					Confidence:       confidence,
					SemanticScore:    0.7,
					ContextRelevance: 0.8,
					Source:           "contextual",
				})
			}
		}
	}

	return suggestions
}

// analyzeSemanticClusters uses semantic clustering for tag suggestions
func (sa *SemanticAnalyzer) analyzeSemanticClusters(words []string) []SemanticSuggestion {
	var suggestions []SemanticSuggestion

	// Check words against semantic clusters
	for clusterTag, clusterWords := range sa.semanticClusters {
		matchScore := sa.calculateClusterMatch(words, clusterWords)
		
		if matchScore > 0.2 {
			suggestions = append(suggestions, SemanticSuggestion{
				Tag:              clusterTag,
				Confidence:       math.Min(0.8, matchScore),
				SemanticScore:    matchScore,
				ContextRelevance: 0.6,
				Source:           "cluster",
				RelatedTerms:     sa.getMatchingTerms(words, clusterWords),
			})
		}
	}

	return suggestions
}

// Helper methods

func (sa *SemanticAnalyzer) extractWords(text string) []string {
	var words []string
	
	// Simple word extraction with filtering
	wordSlice := strings.Fields(text)
	for _, word := range wordSlice {
		// Clean word
		cleaned := strings.TrimFunc(word, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})
		
		// Filter short words and common stop words
		if len(cleaned) > 2 && !sa.isStopWord(cleaned) {
			words = append(words, cleaned)
		}
	}
	
	return words
}

func (sa *SemanticAnalyzer) isStopWord(word string) bool {
	stopWords := []string{
		"the", "and", "for", "are", "but", "not", "you", "all", "can", "had",
		"her", "was", "one", "our", "out", "day", "get", "has", "him", "his",
		"how", "man", "new", "now", "old", "see", "two", "way", "who", "boy",
		"did", "its", "let", "put", "say", "she", "too", "use", "this", "that",
	}
	
	for _, stop := range stopWords {
		if word == stop {
			return true
		}
	}
	return false
}

func (sa *SemanticAnalyzer) calculateDocumentVector(words []string) []float64 {
	if len(words) == 0 {
		return nil
	}

	vectorSize := 50 // Simplified vector size
	docVector := make([]float64, vectorSize)
	count := 0

	for _, word := range words {
		if vector, exists := sa.wordVectors[word]; exists {
			for i, val := range vector {
				docVector[i] += val
			}
			count++
		}
	}

	if count == 0 {
		return nil
	}

	// Normalize by averaging
	for i := range docVector {
		docVector[i] /= float64(count)
	}

	return docVector
}

func (sa *SemanticAnalyzer) getTagVector(tag string) []float64 {
	if vector, exists := sa.wordVectors[tag]; exists {
		return vector
	}
	
	// Generate synthetic vector for unknown tags
	return sa.generateSyntheticVector(tag)
}

func (sa *SemanticAnalyzer) calculateCosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	var dotProduct, norm1, norm2 float64
	
	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

func (sa *SemanticAnalyzer) hasDirectWordMatch(tag string, words []string) bool {
	tagWords := strings.Fields(strings.ReplaceAll(tag, "-", " "))
	
	for _, tagWord := range tagWords {
		for _, word := range words {
			if strings.Contains(word, tagWord) || strings.Contains(tagWord, word) {
				return true
			}
		}
	}
	
	return false
}

func (sa *SemanticAnalyzer) calculateContextRelevance(tag string, fullText string) float64 {
	// Simple context relevance based on word proximity and co-occurrence
	tagWords := strings.Fields(strings.ReplaceAll(tag, "-", " "))
	relevance := 0.0
	
	for _, tagWord := range tagWords {
		if strings.Contains(fullText, tagWord) {
			relevance += 0.3
		}
	}
	
	return math.Min(1.0, relevance)
}

func (sa *SemanticAnalyzer) analyzeURLContext(url string) []SemanticSuggestion {
	var suggestions []SemanticSuggestion

	urlLower := strings.ToLower(url)

	// Domain-based suggestions
	domainPatterns := map[string][]string{
		"github.com":       {"programming", "open-source", "code", "repository"},
		"stackoverflow.com": {"programming", "q-and-a", "tutorial", "debugging"},
		"medium.com":       {"blog", "article", "tutorial", "opinion"},
		"dev.to":          {"programming", "blog", "tutorial", "community"},
		"youtube.com":      {"video", "tutorial", "entertainment", "educational"},
		"docs.":           {"documentation", "reference", "guide"},
		"api.":            {"api", "reference", "documentation"},
		"blog":            {"blog", "article", "personal"},
		"news":            {"news", "current-events", "journalism"},
	}

	for pattern, tags := range domainPatterns {
		if strings.Contains(urlLower, pattern) {
			for _, tag := range tags {
				suggestions = append(suggestions, SemanticSuggestion{
					Tag:              tag,
					Confidence:       0.7,
					SemanticScore:    0.6,
					ContextRelevance: 0.9,
					Source:           "contextual",
				})
			}
		}
	}

	return suggestions
}

func (sa *SemanticAnalyzer) calculateClusterMatch(words []string, clusterWords []string) float64 {
	matches := 0
	for _, word := range words {
		for _, clusterWord := range clusterWords {
			if strings.Contains(word, clusterWord) || strings.Contains(clusterWord, word) {
				matches++
				break
			}
		}
	}

	if len(words) == 0 {
		return 0.0
	}

	return float64(matches) / float64(len(words))
}

func (sa *SemanticAnalyzer) getMatchingTerms(words []string, clusterWords []string) []string {
	var matching []string
	
	for _, word := range words {
		for _, clusterWord := range clusterWords {
			if strings.Contains(word, clusterWord) || strings.Contains(clusterWord, word) {
				matching = append(matching, clusterWord)
				break
			}
		}
	}
	
	return matching
}

func (sa *SemanticAnalyzer) calculateSemanticRelevance(tag string, words []string, content string) float64 {
	score := 0.0
	
	// Direct word match
	if sa.hasDirectWordMatch(tag, words) {
		score += 0.4
	}
	
	// Semantic cluster match
	if clusterWords, exists := sa.semanticClusters[tag]; exists {
		clusterScore := sa.calculateClusterMatch(words, clusterWords)
		score += clusterScore * 0.3
	}
	
	// Context relevance
	contextScore := sa.calculateContextRelevance(tag, content)
	score += contextScore * 0.3
	
	return math.Min(1.0, score)
}

func (sa *SemanticAnalyzer) deduplicateAndSort(suggestions []SemanticSuggestion) []SemanticSuggestion {
	// Remove duplicates by tag, keeping highest confidence
	tagMap := make(map[string]SemanticSuggestion)
	
	for _, suggestion := range suggestions {
		if existing, exists := tagMap[suggestion.Tag]; exists {
			if suggestion.Confidence > existing.Confidence {
				tagMap[suggestion.Tag] = suggestion
			}
		} else {
			tagMap[suggestion.Tag] = suggestion
		}
	}
	
	// Convert back to slice
	var deduplicated []SemanticSuggestion
	for _, suggestion := range tagMap {
		deduplicated = append(deduplicated, suggestion)
	}
	
	// Sort by confidence
	sort.Slice(deduplicated, func(i, j int) bool {
		return deduplicated[i].Confidence > deduplicated[j].Confidence
	})
	
	return deduplicated
}

// Initialize word vectors with simplified embeddings
func (sa *SemanticAnalyzer) initializeWordVectors() {
	// This is a simplified word vector system
	// In production, you'd load pre-trained embeddings
	
	techWords := []string{"programming", "code", "software", "development", "web", "api", "database"}
	designWords := []string{"design", "ui", "ux", "graphics", "visual", "layout", "typography"}
	businessWords := []string{"business", "startup", "marketing", "finance", "strategy", "management"}
	
	sa.addWordCluster(techWords, []float64{0.8, 0.2, 0.1})
	sa.addWordCluster(designWords, []float64{0.2, 0.8, 0.1})  
	sa.addWordCluster(businessWords, []float64{0.1, 0.2, 0.8})
}

func (sa *SemanticAnalyzer) addWordCluster(words []string, baseVector []float64) {
	for i, word := range words {
		vector := make([]float64, 50)
		
		// Set base semantic dimensions
		for j, val := range baseVector {
			vector[j] = val
		}
		
		// Add word-specific variations
		for j := len(baseVector); j < len(vector); j++ {
			vector[j] = math.Sin(float64(i+j)) * 0.3
		}
		
		sa.wordVectors[word] = vector
	}
}

func (sa *SemanticAnalyzer) generateSyntheticVector(word string) []float64 {
	vector := make([]float64, 50)
	
	// Generate pseudo-random but consistent vector based on word
	for i := 0; i < len(vector); i++ {
		// Simple hash-based generation
		hash := 0
		for _, r := range word {
			hash = hash*31 + int(r)
		}
		vector[i] = math.Sin(float64(hash+i)) * 0.5
	}
	
	return vector
}

// Pre-defined semantic clusters
func getSemanticClusters() map[string][]string {
	return map[string][]string{
		"programming": {"code", "coding", "development", "software", "programming", "developer", "tech", "computer", "algorithm", "function", "variable", "class", "method"},
		"web-development": {"html", "css", "javascript", "react", "vue", "angular", "frontend", "backend", "fullstack", "web", "website", "browser", "dom", "ajax"},
		"design": {"ui", "ux", "design", "visual", "graphics", "layout", "typography", "color", "interface", "user", "experience", "wireframe", "mockup"},
		"tutorial": {"tutorial", "guide", "howto", "learn", "learning", "education", "course", "lesson", "step", "instruction", "example", "demo"},
		"reference": {"documentation", "docs", "reference", "api", "manual", "guide", "specification", "cheatsheet", "lookup"},
		"news": {"news", "article", "blog", "post", "update", "announcement", "release", "breaking", "latest", "current"},
		"tool": {"tool", "utility", "app", "application", "software", "program", "service", "platform", "system"},
		"framework": {"framework", "library", "package", "module", "component", "plugin", "extension", "addon"},
		"database": {"database", "db", "sql", "nosql", "mysql", "postgresql", "mongodb", "data", "storage", "query"},
		"security": {"security", "encryption", "authentication", "authorization", "vulnerability", "exploit", "patch", "privacy", "protection"},
		"performance": {"performance", "optimization", "speed", "fast", "slow", "benchmark", "profiling", "cache", "memory", "cpu"},
		"testing": {"test", "testing", "unit", "integration", "automation", "qa", "quality", "bug", "debug", "mock"},
		"business": {"business", "startup", "entrepreneur", "company", "corporate", "enterprise", "strategy", "market", "customer"},
		"finance": {"finance", "money", "investment", "trading", "stock", "crypto", "bitcoin", "blockchain", "economics", "budget"},
		"productivity": {"productivity", "workflow", "automation", "efficiency", "organization", "planning", "time", "management", "gtd"},
		"health": {"health", "fitness", "medicine", "wellness", "nutrition", "exercise", "mental", "physical", "medical", "doctor"},
		"science": {"science", "research", "study", "analysis", "data", "experiment", "hypothesis", "theory", "academic", "paper"},
		"ai": {"ai", "artificial", "intelligence", "machine", "learning", "neural", "network", "deep", "algorithm", "model", "prediction"},
		"mobile": {"mobile", "app", "ios", "android", "phone", "smartphone", "tablet", "responsive", "native", "hybrid"},
	}
}

// Pre-defined context patterns
func getContextPatterns() map[string][]string {
	return map[string][]string{
		"how to": {"tutorial", "guide", "howto"},
		"getting started": {"beginner", "tutorial", "guide"},
		"best practices": {"best-practices", "guide", "reference"},
		"cheat sheet": {"cheatsheet", "reference", "quick-reference"},
		"vs": {"comparison", "versus", "alternatives"},
		"review": {"review", "analysis", "opinion"},
		"introduction": {"beginner", "introduction", "basics"},
		"advanced": {"advanced", "expert", "deep-dive"},
		"free": {"free", "open-source", "gratis"},
		"open source": {"open-source", "github", "community"},
		"case study": {"case-study", "example", "real-world"},
		"benchmark": {"performance", "benchmark", "comparison"},
		"interview": {"interview", "questions", "preparation"},
		"roadmap": {"roadmap", "learning-path", "guide"},
	}
}