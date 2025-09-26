package lidarr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a Lidarr API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Config holds Lidarr client configuration
type Config struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// Artist represents a Lidarr artist
type Artist struct {
	ID                int    `json:"id"`
	Status            string `json:"status"`
	Ended             bool   `json:"ended"`
	ArtistName        string `json:"artistName"`
	ForeignArtistID   string `json:"foreignArtistId"`
	MBid              string `json:"mbId"`
	AMGid             int    `json:"amgId"`
	DiscogsID         int    `json:"discogsId"`
	AllMusicID        string `json:"allMusicId"`
	Overview          string `json:"overview"`
	ArtistType        string `json:"artistType"`
	Disambiguation    string `json:"disambiguation"`
	Links             []Link `json:"links"`
	Images            []Image `json:"images"`
	Path              string `json:"path"`
	QualityProfileID  int    `json:"qualityProfileId"`
	MetadataProfileID int    `json:"metadataProfileId"`
	Monitored         bool   `json:"monitored"`
	MonitorNewItems   string `json:"monitorNewItems"`
	RootFolderPath    string `json:"rootFolderPath"`
	Genres            []string `json:"genres"`
	CleanName         string `json:"cleanName"`
	SortName          string `json:"sortName"`
	Tags              []int  `json:"tags"`
	Added             string `json:"added"`
	AddOptions        *AddOptions `json:"addOptions,omitempty"`
}

// AddOptions represents options for adding an artist
type AddOptions struct {
	Monitor                string `json:"monitor"`
	SearchForMissingAlbums bool   `json:"searchForMissingAlbums"`
}

// Link represents an external link
type Link struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// Image represents an artist image
type Image struct {
	CoverType string `json:"coverType"`
	URL       string `json:"url"`
}

// Command represents a Lidarr command
type Command struct {
	Name     string                 `json:"name"`
	Body     map[string]interface{} `json:"body,omitempty"`
	Priority string                 `json:"priority"`
	Status   string                 `json:"status"`
	Queued   string                 `json:"queued"`
	Started  string                 `json:"started"`
	Ended    string                 `json:"ended"`
	Duration string                 `json:"duration"`
	ID       int                    `json:"id"`
}

// NewClient creates a new Lidarr API client
func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &Client{
		baseURL: strings.TrimSuffix(config.BaseURL, "/"),
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// makeRequest makes an HTTP request to the Lidarr API
func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := fmt.Sprintf("%s/api/v1%s", c.baseURL, endpoint)
	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

// SearchArtist searches for an artist by MusicBrainz ID
func (c *Client) SearchArtist(mbid string) ([]Artist, error) {
	searchTerm := fmt.Sprintf("lidarr:%s", mbid)
	endpoint := fmt.Sprintf("/artist/lookup?term=%s", url.QueryEscape(searchTerm))

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var artists []Artist
	if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return artists, nil
}

// AddArtist adds an artist to Lidarr
func (c *Client) AddArtist(artist Artist) (*Artist, error) {
	resp, err := c.makeRequest("POST", "/artist", artist)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to add artist, status: %d", resp.StatusCode)
	}

	var addedArtist Artist
	if err := json.NewDecoder(resp.Body).Decode(&addedArtist); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &addedArtist, nil
}

// AddArtistByMBID adds an artist to Lidarr using MusicBrainz ID
func (c *Client) AddArtistByMBID(mbid, rootFolderPath string, qualityProfileID, metadataProfileID int, monitor bool, searchForMissing bool) (*Artist, error) {
	// First search for the artist
	artists, err := c.SearchArtist(mbid)
	if err != nil {
		return nil, fmt.Errorf("failed to search for artist: %w", err)
	}

	if len(artists) == 0 {
		return nil, fmt.Errorf("artist not found with MBID: %s", mbid)
	}

	// Take the first result and configure it
	artist := artists[0]
	artist.RootFolderPath = rootFolderPath
	artist.QualityProfileID = qualityProfileID
	artist.MetadataProfileID = metadataProfileID
	artist.Monitored = monitor
	artist.MonitorNewItems = "all"

	if searchForMissing {
		artist.AddOptions = &AddOptions{
			Monitor:                "all",
			SearchForMissingAlbums: true,
		}
	}

	// Add the artist
	return c.AddArtist(artist)
}

// GetRootFolders gets available root folders
func (c *Client) GetRootFolders() ([]RootFolder, error) {
	resp, err := c.makeRequest("GET", "/rootfolder", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var folders []RootFolder
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return folders, nil
}

// GetQualityProfiles gets available quality profiles
func (c *Client) GetQualityProfiles() ([]QualityProfile, error) {
	resp, err := c.makeRequest("GET", "/qualityprofile", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var profiles []QualityProfile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return profiles, nil
}

// GetMetadataProfiles gets available metadata profiles
func (c *Client) GetMetadataProfiles() ([]MetadataProfile, error) {
	resp, err := c.makeRequest("GET", "/metadataprofile", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var profiles []MetadataProfile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return profiles, nil
}

// TestConnection tests the connection to Lidarr
func (c *Client) TestConnection() error {
	resp, err := c.makeRequest("GET", "/system/status", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("connection test failed with status %d", resp.StatusCode)
	}

	return nil
}

// RootFolder represents a Lidarr root folder
type RootFolder struct {
	ID                int    `json:"id"`
	Path              string `json:"path"`
	Accessible        bool   `json:"accessible"`
	FreeSpace         int64  `json:"freeSpace"`
	UnmappedFolders   []UnmappedFolder `json:"unmappedFolders"`
}

// UnmappedFolder represents an unmapped folder
type UnmappedFolder struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// QualityProfile represents a Lidarr quality profile
type QualityProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MetadataProfile represents a Lidarr metadata profile
type MetadataProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}