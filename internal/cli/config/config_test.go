package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Default(t *testing.T) {
	config := Default()

	if config.RedirectURI != "http://127.0.0.1:4000" {
		t.Errorf("Expected default redirect URI, got %s", config.RedirectURI)
	}

	if config.DefaultOutput != "text" {
		t.Errorf("Expected default output format 'text', got %s", config.DefaultOutput)
	}

	if !config.ColorOutput {
		t.Error("Expected color output to be enabled by default")
	}

	if !config.CacheEnabled {
		t.Error("Expected cache to be enabled by default")
	}

	if config.CacheTTL != "1h" {
		t.Errorf("Expected default cache TTL '1h', got %s", config.CacheTTL)
	}
}

func TestConfig_SetCredentials(t *testing.T) {
	// Reset current config
	current = nil

	SetCredentials("test_client_id", "test_client_secret", "http://localhost:3000/callback")

	config := Get()
	if config.ClientID != "test_client_id" {
		t.Errorf("Expected client ID 'test_client_id', got %s", config.ClientID)
	}

	if config.ClientSecret != "test_client_secret" {
		t.Errorf("Expected client secret 'test_client_secret', got %s", config.ClientSecret)
	}

	if config.RedirectURI != "http://localhost:3000/callback" {
		t.Errorf("Expected redirect URI 'http://localhost:3000/callback', got %s", config.RedirectURI)
	}
}

func TestConfig_SetTokens(t *testing.T) {
	// Reset current config
	current = nil

	SetTokens("access_token", "refresh_token", "Bearer", "2023-12-31T23:59:59Z")

	config := Get()
	if config.AccessToken != "access_token" {
		t.Errorf("Expected access token 'access_token', got %s", config.AccessToken)
	}

	if config.RefreshToken != "refresh_token" {
		t.Errorf("Expected refresh token 'refresh_token', got %s", config.RefreshToken)
	}

	if config.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got %s", config.TokenType)
	}

	if config.ExpiresAt != "2023-12-31T23:59:59Z" {
		t.Errorf("Expected expires at '2023-12-31T23:59:59Z', got %s", config.ExpiresAt)
	}
}

func TestConfig_HasCredentials(t *testing.T) {
	// Reset current config
	current = nil

	// Test without credentials
	if HasCredentials() {
		t.Error("Expected HasCredentials to return false when no credentials are set")
	}

	// Set credentials
	SetCredentials("client_id", "client_secret", "")

	if !HasCredentials() {
		t.Error("Expected HasCredentials to return true when credentials are set")
	}
}

func TestConfig_IsAuthenticated(t *testing.T) {
	// Reset current config
	current = nil

	// Test without authentication
	if IsAuthenticated() {
		t.Error("Expected IsAuthenticated to return false when not authenticated")
	}

	// Set tokens
	SetTokens("access_token", "refresh_token", "Bearer", "")

	if !IsAuthenticated() {
		t.Error("Expected IsAuthenticated to return true when tokens are set")
	}
}

func TestConfig_ClearTokens(t *testing.T) {
	// Reset current config
	current = nil

	// Set tokens first
	SetTokens("access_token", "refresh_token", "Bearer", "2023-12-31T23:59:59Z")

	// Verify tokens are set
	if !IsAuthenticated() {
		t.Error("Expected to be authenticated before clearing tokens")
	}

	// Clear tokens
	ClearTokens()

	// Verify tokens are cleared
	config := Get()
	if config.AccessToken != "" {
		t.Error("Expected access token to be cleared")
	}

	if config.RefreshToken != "" {
		t.Error("Expected refresh token to be cleared")
	}

	if config.TokenType != "" {
		t.Error("Expected token type to be cleared")
	}

	if config.ExpiresAt != "" {
		t.Error("Expected expires at to be cleared")
	}

	if IsAuthenticated() {
		t.Error("Expected IsAuthenticated to return false after clearing tokens")
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_config.yaml")

	// Reset current config
	current = nil

	// Initialize with test file
	if err := Init(tmpFile, false, "text"); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Set some test data
	SetCredentials("test_client_id", "test_client_secret", "http://127.0.0.1:4000")
	SetTokens("test_access_token", "test_refresh_token", "Bearer", "2023-12-31T23:59:59Z")

	// Save config
	if err := Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Reset current config
	current = nil

	// Initialize again (should load from file)
	if err := Init(tmpFile, false, "text"); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	// Verify loaded data
	config := Get()
	if config.ClientID != "test_client_id" {
		t.Errorf("Expected loaded client ID 'test_client_id', got %s", config.ClientID)
	}

	if config.AccessToken != "test_access_token" {
		t.Errorf("Expected loaded access token 'test_access_token', got %s", config.AccessToken)
	}

	if !IsAuthenticated() {
		t.Error("Expected to be authenticated after loading config")
	}

	if !HasCredentials() {
		t.Error("Expected to have credentials after loading config")
	}
}