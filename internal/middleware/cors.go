package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORSMiddleware adds CORS headers for browser extension and frontend compatibility
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowedOrigin := getAllowedOrigin(origin)
		
		// Set CORS headers with dynamic origin
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
		
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// HTTPS enforcement in production
		if os.Getenv("ENV") == "production" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// getAllowedOrigin returns the appropriate origin based on environment and request
func getAllowedOrigin(requestOrigin string) string {
	// Get allowed origins from environment variable
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		// Default allowed origins for development and common deployments
		allowedOrigins = "http://localhost:3000,http://localhost:5173,http://localhost:8080,https://bookmark.arcbjorn.com,https://bookmarks.yourdomain.com"
	}

	// Split allowed origins
	origins := strings.Split(allowedOrigins, ",")
	
	// Check if request origin is in allowed list
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == requestOrigin {
			return requestOrigin
		}
		
		// Allow wildcard subdomains in production
		if strings.HasPrefix(origin, "*.") && requestOrigin != "" {
			domain := origin[2:] // Remove *.
			if strings.HasSuffix(requestOrigin, domain) {
				return requestOrigin
			}
		}
	}
	
	// For browser extensions (chrome-extension://, moz-extension://)
	if strings.HasPrefix(requestOrigin, "chrome-extension://") || 
	   strings.HasPrefix(requestOrigin, "moz-extension://") ||
	   strings.HasPrefix(requestOrigin, "safari-web-extension://") {
		return requestOrigin
	}
	
	// Default to first allowed origin if no match
	if len(origins) > 0 {
		return strings.TrimSpace(origins[0])
	}
	
	// Fallback for development
	if os.Getenv("ENV") != "production" {
		return "*"
	}
	
	return "null"
}