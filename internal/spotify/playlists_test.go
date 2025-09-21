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

// Mock playlist responses
var mockPlaylistResponse = `{
	"collaborative": false,
	"description": "Test playlist description",
	"external_urls": {"spotify": "https://open.spotify.com/playlist/37i9dQZF1DX0XUsuxWHRQd"},
	"followers": {"href": null, "total": 100},
	"href": "https://api.spotify.com/v1/playlists/37i9dQZF1DX0XUsuxWHRQd",
	"id": "37i9dQZF1DX0XUsuxWHRQd",
	"images": [{
		"url": "https://example.com/image.jpg",
		"height": 640,
		"width": 640
	}],
	"name": "Test Playlist",
	"owner": {
		"id": "testuser",
		"display_name": "Test User",
		"type": "user",
		"uri": "spotify:user:testuser",
		"href": "https://api.spotify.com/v1/users/testuser",
		"external_urls": {"spotify": "https://open.spotify.com/user/testuser"}
	},
	"public": true,
	"snapshot_id": "MTEsOGZmN2ZmYmIwNzE0NDU3NmZhNTEwNzBkNTU3MTlkYjgwYTMwNzFjMQ==",
	"tracks": {
		"href": "https://api.spotify.com/v1/playlists/37i9dQZF1DX0XUsuxWHRQd/tracks",
		"items": [{
			"added_at": "2023-01-01T00:00:00Z",
			"added_by": {
				"id": "testuser",
				"type": "user",
				"uri": "spotify:user:testuser",
				"href": "https://api.spotify.com/v1/users/testuser",
				"external_urls": {"spotify": "https://open.spotify.com/user/testuser"}
			},
			"is_local": false,
			"track": {
				"id": "6iV5W9uYEdYUVa79Axb7Rh",
				"name": "Test Track",
				"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
				"duration_ms": 180000,
				"explicit": false,
				"popularity": 75,
				"external_urls": {"spotify": "https://open.spotify.com/track/6iV5W9uYEdYUVa79Axb7Rh"},
				"href": "https://api.spotify.com/v1/tracks/6iV5W9uYEdYUVa79Axb7Rh",
				"type": "track",
				"uri": "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
				"available_markets": ["US"],
				"disc_number": 1,
				"track_number": 1,
				"is_local": false,
				"preview_url": null,
				"external_ids": {}
			}
		}],
		"limit": 100,
		"next": null,
		"offset": 0,
		"previous": null,
		"total": 1
	},
	"type": "playlist",
	"uri": "spotify:playlist:37i9dQZF1DX0XUsuxWHRQd"
}`

var mockPlaylistTracksResponse = `{
	"href": "https://api.spotify.com/v1/playlists/37i9dQZF1DX0XUsuxWHRQd/tracks",
	"items": [{
		"added_at": "2023-01-01T00:00:00Z",
		"added_by": {
			"id": "testuser",
			"type": "user",
			"uri": "spotify:user:testuser",
			"href": "https://api.spotify.com/v1/users/testuser",
			"external_urls": {"spotify": "https://open.spotify.com/user/testuser"}
		},
		"is_local": false,
		"track": {
			"id": "6iV5W9uYEdYUVa79Axb7Rh",
			"name": "Test Track 1",
			"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
			"duration_ms": 180000,
			"explicit": false,
			"popularity": 75,
			"external_urls": {"spotify": "https://open.spotify.com/track/6iV5W9uYEdYUVa79Axb7Rh"},
			"href": "https://api.spotify.com/v1/tracks/6iV5W9uYEdYUVa79Axb7Rh",
			"type": "track",
			"uri": "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
			"available_markets": ["US"],
			"disc_number": 1,
			"track_number": 1,
			"is_local": false,
			"preview_url": null,
			"external_ids": {}
		}
	}, {
		"added_at": "2023-01-02T00:00:00Z",
		"added_by": {
			"id": "testuser",
			"type": "user",
			"uri": "spotify:user:testuser",
			"href": "https://api.spotify.com/v1/users/testuser",
			"external_urls": {"spotify": "https://open.spotify.com/user/testuser"}
		},
		"is_local": false,
		"track": {
			"id": "7iV5W9uYEdYUVa79Axb7Rh",
			"name": "Test Track 2",
			"artists": [{"id": "2301WleyT98MSxVHPZCA6M", "name": "Another Artist", "type": "artist", "uri": "spotify:artist:2301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/2301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/2301WleyT98MSxVHPZCA6M"}}],
			"duration_ms": 210000,
			"explicit": true,
			"popularity": 80,
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
		}
	}],
	"limit": 100,
	"next": null,
	"offset": 0,
	"previous": null,
	"total": 2
}`

