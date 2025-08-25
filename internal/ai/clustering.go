package ai

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"torimemo/internal/models"
)

// BookmarkClusterer provides AI-powered bookmark clustering capabilities
type BookmarkClusterer struct {
	semanticAnalyzer *SemanticAnalyzer
	similarityEngine *SimilarityEngine
	// Clustering parameters
	minClusterSize   int
	maxClusters      int
	similarityThreshold float64
}

// NewBookmarkClusterer creates a new bookmark clustering engine
func NewBookmarkClusterer() *BookmarkClusterer {
	return &BookmarkClusterer{
		semanticAnalyzer:    NewSemanticAnalyzer(),
		similarityEngine:    NewSimilarityEngine(),
		minClusterSize:      3,  // Minimum bookmarks per cluster
		maxClusters:         20, // Maximum number of clusters
		similarityThreshold: 0.4, // Minimum similarity for clustering
	}
}

// BookmarkCluster represents a group of related bookmarks
type BookmarkCluster struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Bookmarks       []models.Bookmark  `json:"bookmarks"`
	CenterPoint     []float64          `json:"-"` // Internal cluster center
	CommonTags      []string           `json:"common_tags"`
	Themes          []ClusterTheme     `json:"themes"`
	Confidence      float64            `json:"confidence"`
	CreatedAt       time.Time          `json:"created_at"`
	Quality         ClusterQuality     `json:"quality"`
}

// ClusterTheme represents a semantic theme within a cluster
type ClusterTheme struct {
	Name        string    `json:"name"`
	Keywords    []string  `json:"keywords"`
	Strength    float64   `json:"strength"`    // 0-1, how strong this theme is
	Coverage    float64   `json:"coverage"`    // 0-1, how many bookmarks match this theme
}

// ClusterQuality provides metrics about cluster quality
type ClusterQuality struct {
	Cohesion     float64 `json:"cohesion"`      // How similar bookmarks within cluster are
	Separation   float64 `json:"separation"`    // How different this cluster is from others
	Silhouette   float64 `json:"silhouette"`    // Combined quality metric
	Purity       float64 `json:"purity"`        // Domain/topic purity
}

// ClusteringRequest represents a request to cluster bookmarks
type ClusteringRequest struct {
	BookmarkIDs     []int   `json:"bookmark_ids,omitempty"`  // Specific bookmarks to cluster
	UserID          int     `json:"user_id"`
	MinClusterSize  int     `json:"min_cluster_size,omitempty"`
	MaxClusters     int     `json:"max_clusters,omitempty"`
	SimilarityThreshold float64 `json:"similarity_threshold,omitempty"`
	ClusteringMethod string  `json:"clustering_method,omitempty"` // "semantic", "domain", "hybrid"
}

