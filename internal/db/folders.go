package db

import (
	"database/sql"
	"fmt"
	"strings"

	"torimemo/internal/models"
)

// FolderRepository handles folder database operations
type FolderRepository struct {
	db *DB
}

// NewFolderRepository creates a new folder repository
func NewFolderRepository(db *DB) *FolderRepository {
	return &FolderRepository{db: db}
}

// Create creates a new folder
func (r *FolderRepository) Create(req *models.CreateFolderRequest) (*models.Folder, error) {
	// Set defaults
	color := "#666666"
	if req.Color != nil {
		color = *req.Color
	}
	
	icon := "📁"
	if req.Icon != nil {
		icon = *req.Icon
	}
	
	sortOrder := 0
	if req.SortOrder != nil {
		sortOrder = *req.SortOrder
	}

	// Insert folder
	query := `
		INSERT INTO folders (name, description, color, icon, parent_id, sort_order)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, req.Name, req.Description, color, icon, req.ParentID, sortOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get folder ID: %w", err)
	}

	return r.GetByID(int(id))
}

// GetByID retrieves a folder by ID
func (r *FolderRepository) GetByID(id int) (*models.Folder, error) {
	query := `
		SELECT f.id, f.name, f.description, f.color, f.icon, f.parent_id, f.sort_order,
		       f.created_at, f.updated_at,
		       COUNT(bf.bookmark_id) as bookmark_count
		FROM folders f
		LEFT JOIN bookmark_folders bf ON f.id = bf.folder_id
		WHERE f.id = ?
		GROUP BY f.id
	`

	var folder models.Folder
	err := r.db.QueryRow(query, id).Scan(
		&folder.ID, &folder.Name, &folder.Description, &folder.Color, &folder.Icon,
		&folder.ParentID, &folder.SortOrder, &folder.CreatedAt, &folder.UpdatedAt,
		&folder.BookmarkCount,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}

	// Build folder path
	folder.Path, _ = r.buildFolderPath(id)

	return &folder, nil
}

// List retrieves all folders with optional parent filter
func (r *FolderRepository) List(parentID *int) ([]*models.Folder, error) {
	var query string
	var args []interface{}

	if parentID == nil {
		// Get root folders (parent_id IS NULL)
		query = `
			SELECT f.id, f.name, f.description, f.color, f.icon, f.parent_id, f.sort_order,
			       f.created_at, f.updated_at,
			       COUNT(bf.bookmark_id) as bookmark_count
			FROM folders f
			LEFT JOIN bookmark_folders bf ON f.id = bf.folder_id
			WHERE f.parent_id IS NULL
			GROUP BY f.id
			ORDER BY f.sort_order, f.name
		`
	} else {
		// Get subfolders of specific parent
		query = `
			SELECT f.id, f.name, f.description, f.color, f.icon, f.parent_id, f.sort_order,
			       f.created_at, f.updated_at,
			       COUNT(bf.bookmark_id) as bookmark_count
			FROM folders f
			LEFT JOIN bookmark_folders bf ON f.id = bf.folder_id
			WHERE f.parent_id = ?
			GROUP BY f.id
			ORDER BY f.sort_order, f.name
		`
		args = []interface{}{*parentID}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}
	defer rows.Close()

	var folders []*models.Folder
	for rows.Next() {
		var folder models.Folder
		err := rows.Scan(
			&folder.ID, &folder.Name, &folder.Description, &folder.Color, &folder.Icon,
			&folder.ParentID, &folder.SortOrder, &folder.CreatedAt, &folder.UpdatedAt,
			&folder.BookmarkCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}

		// Build folder path
		folder.Path, _ = r.buildFolderPath(folder.ID)

		folders = append(folders, &folder)
	}

	return folders, nil
}

// GetTree retrieves the complete folder tree
func (r *FolderRepository) GetTree() ([]*models.FolderTree, error) {
	// Get all folders
	query := `
		SELECT f.id, f.name, f.description, f.color, f.icon, f.parent_id, f.sort_order,
		       f.created_at, f.updated_at,
		       COUNT(bf.bookmark_id) as bookmark_count
		FROM folders f
		LEFT JOIN bookmark_folders bf ON f.id = bf.folder_id
		GROUP BY f.id
		ORDER BY f.sort_order, f.name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder tree: %w", err)
	}
	defer rows.Close()

	folderMap := make(map[int]*models.Folder)
	var rootFolders []*models.Folder

	for rows.Next() {
		var folder models.Folder
		err := rows.Scan(
			&folder.ID, &folder.Name, &folder.Description, &folder.Color, &folder.Icon,
			&folder.ParentID, &folder.SortOrder, &folder.CreatedAt, &folder.UpdatedAt,
			&folder.BookmarkCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}

		folderMap[folder.ID] = &folder

		if folder.ParentID == nil {
			rootFolders = append(rootFolders, &folder)
		}
	}

	// Build tree structure
	var tree []*models.FolderTree
	for _, folder := range rootFolders {
		treeNode := r.buildTreeNode(folder, folderMap)
		tree = append(tree, treeNode)
	}

	return tree, nil
}

