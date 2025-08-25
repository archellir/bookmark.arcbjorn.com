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
)

//go:embed web/dist/*
var staticFiles embed.FS

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get the embedded filesystem
	webFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		log.Fatal("Failed to create sub filesystem:", err)
	}

	// Setup routes
	mux := http.NewServeMux()

	// API routes (to be implemented)
	mux.HandleFunc("/api/", handleAPI)

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

	fmt.Printf("🚀 Torimemo server starting on port %s\n", port)
	fmt.Printf("📍 http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Basic API response for now
	switch r.URL.Path {
	case "/api/health":
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok","message":"Torimemo API is running"}`)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"endpoint not found"}`)
	}
}