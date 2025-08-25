package ai

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ONNXInferenceEngine provides advanced content understanding using ONNX models
type ONNXInferenceEngine struct {
	modelPath        string
	textEncoder      *TextEncoder
	sentimentModel   *SentimentModel
	topicModel       *TopicModel
	qualityModel     *QualityModel
	isInitialized    bool
	supportedModels  []string
}

// TextEncoder handles text preprocessing for ONNX models
type TextEncoder struct {
	vocabulary   map[string]int
	maxSeqLength int
	padToken     int
	unkToken     int
}

// SentimentModel analyzes content sentiment
type SentimentModel struct {
	modelFile string
	labels    []string // ["negative", "neutral", "positive"]
}

// TopicModel performs topic classification
type TopicModel struct {
	modelFile string
	topics    []string
	threshold float64
}

// QualityModel assesses content quality
type QualityModel struct {
	modelFile string
	features  []string
	weights   map[string]float64
}

// ONNXPrediction represents a model prediction
type ONNXPrediction struct {
	ModelType   string             `json:"model_type"`
	Predictions []LabelConfidence  `json:"predictions"`
	TopResult   LabelConfidence    `json:"top_result"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LabelConfidence represents a label with confidence score
type LabelConfidence struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	Score      float64 `json:"raw_score"`
}

// ContentAnalysisResult represents comprehensive content analysis
type ContentAnalysisResult struct {
	Text            string                     `json:"text"`
	Sentiment       *ONNXPrediction            `json:"sentiment,omitempty"`
	Topics          *ONNXPrediction            `json:"topics,omitempty"`
	Quality         *QualityAssessment         `json:"quality,omitempty"`
	ContentFeatures *ContentFeatures           `json:"content_features,omitempty"`
	ProcessingTime  time.Duration              `json:"processing_time"`
	ModelVersions   map[string]string          `json:"model_versions,omitempty"`
}

// QualityAssessment provides detailed content quality metrics
type QualityAssessment struct {
	OverallScore    float64            `json:"overall_score"`
	Readability     float64            `json:"readability"`
	Informativeness float64            `json:"informativeness"`
	Credibility     float64            `json:"credibility"`
	Uniqueness      float64            `json:"uniqueness"`
	Engagement      float64            `json:"engagement"`
	TechnicalDepth  float64            `json:"technical_depth"`
	Factors         map[string]float64 `json:"factors"`
	Recommendations []string           `json:"recommendations,omitempty"`
}

// ContentFeatures represents extracted content features
type ContentFeatures struct {
	WordCount       int                `json:"word_count"`
	SentenceCount   int                `json:"sentence_count"`
	ParagraphCount  int                `json:"paragraph_count"`
	AvgWordsPerSent float64            `json:"avg_words_per_sentence"`
	ReadingLevel    string             `json:"reading_level"`
	KeyPhrases      []string           `json:"key_phrases"`
	NamedEntities   []NamedEntity      `json:"named_entities,omitempty"`
	TechnicalTerms  []string           `json:"technical_terms,omitempty"`
	LinkDensity     float64            `json:"link_density"`
	CodeBlocks      int                `json:"code_blocks"`
}

// NamedEntity represents a detected named entity
type NamedEntity struct {
	Text       string  `json:"text"`
	Type       string  `json:"type"` // PERSON, ORG, TECH, etc.
	Confidence float64 `json:"confidence"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
}

// NewONNXInferenceEngine creates a new ONNX inference engine
func NewONNXInferenceEngine(modelPath string) *ONNXInferenceEngine {
	return &ONNXInferenceEngine{
		modelPath: modelPath,
		textEncoder: &TextEncoder{
			vocabulary:   make(map[string]int),
			maxSeqLength: 512,
			padToken:     0,
			unkToken:     1,
		},
		sentimentModel: &SentimentModel{
			modelFile: filepath.Join(modelPath, "sentiment.onnx"),
			labels:    []string{"negative", "neutral", "positive"},
		},
		topicModel: &TopicModel{
			modelFile: filepath.Join(modelPath, "topics.onnx"),
			topics:    getDefaultTopics(),
			threshold: 0.1,
		},
		qualityModel: &QualityModel{
			modelFile: filepath.Join(modelPath, "quality.onnx"),
			features:  getQualityFeatures(),
			weights:   getQualityWeights(),
		},
		isInitialized:   false,
		supportedModels: []string{"sentiment", "topics", "quality"},
	}
}

