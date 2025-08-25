package ai

import (
	"net/url"
	"regexp"
	"strings"

	"torimemo/internal/models"
)

// DomainRule defines a rule for categorizing domains
type DomainRule struct {
	Pattern   string
	Tags      []string
	Category  string
	Priority  int
	IsRegex   bool
}

// ContentRule defines a rule for categorizing based on content
type ContentRule struct {
	Keywords  []string
	Tags      []string
	Category  string
	Priority  int
}

// Categorizer handles AI-powered bookmark categorization
type Categorizer struct {
	domainRules  []DomainRule
	contentRules []ContentRule
}

// NewCategorizer creates a new AI categorizer
func NewCategorizer() *Categorizer {
	c := &Categorizer{
		domainRules:  getDefaultDomainRules(),
		contentRules: getDefaultContentRules(),
	}
	return c
}

// TagSuggestion represents AI-suggested tags for a bookmark
type TagSuggestion struct {
	URL        string   `json:"url"`
	Tags       []string `json:"tags"`
	Category   string   `json:"category"`
	Confidence float64  `json:"confidence"`
	Source     string   `json:"source"`
}

// CategorizeBookmark analyzes a bookmark and suggests tags and category
func (c *Categorizer) CategorizeBookmark(bookmark *models.Bookmark) (*TagSuggestion, error) {
	suggestions := &TagSuggestion{
		URL:        bookmark.URL,
		Tags:       make([]string, 0),
		Category:   "general",
		Confidence: 0.0,
		Source:     "rule-based",
	}

	// Extract domain from URL
	parsedURL, err := url.Parse(bookmark.URL)
	if err != nil {
		return suggestions, err
	}
	domain := parsedURL.Hostname()

	// Apply domain rules
	domainTags := c.categorizeDomain(domain)
	suggestions.Tags = append(suggestions.Tags, domainTags...)

	// Apply content rules
	description := ""
	if bookmark.Description != nil {
		description = *bookmark.Description
	}
	contentTags := c.categorizeContent(bookmark.Title, description)
	suggestions.Tags = append(suggestions.Tags, contentTags...)

	// Remove duplicates and set category
	suggestions.Tags = c.removeDuplicates(suggestions.Tags)
	suggestions.Category = c.determineCategory(suggestions.Tags, domain)
	suggestions.Confidence = c.calculateConfidence(suggestions.Tags, domain)

	return suggestions, nil
}

// categorizeDomain applies domain-based rules
func (c *Categorizer) categorizeDomain(domain string) []string {
	tags := make([]string, 0)
	
	for _, rule := range c.domainRules {
		matched := false
		
		if rule.IsRegex {
			regex, err := regexp.Compile(rule.Pattern)
			if err == nil && regex.MatchString(domain) {
				matched = true
			}
		} else {
			if strings.Contains(domain, rule.Pattern) {
				matched = true
			}
		}
		
		if matched {
			tags = append(tags, rule.Tags...)
		}
	}
	
	return tags
}

// categorizeContent applies content-based rules
func (c *Categorizer) categorizeContent(title, description string) []string {
	tags := make([]string, 0)
	content := strings.ToLower(title + " " + description)
	
	for _, rule := range c.contentRules {
		for _, keyword := range rule.Keywords {
			if strings.Contains(content, strings.ToLower(keyword)) {
				tags = append(tags, rule.Tags...)
				break
			}
		}
	}
	
	return tags
}

// determineCategory determines the primary category
func (c *Categorizer) determineCategory(tags []string, domain string) string {
	// Category priority mapping
	categoryMap := map[string]int{
		"development": 10,
		"design":     8,
		"business":   7,
		"education":  6,
		"news":       5,
		"social":     4,
		"entertainment": 3,
		"reference":  2,
		"general":    1,
	}
	
	bestCategory := "general"
	bestScore := 0
	
	for _, tag := range tags {
		if score, exists := categoryMap[tag]; exists && score > bestScore {
			bestCategory = tag
			bestScore = score
		}
	}
	
	return bestCategory
}

