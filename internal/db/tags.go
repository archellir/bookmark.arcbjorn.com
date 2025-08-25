package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"torimemo/internal/models"
)

// TagRepository handles tag database operations
type TagRepository struct {
	db *DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *DB) *TagRepository {
	return &TagRepository{db: db}
}

// Create creates a new tag
func (r *TagRepository) Create(req *models.CreateTagRequest) (*models.Tag, error) {
	query := `INSERT INTO tags (name, color, created_at) VALUES (?, ?, ?)`
	
	color := req.Color
	if color == "" {
		color = "#007acc" // Default color
	}

	result, err := r.db.Exec(query, req.Name, color, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	tagID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get tag ID: %w", err)
	}

	return r.GetByID(int(tagID))
}

// GetByID retrieves a tag by ID
func (r *TagRepository) GetByID(id int) (*models.Tag, error) {
	tag := &models.Tag{}
	query := `SELECT id, name, color, created_at FROM tags WHERE id = ?`
	
	err := r.db.QueryRow(query, id).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

// GetByName retrieves a tag by name
func (r *TagRepository) GetByName(name string) (*models.Tag, error) {
	tag := &models.Tag{}
	query := `SELECT id, name, color, created_at FROM tags WHERE name = ?`
	
	err := r.db.QueryRow(query, name).Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("tag not found")
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}

	return tag, nil
}

// List retrieves all tags with optional filtering
func (r *TagRepository) List(search string) ([]models.TagWithCount, error) {
	var conditions []string
	var args []interface{}

	baseQuery := `
		SELECT t.id, t.name, t.color, t.created_at, COUNT(bt.bookmark_id) as bookmark_count
		FROM tags t
		LEFT JOIN bookmark_tags bt ON t.id = bt.tag_id
	`

	if search != "" {
		conditions = append(conditions, "t.name LIKE ?")
		args = append(args, "%"+search+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf("%s %s GROUP BY t.id, t.name, t.color, t.created_at ORDER BY bookmark_count DESC, t.name ASC", baseQuery, whereClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []models.TagWithCount
	for rows.Next() {
		var tag models.TagWithCount
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt, &tag.BookmarkCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// GetTagCloud returns tags formatted for tag cloud display
func (r *TagRepository) GetTagCloud(limit int) ([]models.TagCloudItem, error) {
	query := `
		SELECT t.name, t.color, COUNT(bt.bookmark_id) as count
		FROM tags t
		LEFT JOIN bookmark_tags bt ON t.id = bt.tag_id
		GROUP BY t.id, t.name, t.color
		HAVING count > 0
		ORDER BY count DESC, t.name ASC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query tag cloud: %w", err)
	}
	defer rows.Close()

	var items []models.TagCloudItem
	var maxCount int

	for rows.Next() {
		var item models.TagCloudItem
		err := rows.Scan(&item.Name, &item.Color, &item.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag cloud item: %w", err)
		}

		if item.Count > maxCount {
			maxCount = item.Count
		}

		items = append(items, item)
	}

	// Calculate relative sizes (0.3 to 1.0)
	for i := range items {
		if maxCount > 0 {
			items[i].Size = 0.3 + (0.7 * float64(items[i].Count) / float64(maxCount))
		} else {
			items[i].Size = 0.5
		}
	}

	return items, rows.Err()
}

// Update updates an existing tag
func (r *TagRepository) Update(id int, req *models.UpdateTagRequest) (*models.Tag, error) {
	var setParts []string
	var args []interface{}

	if req.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Color != nil {
		setParts = append(setParts, "color = ?")
		args = append(args, *req.Color)
	}

	if len(setParts) == 0 {
		return r.GetByID(id) // Nothing to update
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE tags SET %s WHERE id = ?", strings.Join(setParts, ", "))

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("tag not found")
	}

	return r.GetByID(id)
}

// Delete deletes a tag by ID
func (r *TagRepository) Delete(id int) error {
	result, err := r.db.Exec("DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tag not found")
	}

	return nil
}

// GetPopularTags returns the most used tags
func (r *TagRepository) GetPopularTags(limit int) ([]models.TagWithCount, error) {
	query := `
		SELECT t.id, t.name, t.color, t.created_at, COUNT(bt.bookmark_id) as bookmark_count
		FROM tags t
		JOIN bookmark_tags bt ON t.id = bt.tag_id
		GROUP BY t.id, t.name, t.color, t.created_at
		ORDER BY bookmark_count DESC, t.name ASC
		LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query popular tags: %w", err)
	}
	defer rows.Close()

	var tags []models.TagWithCount
	for rows.Next() {
		var tag models.TagWithCount
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt, &tag.BookmarkCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan popular tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// GetUnusedTags returns tags that are not used by any bookmarks
func (r *TagRepository) GetUnusedTags() ([]models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.color, t.created_at
		FROM tags t
		LEFT JOIN bookmark_tags bt ON t.id = bt.tag_id
		WHERE bt.tag_id IS NULL
		ORDER BY t.name ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query unused tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Color, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan unused tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// CleanupUnusedTags removes tags that are not associated with any bookmarks
func (r *TagRepository) CleanupUnusedTags() (int, error) {
	result, err := r.db.Exec(`
		DELETE FROM tags 
		WHERE id NOT IN (
			SELECT DISTINCT tag_id FROM bookmark_tags
		)
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup unused tags: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}