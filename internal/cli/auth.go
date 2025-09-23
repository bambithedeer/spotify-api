package cli

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/bambithedeer/spotify-api/internal/auth"
	"github.com/bambithedeer/spotify-api/internal/cli/client"
	"github.com/bambithedeer/spotify-api/internal/cli/config"
	"github.com/bambithedeer/spotify-api/internal/cli/utils"
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long: `Commands for managing Spotify API authentication.

This includes setting up API credentials, logging in with user accounts,
and managing authentication tokens.`,
	Example: `  # Set up API credentials
  spotify-cli auth setup

  # Login with user account (requires user interaction)
  spotify-cli auth login

  # Login with client credentials (app-only access)
  spotify-cli auth client-credentials

  # Check authentication status
  spotify-cli auth status

  # Logout and clear tokens
  spotify-cli auth logout`,
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Set up Spotify API credentials",
	Long: `Set up your Spotify API credentials (Client ID and Client Secret).

You can get these credentials by creating a Spotify app at:
https://developer.spotify.com/dashboard

This command will prompt you to enter your credentials interactively.`,
	Example: `  spotify-cli auth setup`,
	RunE:    runSetup,
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with user account authorization",
	Long: `Login using Spotify's Authorization Code flow to access user data.

This command will:
1. Open your browser to Spotify's authorization page
2. Start a local server to receive the authorization code
3. Exchange the code for access and refresh tokens
4. Save the tokens for future use

Requires API credentials to be set up first with 'auth setup'.`,
	Example: `  spotify-cli auth login`,
	RunE:    runLogin,
}

var clientCredentialsCmd = &cobra.Command{
	Use:   "client-credentials",
	Short: "Login with client credentials (app-only access)",
	Long: `Login using Client Credentials flow for app-only access to public data.

This flow provides access to:
- Search for tracks, albums, artists, playlists
- Get track, album, artist, and playlist information
- Browse featured playlists and categories

This does NOT provide access to user-specific data like:
- User's library, playlists, or profile
- Playback control
- Following/unfollowing

Requires API credentials to be set up first with 'auth setup'.`,
	Example: `  spotify-cli auth client-credentials`,
	RunE:    runClientCredentials,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long: `Display information about the current authentication state.

Shows:
- Whether API credentials are configured
- Current authentication status
- Token expiration time (if available)
- Available scopes (if available)`,
	Example: `  spotify-cli auth status`,
	RunE:    runStatus,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear stored tokens",
	Long: `Clear all stored authentication tokens.

This will remove:
- Access tokens
- Refresh tokens
- Token expiration information

API credentials (Client ID and Client Secret) will be preserved.`,
	Example: `  spotify-cli auth logout`,
	RunE:    runLogout,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(setupCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(clientCredentialsCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(logoutCmd)
}

func runSetup(cmd *cobra.Command, args []string) error {
	fmt.Println("Setting up Spotify API credentials")
	fmt.Println()
	fmt.Println("You can get these credentials by creating a Spotify app at:")
	fmt.Println("https://developer.spotify.com/dashboard")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Get Client ID
	fmt.Print("Enter your Client ID: ")
	clientID, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read client ID: %w", err)
	}
	clientID = strings.TrimSpace(clientID)

	if clientID == "" {
		return fmt.Errorf("client ID cannot be empty")
	}

	// Get Client Secret
	fmt.Print("Enter your Client Secret: ")
	clientSecret, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read client secret: %w", err)
	}
	clientSecret = strings.TrimSpace(clientSecret)

	if clientSecret == "" {
		return fmt.Errorf("client secret cannot be empty")
	}

	// Get Redirect URI (optional, with default)
	fmt.Printf("Enter your Redirect URI [%s]: ", config.Get().RedirectURI)
	redirectURI, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read redirect URI: %w", err)
	}
	redirectURI = strings.TrimSpace(redirectURI)

	if redirectURI == "" {
		redirectURI = config.Get().RedirectURI
	}

	// Validate redirect URI
	if _, err := url.Parse(redirectURI); err != nil {
		return fmt.Errorf("invalid redirect URI: %w", err)
	}

	// Save credentials
	config.SetCredentials(clientID, clientSecret, redirectURI)
	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	utils.PrintSuccess("Credentials saved successfully!")
	fmt.Println()
	fmt.Println("You can now use:")
	fmt.Println("  spotify-cli auth login                 # For user data access")
	fmt.Println("  spotify-cli auth client-credentials    # For public data access")

	return nil
}