// Initialize loads ONNX models and prepares the inference engine
func (oe *ONNXInferenceEngine) Initialize() error {
	// Create model directory if it doesn't exist
	if err := os.MkdirAll(oe.modelPath, 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}
	
	// Initialize text encoder
	if err := oe.initializeTextEncoder(); err != nil {
		return fmt.Errorf("failed to initialize text encoder: %w", err)
	}
	
	// Initialize models (use fallback implementations if ONNX models not available)
	if err := oe.initializeModels(); err != nil {
		return fmt.Errorf("failed to initialize models: %w", err)
	}
	
	oe.isInitialized = true
	return nil
}

// AnalyzeContent performs comprehensive content analysis using multiple models
func (oe *ONNXInferenceEngine) AnalyzeContent(text, title, url string) (*ContentAnalysisResult, error) {
	startTime := time.Now()
	
	if !oe.isInitialized {
		if err := oe.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize ONNX engine: %w", err)
		}
	}
	
	if text == "" && title == "" {
		return nil, fmt.Errorf("no content to analyze")
	}
	
	// Combine title and text for analysis
	fullText := title
	if text != "" {
		if fullText != "" {
			fullText += ". " + text
		} else {
			fullText = text
		}
	}
	
	result := &ContentAnalysisResult{
		Text:           fullText,
		ProcessingTime: 0,
		ModelVersions:  oe.getModelVersions(),
	}
	
	// Extract content features first
	result.ContentFeatures = oe.extractContentFeatures(fullText, url)
	
	// Perform sentiment analysis
	if sentiment, err := oe.analyzeSentiment(fullText); err == nil {
		result.Sentiment = sentiment
	}
	
	// Perform topic classification
	if topics, err := oe.classifyTopics(fullText); err == nil {
		result.Topics = topics
	}
	
	// Assess content quality
	if quality, err := oe.assessQuality(fullText, title, url, result.ContentFeatures); err == nil {
		result.Quality = quality
	}
	
	result.ProcessingTime = time.Since(startTime)
	return result, nil
}

// GetSupportedAnalysis returns list of supported analysis types
func (oe *ONNXInferenceEngine) GetSupportedAnalysis() []string {
	return []string{"sentiment", "topics", "quality", "features", "entities"}
}

// Internal methods

func (oe *ONNXInferenceEngine) initializeTextEncoder() error {
	// Load or create vocabulary
	vocabFile := filepath.Join(oe.modelPath, "vocab.json")
	
	if _, err := os.Stat(vocabFile); err == nil {
		// Load existing vocabulary
		return oe.loadVocabulary(vocabFile)
	}
	
	// Create default vocabulary for tech content
	oe.textEncoder.vocabulary = createDefaultVocabulary()
	
	// Save vocabulary
	return oe.saveVocabulary(vocabFile)
}

func (oe *ONNXInferenceEngine) initializeModels() error {
	// In a full implementation, this would load actual ONNX models
	// For now, we'll use fallback implementations
	
	// Check if ONNX models exist
	modelsExist := oe.checkModelsExist()
	
	if !modelsExist {
		// Create fallback models with reasonable defaults
		return oe.createFallbackModels()
	}
	
	// Load actual ONNX models (placeholder for real implementation)
	return oe.loadONNXModels()
}

func (oe *ONNXInferenceEngine) checkModelsExist() bool {
	for _, model := range oe.supportedModels {
		modelFile := filepath.Join(oe.modelPath, model+".onnx")
		if _, err := os.Stat(modelFile); err != nil {
			return false
		}
	}
	return true
}

func (oe *ONNXInferenceEngine) createFallbackModels() error {
	// Create model metadata files to indicate fallback mode
	for _, model := range oe.supportedModels {
		metaFile := filepath.Join(oe.modelPath, model+"_meta.json")
		metadata := map[string]interface{}{
			"model_type": model,
			"version":    "fallback-1.0",
			"created":    time.Now().Format(time.RFC3339),
			"mode":       "rule_based_fallback",
		}
		
		data, _ := json.MarshalIndent(metadata, "", "  ")
		if err := os.WriteFile(metaFile, data, 0644); err != nil {
			return fmt.Errorf("failed to create metadata for %s: %w", model, err)
		}
	}
	
	return nil
}

