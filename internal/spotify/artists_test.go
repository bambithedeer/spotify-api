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

// Mock artist responses
var mockArtistResponse = `{
	"id": "1301WleyT98MSxVHPZCA6M",
	"name": "Test Artist",
	"type": "artist",
	"uri": "spotify:artist:1301WleyT98MSxVHPZCA6M",
	"href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M",
	"external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"},
	"followers": {"href": null, "total": 1000000},
	"genres": ["rock", "pop", "alternative"],
	"images": [{
		"url": "https://example.com/image.jpg",
		"height": 640,
		"width": 640
	}],
	"popularity": 85
}`

var mockMultipleArtistsResponse = `{
	"artists": [{
		"id": "1301WleyT98MSxVHPZCA6M",
		"name": "Test Artist 1",
		"type": "artist",
		"uri": "spotify:artist:1301WleyT98MSxVHPZCA6M",
		"href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M",
		"external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"},
		"followers": {"href": null, "total": 1000000},
		"genres": ["rock", "pop"],
		"images": [],
		"popularity": 85
	}, {
		"id": "2301WleyT98MSxVHPZCA6M",
		"name": "Test Artist 2",
		"type": "artist",
		"uri": "spotify:artist:2301WleyT98MSxVHPZCA6M",
		"href": "https://api.spotify.com/v1/artists/2301WleyT98MSxVHPZCA6M",
		"external_urls": {"spotify": "https://open.spotify.com/artist/2301WleyT98MSxVHPZCA6M"},
		"followers": {"href": null, "total": 500000},
		"genres": ["jazz", "blues"],
		"images": [],
		"popularity": 70
	}]
}`

var mockArtistAlbumsResponse = `{
	"href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M/albums",
	"items": [{
		"id": "4iV5W9uYEdYUVa79Axb7Rh",
		"name": "Test Album",
		"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
		"album_type": "album",
		"total_tracks": 12,
		"available_markets": ["US"],
		"external_urls": {"spotify": "https://open.spotify.com/album/4iV5W9uYEdYUVa79Axb7Rh"},
		"href": "https://api.spotify.com/v1/albums/4iV5W9uYEdYUVa79Axb7Rh",
		"images": [],
		"release_date": "2023-01-01",
		"release_date_precision": "day",
		"type": "album",
		"uri": "spotify:album:4iV5W9uYEdYUVa79Axb7Rh"
	}],
	"limit": 20,
	"next": null,
	"offset": 0,
	"previous": null,
	"total": 1
}`

var mockArtistTopTracksResponse = `{
	"tracks": [{
		"id": "6iV5W9uYEdYUVa79Axb7Rh",
		"name": "Top Track 1",
		"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
		"duration_ms": 200000,
		"explicit": false,
		"popularity": 90,
		"external_urls": {"spotify": "https://open.spotify.com/track/6iV5W9uYEdYUVa79Axb7Rh"},
		"href": "https://api.spotify.com/v1/tracks/6iV5W9uYEdYUVa79Axb7Rh",
		"type": "track",
		"uri": "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
		"available_markets": ["US"],
		"disc_number": 1,
		"track_number": 1,
		"is_local": false,
		"preview_url": "https://example.com/preview.mp3",
		"external_ids": {"isrc": "TEST123456789"}
	}, {
		"id": "7iV5W9uYEdYUVa79Axb7Rh",
		"name": "Top Track 2",
		"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
		"duration_ms": 180000,
		"explicit": false,
		"popularity": 85,
		"external_urls": {"spotify": "https://open.spotify.com/track/7iV5W9uYEdYUVa79Axb7Rh"},
		"href": "https://api.spotify.com/v1/tracks/7iV5W9uYEdYUVa79Axb7Rh",
		"type": "track",
		"uri": "spotify:track:7iV5W9uYEdYUVa79Axb7Rh",
		"available_markets": ["US"],
		"disc_number": 1,
		"track_number": 2,
		"is_local": false,
		"preview_url": null,
		"external_ids": {}
	}]
}`