func runLogin(cmd *cobra.Command, args []string) error {
	if !config.HasCredentials() {
		return fmt.Errorf("credentials not configured. Run 'spotify-cli auth setup' first")
	}

	cfg := config.Get()
	authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI)

	// Parse redirect URI to get port
	redirectURL, err := url.Parse(cfg.RedirectURI)
	if err != nil {
		return fmt.Errorf("invalid redirect URI: %w", err)
	}


	// Generate random state
	state, err := generateRandomString(32)
	if err != nil {
		return fmt.Errorf("failed to generate state: %w", err)
	}

	// Define scopes for full user access
	scopes := []string{
		"user-read-private",
		"user-read-email",
		"user-library-read",
		"user-library-modify",
		"user-read-playback-state",
		"user-modify-playback-state",
		"user-read-currently-playing",
		"playlist-read-private",
		"playlist-read-collaborative",
		"playlist-modify-public",
		"playlist-modify-private",
		"user-follow-read",
		"user-follow-modify",
		"user-read-recently-played",
		"user-top-read",
	}

	// Get authorization URL
	authURL := authClient.GetAuthorizationURL(scopes, state)

	fmt.Println("Opening browser for Spotify authorization...")
	fmt.Println()
	fmt.Println("If the browser doesn't open automatically, visit this URL:")
	fmt.Println(authURL)
	fmt.Println()

	// Open browser
	if err := openBrowser(authURL); err != nil {
		utils.PrintWarning("Failed to open browser automatically")
	}

	// Start local server to receive callback
	authCode := make(chan string, 1)
	authError := make(chan error, 1)

	server := &http.Server{
		Addr: redirectURL.Host,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check state parameter
			if r.URL.Query().Get("state") != state {
				authError <- fmt.Errorf("invalid state parameter")
				http.Error(w, "Invalid state parameter", http.StatusBadRequest)
				return
			}

			// Check for error
			if errorParam := r.URL.Query().Get("error"); errorParam != "" {
				authError <- fmt.Errorf("authorization error: %s", errorParam)
				http.Error(w, "Authorization error: "+errorParam, http.StatusBadRequest)
				return
			}

			// Get authorization code
			code := r.URL.Query().Get("code")
			if code == "" {
				authError <- fmt.Errorf("no authorization code received")
				http.Error(w, "No authorization code received", http.StatusBadRequest)
				return
			}

			// Send success response
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprint(w, `
				<html>
					<head><title>Spotify CLI Authorization</title></head>
					<body>
						<h1>Authorization Successful!</h1>
						<p>You can now close this browser window and return to the CLI.</p>
					</body>
				</html>
			`)

			authCode <- code
		}),
	}

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			authError <- fmt.Errorf("failed to start callback server: %w", err)
		}
	}()

	fmt.Printf("Waiting for authorization (listening on %s)...\n", redirectURL.Host)

	// Wait for authorization or timeout
	var code string
	select {
	case code = <-authCode:
		// Success - continue
	case err := <-authError:
		server.Shutdown(context.Background())
		return err
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		return fmt.Errorf("authorization timeout after 5 minutes")
	}

	// Shutdown server
	server.Shutdown(context.Background())

	fmt.Println("Authorization code received, exchanging for tokens...")

	// Exchange code for tokens
	token, err := authClient.ExchangeCode(code)
	if err != nil {
		return fmt.Errorf("failed to exchange authorization code: %w", err)
	}

	// Save tokens
	expiresAt := ""
	if !token.Expiry.IsZero() {
		expiresAt = token.Expiry.Format(time.RFC3339)
	}

	config.SetTokens(token.AccessToken, token.RefreshToken, token.TokenType, expiresAt)
	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	utils.PrintSuccess("Login successful!")

	if token.Scope != "" {
		fmt.Printf("Granted scopes: %s\n", token.Scope)
	}

	if !token.Expiry.IsZero() {
		fmt.Printf("Token expires: %s\n", token.Expiry.Format("2006-01-02 15:04:05 MST"))
	}

	return nil
}

