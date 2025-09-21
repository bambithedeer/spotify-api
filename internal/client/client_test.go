package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bambithedeer/spotify-api/internal/auth"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test_id", "test_secret", "http://localhost:8080/callback")

	if client.baseURL != SpotifyAPIBaseURL {
		t.Errorf("Expected baseURL %s, got %s", SpotifyAPIBaseURL, client.baseURL)
	}

	if client.authClient == nil {
		t.Error("Expected authClient to be initialized")
	}
}

func TestSetToken(t *testing.T) {
	client := NewClient("test_id", "test_secret", "http://localhost:8080/callback")
	token := &auth.Token{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}

	client.SetToken(token)

	if client.GetToken() != token {
		t.Error("Expected token to be set")
	}
}

func TestRefreshTokenIfNeeded(t *testing.T) {
	client := NewClient("test_id", "test_secret", "http://localhost:8080/callback")

	// Test with no token
	err := client.RefreshTokenIfNeeded()
	if err == nil {
		t.Error("Expected error when no token is set")
	}

	// Test with valid token
	validToken := &auth.Token{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	client.SetToken(validToken)

	err = client.RefreshTokenIfNeeded()
	if err != nil {
		t.Errorf("Expected no error with valid token, got %v", err)
	}

	// Test with expired token but no refresh token
	expiredToken := &auth.Token{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(-time.Hour),
	}
	client.SetToken(expiredToken)

	err = client.RefreshTokenIfNeeded()
	if err == nil {
		t.Error("Expected error with expired token and no refresh token")
	}
}

func TestMakeRequest(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test_token" {
			t.Errorf("Expected Authorization header 'Bearer test_token', got %s", authHeader)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": "response"}`))
	}))
	defer server.Close()

	client := NewClient("test_id", "test_secret", "http://localhost:8080/callback")
	client.baseURL = server.URL // Use test server

	// Set a valid token
	token := &auth.Token{
		AccessToken: "test_token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
	}
	client.SetToken(token)

	resp, err := client.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMakeRequestWithoutToken(t *testing.T) {
	client := NewClient("test_id", "test_secret", "http://localhost:8080/callback")

	_, err := client.Get(context.Background(), "/test")
	if err == nil {
		t.Error("Expected error when making request without token")
	}
}

func TestGetAuthorizationURL(t *testing.T) {
	client := NewClient("test_id", "test_secret", "http://localhost:8080/callback")
	scopes := []string{"user-read-private", "user-read-email"}
	state := "test_state"

	url := client.GetAuthorizationURL(scopes, state)

	if url == "" {
		t.Error("Expected authorization URL to be returned")
	}
}