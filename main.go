package main

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"torimemo/internal/db"
	"torimemo/internal/handlers"
	"torimemo/internal/logger"
	"torimemo/internal/middleware"
)

//go:embed web/dist/*
var staticFiles embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./torimemo.db"
	}

	// Initialize database
	database, err := db.NewDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize repositories
	bookmarkRepo := db.NewBookmarkRepository(database)
	tagRepo := db.NewTagRepository(database)

	// Initialize handlers
	bookmarkHandler := handlers.NewBookmarkHandler(bookmarkRepo)
	tagHandler := handlers.NewTagHandler(tagRepo)
	importExportHandler := handlers.NewImportExportHandler(bookmarkRepo, tagRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(bookmarkRepo, tagRepo, database)
	advancedSearchHandler := handlers.NewAdvancedSearchHandler(bookmarkRepo)

	// Initialize logger
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		logger.SetLevel(logger.DEBUG)
	case "WARN":
		logger.SetLevel(logger.WARN)  
	case "ERROR":
		logger.SetLevel(logger.ERROR)
	default:
		logger.SetLevel(logger.INFO)
	}

	logger.Info("Starting Torimemo server", map[string]interface{}{
		"port": port,
		"db_path": dbPath,
		"log_level": logLevel,
	})

	// Get the embedded filesystem
	webFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		log.Fatal("Failed to create sub filesystem:", err)
	}

	// Setup routes
	mux := http.NewServeMux()

	// API routes
	mux.Handle("/api/bookmarks", bookmarkHandler)
	mux.Handle("/api/bookmarks/", bookmarkHandler)
	mux.Handle("/api/tags", tagHandler)
	mux.Handle("/api/tags/", tagHandler)
	mux.Handle("/api/export", importExportHandler)
	mux.Handle("/api/import", importExportHandler)
	mux.Handle("/api/analytics", analyticsHandler)
	mux.Handle("/api/search/advanced", advancedSearchHandler)
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/stats", handleStats(database))

	// Serve static files and SPA
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle SPA routing - serve index.html for routes that don't exist
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists
		if _, err := fs.Stat(webFS, path); err != nil {
			// File doesn't exist, serve index.html for SPA routing
			path = "index.html"
		}

		// Serve the file
		file, err := webFS.Open(path)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer file.Close()

		// Get file info for content type
		info, err := file.Stat()
		if err != nil {
			http.Error(w, "File info error", http.StatusInternalServerError)
			return
		}

		// Set content type based on file extension
		ext := filepath.Ext(path)
		switch ext {
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		}

		http.ServeContent(w, r, info.Name(), info.ModTime(), file.(interface{
			io.Reader
			io.Seeker
			io.Closer
		}))
	})

	// Apply middleware
	handler := middleware.LoggingMiddleware(middleware.CORSMiddleware(mux))

	fmt.Printf("🚀 Torimemo server starting on port %s\n", port)
	fmt.Printf("📊 Database: %s\n", dbPath)
	fmt.Printf("📍 http://localhost:%s\n", port)
	fmt.Printf("🔍 API: http://localhost:%s/api/health\n", port)
	
	logger.Info("Server ready to accept connections")
	log.Fatal(http.ListenAndServe(":"+port, handler))
}

// handleHealth provides API health check
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{
		"status": "ok", 
		"message": "Torimemo API is running",
		"version": "1.0.0",
		"features": [
			"bookmark_management",
			"full_text_search", 
			"tag_system",
			"learning_system_ready"
		]
	}`)
}

// handleStats provides database statistics
func handleStats(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		stats, err := database.GetDBStats()
		if err != nil {
			http.Error(w, `{"error":"Failed to get stats"}`, http.StatusInternalServerError)
			return
		}

		// Convert to JSON manually for simple response
		fmt.Fprintf(w, `{
			"bookmarks": %v,
			"tags": %v,
			"bookmark_tags": %v,
			"learned_patterns": %v,
			"tag_corrections": %v,
			"domain_profiles": %v,
			"file_size_bytes": %v,
			"sqlite_version": "%v"
		}`, 
			getIntValue(stats, "bookmarks"),
			getIntValue(stats, "tags"), 
			getIntValue(stats, "bookmark_tags"),
			getIntValue(stats, "learned_patterns"),
			getIntValue(stats, "tag_corrections"),
			getIntValue(stats, "domain_profiles"),
			getIntValue(stats, "file_size_bytes"),
			getStringValue(stats, "sqlite_version"),
		)
	}
}

func getIntValue(stats map[string]interface{}, key string) int {
	if val, ok := stats[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
		if int64Val, ok := val.(int64); ok {
			return int(int64Val)
		}
	}
	return 0
}

func getStringValue(stats map[string]interface{}, key string) string {
	if val, ok := stats[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}