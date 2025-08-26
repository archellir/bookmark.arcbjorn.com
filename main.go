package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"torimemo/internal/db"
	"torimemo/internal/handlers"
	"torimemo/internal/logger"
	"torimemo/internal/services"
	"torimemo/internal/middleware"
)

//go:embed web/dist/*
var staticFiles embed.FS

func main() {
	// Configure container-aware GOMAXPROCS for better performance in containerized environments
	runtime.SetDefaultGOMAXPROCS()
	
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
	learningRepo := db.NewLearningRepository(database)
	folderRepo := db.NewFolderRepository(database)
	userRepo := db.NewUserRepository(database)

	// Initialize handlers
	bookmarkHandler := handlers.NewBookmarkHandler(bookmarkRepo, learningRepo)
	tagHandler := handlers.NewTagHandler(tagRepo)
	importExportHandler := handlers.NewImportExportHandler(bookmarkRepo, tagRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(bookmarkRepo, tagRepo, database)
	advancedSearchHandler := handlers.NewAdvancedSearchHandler(bookmarkRepo)
	learningHandler := handlers.NewLearningHandler(bookmarkRepo, learningRepo)
	importHandler := handlers.NewImportHandler(bookmarkRepo)
	archiveHandler := handlers.NewArchiveHandler(bookmarkRepo)
	folderHandler := handlers.NewFolderHandler(folderRepo, bookmarkRepo)
	duplicateHandler := handlers.NewDuplicateHandler(bookmarkRepo)
	aiDuplicateHandler := handlers.NewAIDuplicatesHandler(bookmarkRepo)
	aiClusteringHandler := handlers.NewAIClusteringHandler(bookmarkRepo)
	aiPredictiveHandler := handlers.NewAIPredictiveHandler(bookmarkRepo, learningRepo)
	authHandler := handlers.NewAuthHandler(userRepo)
	aiFeedbackHandler := handlers.NewAIFeedbackHandler(bookmarkRepo, learningRepo)
	
	// Initialize health checker service
	healthChecker := services.NewHealthChecker(bookmarkRepo)
	healthHandler := handlers.NewHealthHandler(healthChecker)

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

	// Static files would normally be embedded here in production

	// Setup routes
	mux := http.NewServeMux()

	// API routes
	mux.Handle("/api/bookmarks", bookmarkHandler)
	mux.Handle("/api/bookmarks/", bookmarkHandler)
	mux.Handle("/api/tags", tagHandler)
	mux.Handle("/api/tags/", tagHandler)
	mux.Handle("/api/folders", folderHandler)
	mux.Handle("/api/folders/", folderHandler)
	mux.Handle("/api/export", importExportHandler)
	mux.Handle("/api/import", importExportHandler)
	mux.Handle("/api/analytics", analyticsHandler)
	mux.Handle("/api/search/advanced", advancedSearchHandler)
	mux.Handle("/api/learning", learningHandler)
	mux.Handle("/api/learning/", learningHandler)
	importHandler.RegisterRoutes(mux)
	healthHandler.RegisterRoutes(mux)
	archiveHandler.RegisterRoutes(mux)
	folderHandler.RegisterRoutes(mux)
	duplicateHandler.RegisterRoutes(mux)
	aiDuplicateHandler.RegisterRoutes(mux)
	aiClusteringHandler.RegisterRoutes(mux)
	aiPredictiveHandler.RegisterRoutes(mux)
	authHandler.RegisterRoutes(mux)
	aiFeedbackHandler.RegisterRoutes(mux)
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/stats", handleStats(database))

	// Static file serving would be implemented here in production
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Placeholder for static file serving
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Static files not configured in this build"))
	})

	// Initialize middlewares
	authMiddleware := middleware.NewAuthMiddleware(userRepo)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute, 20) // 100 req/min, burst of 20
	
	// Apply middleware stack
	corsHandler := middleware.CORSMiddleware(mux)
	rateLimitedHandler := rateLimiter.Middleware(corsHandler)
	optionalAuthHandler := authMiddleware.OptionalAuth(rateLimitedHandler)
	handler := middleware.LoggingMiddleware(optionalAuthHandler)

	fmt.Printf("🚀 Torimemo server starting on port %s\n", port)
	fmt.Printf("📊 Database: %s\n", dbPath)
	fmt.Printf("📍 http://localhost:%s\n", port)
	fmt.Printf("🔍 API: http://localhost:%s/api/health\n", port)
	
	logger.Info("Server ready to accept connections")
	
	// Start health checker service
	healthChecker.Start()
	defer healthChecker.Stop()
	
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
			"ai_auto_categorization",
			"learning_system_active"
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