func (oe *ONNXInferenceEngine) loadONNXModels() error {
	// Placeholder for actual ONNX model loading
	// In real implementation, would use ONNX Runtime Go bindings
	fmt.Printf("Loading ONNX models from %s (placeholder)\n", oe.modelPath)
	return nil
}

func (oe *ONNXInferenceEngine) analyzeSentiment(text string) (*ONNXPrediction, error) {
	// Fallback rule-based sentiment analysis
	words := strings.Fields(strings.ToLower(text))
	
	positiveScore := 0.0
	negativeScore := 0.0
	
	// Simple sentiment lexicon
	positiveWords := map[string]float64{
		"good": 0.8, "great": 0.9, "excellent": 1.0, "amazing": 0.9, "awesome": 0.8,
		"love": 0.7, "like": 0.6, "best": 0.8, "perfect": 0.9, "wonderful": 0.8,
		"fantastic": 0.9, "outstanding": 0.9, "impressive": 0.7, "helpful": 0.6,
		"useful": 0.6, "easy": 0.5, "clear": 0.5, "simple": 0.5, "fast": 0.6,
		"efficient": 0.7, "powerful": 0.7, "innovative": 0.8, "creative": 0.7,
	}
	
	negativeWords := map[string]float64{
		"bad": 0.8, "terrible": 0.9, "awful": 1.0, "horrible": 0.9, "hate": 0.8,
		"worst": 0.9, "useless": 0.8, "broken": 0.7, "difficult": 0.6, "hard": 0.5,
		"slow": 0.6, "complex": 0.4, "confusing": 0.7, "frustrating": 0.8,
		"disappointing": 0.7, "poor": 0.6, "weak": 0.5, "limited": 0.4,
	}
	
	for _, word := range words {
		if score, exists := positiveWords[word]; exists {
			positiveScore += score
		}
		if score, exists := negativeWords[word]; exists {
			negativeScore += score
		}
	}
	
	// Calculate sentiment scores
	total := positiveScore + negativeScore
	if total == 0 {
		// Neutral
		return &ONNXPrediction{
			ModelType: "sentiment",
			Predictions: []LabelConfidence{
				{Label: "neutral", Confidence: 0.8, Score: 0.5},
				{Label: "positive", Confidence: 0.1, Score: 0.3},
				{Label: "negative", Confidence: 0.1, Score: 0.2},
			},
			TopResult: LabelConfidence{Label: "neutral", Confidence: 0.8, Score: 0.5},
		}, nil
	}
	
	posConfidence := positiveScore / total
	negConfidence := negativeScore / total
	neuConfidence := 1.0 - math.Abs(posConfidence-negConfidence)
	
	predictions := []LabelConfidence{
		{Label: "positive", Confidence: posConfidence, Score: positiveScore},
		{Label: "negative", Confidence: negConfidence, Score: negativeScore},
		{Label: "neutral", Confidence: neuConfidence, Score: 1.0 - total},
	}
	
	// Sort by confidence
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Confidence > predictions[j].Confidence
	})
	
	return &ONNXPrediction{
		ModelType:   "sentiment",
		Predictions: predictions,
		TopResult:   predictions[0],
		Metadata: map[string]interface{}{
			"positive_score": positiveScore,
			"negative_score": negativeScore,
			"word_count":     len(words),
		},
	}, nil
}

