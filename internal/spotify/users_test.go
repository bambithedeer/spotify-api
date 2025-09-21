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

// Mock user responses
var mockCurrentUserResponse = `{
	"id": "testuser",
	"display_name": "Test User",
	"email": "test@example.com",
	"external_urls": {"spotify": "https://open.spotify.com/user/testuser"},
	"followers": {"href": null, "total": 42},
	"href": "https://api.spotify.com/v1/users/testuser",
	"images": [{
		"url": "https://example.com/avatar.jpg",
		"height": 300,
		"width": 300
	}],
	"type": "user",
	"uri": "spotify:user:testuser",
	"country": "US",
	"product": "premium"
}`

var mockUserResponse = `{
	"id": "otheruser",
	"display_name": "Other User",
	"external_urls": {"spotify": "https://open.spotify.com/user/otheruser"},
	"followers": {"href": null, "total": 123},
	"href": "https://api.spotify.com/v1/users/otheruser",
	"images": [],
	"type": "user",
	"uri": "spotify:user:otheruser"
}`

var mockFollowedArtistsResponse = `{
	"artists": {
		"href": "https://api.spotify.com/v1/me/following?type=artist",
		"items": [{
			"id": "1301WleyT98MSxVHPZCA6M",
			"name": "Test Artist",
			"genres": ["rock", "alternative"],
			"popularity": 75,
			"type": "artist",
			"uri": "spotify:artist:1301WleyT98MSxVHPZCA6M",
			"href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M",
			"external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"},
			"images": [{
				"url": "https://example.com/artist.jpg",
				"height": 640,
				"width": 640
			}],
			"followers": {"href": null, "total": 1000000}
		}],
		"limit": 20,
		"next": null,
		"cursors": {
			"after": "1301WleyT98MSxVHPZCA6M"
		},
		"total": 1
	}
}`

var mockTopArtistsResponse = `{
	"href": "https://api.spotify.com/v1/me/top/artists",
	"items": [{
		"id": "2301WleyT98MSxVHPZCA6M",
		"name": "Top Artist",
		"genres": ["pop", "electronic"],
		"popularity": 85,
		"type": "artist",
		"uri": "spotify:artist:2301WleyT98MSxVHPZCA6M",
		"href": "https://api.spotify.com/v1/artists/2301WleyT98MSxVHPZCA6M",
		"external_urls": {"spotify": "https://open.spotify.com/artist/2301WleyT98MSxVHPZCA6M"},
		"images": [],
		"followers": {"href": null, "total": 2000000}
	}],
	"limit": 20,
	"next": null,
	"offset": 0,
	"previous": null,
	"total": 1
}`

var mockTopTracksResponse = `{
	"href": "https://api.spotify.com/v1/me/top/tracks",
	"items": [{
		"id": "7iV5W9uYEdYUVa79Axb7Rh",
		"name": "Top Track",
		"artists": [{
			"id": "2301WleyT98MSxVHPZCA6M",
			"name": "Top Artist",
			"type": "artist",
			"uri": "spotify:artist:2301WleyT98MSxVHPZCA6M",
			"href": "https://api.spotify.com/v1/artists/2301WleyT98MSxVHPZCA6M",
			"external_urls": {"spotify": "https://open.spotify.com/artist/2301WleyT98MSxVHPZCA6M"}
		}],
		"album": {
			"id": "5iV5W9uYEdYUVa79Axb7Rh",
			"name": "Top Album",
			"type": "album",
			"uri": "spotify:album:5iV5W9uYEdYUVa79Axb7Rh",
			"href": "https://api.spotify.com/v1/albums/5iV5W9uYEdYUVa79Axb7Rh",
			"external_urls": {"spotify": "https://open.spotify.com/album/5iV5W9uYEdYUVa79Axb7Rh"}
		},
		"duration_ms": 180000,
		"explicit": false,
		"external_urls": {"spotify": "https://open.spotify.com/track/7iV5W9uYEdYUVa79Axb7Rh"},
		"href": "https://api.spotify.com/v1/tracks/7iV5W9uYEdYUVa79Axb7Rh",
		"type": "track",
		"uri": "spotify:track:7iV5W9uYEdYUVa79Axb7Rh",
		"track_number": 1,
		"disc_number": 1,
		"is_local": false,
		"preview_url": null,
		"popularity": 90
	}],
	"limit": 20,
	"next": null,
	"offset": 0,
	"previous": null,
	"total": 1
}`

var mockFollowCheckResponse = `[true, false]`

func createTestUsersService() (*UsersService, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"status": 401, "message": "Unauthorized"}}`))
			return
		}

		switch {
		case r.URL.Path == "/me" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockCurrentUserResponse))
		case r.URL.Path == "/users/otheruser" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockUserResponse))
		case r.URL.Path == "/me/following" && r.Method == "GET" && strings.Contains(r.URL.RawQuery, "type=artist"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockFollowedArtistsResponse))
		case r.URL.Path == "/me/following" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/following" && r.Method == "DELETE":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/following/contains" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockFollowCheckResponse))
		case r.URL.Path == "/me/top/artists" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockTopArtistsResponse))
		case r.URL.Path == "/me/top/tracks" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockTopTracksResponse))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": {"status": 404, "message": "Not Found"}}`))
		}
	}))

	// Create a client with minimal setup for testing
	spotifyClient := &client.Client{}
	requestBuilder := api.NewRequestBuilder(spotifyClient)

	service := NewUsersService(requestBuilder)
	return service, server
}

