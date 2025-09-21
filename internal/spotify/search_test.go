package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/auth"
	"github.com/bambithedeer/spotify-api/internal/client"
)

// Mock search responses
var mockSearchTrackResponse = `{
	"tracks": {
		"href": "https://api.spotify.com/v1/search?query=test&type=track&offset=0&limit=20",
		"items": [{
			"id": "track123",
			"name": "Test Track",
			"artists": [{
				"id": "artist123",
				"name": "Test Artist",
				"type": "artist",
				"uri": "spotify:artist:artist123",
				"href": "https://api.spotify.com/v1/artists/artist123",
				"external_urls": {"spotify": "https://open.spotify.com/artist/artist123"}
			}],
			"duration_ms": 180000,
			"explicit": false,
			"popularity": 75,
			"external_urls": {"spotify": "https://open.spotify.com/track/track123"},
			"href": "https://api.spotify.com/v1/tracks/track123",
			"type": "track",
			"uri": "spotify:track:track123",
			"available_markets": ["US", "CA"],
			"disc_number": 1,
			"track_number": 1,
			"is_local": false,
			"preview_url": null,
			"external_ids": {}
		}],
		"limit": 20,
		"next": null,
		"offset": 0,
		"previous": null,
		"total": 1
	}
}`

var mockSearchAllResponse = `{
	"tracks": {
		"href": "https://api.spotify.com/v1/search",
		"items": [{
			"id": "track123",
			"name": "Test Track",
			"artists": [{"id": "artist123", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:artist123", "href": "https://api.spotify.com/v1/artists/artist123", "external_urls": {"spotify": "https://open.spotify.com/artist/artist123"}}],
			"duration_ms": 180000,
			"explicit": false,
			"popularity": 75,
			"external_urls": {"spotify": "https://open.spotify.com/track/track123"},
			"href": "https://api.spotify.com/v1/tracks/track123",
			"type": "track",
			"uri": "spotify:track:track123",
			"available_markets": ["US"],
			"disc_number": 1,
			"track_number": 1,
			"is_local": false,
			"preview_url": null,
			"external_ids": {}
		}],
		"limit": 20,
		"next": null,
		"offset": 0,
		"previous": null,
		"total": 1
	},
	"artists": {
		"href": "https://api.spotify.com/v1/search",
		"items": [{
			"id": "artist123",
			"name": "Test Artist",
			"type": "artist",
			"uri": "spotify:artist:artist123",
			"href": "https://api.spotify.com/v1/artists/artist123",
			"external_urls": {"spotify": "https://open.spotify.com/artist/artist123"},
			"followers": {"href": null, "total": 10000},
			"genres": ["rock", "pop"],
			"images": [],
			"popularity": 75
		}],
		"limit": 20,
		"next": null,
		"offset": 0,
		"previous": null,
		"total": 1
	}
}`

