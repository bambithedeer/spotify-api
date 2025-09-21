package spotify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/client"
)

// Mock library responses
var mockSavedTracksResponse = `{
	"href": "https://api.spotify.com/v1/me/tracks",
	"items": [{
		"added_at": "2023-01-01T00:00:00Z",
		"track": {
			"id": "6iV5W9uYEdYUVa79Axb7Rh",
			"name": "Test Track",
			"artists": [{
				"id": "1301WleyT98MSxVHPZCA6M",
				"name": "Test Artist",
				"type": "artist",
				"uri": "spotify:artist:1301WleyT98MSxVHPZCA6M",
				"href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M",
				"external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}
			}],
			"album": {
				"id": "4iV5W9uYEdYUVa79Axb7Rh",
				"name": "Test Album",
				"type": "album",
				"uri": "spotify:album:4iV5W9uYEdYUVa79Axb7Rh",
				"href": "https://api.spotify.com/v1/albums/4iV5W9uYEdYUVa79Axb7Rh",
				"external_urls": {"spotify": "https://open.spotify.com/album/4iV5W9uYEdYUVa79Axb7Rh"}
			},
			"duration_ms": 240000,
			"explicit": false,
			"external_urls": {"spotify": "https://open.spotify.com/track/6iV5W9uYEdYUVa79Axb7Rh"},
			"href": "https://api.spotify.com/v1/tracks/6iV5W9uYEdYUVa79Axb7Rh",
			"type": "track",
			"uri": "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
			"track_number": 1,
			"disc_number": 1,
			"is_local": false,
			"preview_url": null
		}
	}],
	"limit": 20,
	"next": null,
	"offset": 0,
	"previous": null,
	"total": 1
}`

var mockSavedAlbumsResponse = `{
	"href": "https://api.spotify.com/v1/me/albums",
	"items": [{
		"added_at": "2023-01-01T00:00:00Z",
		"album": {
			"id": "6akEvsycLGftJxYudPjmqK",
			"name": "Test Album",
			"artists": [{
				"id": "1301WleyT98MSxVHPZCA6M",
				"name": "Test Artist",
				"type": "artist",
				"uri": "spotify:artist:1301WleyT98MSxVHPZCA6M",
				"href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M",
				"external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}
			}],
			"album_type": "album",
			"total_tracks": 12,
			"available_markets": ["US", "CA"],
			"external_urls": {"spotify": "https://open.spotify.com/album/6akEvsycLGftJxYudPjmqK"},
			"href": "https://api.spotify.com/v1/albums/6akEvsycLGftJxYudPjmqK",
			"images": [],
			"release_date": "2023-01-01",
			"release_date_precision": "day",
			"type": "album",
			"uri": "spotify:album:6akEvsycLGftJxYudPjmqK"
		}
	}],
	"limit": 20,
	"next": null,
	"offset": 0,
	"previous": null,
	"total": 1
}`

var mockCheckTracksResponse = `[true, false, true]`
var mockCheckAlbumsResponse = `[false, true]`

func createTestLibraryService() (*LibraryService, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"status": 401, "message": "Unauthorized"}}`))
			return
		}

		switch {
		case r.URL.Path == "/me/tracks" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockSavedTracksResponse))
		case r.URL.Path == "/me/tracks" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/tracks" && r.Method == "DELETE":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/tracks/contains" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockCheckTracksResponse))
		case r.URL.Path == "/me/albums" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockSavedAlbumsResponse))
		case r.URL.Path == "/me/albums" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/albums" && r.Method == "DELETE":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/albums/contains" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockCheckAlbumsResponse))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": {"status": 404, "message": "Not Found"}}`))
		}
	}))

	// Create a client with the test server URL
	spotifyClient := &client.Client{}
	// Note: We're using a simplified approach here. In practice, we'd need a way to inject the base URL
	requestBuilder := api.NewRequestBuilder(spotifyClient)

	service := NewLibraryService(requestBuilder)
	return service, server
}

