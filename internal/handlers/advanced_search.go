package handlers

import (
	"encoding/json/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

// AdvancedSearchHandler handles advanced search operations
type AdvancedSearchHandler struct {
	bookmarkRepo *db.BookmarkRepository
}

// NewAdvancedSearchHandler creates a new advanced search handler
func NewAdvancedSearchHandler(bookmarkRepo *db.BookmarkRepository) *AdvancedSearchHandler {
	return &AdvancedSearchHandler{bookmarkRepo: bookmarkRepo}
}

// AdvancedSearchRequest represents advanced search parameters
type AdvancedSearchRequest struct {
	Query        string    `json:"query"`
	Tags         []string  `json:"tags"`
	ExcludeTags  []string  `json:"exclude_tags"`
	Domain       string    `json:"domain"`
	FavoritesOnly *bool    `json:"favorites_only"`
	DateFrom     *time.Time `json:"date_from"`
	DateTo       *time.Time `json:"date_to"`
	SortBy       string    `json:"sort_by"` // created_at, updated_at, title, relevance
	SortOrder    string    `json:"sort_order"` // asc, desc
	Page         int       `json:"page"`
	Limit        int       `json:"limit"`
}

// AdvancedSearchResponse represents advanced search results
type AdvancedSearchResponse struct {
	Bookmarks     []models.Bookmark `json:"bookmarks"`
	Total         int               `json:"total"`
	Page          int               `json:"page"`
	Limit         int               `json:"limit"`
	HasMore       bool              `json:"has_more"`
	SearchSummary SearchSummary     `json:"search_summary"`
}

type SearchSummary struct {
	QueryTerms    []string          `json:"query_terms"`
	TagsFiltered  []string          `json:"tags_filtered"`
	DomainFilter  string            `json:"domain_filter,omitempty"`
	DateRange     *DateRange        `json:"date_range,omitempty"`
	ResultsByType map[string]int    `json:"results_by_type"`
}

type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// ServeHTTP implements the http.Handler interface
func (h *AdvancedSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		h.writeError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req AdvancedSearchRequest
	if err := json.UnmarshalRead(r.Body, &req); err != nil {
		// Try to parse from query parameters as fallback
		req = h.parseQueryParams(r)
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.SortBy == "" {
		if req.Query != "" {
			req.SortBy = "relevance"
		} else {
			req.SortBy = "created_at"
		}
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Perform search
	results, err := h.performAdvancedSearch(&req)
	if err != nil {
		h.writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.MarshalWrite(w, results)
}

func (h *AdvancedSearchHandler) parseQueryParams(r *http.Request) AdvancedSearchRequest {
	req := AdvancedSearchRequest{}
	
	req.Query = r.URL.Query().Get("q")
	req.Domain = r.URL.Query().Get("domain")
	
	if tags := r.URL.Query().Get("tags"); tags != "" {
		req.Tags = strings.Split(tags, ",")
	}
	
	if excludeTags := r.URL.Query().Get("exclude_tags"); excludeTags != "" {
		req.ExcludeTags = strings.Split(excludeTags, ",")
	}
	
	if favStr := r.URL.Query().Get("favorites"); favStr != "" {
		fav := favStr == "true"
		req.FavoritesOnly = &fav
	}
	
	if dateFrom := r.URL.Query().Get("date_from"); dateFrom != "" {
		if parsed, err := time.Parse("2006-01-02", dateFrom); err == nil {
			req.DateFrom = &parsed
		}
	}
	
	if dateTo := r.URL.Query().Get("date_to"); dateTo != "" {
		if parsed, err := time.Parse("2006-01-02", dateTo); err == nil {
			req.DateTo = &parsed
		}
	}
	
	req.SortBy = r.URL.Query().Get("sort_by")
	req.SortOrder = r.URL.Query().Get("sort_order")
	
	if page, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && page > 0 {
		req.Page = page
	}
	
	if limit, err := strconv.Atoi(r.URL.Query().Get("limit")); err == nil && limit > 0 {
		req.Limit = limit
	}
	
	return req
}

func (h *AdvancedSearchHandler) performAdvancedSearch(req *AdvancedSearchRequest) (*AdvancedSearchResponse, error) {
	// Build base query
	var conditions []string
	var args []interface{}
	var joins []string

	// Text search
	if req.Query != "" {
		if req.SortBy == "relevance" {
			// Use FTS for relevance sorting
			conditions = append(conditions, "b.id IN (SELECT rowid FROM bookmarks_fts WHERE bookmarks_fts MATCH ?)")
			args = append(args, req.Query)
		} else {
			// Use LIKE for other sorting
			conditions = append(conditions, "(b.title LIKE ? OR b.description LIKE ? OR b.url LIKE ?)")
			likeQuery := "%" + req.Query + "%"
			args = append(args, likeQuery, likeQuery, likeQuery)
		}
	}

	// Tag filters
	if len(req.Tags) > 0 {
		tagPlaceholders := strings.Repeat("?,", len(req.Tags))
		tagPlaceholders = tagPlaceholders[:len(tagPlaceholders)-1] // Remove trailing comma
		
		joins = append(joins, "JOIN bookmark_tags bt ON b.id = bt.bookmark_id")
		joins = append(joins, "JOIN tags t ON bt.tag_id = t.id")
		conditions = append(conditions, "t.name IN ("+tagPlaceholders+")")
		
		for _, tag := range req.Tags {
			args = append(args, tag)
		}
	}

	// Exclude tags
	if len(req.ExcludeTags) > 0 {
		excludePlaceholders := strings.Repeat("?,", len(req.ExcludeTags))
		excludePlaceholders = excludePlaceholders[:len(excludePlaceholders)-1]
		
		conditions = append(conditions, "b.id NOT IN (SELECT DISTINCT bt2.bookmark_id FROM bookmark_tags bt2 JOIN tags t2 ON bt2.tag_id = t2.id WHERE t2.name IN ("+excludePlaceholders+"))")
		
		for _, tag := range req.ExcludeTags {
			args = append(args, tag)
		}
	}

	// Domain filter
	if req.Domain != "" {
		conditions = append(conditions, "b.url LIKE ?")
		args = append(args, "%"+req.Domain+"%")
	}

	// Favorites filter
	if req.FavoritesOnly != nil {
		conditions = append(conditions, "b.is_favorite = ?")
		args = append(args, *req.FavoritesOnly)
	}

	// Date range filters
	if req.DateFrom != nil {
		conditions = append(conditions, "b.created_at >= ?")
		args = append(args, *req.DateFrom)
	}
	
	if req.DateTo != nil {
		conditions = append(conditions, "b.created_at <= ?")
		args = append(args, *req.DateTo)
	}

	// Build final query
	baseQuery := "FROM bookmarks b"
	if len(joins) > 0 {
		baseQuery += " " + strings.Join(joins, " ")
	}
	
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total results
	countQuery := "SELECT COUNT(DISTINCT b.id) " + baseQuery + whereClause
	var total int
	err := h.bookmarkRepo.GetDB().QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Build sort clause
	sortClause := h.buildSortClause(req.SortBy, req.SortOrder, req.Query != "" && req.SortBy == "relevance")

	// Get results with pagination
	offset := (req.Page - 1) * req.Limit
	selectQuery := `SELECT DISTINCT b.id, b.title, b.url, b.description, b.favicon_url, 
					b.created_at, b.updated_at, b.is_favorite ` + 
					baseQuery + whereClause + sortClause + 
					" LIMIT ? OFFSET ?"
	
	args = append(args, req.Limit, offset)
	
	rows, err := h.bookmarkRepo.GetDB().Query(selectQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []models.Bookmark
	for rows.Next() {
		var bookmark models.Bookmark
		err := rows.Scan(
			&bookmark.ID, &bookmark.Title, &bookmark.URL, &bookmark.Description,
			&bookmark.FaviconURL, &bookmark.CreatedAt, &bookmark.UpdatedAt, &bookmark.IsFavorite,
		)
		if err != nil {
			return nil, err
		}
		
		// Load tags for each bookmark
		tags, err := h.bookmarkRepo.GetBookmarkTags(bookmark.ID)
		if err == nil {
			bookmark.Tags = tags
		}
		
		bookmarks = append(bookmarks, bookmark)
	}

	// Build search summary
	summary := h.buildSearchSummary(req, bookmarks)

	return &AdvancedSearchResponse{
		Bookmarks:     bookmarks,
		Total:         total,
		Page:          req.Page,
		Limit:         req.Limit,
		HasMore:       offset+len(bookmarks) < total,
		SearchSummary: *summary,
	}, nil
}

func (h *AdvancedSearchHandler) buildSortClause(sortBy, sortOrder string, useRelevance bool) string {
	orderDir := "DESC"
	if strings.ToLower(sortOrder) == "asc" {
		orderDir = "ASC"
	}

	switch sortBy {
	case "relevance":
		if useRelevance {
			return " ORDER BY rank " + orderDir
		}
		return " ORDER BY b.created_at " + orderDir
	case "title":
		return " ORDER BY b.title " + orderDir
	case "updated_at":
		return " ORDER BY b.updated_at " + orderDir
	case "created_at":
		fallthrough
	default:
		return " ORDER BY b.created_at " + orderDir
	}
}

func (h *AdvancedSearchHandler) buildSearchSummary(req *AdvancedSearchRequest, bookmarks []models.Bookmark) *SearchSummary {
	summary := &SearchSummary{
		TagsFiltered:  req.Tags,
		DomainFilter:  req.Domain,
		ResultsByType: make(map[string]int),
	}

	// Parse query terms
	if req.Query != "" {
		summary.QueryTerms = strings.Fields(req.Query)
	}

	// Date range
	if req.DateFrom != nil || req.DateTo != nil {
		summary.DateRange = &DateRange{}
		if req.DateFrom != nil {
			summary.DateRange.From = req.DateFrom.Format("2006-01-02")
		}
		if req.DateTo != nil {
			summary.DateRange.To = req.DateTo.Format("2006-01-02")
		}
	}

	// Count results by type
	summary.ResultsByType["total"] = len(bookmarks)
	for _, bookmark := range bookmarks {
		if bookmark.IsFavorite {
			summary.ResultsByType["favorites"]++
		}
		if bookmark.Description != nil && *bookmark.Description != "" {
			summary.ResultsByType["with_description"]++
		}
		if len(bookmark.Tags) > 0 {
			summary.ResultsByType["with_tags"]++
		}
	}

	return summary
}

func (h *AdvancedSearchHandler) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	json.MarshalWrite(w, map[string]interface{}{
		"error":  message,
		"status": statusCode,
	})
}