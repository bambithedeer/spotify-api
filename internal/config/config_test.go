package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Spotify.RedirectURI != "http://localhost:8080/callback" {
		t.Errorf("Expected default redirect URI, got %s", config.Spotify.RedirectURI)
	}

	if len(config.Spotify.Scopes) == 0 {
		t.Error("Expected default scopes to be set")
	}

	if config.Logging.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", config.Logging.Level)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid config",
			config: &Config{
				Spotify: SpotifyConfig{
					ClientID:     "test_id",
					ClientSecret: "test_secret",
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			expectError: false,
		},
		{
			name: "missing client ID",
			config: &Config{
				Spotify: SpotifyConfig{
					ClientSecret: "test_secret",
				},
				Logging: LoggingConfig{
					Level:  "info",
					Format: "text",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if tt.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
			}
		})
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save original env vars
	originalClientID := os.Getenv("SPOTIFY_CLIENT_ID")
	originalClientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")

	defer func() {
		os.Setenv("SPOTIFY_CLIENT_ID", originalClientID)
		os.Setenv("SPOTIFY_CLIENT_SECRET", originalClientSecret)
	}()

	// Set test env vars
	os.Setenv("SPOTIFY_CLIENT_ID", "test_client_id")
	os.Setenv("SPOTIFY_CLIENT_SECRET", "test_client_secret")

	config := DefaultConfig()
	loadFromEnv(config)

	if config.Spotify.ClientID != "test_client_id" {
		t.Errorf("Expected client ID from env, got %s", config.Spotify.ClientID)
	}

	if config.Spotify.ClientSecret != "test_client_secret" {
		t.Errorf("Expected client secret from env, got %s", config.Spotify.ClientSecret)
	}
}