var mockRelatedArtistsResponse = `{
	"artists": [{
		"id": "2301WleyT98MSxVHPZCA6M",
		"name": "Related Artist 1",
		"type": "artist",
		"uri": "spotify:artist:2301WleyT98MSxVHPZCA6M",
		"href": "https://api.spotify.com/v1/artists/2301WleyT98MSxVHPZCA6M",
		"external_urls": {"spotify": "https://open.spotify.com/artist/2301WleyT98MSxVHPZCA6M"},
		"followers": {"href": null, "total": 800000},
		"genres": ["rock", "alternative"],
		"images": [],
		"popularity": 80
	}, {
		"id": "3301WleyT98MSxVHPZCA6M",
		"name": "Related Artist 2",
		"type": "artist",
		"uri": "spotify:artist:3301WleyT98MSxVHPZCA6M",
		"href": "https://api.spotify.com/v1/artists/3301WleyT98MSxVHPZCA6M",
		"external_urls": {"spotify": "https://open.spotify.com/artist/3301WleyT98MSxVHPZCA6M"},
		"followers": {"href": null, "total": 600000},
		"genres": ["pop", "indie"],
		"images": [],
		"popularity": 75
	}]
}`

func createTestArtistsService() (*ArtistsService, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"status": 401, "message": "Unauthorized"}}`))
			return
		}

		switch {
		case r.URL.Path == "/artists/1301WleyT98MSxVHPZCA6M":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockArtistResponse))
		case r.URL.Path == "/artists" && strings.Contains(r.URL.RawQuery, "ids="):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockMultipleArtistsResponse))
		case r.URL.Path == "/artists/1301WleyT98MSxVHPZCA6M/albums":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockArtistAlbumsResponse))
		case r.URL.Path == "/artists/1301WleyT98MSxVHPZCA6M/top-tracks":
			if r.URL.Query().Get("market") == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": {"status": 400, "message": "Market parameter required"}}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockArtistTopTracksResponse))
		case r.URL.Path == "/artists/1301WleyT98MSxVHPZCA6M/related-artists":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockRelatedArtistsResponse))
		case strings.HasPrefix(r.URL.Path, "/artists/invalid"):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": {"status": 404, "message": "Artist not found"}}`))
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

	// Create request builder and artists service
	builder := api.NewRequestBuilder(client)
	service := NewArtistsService(builder)

	return service, server
}

func TestArtistsService_GetArtist(t *testing.T) {
	service, server := createTestArtistsService()
	defer server.Close()

	ctx := context.Background()

	// Test successful artist retrieval
	artist, err := service.GetArtist(ctx, "1301WleyT98MSxVHPZCA6M")
	if err != nil {
		t.Fatalf("GetArtist failed: %v", err)
	}

	if artist == nil {
		t.Fatal("Expected artist result, got nil")
	}

	if artist.ID != "1301WleyT98MSxVHPZCA6M" {
		t.Errorf("Expected artist ID '1301WleyT98MSxVHPZCA6M', got %s", artist.ID)
	}

	if artist.Name != "Test Artist" {
		t.Errorf("Expected artist name 'Test Artist', got %s", artist.Name)
	}

	if artist.Popularity != 85 {
		t.Errorf("Expected popularity 85, got %d", artist.Popularity)
	}

	if len(artist.Genres) != 3 {
		t.Errorf("Expected 3 genres, got %d", len(artist.Genres))
	}

	if artist.Followers.Total != 1000000 {
		t.Errorf("Expected followers 1000000, got %d", artist.Followers.Total)
	}
}