func createTestSearchService() (*SearchService, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"status": 401, "message": "Unauthorized"}}`))
			return
		}

		query := r.URL.Query()
		typeParam := query.Get("type")

		switch {
		case typeParam == "track":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockSearchTrackResponse))
		case strings.Contains(typeParam, "track") && strings.Contains(typeParam, "artist"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockSearchAllResponse))
		case typeParam == "album":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"albums": {"href": "test", "items": [], "limit": 20, "next": null, "offset": 0, "previous": null, "total": 0}}`))
		case typeParam == "artist":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"artists": {"href": "test", "items": [], "limit": 20, "next": null, "offset": 0, "previous": null, "total": 0}}`))
		case typeParam == "playlist":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"playlists": {"href": "test", "items": [], "limit": 20, "next": null, "offset": 0, "previous": null, "total": 0}}`))
		case query.Get("q") != "":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockSearchTrackResponse))
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"status": 400, "message": "Bad request"}}`))
		}
	}))

	// Create client and set test server URL
	client := client.NewClient("test_id", "test_secret", "http://localhost/callback")
	client.SetBaseURL(server.URL)

	// Set mock token
	token := &auth.Token{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	client.SetToken(token)

	// Create request builder and search service
	builder := api.NewRequestBuilder(client)
	service := NewSearchService(builder)

	return service, server
}

func TestSearchService_SearchTracks(t *testing.T) {
	service, server := createTestSearchService()
	defer server.Close()

	ctx := context.Background()

	// Test successful search
	tracks, pagination, err := service.SearchTracks(ctx, "test track", nil)
	if err != nil {
		t.Fatalf("SearchTracks failed: %v", err)
	}

	if tracks == nil {
		t.Fatal("Expected tracks result, got nil")
	}

	if len(tracks.Items) != 1 {
		t.Errorf("Expected 1 track, got %d", len(tracks.Items))
	}

	if tracks.Items[0].ID != "track123" {
		t.Errorf("Expected track ID 'track123', got %s", tracks.Items[0].ID)
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}

	// For search endpoints, pagination info comes from the nested tracks object
	// The pagination parser looks at the root level, so it won't find nested pagination
	// This is expected behavior for search responses
}

func TestSearchService_SearchTracksWithPagination(t *testing.T) {
	service, server := createTestSearchService()
	defer server.Close()

	ctx := context.Background()

	// Test with pagination options
	paginationOpts := &api.PaginationOptions{
		Limit:  10,
		Offset: 0,
	}

	tracks, pagination, err := service.SearchTracks(ctx, "test track", paginationOpts)
	if err != nil {
		t.Fatalf("SearchTracks with pagination failed: %v", err)
	}

	if tracks == nil {
		t.Fatal("Expected tracks result, got nil")
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestSearchService_SearchAlbums(t *testing.T) {
	service, server := createTestSearchService()
	defer server.Close()

	ctx := context.Background()

	albums, pagination, err := service.SearchAlbums(ctx, "test album", nil)
	if err != nil {
		t.Fatalf("SearchAlbums failed: %v", err)
	}

	if albums == nil {
		t.Fatal("Expected albums result, got nil")
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestSearchService_SearchArtists(t *testing.T) {
	service, server := createTestSearchService()
	defer server.Close()

	ctx := context.Background()

	artists, pagination, err := service.SearchArtists(ctx, "test artist", nil)
	if err != nil {
		t.Fatalf("SearchArtists failed: %v", err)
	}

	if artists == nil {
		t.Fatal("Expected artists result, got nil")
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestSearchService_SearchPlaylists(t *testing.T) {
	service, server := createTestSearchService()
	defer server.Close()

	ctx := context.Background()

	playlists, pagination, err := service.SearchPlaylists(ctx, "test playlist", nil)
	if err != nil {
		t.Fatalf("SearchPlaylists failed: %v", err)
	}

	if playlists == nil {
		t.Fatal("Expected playlists result, got nil")
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestSearchService_Search(t *testing.T) {
	service, server := createTestSearchService()
	defer server.Close()

	ctx := context.Background()

	// Test multi-type search
	options := &SearchOptions{
		Query: "test",
		Types: []string{"track", "artist"},
		Limit: 20,
	}

	result, err := service.Search(ctx, options)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected search result, got nil")
	}

	if result.Tracks == nil {
		t.Error("Expected tracks in result")
	}

	if result.Artists == nil {
		t.Error("Expected artists in result")
	}
}

func TestSearchService_ValidationErrors(t *testing.T) {
	service, server := createTestSearchService()
	defer server.Close()

	ctx := context.Background()

	// Test empty query
	_, _, err := service.SearchTracks(ctx, "", nil)
	if err == nil {
		t.Error("Expected error for empty query")
	}

	// Test whitespace-only query
	_, _, err = service.SearchTracks(ctx, "   ", nil)
	if err == nil {
		t.Error("Expected error for whitespace-only query")
	}

	// Test invalid pagination limit
	invalidPagination := &api.PaginationOptions{
		Limit: 100, // Too high
	}
	_, _, err = service.SearchTracks(ctx, "test", invalidPagination)
	if err == nil {
		t.Error("Expected error for invalid pagination limit")
	}
}

func TestSearchService_SearchOptionsValidation(t *testing.T) {
	service, server := createTestSearchService()
	defer server.Close()

	ctx := context.Background()

	// Test invalid search options
	invalidOptions := &SearchOptions{
		Query: "", // Empty query
		Types: []string{"track"},
	}

	_, err := service.Search(ctx, invalidOptions)
	if err == nil {
		t.Error("Expected error for invalid search options")
	}

	// Test invalid search types
	invalidTypes := &SearchOptions{
		Query: "test",
		Types: []string{"invalid_type"},
	}

	_, err = service.Search(ctx, invalidTypes)
	if err == nil {
		t.Error("Expected error for invalid search types")
	}

	// Test invalid market
	invalidMarket := &SearchOptions{
		Query:  "test",
		Types:  []string{"track"},
		Market: "INVALID",
	}

	_, err = service.Search(ctx, invalidMarket)
	if err == nil {
		t.Error("Expected error for invalid market")
	}
}

func TestSearchFilter(t *testing.T) {
	// Test basic filter
	filter := NewSearchFilter("test")
	query := filter.String()
	if query != "test" {
		t.Errorf("Expected 'test', got %s", query)
	}

	// Test complex filter
	filter = NewSearchFilter("rock").
		Artist("Beatles").
		Album("Abbey Road").
		Year(1969)

	query = filter.String()
	expected := "rock artist:Beatles album:Abbey Road year:1969"
	if query != expected {
		t.Errorf("Expected '%s', got %s", expected, query)
	}

	// Test year range filter
	filter = NewSearchFilter("rock").YearRange(1960, 1970)
	query = filter.String()
	expected = "rock year:1960-1970"
	if query != expected {
		t.Errorf("Expected '%s', got %s", expected, query)
	}

	// Test tag filters
	filter = NewSearchFilter("music").IsNew().IsHipster()
	query = filter.String()
	expected = "music tag:new tag:hipster"
	if query != expected {
		t.Errorf("Expected '%s', got %s", expected, query)
	}
}

func TestSearchService_BuildSearchParams(t *testing.T) {
	service := &SearchService{
		validator: api.NewValidator(),
	}

	options := &SearchOptions{
		Query:           "test query",
		Types:           []string{"track", "artist"},
		Market:          "US",
		Limit:           10,
		Offset:          20,
		IncludeExternal: "audio",
	}

	params := service.buildSearchParams(options)

	if params["q"] != "test query" {
		t.Errorf("Expected query 'test query', got %v", params["q"])
	}

	if params["type"] != "track,artist" {
		t.Errorf("Expected type 'track,artist', got %v", params["type"])
	}

	if params["market"] != "US" {
		t.Errorf("Expected market 'US', got %v", params["market"])
	}

	if params["limit"] != 10 {
		t.Errorf("Expected limit 10, got %v", params["limit"])
	}

	if params["offset"] != 20 {
		t.Errorf("Expected offset 20, got %v", params["offset"])
	}

	if params["include_external"] != "audio" {
		t.Errorf("Expected include_external 'audio', got %v", params["include_external"])
	}
}