package api

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/bambithedeer/spotify-api/internal/auth"
	"github.com/bambithedeer/spotify-api/internal/client"
	"github.com/bambithedeer/spotify-api/internal/models"
)

// Mock response data
var mockTrackResponse = `{
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
	"available_markets": ["US", "CA"],
	"disc_number": 1,
	"track_number": 1,
	"is_local": false,
	"preview_url": null,
	"external_ids": {}
}`

var mockPaginatedResponse = `{
	"href": "https://api.spotify.com/v1/tracks",
	"items": [` + mockTrackResponse + `],
	"limit": 20,
	"next": "https://api.spotify.com/v1/tracks?offset=20&limit=20",
	"offset": 0,
	"previous": null,
	"total": 100
}`

var mockErrorResponse = `{
	"error": {
		"status": 400,
		"message": "Invalid request"
	}
}`

func TestResponseHandler_ParseResponse(t *testing.T) {
	handler := NewResponseHandler()

	// Test successful response
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(mockTrackResponse)),
	}

	var track models.Track
	err := handler.ParseResponse(resp, &track)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if track.ID != "track123" {
		t.Errorf("Expected track ID 'track123', got %s", track.ID)
	}

	if track.Name != "Test Track" {
		t.Errorf("Expected track name 'Test Track', got %s", track.Name)
	}
}

func TestResponseHandler_ParseErrorResponse(t *testing.T) {
	handler := NewResponseHandler()

	// Test error response
	resp := &http.Response{
		StatusCode: 400,
		Body:       io.NopCloser(strings.NewReader(mockErrorResponse)),
	}

	var track models.Track
	err := handler.ParseResponse(resp, &track)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedMsg := "HTTP 400: Invalid request"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain '%s', got %s", expectedMsg, err.Error())
	}
}

func TestResponseHandler_ParsePaginatedResponse(t *testing.T) {
	handler := NewResponseHandler()

	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(mockPaginatedResponse)),
	}

	var paging models.Paging[models.Track]
	pagination, err := handler.ParsePaginatedResponse(resp, &paging)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if pagination.Total != 100 {
		t.Errorf("Expected total 100, got %d", pagination.Total)
	}

	if pagination.Limit != 20 {
		t.Errorf("Expected limit 20, got %d", pagination.Limit)
	}

	if !pagination.HasNext() {
		t.Error("Expected to have next page")
	}

	if pagination.HasPrevious() {
		t.Error("Expected to not have previous page")
	}

	nextOffset := pagination.GetNextOffset()
	if nextOffset != 20 {
		t.Errorf("Expected next offset 20, got %d", nextOffset)
	}
}

func TestQueryParams_ToURLValues(t *testing.T) {
	params := QueryParams{
		"string_param": "test",
		"int_param":    42,
		"bool_param":   true,
		"array_param":  []string{"a", "b", "c"},
		"int_array":    []int{1, 2, 3},
		"empty_string": "",
		"empty_array":  []string{},
	}

	values := params.ToURLValues()

	if values.Get("string_param") != "test" {
		t.Errorf("Expected string_param 'test', got %s", values.Get("string_param"))
	}

	if values.Get("int_param") != "42" {
		t.Errorf("Expected int_param '42', got %s", values.Get("int_param"))
	}

	if values.Get("bool_param") != "true" {
		t.Errorf("Expected bool_param 'true', got %s", values.Get("bool_param"))
	}

	if values.Get("array_param") != "a,b,c" {
		t.Errorf("Expected array_param 'a,b,c', got %s", values.Get("array_param"))
	}

	if values.Get("int_array") != "1,2,3" {
		t.Errorf("Expected int_array '1,2,3', got %s", values.Get("int_array"))
	}

	// Empty values should not be included
	if values.Has("empty_string") {
		t.Error("Expected empty_string to not be included")
	}

	if values.Has("empty_array") {
		t.Error("Expected empty_array to not be included")
	}
}

func TestPaginationOptions_Merge(t *testing.T) {
	options := &PaginationOptions{
		Limit:  50,
		Offset: 100,
	}

	params := QueryParams{
		"existing": "value",
	}

	merged := options.Merge(params)

	if merged["limit"] != 50 {
		t.Errorf("Expected limit 50, got %v", merged["limit"])
	}

	if merged["offset"] != 100 {
		t.Errorf("Expected offset 100, got %v", merged["offset"])
	}

	if merged["existing"] != "value" {
		t.Errorf("Expected existing param to be preserved")
	}
}

