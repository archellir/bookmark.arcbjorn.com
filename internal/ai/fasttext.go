package ai

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FastTextClassifier provides lightweight text classification using FastText-style approach
type FastTextClassifier struct {
	vocabulary    map[string]int     // word -> index
	wordVectors   [][]float32        // word embeddings
	labelVectors  [][]float32        // label embeddings  
	labels        []string           // label names
	ngrams        map[string]int     // n-gram -> index
	vectorSize    int                // embedding dimension
	threshold     float64            // classification threshold
	modelPath     string             // path to model files
	isInitialized bool               // model initialization status
}

// FastTextPrediction represents a classification result
type FastTextPrediction struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	Score      float64 `json:"score"`
}

// FastTextResult represents classification results for a text
type FastTextResult struct {
	Text        string                `json:"text"`
	Predictions []FastTextPrediction  `json:"predictions"`
	TopLabel    string                `json:"top_label"`
	TopScore    float64               `json:"top_score"`
	ProcessTime time.Duration         `json:"process_time"`
}

// NewFastTextClassifier creates a new FastText classifier
func NewFastTextClassifier(modelPath string) *FastTextClassifier {
	return &FastTextClassifier{
		vocabulary:    make(map[string]int),
		wordVectors:   [][]float32{},
		labelVectors:  [][]float32{},
		labels:        []string{},
		ngrams:        make(map[string]int),
		vectorSize:    100, // Default dimension
		threshold:     0.0,
		modelPath:     modelPath,
		isInitialized: false,
	}
}

// Initialize loads or creates the FastText model
func (ft *FastTextClassifier) Initialize() error {
	// Check if pre-trained model exists
	if ft.modelExists() {
		return ft.loadModel()
	}
	
	// Create default model with common bookmark categories
	return ft.createDefaultModel()
}

// ClassifyText classifies text into categories
func (ft *FastTextClassifier) ClassifyText(text string, k int) (*FastTextResult, error) {
	startTime := time.Now()
	
	if !ft.isInitialized {
		if err := ft.Initialize(); err != nil {
			return nil, fmt.Errorf("failed to initialize classifier: %w", err)
		}
	}
	
	if k <= 0 {
		k = 5 // Default top-k
	}
	
	// Preprocess text
	words := ft.preprocessText(text)
	if len(words) == 0 {
		return &FastTextResult{
			Text:        text,
			Predictions: []FastTextPrediction{},
			ProcessTime: time.Since(startTime),
		}, nil
	}
	
	// Get text vector representation
	textVector := ft.getTextVector(words)
	
	// Calculate similarities with all labels
	var predictions []FastTextPrediction
	for i, label := range ft.labels {
		if i < len(ft.labelVectors) {
			similarity := ft.cosineSimilarity(textVector, ft.labelVectors[i])
			confidence := ft.scoreToConfidence(similarity)
			
			predictions = append(predictions, FastTextPrediction{
				Label:      label,
				Confidence: confidence,
				Score:      similarity,
			})
		}
	}
	
	// Sort by score
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Score > predictions[j].Score
	})
	
	// Keep top-k predictions above threshold
	var filteredPredictions []FastTextPrediction
	for i, pred := range predictions {
		if i < k && pred.Score > ft.threshold {
			filteredPredictions = append(filteredPredictions, pred)
		}
	}
	
	result := &FastTextResult{
		Text:        text,
		Predictions: filteredPredictions,
		ProcessTime: time.Since(startTime),
	}
	
	if len(filteredPredictions) > 0 {
		result.TopLabel = filteredPredictions[0].Label
		result.TopScore = filteredPredictions[0].Score
	}
	
	return result, nil
}

// GetSupportedLabels returns all supported classification labels
func (ft *FastTextClassifier) GetSupportedLabels() []string {
	return ft.labels
}

// TrainFromData trains the classifier from labeled data
func (ft *FastTextClassifier) TrainFromData(trainingData []TrainingExample) error {
	if len(trainingData) == 0 {
		return fmt.Errorf("no training data provided")
	}
	
	// Build vocabulary from training data
	ft.buildVocabulary(trainingData)
	
	// Extract unique labels
	labelSet := make(map[string]bool)
	for _, example := range trainingData {
		for _, label := range example.Labels {
			labelSet[label] = true
		}
	}
	
	ft.labels = []string{}
	for label := range labelSet {
		ft.labels = append(ft.labels, label)
	}
	sort.Strings(ft.labels)
	
	// Initialize vectors
	ft.initializeVectors()
	
	// Simple training: average word vectors for each label
	ft.trainVectors(trainingData)
	
	ft.isInitialized = true
	
	// Save model
	return ft.saveModel()
}

