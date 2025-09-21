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

// Mock track responses
var mockTrackResponse = `{
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
	"duration_ms": 180000,
	"explicit": false,
	"popularity": 75,
	"external_urls": {"spotify": "https://open.spotify.com/track/6iV5W9uYEdYUVa79Axb7Rh"},
	"href": "https://api.spotify.com/v1/tracks/6iV5W9uYEdYUVa79Axb7Rh",
	"type": "track",
	"uri": "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
	"available_markets": ["US", "CA"],
	"disc_number": 1,
	"track_number": 1,
	"is_local": false,
	"preview_url": "https://example.com/preview.mp3",
	"external_ids": {"isrc": "TEST123456789"}
}`

var mockMultipleTracksResponse = `{
	"tracks": [{
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
	}, {
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
	}]
}`

var mockAudioFeaturesResponse = `{
	"acousticness": 0.00242,
	"analysis_url": "https://api.spotify.com/v1/audio-analysis/6iV5W9uYEdYUVa79Axb7Rh",
	"danceability": 0.585,
	"duration_ms": 180000,
	"energy": 0.842,
	"id": "6iV5W9uYEdYUVa79Axb7Rh",
	"instrumentalness": 0.00686,
	"key": 9,
	"liveness": 0.0866,
	"loudness": -5.883,
	"mode": 0,
	"speechiness": 0.0556,
	"tempo": 118.211,
	"time_signature": 4,
	"track_href": "https://api.spotify.com/v1/tracks/6iV5W9uYEdYUVa79Axb7Rh",
	"type": "audio_features",
	"uri": "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
	"valence": 0.428
}`

var mockMultipleAudioFeaturesResponse = `{
	"audio_features": [{
		"acousticness": 0.00242,
		"analysis_url": "https://api.spotify.com/v1/audio-analysis/6iV5W9uYEdYUVa79Axb7Rh",
		"danceability": 0.585,
		"duration_ms": 180000,
		"energy": 0.842,
		"id": "6iV5W9uYEdYUVa79Axb7Rh",
		"instrumentalness": 0.00686,
		"key": 9,
		"liveness": 0.0866,
		"loudness": -5.883,
		"mode": 0,
		"speechiness": 0.0556,
		"tempo": 118.211,
		"time_signature": 4,
		"track_href": "https://api.spotify.com/v1/tracks/6iV5W9uYEdYUVa79Axb7Rh",
		"type": "audio_features",
		"uri": "spotify:track:6iV5W9uYEdYUVa79Axb7Rh",
		"valence": 0.428
	}, {
		"acousticness": 0.1,
		"analysis_url": "https://api.spotify.com/v1/audio-analysis/7iV5W9uYEdYUVa79Axb7Rh",
		"danceability": 0.7,
		"duration_ms": 210000,
		"energy": 0.9,
		"id": "7iV5W9uYEdYUVa79Axb7Rh",
		"instrumentalness": 0.0,
		"key": 5,
		"liveness": 0.1,
		"loudness": -4.0,
		"mode": 1,
		"speechiness": 0.05,
		"tempo": 125.0,
		"time_signature": 4,
		"track_href": "https://api.spotify.com/v1/tracks/7iV5W9uYEdYUVa79Axb7Rh",
		"type": "audio_features",
		"uri": "spotify:track:7iV5W9uYEdYUVa79Axb7Rh",
		"valence": 0.8
	}]
}`

var mockAudioAnalysisResponse = `{
	"meta": {
		"analyzer_version": "4.0.0",
		"platform": "Linux",
		"detailed_status": "OK",
		"status_code": 0,
		"timestamp": 1495193577,
		"analysis_time": 6.93906,
		"input_total_time": 255.349
	},
	"track": {
		"num_samples": 4585515,
		"duration": 255.349,
		"sample_md5": "string",
		"offset_seconds": 0,
		"window_seconds": 0,
		"analysis_sample_rate": 22050,
		"analysis_channels": 1,
		"end_of_fade_in": 0,
		"start_of_fade_out": 250.0,
		"loudness": -5.883,
		"tempo": 118.211,
		"tempo_confidence": 0.73,
		"time_signature": 4,
		"time_signature_confidence": 0.994,
		"key": 9,
		"key_confidence": 0.408,
		"mode": 0,
		"mode_confidence": 0.485
	},
	"bars": [],
	"beats": [],
	"sections": [],
	"segments": [],
	"tatums": []
}`

