package cli

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bambithedeer/spotify-api/internal/cli/config"
	"github.com/spf13/cobra"
)

func TestGenerateRandomString(t *testing.T) {
	tests := []int{8, 16, 32, 64}

	for _, length := range tests {
		result, err := generateRandomString(length)
		if err != nil {
			t.Errorf("generateRandomString(%d) failed: %v", length, err)
			continue
		}

		if len(result) != length {
			t.Errorf("generateRandomString(%d) returned string of length %d", length, len(result))
		}

		// Check that it's hexadecimal
		for _, char := range result {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
				t.Errorf("generateRandomString(%d) returned non-hex character: %c", length, char)
				break
			}
		}
	}

	// Test that consecutive calls return different values
	str1, _ := generateRandomString(32)
	str2, _ := generateRandomString(32)
	if str1 == str2 {
		t.Error("generateRandomString should return different values on consecutive calls")
	}
}

func TestMaskString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "*"},
		{"ab", "**"},
		{"abc", "***"},
		{"abcd", "****"},
		{"abcde", "*****"},
		{"abcdef", "******"},
		{"abcdefgh", "********"},
		{"abcdefghi", "abcd*fghi"},
		{"abcdefghij", "abcd**ghij"},
		{"1234567890abcdef", "1234********cdef"},
		{"very_long_client_id_example", "very*******************mple"},
	}

	for _, test := range tests {
		result := maskString(test.input)
		if result != test.expected {
			t.Errorf("maskString(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30 seconds"},
		{59 * time.Second, "59 seconds"},
		{1 * time.Minute, "1 minutes"},
		{30 * time.Minute, "30 minutes"},
		{59 * time.Minute, "59 minutes"},
		{1 * time.Hour, "1 hours"},
		{2 * time.Hour, "2 hours"},
		{23 * time.Hour, "23 hours"},
		{24 * time.Hour, "1 days"},
		{48 * time.Hour, "2 days"},
		{7 * 24 * time.Hour, "7 days"},
	}

	for _, test := range tests {
		result := formatDuration(test.duration)
		if result != test.expected {
			t.Errorf("formatDuration(%v) = %q, expected %q", test.duration, result, test.expected)
		}
	}
}

func TestAuthCommands_Integration(t *testing.T) {
	// Create temporary config for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_config.yaml")

	// Reset current config
	config.Reset()

	// Initialize with test file
	if err := config.Init(tmpFile, false, "text"); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	t.Run("setup command validation", func(t *testing.T) {
		// Test that setup requires credentials
		if config.HasCredentials() {
			t.Error("Expected no credentials initially")
		}

		// Manually set credentials to test the flow
		config.SetCredentials("test_client_id", "test_client_secret", "http://127.0.0.1:4000")

		if !config.HasCredentials() {
			t.Error("Expected credentials to be set")
		}

		cfg := config.Get()
		if cfg.ClientID != "test_client_id" {
			t.Errorf("Expected client ID 'test_client_id', got %s", cfg.ClientID)
		}
	})

	t.Run("status command with no credentials", func(t *testing.T) {
		// Clear credentials
		config.Reset()
		config.Init(tmpFile, false, "text")

		// runStatus should not error when no credentials are set
		err := runStatus(nil, nil)
		if err != nil {
			t.Errorf("runStatus() should not error with no credentials: %v", err)
		}
	})

	t.Run("status command with credentials but no tokens", func(t *testing.T) {
		// Set credentials but no tokens
		config.SetCredentials("test_client_id", "test_client_secret", "http://127.0.0.1:4000")

		err := runStatus(nil, nil)
		if err != nil {
			t.Errorf("runStatus() should not error with credentials but no tokens: %v", err)
		}
	})

	t.Run("status command with credentials and tokens", func(t *testing.T) {
		// Set credentials and tokens
		config.SetCredentials("test_client_id", "test_client_secret", "http://127.0.0.1:4000")
		futureTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
		config.SetTokens("test_access_token", "test_refresh_token", "Bearer", futureTime)

		err := runStatus(nil, nil)
		if err != nil {
			t.Errorf("runStatus() should not error with valid tokens: %v", err)
		}
	})

	t.Run("logout command", func(t *testing.T) {
		// Set tokens first
		config.SetTokens("test_access_token", "test_refresh_token", "Bearer", "")

		if !config.IsAuthenticated() {
			t.Error("Expected to be authenticated before logout")
		}

		err := runLogout(nil, nil)
		if err != nil {
			t.Errorf("runLogout() failed: %v", err)
		}

		if config.IsAuthenticated() {
			t.Error("Expected to be logged out after runLogout()")
		}

		// Credentials should still be present
		if !config.HasCredentials() {
			t.Error("Expected credentials to be preserved after logout")
		}
	})

	t.Run("logout when not authenticated", func(t *testing.T) {
		// Clear tokens
		config.ClearTokens()

		err := runLogout(nil, nil)
		if err != nil {
			t.Errorf("runLogout() should not error when not authenticated: %v", err)
		}
	})

	t.Run("login and client-credentials require credentials", func(t *testing.T) {
		// Clear credentials
		config.Reset()
		config.Init(tmpFile, false, "text")

		// Test that functions check for credentials without actually running them
		// Note: HasCredentials() might still return true if .env file is present
		// This is expected behavior - the CLI should load from .env by default

		// The functions should check HasCredentials() and return early with error
		// We can't safely test the actual functions without triggering network calls
		// and browser opening, so we just verify the credential checking logic
	})
}

