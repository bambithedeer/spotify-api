package client

import (
	"fmt"
	"time"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/auth"
	"github.com/bambithedeer/spotify-api/internal/cli/config"
	"github.com/bambithedeer/spotify-api/internal/client"
	"github.com/bambithedeer/spotify-api/internal/spotify"
)

// SpotifyClient wraps the Spotify API client for CLI use
type SpotifyClient struct {
	client *client.Client

	// Services
	Search    *spotify.SearchService
	Albums    *spotify.AlbumsService
	Artists   *spotify.ArtistsService
	Tracks    *spotify.TracksService
	Playlists *spotify.PlaylistsService
	Library   *spotify.LibraryService
	Users     *spotify.UsersService
	Player    *spotify.PlayerService
}

// NewSpotifyClient creates a new Spotify client for CLI use
func NewSpotifyClient() (*SpotifyClient, error) {
	cfg := config.Get()

	if !config.HasCredentials() {
		return nil, fmt.Errorf("Spotify API credentials not configured. Run 'spotify-cli auth setup' first")
	}

	// Create the underlying client
	spotifyClient := client.NewClient(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI)

	// Set token if available
	if config.IsAuthenticated() {
		token, err := parseToken(cfg)
		if err != nil {
			return nil, fmt.Errorf("invalid token configuration: %w", err)
		}
		spotifyClient.SetToken(token)
	}

	// Create service instances
	sc := &SpotifyClient{
		client: spotifyClient,
	}

	sc.initServices()

	return sc, nil
}

// NewUnauthenticatedClient creates a client that can be used for authentication
func NewUnauthenticatedClient() (*SpotifyClient, error) {
	cfg := config.Get()

	if !config.HasCredentials() {
		return nil, fmt.Errorf("Spotify API credentials not configured")
	}

	spotifyClient := client.NewClient(cfg.ClientID, cfg.ClientSecret, cfg.RedirectURI)

	sc := &SpotifyClient{
		client: spotifyClient,
	}

	return sc, nil
}

// GetClient returns the underlying Spotify client
func (sc *SpotifyClient) GetClient() *client.Client {
	return sc.client
}

// IsAuthenticated returns true if the client is authenticated
func (sc *SpotifyClient) IsAuthenticated() bool {
	return sc.client.GetToken() != nil
}

// Authenticate performs client credentials authentication for public data access
func (sc *SpotifyClient) Authenticate() error {
	return sc.client.AuthenticateClientCredentials()
}

// SaveToken saves the current token to configuration
func (sc *SpotifyClient) SaveToken() error {
	token := sc.client.GetToken()
	if token == nil {
		return fmt.Errorf("no token to save")
	}

	expiresAt := ""
	if !token.Expiry.IsZero() {
		expiresAt = token.Expiry.Format(time.RFC3339)
	}

	config.SetTokens(token.AccessToken, token.RefreshToken, token.TokenType, expiresAt)
	return config.Save()
}

// initServices initializes all service instances
func (sc *SpotifyClient) initServices() {
	requestBuilder := api.NewRequestBuilder(sc.client)

	sc.Search = spotify.NewSearchService(requestBuilder)
	sc.Albums = spotify.NewAlbumsService(requestBuilder)
	sc.Artists = spotify.NewArtistsService(requestBuilder)
	sc.Tracks = spotify.NewTracksService(requestBuilder)
	sc.Playlists = spotify.NewPlaylistsService(requestBuilder)
	sc.Library = spotify.NewLibraryService(requestBuilder)
	sc.Users = spotify.NewUsersService(requestBuilder)
	sc.Player = spotify.NewPlayerService(requestBuilder)
}

// parseToken converts config token data to auth.Token
func parseToken(cfg *config.Config) (*auth.Token, error) {
	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("no access token")
	}

	token := &auth.Token{
		AccessToken:  cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		TokenType:    cfg.TokenType,
	}

	if cfg.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, cfg.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at format: %w", err)
		}
		token.Expiry = expiresAt
	}

	return token, nil
}