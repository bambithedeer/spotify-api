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

// Mock album responses
var mockAlbumResponse = `{
	"id": "4iV5W9uYEdYUVa79Axb7Rh",
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
	"external_urls": {"spotify": "https://open.spotify.com/album/4iV5W9uYEdYUVa79Axb7Rh"},
	"href": "https://api.spotify.com/v1/albums/4iV5W9uYEdYUVa79Axb7Rh",
	"images": [],
	"release_date": "2023-01-01",
	"release_date_precision": "day",
	"type": "album",
	"uri": "spotify:album:4iV5W9uYEdYUVa79Axb7Rh",
	"genres": ["rock", "pop"],
	"label": "Test Records",
	"popularity": 75,
	"tracks": {
		"href": "https://api.spotify.com/v1/albums/4iV5W9uYEdYUVa79Axb7Rh/tracks",
		"items": [{
			"id": "6iV5W9uYEdYUVa79Axb7Rh",
			"name": "Test Track",
			"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
			"duration_ms": 180000,
			"explicit": false,
			"external_urls": {"spotify": "https://open.spotify.com/track/6iV5W9uYEdYUVa79Axb7Rh"},
			"href": "https://api.spotify.com/v1/tracks/6iV5W9uYEdYUVa79Axb7Rh",
			"type": "track",
			"uri": "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
			"available_markets": ["US"],
			"disc_number": 1,
			"track_number": 1,
			"is_local": false,
			"preview_url": null
		}],
		"limit": 50,
		"next": null,
		"offset": 0,
		"previous": null,
		"total": 1
	}
}`

var mockMultipleAlbumsResponse = `{
	"albums": [{
		"id": "4iV5W9uYEdYUVa79Axb7Rh",
		"name": "Test Album 1",
		"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
		"album_type": "album",
		"total_tracks": 10,
		"available_markets": ["US"],
		"external_urls": {"spotify": "https://open.spotify.com/album/4iV5W9uYEdYUVa79Axb7Rh"},
		"href": "https://api.spotify.com/v1/albums/4iV5W9uYEdYUVa79Axb7Rh",
		"images": [],
		"release_date": "2023-01-01",
		"release_date_precision": "day",
		"type": "album",
		"uri": "spotify:album:4iV5W9uYEdYUVa79Axb7Rh"
	}, {
		"id": "5iV5W9uYEdYUVa79Axb7Rh",
		"name": "Test Album 2",
		"artists": [{"id": "2301WleyT98MSxVHPZCA6M", "name": "Another Artist", "type": "artist", "uri": "spotify:artist:2301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/2301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/2301WleyT98MSxVHPZCA6M"}}],
		"album_type": "album",
		"total_tracks": 8,
		"available_markets": ["US"],
		"external_urls": {"spotify": "https://open.spotify.com/album/5iV5W9uYEdYUVa79Axb7Rh"},
		"href": "https://api.spotify.com/v1/albums/5iV5W9uYEdYUVa79Axb7Rh",
		"images": [],
		"release_date": "2023-02-01",
		"release_date_precision": "day",
		"type": "album",
		"uri": "spotify:album:5iV5W9uYEdYUVa79Axb7Rh"
	}]
}`

var mockAlbumTracksResponse = `{
	"href": "https://api.spotify.com/v1/albums/4iV5W9uYEdYUVa79Axb7Rh/tracks",
	"items": [{
		"id": "6iV5W9uYEdYUVa79Axb7Rh",
		"name": "Test Track 1",
		"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
		"duration_ms": 180000,
		"explicit": false,
		"external_urls": {"spotify": "https://open.spotify.com/track/6iV5W9uYEdYUVa79Axb7Rh"},
		"href": "https://api.spotify.com/v1/tracks/6iV5W9uYEdYUVa79Axb7Rh",
		"type": "track",
		"uri": "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
		"available_markets": ["US"],
		"disc_number": 1,
		"track_number": 1,
		"is_local": false,
		"preview_url": null
	}, {
		"id": "7iV5W9uYEdYUVa79Axb7Rh",
		"name": "Test Track 2",
		"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
		"duration_ms": 200000,
		"explicit": false,
		"external_urls": {"spotify": "https://open.spotify.com/track/7iV5W9uYEdYUVa79Axb7Rh"},
		"href": "https://api.spotify.com/v1/tracks/7iV5W9uYEdYUVa79Axb7Rh",
		"type": "track",
		"uri": "spotify:track:7iV5W9uYEdYUVa79Axb7Rh",
		"available_markets": ["US"],
		"disc_number": 1,
		"track_number": 2,
		"is_local": false,
		"preview_url": null
	}],
	"limit": 50,
	"next": null,
	"offset": 0,
	"previous": null,
	"total": 2
}`

