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

// Mock player responses
var mockPlaybackStateResponse = `{
	"device": {
		"id": "ed01a3cf-6a7b-4d16-8b32-ed8bb8e93345",
		"is_active": true,
		"is_private_session": false,
		"is_restricted": false,
		"name": "Test Speaker",
		"type": "Computer",
		"volume_percent": 75,
		"supports_volume": true
	},
	"repeat_state": "off",
	"shuffle_state": false,
	"context": {
		"type": "playlist",
		"href": "https://api.spotify.com/v1/playlists/37i9dQZF1DX0XUsuxWHRQd",
		"external_urls": {
			"spotify": "https://open.spotify.com/playlist/37i9dQZF1DX0XUsuxWHRQd"
		},
		"uri": "spotify:playlist:37i9dQZF1DX0XUsuxWHRQd"
	},
	"timestamp": 1234567890123,
	"progress_ms": 45000,
	"is_playing": true,
	"item": {
		"id": "6iV5W9uYEdYUVa79Axb7Rh",
		"name": "Test Track",
		"type": "track",
		"uri": "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
		"duration_ms": 180000
	},
	"currently_playing_type": "track",
	"actions": {
		"interrupting_playback": false,
		"pausing": true,
		"resuming": true,
		"seeking": true,
		"skipping_next": true,
		"skipping_prev": true,
		"toggling_repeat_context": true,
		"toggling_repeat_track": true,
		"toggling_shuffle": true,
		"transferring_playback": true
	}
}`

var mockCurrentlyPlayingResponse = `{
	"context": {
		"type": "playlist",
		"href": "https://api.spotify.com/v1/playlists/37i9dQZF1DX0XUsuxWHRQd",
		"external_urls": {
			"spotify": "https://open.spotify.com/playlist/37i9dQZF1DX0XUsuxWHRQd"
		},
		"uri": "spotify:playlist:37i9dQZF1DX0XUsuxWHRQd"
	},
	"timestamp": 1234567890123,
	"progress_ms": 60000,
	"is_playing": true,
	"item": {
		"id": "7iV5W9uYEdYUVa79Axb7Rh",
		"name": "Currently Playing Track",
		"type": "track",
		"uri": "spotify:track:7iV5W9uYEdYUVa79Axb7Rh",
		"duration_ms": 210000
	},
	"currently_playing_type": "track",
	"actions": {
		"interrupting_playback": false,
		"pausing": true,
		"resuming": false,
		"seeking": true,
		"skipping_next": true,
		"skipping_prev": true,
		"toggling_repeat_context": true,
		"toggling_repeat_track": true,
		"toggling_shuffle": true,
		"transferring_playback": true
	}
}`

var mockDevicesResponse = `{
	"devices": [{
		"id": "ed01a3cf-6a7b-4d16-8b32-ed8bb8e93345",
		"is_active": true,
		"is_private_session": false,
		"is_restricted": false,
		"name": "Test Speaker",
		"type": "Computer",
		"volume_percent": 75,
		"supports_volume": true
	}, {
		"id": "2c01a3cf-6a7b-4d16-8b32-ed8bb8e93346",
		"is_active": false,
		"is_private_session": false,
		"is_restricted": false,
		"name": "Test Phone",
		"type": "Smartphone",
		"volume_percent": 50,
		"supports_volume": true
	}]
}`

var mockRecentlyPlayedResponse = `{
	"href": "https://api.spotify.com/v1/me/player/recently-played",
	"items": [{
		"track": {
			"id": "8iV5W9uYEdYUVa79Axb7Rh",
			"name": "Recently Played Track",
			"type": "track",
			"uri": "spotify:track:8iV5W9uYEdYUVa79Axb7Rh",
			"duration_ms": 195000
		},
		"played_at": "2023-01-01T12:00:00Z",
		"context": {
			"type": "playlist",
			"href": "https://api.spotify.com/v1/playlists/37i9dQZF1DX0XUsuxWHRQd",
			"external_urls": {
				"spotify": "https://open.spotify.com/playlist/37i9dQZF1DX0XUsuxWHRQd"
			},
			"uri": "spotify:playlist:37i9dQZF1DX0XUsuxWHRQd"
		}
	}],
	"next": null,
	"cursors": {
		"after": "1672574400000",
		"before": "1672574400000"
	},
	"limit": 20,
	"total": 1
}`

func createTestPlayerService() (*PlayerService, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"status": 401, "message": "Unauthorized"}}`))
			return
		}

		switch {
		case r.URL.Path == "/me/player" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockPlaybackStateResponse))
		case r.URL.Path == "/me/player/currently-playing" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockCurrentlyPlayingResponse))
		case r.URL.Path == "/me/player/devices" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockDevicesResponse))
		case r.URL.Path == "/me/player/play" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/player/pause" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/player/next" && r.Method == "POST":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/player/previous" && r.Method == "POST":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/player/seek" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/player/repeat" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/player/shuffle" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/player/volume" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/player" && r.Method == "PUT":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/player/queue" && r.Method == "POST":
			w.WriteHeader(http.StatusOK)
		case r.URL.Path == "/me/player/recently-played" && r.Method == "GET":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockRecentlyPlayedResponse))
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": {"status": 404, "message": "Not Found"}}`))
		}
	}))

	// Create a client with minimal setup for testing
	spotifyClient := &client.Client{}
	requestBuilder := api.NewRequestBuilder(spotifyClient)

	service := NewPlayerService(requestBuilder)
	return service, server
}

