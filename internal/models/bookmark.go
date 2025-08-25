package models

import (
	"time"
)

// Bookmark represents a bookmarked URL with metadata
type Bookmark struct {
	ID          int       `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	URL         string    `json:"url" db:"url"`
	Description *string   `json:"description,omitempty" db:"description"`
	FaviconURL  *string   `json:"favicon_url,omitempty" db:"favicon_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	IsFavorite  bool      `json:"is_favorite" db:"is_favorite"`
	Tags        []Tag     `json:"tags,omitempty"`
}

// CreateBookmarkRequest represents the request to create a new bookmark
type CreateBookmarkRequest struct {
	Title       string   `json:"title"`
	URL         string   `json:"url" validate:"required,url"`
	Description *string  `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// UpdateBookmarkRequest represents the request to update a bookmark
type UpdateBookmarkRequest struct {
	Title       *string  `json:"title,omitempty"`
	URL         *string  `json:"url,omitempty"`
	Description *string  `json:"description,omitempty"`
	IsFavorite  *bool    `json:"is_favorite,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// BookmarkWithTags represents a bookmark with its associated tags
type BookmarkWithTags struct {
	Bookmark
	TagNames []string `json:"tag_names,omitempty"`
}

// BookmarkListResponse represents the paginated response for bookmarks
type BookmarkListResponse struct {
	Bookmarks    []Bookmark `json:"bookmarks"`
	Total        int        `json:"total"`
	Page         int        `json:"page"`
	Limit        int        `json:"limit"`
	HasMore      bool       `json:"has_more"`
	TotalPages   int        `json:"total_pages"`
	TagCount     int        `json:"tag_count"`
	FavoriteCount int       `json:"favorite_count"`
}

// SearchResult represents a full-text search result
type SearchResult struct {
	Bookmark
	Rank  float64 `json:"rank"`
	Snippet string `json:"snippet,omitempty"`
}