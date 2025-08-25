package ai

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// PredictiveTagEngine provides AI-powered predictive tag suggestions
type PredictiveTagEngine struct {
	semanticAnalyzer *SemanticAnalyzer
	userPatterns     map[int]*UserTaggingPattern // userID -> patterns
	contextCache     map[string][]PredictiveTag  // URL/content hash -> predicted tags
	lastUpdated      time.Time
}

// NewPredictiveTagEngine creates a new predictive tag engine
func NewPredictiveTagEngine() *PredictiveTagEngine {
	return &PredictiveTagEngine{
		semanticAnalyzer: NewSemanticAnalyzer(),
		userPatterns:     make(map[int]*UserTaggingPattern),
		contextCache:     make(map[string][]PredictiveTag),
		lastUpdated:      time.Now(),
	}
}

// UserTaggingPattern represents learned user tagging behavior
type UserTaggingPattern struct {
	UserID               int                        `json:"user_id"`
	TagFrequency         map[string]int             `json:"tag_frequency"`
	TagCooccurrence      map[string]map[string]int  `json:"tag_cooccurrence"`
	DomainTagPreferences map[string][]string        `json:"domain_tag_preferences"`
	TimeBasedPatterns    map[string][]string        `json:"time_based_patterns"` // hour/day -> preferred tags
	TagSequences         [][]string                 `json:"tag_sequences"`       // Common tag sequences
	RecentTags           []TimestampedTag           `json:"recent_tags"`         // Recent usage for recency boost
	TagCategories        map[string][]string        `json:"tag_categories"`      // Auto-detected tag categories
	PersonalVocabulary   []string                   `json:"personal_vocabulary"` // User's unique tag vocabulary
	LastUpdated          time.Time                  `json:"last_updated"`
}

// TimestampedTag represents a tag with usage timestamp
type TimestampedTag struct {
	Tag       string    `json:"tag"`
	Timestamp time.Time `json:"timestamp"`
	Context   string    `json:"context"` // URL or content context
}