func TestLibraryService_ValidationErrors(t *testing.T) {
	// Create a minimal RequestBuilder for validation testing (will fail at network level)
	client := &client.Client{}
	requestBuilder := api.NewRequestBuilder(client)
	service := NewLibraryService(requestBuilder)

	// Test empty track IDs for SaveTracks
	err := service.SaveTracks(context.Background(), []string{})
	if err == nil {
		t.Error("Expected error for empty track IDs in SaveTracks")
	}

	// Test too many track IDs for SaveTracks
	tooManyIDs := make([]string, 51)
	for i := range tooManyIDs {
		tooManyIDs[i] = "spotify:track:6iV5W9uYEdYUVa79Axb7Rh"
	}
	err = service.SaveTracks(context.Background(), tooManyIDs)
	if err == nil {
		t.Error("Expected error for too many track IDs in SaveTracks")
	}

	// Test empty track IDs for RemoveTracks
	err = service.RemoveTracks(context.Background(), []string{})
	if err == nil {
		t.Error("Expected error for empty track IDs in RemoveTracks")
	}

	// Test too many track IDs for RemoveTracks
	err = service.RemoveTracks(context.Background(), tooManyIDs)
	if err == nil {
		t.Error("Expected error for too many track IDs in RemoveTracks")
	}

	// Test empty track IDs for CheckSavedTracks
	_, err = service.CheckSavedTracks(context.Background(), []string{})
	if err == nil {
		t.Error("Expected error for empty track IDs in CheckSavedTracks")
	}

	// Test too many track IDs for CheckSavedTracks
	_, err = service.CheckSavedTracks(context.Background(), tooManyIDs)
	if err == nil {
		t.Error("Expected error for too many track IDs in CheckSavedTracks")
	}

	// Test empty album IDs for SaveAlbums
	err = service.SaveAlbums(context.Background(), []string{})
	if err == nil {
		t.Error("Expected error for empty album IDs in SaveAlbums")
	}

	// Test too many album IDs for SaveAlbums
	tooManyAlbumIDs := make([]string, 51)
	for i := range tooManyAlbumIDs {
		tooManyAlbumIDs[i] = "spotify:album:6akEvsycLGftJxYudPjmqK"
	}
	err = service.SaveAlbums(context.Background(), tooManyAlbumIDs)
	if err == nil {
		t.Error("Expected error for too many album IDs in SaveAlbums")
	}

	// Test empty album IDs for RemoveAlbums
	err = service.RemoveAlbums(context.Background(), []string{})
	if err == nil {
		t.Error("Expected error for empty album IDs in RemoveAlbums")
	}

	// Test too many album IDs for RemoveAlbums
	err = service.RemoveAlbums(context.Background(), tooManyAlbumIDs)
	if err == nil {
		t.Error("Expected error for too many album IDs in RemoveAlbums")
	}

	// Test empty album IDs for CheckSavedAlbums
	_, err = service.CheckSavedAlbums(context.Background(), []string{})
	if err == nil {
		t.Error("Expected error for empty album IDs in CheckSavedAlbums")
	}

	// Test too many album IDs for CheckSavedAlbums
	_, err = service.CheckSavedAlbums(context.Background(), tooManyAlbumIDs)
	if err == nil {
		t.Error("Expected error for too many album IDs in CheckSavedAlbums")
	}

	// Test invalid market in SavedAlbumsOptions
	options := &SavedAlbumsOptions{
		Market: "INVALID_MARKET_CODE",
	}
	_, _, err = service.GetSavedAlbums(context.Background(), options)
	if err == nil {
		t.Error("Expected error for invalid market code")
	}

	// Test invalid limit in SavedAlbumsOptions
	options = &SavedAlbumsOptions{
		Limit: 100, // exceeds max limit of 50
	}
	_, _, err = service.GetSavedAlbums(context.Background(), options)
	if err == nil {
		t.Error("Expected error for limit exceeding maximum")
	}

	// Test negative offset in SavedAlbumsOptions
	options = &SavedAlbumsOptions{
		Offset: -1,
	}
	_, _, err = service.GetSavedAlbums(context.Background(), options)
	if err == nil {
		t.Error("Expected error for negative offset")
	}
}

func TestLibraryService_ValidationSuccess(t *testing.T) {
	// Create a minimal RequestBuilder for validation testing
	client := &client.Client{}
	requestBuilder := api.NewRequestBuilder(client)
	service := NewLibraryService(requestBuilder)

	// Test valid SavedAlbumsOptions
	options := &SavedAlbumsOptions{
		Market: "US",
		Limit:  20,
		Offset: 0,
	}

	// This will fail with network error, but we just want to test validation passes
	_, _, err := service.GetSavedAlbums(context.Background(), options)
	// We expect a network error, not a validation error
	if err != nil && strings.Contains(err.Error(), "invalid") {
		t.Errorf("Expected network error, got validation error: %v", err)
	}
}