package models

import "time"

// Tag represents a bookmark tag for categorization
type Tag struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Color     string    `json:"color" db:"color"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Count     int       `json:"count,omitempty"`
}

// CreateTagRequest represents the request to create a new tag
type CreateTagRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=50"`
	Color string `json:"color,omitempty"`
}

// UpdateTagRequest represents the request to update a tag
type UpdateTagRequest struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

// TagWithCount represents a tag with its bookmark count
type TagWithCount struct {
	Tag
	BookmarkCount int `json:"bookmark_count"`
}

// TagCloudItem represents a tag for tag cloud display
type TagCloudItem struct {
	Name  string  `json:"name"`
	Count int     `json:"count"`
	Size  float64 `json:"size"` // Relative size for display (0.0 - 1.0)
	Color string  `json:"color"`
}