func TestArtistsService_GetArtists(t *testing.T) {
	service, server := createTestArtistsService()
	defer server.Close()

	ctx := context.Background()

	// Test multiple artists retrieval
	artistIDs := []string{"1301WleyT98MSxVHPZCA6M", "2301WleyT98MSxVHPZCA6M"}
	artists, err := service.GetArtists(ctx, artistIDs)
	if err != nil {
		t.Fatalf("GetArtists failed: %v", err)
	}

	if len(artists) != 2 {
		t.Errorf("Expected 2 artists, got %d", len(artists))
	}

	if artists[0].ID != "1301WleyT98MSxVHPZCA6M" {
		t.Errorf("Expected first artist ID '1301WleyT98MSxVHPZCA6M', got %s", artists[0].ID)
	}

	if artists[1].ID != "2301WleyT98MSxVHPZCA6M" {
		t.Errorf("Expected second artist ID '2301WleyT98MSxVHPZCA6M', got %s", artists[1].ID)
	}

	if artists[0].Name != "Test Artist 1" {
		t.Errorf("Expected first artist name 'Test Artist 1', got %s", artists[0].Name)
	}

	if artists[1].Name != "Test Artist 2" {
		t.Errorf("Expected second artist name 'Test Artist 2', got %s", artists[1].Name)
	}
}

func TestArtistsService_GetArtistAlbums(t *testing.T) {
	service, server := createTestArtistsService()
	defer server.Close()

	ctx := context.Background()

	// Test artist albums retrieval
	options := &ArtistAlbumsOptions{
		IncludeGroups: []string{"album", "single"},
		Market:        "US",
		Limit:         20,
		Offset:        0,
	}

	albums, pagination, err := service.GetArtistAlbums(ctx, "1301WleyT98MSxVHPZCA6M", options)
	if err != nil {
		t.Fatalf("GetArtistAlbums failed: %v", err)
	}

	if albums == nil {
		t.Fatal("Expected albums result, got nil")
	}

	if len(albums.Items) != 1 {
		t.Errorf("Expected 1 album, got %d", len(albums.Items))
	}

	if albums.Items[0].ID != "4iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected album ID '4iV5W9uYEdYUVa79Axb7Rh', got %s", albums.Items[0].ID)
	}

	if albums.Items[0].Name != "Test Album" {
		t.Errorf("Expected album name 'Test Album', got %s", albums.Items[0].Name)
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestArtistsService_GetArtistAlbumsWithoutOptions(t *testing.T) {
	service, server := createTestArtistsService()
	defer server.Close()

	ctx := context.Background()

	// Test artist albums retrieval without options
	albums, pagination, err := service.GetArtistAlbums(ctx, "1301WleyT98MSxVHPZCA6M", nil)
	if err != nil {
		t.Fatalf("GetArtistAlbums without options failed: %v", err)
	}

	if albums == nil {
		t.Fatal("Expected albums result, got nil")
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestArtistsService_GetArtistTopTracks(t *testing.T) {
	service, server := createTestArtistsService()
	defer server.Close()

	ctx := context.Background()

	// Test top tracks retrieval
	tracks, err := service.GetArtistTopTracks(ctx, "1301WleyT98MSxVHPZCA6M", "US")
	if err != nil {
		t.Fatalf("GetArtistTopTracks failed: %v", err)
	}

	if len(tracks) != 2 {
		t.Errorf("Expected 2 tracks, got %d", len(tracks))
	}

	if tracks[0].ID != "6iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected first track ID '6iV5W9uYEdYUVa79Axb7Rh', got %s", tracks[0].ID)
	}

	if tracks[1].ID != "7iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected second track ID '7iV5W9uYEdYUVa79Axb7Rh', got %s", tracks[1].ID)
	}

	if tracks[0].Name != "Top Track 1" {
		t.Errorf("Expected first track name 'Top Track 1', got %s", tracks[0].Name)
	}

	if tracks[0].Popularity != 90 {
		t.Errorf("Expected first track popularity 90, got %d", tracks[0].Popularity)
	}
}

