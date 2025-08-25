package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"torimemo/internal/db"
)

// AnalyticsHandler handles analytics and metrics endpoints
type AnalyticsHandler struct {
	bookmarkRepo *db.BookmarkRepository
	tagRepo      *db.TagRepository
	db           *db.DB
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(bookmarkRepo *db.BookmarkRepository, tagRepo *db.TagRepository, database *db.DB) *AnalyticsHandler {
	return &AnalyticsHandler{
		bookmarkRepo: bookmarkRepo,
		tagRepo:      tagRepo,
		db:           database,
	}
}

// AnalyticsResponse represents the analytics dashboard data
type AnalyticsResponse struct {
	Overview      OverviewMetrics      `json:"overview"`
	Growth        GrowthMetrics        `json:"growth"`
	TopTags       []TagMetric          `json:"top_tags"`
	TopDomains    []DomainMetric       `json:"top_domains"`
	Activity      []ActivityMetric     `json:"recent_activity"`
	SearchMetrics SearchMetrics        `json:"search_metrics"`
}

type OverviewMetrics struct {
	TotalBookmarks   int     `json:"total_bookmarks"`
	TotalTags        int     `json:"total_tags"`
	FavoriteCount    int     `json:"favorite_count"`
	FavoritePercent  float64 `json:"favorite_percent"`
	AvgBookmarksPerDay float64 `json:"avg_bookmarks_per_day"`
	DatabaseSizeMB   float64 `json:"database_size_mb"`
}

type GrowthMetrics struct {
	BookmarksThisWeek int `json:"bookmarks_this_week"`
	BookmarksLastWeek int `json:"bookmarks_last_week"`
	GrowthPercent     float64 `json:"growth_percent"`
}

type TagMetric struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Color string `json:"color"`
}

type DomainMetric struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

type ActivityMetric struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type SearchMetrics struct {
	MostSearchedTerms []SearchTermMetric `json:"most_searched_terms"`
	AvgSearchTime     float64           `json:"avg_search_time_ms"`
}

type SearchTermMetric struct {
	Term  string `json:"term"`
	Count int    `json:"count"`
}