// Update updates a folder
func (r *FolderRepository) Update(id int, req *models.UpdateFolderRequest) (*models.Folder, error) {
	// Build dynamic update query
	var setParts []string
	var args []interface{}

	if req.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *req.Name)
	}
	if req.Description != nil {
		setParts = append(setParts, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Color != nil {
		setParts = append(setParts, "color = ?")
		args = append(args, *req.Color)
	}
	if req.Icon != nil {
		setParts = append(setParts, "icon = ?")
		args = append(args, *req.Icon)
	}
	if req.ParentID != nil {
		setParts = append(setParts, "parent_id = ?")
		args = append(args, *req.ParentID)
	}
	if req.SortOrder != nil {
		setParts = append(setParts, "sort_order = ?")
		args = append(args, *req.SortOrder)
	}

	if len(setParts) == 0 {
		return r.GetByID(id) // No changes, return existing
	}

	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE folders SET %s WHERE id = ?", strings.Join(setParts, ", "))
	
	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update folder: %w", err)
	}

	return r.GetByID(id)
}

// Delete deletes a folder and all its subfolders
func (r *FolderRepository) Delete(id int) error {
	// Check if folder exists
	if _, err := r.GetByID(id); err != nil {
		return err
	}

	// SQLite will handle cascade deletion of subfolders and bookmark_folders
	_, err := r.db.Exec("DELETE FROM folders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete folder: %w", err)
	}

	return nil
}

// AddBookmarkToFolder adds a bookmark to a folder
func (r *FolderRepository) AddBookmarkToFolder(bookmarkID, folderID int) error {
	_, err := r.db.Exec(
		"INSERT OR IGNORE INTO bookmark_folders (bookmark_id, folder_id) VALUES (?, ?)",
		bookmarkID, folderID,
	)
	if err != nil {
		return fmt.Errorf("failed to add bookmark to folder: %w", err)
	}
	return nil
}

// RemoveBookmarkFromFolder removes a bookmark from a folder
func (r *FolderRepository) RemoveBookmarkFromFolder(bookmarkID, folderID int) error {
	_, err := r.db.Exec(
		"DELETE FROM bookmark_folders WHERE bookmark_id = ? AND folder_id = ?",
		bookmarkID, folderID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove bookmark from folder: %w", err)
	}
	return nil
}

// GetBookmarkFolders retrieves all folders containing a bookmark
func (r *FolderRepository) GetBookmarkFolders(bookmarkID int) ([]*models.Folder, error) {
	query := `
		SELECT f.id, f.name, f.description, f.color, f.icon, f.parent_id, f.sort_order,
		       f.created_at, f.updated_at, 0 as bookmark_count
		FROM folders f
		JOIN bookmark_folders bf ON f.id = bf.folder_id
		WHERE bf.bookmark_id = ?
		ORDER BY f.sort_order, f.name
	`

	rows, err := r.db.Query(query, bookmarkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmark folders: %w", err)
	}
	defer rows.Close()

	var folders []*models.Folder
	for rows.Next() {
		var folder models.Folder
		err := rows.Scan(
			&folder.ID, &folder.Name, &folder.Description, &folder.Color, &folder.Icon,
			&folder.ParentID, &folder.SortOrder, &folder.CreatedAt, &folder.UpdatedAt,
			&folder.BookmarkCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}

		folder.Path, _ = r.buildFolderPath(folder.ID)
		folders = append(folders, &folder)
	}

	return folders, nil
}

// Helper function to build folder path
func (r *FolderRepository) buildFolderPath(id int) (string, error) {
	var path []string
	currentID := &id

	// Traverse up the hierarchy
	for currentID != nil {
		var name string
		var parentID *int
		
		err := r.db.QueryRow(
			"SELECT name, parent_id FROM folders WHERE id = ?", 
			*currentID,
		).Scan(&name, &parentID)
		
		if err != nil {
			return "", err
		}

		path = append([]string{name}, path...) // Prepend to build path from root
		currentID = parentID
	}

	return strings.Join(path, "/"), nil
}

// Helper function to build tree node recursively
func (r *FolderRepository) buildTreeNode(folder *models.Folder, folderMap map[int]*models.Folder) *models.FolderTree {
	node := &models.FolderTree{
		Folder: *folder,
	}

	// Find children
	for _, f := range folderMap {
		if f.ParentID != nil && *f.ParentID == folder.ID {
			childNode := r.buildTreeNode(f, folderMap)
			node.Children = append(node.Children, *childNode)
		}
	}

	return node
}