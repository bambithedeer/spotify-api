package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bambithedeer/spotify-api/internal/auth"
	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/bambithedeer/spotify-api/internal/ratelimit"
)

const (
	SpotifyAPIBaseURL = "https://api.spotify.com/v1"
	DefaultTimeout    = 30 * time.Second
)

// Client represents a Spotify API client
type Client struct {
	httpClient  *http.Client
	authClient  *auth.Client
	token       *auth.Token
	baseURL     string
	rateLimiter *ratelimit.RateLimiter
	retryConfig *ratelimit.RetryConfig
}

// NewClient creates a new Spotify API client
func NewClient(clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		authClient:  auth.NewClient(clientID, clientSecret, redirectURI),
		baseURL:     SpotifyAPIBaseURL,
		rateLimiter: ratelimit.NewRateLimiter(),
		retryConfig: ratelimit.DefaultRetryConfig(),
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

// DeleteWithBody performs a DELETE request with body to the Spotify API
func (c *Client) DeleteWithBody(ctx context.Context, endpoint string, body io.Reader) (*http.Response, error) {
	return c.makeRequest(ctx, "DELETE", endpoint, body)
}

// makeRequest is the internal method that handles all HTTP requests with rate limiting and retries
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
	// Ensure we have a valid token
	if err := c.RefreshTokenIfNeeded(); err != nil {
		return nil, err
	}

	if c.token == nil {
		return nil, errors.NewAuthError("not authenticated")
	}

	// Implement retry logic with exponential backoff
	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		// Wait for rate limiter
		if err := c.rateLimiter.Wait(ctx); err != nil {
			return nil, errors.WrapNetworkError(err, "rate limiter wait failed")
		}

		// Create a new request for each attempt (body might need to be read multiple times)
		var requestBody io.Reader
		if body != nil {
			// For retry attempts, we need a fresh body reader
			// This is a limitation - callers should pass seekable readers for retries
			requestBody = body
		}

		resp, err := c.executeRequest(ctx, method, endpoint, requestBody)

		// If request succeeded or context was cancelled, return immediately
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			// Network error - should retry
			if attempt < c.retryConfig.MaxRetries {
				delay := c.retryConfig.GetRetryDelay(attempt, nil)
				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
			return nil, err
		}

		// Handle rate limiting
		if resp.StatusCode == http.StatusTooManyRequests {
			if err := c.rateLimiter.HandleRateLimitResponse(resp); err != nil {
				// If this is the last attempt, return the error
				if !c.retryConfig.ShouldRetry(resp, attempt) {
					resp.Body.Close()
					return nil, err
				}
				// Otherwise, wait and retry
				resp.Body.Close()
				delay := c.retryConfig.GetRetryDelay(attempt, resp)
				select {
				case <-time.After(delay):
					continue
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			}
		}

		// Check if we should retry based on status code
		if c.retryConfig.ShouldRetry(resp, attempt) {
			resp.Body.Close()
			delay := c.retryConfig.GetRetryDelay(attempt, resp)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Request succeeded, return response
		return resp, nil
	}

	return nil, errors.NewAPIError("max retries exceeded")
}

// executeRequest performs a single HTTP request without retry logic
func (c *Client) executeRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Response, error) {
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

	// Handle common HTTP errors that shouldn't be retried
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		resp.Body.Close()
		return nil, errors.NewAuthError("unauthorized - token may be invalid")
	case http.StatusForbidden:
		resp.Body.Close()
		return nil, errors.NewAuthError("forbidden - insufficient permissions")
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

// SetRateLimiter allows customization of the rate limiter
func (c *Client) SetRateLimiter(rl *ratelimit.RateLimiter) {
	c.rateLimiter = rl
}

// SetRetryConfig allows customization of the retry configuration
func (c *Client) SetRetryConfig(config *ratelimit.RetryConfig) {
	c.retryConfig = config
}

// GetRateLimiterStatus returns the current rate limiter status
func (c *Client) GetRateLimiterStatus() (availableTokens int, maxTokens int, retryAfter time.Time) {
	return c.rateLimiter.GetStatus()
}

// SetBaseURL sets the base URL for the client (useful for testing)
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}