// ServeHTTP implements the http.Handler interface
func (h *AnalyticsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	analytics, err := h.getAnalytics()
	if err != nil {
		h.writeError(w, "Failed to get analytics", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(analytics)
}

func (h *AnalyticsHandler) getAnalytics() (*AnalyticsResponse, error) {
	// Get overview metrics
	overview, err := h.getOverviewMetrics()
	if err != nil {
		return nil, err
	}

	// Get growth metrics
	growth, err := h.getGrowthMetrics()
	if err != nil {
		return nil, err
	}

	// Get top tags
	topTags, err := h.getTopTags()
	if err != nil {
		return nil, err
	}

	// Get top domains
	topDomains, err := h.getTopDomains()
	if err != nil {
		return nil, err
	}

	// Get recent activity
	activity, err := h.getRecentActivity()
	if err != nil {
		return nil, err
	}

	// Mock search metrics for now (would need search logging)
	searchMetrics := SearchMetrics{
		MostSearchedTerms: []SearchTermMetric{
			{Term: "programming", Count: 15},
			{Term: "javascript", Count: 12},
			{Term: "golang", Count: 8},
		},
		AvgSearchTime: 8.5,
	}

	return &AnalyticsResponse{
		Overview:      *overview,
		Growth:        *growth,
		TopTags:       topTags,
		TopDomains:    topDomains,
		Activity:      activity,
		SearchMetrics: searchMetrics,
	}, nil
}

func (h *AnalyticsHandler) getOverviewMetrics() (*OverviewMetrics, error) {
	// Get total bookmarks (use userID 0 for admin/global analytics)
	bookmarks, err := h.bookmarkRepo.List(1, 10000, "", "", false, 0)
	if err != nil {
		return nil, err
	}

	// Get total tags
	tags, err := h.tagRepo.List("")
	if err != nil {
		return nil, err
	}

	// Count favorites
	favoriteCount := 0
	for _, bookmark := range bookmarks.Bookmarks {
		if bookmark.IsFavorite {
			favoriteCount++
		}
	}

	// Calculate favorite percentage
	var favoritePercent float64
	if len(bookmarks.Bookmarks) > 0 {
		favoritePercent = (float64(favoriteCount) / float64(len(bookmarks.Bookmarks))) * 100
	}

	// Calculate average bookmarks per day (based on first and last bookmark dates)
	var avgPerDay float64
	if len(bookmarks.Bookmarks) > 1 {
		first := bookmarks.Bookmarks[len(bookmarks.Bookmarks)-1].CreatedAt
		last := bookmarks.Bookmarks[0].CreatedAt
		days := last.Sub(first).Hours() / 24
		if days > 0 {
			avgPerDay = float64(len(bookmarks.Bookmarks)) / days
		}
	}

	// Get database size
	stats, err := h.db.GetDBStats()
	if err != nil {
		return nil, err
	}
	
	dbSizeMB := float64(h.getIntValue(stats, "file_size_bytes")) / 1024 / 1024

	return &OverviewMetrics{
		TotalBookmarks:     len(bookmarks.Bookmarks),
		TotalTags:          len(tags),
		FavoriteCount:      favoriteCount,
		FavoritePercent:    favoritePercent,
		AvgBookmarksPerDay: avgPerDay,
		DatabaseSizeMB:     dbSizeMB,
	}, nil
}

func (h *AnalyticsHandler) getGrowthMetrics() (*GrowthMetrics, error) {
	now := time.Now()
	weekAgo := now.AddDate(0, 0, -7)
	twoWeeksAgo := now.AddDate(0, 0, -14)

	// Count bookmarks from this week
	thisWeekCount, err := h.countBookmarksSince(weekAgo)
	if err != nil {
		return nil, err
	}

	// Count bookmarks from last week
	lastWeekCount, err := h.countBookmarksBetween(twoWeeksAgo, weekAgo)
	if err != nil {
		return nil, err
	}

	// Calculate growth percentage
	var growthPercent float64
	if lastWeekCount > 0 {
		growthPercent = ((float64(thisWeekCount) - float64(lastWeekCount)) / float64(lastWeekCount)) * 100
	}

	return &GrowthMetrics{
		BookmarksThisWeek: thisWeekCount,
		BookmarksLastWeek: lastWeekCount,
		GrowthPercent:     growthPercent,
	}, nil
}

func (h *AnalyticsHandler) getTopTags() ([]TagMetric, error) {
	query := `
		SELECT t.name, t.color, COUNT(bt.bookmark_id) as usage_count
		FROM tags t
		LEFT JOIN bookmark_tags bt ON t.id = bt.tag_id
		GROUP BY t.id, t.name, t.color
		ORDER BY usage_count DESC
		LIMIT 10
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []TagMetric
	for rows.Next() {
		var metric TagMetric
		err := rows.Scan(&metric.Name, &metric.Color, &metric.Count)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (h *AnalyticsHandler) getTopDomains() ([]DomainMetric, error) {
	query := `
		SELECT 
			CASE 
				WHEN url LIKE 'https://%' THEN SUBSTR(url, 9)
				WHEN url LIKE 'http://%' THEN SUBSTR(url, 8)
				ELSE url
			END as domain_part,
			COUNT(*) as count
		FROM bookmarks
		GROUP BY domain_part
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []DomainMetric
	for rows.Next() {
		var domainPart string
		var count int
		err := rows.Scan(&domainPart, &count)
		if err != nil {
			return nil, err
		}
		
		// Extract just the domain name
		if idx := 0; idx < len(domainPart) {
			if slashIdx := 0; slashIdx < len(domainPart) && domainPart[slashIdx] == '/' {
				domainPart = domainPart[:slashIdx]
			}
		}
		
		metrics = append(metrics, DomainMetric{
			Domain: domainPart,
			Count:  count,
		})
	}

	return metrics, nil
}

func (h *AnalyticsHandler) getRecentActivity() ([]ActivityMetric, error) {
	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM bookmarks
		WHERE created_at >= DATE('now', '-30 days')
		GROUP BY DATE(created_at)
		ORDER BY date DESC
		LIMIT 30
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []ActivityMetric
	for rows.Next() {
		var metric ActivityMetric
		err := rows.Scan(&metric.Date, &metric.Count)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

func (h *AnalyticsHandler) countBookmarksSince(since time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM bookmarks WHERE created_at >= ?"
	var count int
	err := h.db.QueryRow(query, since).Scan(&count)
	return count, err
}

func (h *AnalyticsHandler) countBookmarksBetween(start, end time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM bookmarks WHERE created_at >= ? AND created_at < ?"
	var count int
	err := h.db.QueryRow(query, start, end).Scan(&count)
	return count, err
}

func (h *AnalyticsHandler) getIntValue(stats map[string]interface{}, key string) int {
	if val, ok := stats[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if int64Val, ok := val.(int64); ok {
			return int(int64Val)
		}
	}
	return 0
}

func (h *AnalyticsHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  message,
		"status": statusCode,
	})
}