func (oe *ONNXInferenceEngine) classifyTopics(text string) (*ONNXPrediction, error) {
	words := strings.Fields(strings.ToLower(text))
	wordSet := make(map[string]bool)
	for _, word := range words {
		wordSet[word] = true
	}
	
	// Topic keywords
	topicKeywords := map[string][]string{
		"programming":     {"code", "coding", "programming", "development", "software", "algorithm", "function"},
		"web-development": {"web", "html", "css", "javascript", "frontend", "backend", "api", "http"},
		"machine-learning": {"machine", "learning", "ai", "neural", "network", "data", "model", "training"},
		"database":        {"database", "sql", "nosql", "data", "storage", "query", "table", "index"},
		"security":        {"security", "encryption", "authentication", "vulnerability", "cyber", "protection"},
		"mobile":          {"mobile", "app", "ios", "android", "smartphone", "tablet", "native"},
		"design":          {"design", "ui", "ux", "interface", "user", "experience", "graphics", "visual"},
		"business":        {"business", "startup", "company", "market", "strategy", "revenue", "growth"},
		"science":         {"science", "research", "study", "analysis", "experiment", "theory", "scientific"},
		"education":       {"education", "learning", "tutorial", "course", "lesson", "teach", "student"},
	}
	
	var topicScores []LabelConfidence
	
	for topic, keywords := range topicKeywords {
		score := 0.0
		matches := 0
		
		for _, keyword := range keywords {
			if wordSet[keyword] {
				score += 1.0
				matches++
			}
		}
		
		if matches > 0 {
			confidence := score / float64(len(keywords))
			topicScores = append(topicScores, LabelConfidence{
				Label:      topic,
				Confidence: confidence,
				Score:      score,
			})
		}
	}
	
	// Sort by confidence
	sort.Slice(topicScores, func(i, j int) bool {
		return topicScores[i].Confidence > topicScores[j].Confidence
	})
	
	// Filter by threshold
	var filteredScores []LabelConfidence
	for _, score := range topicScores {
		if score.Confidence > oe.topicModel.threshold {
			filteredScores = append(filteredScores, score)
		}
	}
	
	result := &ONNXPrediction{
		ModelType:   "topics",
		Predictions: filteredScores,
	}
	
	if len(filteredScores) > 0 {
		result.TopResult = filteredScores[0]
	}
	
	return result, nil
}

func (oe *ONNXInferenceEngine) assessQuality(text, title, url string, features *ContentFeatures) (*QualityAssessment, error) {
	assessment := &QualityAssessment{
		Factors:         make(map[string]float64),
		Recommendations: []string{},
	}
	
	// Assess different quality factors
	assessment.Readability = oe.assessReadability(features)
	assessment.Informativeness = oe.assessInformativeness(text, features)
	assessment.Credibility = oe.assessCredibility(url, features)
	assessment.Uniqueness = oe.assessUniqueness(text)
	assessment.Engagement = oe.assessEngagement(text, features)
	assessment.TechnicalDepth = oe.assessTechnicalDepth(text, features)
	
	// Store individual factors
	assessment.Factors["readability"] = assessment.Readability
	assessment.Factors["informativeness"] = assessment.Informativeness
	assessment.Factors["credibility"] = assessment.Credibility
	assessment.Factors["uniqueness"] = assessment.Uniqueness
	assessment.Factors["engagement"] = assessment.Engagement
	assessment.Factors["technical_depth"] = assessment.TechnicalDepth
	
	// Calculate weighted overall score
	weights := oe.qualityModel.weights
	assessment.OverallScore = 
		assessment.Readability * weights["readability"] +
		assessment.Informativeness * weights["informativeness"] +
		assessment.Credibility * weights["credibility"] +
		assessment.Uniqueness * weights["uniqueness"] +
		assessment.Engagement * weights["engagement"] +
		assessment.TechnicalDepth * weights["technical_depth"]
	
	// Generate recommendations
	assessment.Recommendations = oe.generateQualityRecommendations(assessment)
	
	return assessment, nil
}

func (oe *ONNXInferenceEngine) extractContentFeatures(text, url string) *ContentFeatures {
	words := strings.Fields(text)
	sentences := strings.Split(text, ".")
	paragraphs := strings.Split(text, "\n\n")
	
	features := &ContentFeatures{
		WordCount:      len(words),
		SentenceCount:  len(sentences),
		ParagraphCount: len(paragraphs),
		KeyPhrases:     oe.extractKeyPhrases(text),
		TechnicalTerms: oe.extractTechnicalTerms(text),
		CodeBlocks:     oe.countCodeBlocks(text),
	}
	
	if len(sentences) > 0 {
		features.AvgWordsPerSent = float64(len(words)) / float64(len(sentences))
	}
	
	features.ReadingLevel = oe.calculateReadingLevel(features)
	features.LinkDensity = oe.calculateLinkDensity(text)
	
	return features
}

// Helper methods for quality assessment

func (oe *ONNXInferenceEngine) assessReadability(features *ContentFeatures) float64 {
	if features.AvgWordsPerSent == 0 {
		return 0.5
	}
	
	// Simple readability score based on average sentence length
	avgWordsPerSent := features.AvgWordsPerSent
	
	if avgWordsPerSent <= 15 {
		return 0.9 // Easy to read
	} else if avgWordsPerSent <= 25 {
		return 0.7 // Moderate
	} else {
		return 0.4 // Difficult
	}
}