func TestPlayerService_ValidationErrors(t *testing.T) {
	// Create a minimal RequestBuilder for validation testing
	client := &client.Client{}
	requestBuilder := api.NewRequestBuilder(client)
	service := NewPlayerService(requestBuilder)

	// Test negative position for Seek
	err := service.Seek(context.Background(), -1, "")
	if err == nil {
		t.Error("Expected error for negative position in Seek")
	}

	// Test invalid repeat state
	err = service.SetRepeat(context.Background(), "invalid", "")
	if err == nil {
		t.Error("Expected error for invalid repeat state")
	}

	// Test volume below 0
	err = service.SetVolume(context.Background(), -1, "")
	if err == nil {
		t.Error("Expected error for volume below 0")
	}

	// Test volume above 100
	err = service.SetVolume(context.Background(), 101, "")
	if err == nil {
		t.Error("Expected error for volume above 100")
	}

	// Test empty URI for AddToQueue
	err = service.AddToQueue(context.Background(), "", "")
	if err == nil {
		t.Error("Expected error for empty URI in AddToQueue")
	}

	// Test nil TransferPlaybackRequest
	err = service.TransferPlayback(context.Background(), nil)
	if err == nil {
		t.Error("Expected error for nil TransferPlaybackRequest")
	}

	// Test empty device IDs in TransferPlaybackRequest
	err = service.TransferPlayback(context.Background(), &TransferPlaybackRequest{
		DeviceIDs: []string{},
	})
	if err == nil {
		t.Error("Expected error for empty device IDs in TransferPlaybackRequest")
	}

	// Test too many device IDs in TransferPlaybackRequest
	err = service.TransferPlayback(context.Background(), &TransferPlaybackRequest{
		DeviceIDs: []string{"device1", "device2"},
	})
	if err == nil {
		t.Error("Expected error for too many device IDs in TransferPlaybackRequest")
	}

	// Test invalid additional types
	options := &CurrentlyPlayingOptions{
		AdditionalTypes: []string{"invalid_type"},
	}
	_, err = service.GetCurrentlyPlaying(context.Background(), options)
	if err == nil {
		t.Error("Expected error for invalid additional type")
	}

	// Test invalid market
	options = &CurrentlyPlayingOptions{
		Market: "INVALID_MARKET",
	}
	_, err = service.GetCurrentlyPlaying(context.Background(), options)
	if err == nil {
		t.Error("Expected error for invalid market")
	}

	// Test invalid limit for GetRecentlyPlayed
	recentOptions := &RecentlyPlayedOptions{
		Limit: 100, // exceeds max limit of 50
	}
	_, err = service.GetRecentlyPlayed(context.Background(), recentOptions)
	if err == nil {
		t.Error("Expected error for limit exceeding maximum in GetRecentlyPlayed")
	}
}

func TestPlayerService_RepeatStateValidation(t *testing.T) {
	// Create a minimal RequestBuilder for validation testing
	client := &client.Client{}
	requestBuilder := api.NewRequestBuilder(client)
	service := NewPlayerService(requestBuilder)

	// Test valid repeat states
	validStates := []string{"track", "context", "off"}

	for _, state := range validStates {
		// This will fail with network error, but we just want to test validation passes
		err := service.SetRepeat(context.Background(), state, "")
		if err != nil && strings.Contains(err.Error(), "invalid repeat state") {
			t.Errorf("Expected network error for valid repeat state '%s', got validation error: %v", state, err)
		}
	}

	// Test invalid repeat state
	err := service.SetRepeat(context.Background(), "invalid_state", "")
	if err == nil || !strings.Contains(err.Error(), "invalid repeat state") {
		t.Error("Expected validation error for invalid repeat state")
	}
}

func TestPlayerService_AdditionalTypesValidation(t *testing.T) {
	// Create a minimal RequestBuilder for validation testing
	client := &client.Client{}
	requestBuilder := api.NewRequestBuilder(client)
	service := NewPlayerService(requestBuilder)

	// Test valid additional types
	validOptions := &CurrentlyPlayingOptions{
		AdditionalTypes: []string{"track", "episode"},
	}

	// This will fail with network error, but we just want to test validation passes
	_, err := service.GetCurrentlyPlaying(context.Background(), validOptions)
	if err != nil && strings.Contains(err.Error(), "invalid additional type") {
		t.Errorf("Expected network error for valid additional types, got validation error: %v", err)
	}

	// Test invalid additional type
	invalidOptions := &CurrentlyPlayingOptions{
		AdditionalTypes: []string{"invalid_type"},
	}

	_, err = service.GetCurrentlyPlaying(context.Background(), invalidOptions)
	if err == nil || !strings.Contains(err.Error(), "invalid additional type") {
		t.Error("Expected validation error for invalid additional type")
	}
}

func TestPlayerService_VolumeValidation(t *testing.T) {
	// Create a minimal RequestBuilder for validation testing
	client := &client.Client{}
	requestBuilder := api.NewRequestBuilder(client)
	service := NewPlayerService(requestBuilder)

	// Test valid volume levels
	validVolumes := []int{0, 50, 100}

	for _, volume := range validVolumes {
		// This will fail with network error, but we just want to test validation passes
		err := service.SetVolume(context.Background(), volume, "")
		if err != nil && strings.Contains(err.Error(), "volume must be between 0 and 100") {
			t.Errorf("Expected network error for valid volume %d, got validation error: %v", volume, err)
		}
	}

	// Test invalid volume levels
	invalidVolumes := []int{-1, 101}

	for _, volume := range invalidVolumes {
		err := service.SetVolume(context.Background(), volume, "")
		if err == nil || !strings.Contains(err.Error(), "volume must be between 0 and 100") {
			t.Errorf("Expected validation error for invalid volume %d", volume)
		}
	}
}