var mockNewReleasesResponse = `{
	"albums": {
		"href": "https://api.spotify.com/v1/browse/new-releases",
		"items": [{
			"id": "8iV5W9uYEdYUVa79Axb7Rh",
			"name": "New Release Album",
			"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "New Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
			"album_type": "album",
			"total_tracks": 10,
			"available_markets": ["US"],
			"external_urls": {"spotify": "https://open.spotify.com/album/8iV5W9uYEdYUVa79Axb7Rh"},
			"href": "https://api.spotify.com/v1/albums/8iV5W9uYEdYUVa79Axb7Rh",
			"images": [],
			"release_date": "2023-12-01",
			"release_date_precision": "day",
			"type": "album",
			"uri": "spotify:album:8iV5W9uYEdYUVa79Axb7Rh"
		}],
		"limit": 20,
		"next": null,
		"offset": 0,
		"previous": null,
		"total": 1
	}
}`

func createTestAlbumsService() (*AlbumsService, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"status": 401, "message": "Unauthorized"}}`))
			return
		}

		switch {
		case r.URL.Path == "/albums/4iV5W9uYEdYUVa79Axb7Rh":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockAlbumResponse))
		case r.URL.Path == "/albums" && strings.Contains(r.URL.RawQuery, "ids="):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockMultipleAlbumsResponse))
		case r.URL.Path == "/albums/4iV5W9uYEdYUVa79Axb7Rh/tracks":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockAlbumTracksResponse))
		case r.URL.Path == "/browse/new-releases":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockNewReleasesResponse))
		case strings.HasPrefix(r.URL.Path, "/artists/") && strings.HasSuffix(r.URL.Path, "/albums"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"href": "test", "items": [], "limit": 20, "next": null, "offset": 0, "previous": null, "total": 0}`))
		case strings.HasPrefix(r.URL.Path, "/albums/invalid"):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": {"status": 404, "message": "Album not found"}}`))
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

	// Create request builder and albums service
	builder := api.NewRequestBuilder(client)
	service := NewAlbumsService(builder)

	return service, server
}

func TestAlbumsService_GetAlbum(t *testing.T) {
	service, server := createTestAlbumsService()
	defer server.Close()

	ctx := context.Background()

	// Test successful album retrieval
	album, err := service.GetAlbum(ctx, "4iV5W9uYEdYUVa79Axb7Rh", "")
	if err != nil {
		t.Fatalf("GetAlbum failed: %v", err)
	}

	if album == nil {
		t.Fatal("Expected album result, got nil")
	}

	if album.ID != "4iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected album ID '4iV5W9uYEdYUVa79Axb7Rh', got %s", album.ID)
	}

	if album.Name != "Test Album" {
		t.Errorf("Expected album name 'Test Album', got %s", album.Name)
	}

	if len(album.Artists) == 0 {
		t.Error("Expected album to have artists")
	}

	if album.TotalTracks != 12 {
		t.Errorf("Expected total tracks 12, got %d", album.TotalTracks)
	}
}

func TestAlbumsService_GetAlbumWithMarket(t *testing.T) {
	service, server := createTestAlbumsService()
	defer server.Close()

	ctx := context.Background()

	// Test album retrieval with market
	album, err := service.GetAlbum(ctx, "4iV5W9uYEdYUVa79Axb7Rh", "US")
	if err != nil {
		t.Fatalf("GetAlbum with market failed: %v", err)
	}

	if album == nil {
		t.Fatal("Expected album result, got nil")
	}

	if album.ID != "4iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected album ID '4iV5W9uYEdYUVa79Axb7Rh', got %s", album.ID)
	}
}

func TestAlbumsService_GetAlbums(t *testing.T) {
	service, server := createTestAlbumsService()
	defer server.Close()

	ctx := context.Background()

	// Test multiple albums retrieval
	albumIDs := []string{"4iV5W9uYEdYUVa79Axb7Rh", "5iV5W9uYEdYUVa79Axb7Rh"}
	albums, err := service.GetAlbums(ctx, albumIDs, "")
	if err != nil {
		t.Fatalf("GetAlbums failed: %v", err)
	}

	if len(albums) != 2 {
		t.Errorf("Expected 2 albums, got %d", len(albums))
	}

	if albums[0].ID != "4iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected first album ID '4iV5W9uYEdYUVa79Axb7Rh', got %s", albums[0].ID)
	}

	if albums[1].ID != "5iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected second album ID '5iV5W9uYEdYUVa79Axb7Rh', got %s", albums[1].ID)
	}
}