func TestPaginationOptions_ValidateLimit(t *testing.T) {
	options := &PaginationOptions{Limit: 50}

	// Valid limit
	err := options.ValidateLimit(1, 100)
	if err != nil {
		t.Errorf("Expected no error for valid limit, got %v", err)
	}

	// Invalid limit (too low)
	options.Limit = 0
	err = options.ValidateLimit(1, 100)
	if err == nil {
		t.Error("Expected error for limit too low")
	}

	// Invalid limit (too high)
	options.Limit = 150
	err = options.ValidateLimit(1, 100)
	if err == nil {
		t.Error("Expected error for limit too high")
	}
}

func TestRequestBuilder_Integration(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tracks/track123":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockTrackResponse))
		case "/search":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockPaginatedResponse))
		case "/playlists":
			if r.Method == "POST" {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id": "playlist123", "name": "New Playlist"}`))
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client and builder
	client := client.NewClient("test", "test", "http://localhost/callback")
	client.SetBaseURL(server.URL)

	// Set a mock token for authentication
	mockToken := &auth.Token{
		AccessToken: "mock_token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	client.SetToken(mockToken)

	builder := NewRequestBuilder(client)

	ctx := context.Background()

	// Test GET request
	var track models.Track
	err := builder.Get(ctx, "/tracks/track123", nil, &track)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}

	if track.ID != "track123" {
		t.Errorf("Expected track ID 'track123', got %s", track.ID)
	}

	// Test GET with pagination
	var paging models.Paging[models.Track]
	pagination, err := builder.GetPaginated(ctx, "/search", QueryParams{"q": "test"}, &paging)
	if err != nil {
		t.Fatalf("Paginated GET request failed: %v", err)
	}

	if pagination.Total != 100 {
		t.Errorf("Expected total 100, got %d", pagination.Total)
	}

	// Test POST request
	requestBody := map[string]interface{}{
		"name":   "New Playlist",
		"public": true,
	}

	var playlist map[string]interface{}
	err = builder.Post(ctx, "/playlists", requestBody, &playlist)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}

	if playlist["id"] != "playlist123" {
		t.Errorf("Expected playlist ID 'playlist123', got %v", playlist["id"])
	}
}

func TestBatch_Operations(t *testing.T) {
	batch := NewBatch()

	// Add operations
	batch.AddGet("/tracks/1", QueryParams{"market": "US"}).
		AddPost("/playlists", map[string]string{"name": "Test"}).
		AddPut("/playlists/1", map[string]string{"name": "Updated"}).
		AddDelete("/playlists/1/tracks", QueryParams{"uris": "spotify:track:1"})

	if len(batch.operations) != 4 {
		t.Errorf("Expected 4 operations, got %d", len(batch.operations))
	}

	// Verify operation types
	expectedMethods := []string{"GET", "POST", "PUT", "DELETE"}
	for i, op := range batch.operations {
		if op.Method != expectedMethods[i] {
			t.Errorf("Expected method %s at position %d, got %s", expectedMethods[i], i, op.Method)
		}
	}
}

func TestRequestBuilder_BuildURL(t *testing.T) {
	client := client.NewClient("test", "test", "http://localhost/callback")
	builder := NewRequestBuilder(client)

	// Test without parameters
	url := builder.buildURL("/tracks", nil)
	if url != "/tracks" {
		t.Errorf("Expected '/tracks', got %s", url)
	}

	// Test with parameters
	params := QueryParams{
		"market": "US",
		"limit":  20,
	}
	url = builder.buildURL("/tracks", params)
	expected := "/tracks?limit=20&market=US"
	if url != expected {
		t.Errorf("Expected '%s', got %s", expected, url)
	}

	// Test with existing query parameters
	url = builder.buildURL("/tracks?existing=param", params)
	if !strings.Contains(url, "existing=param") {
		t.Error("Expected existing parameter to be preserved")
	}
	if !strings.Contains(url, "market=US") {
		t.Error("Expected new parameters to be added")
	}
}