// Training data structure
type TrainingExample struct {
	Text   string   `json:"text"`
	Labels []string `json:"labels"`
}

// Internal methods

func (ft *FastTextClassifier) modelExists() bool {
	vocabFile := filepath.Join(ft.modelPath, "vocab.txt")
	vectorFile := filepath.Join(ft.modelPath, "vectors.bin")
	labelFile := filepath.Join(ft.modelPath, "labels.txt")
	
	_, err1 := os.Stat(vocabFile)
	_, err2 := os.Stat(vectorFile)
	_, err3 := os.Stat(labelFile)
	
	return err1 == nil && err2 == nil && err3 == nil
}

func (ft *FastTextClassifier) createDefaultModel() error {
	// Create default bookmark classification categories
	ft.labels = []string{
		"programming", "web-development", "design", "tutorial", "reference",
		"news", "blog", "documentation", "tool", "framework", "library",
		"article", "guide", "javascript", "python", "react", "database",
		"api", "security", "performance", "testing", "devops", "ai",
		"machine-learning", "data-science", "mobile", "business", "finance",
		"productivity", "health", "science", "technology", "education",
		"entertainment", "gaming", "social", "media", "video", "music",
	}
	
	// Create basic vocabulary with common tech terms
	commonWords := []string{
		"code", "coding", "development", "web", "app", "application", "software",
		"tutorial", "guide", "learn", "learning", "documentation", "docs", "api",
		"framework", "library", "tool", "service", "database", "data", "algorithm",
		"javascript", "python", "react", "vue", "angular", "node", "golang", "rust",
		"html", "css", "design", "ui", "ux", "frontend", "backend", "fullstack",
		"security", "performance", "testing", "deploy", "devops", "cloud", "aws",
		"docker", "kubernetes", "git", "github", "open", "source", "project",
		"artificial", "intelligence", "machine", "learning", "neural", "network",
		"business", "startup", "finance", "productivity", "automation", "workflow",
	}
	
	// Build vocabulary
	ft.vocabulary = make(map[string]int)
	for i, word := range commonWords {
		ft.vocabulary[word] = i
	}
	
	// Initialize vectors with random values
	ft.vectorSize = 50 // Smaller for default model
	ft.initializeVectors()
	
	// Create simple rule-based label vectors
	ft.createRuleBasedVectors()
	
	ft.isInitialized = true
	
	// Use os.Root for safer file operations with security boundaries
	root, err := os.OpenRoot(".")
	if err != nil {
		// Fallback to standard operations if os.Root fails
		if err := os.MkdirAll(ft.modelPath, 0755); err != nil {
			return fmt.Errorf("failed to create model directory: %w", err)
		}
	} else {
		// Use os.Root for secure directory creation
		if err := root.MkdirAll(ft.modelPath, 0755); err != nil {
			return fmt.Errorf("failed to create model directory: %w", err)
		}
	}
	
	return ft.saveModel()
}

func (ft *FastTextClassifier) createRuleBasedVectors() {
	// Create label vectors based on keyword associations
	labelKeywords := map[string][]string{
		"programming":     {"code", "coding", "development", "software", "algorithm"},
		"web-development": {"web", "html", "css", "javascript", "frontend", "backend"},
		"design":          {"design", "ui", "ux", "graphics", "visual", "layout"},
		"tutorial":        {"tutorial", "guide", "learn", "learning", "howto"},
		"reference":       {"documentation", "docs", "reference", "api", "manual"},
		"framework":       {"framework", "library", "tool", "service"},
		"database":        {"database", "data", "sql", "storage", "query"},
		"javascript":      {"javascript", "js", "node", "react", "vue", "angular"},
		"python":          {"python", "django", "flask", "pandas", "numpy"},
		"ai":              {"artificial", "intelligence", "machine", "learning", "neural"},
		"business":        {"business", "startup", "finance", "productivity"},
	}
	
	ft.labelVectors = make([][]float32, len(ft.labels))
	
	for i, label := range ft.labels {
		vector := make([]float32, ft.vectorSize)
		
		// Find keywords for this label
		keywords := labelKeywords[label]
		if keywords == nil {
			// Default random vector for unknown labels
			for j := 0; j < ft.vectorSize; j++ {
				vector[j] = float32(0.1 * math.Sin(float64(i*j)))
			}
		} else {
			// Average word vectors for keywords
			validWords := 0
			for _, keyword := range keywords {
				if wordIdx, exists := ft.vocabulary[keyword]; exists {
					if wordIdx < len(ft.wordVectors) {
						for j := 0; j < ft.vectorSize; j++ {
							vector[j] += ft.wordVectors[wordIdx][j]
						}
						validWords++
					}
				}
			}
			
			// Normalize
			if validWords > 0 {
				for j := 0; j < ft.vectorSize; j++ {
					vector[j] /= float32(validWords)
				}
			}
		}
		
		ft.labelVectors[i] = vector
	}
}