var mockRecommendationsResponse = `{
	"tracks": [{
		"id": "8iV5W9uYEdYUVa79Axb7Rh",
		"name": "Recommended Track",
		"artists": [{"id": "1301WleyT98MSxVHPZCA6M", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:1301WleyT98MSxVHPZCA6M", "href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M", "external_urls": {"spotify": "https://open.spotify.com/artist/1301WleyT98MSxVHPZCA6M"}}],
		"duration_ms": 190000,
		"explicit": false,
		"popularity": 78,
		"external_urls": {"spotify": "https://open.spotify.com/track/8iV5W9uYEdYUVa79Axb7Rh"},
		"href": "https://api.spotify.com/v1/tracks/8iV5W9uYEdYUVa79Axb7Rh",
		"type": "track",
		"uri": "spotify:track:8iV5W9uYEdYUVa79Axb7Rh",
		"available_markets": ["US"],
		"disc_number": 1,
		"track_number": 1,
		"is_local": false,
		"preview_url": null,
		"external_ids": {}
	}],
	"seeds": [{
		"initialPoolSize": 500,
		"afterFilteringSize": 380,
		"afterRelinkingSize": 365,
		"id": "1301WleyT98MSxVHPZCA6M",
		"type": "ARTIST",
		"href": "https://api.spotify.com/v1/artists/1301WleyT98MSxVHPZCA6M"
	}]
}`

func createTestTracksService() (*TracksService, *httptest.Server) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test_token" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": {"status": 401, "message": "Unauthorized"}}`))
			return
		}

		switch {
		case r.URL.Path == "/tracks/6iV5W9uYEdYUVa79Axb7Rh":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockTrackResponse))
		case r.URL.Path == "/tracks" && strings.Contains(r.URL.RawQuery, "ids="):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockMultipleTracksResponse))
		case r.URL.Path == "/audio-features/6iV5W9uYEdYUVa79Axb7Rh":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockAudioFeaturesResponse))
		case r.URL.Path == "/audio-features" && strings.Contains(r.URL.RawQuery, "ids="):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockMultipleAudioFeaturesResponse))
		case r.URL.Path == "/audio-analysis/6iV5W9uYEdYUVa79Axb7Rh":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockAudioAnalysisResponse))
		case r.URL.Path == "/recommendations":
			// Check for required seed parameters
			query := r.URL.Query()
			if query.Get("seed_artists") == "" && query.Get("seed_genres") == "" && query.Get("seed_tracks") == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error": {"status": 400, "message": "At least one seed required"}}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(mockRecommendationsResponse))
		case strings.HasPrefix(r.URL.Path, "/tracks/invalid"):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": {"status": 404, "message": "Track not found"}}`))
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

	// Create request builder and tracks service
	builder := api.NewRequestBuilder(client)
	service := NewTracksService(builder)

	return service, server
}

func TestTracksService_GetTrack(t *testing.T) {
	service, server := createTestTracksService()
	defer server.Close()

	ctx := context.Background()

	// Test successful track retrieval
	track, err := service.GetTrack(ctx, "6iV5W9uYEdYUVa79Axb7Rh", "")
	if err != nil {
		t.Fatalf("GetTrack failed: %v", err)
	}

	if track == nil {
		t.Fatal("Expected track result, got nil")
	}

	if track.ID != "6iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected track ID '6iV5W9uYEdYUVa79Axb7Rh', got %s", track.ID)
	}

	if track.Name != "Test Track" {
		t.Errorf("Expected track name 'Test Track', got %s", track.Name)
	}

	if track.DurationMs != 180000 {
		t.Errorf("Expected duration 180000ms, got %d", track.DurationMs)
	}

	if track.Popularity != 75 {
		t.Errorf("Expected popularity 75, got %d", track.Popularity)
	}

	if len(track.Artists) == 0 {
		t.Error("Expected track to have artists")
	}
}

