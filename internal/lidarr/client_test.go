package lidarr

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	config := Config{
		BaseURL: "http://localhost:8686",
		APIKey:  "test-api-key",
		Timeout: 10 * time.Second,
	}

	client := NewClient(config)
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.baseURL != config.BaseURL {
		t.Errorf("expected baseURL %s, got %s", config.BaseURL, client.baseURL)
	}

	if client.apiKey != config.APIKey {
		t.Errorf("expected apiKey %s, got %s", config.APIKey, client.apiKey)
	}

	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}

	if client.httpClient.Timeout != config.Timeout {
		t.Errorf("expected timeout %v, got %v", config.Timeout, client.httpClient.Timeout)
	}
}

func TestNewClientDefaultTimeout(t *testing.T) {
	config := Config{
		BaseURL: "http://localhost:8686",
		APIKey:  "test-api-key",
		// Timeout not set
	}

	client := NewClient(config)
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	expectedTimeout := 30 * time.Second
	if client.httpClient.Timeout != expectedTimeout {
		t.Errorf("expected default timeout %v, got %v", expectedTimeout, client.httpClient.Timeout)
	}
}

func TestNewClientTrimsBaseURL(t *testing.T) {
	config := Config{
		BaseURL: "http://localhost:8686/",
		APIKey:  "test-api-key",
	}

	client := NewClient(config)
	expected := "http://localhost:8686"
	if client.baseURL != expected {
		t.Errorf("expected baseURL %s, got %s", expected, client.baseURL)
	}
}

// Mock tests for API functionality (integration tests would require running Lidarr instance)

func TestAddArtistByMBIDValidation(t *testing.T) {
	config := Config{
		BaseURL: "http://localhost:8686",
		APIKey:  "test-api-key",
	}

	client := NewClient(config)

	// Test with empty MBID
	_, err := client.AddArtistByMBID("", "/music", 1, 1, true, false)
	if err == nil {
		t.Error("expected error for empty MBID")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected string
	}{
		{
			name: "trailing slash removal",
			config: Config{
				BaseURL: "http://localhost:8686/",
				APIKey:  "test-key",
			},
			expected: "http://localhost:8686",
		},
		{
			name: "no trailing slash",
			config: Config{
				BaseURL: "http://localhost:8686",
				APIKey:  "test-key",
			},
			expected: "http://localhost:8686",
		},
		{
			name: "multiple trailing slashes",
			config: Config{
				BaseURL: "http://localhost:8686///",
				APIKey:  "test-key",
			},
			expected: "http://localhost:8686//",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			if client.baseURL != tt.expected {
				t.Errorf("expected baseURL %s, got %s", tt.expected, client.baseURL)
			}
		})
	}
}

// Integration tests (require running Lidarr instance)
// These tests are skipped in short mode

func TestSearchArtistIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This test requires a running Lidarr instance
	// You would need to set up environment variables or test configuration
	t.Skip("integration test requires running Lidarr instance")
}

func TestTestConnectionIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This test requires a running Lidarr instance
	t.Skip("integration test requires running Lidarr instance")
}

func TestGetRootFoldersIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This test requires a running Lidarr instance
	t.Skip("integration test requires running Lidarr instance")
}

func TestGetQualityProfilesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This test requires a running Lidarr instance
	t.Skip("integration test requires running Lidarr instance")
}

func TestGetMetadataProfilesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// This test requires a running Lidarr instance
	t.Skip("integration test requires running Lidarr instance")
}