var mockUserPlaylistsResponse = `{
	"href": "https://api.spotify.com/v1/me/playlists",
	"items": [{
		"collaborative": false,
		"description": "My first playlist",
		"external_urls": {"spotify": "https://open.spotify.com/playlist/37i9dQZF1DX0XUsuxWHRQd"},
		"href": "https://api.spotify.com/v1/playlists/37i9dQZF1DX0XUsuxWHRQd",
		"id": "37i9dQZF1DX0XUsuxWHRQd",
		"images": [],
		"name": "My Playlist 1",
		"owner": {
			"id": "testuser",
			"display_name": "Test User",
			"type": "user",
			"uri": "spotify:user:testuser",
			"href": "https://api.spotify.com/v1/users/testuser",
			"external_urls": {"spotify": "https://open.spotify.com/user/testuser"}
		},
		"public": true,
		"snapshot_id": "MTEsOGZmN2ZmYmIwNzE0NDU3NmZhNTEwNzBkNTU3MTlkYjgwYTMwNzFjMQ==",
		"tracks": {
			"href": "https://api.spotify.com/v1/playlists/37i9dQZF1DX0XUsuxWHRQd/tracks",
			"total": 5
		},
		"type": "playlist",
		"uri": "spotify:playlist:37i9dQZF1DX0XUsuxWHRQd"
	}],
	"limit": 20,
	"next": null,
	"offset": 0,
	"previous": null,
	"total": 1
}`

var mockCreatedPlaylistResponse = `{
	"collaborative": false,
	"description": "New playlist description",
	"external_urls": {"spotify": "https://open.spotify.com/playlist/1BxfuPKGuaTgP6aM0NrF0N"},
	"followers": {"href": null, "total": 0},
	"href": "https://api.spotify.com/v1/playlists/1BxfuPKGuaTgP6aM0NrF0N",
	"id": "1BxfuPKGuaTgP6aM0NrF0N",
	"images": [],
	"name": "New Playlist",
	"owner": {
		"id": "testuser",
		"display_name": "Test User",
		"type": "user",
		"uri": "spotify:user:testuser",
		"href": "https://api.spotify.com/v1/users/testuser",
		"external_urls": {"spotify": "https://open.spotify.com/user/testuser"}
	},
	"public": false,
	"snapshot_id": "MTEsOGZmN2ZmYmIwNzE0NDU3NmZhNTEwNzBkNTU3MTlkYjgwYTMwNzFjMQ==",
	"tracks": {
		"href": "https://api.spotify.com/v1/playlists/1BxfuPKGuaTgP6aM0NrF0N/tracks",
		"items": [],
		"limit": 100,
		"next": null,
		"offset": 0,
		"previous": null,
		"total": 0
	},
	"type": "playlist",
	"uri": "spotify:playlist:1BxfuPKGuaTgP6aM0NrF0N"
}`

var mockSnapshotResponse = `{
	"snapshot_id": "MTEsOGZmN2ZmYmIwNzE0NDU3NmZhNTEwNzBkNTU3MTlkYjgwYTMwNzFjMQ=="
}`

func createTestPlaylistsService() (*PlaylistsService, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"status": 401, "message": "Unauthorized"}}`))
			return
		}

		switch {
		case r.URL.Path == "/playlists/37i9dQZF1DX0XUsuxWHRQd" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockPlaylistResponse))
		case r.URL.Path == "/playlists/37i9dQZF1DX0XUsuxWHRQd/tracks" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockPlaylistTracksResponse))
		case r.URL.Path == "/playlists/37i9dQZF1DX0XUsuxWHRQd/tracks" && r.Method == "POST":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(mockSnapshotResponse))
		case r.URL.Path == "/playlists/37i9dQZF1DX0XUsuxWHRQd/tracks" && r.Method == "DELETE":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockSnapshotResponse))
		case r.URL.Path == "/playlists/37i9dQZF1DX0XUsuxWHRQd/tracks" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockSnapshotResponse))
		case r.URL.Path == "/playlists/37i9dQZF1DX0XUsuxWHRQd" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(``)) // Update returns empty response
		case r.URL.Path == "/me/playlists":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockUserPlaylistsResponse))
		case r.URL.Path == "/users/testuser/playlists" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockUserPlaylistsResponse))
		case r.URL.Path == "/users/testuser/playlists" && r.Method == "POST":
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(mockCreatedPlaylistResponse))
		case strings.HasPrefix(r.URL.Path, "/playlists/invalid"):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": {"status": 404, "message": "Playlist not found"}}`))
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

	// Create request builder and playlists service
	builder := api.NewRequestBuilder(client)
	service := NewPlaylistsService(builder)

	return service, server
}

