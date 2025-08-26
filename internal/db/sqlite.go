package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps the sql.DB connection with helper methods
type DB struct {
	*sql.DB
	dbPath string
}

// NewDB creates a new database connection and runs migrations
func NewDB(dbPath string) (*DB, error) {
	// Use os.Root for safer file operations with security boundaries
	root, err := os.OpenRoot(".")
	if err != nil {
		// Fallback to standard operations if os.Root fails
		if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	} else {
		// Use os.Root for secure directory creation
		if err := root.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	// Open database connection
	sqlDB, err := sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{
		DB:     sqlDB,
		dbPath: dbPath,
	}

	// Run migrations
	if err := db.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Printf("Database initialized at %s", dbPath)
	return db, nil
}

// runMigrations executes all migration files in order
func (db *DB) runMigrations() error {
	// Create migrations table if it doesn't exist
	if err := db.createMigrationsTable(); err != nil {
		return err
	}

	// Get list of applied migrations
	appliedMigrations, err := db.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Read migration files from embedded filesystem
	migrationFiles, err := db.getMigrationFiles()
	if err != nil {
		return err
	}

	// Execute pending migrations
	for _, filename := range migrationFiles {
		if appliedMigrations[filename] {
			continue // Skip already applied migrations
		}

		log.Printf("Running migration: %s", filename)
		
		// Read migration content
		content, err := migrationsFS.ReadFile(filepath.Join("migrations", filename))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute migration in a transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", filename, err)
		}

		// Execute migration SQL
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		// Record migration as applied
		if _, err := tx.Exec("INSERT INTO migrations (filename, applied_at) VALUES (?, datetime('now'))", filename); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}

		log.Printf("Migration completed: %s", filename)
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (db *DB) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			filename TEXT NOT NULL UNIQUE,
			applied_at DATETIME NOT NULL
		)
	`
	_, err := db.Exec(query)
	return err
}

// getAppliedMigrations returns a map of applied migration filenames
func (db *DB) getAppliedMigrations() (map[string]bool, error) {
	rows, err := db.Query("SELECT filename FROM migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}

	return applied, rows.Err()
}

// getMigrationFiles returns sorted list of migration filenames
func (db *DB) getMigrationFiles() ([]string, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	// Sort to ensure consistent order
	sort.Strings(files)
	return files, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// GetDBStats returns database statistics
func (db *DB) GetDBStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get table counts
	tables := []struct {
		name  string
		table string
	}{
		{"bookmarks", "bookmarks"},
		{"tags", "tags"},
		{"bookmark_tags", "bookmark_tags"},
		{"learned_patterns", "learned_patterns"},
		{"tag_corrections", "tag_corrections"},
		{"domain_profiles", "domain_profiles"},
	}

	for _, t := range tables {
		var count int
		err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", t.table)).Scan(&count)
		if err != nil {
			continue // Table might not exist yet
		}
		stats[t.name] = count
	}

	// Get database file size
	if fileInfo, err := os.Stat(db.dbPath); err == nil {
		stats["file_size_bytes"] = fileInfo.Size()
	}

	// Get SQLite version
	var version string
	if err := db.QueryRow("SELECT sqlite_version()").Scan(&version); err == nil {
		stats["sqlite_version"] = version
	}

	return stats, nil
}