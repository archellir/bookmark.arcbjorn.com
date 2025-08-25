package models

import "time"

// Folder represents a bookmark folder/collection
type Folder struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	Color       string    `json:"color" db:"color"`
	Icon        string    `json:"icon" db:"icon"`
	ParentID    *int      `json:"parent_id" db:"parent_id"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	
	// Computed fields
	Children     []Folder `json:"children,omitempty"`
	BookmarkCount int     `json:"bookmark_count"`
	Path         string   `json:"path"` // Full path like "Work/Projects/Web"
}

// FolderTree represents the hierarchical folder structure
type FolderTree struct {
	Folder   Folder   `json:"folder"`
	Children []FolderTree `json:"children"`
}

// CreateFolderRequest represents the request to create a new folder
type CreateFolderRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	ParentID    *int    `json:"parent_id,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

// UpdateFolderRequest represents the request to update a folder
type UpdateFolderRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	ParentID    *int    `json:"parent_id,omitempty"`
	SortOrder   *int    `json:"sort_order,omitempty"`
}

// MoveFolderRequest represents the request to move folders to a different parent
type MoveFolderRequest struct {
	ParentID  *int `json:"parent_id,omitempty"`
	SortOrder *int `json:"sort_order,omitempty"`
}

// FolderStats represents statistics about a folder
type FolderStats struct {
	ID              int `json:"id"`
	BookmarkCount   int `json:"bookmark_count"`
	SubfolderCount  int `json:"subfolder_count"`
	TotalBookmarks  int `json:"total_bookmarks"` // Including subfolders
}