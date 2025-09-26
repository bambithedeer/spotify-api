package musicbrainz

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	BaseURL   = "https://musicbrainz.org/ws/2"
	UserAgent = "spotify-cli/1.0 (https://github.com/bambithedeer/spotify-api)"
	RateLimit = 1 * time.Second // MusicBrainz rate limit: 1 request per second
)

// Client represents a MusicBrainz API client
type Client struct {
	httpClient  *http.Client
	rateLimiter *time.Ticker
	userAgent   string
}

// Artist represents a MusicBrainz artist
type Artist struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	SortName string   `json:"sort-name"`
	Type     string   `json:"type"`
	Country  string   `json:"country"`
	Aliases  []Alias  `json:"aliases"`
	Score    int      `json:"score"`
}

// Alias represents an artist alias
type Alias struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Primary bool  `json:"primary"`
}

// SearchResponse represents the response from a MusicBrainz search
type SearchResponse struct {
	Created string   `json:"created"`
	Count   int      `json:"count"`
	Offset  int      `json:"offset"`
	Artists []Artist `json:"artists"`
}

// NewClient creates a new MusicBrainz API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: time.NewTicker(RateLimit),
		userAgent:   UserAgent,
	}
}

// SearchArtist searches for artists by name
func (c *Client) SearchArtist(artistName string) (*SearchResponse, error) {
	// Wait for rate limiter
	<-c.rateLimiter.C

	// Build search query
	query := fmt.Sprintf("artist:%s", url.QueryEscape(strings.ToLower(artistName)))

	// Construct URL
	searchURL := fmt.Sprintf("%s/artist/?query=%s&fmt=json&limit=10", BaseURL, query)

	// Create request
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResp, nil
}

// GetBestMatch returns the best matching artist from search results
func (c *Client) GetBestMatch(artistName string) (*Artist, error) {
	searchResp, err := c.SearchArtist(artistName)
	if err != nil {
		return nil, err
	}

	if len(searchResp.Artists) == 0 {
		return nil, fmt.Errorf("no artists found for '%s'", artistName)
	}

	// Return the first result (highest score)
	bestMatch := &searchResp.Artists[0]
	return bestMatch, nil
}

// GetArtistMBID returns the MusicBrainz ID for an artist
func (c *Client) GetArtistMBID(artistName string) (string, error) {
	artist, err := c.GetBestMatch(artistName)
	if err != nil {
		return "", err
	}
	return artist.ID, nil
}

// Close cleans up the client resources
func (c *Client) Close() {
	if c.rateLimiter != nil {
		c.rateLimiter.Stop()
	}
}