func (ft *FastTextClassifier) preprocessText(text string) []string {
	// Simple preprocessing: lowercase, split, filter
	text = strings.ToLower(text)
	words := strings.Fields(text)
	
	var cleanWords []string
	for _, word := range words {
		// Remove punctuation and short words
		cleaned := strings.Trim(word, ".,!?;:\"'()[]{}/<>")
		if len(cleaned) > 2 && !ft.isStopWord(cleaned) {
			cleanWords = append(cleanWords, cleaned)
		}
	}
	
	return cleanWords
}

func (ft *FastTextClassifier) isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "and": true, "for": true, "are": true, "but": true,
		"not": true, "you": true, "all": true, "can": true, "had": true,
		"her": true, "was": true, "one": true, "our": true, "out": true,
		"day": true, "get": true, "has": true, "him": true, "his": true,
		"how": true, "man": true, "new": true, "now": true, "old": true,
		"see": true, "two": true, "way": true, "who": true, "boy": true,
		"did": true, "its": true, "let": true, "put": true, "say": true,
		"she": true, "too": true, "use": true, "this": true, "that": true,
		"with": true, "have": true, "from": true, "they": true, "know": true,
		"want": true, "been": true, "good": true, "much": true, "some": true,
		"time": true, "very": true, "when": true, "come": true, "here": true,
		"just": true, "like": true, "long": true, "make": true, "many": true,
		"over": true, "such": true, "take": true, "than": true, "them": true,
		"well": true, "were": true,
	}
	
	return stopWords[word]
}

func (ft *FastTextClassifier) getTextVector(words []string) []float32 {
	vector := make([]float32, ft.vectorSize)
	count := 0
	
	for _, word := range words {
		if wordIdx, exists := ft.vocabulary[word]; exists {
			if wordIdx < len(ft.wordVectors) {
				for i := 0; i < ft.vectorSize; i++ {
					vector[i] += ft.wordVectors[wordIdx][i]
				}
				count++
			}
		}
	}
	
	// Average and normalize
	if count > 0 {
		for i := 0; i < ft.vectorSize; i++ {
			vector[i] /= float32(count)
		}
	}
	
	return ft.normalizeVector(vector)
}

func (ft *FastTextClassifier) normalizeVector(vector []float32) []float32 {
	var norm float32
	for _, val := range vector {
		norm += val * val
	}
	norm = float32(math.Sqrt(float64(norm)))
	
	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}
	
	return vector
}

func (ft *FastTextClassifier) cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}
	
	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}
	
	if normA == 0 || normB == 0 {
		return 0.0
	}
	
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (ft *FastTextClassifier) scoreToConfidence(score float64) float64 {
	// Convert similarity score to confidence (0-1)
	// Apply sigmoid-like transformation
	return 1.0 / (1.0 + math.Exp(-5*(score-0.5)))
}

func (ft *FastTextClassifier) buildVocabulary(trainingData []TrainingExample) {
	wordCount := make(map[string]int)
	
	// Count word frequencies
	for _, example := range trainingData {
		words := ft.preprocessText(example.Text)
		for _, word := range words {
			wordCount[word]++
		}
	}
	
	// Keep words with minimum frequency
	minFreq := 2
	ft.vocabulary = make(map[string]int)
	idx := 0
	
	for word, count := range wordCount {
		if count >= minFreq {
			ft.vocabulary[word] = idx
			idx++
		}
	}
}