// PredictiveTag represents a predicted tag suggestion
type PredictiveTag struct {
	Tag             string                 `json:"tag"`
	Confidence      float64                `json:"confidence"`
	PredictionType  string                 `json:"prediction_type"` // "frequency", "cooccurrence", "semantic", "temporal", "sequence"
	Reason          string                 `json:"reason"`
	RelatedTags     []string               `json:"related_tags,omitempty"`
	UserRelevance   float64                `json:"user_relevance"`   // How relevant to this specific user
	ContextMatch    float64                `json:"context_match"`    // How well it matches current context
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PredictionContext provides context for tag prediction
type PredictionContext struct {
	UserID      int      `json:"user_id"`
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ExistingTags []string `json:"existing_tags"`
	Time        time.Time `json:"time"`
	Location    string   `json:"location,omitempty"` // Geographic context if available
}

// PredictiveAnalysisRequest represents a request for predictive analysis
type PredictiveAnalysisRequest struct {
	Context         PredictionContext `json:"context"`
	MaxSuggestions  int               `json:"max_suggestions"`
	MinConfidence   float64           `json:"min_confidence"`
	IncludeMetadata bool              `json:"include_metadata"`
}

// PredictiveAnalysisResult represents the result of predictive analysis
type PredictiveAnalysisResult struct {
	Predictions       []PredictiveTag          `json:"predictions"`
	UserPattern       *UserTaggingPattern      `json:"user_pattern,omitempty"`
	ConfidenceScore   float64                  `json:"confidence_score"`
	ProcessingTimeMs  int64                    `json:"processing_time_ms"`
	PredictionSummary PredictionSummary        `json:"prediction_summary"`
}

// PredictionSummary provides insights about the predictions
type PredictionSummary struct {
	TotalCandidates    int     `json:"total_candidates"`
	HighConfidenceCount int    `json:"high_confidence_count"`
	SemanticCount      int     `json:"semantic_count"`
	PatternBasedCount  int     `json:"pattern_based_count"`
	NovelSuggestions   int     `json:"novel_suggestions"` // Tags not in user's vocabulary
	Insights           []string `json:"insights"`
}

// PredictTags performs intelligent tag prediction based on context
func (pte *PredictiveTagEngine) PredictTags(req PredictiveAnalysisRequest) (*PredictiveAnalysisResult, error) {
	startTime := time.Now()

	// Get or create user pattern
	userPattern := pte.getUserPattern(req.Context.UserID)

	// Generate predictions from multiple sources
	var allPredictions []PredictiveTag

	// 1. Frequency-based predictions
	frequencyPreds := pte.predictFromFrequency(userPattern, req.Context)
	allPredictions = append(allPredictions, frequencyPreds...)

	// 2. Co-occurrence based predictions
	cooccurrencePreds := pte.predictFromCooccurrence(userPattern, req.Context)
	allPredictions = append(allPredictions, cooccurrencePreds...)

	// 3. Domain-based predictions
	domainPreds := pte.predictFromDomain(userPattern, req.Context)
	allPredictions = append(allPredictions, domainPreds...)

	// 4. Semantic predictions
	semanticPreds := pte.predictFromSemantics(req.Context)
	allPredictions = append(allPredictions, semanticPreds...)

	// 5. Temporal pattern predictions
	temporalPreds := pte.predictFromTemporal(userPattern, req.Context)
	allPredictions = append(allPredictions, temporalPreds...)

	// 6. Sequence-based predictions
	sequencePreds := pte.predictFromSequences(userPattern, req.Context)
	allPredictions = append(allPredictions, sequencePreds...)

	// Merge and rank predictions
	finalPredictions := pte.mergeAndRankPredictions(allPredictions, req.Context, userPattern)

	// Filter by confidence and limit
	filteredPredictions := pte.filterPredictions(finalPredictions, req.MinConfidence, req.MaxSuggestions)

	// Generate analysis summary
	summary := pte.generatePredictionSummary(allPredictions, filteredPredictions, userPattern)

	result := &PredictiveAnalysisResult{
		Predictions:       filteredPredictions,
		ConfidenceScore:   pte.calculateOverallConfidence(filteredPredictions),
		ProcessingTimeMs:  time.Since(startTime).Milliseconds(),
		PredictionSummary: summary,
	}

	if req.IncludeMetadata {
		result.UserPattern = userPattern
	}

	return result, nil
}

// LearnFromUserFeedback updates user patterns based on tag selections
func (pte *PredictiveTagEngine) LearnFromUserFeedback(userID int, context PredictionContext, selectedTags, rejectedTags []string) error {
	userPattern := pte.getUserPattern(userID)
	
	// Update tag frequency
	for _, tag := range selectedTags {
		userPattern.TagFrequency[tag]++
		
		// Add to recent tags
		userPattern.RecentTags = append(userPattern.RecentTags, TimestampedTag{
			Tag:       tag,
			Timestamp: time.Now(),
			Context:   context.URL,
		})
	}

	// Update co-occurrence patterns
	for i, tag1 := range selectedTags {
		if userPattern.TagCooccurrence[tag1] == nil {
			userPattern.TagCooccurrence[tag1] = make(map[string]int)
		}
		
		for j, tag2 := range selectedTags {
			if i != j {
				userPattern.TagCooccurrence[tag1][tag2]++
			}
		}
	}

	// Update domain preferences
	domain := pte.extractDomain(context.URL)
	if domain != "" {
		domainTags := userPattern.DomainTagPreferences[domain]
		for _, tag := range selectedTags {
			if !pte.containsString(domainTags, tag) {
				userPattern.DomainTagPreferences[domain] = append(domainTags, tag)
			}
		}
	}

	// Update time-based patterns
	timeKey := context.Time.Format("15") // Hour of day
	timeTags := userPattern.TimeBasedPatterns[timeKey]
	for _, tag := range selectedTags {
		if !pte.containsString(timeTags, tag) {
			userPattern.TimeBasedPatterns[timeKey] = append(timeTags, tag)
		}
	}

	// Learn tag sequences if multiple tags were selected
	if len(selectedTags) > 1 {
		userPattern.TagSequences = append(userPattern.TagSequences, selectedTags)
		// Keep only recent sequences (last 100)
		if len(userPattern.TagSequences) > 100 {
			userPattern.TagSequences = userPattern.TagSequences[len(userPattern.TagSequences)-100:]
		}
	}

	// Update personal vocabulary
	for _, tag := range selectedTags {
		if !pte.containsString(userPattern.PersonalVocabulary, tag) {
			userPattern.PersonalVocabulary = append(userPattern.PersonalVocabulary, tag)
		}
	}

	// Trim recent tags to keep only last 50
	if len(userPattern.RecentTags) > 50 {
		userPattern.RecentTags = userPattern.RecentTags[len(userPattern.RecentTags)-50:]
	}

	userPattern.LastUpdated = time.Now()
	pte.userPatterns[userID] = userPattern

	return nil
}

// Prediction methods

func (pte *PredictiveTagEngine) predictFromFrequency(pattern *UserTaggingPattern, context PredictionContext) []PredictiveTag {
	var predictions []PredictiveTag

	// Get top frequent tags
	type tagFreq struct {
		tag   string
		freq  int
	}

	var sortedTags []tagFreq
	for tag, freq := range pattern.TagFrequency {
		// Skip if tag is already in existing tags
		if pte.containsString(context.ExistingTags, tag) {
			continue
		}
		sortedTags = append(sortedTags, tagFreq{tag, freq})
	}

	sort.Slice(sortedTags, func(i, j int) bool {
		return sortedTags[i].freq > sortedTags[j].freq
	})

	// Convert to predictions (top 10)
	limit := 10
	if len(sortedTags) < limit {
		limit = len(sortedTags)
	}

	totalUsage := 0
	for _, tf := range sortedTags {
		totalUsage += tf.freq
	}

	for i := 0; i < limit; i++ {
		tf := sortedTags[i]
		confidence := float64(tf.freq) / float64(totalUsage)
		
		predictions = append(predictions, PredictiveTag{
			Tag:            tf.tag,
			Confidence:     confidence,
			PredictionType: "frequency",
			Reason:         fmt.Sprintf("Used %d times (%.1f%% of your tags)", tf.freq, confidence*100),
			UserRelevance:  confidence,
			ContextMatch:   0.5, // Neutral for frequency-based
		})
	}

	return predictions
}

func (pte *PredictiveTagEngine) predictFromCooccurrence(pattern *UserTaggingPattern, context PredictionContext) []PredictiveTag {
	var predictions []PredictiveTag

	if len(context.ExistingTags) == 0 {
		return predictions
	}

	// Analyze co-occurrence with existing tags
	cooccurrenceScores := make(map[string]float64)

	for _, existingTag := range context.ExistingTags {
		if cooccurringTags, exists := pattern.TagCooccurrence[existingTag]; exists {
			for tag, count := range cooccurringTags {
				// Skip if already in existing tags
				if pte.containsString(context.ExistingTags, tag) {
					continue
				}
				cooccurrenceScores[tag] += float64(count)
			}
		}
	}

	// Convert to predictions
	type tagScore struct {
		tag   string
		score float64
	}

	var sortedTags []tagScore
	for tag, score := range cooccurrenceScores {
		sortedTags = append(sortedTags, tagScore{tag, score})
	}

	sort.Slice(sortedTags, func(i, j int) bool {
		return sortedTags[i].score > sortedTags[j].score
	})

	// Create predictions (top 10)
	limit := 10
	if len(sortedTags) < limit {
		limit = len(sortedTags)
	}

	maxScore := 0.0
	if len(sortedTags) > 0 {
		maxScore = sortedTags[0].score
	}

	for i := 0; i < limit; i++ {
		ts := sortedTags[i]
		confidence := ts.score / maxScore
		
		relatedTags := []string{}
		for _, et := range context.ExistingTags {
			if pattern.TagCooccurrence[et][ts.tag] > 0 {
				relatedTags = append(relatedTags, et)
			}
		}

		predictions = append(predictions, PredictiveTag{
			Tag:            ts.tag,
			Confidence:     confidence,
			PredictionType: "cooccurrence",
			Reason:         fmt.Sprintf("Often used with %s", strings.Join(relatedTags, ", ")),
			RelatedTags:    relatedTags,
			UserRelevance:  confidence,
			ContextMatch:   0.7, // Higher for co-occurrence
		})
	}

	return predictions
}

func (pte *PredictiveTagEngine) predictFromDomain(pattern *UserTaggingPattern, context PredictionContext) []PredictiveTag {
	var predictions []PredictiveTag

	domain := pte.extractDomain(context.URL)
	if domain == "" {
		return predictions
	}

	domainTags, exists := pattern.DomainTagPreferences[domain]
	if !exists || len(domainTags) == 0 {
		return predictions
	}

	// Score tags based on frequency in this domain
	for _, tag := range domainTags {
		// Skip if already in existing tags
		if pte.containsString(context.ExistingTags, tag) {
			continue
		}

		// Calculate confidence based on frequency in this domain vs overall
		overallFreq := pattern.TagFrequency[tag]
		domainUsage := pte.countTagInDomain(pattern, tag, domain)
		
		confidence := 0.5 // Base confidence
		if overallFreq > 0 {
			confidence = float64(domainUsage) / float64(overallFreq)
		}

		predictions = append(predictions, PredictiveTag{
			Tag:            tag,
			Confidence:     confidence,
			PredictionType: "domain",
			Reason:         fmt.Sprintf("Previously used for %s content", domain),
			UserRelevance:  confidence,
			ContextMatch:   0.8, // High for domain match
			Metadata: map[string]interface{}{
				"domain":      domain,
				"domain_usage": domainUsage,
			},
		})
	}

	return predictions
}

func (pte *PredictiveTagEngine) predictFromSemantics(context PredictionContext) []PredictiveTag {
	var predictions []PredictiveTag

	// Use semantic analyzer to get semantic suggestions
	semanticSuggestions := pte.semanticAnalyzer.AnalyzeSemanticContent(context.Title, context.Description, context.URL)

	for _, semSugg := range semanticSuggestions {
		// Skip if already in existing tags
		if pte.containsString(context.ExistingTags, semSugg.Tag) {
			continue
		}

		predictions = append(predictions, PredictiveTag{
			Tag:            semSugg.Tag,
			Confidence:     semSugg.Confidence,
			PredictionType: "semantic",
			Reason:         "Semantically related to content",
			RelatedTags:    semSugg.RelatedTerms,
			UserRelevance:  0.5, // Neutral - not user-specific
			ContextMatch:   semSugg.ContextRelevance,
			Metadata: map[string]interface{}{
				"semantic_score": semSugg.SemanticScore,
				"source":         semSugg.Source,
			},
		})
	}

	return predictions
}

func (pte *PredictiveTagEngine) predictFromTemporal(pattern *UserTaggingPattern, context PredictionContext) []PredictiveTag {
	var predictions []PredictiveTag

	// Current time-based predictions
	hour := context.Time.Format("15")
	dayOfWeek := context.Time.Weekday().String()

	// Check hourly patterns
	if hourTags, exists := pattern.TimeBasedPatterns[hour]; exists {
		for _, tag := range hourTags {
			if pte.containsString(context.ExistingTags, tag) {
				continue
			}

			predictions = append(predictions, PredictiveTag{
				Tag:            tag,
				Confidence:     0.6,
				PredictionType: "temporal",
				Reason:         fmt.Sprintf("Often used at %s:00", hour),
				UserRelevance:  0.7,
				ContextMatch:   0.5,
				Metadata: map[string]interface{}{
					"time_pattern": "hourly",
					"hour":         hour,
				},
			})
		}
	}

	// Check day-of-week patterns
	if dayTags, exists := pattern.TimeBasedPatterns[dayOfWeek]; exists {
		for _, tag := range dayTags {
			if pte.containsString(context.ExistingTags, tag) {
				continue
			}

			predictions = append(predictions, PredictiveTag{
				Tag:            tag,
				Confidence:     0.5,
				PredictionType: "temporal",
				Reason:         fmt.Sprintf("Often used on %s", dayOfWeek),
				UserRelevance:  0.6,
				ContextMatch:   0.4,
				Metadata: map[string]interface{}{
					"time_pattern": "weekly",
					"day":          dayOfWeek,
				},
			})
		}
	}

	return predictions
}

func (pte *PredictiveTagEngine) predictFromSequences(pattern *UserTaggingPattern, context PredictionContext) []PredictiveTag {
	var predictions []PredictiveTag

	if len(context.ExistingTags) == 0 {
		return predictions
	}

	// Find sequences that start with existing tags
	sequenceMatches := make(map[string]int)

	for _, sequence := range pattern.TagSequences {
		for _, existingTag := range context.ExistingTags {
			for i, seqTag := range sequence {
				if seqTag == existingTag && i < len(sequence)-1 {
					// Next tag in sequence
					nextTag := sequence[i+1]
					if !pte.containsString(context.ExistingTags, nextTag) {
						sequenceMatches[nextTag]++
					}
				}
			}
		}
	}

	// Convert to predictions
	type seqMatch struct {
		tag   string
		count int
	}

	var sortedMatches []seqMatch
	for tag, count := range sequenceMatches {
		sortedMatches = append(sortedMatches, seqMatch{tag, count})
	}

	sort.Slice(sortedMatches, func(i, j int) bool {
		return sortedMatches[i].count > sortedMatches[j].count
	})

	// Create predictions
	maxCount := 0
	if len(sortedMatches) > 0 {
		maxCount = sortedMatches[0].count
	}

	for _, sm := range sortedMatches {
		confidence := float64(sm.count) / float64(maxCount)
		
		predictions = append(predictions, PredictiveTag{
			Tag:            sm.tag,
			Confidence:     confidence,
			PredictionType: "sequence",
			Reason:         fmt.Sprintf("Follows your tagging patterns (%d times)", sm.count),
			UserRelevance:  confidence,
			ContextMatch:   0.6,
		})
	}

	return predictions
}

// Helper methods

func (pte *PredictiveTagEngine) getUserPattern(userID int) *UserTaggingPattern {
	if pattern, exists := pte.userPatterns[userID]; exists {
		return pattern
	}

	// Create new pattern for user
	pattern := &UserTaggingPattern{
		UserID:               userID,
		TagFrequency:         make(map[string]int),
		TagCooccurrence:      make(map[string]map[string]int),
		DomainTagPreferences: make(map[string][]string),
		TimeBasedPatterns:    make(map[string][]string),
		TagSequences:         [][]string{},
		RecentTags:           []TimestampedTag{},
		TagCategories:        make(map[string][]string),
		PersonalVocabulary:   []string{},
		LastUpdated:          time.Now(),
	}

	pte.userPatterns[userID] = pattern
	return pattern
}

func (pte *PredictiveTagEngine) mergeAndRankPredictions(predictions []PredictiveTag, context PredictionContext, pattern *UserTaggingPattern) []PredictiveTag {
	// Group predictions by tag
	tagGroups := make(map[string][]PredictiveTag)
	for _, pred := range predictions {
		tagGroups[pred.Tag] = append(tagGroups[pred.Tag], pred)
	}

	var mergedPredictions []PredictiveTag

	// Merge predictions for each tag
	for _, preds := range tagGroups {
		if len(preds) == 1 {
			mergedPredictions = append(mergedPredictions, preds[0])
		} else {
			// Combine multiple predictions for same tag
			merged := pte.combinePredictions(preds, pattern)
			mergedPredictions = append(mergedPredictions, merged)
		}
	}

	// Apply recency boost
	for i := range mergedPredictions {
		recencyBoost := pte.calculateRecencyBoost(mergedPredictions[i].Tag, pattern)
		mergedPredictions[i].Confidence *= (1.0 + recencyBoost)
		mergedPredictions[i].UserRelevance *= (1.0 + recencyBoost)
	}

	// Sort by combined score
	sort.Slice(mergedPredictions, func(i, j int) bool {
		scoreI := (mergedPredictions[i].Confidence * 0.4) +
				 (mergedPredictions[i].UserRelevance * 0.3) +
				 (mergedPredictions[i].ContextMatch * 0.3)
		scoreJ := (mergedPredictions[j].Confidence * 0.4) +
				 (mergedPredictions[j].UserRelevance * 0.3) +
				 (mergedPredictions[j].ContextMatch * 0.3)
		return scoreI > scoreJ
	})

	return mergedPredictions
}

func (pte *PredictiveTagEngine) combinePredictions(predictions []PredictiveTag, pattern *UserTaggingPattern) PredictiveTag {
	// Combine predictions using weighted average
	combined := predictions[0] // Start with first prediction
	
	totalWeight := 0.0
	weightedConfidence := 0.0
	weightedUserRelevance := 0.0
	weightedContextMatch := 0.0
	
	var allTypes []string
	var allReasons []string

	for _, pred := range predictions {
		weight := pte.getPredictionTypeWeight(pred.PredictionType)
		totalWeight += weight
		
		weightedConfidence += pred.Confidence * weight
		weightedUserRelevance += pred.UserRelevance * weight
		weightedContextMatch += pred.ContextMatch * weight
		
		allTypes = append(allTypes, pred.PredictionType)
		allReasons = append(allReasons, pred.Reason)
	}

	if totalWeight > 0 {
		combined.Confidence = weightedConfidence / totalWeight
		combined.UserRelevance = weightedUserRelevance / totalWeight
		combined.ContextMatch = weightedContextMatch / totalWeight
	}

	combined.PredictionType = strings.Join(allTypes, "+")
	combined.Reason = strings.Join(allReasons, "; ")

	return combined
}

func (pte *PredictiveTagEngine) getPredictionTypeWeight(predType string) float64 {
	weights := map[string]float64{
		"frequency":    0.8,
		"cooccurrence": 1.0,
		"domain":       0.9,
		"semantic":     0.7,
		"temporal":     0.6,
		"sequence":     0.8,
	}
	
	if weight, exists := weights[predType]; exists {
		return weight
	}
	return 0.5
}

func (pte *PredictiveTagEngine) calculateRecencyBoost(tag string, pattern *UserTaggingPattern) float64 {
	// Boost tags used recently
	for _, recentTag := range pattern.RecentTags {
		if recentTag.Tag == tag {
			daysSince := time.Since(recentTag.Timestamp).Hours() / 24
			if daysSince < 7 {
				return (7 - daysSince) / 7 * 0.2 // Up to 20% boost
			}
		}
	}
	return 0.0
}

func (pte *PredictiveTagEngine) filterPredictions(predictions []PredictiveTag, minConfidence float64, maxSuggestions int) []PredictiveTag {
	var filtered []PredictiveTag

	for _, pred := range predictions {
		if pred.Confidence >= minConfidence {
			filtered = append(filtered, pred)
		}
	}

	if len(filtered) > maxSuggestions {
		filtered = filtered[:maxSuggestions]
	}

	return filtered
}

func (pte *PredictiveTagEngine) calculateOverallConfidence(predictions []PredictiveTag) float64 {
	if len(predictions) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, pred := range predictions {
		totalConfidence += pred.Confidence
	}

	return totalConfidence / float64(len(predictions))
}

func (pte *PredictiveTagEngine) generatePredictionSummary(allPreds, finalPreds []PredictiveTag, pattern *UserTaggingPattern) PredictionSummary {
	summary := PredictionSummary{
		TotalCandidates: len(allPreds),
	}

	// Count by type and confidence
	typeCount := make(map[string]int)
	for _, pred := range finalPreds {
		if pred.Confidence >= 0.7 {
			summary.HighConfidenceCount++
		}
		
		for _, predType := range strings.Split(pred.PredictionType, "+") {
			typeCount[predType]++
		}
		
		// Check if tag is in user's vocabulary
		if !pte.containsString(pattern.PersonalVocabulary, pred.Tag) {
			summary.NovelSuggestions++
		}
	}

	summary.SemanticCount = typeCount["semantic"]
	summary.PatternBasedCount = typeCount["frequency"] + typeCount["cooccurrence"] + typeCount["sequence"]

	// Generate insights
	var insights []string
	
	if summary.HighConfidenceCount > 0 {
		insights = append(insights, fmt.Sprintf("%d high-confidence suggestions based on your patterns", summary.HighConfidenceCount))
	}
	
	if summary.NovelSuggestions > 0 {
		insights = append(insights, fmt.Sprintf("%d new tag suggestions to expand your vocabulary", summary.NovelSuggestions))
	}
	
	if summary.SemanticCount > 0 {
		insights = append(insights, fmt.Sprintf("%d AI-powered semantic suggestions", summary.SemanticCount))
	}

	summary.Insights = insights
	return summary
}

// Utility methods

func (pte *PredictiveTagEngine) extractDomain(url string) string {
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

func (pte *PredictiveTagEngine) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (pte *PredictiveTagEngine) countTagInDomain(pattern *UserTaggingPattern, tag, domain string) int {
	// This is simplified - in a real implementation, you'd track tag usage per domain
	domainTags := pattern.DomainTagPreferences[domain]
	for _, domainTag := range domainTags {
		if domainTag == tag {
			return pattern.TagFrequency[tag] / 2 // Rough estimate
		}
	}
	return 0
}