func TestTracksService_GetTrackWithMarket(t *testing.T) {
	service, server := createTestTracksService()
	defer server.Close()

	ctx := context.Background()

	// Test track retrieval with market
	track, err := service.GetTrack(ctx, "6iV5W9uYEdYUVa79Axb7Rh", "US")
	if err != nil {
		t.Fatalf("GetTrack with market failed: %v", err)
	}

	if track == nil {
		t.Fatal("Expected track result, got nil")
	}

	if track.ID != "6iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected track ID '6iV5W9uYEdYUVa79Axb7Rh', got %s", track.ID)
	}
}

func TestTracksService_GetTracks(t *testing.T) {
	service, server := createTestTracksService()
	defer server.Close()

	ctx := context.Background()

	// Test multiple tracks retrieval
	trackIDs := []string{"6iV5W9uYEdYUVa79Axb7Rh", "7iV5W9uYEdYUVa79Axb7Rh"}
	tracks, err := service.GetTracks(ctx, trackIDs, "")
	if err != nil {
		t.Fatalf("GetTracks failed: %v", err)
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

	if tracks[0].Name != "Test Track 1" {
		t.Errorf("Expected first track name 'Test Track 1', got %s", tracks[0].Name)
	}

	if tracks[1].Name != "Test Track 2" {
		t.Errorf("Expected second track name 'Test Track 2', got %s", tracks[1].Name)
	}
}

func TestTracksService_GetTrackAudioFeatures(t *testing.T) {
	service, server := createTestTracksService()
	defer server.Close()

	ctx := context.Background()

	// Test audio features retrieval
	features, err := service.GetTrackAudioFeatures(ctx, "6iV5W9uYEdYUVa79Axb7Rh")
	if err != nil {
		t.Fatalf("GetTrackAudioFeatures failed: %v", err)
	}

	if features == nil {
		t.Fatal("Expected audio features result, got nil")
	}

	if features.ID != "6iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected features ID '6iV5W9uYEdYUVa79Axb7Rh', got %s", features.ID)
	}

	if features.Danceability != 0.585 {
		t.Errorf("Expected danceability 0.585, got %f", features.Danceability)
	}

	if features.Energy != 0.842 {
		t.Errorf("Expected energy 0.842, got %f", features.Energy)
	}

	if features.Tempo != 118.211 {
		t.Errorf("Expected tempo 118.211, got %f", features.Tempo)
	}
}

func TestTracksService_GetTracksAudioFeatures(t *testing.T) {
	service, server := createTestTracksService()
	defer server.Close()

	ctx := context.Background()

	// Test multiple audio features retrieval
	trackIDs := []string{"6iV5W9uYEdYUVa79Axb7Rh", "7iV5W9uYEdYUVa79Axb7Rh"}
	features, err := service.GetTracksAudioFeatures(ctx, trackIDs)
	if err != nil {
		t.Fatalf("GetTracksAudioFeatures failed: %v", err)
	}

	if len(features) != 2 {
		t.Errorf("Expected 2 audio features, got %d", len(features))
	}

	if features[0].ID != "6iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected first features ID '6iV5W9uYEdYUVa79Axb7Rh', got %s", features[0].ID)
	}

	if features[1].ID != "7iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected second features ID '7iV5W9uYEdYUVa79Axb7Rh', got %s", features[1].ID)
	}
}

