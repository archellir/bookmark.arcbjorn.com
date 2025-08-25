package services

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// URLService provides URL normalization and expansion utilities
type URLService struct {
	client *http.Client
}

// NewURLService creates a new URL service
func NewURLService() *URLService {
	return &URLService{
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Stop after 10 redirects
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

// URLNormalizationResult contains normalized URL and variations
type URLNormalizationResult struct {
	Normalized  string   `json:"normalized"`
	Variations  []string `json:"variations"`
	IsShortURL  bool     `json:"is_short_url"`
	ExpandedURL string   `json:"expanded_url,omitempty"`
}

// NormalizeURL normalizes a URL and generates common variations
func (s *URLService) NormalizeURL(rawURL string) (*URLNormalizationResult, error) {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure scheme is present
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
		rawURL = parsedURL.String()
		parsedURL, _ = url.Parse(rawURL)
	}

	result := &URLNormalizationResult{
		IsShortURL: s.isShortURL(parsedURL.Host),
	}

	// If it's a short URL, try to expand it
	if result.IsShortURL {
		expanded, err := s.expandURL(rawURL)
		if err == nil && expanded != rawURL {
			result.ExpandedURL = expanded
			// Use expanded URL for normalization
			expandedParsed, err := url.Parse(expanded)
			if err == nil {
				parsedURL = expandedParsed
			}
		}
	}

	// Normalize the URL
	normalized := s.normalizeURL(parsedURL)
	result.Normalized = normalized

	// Generate variations
	result.Variations = s.generateURLVariations(parsedURL)

	return result, nil
}

// normalizeURL applies standard normalization rules
func (s *URLService) normalizeURL(u *url.URL) string {
	// Create a copy to avoid modifying original
	normalized := *u

	// Convert scheme to lowercase
	normalized.Scheme = strings.ToLower(normalized.Scheme)

	// Convert host to lowercase and remove www prefix
	normalized.Host = strings.ToLower(normalized.Host)
	if strings.HasPrefix(normalized.Host, "www.") {
		normalized.Host = normalized.Host[4:]
	}

	// Remove default ports
	if (normalized.Scheme == "http" && strings.HasSuffix(normalized.Host, ":80")) ||
		(normalized.Scheme == "https" && strings.HasSuffix(normalized.Host, ":443")) {
		normalized.Host = strings.Split(normalized.Host, ":")[0]
	}

	// Remove trailing slash from path (except for root)
	if normalized.Path != "/" && strings.HasSuffix(normalized.Path, "/") {
		normalized.Path = strings.TrimSuffix(normalized.Path, "/")
	}

	// Remove common tracking parameters
	values := normalized.Query()
	trackingParams := []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "gclid", "msclkid", "ref", "source",
		"_ga", "_gid", "mc_cid", "mc_eid",
	}

	for _, param := range trackingParams {
		values.Del(param)
	}
	normalized.RawQuery = values.Encode()

	// Remove fragment
	normalized.Fragment = ""

	return normalized.String()
}

// generateURLVariations creates common URL variations for duplicate detection
func (s *URLService) generateURLVariations(u *url.URL) []string {
	variations := make([]string, 0)
	
	// Original normalized
	normalized := s.normalizeURL(u)
	variations = append(variations, normalized)

	// With/without www
	withWWW := *u
	withoutWWW := *u
	
	if strings.HasPrefix(u.Host, "www.") {
		withoutWWW.Host = u.Host[4:]
	} else {
		withWWW.Host = "www." + u.Host
	}
	
	variations = append(variations, s.normalizeURL(&withWWW))
	variations = append(variations, s.normalizeURL(&withoutWWW))

	// HTTP/HTTPS variations
	httpVar := *u
	httpsVar := *u
	httpVar.Scheme = "http"
	httpsVar.Scheme = "https"
	
	variations = append(variations, s.normalizeURL(&httpVar))
	variations = append(variations, s.normalizeURL(&httpsVar))

	// With/without trailing slash
	withSlash := *u
	withoutSlash := *u
	
	if !strings.HasSuffix(u.Path, "/") && u.Path != "" {
		withSlash.Path = u.Path + "/"
	}
	if strings.HasSuffix(u.Path, "/") && u.Path != "/" {
		withoutSlash.Path = strings.TrimSuffix(u.Path, "/")
	}
	
	variations = append(variations, s.normalizeURL(&withSlash))
	variations = append(variations, s.normalizeURL(&withoutSlash))

	// Remove duplicates
	seen := make(map[string]bool)
	unique := make([]string, 0)
	
	for _, v := range variations {
		if !seen[v] {
			seen[v] = true
			unique = append(unique, v)
		}
	}

	return unique
}

