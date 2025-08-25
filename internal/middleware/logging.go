package middleware

import (
	"net/http"
	"time"
	"torimemo/internal/logger"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// LoggingMiddleware logs HTTP requests in structured format
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}
		
		// Process request
		next.ServeHTTP(wrapped, r)
		
		// Log request details
		duration := time.Since(start)
		
		logData := map[string]interface{}{
			"method":      r.Method,
			"url":         r.URL.String(),
			"status":      wrapped.statusCode,
			"bytes":       wrapped.bytes,
			"duration_ms": duration.Milliseconds(),
			"user_agent":  r.UserAgent(),
			"remote_addr": r.RemoteAddr,
		}
		
		// Add query params for GET requests
		if r.Method == "GET" && len(r.URL.RawQuery) > 0 {
			logData["query"] = r.URL.RawQuery
		}
		
		level := logger.INFO
		message := "HTTP Request"
		
		// Log errors and slow requests differently
		if wrapped.statusCode >= 400 {
			level = logger.WARN
			if wrapped.statusCode >= 500 {
				level = logger.ERROR
			}
			message = "HTTP Error"
		} else if duration > 1*time.Second {
			level = logger.WARN
			message = "Slow Request"
		}
		
		switch level {
		case logger.ERROR:
			logger.Error(message, logData)
		case logger.WARN:
			logger.Warn(message, logData)
		default:
			logger.Info(message, logData)
		}
	})
}