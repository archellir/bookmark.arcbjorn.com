-- Learning system tables for AI auto-categorization

-- Store learned patterns with confidence scores
CREATE TABLE learned_patterns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url_pattern TEXT NOT NULL,          -- "youtube.com/*/react*"
    domain TEXT NOT NULL,               -- "youtube.com"
    confirmed_tags TEXT NOT NULL,       -- JSON: ["Tutorial", "React"]
    rejected_tags TEXT DEFAULT '[]',    -- JSON: ["Video"]
    confidence REAL DEFAULT 0.0,       -- 0.0 to 1.0
    sample_count INTEGER DEFAULT 1,    -- How many times pattern seen
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Track user corrections for learning
CREATE TABLE tag_corrections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    bookmark_id INTEGER NOT NULL,
    original_tags TEXT NOT NULL,        -- AI suggested tags (JSON)
    final_tags TEXT NOT NULL,          -- User final tags (JSON)
    correction_type TEXT NOT NULL,     -- "kept", "added", "removed"
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (bookmark_id) REFERENCES bookmarks(id) ON DELETE CASCADE
);

-- Domain-specific user preferences
CREATE TABLE domain_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain TEXT NOT NULL UNIQUE,
    common_tags TEXT DEFAULT '[]',      -- JSON: frequently used tags
    ignored_tags TEXT DEFAULT '[]',     -- JSON: tags user never uses
    custom_mappings TEXT DEFAULT '{}', -- JSON: user's preferred tag names
    bookmark_count INTEGER DEFAULT 0,  -- Number of bookmarks for this domain
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for learning system
CREATE INDEX idx_learned_patterns_domain ON learned_patterns(domain);
CREATE INDEX idx_learned_patterns_url_pattern ON learned_patterns(url_pattern);
CREATE INDEX idx_learned_patterns_confidence ON learned_patterns(confidence);
CREATE INDEX idx_tag_corrections_bookmark_id ON tag_corrections(bookmark_id);
CREATE INDEX idx_tag_corrections_created_at ON tag_corrections(created_at);
CREATE INDEX idx_domain_profiles_domain ON domain_profiles(domain);
CREATE INDEX idx_domain_profiles_bookmark_count ON domain_profiles(bookmark_count);