// isShortURL checks if a host is a known URL shortener
func (s *URLService) isShortURL(host string) bool {
	shorteners := []string{
		"bit.ly", "tinyurl.com", "t.co", "goo.gl", "ow.ly",
		"short.link", "tiny.cc", "is.gd", "buff.ly", "ift.tt",
		"youtu.be", "amzn.to", "fb.me", "li.st", "tr.im",
		"cutt.ly", "rebrand.ly", "bl.ink", "switchy.io",
		"short.io", "tiny.one", "link.do", "clck.ru",
	}

	host = strings.ToLower(host)
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}

	for _, shortener := range shorteners {
		if host == shortener {
			return true
		}
	}

	// Check for custom short domains (usually short)
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		domain := parts[len(parts)-2]
		// If domain part is very short, it might be a custom shortener
		if len(domain) <= 3 && !isCommonTLD(host) {
			return true
		}
	}

	return false
}

// expandURL attempts to expand a short URL by following redirects
func (s *URLService) expandURL(shortURL string) (string, error) {
	// Create a HEAD request to avoid downloading content
	req, err := http.NewRequest("HEAD", shortURL, nil)
	if err != nil {
		return shortURL, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Torimemo/1.0; +https://torimemo.app)")

	resp, err := s.client.Do(req)
	if err != nil {
		// If HEAD fails, try GET
		resp, err = s.client.Get(shortURL)
		if err != nil {
			return shortURL, err
		}
	}
	defer resp.Body.Close()

	// The final URL after redirects
	finalURL := resp.Request.URL.String()
	
	// If no redirect occurred, return original
	if finalURL == shortURL {
		return shortURL, fmt.Errorf("no redirect found")
	}

	return finalURL, nil
}

// isCommonTLD checks if a domain has a common TLD (not a short URL)
func isCommonTLD(host string) bool {
	commonTLDs := []string{
		".com", ".org", ".net", ".edu", ".gov", ".mil",
		".int", ".co.uk", ".de", ".fr", ".jp", ".cn",
	}

	for _, tld := range commonTLDs {
		if strings.HasSuffix(host, tld) {
			return true
		}
	}
	return false
}

// FindSimilarURLs finds URLs that might be duplicates of the given URL
func (s *URLService) FindSimilarURLs(targetURL string, candidateURLs []string) []string {
	result, err := s.NormalizeURL(targetURL)
	if err != nil {
		return nil
	}

	var similar []string
	targetVariations := make(map[string]bool)
	
	// Add all variations of target URL to map for quick lookup
	for _, variation := range result.Variations {
		targetVariations[variation] = true
	}
	
	// Check expanded URL variations if it's a short URL
	if result.ExpandedURL != "" {
		expandedResult, err := s.NormalizeURL(result.ExpandedURL)
		if err == nil {
			for _, variation := range expandedResult.Variations {
				targetVariations[variation] = true
			}
		}
	}

	// Check each candidate URL
	for _, candidateURL := range candidateURLs {
		candidateResult, err := s.NormalizeURL(candidateURL)
		if err != nil {
			continue
		}

		// Check if any variation matches
		for _, variation := range candidateResult.Variations {
			if targetVariations[variation] {
				similar = append(similar, candidateURL)
				break
			}
		}

		// Also check expanded URL if candidate is short URL
		if candidateResult.ExpandedURL != "" {
			expandedResult, err := s.NormalizeURL(candidateResult.ExpandedURL)
			if err == nil {
				for _, variation := range expandedResult.Variations {
					if targetVariations[variation] {
						similar = append(similar, candidateURL)
						break
					}
				}
			}
		}
	}

	return similar
}

// GenerateShortURL creates a custom short URL (simple implementation)
func (s *URLService) GenerateShortURL(originalURL string, baseURL string) (string, error) {
	// Simple hash-based short code generation
	hash := s.generateShortCode(originalURL)
	return fmt.Sprintf("%s/s/%s", strings.TrimSuffix(baseURL, "/"), hash), nil
}

// generateShortCode generates a short code from URL
func (s *URLService) generateShortCode(url string) string {
	// Simple base62 encoding of hash
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	
	// Simple hash function (for demo - in production use crypto/rand)
	hash := 0
	for _, char := range url {
		hash = (hash*31 + int(char)) % 238328 // Keep it reasonable size
	}
	
	// Convert to base62
	result := ""
	for hash > 0 {
		result = string(chars[hash%62]) + result
		hash /= 62
	}
	
	// Ensure minimum length
	for len(result) < 6 {
		result = "0" + result
	}
	
	return result
}