func (oe *ONNXInferenceEngine) assessInformativeness(text string, features *ContentFeatures) float64 {
	score := 0.0
	
	// Word count factor
	if features.WordCount > 500 {
		score += 0.3
	} else if features.WordCount > 200 {
		score += 0.2
	}
	
	// Technical terms factor
	if len(features.TechnicalTerms) > 5 {
		score += 0.2
	}
	
	// Code blocks factor
	if features.CodeBlocks > 0 {
		score += 0.2
	}
	
	// Key phrases factor
	if len(features.KeyPhrases) > 3 {
		score += 0.3
	}
	
	return math.Min(1.0, score)
}

func (oe *ONNXInferenceEngine) assessCredibility(url string, features *ContentFeatures) float64 {
	score := 0.5 // Base score
	
	// Domain credibility
	if strings.Contains(url, ".edu") || strings.Contains(url, ".gov") {
		score += 0.3
	} else if strings.Contains(url, "github.com") || strings.Contains(url, "stackoverflow.com") {
		score += 0.2
	}
	
	// Content length factor
	if features.WordCount > 300 {
		score += 0.1
	}
	
	// Structure factor
	if features.ParagraphCount > 2 {
		score += 0.1
	}
	
	return math.Min(1.0, score)
}

func (oe *ONNXInferenceEngine) assessUniqueness(text string) float64 {
	// Simple uniqueness assessment based on vocabulary diversity
	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		return 0.0
	}
	
	uniqueWords := make(map[string]bool)
	for _, word := range words {
		uniqueWords[word] = true
	}
	
	diversity := float64(len(uniqueWords)) / float64(len(words))
	return math.Min(1.0, diversity * 1.5) // Scale up diversity score
}

func (oe *ONNXInferenceEngine) assessEngagement(text string, features *ContentFeatures) float64 {
	score := 0.0
	
	// Question marks (engagement indicator)
	questionMarks := strings.Count(text, "?")
	if questionMarks > 0 {
		score += 0.2
	}
	
	// Exclamation marks (enthusiasm)
	exclamationMarks := strings.Count(text, "!")
	if exclamationMarks > 0 && exclamationMarks < 5 { // Not too many
		score += 0.1
	}
	
	// Code blocks (interactive content)
	if features.CodeBlocks > 0 {
		score += 0.3
	}
	
	// Reading level (not too difficult)
	if features.ReadingLevel == "intermediate" {
		score += 0.2
	} else if features.ReadingLevel == "beginner" {
		score += 0.2
	}
	
	return math.Min(1.0, score)
}

func (oe *ONNXInferenceEngine) assessTechnicalDepth(text string, features *ContentFeatures) float64 {
	score := 0.0
	
	// Technical terms density
	if len(features.TechnicalTerms) > 0 {
		density := float64(len(features.TechnicalTerms)) / float64(features.WordCount) * 100
		if density > 10 {
			score += 0.4 // High technical density
		} else if density > 5 {
			score += 0.3 // Moderate technical density
		} else {
			score += 0.1 // Low technical density
		}
	}
	
	// Code blocks
	if features.CodeBlocks > 2 {
		score += 0.3
	} else if features.CodeBlocks > 0 {
		score += 0.2
	}
	
	// Length factor
	if features.WordCount > 1000 {
		score += 0.3
	}
	
	return math.Min(1.0, score)
}

func (oe *ONNXInferenceEngine) extractKeyPhrases(text string) []string {
	// Simple key phrase extraction based on frequency and positioning
	words := strings.Fields(strings.ToLower(text))
	phrases := []string{}
	
	// Look for 2-3 word combinations that appear multiple times
	for i := 0; i < len(words)-1; i++ {
		if len(words[i]) > 3 && len(words[i+1]) > 3 {
			phrase := words[i] + " " + words[i+1]
			if strings.Count(text, phrase) > 1 {
				phrases = append(phrases, phrase)
			}
		}
	}
	
	// Remove duplicates and return top phrases
	unique := make(map[string]bool)
	var result []string
	for _, phrase := range phrases {
		if !unique[phrase] && len(result) < 5 {
			unique[phrase] = true
			result = append(result, phrase)
		}
	}
	
	return result
}