func TestArtistsService_GetRelatedArtists(t *testing.T) {
	service, server := createTestArtistsService()
	defer server.Close()

	ctx := context.Background()

	// Test related artists retrieval
	artists, err := service.GetRelatedArtists(ctx, "1301WleyT98MSxVHPZCA6M")
	if err != nil {
		t.Fatalf("GetRelatedArtists failed: %v", err)
	}

	if len(artists) != 2 {
		t.Errorf("Expected 2 artists, got %d", len(artists))
	}

	if artists[0].ID != "2301WleyT98MSxVHPZCA6M" {
		t.Errorf("Expected first artist ID '2301WleyT98MSxVHPZCA6M', got %s", artists[0].ID)
	}

	if artists[1].ID != "3301WleyT98MSxVHPZCA6M" {
		t.Errorf("Expected second artist ID '3301WleyT98MSxVHPZCA6M', got %s", artists[1].ID)
	}

	if artists[0].Name != "Related Artist 1" {
		t.Errorf("Expected first artist name 'Related Artist 1', got %s", artists[0].Name)
	}

	if artists[1].Name != "Related Artist 2" {
		t.Errorf("Expected second artist name 'Related Artist 2', got %s", artists[1].Name)
	}
}

func TestArtistsService_ValidationErrors(t *testing.T) {
	service, server := createTestArtistsService()
	defer server.Close()

	ctx := context.Background()

	// Test invalid artist ID
	_, err := service.GetArtist(ctx, "invalid")
	if err == nil {
		t.Error("Expected error for invalid artist ID")
	}

	// Test empty artist IDs for multiple artists
	_, err = service.GetArtists(ctx, []string{})
	if err == nil {
		t.Error("Expected error for empty artist IDs")
	}

	// Test too many artist IDs
	tooManyIDs := make([]string, 51)
	for i := range tooManyIDs {
		tooManyIDs[i] = "1301WleyT98MSxVHPZCA6M"
	}
	_, err = service.GetArtists(ctx, tooManyIDs)
	if err == nil {
		t.Error("Expected error for too many artist IDs")
	}

	// Test top tracks without market
	_, err = service.GetArtistTopTracks(ctx, "1301WleyT98MSxVHPZCA6M", "")
	if err == nil {
		t.Error("Expected error for missing market parameter")
	}

	// Test invalid market
	_, err = service.GetArtistTopTracks(ctx, "1301WleyT98MSxVHPZCA6M", "INVALID")
	if err == nil {
		t.Error("Expected error for invalid market")
	}

	// Test invalid pagination limit for albums
	invalidOptions := &ArtistAlbumsOptions{
		Limit: 100, // Too high
	}
	_, _, err = service.GetArtistAlbums(ctx, "1301WleyT98MSxVHPZCA6M", invalidOptions)
	if err == nil {
		t.Error("Expected error for invalid pagination limit")
	}
}

func TestArtistsService_IncludeGroupsValidation(t *testing.T) {
	service, server := createTestArtistsService()
	defer server.Close()

	ctx := context.Background()

	// Test valid include groups
	validOptions := &ArtistAlbumsOptions{
		IncludeGroups: []string{"album", "single", "appears_on", "compilation"},
	}

	_, _, err := service.GetArtistAlbums(ctx, "1301WleyT98MSxVHPZCA6M", validOptions)
	if err != nil {
		t.Errorf("Expected no error for valid include groups, got %v", err)
	}

	// Test invalid include group
	invalidOptions := &ArtistAlbumsOptions{
		IncludeGroups: []string{"album", "invalid_group"},
	}

	_, _, err = service.GetArtistAlbums(ctx, "1301WleyT98MSxVHPZCA6M", invalidOptions)
	if err == nil {
		t.Error("Expected error for invalid include group")
	}
}

func TestArtistsService_ValidateIncludeGroups(t *testing.T) {
	service := &ArtistsService{
		validator: api.NewValidator(),
	}

	// Test valid groups
	validGroups := []string{"album", "single", "appears_on", "compilation"}
	err := service.validateIncludeGroups(validGroups)
	if err != nil {
		t.Errorf("Expected no error for valid groups, got %v", err)
	}

	// Test invalid group
	invalidGroups := []string{"album", "invalid"}
	err = service.validateIncludeGroups(invalidGroups)
	if err == nil {
		t.Error("Expected error for invalid group")
	}
}