func (ft *FastTextClassifier) initializeVectors() {
	vocabSize := len(ft.vocabulary)
	labelSize := len(ft.labels)
	
	// Initialize word vectors with small random values
	ft.wordVectors = make([][]float32, vocabSize)
	for i := 0; i < vocabSize; i++ {
		vector := make([]float32, ft.vectorSize)
		for j := 0; j < ft.vectorSize; j++ {
			// Simple pseudo-random initialization
			vector[j] = float32(0.1 * math.Sin(float64(i*j+j)))
		}
		ft.wordVectors[i] = vector
	}
	
	// Initialize label vectors
	ft.labelVectors = make([][]float32, labelSize)
	for i := 0; i < labelSize; i++ {
		vector := make([]float32, ft.vectorSize)
		for j := 0; j < ft.vectorSize; j++ {
			vector[j] = float32(0.1 * math.Cos(float64(i*j+i)))
		}
		ft.labelVectors[i] = vector
	}
}

func (ft *FastTextClassifier) trainVectors(trainingData []TrainingExample) {
	// Simple training: average word vectors for each label
	labelWordVectors := make(map[string][][]float32)
	
	// Collect word vectors for each label
	for _, example := range trainingData {
		words := ft.preprocessText(example.Text)
		textVector := ft.getTextVector(words)
		
		for _, label := range example.Labels {
			if labelWordVectors[label] == nil {
				labelWordVectors[label] = [][]float32{}
			}
			labelWordVectors[label] = append(labelWordVectors[label], textVector)
		}
	}
	
	// Update label vectors by averaging
	for labelIdx, label := range ft.labels {
		if vectors, exists := labelWordVectors[label]; exists && len(vectors) > 0 {
			avgVector := make([]float32, ft.vectorSize)
			
			for _, vector := range vectors {
				for i := 0; i < ft.vectorSize; i++ {
					avgVector[i] += vector[i]
				}
			}
			
			// Average and normalize
			for i := 0; i < ft.vectorSize; i++ {
				avgVector[i] /= float32(len(vectors))
			}
			
			ft.labelVectors[labelIdx] = ft.normalizeVector(avgVector)
		}
	}
}

