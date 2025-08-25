package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"torimemo/internal/db"
	"torimemo/internal/models"
)

// HealthStatus represents the health status of a bookmark
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusBroken    HealthStatus = "broken"
	HealthStatusSlow      HealthStatus = "slow"
	HealthStatusRedirect  HealthStatus = "redirect"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// BookmarkHealth represents the health check result for a bookmark
type BookmarkHealth struct {
	ID           int          `json:"id"`
	URL          string       `json:"url"`
	Status       HealthStatus `json:"status"`
	StatusCode   int          `json:"status_code"`
	ResponseTime int64        `json:"response_time_ms"`
	RedirectURL  string       `json:"redirect_url,omitempty"`
	Error        string       `json:"error,omitempty"`
	LastChecked  time.Time    `json:"last_checked"`
}

// HealthChecker manages bookmark health checking
type HealthChecker struct {
	repo           *db.BookmarkRepository
	client         *http.Client
	checkInterval  time.Duration
	batchSize      int
	maxConcurrent  int
	slowThreshold  time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.RWMutex
	healthData     map[int]*BookmarkHealth
	running        bool
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(repo *db.BookmarkRepository) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &HealthChecker{
		repo:          repo,
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Allow up to 5 redirects
				if len(via) >= 5 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
		checkInterval: 24 * time.Hour, // Check once per day
		batchSize:     50,             // Process 50 bookmarks at a time
		maxConcurrent: 10,             // Max 10 concurrent requests
		slowThreshold: 5 * time.Second, // Consider > 5s as slow
		ctx:           ctx,
		cancel:        cancel,
		healthData:    make(map[int]*BookmarkHealth),
		running:       false,
	}
}

// Start begins the health checking service
func (hc *HealthChecker) Start() {
	hc.mu.Lock()
	if hc.running {
		hc.mu.Unlock()
		return
	}
	hc.running = true
	hc.mu.Unlock()

	log.Println("Starting bookmark health checker...")
	
	// Initial check
	go hc.runHealthCheck()
	
	// Periodic checks
	ticker := time.NewTicker(hc.checkInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-hc.ctx.Done():
				return
			case <-ticker.C:
				go hc.runHealthCheck()
			}
		}
	}()
}

// Stop stops the health checking service
func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	
	if !hc.running {
		return
	}
	
	log.Println("Stopping bookmark health checker...")
	hc.cancel()
	hc.running = false
}

// runHealthCheck performs a health check on all bookmarks
func (hc *HealthChecker) runHealthCheck() {
	log.Println("Running bookmark health check...")
	
	// Get all bookmarks
	response, err := hc.repo.ListAllUsers(1, 10000, "", "", false) // Get all bookmarks
	if err != nil {
		log.Printf("Failed to get bookmarks for health check: %v", err)
		return
	}
	
	bookmarks := response.Bookmarks
	log.Printf("Checking health of %d bookmarks", len(bookmarks))
	
	// Process in batches
	for i := 0; i < len(bookmarks); i += hc.batchSize {
		end := i + hc.batchSize
		if end > len(bookmarks) {
			end = len(bookmarks)
		}
		
		batch := bookmarks[i:end]
		hc.checkBatch(batch)
		
		// Small delay between batches to avoid overwhelming servers
		time.Sleep(1 * time.Second)
	}
	
	log.Printf("Health check completed. Checked %d bookmarks", len(bookmarks))
}

// checkBatch checks a batch of bookmarks concurrently
func (hc *HealthChecker) checkBatch(bookmarks []models.Bookmark) {
	semaphore := make(chan struct{}, hc.maxConcurrent)
	var wg sync.WaitGroup
	
	for _, bookmark := range bookmarks {
		wg.Add(1)
		go func(b models.Bookmark) {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			hc.checkBookmark(b)
		}(bookmark)
	}
	
	wg.Wait()
}

