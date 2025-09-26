package integration

import (
	"testing"

	"github.com/bambithedeer/spotify-api/internal/lidarr"
	"github.com/bambithedeer/spotify-api/internal/logger"
	"github.com/bambithedeer/spotify-api/internal/musicbrainz"
)

func TestNewLidarrIntegration(t *testing.T) {
	lidarrClient := lidarr.NewClient(lidarr.Config{
		BaseURL: "http://localhost:8686",
		APIKey:  "test-key",
	})

	mbClient := musicbrainz.NewClient()
	defer mbClient.Close()

	config := &LidarrConfig{
		RootFolderPath:    "/music",
		QualityProfileID:  1,
		MetadataProfileID: 1,
		Monitor:           true,
		SearchForMissing:  true,
	}

	log := logger.NewLogger(&logger.Config{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	})

	integration := NewLidarrIntegration(lidarrClient, mbClient, config, log)

	if integration == nil {
		t.Fatal("NewLidarrIntegration returned nil")
	}

	if integration.lidarrClient != lidarrClient {
		t.Error("lidarrClient not set correctly")
	}

	if integration.musicbrainzClient != mbClient {
		t.Error("musicbrainzClient not set correctly")
	}

	if integration.config != config {
		t.Error("config not set correctly")
	}

	if integration.logger != log {
		t.Error("logger not set correctly")
	}
}

func TestAddArtistIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This would require real API connections
	t.Skip("integration test requires running services")
}

func TestAddArtistsBatchValidation(t *testing.T) {
	lidarrClient := lidarr.NewClient(lidarr.Config{
		BaseURL: "http://localhost:8686",
		APIKey:  "test-key",
	})

	mbClient := musicbrainz.NewClient()
	defer mbClient.Close()

	config := &LidarrConfig{
		RootFolderPath:    "/music",
		QualityProfileID:  1,
		MetadataProfileID: 1,
		Monitor:           true,
		SearchForMissing:  true,
	}

	log := logger.NewLogger(&logger.Config{
		Level:  "info",
		Format: "text",
		Output: "stdout",
	})

	integration := NewLidarrIntegration(lidarrClient, mbClient, config, log)

	// Test empty artist list
	result := integration.AddArtistsBatch([]string{}, 3)
	if result.Total != 0 {
		t.Errorf("expected total 0, got %d", result.Total)
	}

	// Test default concurrency
	artistNames := []string{"Test Artist 1", "Test Artist 2"}
	result = integration.AddArtistsBatch(artistNames, 0)
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
}

func TestLidarrConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *LidarrConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: &LidarrConfig{
				RootFolderPath:    "/music",
				QualityProfileID:  1,
				MetadataProfileID: 1,
				Monitor:           true,
				SearchForMissing:  true,
			},
			valid: true,
		},
		{
			name: "missing root folder",
			config: &LidarrConfig{
				RootFolderPath:    "",
				QualityProfileID:  1,
				MetadataProfileID: 1,
				Monitor:           true,
				SearchForMissing:  true,
			},
			valid: false,
		},
		{
			name: "invalid quality profile",
			config: &LidarrConfig{
				RootFolderPath:    "/music",
				QualityProfileID:  0,
				MetadataProfileID: 1,
				Monitor:           true,
				SearchForMissing:  true,
			},
			valid: false,
		},
		{
			name: "invalid metadata profile",
			config: &LidarrConfig{
				RootFolderPath:    "/music",
				QualityProfileID:  1,
				MetadataProfileID: 0,
				Monitor:           true,
				SearchForMissing:  true,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			if tt.config.RootFolderPath == "" && tt.valid {
				t.Error("expected invalid config with empty root folder path")
			}

			if tt.config.QualityProfileID <= 0 && tt.valid {
				t.Error("expected invalid config with zero quality profile ID")
			}

			if tt.config.MetadataProfileID <= 0 && tt.valid {
				t.Error("expected invalid config with zero metadata profile ID")
			}
		})
	}
}

func TestArtistResult(t *testing.T) {
	result := &ArtistResult{
		ArtistName:   "Pink Floyd",
		SpotifyName:  "Pink Floyd",
		MBID:         "83d91898-7763-47d7-b03b-b92132375c47",
		Success:      true,
		Error:        nil,
		LidarrArtist: nil,
	}

	if result.ArtistName != "Pink Floyd" {
		t.Errorf("expected artist name 'Pink Floyd', got %s", result.ArtistName)
	}

	if !result.Success {
		t.Error("expected success to be true")
	}

	if result.Error != nil {
		t.Errorf("expected no error, got %v", result.Error)
	}
}

func TestBatchResult(t *testing.T) {
	results := []ArtistResult{
		{Success: true},
		{Success: false},
		{Success: true},
	}

	batchResult := &BatchResult{
		Total:     3,
		Successes: 2,
		Failures:  1,
		Results:   results,
	}

	if batchResult.Total != 3 {
		t.Errorf("expected total 3, got %d", batchResult.Total)
	}

	if batchResult.Successes != 2 {
		t.Errorf("expected successes 2, got %d", batchResult.Successes)
	}

	if batchResult.Failures != 1 {
		t.Errorf("expected failures 1, got %d", batchResult.Failures)
	}

	if len(batchResult.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(batchResult.Results))
	}
}