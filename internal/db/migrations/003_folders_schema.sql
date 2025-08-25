-- Add bookmark folders/collections support

-- Create folders table with hierarchical support
CREATE TABLE folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    color TEXT DEFAULT '#666666',
    icon TEXT DEFAULT '📁',
    parent_id INTEGER REFERENCES folders(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Ensure folder names are unique within the same parent
    UNIQUE(name, parent_id)
);

-- Create many-to-many relationship between bookmarks and folders
CREATE TABLE bookmark_folders (
    bookmark_id INTEGER NOT NULL,
    folder_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (bookmark_id, folder_id),
    FOREIGN KEY (bookmark_id) REFERENCES bookmarks(id) ON DELETE CASCADE,
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_folders_parent_id ON folders(parent_id);
CREATE INDEX idx_folders_name ON folders(name);
CREATE INDEX idx_folders_sort_order ON folders(sort_order);
CREATE INDEX idx_bookmark_folders_bookmark_id ON bookmark_folders(bookmark_id);
CREATE INDEX idx_bookmark_folders_folder_id ON bookmark_folders(folder_id);

-- Create some default folders
INSERT INTO folders (name, description, color, icon, sort_order) VALUES 
    ('Work', 'Work-related bookmarks', '#007acc', '💼', 1),
    ('Personal', 'Personal bookmarks', '#28a745', '👤', 2),
    ('Learning', 'Educational resources', '#ffc107', '📚', 3),
    ('Tools', 'Development tools and utilities', '#6f42c1', '🔧', 4),
    ('Archive', 'Old or rarely accessed bookmarks', '#6c757d', '🗄️', 5);

-- Create trigger to update folder updated_at timestamp
CREATE TRIGGER folders_update_timestamp 
AFTER UPDATE ON folders 
FOR EACH ROW 
BEGIN
    UPDATE folders SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;