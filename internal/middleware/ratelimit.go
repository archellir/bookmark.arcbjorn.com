package middleware

import (
	"net/http"
	"sync"
	"time"
	"encoding/json"
	"strings"
	"net"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	clients map[string]*client
	mutex   *sync.RWMutex
	rate    time.Duration // how often to replenish tokens
	capacity int         // maximum number of tokens
	cleanup  time.Duration // how often to cleanup old clients
}

type client struct {
	tokens    int
	lastSeen  time.Time
	mutex     *sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: requests per period (e.g., 100 requests per minute)
// period: time period (e.g., time.Minute)
// burst: maximum burst requests allowed
func NewRateLimiter(rate int, period time.Duration, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients:  make(map[string]*client),
		mutex:    &sync.RWMutex{},
		rate:     period / time.Duration(rate),
		capacity: burst,
		cleanup:  time.Minute * 5, // cleanup every 5 minutes
	}

	// Start cleanup goroutine
	go rl.cleanupClients()

	return rl
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := rl.getClientIP(r)

		// Check if request is allowed
		if !rl.allow(clientIP) {
			rl.writeRateLimitError(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allow checks if the client is allowed to make a request
func (rl *RateLimiter) allow(clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	c, exists := rl.clients[clientIP]
	if !exists {
		// New client
		c = &client{
			tokens:   rl.capacity - 1, // consume one token
			lastSeen: time.Now(),
			mutex:    &sync.Mutex{},
		}
		rl.clients[clientIP] = c
		return true
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Calculate tokens to add based on time passed
	now := time.Now()
	elapsed := now.Sub(c.lastSeen)
	tokensToAdd := int(elapsed / rl.rate)

	if tokensToAdd > 0 {
		c.tokens += tokensToAdd
		if c.tokens > rl.capacity {
			c.tokens = rl.capacity
		}
		c.lastSeen = now
	}

	// Check if we have tokens available
	if c.tokens > 0 {
		c.tokens--
		c.lastSeen = now
		return true
	}

	return false
}

// getClientIP extracts the client IP address from the request
func (rl *RateLimiter) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for reverse proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP if there are multiple
		if idx := strings.Index(forwarded, ","); idx != -1 {
			forwarded = forwarded[:idx]
		}
		forwarded = strings.TrimSpace(forwarded)
		if forwarded != "" {
			return forwarded
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return strings.TrimSpace(realIP)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	
	return ip
}

// writeRateLimitError writes a rate limit exceeded response
func (rl *RateLimiter) writeRateLimitError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "60") // Suggest retry after 60 seconds
	w.WriteHeader(http.StatusTooManyRequests)

	response := map[string]interface{}{
		"error":   "Rate limit exceeded",
		"message": "Too many requests. Please try again later.",
		"status":  http.StatusTooManyRequests,
	}

	json.NewEncoder(w).Encode(response)
}

// cleanupClients removes old inactive clients to prevent memory leaks
func (rl *RateLimiter) cleanupClients() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mutex.Lock()
			
			cutoff := time.Now().Add(-time.Hour) // Remove clients inactive for 1 hour
			for ip, client := range rl.clients {
				client.mutex.Lock()
				if client.lastSeen.Before(cutoff) {
					delete(rl.clients, ip)
				}
				client.mutex.Unlock()
			}
			
			rl.mutex.Unlock()
		}
	}
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	activeClients := 0
	totalTokens := 0

	for _, client := range rl.clients {
		client.mutex.Lock()
		activeClients++
		totalTokens += client.tokens
		client.mutex.Unlock()
	}

	return map[string]interface{}{
		"active_clients": activeClients,
		"total_tokens":   totalTokens,
		"rate_per_sec":   float64(time.Second) / float64(rl.rate),
		"capacity":       rl.capacity,
	}
}