// calculateConfidence calculates confidence score
func (c *Categorizer) calculateConfidence(tags []string, domain string) float64 {
	if len(tags) == 0 {
		return 0.1
	}
	
	confidence := 0.5 // Base confidence for having any tags
	confidence += float64(len(tags)) * 0.1 // More tags = higher confidence
	
	// Known domains get higher confidence
	knownDomains := []string{"github.com", "stackoverflow.com", "medium.com", "dev.to", "youtube.com"}
	for _, known := range knownDomains {
		if strings.Contains(domain, known) {
			confidence += 0.3
			break
		}
	}
	
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	return confidence
}

// removeDuplicates removes duplicate tags
func (c *Categorizer) removeDuplicates(tags []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	
	for _, tag := range tags {
		if !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}
	
	return result
}

// getDefaultDomainRules returns the default domain categorization rules
func getDefaultDomainRules() []DomainRule {
	return []DomainRule{
		{Pattern: "github.com", Tags: []string{"development", "code", "programming"}, Category: "development", Priority: 10, IsRegex: false},
		{Pattern: "stackoverflow.com", Tags: []string{"development", "programming", "qa"}, Category: "development", Priority: 10, IsRegex: false},
		{Pattern: "dev.to", Tags: []string{"development", "programming", "blog"}, Category: "development", Priority: 9, IsRegex: false},
		{Pattern: "medium.com", Tags: []string{"blog", "article", "reading"}, Category: "reference", Priority: 7, IsRegex: false},
		{Pattern: "youtube.com", Tags: []string{"video", "entertainment", "tutorial"}, Category: "entertainment", Priority: 8, IsRegex: false},
		{Pattern: "youtu.be", Tags: []string{"video", "entertainment", "tutorial"}, Category: "entertainment", Priority: 8, IsRegex: false},
		{Pattern: "linkedin.com", Tags: []string{"professional", "networking", "business"}, Category: "business", Priority: 8, IsRegex: false},
		{Pattern: "twitter.com", Tags: []string{"social", "news", "micro-blogging"}, Category: "social", Priority: 6, IsRegex: false},
		{Pattern: "x.com", Tags: []string{"social", "news", "micro-blogging"}, Category: "social", Priority: 6, IsRegex: false},
		{Pattern: "reddit.com", Tags: []string{"social", "discussion", "community"}, Category: "social", Priority: 6, IsRegex: false},
		{Pattern: "dribbble.com", Tags: []string{"design", "inspiration", "ui"}, Category: "design", Priority: 9, IsRegex: false},
		{Pattern: "behance.net", Tags: []string{"design", "portfolio", "creative"}, Category: "design", Priority: 9, IsRegex: false},
		{Pattern: "figma.com", Tags: []string{"design", "ui", "tool"}, Category: "design", Priority: 9, IsRegex: false},
		{Pattern: "docs.google.com", Tags: []string{"document", "collaboration", "productivity"}, Category: "reference", Priority: 7, IsRegex: false},
		{Pattern: "notion.so", Tags: []string{"productivity", "notes", "workspace"}, Category: "reference", Priority: 8, IsRegex: false},
		{Pattern: "wikipedia.org", Tags: []string{"reference", "encyclopedia", "education"}, Category: "education", Priority: 9, IsRegex: false},
		{Pattern: "coursera.org", Tags: []string{"education", "course", "learning"}, Category: "education", Priority: 10, IsRegex: false},
		{Pattern: "udemy.com", Tags: []string{"education", "course", "tutorial"}, Category: "education", Priority: 9, IsRegex: false},
		{Pattern: "arxiv.org", Tags: []string{"research", "academic", "paper"}, Category: "education", Priority: 10, IsRegex: false},
		{Pattern: "hackernews.ycombinator.com", Tags: []string{"news", "tech", "startup"}, Category: "news", Priority: 8, IsRegex: false},
		{Pattern: "techcrunch.com", Tags: []string{"news", "tech", "startup"}, Category: "news", Priority: 7, IsRegex: false},
		{Pattern: "aws.amazon.com", Tags: []string{"cloud", "infrastructure", "development"}, Category: "development", Priority: 9, IsRegex: false},
		{Pattern: "cloud.google.com", Tags: []string{"cloud", "infrastructure", "development"}, Category: "development", Priority: 9, IsRegex: false},
		{Pattern: "azure.microsoft.com", Tags: []string{"cloud", "infrastructure", "development"}, Category: "development", Priority: 9, IsRegex: false},
		
		// Regex patterns for broader matching
		{Pattern: `.*\.(edu|ac\.[a-z]{2})$`, Tags: []string{"education", "academic"}, Category: "education", Priority: 8, IsRegex: true},
		{Pattern: `.*\.(gov|mil)$`, Tags: []string{"government", "official"}, Category: "reference", Priority: 7, IsRegex: true},
		{Pattern: `.*\.(org)$`, Tags: []string{"organization", "nonprofit"}, Category: "reference", Priority: 5, IsRegex: true},
	}
}

