package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bambithedeer/spotify-api/internal/auth"
	"github.com/bambithedeer/spotify-api/internal/errors"
)

const (
	SpotifyAPIBaseURL = "https://api.spotify.com/v1"
	DefaultTimeout    = 30 * time.Second
)

// Client represents a Spotify API client
type Client struct {
	httpClient *http.Client
	authClient *auth.Client
	token      *auth.Token
	baseURL    string
}

// NewClient creates a new Spotify API client
func NewClient(clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		authClient: auth.NewClient(clientID, clientSecret, redirectURI),
		baseURL:    SpotifyAPIBaseURL,
	}
}

// AuthenticateClientCredentials authenticates using client credentials flow
// This provides access to public data only (no user-specific data)
func (c *Client) AuthenticateClientCredentials() error {
	token, err := c.authClient.ClientCredentials()
	if err != nil {
		return errors.WrapAuthError(err, "client credentials authentication failed")
	}

	c.token = token
	return nil
}

// SetToken sets the access token (for when user has already authenticated)
func (c *Client) SetToken(token *auth.Token) {
	c.token = token
}

// GetToken returns the current token
func (c *Client) GetToken() *auth.Token {
	return c.token
}

// RefreshTokenIfNeeded refreshes the token if it's expired
func (c *Client) RefreshTokenIfNeeded() error {
	if c.token == nil {
		return errors.NewAuthError("no token available")
	}

	if !c.token.IsExpired() {
		return nil // Token is still valid
	}

	if c.token.RefreshToken == "" {
		return errors.NewAuthError("token expired and no refresh token available")
	}

	newToken, err := c.authClient.RefreshToken(c.token.RefreshToken)
	if err != nil {
		return errors.WrapAuthError(err, "failed to refresh token")
	}

	c.token = newToken
	return nil
}

// Get performs a GET request to the Spotify API
func (c *Client) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.makeRequest(ctx, "GET", endpoint, nil)
}

// Post performs a POST request to the Spotify API
func (c *Client) Post(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	return c.makeRequest(ctx, "POST", endpoint, body)
}

// Put performs a PUT request to the Spotify API
func (c *Client) Put(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	return c.makeRequest(ctx, "PUT", endpoint, body)
}

// Delete performs a DELETE request to the Spotify API
func (c *Client) Delete(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.makeRequest(ctx, "DELETE", endpoint, nil)
}

// makeRequest is the internal method that handles all HTTP requests
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	// Ensure we have a valid token
	if err := c.RefreshTokenIfNeeded(); err != nil {
		return nil, err
	}

	if c.token == nil {
		return nil, errors.NewAuthError("not authenticated")
	}

	// Build the full URL
	url := c.baseURL + endpoint

	// Create the request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, errors.WrapNetworkError(err, "failed to create request")
	}

	// Add authentication header
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", c.token.TokenType, c.token.AccessToken))
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.WrapNetworkError(err, "request failed")
	}

	// Handle common HTTP errors
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return nil, errors.NewAuthError("unauthorized - token may be invalid")
	case http.StatusForbidden:
		return nil, errors.NewAuthError("forbidden - insufficient permissions")
	case http.StatusTooManyRequests:
		return nil, errors.NewAPIError("rate limited - too many requests")
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return nil, errors.NewAPIError("Spotify API error - service unavailable")
	}

	return resp, nil
}

// GetAuthorizationURL returns the authorization URL for user authentication
func (c *Client) GetAuthorizationURL(scopes []string, state string) string {
	return c.authClient.GetAuthorizationURL(scopes, state)
}

// ExchangeCode exchanges an authorization code for tokens
func (c *Client) ExchangeCode(code string) error {
	token, err := c.authClient.ExchangeCode(code)
	if err != nil {
		return errors.WrapAuthError(err, "failed to exchange authorization code")
	}

	c.token = token
	return nil
}