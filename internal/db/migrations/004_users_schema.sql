-- Add user authentication support

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

-- Add user_id to existing tables
ALTER TABLE bookmarks ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE tags ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE folders ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;

-- Create indexes for performance
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_is_active ON users(is_active);
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);
CREATE INDEX idx_tags_user_id ON tags(user_id);
CREATE INDEX idx_folders_user_id ON folders(user_id);

-- Create sessions table for JWT blacklisting (optional security feature)
CREATE TABLE user_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_revoked BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_token_hash ON user_sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX idx_sessions_user_id ON user_sessions(user_id);

-- Create trigger to update user updated_at timestamp
CREATE TRIGGER users_update_timestamp 
AFTER UPDATE ON users 
FOR EACH ROW 
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Create default admin user (password: 'admin123' - should be changed)
-- Password hash for 'admin123' using bcrypt cost 12
INSERT INTO users (username, email, password_hash, full_name, is_admin) VALUES 
    ('admin', 'admin@torimemo.app', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMaQJfO69K3WbL9gqvKqJ1VxZm', 'Administrator', TRUE);

-- Update existing data to belong to admin user (user_id = 1)
UPDATE bookmarks SET user_id = 1 WHERE user_id IS NULL;
UPDATE tags SET user_id = 1 WHERE user_id IS NULL;  
UPDATE folders SET user_id = 1 WHERE user_id IS NULL;