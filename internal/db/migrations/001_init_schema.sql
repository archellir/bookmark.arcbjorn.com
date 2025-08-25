-- Complete initial schema for Torimemo bookmark manager with authentication

-- Create users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    full_name TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login_at DATETIME
);

-- Create bookmarks table with user association
CREATE TABLE bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    description TEXT,
    favicon_url TEXT,
    user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_favorite BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(url, user_id)
);

-- Create tags table with user association
CREATE TABLE tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    color TEXT DEFAULT '#007acc',
    user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(name, user_id)
);

-- Create folders table with hierarchical support and user association
CREATE TABLE folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    color TEXT DEFAULT '#666666',
    icon TEXT DEFAULT '📁',
    parent_id INTEGER REFERENCES folders(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL,
    sort_order INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(name, parent_id, user_id)
);

-- Create many-to-many relationship between bookmarks and tags
CREATE TABLE bookmark_tags (
    bookmark_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (bookmark_id, tag_id),
    FOREIGN KEY (bookmark_id) REFERENCES bookmarks(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
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

-- Create sessions table for JWT blacklisting
CREATE TABLE user_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_revoked BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Learning system tables for AI auto-categorization
CREATE TABLE learned_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url_pattern TEXT NOT NULL,
    domain TEXT NOT NULL,
    confirmed_tags TEXT NOT NULL,
    rejected_tags TEXT DEFAULT '[]',
    confidence REAL DEFAULT 0.0,
    sample_count INTEGER DEFAULT 1,
    user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Track user corrections for learning
CREATE TABLE tag_corrections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bookmark_id INTEGER NOT NULL,
    original_tags TEXT NOT NULL,
    final_tags TEXT NOT NULL,
    correction_type TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bookmark_id) REFERENCES bookmarks(id) ON DELETE CASCADE
);

-- Domain-specific user preferences
CREATE TABLE domain_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain TEXT NOT NULL,
    common_tags TEXT DEFAULT '[]',
    ignored_tags TEXT DEFAULT '[]',
    custom_mappings TEXT DEFAULT '{}',
    bookmark_count INTEGER DEFAULT 0,
    user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(domain, user_id)
);

-- Create indexes for performance

-- User indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Bookmark indexes
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_bookmarks_created_at ON bookmarks(created_at);
CREATE INDEX idx_bookmarks_is_favorite ON bookmarks(is_favorite);
CREATE INDEX idx_bookmarks_url ON bookmarks(url);

-- Tag indexes
CREATE INDEX idx_tags_user_id ON tags(user_id);
CREATE INDEX idx_tags_name ON tags(name);

-- Folder indexes
CREATE INDEX idx_folders_user_id ON folders(user_id);
CREATE INDEX idx_folders_parent_id ON folders(parent_id);
CREATE INDEX idx_folders_name ON folders(name);
CREATE INDEX idx_folders_sort_order ON folders(sort_order);

-- Junction table indexes
CREATE INDEX idx_bookmark_tags_bookmark_id ON bookmark_tags(bookmark_id);
CREATE INDEX idx_bookmark_tags_tag_id ON bookmark_tags(tag_id);
CREATE INDEX idx_bookmark_folders_bookmark_id ON bookmark_folders(bookmark_id);
CREATE INDEX idx_bookmark_folders_folder_id ON bookmark_folders(folder_id);

-- Session indexes
CREATE INDEX idx_sessions_token_hash ON user_sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);

-- Learning system indexes
CREATE INDEX idx_learned_patterns_user_id ON learned_patterns(user_id);
CREATE INDEX idx_learned_patterns_domain ON learned_patterns(domain);
CREATE INDEX idx_learned_patterns_url_pattern ON learned_patterns(url_pattern);
CREATE INDEX idx_learned_patterns_confidence ON learned_patterns(confidence);
CREATE INDEX idx_tag_corrections_bookmark_id ON tag_corrections(bookmark_id);
CREATE INDEX idx_tag_corrections_created_at ON tag_corrections(created_at);
CREATE INDEX idx_domain_profiles_user_id ON domain_profiles(user_id);
CREATE INDEX idx_domain_profiles_domain ON domain_profiles(domain);
CREATE INDEX idx_domain_profiles_bookmark_count ON domain_profiles(bookmark_count);

-- Create full-text search virtual table
CREATE VIRTUAL TABLE bookmarks_fts USING fts5(
    title, 
    url, 
    description,
    content=bookmarks,
    content_rowid=id
);

-- Create triggers to keep FTS table in sync
CREATE TRIGGER bookmarks_ai AFTER INSERT ON bookmarks BEGIN
  INSERT INTO bookmarks_fts(rowid, title, url, description) 
  VALUES (new.id, new.title, new.url, new.description);
END;

CREATE TRIGGER bookmarks_ad AFTER DELETE ON bookmarks BEGIN
  INSERT INTO bookmarks_fts(bookmarks_fts, rowid, title, url, description) 
  VALUES('delete', old.id, old.title, old.url, old.description);
END;

CREATE TRIGGER bookmarks_au AFTER UPDATE ON bookmarks BEGIN
  INSERT INTO bookmarks_fts(bookmarks_fts, rowid, title, url, description) 
  VALUES('delete', old.id, old.title, old.url, old.description);
  INSERT INTO bookmarks_fts(rowid, title, url, description) 
  VALUES (new.id, new.title, new.url, new.description);
END;

-- Create trigger to update user updated_at timestamp
CREATE TRIGGER users_update_timestamp 
AFTER UPDATE ON users 
FOR EACH ROW 
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create trigger to update folder updated_at timestamp
CREATE TRIGGER folders_update_timestamp 
AFTER UPDATE ON folders 
FOR EACH ROW 
BEGIN
    UPDATE folders SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create default admin user (password: 'admin123' - should be changed)
-- Password hash for 'admin123' using bcrypt cost 12
INSERT INTO users (username, email, password_hash, full_name, is_admin) VALUES 
    ('admin', 'admin@torimemo.app', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMaQJfO69K3WbL9gqvKqJ1VxZm', 'Administrator', TRUE);

-- Create some default folders for the admin user
INSERT INTO folders (name, description, color, icon, user_id, sort_order) VALUES 
    ('Work', 'Work-related bookmarks', '#007acc', '💼', 1, 1),
    ('Personal', 'Personal bookmarks', '#28a745', '👤', 1, 2),
    ('Learning', 'Educational resources', '#ffc107', '📚', 1, 3),
    ('Tools', 'Development tools and utilities', '#6f42c1', '🔧', 1, 4),
    ('Archive', 'Old or rarely accessed bookmarks', '#6c757d', '🗄️', 1, 5);