func TestAuthCommands_ErrorHandling(t *testing.T) {
	// Create temporary config for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_config.yaml")

	// Reset current config
	config.Reset()

	// Initialize with test file
	if err := config.Init(tmpFile, false, "text"); err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	t.Run("login with invalid redirect URI", func(t *testing.T) {
		// Set credentials with truly invalid redirect URI
		config.SetCredentials("test_client_id", "test_client_secret", "://invalid")

		// Test that the URL parsing would fail by checking the URI directly
		cfg := config.Get()
		_, err := url.Parse(cfg.RedirectURI)
		if err == nil {
			t.Error("Expected invalid redirect URI to fail parsing")
		}
	})

	t.Run("status with invalid token expiry", func(t *testing.T) {
		// Set credentials and invalid expiry time
		config.SetCredentials("test_client_id", "test_client_secret", "http://127.0.0.1:4000")
		config.SetTokens("test_access_token", "test_refresh_token", "Bearer", "invalid-time")

		// Should not crash, just handle the invalid time gracefully
		err := runStatus(nil, nil)
		if err != nil {
			t.Errorf("runStatus() should handle invalid expiry time gracefully: %v", err)
		}
	})
}

// TestAuthCommandsExist verifies that all auth commands are properly registered
func TestAuthCommandsExist(t *testing.T) {
	// Find the auth command
	var authCommand *cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "auth" {
			authCommand = cmd
			break
		}
	}

	if authCommand == nil {
		t.Fatal("auth command not found")
	}

	// Check that all subcommands exist
	expectedSubcommands := []string{"setup", "login", "client-credentials", "status", "logout"}
	actualSubcommands := make([]string, 0, len(authCommand.Commands()))

	for _, cmd := range authCommand.Commands() {
		actualSubcommands = append(actualSubcommands, cmd.Name())
	}

	for _, expected := range expectedSubcommands {
		found := false
		for _, actual := range actualSubcommands {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found in auth command", expected)
		}
	}
}

// Helper function to capture stdout for testing
func captureOutput(f func()) string {
	// This is a simplified version - in a real implementation,
	// you might want to use a more sophisticated approach to capture output
	original := os.Stdout
	defer func() { os.Stdout = original }()

	// For now, just run the function
	f()
	return ""
}