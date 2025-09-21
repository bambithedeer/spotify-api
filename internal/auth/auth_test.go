package auth

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test_id", "test_secret", "http://localhost:8080/callback")

	if client.ClientID != "test_id" {
		t.Errorf("Expected ClientID 'test_id', got %s", client.ClientID)
	}

	if client.ClientSecret != "test_secret" {
		t.Errorf("Expected ClientSecret 'test_secret', got %s", client.ClientSecret)
	}

	if client.RedirectURI != "http://localhost:8080/callback" {
		t.Errorf("Expected RedirectURI 'http://localhost:8080/callback', got %s", client.RedirectURI)
	}
}

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name     string
		expiry   time.Time
		expected bool
	}{
		{"not expired", time.Now().Add(time.Hour), false},
		{"expired", time.Now().Add(-time.Hour), true},
		{"just expired", time.Now().Add(-time.Second), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &Token{Expiry: tt.expiry}
			if token.IsExpired() != tt.expected {
				t.Errorf("Expected IsExpired() to be %v", tt.expected)
			}
		})
	}
}

func TestGetAuthorizationURL(t *testing.T) {
	client := NewClient("test_id", "test_secret", "http://localhost:8080/callback")
	scopes := []string{"user-read-private", "user-read-email"}
	state := "test_state"

	url := client.GetAuthorizationURL(scopes, state)

	expected := "https://accounts.spotify.com/authorize?client_id=test_id&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&response_type=code&scope=user-read-private+user-read-email&state=test_state"

	if url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}
}

func TestClientCredentials(t *testing.T) {
	// This would be an integration test requiring real Spotify credentials
	// For unit testing, we would need to refactor to allow dependency injection of HTTP client
	t.Skip("Integration test - requires real Spotify API credentials")
}

func TestBasicAuth(t *testing.T) {
	client := NewClient("test_id", "test_secret", "http://localhost:8080/callback")
	auth := client.basicAuth()

	// test_id:test_secret base64 encoded should be dGVzdF9pZDp0ZXN0X3NlY3JldA==
	expected := "dGVzdF9pZDp0ZXN0X3NlY3JldA=="
	if auth != expected {
		t.Errorf("Expected basic auth %s, got %s", expected, auth)
	}
}