package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"torimemo/internal/models"
)

// BookmarkRepository handles bookmark database operations
type BookmarkRepository struct {
	db *DB
}

// NewBookmarkRepository creates a new bookmark repository
func NewBookmarkRepository(db *DB) *BookmarkRepository {
	return &BookmarkRepository{db: db}
}

// Create creates a new bookmark
func (r *BookmarkRepository) Create(req *models.CreateBookmarkRequest) (*models.Bookmark, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert bookmark
	query := `
		INSERT INTO bookmarks (title, url, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := tx.Exec(query, req.Title, req.URL, req.Description, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to create bookmark: %w", err)
	}

	bookmarkID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark ID: %w", err)
	}

	// Add tags if provided
	if len(req.Tags) > 0 {
		if err := r.addTagsToBookmark(tx, int(bookmarkID), req.Tags); err != nil {
			return nil, fmt.Errorf("failed to add tags: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Fetch and return the created bookmark
	return r.GetByID(int(bookmarkID))
}

// GetByID retrieves a bookmark by ID
func (r *BookmarkRepository) GetByID(id int) (*models.Bookmark, error) {
	bookmark := &models.Bookmark{}
	query := `
		SELECT id, title, url, description, favicon_url, created_at, updated_at, is_favorite
		FROM bookmarks WHERE id = ?
	`
	err := r.db.QueryRow(query, id).Scan(
		&bookmark.ID, &bookmark.Title, &bookmark.URL, &bookmark.Description,
		&bookmark.FaviconURL, &bookmark.CreatedAt, &bookmark.UpdatedAt, &bookmark.IsFavorite,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bookmark not found")
		}
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	// Load tags
	tags, err := r.getBookmarkTags(bookmark.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	bookmark.Tags = tags

	return bookmark, nil
}

// GetByURL retrieves a bookmark by URL
func (r *BookmarkRepository) GetByURL(url string) (*models.Bookmark, error) {
	bookmark := &models.Bookmark{}
	query := `
		SELECT id, title, url, description, favicon_url, created_at, updated_at, is_favorite
		FROM bookmarks WHERE url = ?
	`
	err := r.db.QueryRow(query, url).Scan(
		&bookmark.ID, &bookmark.Title, &bookmark.URL, &bookmark.Description,
		&bookmark.FaviconURL, &bookmark.CreatedAt, &bookmark.UpdatedAt, &bookmark.IsFavorite,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, return nil without error
		}
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	// Load tags
	tags, err := r.getBookmarkTags(bookmark.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load tags: %w", err)
	}
	bookmark.Tags = tags

	return bookmark, nil
}

// GetDB returns the underlying database connection
func (r *BookmarkRepository) GetDB() *DB {
	return r.db
}

// GetBookmarkTags is a public method to get tags for a bookmark
func (r *BookmarkRepository) GetBookmarkTags(bookmarkID int) ([]models.Tag, error) {
	return r.getBookmarkTags(bookmarkID)
}


// List retrieves bookmarks with pagination and filtering
func (r *BookmarkRepository) List(page, limit int, searchQuery, tagFilter string, favoritesOnly bool) (*models.BookmarkListResponse, error) {
	offset := (page - 1) * limit

	// Build query conditions
	var conditions []string
	var args []interface{}
	
	if searchQuery != "" {
		conditions = append(conditions, "id IN (SELECT rowid FROM bookmarks_fts WHERE bookmarks_fts MATCH ?)")
		args = append(args, searchQuery)
	}

	if tagFilter != "" {
		conditions = append(conditions, `id IN (
			SELECT bt.bookmark_id FROM bookmark_tags bt
			JOIN tags t ON bt.tag_id = t.id
			WHERE t.name = ?
		)`)
		args = append(args, tagFilter)
	}

	if favoritesOnly {
		conditions = append(conditions, "is_favorite = 1")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM bookmarks %s", whereClause)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get bookmarks
	query := fmt.Sprintf(`
		SELECT id, title, url, description, favicon_url, created_at, updated_at, is_favorite
		FROM bookmarks %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, whereClause)
	
	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query bookmarks: %w", err)
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
			return nil, fmt.Errorf("failed to scan bookmark: %w", err)
		}

		// Load tags for each bookmark
		tags, err := r.getBookmarkTags(bookmark.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load tags for bookmark %d: %w", bookmark.ID, err)
		}
		bookmark.Tags = tags

		bookmarks = append(bookmarks, bookmark)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate bookmarks: %w", err)
	}

	// Get statistics
	var tagCount, favoriteCount int
	r.db.QueryRow("SELECT COUNT(DISTINCT id) FROM tags").Scan(&tagCount)
	r.db.QueryRow("SELECT COUNT(*) FROM bookmarks WHERE is_favorite = 1").Scan(&favoriteCount)

	totalPages := (total + limit - 1) / limit
	hasMore := page < totalPages

	return &models.BookmarkListResponse{
		Bookmarks:     bookmarks,
		Total:         total,
		Page:          page,
		Limit:         limit,
		HasMore:       hasMore,
		TotalPages:    totalPages,
		TagCount:      tagCount,
		FavoriteCount: favoriteCount,
	}, nil
}