// getDefaultContentRules returns the default content categorization rules
func getDefaultContentRules() []ContentRule {
	return []ContentRule{
		{Keywords: []string{"javascript", "js", "typescript", "react", "vue", "angular"}, Tags: []string{"javascript", "frontend", "development"}, Category: "development", Priority: 10},
		{Keywords: []string{"python", "django", "flask", "pandas", "numpy"}, Tags: []string{"python", "development", "backend"}, Category: "development", Priority: 10},
		{Keywords: []string{"golang", "go", "goroutines", "gin", "echo"}, Tags: []string{"golang", "development", "backend"}, Category: "development", Priority: 10},
		{Keywords: []string{"rust", "cargo", "tokio", "actix"}, Tags: []string{"rust", "development", "systems"}, Category: "development", Priority: 10},
		{Keywords: []string{"docker", "kubernetes", "k8s", "containerization"}, Tags: []string{"devops", "containers", "infrastructure"}, Category: "development", Priority: 9},
		{Keywords: []string{"machine learning", "ml", "ai", "neural network", "tensorflow", "pytorch"}, Tags: []string{"ai", "machine-learning", "data-science"}, Category: "development", Priority: 9},
		{Keywords: []string{"css", "sass", "scss", "tailwind", "bootstrap"}, Tags: []string{"css", "frontend", "design"}, Category: "design", Priority: 8},
		{Keywords: []string{"ui", "ux", "user interface", "user experience", "design system"}, Tags: []string{"ui", "ux", "design"}, Category: "design", Priority: 9},
		{Keywords: []string{"api", "rest", "graphql", "endpoint", "microservice"}, Tags: []string{"api", "backend", "development"}, Category: "development", Priority: 8},
		{Keywords: []string{"database", "sql", "mongodb", "postgresql", "mysql"}, Tags: []string{"database", "backend", "data"}, Category: "development", Priority: 8},
		{Keywords: []string{"startup", "entrepreneur", "business plan", "venture capital"}, Tags: []string{"startup", "business", "entrepreneurship"}, Category: "business", Priority: 7},
		{Keywords: []string{"marketing", "seo", "social media", "advertising"}, Tags: []string{"marketing", "business", "growth"}, Category: "business", Priority: 7},
		{Keywords: []string{"tutorial", "how to", "guide", "course", "learn"}, Tags: []string{"tutorial", "education", "learning"}, Category: "education", Priority: 8},
		{Keywords: []string{"research", "study", "paper", "academic", "journal"}, Tags: []string{"research", "academic", "education"}, Category: "education", Priority: 9},
		{Keywords: []string{"news", "breaking", "update", "announcement"}, Tags: []string{"news", "current-events"}, Category: "news", Priority: 6},
		{Keywords: []string{"game", "gaming", "entertainment", "fun", "hobby"}, Tags: []string{"gaming", "entertainment", "hobby"}, Category: "entertainment", Priority: 5},
		{Keywords: []string{"tool", "utility", "software", "app", "application"}, Tags: []string{"tools", "software", "productivity"}, Category: "reference", Priority: 6},
		{Keywords: []string{"recipe", "cooking", "food", "kitchen", "ingredient"}, Tags: []string{"cooking", "food", "recipe"}, Category: "reference", Priority: 5},
		{Keywords: []string{"fitness", "health", "workout", "exercise", "wellness"}, Tags: []string{"health", "fitness", "wellness"}, Category: "reference", Priority: 5},
		{Keywords: []string{"travel", "vacation", "trip", "destination", "tourism"}, Tags: []string{"travel", "tourism", "adventure"}, Category: "reference", Priority: 5},
	}
}