func (ft *FastTextClassifier) saveModel() error {
	// Use os.Root for safer file operations with security boundaries
	root, err := os.OpenRoot(".")
	if err != nil {
		// Fallback to standard file operations
		return ft.saveModelStandard()
	}
	
	// Save vocabulary using os.Root
	vocabFile := filepath.Join(ft.modelPath, "vocab.txt")
	file, err := root.OpenFile(vocabFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create vocab file: %w", err)
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	for word, idx := range ft.vocabulary {
		fmt.Fprintf(writer, "%s %d\n", word, idx)
	}
	writer.Flush()
	
	// Save labels
	labelFile := filepath.Join(ft.modelPath, "labels.txt")
	file, err = root.OpenFile(labelFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create labels file: %w", err)
	}
	defer file.Close()
	
	writer = bufio.NewWriter(file)
	for _, label := range ft.labels {
		fmt.Fprintf(writer, "%s\n", label)
	}
	writer.Flush()
	
	// Save model metadata
	metaFile := filepath.Join(ft.modelPath, "meta.txt")
	file, err = root.OpenFile(metaFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create meta file: %w", err)
	}
	defer file.Close()
	
	fmt.Fprintf(file, "vector_size %d\n", ft.vectorSize)
	fmt.Fprintf(file, "vocab_size %d\n", len(ft.vocabulary))
	fmt.Fprintf(file, "label_size %d\n", len(ft.labels))
	fmt.Fprintf(file, "threshold %f\n", ft.threshold)
	
	return nil
}

// saveModelStandard is a fallback for when os.Root is not available
func (ft *FastTextClassifier) saveModelStandard() error {
	// Save vocabulary
	vocabFile := filepath.Join(ft.modelPath, "vocab.txt")
	file, err := os.Create(vocabFile)
	if err != nil {
		return fmt.Errorf("failed to create vocab file: %w", err)
	}
	defer file.Close()
	
	writer := bufio.NewWriter(file)
	for word, idx := range ft.vocabulary {
		fmt.Fprintf(writer, "%s %d\n", word, idx)
	}
	writer.Flush()
	
	// Save labels
	labelFile := filepath.Join(ft.modelPath, "labels.txt")
	file, err = os.Create(labelFile)
	if err != nil {
		return fmt.Errorf("failed to create labels file: %w", err)
	}
	defer file.Close()
	
	writer = bufio.NewWriter(file)
	for _, label := range ft.labels {
		fmt.Fprintf(writer, "%s\n", label)
	}
	writer.Flush()
	
	// Save model metadata
	metaFile := filepath.Join(ft.modelPath, "meta.txt")
	file, err = os.Create(metaFile)
	if err != nil {
		return fmt.Errorf("failed to create meta file: %w", err)
	}
	defer file.Close()
	
	writer = bufio.NewWriter(file)
	fmt.Fprintf(writer, "vocab_size=%d\n", len(ft.vocabulary))
	fmt.Fprintf(writer, "labels=%d\n", len(ft.labels))
	fmt.Fprintf(writer, "vector_size=%d\n", ft.vectorSize)
	writer.Flush()
	
	return nil
}

func (ft *FastTextClassifier) loadModel() error {
	// Use os.Root for safer file operations with security boundaries
	root, err := os.OpenRoot(".")
	if err != nil {
		// Fallback to standard file operations
		return ft.loadModelStandard()
	}
	
	// Load vocabulary using os.Root
	vocabFile := filepath.Join(ft.modelPath, "vocab.txt")
	file, err := root.Open(vocabFile)
	if err != nil {
		return fmt.Errorf("failed to open vocab file: %w", err)
	}
	defer file.Close()
	
	ft.vocabulary = make(map[string]int)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) == 2 {
			word := parts[0]
			idx, err := strconv.Atoi(parts[1])
			if err == nil {
				ft.vocabulary[word] = idx
			}
		}
	}
	
	// Load labels
	labelFile := filepath.Join(ft.modelPath, "labels.txt")
	file, err = root.Open(labelFile)
	if err != nil {
		return fmt.Errorf("failed to open labels file: %w", err)
	}
	defer file.Close()
	
	ft.labels = []string{}
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		ft.labels = append(ft.labels, strings.TrimSpace(scanner.Text()))
	}
	
	// Load metadata
	metaFile := filepath.Join(ft.modelPath, "meta.txt")
	file, err = root.Open(metaFile)
	if err == nil {
		defer file.Close()
		scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			parts := strings.Fields(scanner.Text())
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				
				switch key {
				case "vector_size":
					if val, err := strconv.Atoi(value); err == nil {
						ft.vectorSize = val
					}
				case "threshold":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						ft.threshold = val
					}
				}
			}
		}
	}
	
	// Initialize vectors (would load from binary file in full implementation)
	ft.initializeVectors()
	ft.createRuleBasedVectors()
	
	ft.isInitialized = true
	return nil
}

// loadModelStandard is a fallback for when os.Root is not available
func (ft *FastTextClassifier) loadModelStandard() error {
	// Load vocabulary
	vocabFile := filepath.Join(ft.modelPath, "vocab.txt")
	file, err := os.Open(vocabFile)
	if err != nil {
		return fmt.Errorf("failed to open vocab file: %w", err)
	}
	defer file.Close()
	
	ft.vocabulary = make(map[string]int)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) == 2 {
			word := parts[0]
			idx, err := strconv.Atoi(parts[1])
			if err == nil {
				ft.vocabulary[word] = idx
			}
		}
	}
	
	// Load labels
	labelFile := filepath.Join(ft.modelPath, "labels.txt")
	file, err = os.Open(labelFile)
	if err != nil {
		return fmt.Errorf("failed to open labels file: %w", err)
	}
	defer file.Close()
	
	ft.labels = []string{}
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		ft.labels = append(ft.labels, strings.TrimSpace(scanner.Text()))
	}
	
	// Load metadata
	metaFile := filepath.Join(ft.modelPath, "meta.txt")
	file, err = os.Open(metaFile)
	if err == nil {
		defer file.Close()
		scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			parts := strings.Split(scanner.Text(), "=")
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				switch key {
				case "vector_size":
					if val, err := strconv.Atoi(value); err == nil {
						ft.vectorSize = val
					}
				case "threshold":
					if val, err := strconv.ParseFloat(value, 64); err == nil {
						ft.threshold = val
					}
				}
			}
		}
	}
	
	// Initialize vectors (would load from binary file in full implementation)
	ft.initializeVectors()
	ft.createRuleBasedVectors()
	
	ft.isInitialized = true
	return nil
}