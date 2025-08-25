package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ContentFetcher handles fetching page content for bookmarks
type ContentFetcher struct {
	client *http.Client
}

// NewContentFetcher creates a new content fetcher
func NewContentFetcher() *ContentFetcher {
	return &ContentFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
			},
		},
	}
}

// PageContent represents extracted page content
type PageContent struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	FaviconURL  string `json:"favicon_url"`
	Language    string `json:"language"`
	Keywords    string `json:"keywords"`
}

// FetchContent fetches and extracts content from a URL
func (f *ContentFetcher) FetchContent(targetURL string) (*PageContent, error) {
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Create request with appropriate headers
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent to avoid bot blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Torimemo/1.0; +https://github.com/torimemo/bookmark-manager)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Make request
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") {
		return nil, fmt.Errorf("non-HTML content type: %s", contentType)
	}

	// Read response body (limit to 1MB)
	limitedReader := io.LimitReader(resp.Body, 1024*1024)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Extract content from HTML
	content := f.extractContent(string(body), parsedURL)
	
	return content, nil
}

// extractContent extracts title, description, and favicon from HTML
func (f *ContentFetcher) extractContent(html string, parsedURL *url.URL) *PageContent {
	content := &PageContent{}

	// Extract title
	content.Title = f.extractTitle(html)

	// Extract meta description
	content.Description = f.extractMetaDescription(html)

	// Extract meta keywords
	content.Keywords = f.extractMetaKeywords(html)

	// Extract language
	content.Language = f.extractLanguage(html)

	// Determine favicon URL
	content.FaviconURL = f.extractFaviconURL(html, parsedURL)

	return content
}

// extractTitle extracts the page title
func (f *ContentFetcher) extractTitle(html string) string {
	// Try <title> tag first
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]*)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		title := strings.TrimSpace(f.decodeHTMLEntities(matches[1]))
		if title != "" {
			return title
		}
	}

	// Try og:title meta tag
	ogTitleRegex := regexp.MustCompile(`<meta[^>]*property="og:title"[^>]*content="([^"]*)"[^>]*>`)
	matches = ogTitleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		title := strings.TrimSpace(f.decodeHTMLEntities(matches[1]))
		if title != "" {
			return title
		}
	}

	// Try twitter:title meta tag
	twitterTitleRegex := regexp.MustCompile(`<meta[^>]*name="twitter:title"[^>]*content="([^"]*)"[^>]*>`)
	matches = twitterTitleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		title := strings.TrimSpace(f.decodeHTMLEntities(matches[1]))
		if title != "" {
			return title
		}
	}

	return ""
}

// extractMetaDescription extracts the page description
func (f *ContentFetcher) extractMetaDescription(html string) string {
	// Try meta description tag
	descRegex := regexp.MustCompile(`<meta[^>]*name="description"[^>]*content="([^"]*)"[^>]*>`)
	matches := descRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		desc := strings.TrimSpace(f.decodeHTMLEntities(matches[1]))
		if desc != "" {
			return desc
		}
	}

	// Try og:description meta tag
	ogDescRegex := regexp.MustCompile(`<meta[^>]*property="og:description"[^>]*content="([^"]*)"[^>]*>`)
	matches = ogDescRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		desc := strings.TrimSpace(f.decodeHTMLEntities(matches[1]))
		if desc != "" {
			return desc
		}
	}

	// Try twitter:description meta tag
	twitterDescRegex := regexp.MustCompile(`<meta[^>]*name="twitter:description"[^>]*content="([^"]*)"[^>]*>`)
	matches = twitterDescRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		desc := strings.TrimSpace(f.decodeHTMLEntities(matches[1]))
		if desc != "" {
			return desc
		}
	}

	return ""
}

// extractMetaKeywords extracts meta keywords
func (f *ContentFetcher) extractMetaKeywords(html string) string {
	keywordsRegex := regexp.MustCompile(`<meta[^>]*name="keywords"[^>]*content="([^"]*)"[^>]*>`)
	matches := keywordsRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(f.decodeHTMLEntities(matches[1]))
	}
	return ""
}

// extractLanguage extracts the page language
func (f *ContentFetcher) extractLanguage(html string) string {
	// Try html lang attribute
	langRegex := regexp.MustCompile(`<html[^>]*lang="([^"]*)"[^>]*>`)
	matches := langRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Try meta language tag
	metaLangRegex := regexp.MustCompile(`<meta[^>]*http-equiv="content-language"[^>]*content="([^"]*)"[^>]*>`)
	matches = metaLangRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return "en" // Default to English
}

// extractFaviconURL determines the favicon URL
func (f *ContentFetcher) extractFaviconURL(html string, parsedURL *url.URL) string {
	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Try link rel="icon" tags
	iconRegex := regexp.MustCompile(`<link[^>]*rel="(?:icon|shortcut icon)"[^>]*href="([^"]*)"[^>]*>`)
	matches := iconRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		iconURL := matches[1]
		if strings.HasPrefix(iconURL, "http") {
			return iconURL
		} else if strings.HasPrefix(iconURL, "//") {
			return parsedURL.Scheme + ":" + iconURL
		} else if strings.HasPrefix(iconURL, "/") {
			return baseURL + iconURL
		} else {
			return baseURL + "/" + iconURL
		}
	}

	// Try apple-touch-icon
	appleIconRegex := regexp.MustCompile(`<link[^>]*rel="apple-touch-icon"[^>]*href="([^"]*)"[^>]*>`)
	matches = appleIconRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		iconURL := matches[1]
		if strings.HasPrefix(iconURL, "http") {
			return iconURL
		} else if strings.HasPrefix(iconURL, "//") {
			return parsedURL.Scheme + ":" + iconURL
		} else if strings.HasPrefix(iconURL, "/") {
			return baseURL + iconURL
		} else {
			return baseURL + "/" + iconURL
		}
	}

	// Default to /favicon.ico
	return baseURL + "/favicon.ico"
}

// decodeHTMLEntities performs basic HTML entity decoding
func (f *ContentFetcher) decodeHTMLEntities(s string) string {
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&apos;", "'")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return s
}

// FetchContentWithTimeout fetches content with a custom timeout
func (f *ContentFetcher) FetchContentWithTimeout(targetURL string, timeout time.Duration) (*PageContent, error) {
	originalTimeout := f.client.Timeout
	f.client.Timeout = timeout
	defer func() {
		f.client.Timeout = originalTimeout
	}()

	return f.FetchContent(targetURL)
}

// IsValidURL checks if a URL is valid and fetchable
func (f *ContentFetcher) IsValidURL(targetURL string) bool {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return false
	}

	// Must have scheme and host
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}

	// Only allow http and https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return false
	}

	return true
}