func TestUsersService_ValidationErrors(t *testing.T) {
	// Create a minimal RequestBuilder for validation testing
	client := &client.Client{}
	requestBuilder := api.NewRequestBuilder(client)
	service := NewUsersService(requestBuilder)

	// Test empty user ID
	_, err := service.GetUser(context.Background(), "")
	if err == nil {
		t.Error("Expected error for empty user ID")
	}

	// Test empty artist IDs for FollowArtists
	err = service.FollowArtists(context.Background(), []string{})
	if err == nil {
		t.Error("Expected error for empty artist IDs in FollowArtists")
	}

	// Test too many artist IDs for FollowArtists
	tooManyIDs := make([]string, 51)
	for i := range tooManyIDs {
		tooManyIDs[i] = "spotify:artist:1301WleyT98MSxVHPZCA6M"
	}
	err = service.FollowArtists(context.Background(), tooManyIDs)
	if err == nil {
		t.Error("Expected error for too many artist IDs in FollowArtists")
	}

	// Test empty artist IDs for UnfollowArtists
	err = service.UnfollowArtists(context.Background(), []string{})
	if err == nil {
		t.Error("Expected error for empty artist IDs in UnfollowArtists")
	}

	// Test too many artist IDs for UnfollowArtists
	err = service.UnfollowArtists(context.Background(), tooManyIDs)
	if err == nil {
		t.Error("Expected error for too many artist IDs in UnfollowArtists")
	}

	// Test empty artist IDs for CheckFollowingArtists
	_, err = service.CheckFollowingArtists(context.Background(), []string{})
	if err == nil {
		t.Error("Expected error for empty artist IDs in CheckFollowingArtists")
	}

	// Test too many artist IDs for CheckFollowingArtists
	_, err = service.CheckFollowingArtists(context.Background(), tooManyIDs)
	if err == nil {
		t.Error("Expected error for too many artist IDs in CheckFollowingArtists")
	}

	// Test invalid time range for GetTopArtists
	options := &TopItemsOptions{
		TimeRange: "invalid_range",
	}
	_, _, err = service.GetTopArtists(context.Background(), options)
	if err == nil {
		t.Error("Expected error for invalid time range")
	}

	// Test invalid limit for GetTopArtists
	options = &TopItemsOptions{
		Limit: 100, // exceeds max limit of 50
	}
	_, _, err = service.GetTopArtists(context.Background(), options)
	if err == nil {
		t.Error("Expected error for limit exceeding maximum")
	}

	// Test negative offset for GetTopArtists
	options = &TopItemsOptions{
		Offset: -1,
	}
	_, _, err = service.GetTopArtists(context.Background(), options)
	if err == nil {
		t.Error("Expected error for negative offset")
	}

	// Test invalid limit for FollowedArtistsOptions
	followOptions := &FollowedArtistsOptions{
		Limit: 100, // exceeds max limit of 50
	}
	_, err = service.GetFollowedArtists(context.Background(), followOptions)
	if err == nil {
		t.Error("Expected error for limit exceeding maximum in GetFollowedArtists")
	}
}

func TestUsersService_TimeRangeValidation(t *testing.T) {
	// Create a minimal RequestBuilder for validation testing
	client := &client.Client{}
	requestBuilder := api.NewRequestBuilder(client)
	service := NewUsersService(requestBuilder)

	// Test valid time ranges
	validRanges := []string{"short_term", "medium_term", "long_term"}

	for _, timeRange := range validRanges {
		options := &TopItemsOptions{
			TimeRange: timeRange,
			Limit:     10,
		}

		// This will fail with network error, but we just want to test validation passes
		_, _, err := service.GetTopArtists(context.Background(), options)
		if err != nil && strings.Contains(err.Error(), "invalid time range") {
			t.Errorf("Expected network error for valid time range '%s', got validation error: %v", timeRange, err)
		}

		_, _, err = service.GetTopTracks(context.Background(), options)
		if err != nil && strings.Contains(err.Error(), "invalid time range") {
			t.Errorf("Expected network error for valid time range '%s', got validation error: %v", timeRange, err)
		}
	}

	// Test invalid time range
	options := &TopItemsOptions{
		TimeRange: "invalid_range",
	}

	_, _, err := service.GetTopArtists(context.Background(), options)
	if err == nil || !strings.Contains(err.Error(), "invalid time range") {
		t.Error("Expected validation error for invalid time range")
	}

	_, _, err = service.GetTopTracks(context.Background(), options)
	if err == nil || !strings.Contains(err.Error(), "invalid time range") {
		t.Error("Expected validation error for invalid time range")
	}
}