// checkBookmark checks the health of a single bookmark
func (hc *HealthChecker) checkBookmark(bookmark models.Bookmark) {
	health := &BookmarkHealth{
		ID:          bookmark.ID,
		URL:         bookmark.URL,
		Status:      HealthStatusUnknown,
		LastChecked: time.Now(),
	}
	
	start := time.Now()
	
	// Create request with context for timeout
	ctx, cancel := context.WithTimeout(hc.ctx, 10*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "HEAD", bookmark.URL, nil)
	if err != nil {
		health.Status = HealthStatusBroken
		health.Error = fmt.Sprintf("Invalid URL: %v", err)
		hc.storeHealth(health)
		return
	}
	
	// Set user agent
	req.Header.Set("User-Agent", "Torimemo-HealthChecker/1.0")
	
	// Perform request
	resp, err := hc.client.Do(req)
	if err != nil {
		health.Status = HealthStatusBroken
		health.Error = err.Error()
		hc.storeHealth(health)
		return
	}
	defer resp.Body.Close()
	
	// Calculate response time
	responseTime := time.Since(start)
	health.ResponseTime = responseTime.Milliseconds()
	health.StatusCode = resp.StatusCode
	
	// Determine health status
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		if responseTime > hc.slowThreshold {
			health.Status = HealthStatusSlow
		} else {
			health.Status = HealthStatusHealthy
		}
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		health.Status = HealthStatusRedirect
		if location := resp.Header.Get("Location"); location != "" {
			health.RedirectURL = location
		}
	case resp.StatusCode >= 400:
		health.Status = HealthStatusBroken
		health.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	default:
		health.Status = HealthStatusUnknown
		health.Error = fmt.Sprintf("Unexpected status code: %d", resp.StatusCode)
	}
	
	hc.storeHealth(health)
}

// storeHealth stores health data in memory
func (hc *HealthChecker) storeHealth(health *BookmarkHealth) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	
	hc.healthData[health.ID] = health
	
	// Log issues
	if health.Status == HealthStatusBroken {
		log.Printf("Broken link detected: %s (%s)", health.URL, health.Error)
	} else if health.Status == HealthStatusSlow {
		log.Printf("Slow link detected: %s (%dms)", health.URL, health.ResponseTime)
	}
}

// GetHealth returns health data for a specific bookmark
func (hc *HealthChecker) GetHealth(bookmarkID int) *BookmarkHealth {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	return hc.healthData[bookmarkID]
}

// GetAllHealth returns all health data
func (hc *HealthChecker) GetAllHealth() map[int]*BookmarkHealth {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	// Create a copy to avoid race conditions
	result := make(map[int]*BookmarkHealth)
	for id, health := range hc.healthData {
		result[id] = health
	}
	
	return result
}

// GetHealthStats returns summary statistics
func (hc *HealthChecker) GetHealthStats() map[string]int {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	stats := map[string]int{
		"total":     0,
		"healthy":   0,
		"broken":    0,
		"slow":      0,
		"redirect":  0,
		"unknown":   0,
		"unchecked": 0,
	}
	
	// Get total bookmarks count
	response, err := hc.repo.ListAllUsers(1, 10000, "", "", false)
	if err == nil {
		stats["total"] = len(response.Bookmarks)
	}
	
	// Count by status
	for _, health := range hc.healthData {
		switch health.Status {
		case HealthStatusHealthy:
			stats["healthy"]++
		case HealthStatusBroken:
			stats["broken"]++
		case HealthStatusSlow:
			stats["slow"]++
		case HealthStatusRedirect:
			stats["redirect"]++
		case HealthStatusUnknown:
			stats["unknown"]++
		}
	}
	
	stats["unchecked"] = stats["total"] - len(hc.healthData)
	
	return stats
}

// CheckBookmarkNow immediately checks a specific bookmark
func (hc *HealthChecker) CheckBookmarkNow(bookmarkID int) *BookmarkHealth {
	// Get bookmark from database
	bookmark, err := hc.repo.GetByID(bookmarkID, 0) // TODO: Pass actual user ID
	if err != nil {
		return &BookmarkHealth{
			ID:          bookmarkID,
			Status:      HealthStatusBroken,
			Error:       "Bookmark not found",
			LastChecked: time.Now(),
		}
	}
	
	hc.checkBookmark(*bookmark)
	return hc.GetHealth(bookmarkID)
}

// GetBrokenBookmarks returns all bookmarks with broken links
func (hc *HealthChecker) GetBrokenBookmarks() []*BookmarkHealth {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	var broken []*BookmarkHealth
	for _, health := range hc.healthData {
		if health.Status == HealthStatusBroken {
			broken = append(broken, health)
		}
	}
	
	return broken
}

// IsValidURL performs basic URL validation
func IsValidURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}