func TestAlbumsService_GetAlbumTracks(t *testing.T) {
	service, server := createTestAlbumsService()
	defer server.Close()

	ctx := context.Background()

	// Test album tracks retrieval
	tracks, pagination, err := service.GetAlbumTracks(ctx, "4iV5W9uYEdYUVa79Axb7Rh", nil, "")
	if err != nil {
		t.Fatalf("GetAlbumTracks failed: %v", err)
	}

	if tracks == nil {
		t.Fatal("Expected tracks result, got nil")
	}

	if len(tracks.Items) != 2 {
		t.Errorf("Expected 2 tracks, got %d", len(tracks.Items))
	}

	if tracks.Items[0].ID != "6iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected first track ID '6iV5W9uYEdYUVa79Axb7Rh', got %s", tracks.Items[0].ID)
	}

	if tracks.Items[1].ID != "7iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected second track ID '7iV5W9uYEdYUVa79Axb7Rh', got %s", tracks.Items[1].ID)
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestAlbumsService_GetAlbumTracksWithPagination(t *testing.T) {
	service, server := createTestAlbumsService()
	defer server.Close()

	ctx := context.Background()

	// Test with pagination options
	paginationOpts := &api.PaginationOptions{
		Limit:  20,
		Offset: 0,
	}

	tracks, pagination, err := service.GetAlbumTracks(ctx, "4iV5W9uYEdYUVa79Axb7Rh", paginationOpts, "US")
	if err != nil {
		t.Fatalf("GetAlbumTracks with pagination failed: %v", err)
	}

	if tracks == nil {
		t.Fatal("Expected tracks result, got nil")
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestAlbumsService_GetNewReleases(t *testing.T) {
	service, server := createTestAlbumsService()
	defer server.Close()

	ctx := context.Background()

	// Test new releases retrieval
	options := &NewReleasesOptions{
		Country: "US",
		Limit:   20,
		Offset:  0,
	}

	albums, pagination, err := service.GetNewReleases(ctx, options)
	if err != nil {
		t.Fatalf("GetNewReleases failed: %v", err)
	}

	if albums == nil {
		t.Fatal("Expected albums result, got nil")
	}

	if len(albums.Items) != 1 {
		t.Errorf("Expected 1 album, got %d", len(albums.Items))
	}

	if albums.Items[0].ID != "8iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected album ID '8iV5W9uYEdYUVa79Axb7Rh', got %s", albums.Items[0].ID)
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestAlbumsService_GetAlbumsByArtist(t *testing.T) {
	service, server := createTestAlbumsService()
	defer server.Close()

	ctx := context.Background()

	// Test artist albums retrieval
	options := &ArtistAlbumsOptions{
		IncludeGroups: []string{"album", "single"},
		Market:        "US",
		Limit:         20,
		Offset:        0,
	}

	albums, pagination, err := service.GetAlbumsByArtist(ctx, "1301WleyT98MSxVHPZCA6M", options)
	if err != nil {
		t.Fatalf("GetAlbumsByArtist failed: %v", err)
	}

	if albums == nil {
		t.Fatal("Expected albums result, got nil")
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestAlbumsService_ValidationErrors(t *testing.T) {
	service, server := createTestAlbumsService()
	defer server.Close()

	ctx := context.Background()

	// Test invalid album ID
	_, err := service.GetAlbum(ctx, "invalid", "")
	if err == nil {
		t.Error("Expected error for invalid album ID")
	}

	// Test empty album IDs for multiple albums
	_, err = service.GetAlbums(ctx, []string{}, "")
	if err == nil {
		t.Error("Expected error for empty album IDs")
	}

	// Test too many album IDs
	tooManyIDs := make([]string, 21)
	for i := range tooManyIDs {
		tooManyIDs[i] = "4iV5W9uYEdYUVa79Axb7Rh"
	}
	_, err = service.GetAlbums(ctx, tooManyIDs, "")
	if err == nil {
		t.Error("Expected error for too many album IDs")
	}

	// Test invalid market
	_, err = service.GetAlbum(ctx, "4iV5W9uYEdYUVa79Axb7Rh", "INVALID")
	if err == nil {
		t.Error("Expected error for invalid market")
	}

	// Test invalid pagination limit
	invalidPagination := &api.PaginationOptions{
		Limit: 100, // Too high
	}
	_, _, err = service.GetAlbumTracks(ctx, "4iV5W9uYEdYUVa79Axb7Rh", invalidPagination, "")
	if err == nil {
		t.Error("Expected error for invalid pagination limit")
	}
}

func TestAlbumsService_IncludeGroupsValidation(t *testing.T) {
	service, server := createTestAlbumsService()
	defer server.Close()

	ctx := context.Background()

	// Test valid include groups
	validOptions := &ArtistAlbumsOptions{
		IncludeGroups: []string{"album", "single", "appears_on", "compilation"},
	}

	_, _, err := service.GetAlbumsByArtist(ctx, "4iV5W9uYEdYUVa79Axb7Rh", validOptions)
	if err != nil {
		t.Errorf("Expected no error for valid include groups, got %v", err)
	}

	// Test invalid include group
	invalidOptions := &ArtistAlbumsOptions{
		IncludeGroups: []string{"album", "invalid_group"},
	}

	_, _, err = service.GetAlbumsByArtist(ctx, "4iV5W9uYEdYUVa79Axb7Rh", invalidOptions)
	if err == nil {
		t.Error("Expected error for invalid include group")
	}
}

func TestAlbumsService_ValidateIncludeGroups(t *testing.T) {
	service := &AlbumsService{
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