// Update updates an existing bookmark
func (r *BookmarkRepository) Update(id int, req *models.UpdateBookmarkRequest) (*models.Bookmark, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Build update query dynamically
	var setParts []string
	var args []interface{}

	if req.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *req.Title)
	}
	if req.URL != nil {
		setParts = append(setParts, "url = ?")
		args = append(args, *req.URL)
	}
	if req.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *req.Description)
	}
	if req.IsFavorite != nil {
		setParts = append(setParts, "is_favorite = ?")
		args = append(args, *req.IsFavorite)
	}

	if len(setParts) > 0 {
		setParts = append(setParts, "updated_at = ?")
		args = append(args, time.Now())
		args = append(args, id)

		query := fmt.Sprintf("UPDATE bookmarks SET %s WHERE id = ?", strings.Join(setParts, ", "))
		if _, err := tx.Exec(query, args...); err != nil {
			return nil, fmt.Errorf("failed to update bookmark: %w", err)
		}
	}

	// Update tags if provided
	if req.Tags != nil {
		// Remove existing tags
		if _, err := tx.Exec("DELETE FROM bookmark_tags WHERE bookmark_id = ?", id); err != nil {
			return nil, fmt.Errorf("failed to remove existing tags: %w", err)
		}

		// Add new tags
		if len(req.Tags) > 0 {
			if err := r.addTagsToBookmark(tx, id, req.Tags); err != nil {
				return nil, fmt.Errorf("failed to add new tags: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Fetch and return the updated bookmark
	return r.GetByID(id)
}

// Delete deletes a bookmark by ID
func (r *BookmarkRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM bookmarks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bookmark not found")
	}

	return nil
}

// Search performs full-text search on bookmarks
func (r *BookmarkRepository) Search(query string, limit int) ([]models.SearchResult, error) {
	sqlQuery := `
		SELECT b.id, b.title, b.url, b.description, b.favicon_url, 
		       b.created_at, b.updated_at, b.is_favorite, 
		       bm25(bookmarks_fts) as rank,
		       snippet(bookmarks_fts, 0, '<mark>', '</mark>', '...', 32) as snippet
		FROM bookmarks b
		JOIN bookmarks_fts ON b.id = bookmarks_fts.rowid
		WHERE bookmarks_fts MATCH ?
		ORDER BY rank
		LIMIT ?
	`

	rows, err := r.db.Query(sqlQuery, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search bookmarks: %w", err)
	}
	defer rows.Close()

	var results []models.SearchResult
	for rows.Next() {
		var result models.SearchResult
		err := rows.Scan(
			&result.ID, &result.Title, &result.URL, &result.Description,
			&result.FaviconURL, &result.CreatedAt, &result.UpdatedAt, &result.IsFavorite,
			&result.Rank, &result.Snippet,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		// Load tags
		tags, err := r.getBookmarkTags(result.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load tags for search result %d: %w", result.ID, err)
		}
		result.Tags = tags

		results = append(results, result)
	}

	return results, rows.Err()
}

// addTagsToBookmark adds tags to a bookmark within a transaction
func (r *BookmarkRepository) addTagsToBookmark(tx *sql.Tx, bookmarkID int, tagNames []string) error {
	for _, tagName := range tagNames {
		if tagName = strings.TrimSpace(tagName); tagName == "" {
			continue
		}

		// Get or create tag
		var tagID int
		err := tx.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
		if err == sql.ErrNoRows {
			// Create new tag
			result, err := tx.Exec("INSERT INTO tags (name, created_at) VALUES (?, ?)", tagName, time.Now())
			if err != nil {
				return fmt.Errorf("failed to create tag %s: %w", tagName, err)
			}
			id, err := result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get tag ID: %w", err)
			}
			tagID = int(id)
		} else if err != nil {
			return fmt.Errorf("failed to get tag %s: %w", tagName, err)
		}

		// Link bookmark to tag
		_, err = tx.Exec("INSERT OR IGNORE INTO bookmark_tags (bookmark_id, tag_id, created_at) VALUES (?, ?, ?)",
			bookmarkID, tagID, time.Now())
		if err != nil {
			return fmt.Errorf("failed to link bookmark to tag %s: %w", tagName, err)
		}
	}

	return nil
}

// getBookmarkTags retrieves all tags for a bookmark
func (r *BookmarkRepository) getBookmarkTags(bookmarkID int) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.color, t.created_at
		FROM tags t
		JOIN bookmark_tags bt ON t.id = bt.tag_id
		WHERE bt.bookmark_id = ?
		ORDER BY t.name
	`
	
	rows, err := r.db.Query(query, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}