func TestTracksService_GetTrackAudioAnalysis(t *testing.T) {
	service, server := createTestTracksService()
	defer server.Close()

	ctx := context.Background()

	// Test audio analysis retrieval
	analysis, err := service.GetTrackAudioAnalysis(ctx, "6iV5W9uYEdYUVa79Axb7Rh")
	if err != nil {
		t.Fatalf("GetTrackAudioAnalysis failed: %v", err)
	}

	if analysis == nil {
		t.Fatal("Expected audio analysis result, got nil")
	}

	if analysis.Track.Duration != 255.349 {
		t.Errorf("Expected track duration 255.349, got %f", analysis.Track.Duration)
	}

	if analysis.Track.Tempo != 118.211 {
		t.Errorf("Expected track tempo 118.211, got %f", analysis.Track.Tempo)
	}

	if analysis.Track.Key != 9 {
		t.Errorf("Expected track key 9, got %d", analysis.Track.Key)
	}
}

func TestTracksService_GetRecommendations(t *testing.T) {
	service, server := createTestTracksService()
	defer server.Close()

	ctx := context.Background()

	// Test recommendations with artist seed
	options := &RecommendationOptions{
		SeedArtists: []string{"1301WleyT98MSxVHPZCA6M"},
		Limit:       20,
		Market:      "US",
	}

	recommendations, err := service.GetRecommendations(ctx, options)
	if err != nil {
		t.Fatalf("GetRecommendations failed: %v", err)
	}

	if recommendations == nil {
		t.Fatal("Expected recommendations result, got nil")
	}

	if len(recommendations.Tracks) != 1 {
		t.Errorf("Expected 1 recommended track, got %d", len(recommendations.Tracks))
	}

	if len(recommendations.Seeds) != 1 {
		t.Errorf("Expected 1 seed, got %d", len(recommendations.Seeds))
	}

	if recommendations.Tracks[0].ID != "8iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected track ID '8iV5W9uYEdYUVa79Axb7Rh', got %s", recommendations.Tracks[0].ID)
	}
}

func TestTracksService_GetRecommendationsWithAudioFeatures(t *testing.T) {
	service, server := createTestTracksService()
	defer server.Close()

	ctx := context.Background()

	// Test recommendations with audio features tuning
	options := &RecommendationOptions{
		SeedGenres: []string{"rock", "pop"},
		Limit:      10,
		AudioFeatures: map[string]interface{}{
			"target_danceability": 0.7,
			"min_energy":          0.5,
			"target_valence":      0.8,
		},
	}

	recommendations, err := service.GetRecommendations(ctx, options)
	if err != nil {
		t.Fatalf("GetRecommendations with audio features failed: %v", err)
	}

	if recommendations == nil {
		t.Fatal("Expected recommendations result, got nil")
	}
}

func TestTracksService_ValidationErrors(t *testing.T) {
	service, server := createTestTracksService()
	defer server.Close()

	ctx := context.Background()

	// Test invalid track ID
	_, err := service.GetTrack(ctx, "invalid", "")
	if err == nil {
		t.Error("Expected error for invalid track ID")
	}

	// Test empty track IDs for multiple tracks
	_, err = service.GetTracks(ctx, []string{}, "")
	if err == nil {
		t.Error("Expected error for empty track IDs")
	}

	// Test too many track IDs
	tooManyIDs := make([]string, 51)
	for i := range tooManyIDs {
		tooManyIDs[i] = "6iV5W9uYEdYUVa79Axb7Rh"
	}
	_, err = service.GetTracks(ctx, tooManyIDs, "")
	if err == nil {
		t.Error("Expected error for too many track IDs")
	}

	// Test too many IDs for audio features
	tooManyFeaturesIDs := make([]string, 101)
	for i := range tooManyFeaturesIDs {
		tooManyFeaturesIDs[i] = "6iV5W9uYEdYUVa79Axb7Rh"
	}
	_, err = service.GetTracksAudioFeatures(ctx, tooManyFeaturesIDs)
	if err == nil {
		t.Error("Expected error for too many audio features IDs")
	}

	// Test invalid market
	_, err = service.GetTrack(ctx, "6iV5W9uYEdYUVa79Axb7Rh", "INVALID")
	if err == nil {
		t.Error("Expected error for invalid market")
	}
}