func TestPlaylistsService_GetPlaylist(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test successful playlist retrieval
	playlist, err := service.GetPlaylist(ctx, "37i9dQZF1DX0XUsuxWHRQd", nil)
	if err != nil {
		t.Fatalf("GetPlaylist failed: %v", err)
	}

	if playlist == nil {
		t.Fatal("Expected playlist result, got nil")
	}

	if playlist.ID != "37i9dQZF1DX0XUsuxWHRQd" {
		t.Errorf("Expected playlist ID '37i9dQZF1DX0XUsuxWHRQd', got %s", playlist.ID)
	}

	if playlist.Name != "Test Playlist" {
		t.Errorf("Expected playlist name 'Test Playlist', got %s", playlist.Name)
	}

	if playlist.Description != "Test playlist description" {
		t.Errorf("Expected playlist description 'Test playlist description', got %s", playlist.Description)
	}

	if !playlist.Public {
		t.Error("Expected playlist to be public")
	}

	if playlist.Owner.ID != "testuser" {
		t.Errorf("Expected owner ID 'testuser', got %s", playlist.Owner.ID)
	}
}

func TestPlaylistsService_GetPlaylistWithOptions(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test playlist retrieval with options
	options := &PlaylistOptions{
		Market:          "US",
		Fields:          "id,name,description",
		AdditionalTypes: []string{"track", "episode"},
	}

	playlist, err := service.GetPlaylist(ctx, "37i9dQZF1DX0XUsuxWHRQd", options)
	if err != nil {
		t.Fatalf("GetPlaylist with options failed: %v", err)
	}

	if playlist == nil {
		t.Fatal("Expected playlist result, got nil")
	}
}