// ClusteringConfig represents configuration for clustering
type ClusteringConfig struct {
	Method              string  `json:"method"`
	MinClusterSize      int     `json:"min_cluster_size"`
	MaxClusters         int     `json:"max_clusters"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
}

// ClusteringResult represents the result of bookmark clustering
type ClusteringResult struct {
	Clusters        []BookmarkCluster `json:"clusters"`
	Unclustered     []models.Bookmark `json:"unclustered"`
	TotalBookmarks  int               `json:"total_bookmarks"`
	ClusterCount    int               `json:"cluster_count"`
	QualityScore    float64           `json:"quality_score"`
	ProcessingTime  time.Duration     `json:"processing_time"`
	Method          string            `json:"method"`
	Summary         ClusteringSummary `json:"summary"`
}

// ClusteringSummary provides high-level insights about the clustering
type ClusteringSummary struct {
	TopDomains      map[string]int    `json:"top_domains"`
	TopTags         map[string]int    `json:"top_tags"`
	TopCategories   map[string]int    `json:"top_categories"`
	TimeDistribution map[string]int   `json:"time_distribution"` // bookmarks by month/year
	Insights        []string          `json:"insights"`
}

// ClusterBookmarksWithConfig performs intelligent clustering with custom configuration
func (bc *BookmarkClusterer) ClusterBookmarksWithConfig(bookmarks []models.Bookmark, config ClusteringConfig) (*ClusteringResult, error) {
	// Apply config to clusterer
	if config.MinClusterSize > 0 {
		bc.minClusterSize = config.MinClusterSize
	}
	if config.MaxClusters > 0 {
		bc.maxClusters = config.MaxClusters
	}
	if config.SimilarityThreshold > 0 {
		bc.similarityThreshold = config.SimilarityThreshold
	}
	
	return bc.ClusterBookmarks(bookmarks, config.Method)
}

// ClusterBookmarks performs intelligent clustering of bookmarks
func (bc *BookmarkClusterer) ClusterBookmarks(bookmarks []models.Bookmark, method string) (*ClusteringResult, error) {
	startTime := time.Now()
	
	result := &ClusteringResult{
		TotalBookmarks: len(bookmarks),
		Method:         method,
		ProcessingTime: 0,
	}

	if len(bookmarks) < bc.minClusterSize {
		result.Unclustered = bookmarks
		result.Summary = bc.generateSummary(bookmarks, nil)
		result.ProcessingTime = time.Since(startTime)
		return result, nil
	}

	var clusters []BookmarkCluster
	var err error

	switch method {
	case "semantic":
		clusters, err = bc.semanticClustering(bookmarks)
	case "domain":
		clusters, err = bc.domainClustering(bookmarks)
	case "hybrid":
		clusters, err = bc.hybridClustering(bookmarks)
	default:
		clusters, err = bc.hybridClustering(bookmarks) // Default to hybrid
	}

	if err != nil {
		return nil, err
	}

	// Post-process clusters
	clusters = bc.refineAndNameClusters(clusters)
	
	// Separate unclustered bookmarks
	clusteredIDs := make(map[int]bool)
	for _, cluster := range clusters {
		for _, bookmark := range cluster.Bookmarks {
			clusteredIDs[bookmark.ID] = true
		}
	}

	for _, bookmark := range bookmarks {
		if !clusteredIDs[bookmark.ID] {
			result.Unclustered = append(result.Unclustered, bookmark)
		}
	}

	result.Clusters = clusters
	result.ClusterCount = len(clusters)
	result.QualityScore = bc.calculateOverallQuality(clusters)
	result.Summary = bc.generateSummary(bookmarks, clusters)
	result.ProcessingTime = time.Since(startTime)

	return result, nil
}

// semanticClustering groups bookmarks by semantic similarity
func (bc *BookmarkClusterer) semanticClustering(bookmarks []models.Bookmark) ([]BookmarkCluster, error) {
	// Create feature vectors for each bookmark
	vectors := make(map[int][]float64)
	for _, bookmark := range bookmarks {
		vector := bc.createSemanticVector(bookmark)
		vectors[bookmark.ID] = vector
	}

	// Use k-means style clustering
	return bc.performKMeansClustering(bookmarks, vectors)
}

// domainClustering groups bookmarks by domain and URL patterns
func (bc *BookmarkClusterer) domainClustering(bookmarks []models.Bookmark) ([]BookmarkCluster, error) {
	domainGroups := make(map[string][]models.Bookmark)
	
	// Group by domain
	for _, bookmark := range bookmarks {
		domain := bc.extractDomain(bookmark.URL)
		domainGroups[domain] = append(domainGroups[domain], bookmark)
	}

	var clusters []BookmarkCluster
	clusterID := 0

	// Convert domain groups to clusters
	for domain, domainBookmarks := range domainGroups {
		if len(domainBookmarks) >= bc.minClusterSize {
			cluster := BookmarkCluster{
				ID:        generateClusterID(clusterID),
				Name:      bc.generateDomainClusterName(domain, domainBookmarks),
				Bookmarks: domainBookmarks,
				CreatedAt: time.Now(),
				CommonTags: bc.extractCommonTags(domainBookmarks),
			}
			cluster.Confidence = bc.calculateDomainClusterConfidence(domainBookmarks)
			cluster.Quality = bc.calculateClusterQuality(cluster, clusters)
			cluster.Themes = bc.extractClusterThemes(domainBookmarks)
			
			clusters = append(clusters, cluster)
			clusterID++
		}
	}

	return clusters, nil
}

// hybridClustering combines semantic and domain-based approaches
func (bc *BookmarkClusterer) hybridClustering(bookmarks []models.Bookmark) ([]BookmarkCluster, error) {
	// Start with domain clustering for strong domain patterns
	domainClusters, err := bc.domainClustering(bookmarks)
	if err != nil {
		return nil, err
	}

	// Extract bookmarks that weren't strongly clustered by domain
	stronglyClusteredIDs := make(map[int]bool)
	var finalClusters []BookmarkCluster

	// Keep high-confidence domain clusters
	for _, cluster := range domainClusters {
		if cluster.Confidence > 0.7 && len(cluster.Bookmarks) >= bc.minClusterSize {
			finalClusters = append(finalClusters, cluster)
			for _, bookmark := range cluster.Bookmarks {
				stronglyClusteredIDs[bookmark.ID] = true
			}
		}
	}

	// Collect remaining bookmarks for semantic clustering
	var remainingBookmarks []models.Bookmark
	for _, bookmark := range bookmarks {
		if !stronglyClusteredIDs[bookmark.ID] {
			remainingBookmarks = append(remainingBookmarks, bookmark)
		}
	}

	// Apply semantic clustering to remaining bookmarks
	if len(remainingBookmarks) >= bc.minClusterSize {
		semanticClusters, err := bc.semanticClustering(remainingBookmarks)
		if err == nil {
			finalClusters = append(finalClusters, semanticClusters...)
		}
	}

	return finalClusters, nil
}

// performKMeansClustering implements simplified k-means clustering
func (bc *BookmarkClusterer) performKMeansClustering(bookmarks []models.Bookmark, vectors map[int][]float64) ([]BookmarkCluster, error) {
	if len(bookmarks) < bc.minClusterSize {
		return nil, nil
	}

	// Determine number of clusters (simplified)
	k := int(math.Min(float64(bc.maxClusters), math.Max(2, float64(len(bookmarks))/5)))
	
	// Initialize random centroids
	centroids := bc.initializeCentroids(vectors, k)
	
	// Perform clustering iterations
	maxIterations := 10
	var clusters []BookmarkCluster
	var assignments map[int][]models.Bookmark

	for iteration := 0; iteration < maxIterations; iteration++ {
		// Assign bookmarks to closest centroids
		assignments = make(map[int][]models.Bookmark)
		
		for _, bookmark := range bookmarks {
			if vector, exists := vectors[bookmark.ID]; exists {
				closestCentroid := bc.findClosestCentroid(vector, centroids)
				assignments[closestCentroid] = append(assignments[closestCentroid], bookmark)
			}
		}

		// Update centroids
		newCentroids := bc.updateCentroids(assignments, vectors)
		
		// Check for convergence (simplified)
		if bc.centroidsConverged(centroids, newCentroids) {
			break
		}
		centroids = newCentroids
	}

	// Convert assignments to clusters
	clusterID := 0
	for centroidID, clusterBookmarks := range assignments {
		if len(clusterBookmarks) >= bc.minClusterSize {
			cluster := BookmarkCluster{
				ID:          generateClusterID(clusterID),
				Name:        bc.generateSemanticClusterName(clusterBookmarks),
				Bookmarks:   clusterBookmarks,
				CenterPoint: centroids[centroidID],
				CreatedAt:   time.Now(),
				CommonTags:  bc.extractCommonTags(clusterBookmarks),
				Themes:      bc.extractClusterThemes(clusterBookmarks),
			}
			cluster.Confidence = bc.calculateSemanticClusterConfidence(cluster, vectors)
			clusters = append(clusters, cluster)
			clusterID++
		}
	}

	// Calculate quality metrics for all clusters
	for i := range clusters {
		clusters[i].Quality = bc.calculateClusterQuality(clusters[i], clusters)
	}

	return clusters, nil
}

// Helper methods for clustering

func (bc *BookmarkClusterer) createSemanticVector(bookmark models.Bookmark) []float64 {
	// Combine title, description, and URL path for analysis
	content := bookmark.Title
	if bookmark.Description != nil {
		content += " " + *bookmark.Description
	}
	
	// Get semantic suggestions as a proxy for content vector
	semanticSuggestions := bc.semanticAnalyzer.AnalyzeSemanticContent(bookmark.Title, 
		func() string {
			if bookmark.Description != nil {
				return *bookmark.Description
			}
			return ""
		}(), bookmark.URL)
	
	// Convert to vector (simplified approach)
	vector := make([]float64, 50)
	
	// Use tags and semantic suggestions to populate vector
	for _, tag := range bookmark.Tags {
		hash := bc.simpleHash(tag.Name)
		vector[hash%50] += 1.0
	}
	
	for _, semSugg := range semanticSuggestions {
		hash := bc.simpleHash(semSugg.Tag)
		vector[hash%50] += semSugg.Confidence
	}
	
	// Normalize vector
	return bc.normalizeVector(vector)
}

func (bc *BookmarkClusterer) extractDomain(url string) string {
	// Simple domain extraction
	url = strings.ToLower(url)
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}
	
	// Remove www prefix
	if strings.HasPrefix(url, "www.") {
		url = url[4:]
	}
	
	return url
}

func (bc *BookmarkClusterer) extractCommonTags(bookmarks []models.Bookmark) []string {
	tagCount := make(map[string]int)
	
	for _, bookmark := range bookmarks {
		for _, tag := range bookmark.Tags {
			tagCount[tag.Name]++
		}
	}
	
	// Find tags that appear in at least 30% of bookmarks
	minCount := int(math.Max(1, float64(len(bookmarks))*0.3))
	var commonTags []string
	
	for tag, count := range tagCount {
		if count >= minCount {
			commonTags = append(commonTags, tag)
		}
	}
	
	// Sort by frequency
	sort.Slice(commonTags, func(i, j int) bool {
		return tagCount[commonTags[i]] > tagCount[commonTags[j]]
	})
	
	// Return top 10
	if len(commonTags) > 10 {
		commonTags = commonTags[:10]
	}
	
	return commonTags
}

func (bc *BookmarkClusterer) extractClusterThemes(bookmarks []models.Bookmark) []ClusterTheme {
	var themes []ClusterTheme
	
	// Analyze common domains
	domainCount := make(map[string]int)
	for _, bookmark := range bookmarks {
		domain := bc.extractDomain(bookmark.URL)
		domainCount[domain]++
	}
	
	// Create domain theme if significant
	for domain, count := range domainCount {
		if float64(count)/float64(len(bookmarks)) > 0.5 {
			themes = append(themes, ClusterTheme{
				Name:     domain + " resources",
				Keywords: []string{domain},
				Strength: float64(count) / float64(len(bookmarks)),
				Coverage: float64(count) / float64(len(bookmarks)),
			})
		}
	}
	
	// Analyze common tag patterns
	tagCount := make(map[string]int)
	for _, bookmark := range bookmarks {
		for _, tag := range bookmark.Tags {
			tagCount[tag.Name]++
		}
	}
	
	var topTags []string
	for tag, count := range tagCount {
		if count >= 2 {
			topTags = append(topTags, tag)
		}
	}
	
	sort.Slice(topTags, func(i, j int) bool {
		return tagCount[topTags[i]] > tagCount[topTags[j]]
	})
	
	if len(topTags) > 0 {
		themeTag := topTags[0]
		themes = append(themes, ClusterTheme{
			Name:     themeTag + " content",
			Keywords: topTags[:int(math.Min(float64(len(topTags)), 5))],
			Strength: float64(tagCount[themeTag]) / float64(len(bookmarks)),
			Coverage: float64(tagCount[themeTag]) / float64(len(bookmarks)),
		})
	}
	
	return themes
}

func (bc *BookmarkClusterer) simpleHash(s string) int {
	hash := 0
	for _, r := range s {
		hash = hash*31 + int(r)
	}
	if hash < 0 {
		hash = -hash
	}
	return hash
}

func (bc *BookmarkClusterer) normalizeVector(vector []float64) []float64 {
	var sum float64
	for _, val := range vector {
		sum += val * val
	}
	
	if sum == 0 {
		return vector
	}
	
	norm := math.Sqrt(sum)
	for i := range vector {
		vector[i] /= norm
	}
	
	return vector
}

func (bc *BookmarkClusterer) initializeCentroids(vectors map[int][]float64, k int) map[int][]float64 {
	centroids := make(map[int][]float64)
	
	vectorDim := 50
	for i := 0; i < k; i++ {
		centroid := make([]float64, vectorDim)
		// Simple random initialization
		for j := 0; j < vectorDim; j++ {
			centroid[j] = math.Sin(float64(i*j)) * 0.5
		}
		centroids[i] = centroid
	}
	
	return centroids
}

func (bc *BookmarkClusterer) findClosestCentroid(vector []float64, centroids map[int][]float64) int {
	maxSimilarity := -1.0
	closestCentroid := 0
	
	for centroidID, centroid := range centroids {
		similarity := bc.cosineSimilarity(vector, centroid)
		if similarity > maxSimilarity {
			maxSimilarity = similarity
			closestCentroid = centroidID
		}
	}
	
	return closestCentroid
}

func (bc *BookmarkClusterer) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}
	
	var dotProduct, normA, normB float64
	
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	
	if normA == 0 || normB == 0 {
		return 0.0
	}
	
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func (bc *BookmarkClusterer) updateCentroids(assignments map[int][]models.Bookmark, vectors map[int][]float64) map[int][]float64 {
	newCentroids := make(map[int][]float64)
	
	for centroidID, bookmarks := range assignments {
		if len(bookmarks) == 0 {
			continue
		}
		
		vectorDim := 50
		centroid := make([]float64, vectorDim)
		count := 0
		
		for _, bookmark := range bookmarks {
			if vector, exists := vectors[bookmark.ID]; exists {
				for i := 0; i < vectorDim; i++ {
					centroid[i] += vector[i]
				}
				count++
			}
		}
		
		if count > 0 {
			for i := range centroid {
				centroid[i] /= float64(count)
			}
		}
		
		newCentroids[centroidID] = centroid
	}
	
	return newCentroids
}

func (bc *BookmarkClusterer) centroidsConverged(old, new map[int][]float64) bool {
	threshold := 0.01
	
	for centroidID, oldCentroid := range old {
		if newCentroid, exists := new[centroidID]; exists {
			similarity := bc.cosineSimilarity(oldCentroid, newCentroid)
			if similarity < (1.0 - threshold) {
				return false
			}
		}
	}
	
	return true
}

// Cluster naming and quality calculation methods

func (bc *BookmarkClusterer) generateSemanticClusterName(bookmarks []models.Bookmark) string {
	commonTags := bc.extractCommonTags(bookmarks)
	
	if len(commonTags) > 0 {
		return strings.Title(commonTags[0]) + " Resources"
	}
	
	// Fallback to domain analysis
	domainCount := make(map[string]int)
	for _, bookmark := range bookmarks {
		domain := bc.extractDomain(bookmark.URL)
		domainCount[domain]++
	}
	
	var topDomain string
	maxCount := 0
	for domain, count := range domainCount {
		if count > maxCount {
			maxCount = count
			topDomain = domain
		}
	}
	
	if topDomain != "" && maxCount > len(bookmarks)/2 {
		return strings.Title(topDomain) + " Collection"
	}
	
	return "Mixed Content Collection"
}

func (bc *BookmarkClusterer) generateDomainClusterName(domain string, bookmarks []models.Bookmark) string {
	return strings.Title(domain) + " Bookmarks"
}

func (bc *BookmarkClusterer) calculateSemanticClusterConfidence(cluster BookmarkCluster, vectors map[int][]float64) float64 {
	if len(cluster.Bookmarks) < 2 {
		return 0.5
	}
	
	// Calculate average intra-cluster similarity
	totalSimilarity := 0.0
	comparisons := 0
	
	for i, bookmarkA := range cluster.Bookmarks {
		for j, bookmarkB := range cluster.Bookmarks {
			if i < j {
				if vectorA, existsA := vectors[bookmarkA.ID]; existsA {
					if vectorB, existsB := vectors[bookmarkB.ID]; existsB {
						similarity := bc.cosineSimilarity(vectorA, vectorB)
						totalSimilarity += similarity
						comparisons++
					}
				}
			}
		}
	}
	
	if comparisons == 0 {
		return 0.5
	}
	
	avgSimilarity := totalSimilarity / float64(comparisons)
	return math.Min(1.0, avgSimilarity)
}

func (bc *BookmarkClusterer) calculateDomainClusterConfidence(bookmarks []models.Bookmark) float64 {
	if len(bookmarks) < 2 {
		return 0.5
	}
	
	domainCount := make(map[string]int)
	for _, bookmark := range bookmarks {
		domain := bc.extractDomain(bookmark.URL)
		domainCount[domain]++
	}
	
	maxCount := 0
	for _, count := range domainCount {
		if count > maxCount {
			maxCount = count
		}
	}
	
	purity := float64(maxCount) / float64(len(bookmarks))
	return math.Min(1.0, purity)
}

func (bc *BookmarkClusterer) calculateClusterQuality(cluster BookmarkCluster, allClusters []BookmarkCluster) ClusterQuality {
	// Simplified quality metrics
	cohesion := bc.calculateCohesion(cluster)
	separation := bc.calculateSeparation(cluster, allClusters)
	silhouette := (cohesion + separation) / 2.0
	purity := bc.calculatePurity(cluster)
	
	return ClusterQuality{
		Cohesion:   cohesion,
		Separation: separation,
		Silhouette: silhouette,
		Purity:     purity,
	}
}

func (bc *BookmarkClusterer) calculateCohesion(cluster BookmarkCluster) float64 {
	// Measure how similar bookmarks within cluster are
	if len(cluster.Bookmarks) < 2 {
		return 1.0
	}
	
	// Use domain similarity as a proxy for cohesion
	domainCount := make(map[string]int)
	for _, bookmark := range cluster.Bookmarks {
		domain := bc.extractDomain(bookmark.URL)
		domainCount[domain]++
	}
	
	maxCount := 0
	for _, count := range domainCount {
		if count > maxCount {
			maxCount = count
		}
	}
	
	return float64(maxCount) / float64(len(cluster.Bookmarks))
}

func (bc *BookmarkClusterer) calculateSeparation(cluster BookmarkCluster, allClusters []BookmarkCluster) float64 {
	// Measure how different this cluster is from others
	if len(allClusters) <= 1 {
		return 1.0
	}
	
	// Use tag overlap as a proxy for separation
	clusterTags := make(map[string]bool)
	for _, bookmark := range cluster.Bookmarks {
		for _, tag := range bookmark.Tags {
			clusterTags[tag.Name] = true
		}
	}
	
	totalOverlap := 0.0
	comparisons := 0
	
	for _, otherCluster := range allClusters {
		if otherCluster.ID == cluster.ID {
			continue
		}
		
		otherTags := make(map[string]bool)
		for _, bookmark := range otherCluster.Bookmarks {
			for _, tag := range bookmark.Tags {
				otherTags[tag.Name] = true
			}
		}
		
		// Calculate tag overlap
		common := 0
		for tag := range clusterTags {
			if otherTags[tag] {
				common++
			}
		}
		
		totalTags := len(clusterTags) + len(otherTags) - common
		if totalTags > 0 {
			overlap := float64(common) / float64(totalTags)
			totalOverlap += overlap
			comparisons++
		}
	}
	
	if comparisons == 0 {
		return 1.0
	}
	
	avgOverlap := totalOverlap / float64(comparisons)
	return 1.0 - avgOverlap // Higher separation means lower overlap
}

func (bc *BookmarkClusterer) calculatePurity(cluster BookmarkCluster) float64 {
	// Domain purity
	domainCount := make(map[string]int)
	for _, bookmark := range cluster.Bookmarks {
		domain := bc.extractDomain(bookmark.URL)
		domainCount[domain]++
	}
	
	maxCount := 0
	for _, count := range domainCount {
		if count > maxCount {
			maxCount = count
		}
	}
	
	return float64(maxCount) / float64(len(cluster.Bookmarks))
}

func (bc *BookmarkClusterer) calculateOverallQuality(clusters []BookmarkCluster) float64 {
	if len(clusters) == 0 {
		return 0.0
	}
	
	totalQuality := 0.0
	for _, cluster := range clusters {
		totalQuality += cluster.Quality.Silhouette
	}
	
	return totalQuality / float64(len(clusters))
}

func (bc *BookmarkClusterer) refineAndNameClusters(clusters []BookmarkCluster) []BookmarkCluster {
	// Remove low-quality clusters
	var refined []BookmarkCluster
	for _, cluster := range clusters {
		if cluster.Quality.Silhouette > 0.3 && len(cluster.Bookmarks) >= bc.minClusterSize {
			refined = append(refined, cluster)
		}
	}
	
	// Update names based on refined analysis
	for i := range refined {
		if refined[i].Name == "" || strings.Contains(refined[i].Name, "Mixed Content") {
			refined[i].Name = bc.generateImprovedClusterName(refined[i])
		}
		refined[i].Description = bc.generateClusterDescription(refined[i])
	}
	
	return refined
}

func (bc *BookmarkClusterer) generateImprovedClusterName(cluster BookmarkCluster) string {
	// Use themes for better naming
	if len(cluster.Themes) > 0 {
		return cluster.Themes[0].Name
	}
	
	// Use common tags
	if len(cluster.CommonTags) > 0 {
		return strings.Title(cluster.CommonTags[0]) + " Collection"
	}
	
	// Fallback to domain
	domainCount := make(map[string]int)
	for _, bookmark := range cluster.Bookmarks {
		domain := bc.extractDomain(bookmark.URL)
		domainCount[domain]++
	}
	
	var topDomain string
	maxCount := 0
	for domain, count := range domainCount {
		if count > maxCount {
			maxCount = count
			topDomain = domain
		}
	}
	
	if topDomain != "" {
		return strings.Title(topDomain) + " Resources"
	}
	
	return "Bookmark Collection"
}

func (bc *BookmarkClusterer) generateClusterDescription(cluster BookmarkCluster) string {
	desc := ""
	
	if len(cluster.CommonTags) > 0 {
		desc = "Common topics: " + strings.Join(cluster.CommonTags[:int(math.Min(float64(len(cluster.CommonTags)), 3))], ", ")
	}
	
	domainCount := make(map[string]int)
	for _, bookmark := range cluster.Bookmarks {
		domain := bc.extractDomain(bookmark.URL)
		domainCount[domain]++
	}
	
	if len(domainCount) == 1 {
		for domain := range domainCount {
			desc += ". All bookmarks from " + domain
		}
	} else if len(domainCount) <= 3 {
		var domains []string
		for domain := range domainCount {
			domains = append(domains, domain)
		}
		desc += ". Sources: " + strings.Join(domains, ", ")
	}
	
	return desc
}

func (bc *BookmarkClusterer) generateSummary(allBookmarks []models.Bookmark, clusters []BookmarkCluster) ClusteringSummary {
	summary := ClusteringSummary{
		TopDomains:      make(map[string]int),
		TopTags:         make(map[string]int),
		TopCategories:   make(map[string]int),
		TimeDistribution: make(map[string]int),
	}
	
	// Analyze all bookmarks
	for _, bookmark := range allBookmarks {
		// Domain analysis
		domain := bc.extractDomain(bookmark.URL)
		summary.TopDomains[domain]++
		
		// Tag analysis
		for _, tag := range bookmark.Tags {
			summary.TopTags[tag.Name]++
		}
		
		// Time analysis
		timeKey := bookmark.CreatedAt.Format("2006-01")
		summary.TimeDistribution[timeKey]++
	}
	
	// Generate insights
	insights := []string{}
	
	if len(clusters) > 0 {
		insights = append(insights, fmt.Sprintf("Found %d distinct bookmark clusters", len(clusters)))
	}
	
	maxDomainCount := 0
	topDomain := ""
	for domain, count := range summary.TopDomains {
		if count > maxDomainCount {
			maxDomainCount = count
			topDomain = domain
		}
	}
	
	if topDomain != "" {
		percentage := (float64(maxDomainCount) / float64(len(allBookmarks))) * 100
		insights = append(insights, fmt.Sprintf("%.1f%% of bookmarks are from %s", percentage, topDomain))
	}
	
	summary.Insights = insights
	
	return summary
}

// Helper function to generate cluster IDs
func generateClusterID(id int) string {
	return fmt.Sprintf("cluster_%d_%d", id, time.Now().Unix())
}