func TestTracksService_RecommendationValidationErrors(t *testing.T) {
	service, server := createTestTracksService()
	defer server.Close()

	ctx := context.Background()

	// Test nil options
	_, err := service.GetRecommendations(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil options")
	}

	// Test no seeds
	_, err = service.GetRecommendations(ctx, &RecommendationOptions{})
	if err == nil {
		t.Error("Expected error for no seeds")
	}

	// Test too many seeds
	_, err = service.GetRecommendations(ctx, &RecommendationOptions{
		SeedArtists: []string{"1301WleyT98MSxVHPZCA6M", "2301WleyT98MSxVHPZCA6M"},
		SeedGenres:  []string{"rock", "pop"},
		SeedTracks:  []string{"6iV5W9uYEdYUVa79Axb7Rh", "7iV5W9uYEdYUVa79Axb7Rh"},
	})
	if err == nil {
		t.Error("Expected error for too many seeds")
	}

	// Test invalid artist seed
	_, err = service.GetRecommendations(ctx, &RecommendationOptions{
		SeedArtists: []string{"invalid"},
	})
	if err == nil {
		t.Error("Expected error for invalid artist seed")
	}

	// Test invalid track seed
	_, err = service.GetRecommendations(ctx, &RecommendationOptions{
		SeedTracks: []string{"invalid"},
	})
	if err == nil {
		t.Error("Expected error for invalid track seed")
	}

	// Test empty genre seed
	_, err = service.GetRecommendations(ctx, &RecommendationOptions{
		SeedGenres: []string{""},
	})
	if err == nil {
		t.Error("Expected error for empty genre seed")
	}

	// Test invalid limit
	_, err = service.GetRecommendations(ctx, &RecommendationOptions{
		SeedGenres: []string{"rock"},
		Limit:      150, // Too high
	})
	if err == nil {
		t.Error("Expected error for invalid limit")
	}

	// Test invalid market
	_, err = service.GetRecommendations(ctx, &RecommendationOptions{
		SeedGenres: []string{"rock"},
		Market:     "INVALID",
	})
	if err == nil {
		t.Error("Expected error for invalid market")
	}
}

func TestTracksService_BuildRecommendationParams(t *testing.T) {
	service := &TracksService{
		validator: api.NewValidator(),
	}

	options := &RecommendationOptions{
		SeedArtists: []string{"1301WleyT98MSxVHPZCA6M"},
		SeedGenres:  []string{"rock", "pop"},
		SeedTracks:  []string{"6iV5W9uYEdYUVa79Axb7Rh"},
		Limit:       20,
		Market:      "US",
		AudioFeatures: map[string]interface{}{
			"target_danceability": 0.7,
			"min_energy":          0.5,
			"invalid_param":       "should_be_ignored",
		},
	}

	params := service.buildRecommendationParams(options)

	if params["seed_artists"] != "1301WleyT98MSxVHPZCA6M" {
		t.Errorf("Expected seed_artists '1301WleyT98MSxVHPZCA6M', got %v", params["seed_artists"])
	}

	if params["seed_genres"] != "rock,pop" {
		t.Errorf("Expected seed_genres 'rock,pop', got %v", params["seed_genres"])
	}

	if params["seed_tracks"] != "6iV5W9uYEdYUVa79Axb7Rh" {
		t.Errorf("Expected seed_tracks '6iV5W9uYEdYUVa79Axb7Rh', got %v", params["seed_tracks"])
	}

	if params["limit"] != 20 {
		t.Errorf("Expected limit 20, got %v", params["limit"])
	}

	if params["market"] != "US" {
		t.Errorf("Expected market 'US', got %v", params["market"])
	}

	if params["target_danceability"] != 0.7 {
		t.Errorf("Expected target_danceability 0.7, got %v", params["target_danceability"])
	}

	if params["min_energy"] != 0.5 {
		t.Errorf("Expected min_energy 0.5, got %v", params["min_energy"])
	}

	// Invalid param should not be included
	if _, exists := params["invalid_param"]; exists {
		t.Error("Expected invalid_param to be filtered out")
	}
}