func runClientCredentials(cmd *cobra.Command, args []string) error {
	if !config.HasCredentials() {
		return fmt.Errorf("credentials not configured. Run 'spotify-cli auth setup' first")
	}

	cfg := config.Get()
	authClient := auth.NewClient(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI)

	fmt.Println("Authenticating with client credentials...")

	token, err := authClient.ClientCredentials()
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save token (no refresh token for client credentials)
	expiresAt := ""
	if !token.Expiry.IsZero() {
		expiresAt = token.Expiry.Format(time.RFC3339)
	}

	config.SetTokens(token.AccessToken, "", token.TokenType, expiresAt)
	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	utils.PrintSuccess("Client credentials authentication successful!")

	if !token.Expiry.IsZero() {
		fmt.Printf("Token expires: %s\n", token.Expiry.Format("2006-01-02 15:04:05 MST"))
	}

	fmt.Println()
	fmt.Println("You now have access to public Spotify data.")
	fmt.Println("Try: spotify-cli search track \"bohemian rhapsody\"")

	return nil
}

// AuthStatus represents the authentication status for structured output
type AuthStatus struct {
	Credentials struct {
		Configured  bool   `json:"configured" yaml:"configured"`
		ClientID    string `json:"client_id,omitempty" yaml:"client_id,omitempty"`
		RedirectURI string `json:"redirect_uri,omitempty" yaml:"redirect_uri,omitempty"`
	} `json:"credentials" yaml:"credentials"`
	Authentication struct {
		Active           bool   `json:"active" yaml:"active"`
		TokenType        string `json:"token_type,omitempty" yaml:"token_type,omitempty"`
		ExpiresAt        string `json:"expires_at,omitempty" yaml:"expires_at,omitempty"`
		TimeRemaining    string `json:"time_remaining,omitempty" yaml:"time_remaining,omitempty"`
		IsExpired        bool   `json:"is_expired,omitempty" yaml:"is_expired,omitempty"`
		IsExpiringSoon   bool   `json:"is_expiring_soon,omitempty" yaml:"is_expiring_soon,omitempty"`
		HasRefreshToken  bool   `json:"has_refresh_token" yaml:"has_refresh_token"`
		TokenValidation  bool   `json:"token_validation" yaml:"token_validation"`
	} `json:"authentication" yaml:"authentication"`
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	// Prepare status data
	status := AuthStatus{}

	// Credentials status
	status.Credentials.Configured = config.HasCredentials()
	if status.Credentials.Configured {
		status.Credentials.ClientID = maskString(cfg.ClientID)
		status.Credentials.RedirectURI = cfg.RedirectURI
	}

	// Authentication status
	status.Authentication.Active = config.IsAuthenticated()
	if cfg.AccessToken != "" {
		status.Authentication.TokenType = cfg.TokenType
		status.Authentication.ExpiresAt = cfg.ExpiresAt
		status.Authentication.IsExpired = config.IsTokenExpired()
		status.Authentication.IsExpiringSoon = config.IsTokenExpiringSoon()
		status.Authentication.HasRefreshToken = cfg.RefreshToken != ""

		if cfg.ExpiresAt != "" {
			if expiresAt, err := time.Parse(time.RFC3339, cfg.ExpiresAt); err == nil {
				if !status.Authentication.IsExpired {
					timeLeft := time.Until(expiresAt)
					status.Authentication.TimeRemaining = formatDuration(timeLeft)
				}
			}
		}

		// Test token validation
		if spotifyClient, err := client.NewSpotifyClient(); err == nil {
			status.Authentication.TokenValidation = spotifyClient.IsAuthenticated()
		}
	}

	// Check if we should output structured data
	if cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml" {
		return utils.Output(status)
	}

	// Text output (existing format)
	fmt.Println("Spotify CLI Authentication Status")
	fmt.Println("=================================")
	fmt.Println()

	// Check credentials
	if status.Credentials.Configured {
		utils.PrintSuccess("API credentials: Configured")
		fmt.Printf("Client ID: %s\n", status.Credentials.ClientID)
		fmt.Printf("Redirect URI: %s\n", status.Credentials.RedirectURI)
	} else {
		utils.PrintWarning("API credentials: Not configured")
		fmt.Println("Run 'spotify-cli auth setup' to configure credentials")
		return nil
	}

	fmt.Println()

	// Check authentication
	if status.Authentication.Active {
		utils.PrintSuccess("Authentication: Active")

		if status.Authentication.TokenType != "" {
			fmt.Printf("Token type: %s\n", status.Authentication.TokenType)
		}

		if status.Authentication.ExpiresAt != "" {
			expiresAt, err := time.Parse(time.RFC3339, status.Authentication.ExpiresAt)
			if err == nil {
				fmt.Printf("Token expires: %s\n", expiresAt.Format("2006-01-02 15:04:05 MST"))

				if status.Authentication.IsExpired {
					utils.PrintWarning("Token has expired - please re-authenticate")
				} else if status.Authentication.IsExpiringSoon {
					utils.PrintWarning("Token expires soon - consider re-authenticating")
					fmt.Printf("Time remaining: %s\n", status.Authentication.TimeRemaining)
				} else {
					fmt.Printf("Time remaining: %s\n", status.Authentication.TimeRemaining)
				}
			}
		}

		// Check if we have refresh token
		if status.Authentication.HasRefreshToken {
			fmt.Println("Refresh token: Available")
		} else {
			fmt.Println("Refresh token: Not available (client credentials flow)")
		}

		// Token validation result
		if status.Authentication.TokenValidation {
			utils.PrintSuccess("Token validation: Passed")
		} else {
			utils.PrintWarning("Token validation: Failed")
		}

		// Verbose output - show additional details
		if cfg.Verbose {
			utils.PrintVerbose("Token details:")
			utils.PrintVerbose("  Access token length: %d characters", len(cfg.AccessToken))
			if cfg.RefreshToken != "" {
				utils.PrintVerbose("  Refresh token length: %d characters", len(cfg.RefreshToken))
			}
			utils.PrintVerbose("  Config file: %s", config.GetConfigFile())
		}
	} else {
		// Check if we have an expired token
		if cfg.AccessToken != "" && status.Authentication.IsExpired {
			utils.PrintWarning("Authentication: Token expired")
			fmt.Println("Your authentication token has expired. Please re-authenticate.")
		} else {
			utils.PrintWarning("Authentication: Not authenticated")
		}

		fmt.Println()
		fmt.Println("Available authentication methods:")
		fmt.Println("  spotify-cli auth login                 # User account access")
		fmt.Println("  spotify-cli auth client-credentials    # Public data access")
	}

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	if !config.IsAuthenticated() {
		fmt.Println("Not currently authenticated.")
		return nil
	}

	config.ClearTokens()
	if err := config.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	utils.PrintSuccess("Logout successful!")
	fmt.Println("Authentication tokens have been cleared.")
	fmt.Println("API credentials have been preserved.")

	return nil
}

// Helper functions

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes)[:length], nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func maskString(s string) string {
	if len(s) <= 8 {
		return strings.Repeat("*", len(s))
	}
	return s[:4] + strings.Repeat("*", len(s)-8) + s[len(s)-4:]
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