func TestPlaylistsService_GetPlaylistTracks(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test playlist tracks retrieval
	tracks, pagination, err := service.GetPlaylistTracks(ctx, "37i9dQZF1DX0XUsuxWHRQd", nil)
	if err != nil {
		t.Fatalf("GetPlaylistTracks failed: %v", err)
	}

	if tracks == nil {
		t.Fatal("Expected tracks result, got nil")
	}

	if len(tracks.Items) != 2 {
		t.Errorf("Expected 2 tracks, got %d", len(tracks.Items))
	}

	track0, ok := tracks.Items[0].Track.(map[string]interface{})
	if !ok || track0["id"] != "6iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected first track ID '6iV5W9uYEdYUVa79Axb7Rh', got %v", track0["id"])
	}

	track1, ok := tracks.Items[1].Track.(map[string]interface{})
	if !ok || track1["id"] != "7iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected second track ID '7iV5W9uYEdYUVa79Axb7Rh', got %v", track1["id"])
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestPlaylistsService_GetPlaylistTracksWithOptions(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test with pagination options
	options := &PlaylistTracksOptions{
		Market:          "US",
		Fields:          "items(track(id,name))",
		Limit:           50,
		Offset:          0,
		AdditionalTypes: []string{"track"},
	}

	tracks, pagination, err := service.GetPlaylistTracks(ctx, "37i9dQZF1DX0XUsuxWHRQd", options)
	if err != nil {
		t.Fatalf("GetPlaylistTracks with options failed: %v", err)
	}

	if tracks == nil {
		t.Fatal("Expected tracks result, got nil")
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestPlaylistsService_GetUserPlaylists(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test user playlists retrieval
	playlists, pagination, err := service.GetUserPlaylists(ctx, nil)
	if err != nil {
		t.Fatalf("GetUserPlaylists failed: %v", err)
	}

	if playlists == nil {
		t.Fatal("Expected playlists result, got nil")
	}

	if len(playlists.Items) != 1 {
		t.Errorf("Expected 1 playlist, got %d", len(playlists.Items))
	}

	if playlists.Items[0].ID != "37i9dQZF1DX0XUsuxWHRQd" {
		t.Errorf("Expected playlist ID '37i9dQZF1DX0XUsuxWHRQd', got %s", playlists.Items[0].ID)
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestPlaylistsService_GetUserPlaylistsByID(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test user playlists by ID retrieval
	playlists, pagination, err := service.GetUserPlaylistsByID(ctx, "testuser", nil)
	if err != nil {
		t.Fatalf("GetUserPlaylistsByID failed: %v", err)
	}

	if playlists == nil {
		t.Fatal("Expected playlists result, got nil")
	}

	if pagination == nil {
		t.Fatal("Expected pagination info, got nil")
	}
}

func TestPlaylistsService_CreatePlaylist(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test playlist creation
	isPublic := false
	request := &CreatePlaylistRequest{
		Name:        "New Playlist",
		Description: "New playlist description",
		Public:      &isPublic,
	}

	playlist, err := service.CreatePlaylist(ctx, "testuser", request)
	if err != nil {
		t.Fatalf("CreatePlaylist failed: %v", err)
	}

	if playlist == nil {
		t.Fatal("Expected playlist result, got nil")
	}

	if playlist.ID != "1BxfuPKGuaTgP6aM0NrF0N" {
		t.Errorf("Expected playlist ID '1BxfuPKGuaTgP6aM0NrF0N', got %s", playlist.ID)
	}

	if playlist.Name != "New Playlist" {
		t.Errorf("Expected playlist name 'New Playlist', got %s", playlist.Name)
	}

	if playlist.Public {
		t.Error("Expected playlist to be private")
	}
}

func TestPlaylistsService_UpdatePlaylist(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test playlist update
	newName := "Updated Playlist"
	newDescription := "Updated description"
	request := &UpdatePlaylistRequest{
		Name:        &newName,
		Description: &newDescription,
	}

	err := service.UpdatePlaylist(ctx, "37i9dQZF1DX0XUsuxWHRQd", request)
	if err != nil {
		t.Fatalf("UpdatePlaylist failed: %v", err)
	}
}

func TestPlaylistsService_AddTracksToPlaylist(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test adding tracks
	request := &AddTracksRequest{
		URIs: []string{
			"spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
			"spotify:track:7iV5W9uYEdYUVa79Axb7Rh",
		},
	}

	response, err := service.AddTracksToPlaylist(ctx, "37i9dQZF1DX0XUsuxWHRQd", request)
	if err != nil {
		t.Fatalf("AddTracksToPlaylist failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected snapshot response, got nil")
	}

	if response.SnapshotID == "" {
		t.Error("Expected snapshot ID to be set")
	}
}

func TestPlaylistsService_RemoveTracksFromPlaylist(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test removing tracks
	request := &RemoveTracksRequest{
		Tracks: []TrackToRemove{
			{
				URI:       "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
				Positions: []int{0},
			},
		},
	}

	response, err := service.RemoveTracksFromPlaylist(ctx, "37i9dQZF1DX0XUsuxWHRQd", request)
	if err != nil {
		t.Fatalf("RemoveTracksFromPlaylist failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected snapshot response, got nil")
	}
}

func TestPlaylistsService_ReorderPlaylistTracks(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test reordering tracks
	request := &ReorderTracksRequest{
		RangeStart:   0,
		InsertBefore: 2,
	}

	response, err := service.ReorderPlaylistTracks(ctx, "37i9dQZF1DX0XUsuxWHRQd", request)
	if err != nil {
		t.Fatalf("ReorderPlaylistTracks failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected snapshot response, got nil")
	}
}

func TestPlaylistsService_ReplacePlaylistTracks(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test replacing tracks
	trackURIs := []string{
		"spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
		"spotify:track:7iV5W9uYEdYUVa79Axb7Rh",
	}

	response, err := service.ReplacePlaylistTracks(ctx, "37i9dQZF1DX0XUsuxWHRQd", trackURIs)
	if err != nil {
		t.Fatalf("ReplacePlaylistTracks failed: %v", err)
	}

	if response == nil {
		t.Fatal("Expected snapshot response, got nil")
	}
}

func TestPlaylistsService_ValidationErrors(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test invalid playlist ID
	_, err := service.GetPlaylist(ctx, "invalid", nil)
	if err == nil {
		t.Error("Expected error for invalid playlist ID")
	}

	// Test empty user ID for creating playlist
	_, err = service.CreatePlaylist(ctx, "", &CreatePlaylistRequest{Name: "Test"})
	if err == nil {
		t.Error("Expected error for empty user ID")
	}

	// Test nil create request
	_, err = service.CreatePlaylist(ctx, "testuser", nil)
	if err == nil {
		t.Error("Expected error for nil create request")
	}

	// Test empty playlist name
	_, err = service.CreatePlaylist(ctx, "testuser", &CreatePlaylistRequest{Name: ""})
	if err == nil {
		t.Error("Expected error for empty playlist name")
	}

	// Test too long description
	longDesc := strings.Repeat("a", 301)
	_, err = service.CreatePlaylist(ctx, "testuser", &CreatePlaylistRequest{
		Name:        "Test",
		Description: longDesc,
	})
	if err == nil {
		t.Error("Expected error for too long description")
	}

	// Test invalid market
	options := &PlaylistOptions{Market: "INVALID"}
	_, err = service.GetPlaylist(ctx, "37i9dQZF1DX0XUsuxWHRQd", options)
	if err == nil {
		t.Error("Expected error for invalid market")
	}

	// Test invalid additional types
	options = &PlaylistOptions{AdditionalTypes: []string{"invalid"}}
	_, err = service.GetPlaylist(ctx, "37i9dQZF1DX0XUsuxWHRQd", options)
	if err == nil {
		t.Error("Expected error for invalid additional types")
	}

	// Test invalid pagination limit
	tracksOptions := &PlaylistTracksOptions{Limit: 200}
	_, _, err = service.GetPlaylistTracks(ctx, "37i9dQZF1DX0XUsuxWHRQd", tracksOptions)
	if err == nil {
		t.Error("Expected error for invalid pagination limit")
	}

	// Test empty tracks for adding
	_, err = service.AddTracksToPlaylist(ctx, "37i9dQZF1DX0XUsuxWHRQd", &AddTracksRequest{URIs: []string{}})
	if err == nil {
		t.Error("Expected error for empty tracks")
	}

	// Test too many tracks for adding
	tooManyTracks := make([]string, 101)
	for i := range tooManyTracks {
		tooManyTracks[i] = "spotify:track:6iV5W9uYEdYUVa79Axb7Rh"
	}
	_, err = service.AddTracksToPlaylist(ctx, "37i9dQZF1DX0XUsuxWHRQd", &AddTracksRequest{URIs: tooManyTracks})
	if err == nil {
		t.Error("Expected error for too many tracks")
	}

	// Test invalid track URI
	_, err = service.AddTracksToPlaylist(ctx, "37i9dQZF1DX0XUsuxWHRQd", &AddTracksRequest{
		URIs: []string{"invalid:uri"},
	})
	if err == nil {
		t.Error("Expected error for invalid track URI")
	}

	// Test too many tracks for replacing
	_, err = service.ReplacePlaylistTracks(ctx, "37i9dQZF1DX0XUsuxWHRQd", tooManyTracks)
	if err == nil {
		t.Error("Expected error for too many tracks for replacing")
	}
}

func TestPlaylistsService_ReorderValidationErrors(t *testing.T) {
	service, server := createTestPlaylistsService()
	defer server.Close()

	ctx := context.Background()

	// Test negative range start
	_, err := service.ReorderPlaylistTracks(ctx, "37i9dQZF1DX0XUsuxWHRQd", &ReorderTracksRequest{
		RangeStart:   -1,
		InsertBefore: 2,
	})
	if err == nil {
		t.Error("Expected error for negative range start")
	}

	// Test negative insert before
	_, err = service.ReorderPlaylistTracks(ctx, "37i9dQZF1DX0XUsuxWHRQd", &ReorderTracksRequest{
		RangeStart:   0,
		InsertBefore: -1,
	})
	if err == nil {
		t.Error("Expected error for negative insert before")
	}

	// Test invalid range length
	rangeLength := 0
	_, err = service.ReorderPlaylistTracks(ctx, "37i9dQZF1DX0XUsuxWHRQd", &ReorderTracksRequest{
		RangeStart:   0,
		InsertBefore: 2,
		RangeLength:  &rangeLength,
	})
	if err == nil {
		t.Error("Expected error for invalid range length")
	}
}