package ai

import (
	"math"
	"strings"
	"time"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

// ConfidenceTuner dynamically adjusts AI confidence based on user feedback
type ConfidenceTuner struct {
	learningRepo *db.LearningRepository
	// Cache for domain confidence scores
	domainConfidence map[string]float64
	// Cache for pattern confidence adjustments
	patternAdjustments map[string]float64
	lastUpdate         time.Time
}

// NewConfidenceTuner creates a new confidence tuning system
func NewConfidenceTuner(learningRepo *db.LearningRepository) *ConfidenceTuner {
	return &ConfidenceTuner{
		learningRepo:       learningRepo,
		domainConfidence:   make(map[string]float64),
		patternAdjustments: make(map[string]float64),
		lastUpdate:         time.Now(),
	}
}

// TuneConfidence adjusts base confidence based on historical performance
func (ct *ConfidenceTuner) TuneConfidence(baseConfidence float64, suggestion *TagSuggestion) float64 {
	// Refresh cache if older than 1 hour
	if time.Since(ct.lastUpdate) > time.Hour {
		ct.refreshCache()
	}

	adjustedConfidence := baseConfidence

	// Apply domain-based confidence adjustment
	if domain := extractDomainFromURL(suggestion.URL); domain != "" {
		if domainAdj, exists := ct.domainConfidence[domain]; exists {
			adjustedConfidence *= (1.0 + domainAdj)
		}
	}

	// Apply tag pattern confidence adjustment
	for _, tag := range suggestion.Tags {
		if patternAdj, exists := ct.patternAdjustments[tag]; exists {
			adjustedConfidence *= (1.0 + patternAdj*0.1) // Reduced impact per tag
		}
	}

	// Apply source-based confidence adjustment
	switch suggestion.Source {
	case "learned-patterns":
		adjustedConfidence *= 1.2 // Boost learned patterns
	case "rule-based+content-fetched":
		adjustedConfidence *= 1.1 // Slight boost for content-fetched
	case "rule-based+domain-profile":
		adjustedConfidence *= 1.05 // Slight boost for domain profiles
	}

	// Ensure confidence stays within bounds
	return math.Max(0.1, math.Min(1.0, adjustedConfidence))
}

// AnalyzeFeedbackPatterns analyzes user feedback to identify performance patterns
func (ct *ConfidenceTuner) AnalyzeFeedbackPatterns() (*FeedbackAnalysis, error) {
	corrections, err := ct.learningRepo.GetTagCorrections(500) // Get more for better analysis
	if err != nil {
		return nil, err
	}

	analysis := &FeedbackAnalysis{
		TotalFeedback:      len(corrections),
		AcceptanceRate:     0.0,
		RejectionRate:      0.0,
		ModificationRate:   0.0,
		TopAcceptedTags:    make(map[string]int),
		TopRejectedTags:    make(map[string]int),
		DomainPerformance:  make(map[string]float64),
		PatternPerformance: make(map[string]float64),
		TimeBasedTrends:    make(map[string]int),
	}

	if len(corrections) == 0 {
		return analysis, nil
	}

	accepted := 0
	rejected := 0
	modified := 0

	// Analyze each correction
	for _, correction := range corrections {
		switch correction.CorrectionType {
		case "accepted", "kept":
			accepted++
			ct.analyzeAcceptedTags(correction, analysis)
		case "rejected":
			rejected++
			ct.analyzeRejectedTags(correction, analysis)
		case "modified":
			modified++
			ct.analyzeModifiedTags(correction, analysis)
		}

		// Time-based analysis
		month := correction.CreatedAt.Format("2006-01")
		analysis.TimeBasedTrends[month]++
	}

	// Calculate rates
	total := float64(len(corrections))
	analysis.AcceptanceRate = float64(accepted) / total
	analysis.RejectionRate = float64(rejected) / total
	analysis.ModificationRate = float64(modified) / total

	return analysis, nil
}

// refreshCache updates the confidence adjustment caches
func (ct *ConfidenceTuner) refreshCache() {
	// Analyze recent feedback patterns
	analysis, err := ct.AnalyzeFeedbackPatterns()
	if err != nil {
		return
	}

	// Update domain confidence based on performance
	for domain, performance := range analysis.DomainPerformance {
		// Convert performance to confidence adjustment (-0.3 to +0.3)
		adjustment := (performance - 0.5) * 0.6
		ct.domainConfidence[domain] = adjustment
	}

	// Update pattern adjustments based on acceptance/rejection rates
	for tag, acceptCount := range analysis.TopAcceptedTags {
		rejectCount := analysis.TopRejectedTags[tag]
		totalCount := acceptCount + rejectCount
		
		if totalCount > 5 { // Minimum sample size
			acceptanceRate := float64(acceptCount) / float64(totalCount)
			// Convert to adjustment factor (-0.5 to +0.5)
			adjustment := (acceptanceRate - 0.5) * 1.0
			ct.patternAdjustments[tag] = adjustment
		}
	}

	ct.lastUpdate = time.Now()
}

// analyzeAcceptedTags processes accepted tag feedback
func (ct *ConfidenceTuner) analyzeAcceptedTags(correction models.TagCorrection, analysis *FeedbackAnalysis) {
	for _, tag := range correction.FinalTags {
		analysis.TopAcceptedTags[tag]++
	}
}

// analyzeRejectedTags processes rejected tag feedback  
func (ct *ConfidenceTuner) analyzeRejectedTags(correction models.TagCorrection, analysis *FeedbackAnalysis) {
	for _, tag := range correction.OriginalTags {
		analysis.TopRejectedTags[tag]++
	}
}

// analyzeModifiedTags processes modified tag feedback
func (ct *ConfidenceTuner) analyzeModifiedTags(correction models.TagCorrection, analysis *FeedbackAnalysis) {
	// Tags that were kept
	originalSet := make(map[string]bool)
	for _, tag := range correction.OriginalTags {
		originalSet[tag] = true
	}

	finalSet := make(map[string]bool)
	for _, tag := range correction.FinalTags {
		finalSet[tag] = true
	}

	// Analyze kept vs removed tags
	for _, tag := range correction.OriginalTags {
		if finalSet[tag] {
			analysis.TopAcceptedTags[tag]++ // Tag was kept
		} else {
			analysis.TopRejectedTags[tag]++ // Tag was removed
		}
	}

	// New tags added
	for _, tag := range correction.FinalTags {
		if !originalSet[tag] {
			analysis.TopAcceptedTags[tag]++ // New tag was added
		}
	}
}

// GetConfidenceReport generates a detailed confidence tuning report
func (ct *ConfidenceTuner) GetConfidenceReport() (*ConfidenceReport, error) {
	analysis, err := ct.AnalyzeFeedbackPatterns()
	if err != nil {
		return nil, err
	}

	report := &ConfidenceReport{
		Analysis:           analysis,
		DomainAdjustments:  ct.domainConfidence,
		PatternAdjustments: ct.patternAdjustments,
		LastUpdated:        ct.lastUpdate,
		Recommendations:    ct.generateRecommendations(analysis),
	}

	return report, nil
}

// generateRecommendations creates actionable recommendations based on analysis
func (ct *ConfidenceTuner) generateRecommendations(analysis *FeedbackAnalysis) []string {
	var recommendations []string

	// Acceptance rate recommendations
	if analysis.AcceptanceRate < 0.6 {
		recommendations = append(recommendations, 
			"AI acceptance rate is low. Consider reducing confidence thresholds or improving domain rules.")
	} else if analysis.AcceptanceRate > 0.9 {
		recommendations = append(recommendations, 
			"AI acceptance rate is very high. Consider being more aggressive with suggestions.")
	}

	// Tag-specific recommendations
	for tag, rejectCount := range analysis.TopRejectedTags {
		if rejectCount > 10 {
			recommendations = append(recommendations, 
				"Tag '"+tag+"' is frequently rejected. Consider reviewing its rules or reducing confidence.")
		}
	}

	// Domain-specific recommendations
	for domain, performance := range analysis.DomainPerformance {
		if performance < 0.3 {
			recommendations = append(recommendations, 
				"Domain '"+domain+"' has poor AI performance. Consider adding domain-specific rules.")
		}
	}

	return recommendations
}

// FeedbackAnalysis represents analyzed user feedback patterns
type FeedbackAnalysis struct {
	TotalFeedback      int                 `json:"total_feedback"`
	AcceptanceRate     float64             `json:"acceptance_rate"`
	RejectionRate      float64             `json:"rejection_rate"`
	ModificationRate   float64             `json:"modification_rate"`
	TopAcceptedTags    map[string]int      `json:"top_accepted_tags"`
	TopRejectedTags    map[string]int      `json:"top_rejected_tags"`
	DomainPerformance  map[string]float64  `json:"domain_performance"`
	PatternPerformance map[string]float64  `json:"pattern_performance"`
	TimeBasedTrends    map[string]int      `json:"time_based_trends"`
}

// ConfidenceReport represents a comprehensive confidence tuning report
type ConfidenceReport struct {
	Analysis           *FeedbackAnalysis   `json:"analysis"`
	DomainAdjustments  map[string]float64  `json:"domain_adjustments"`
	PatternAdjustments map[string]float64  `json:"pattern_adjustments"`
	LastUpdated        time.Time           `json:"last_updated"`
	Recommendations    []string            `json:"recommendations"`
}

// Helper function to extract domain from URL
func extractDomainFromURL(url string) string {
	// Simple domain extraction - can be enhanced
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}
	
	return url
}