func (oe *ONNXInferenceEngine) extractTechnicalTerms(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	
	technicalTerms := map[string]bool{
		"api": true, "algorithm": true, "database": true, "framework": true,
		"library": true, "function": true, "method": true, "class": true,
		"object": true, "array": true, "string": true, "integer": true,
		"boolean": true, "json": true, "xml": true, "http": true, "https": true,
		"css": true, "html": true, "javascript": true, "python": true,
		"java": true, "golang": true, "rust": true, "typescript": true,
		"react": true, "vue": true, "angular": true, "node": true,
		"docker": true, "kubernetes": true, "aws": true, "cloud": true,
		"microservices": true, "authentication": true, "authorization": true,
		"encryption": true, "security": true, "vulnerability": true,
	}
	
	var terms []string
	for _, word := range words {
		if technicalTerms[word] {
			terms = append(terms, word)
		}
	}
	
	return terms
}

func (oe *ONNXInferenceEngine) countCodeBlocks(text string) int {
	// Count code blocks (simplified)
	codeMarkers := []string{"```", "    ", "\t"}
	count := 0
	
	for _, marker := range codeMarkers {
		count += strings.Count(text, marker)
	}
	
	return count / 2 // Assume pairs of markers
}

func (oe *ONNXInferenceEngine) calculateReadingLevel(features *ContentFeatures) string {
	avgWords := features.AvgWordsPerSent
	
	if avgWords <= 12 {
		return "beginner"
	} else if avgWords <= 18 {
		return "intermediate"
	} else {
		return "advanced"
	}
}

func (oe *ONNXInferenceEngine) calculateLinkDensity(text string) float64 {
	httpCount := strings.Count(text, "http")
	words := len(strings.Fields(text))
	
	if words == 0 {
		return 0.0
	}
	
	return float64(httpCount) / float64(words) * 100
}

func (oe *ONNXInferenceEngine) generateQualityRecommendations(assessment *QualityAssessment) []string {
	var recommendations []string
	
	if assessment.Readability < 0.5 {
		recommendations = append(recommendations, "Consider shorter sentences for better readability")
	}
	
	if assessment.Informativeness < 0.4 {
		recommendations = append(recommendations, "Add more detailed information and examples")
	}
	
	if assessment.TechnicalDepth < 0.3 {
		recommendations = append(recommendations, "Include more technical details and code examples")
	}
	
	if assessment.Engagement < 0.4 {
		recommendations = append(recommendations, "Add interactive elements or questions to increase engagement")
	}
	
	return recommendations
}

// Utility methods

func (oe *ONNXInferenceEngine) loadVocabulary(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, &oe.textEncoder.vocabulary)
}

func (oe *ONNXInferenceEngine) saveVocabulary(filename string) error {
	data, err := json.MarshalIndent(oe.textEncoder.vocabulary, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

func (oe *ONNXInferenceEngine) getModelVersions() map[string]string {
	return map[string]string{
		"sentiment": "fallback-1.0",
		"topics":    "fallback-1.0",
		"quality":   "fallback-1.0",
		"engine":    "onnx-fallback-1.0",
	}
}

// Default data

func getDefaultTopics() []string {
	return []string{
		"programming", "web-development", "machine-learning", "database",
		"security", "mobile", "design", "business", "science", "education",
		"tutorial", "reference", "news", "blog", "documentation", "tool",
	}
}

func getQualityFeatures() []string {
	return []string{
		"readability", "informativeness", "credibility", "uniqueness",
		"engagement", "technical_depth",
	}
}

func getQualityWeights() map[string]float64 {
	return map[string]float64{
		"readability":     0.15,
		"informativeness": 0.25,
		"credibility":     0.20,
		"uniqueness":      0.15,
		"engagement":      0.15,
		"technical_depth": 0.10,
	}
}

func createDefaultVocabulary() map[string]int {
	words := []string{
		"<pad>", "<unk>", "the", "and", "for", "are", "but", "not", "you", "all",
		"code", "coding", "development", "web", "app", "software", "tutorial",
		"javascript", "python", "react", "api", "database", "framework", "library",
		"design", "ui", "ux", "html", "css", "programming", "algorithm", "function",
		"security", "performance", "testing", "documentation", "guide", "reference",
	}
	
	vocab := make(map[string]int)
	for i, word := range words {
		vocab[word] = i
	}
	
	return vocab
}