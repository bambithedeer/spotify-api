package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bambithedeer/spotify-api/internal/errors"
)

const (
	SpotifyTokenURL     = "https://accounts.spotify.com/api/token"
	SpotifyAuthorizeURL = "https://accounts.spotify.com/authorize"
)

// TokenResponse represents the response from Spotify token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// Token represents an access token with expiration info
type Token struct {
	AccessToken  string
	TokenType    string
	RefreshToken string
	Expiry       time.Time
	Scope        string
}

// IsExpired checks if the token is expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.Expiry)
}

// Client handles Spotify authentication
type Client struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	httpClient   *http.Client
}

// NewClient creates a new authentication client
func NewClient(clientID, clientSecret, redirectURI string) *Client {
	return &Client{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ClientCredentials performs the Client Credentials flow
// This is used for accessing public data that doesn't require user authorization
func (c *Client) ClientCredentials() (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", SpotifyTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, errors.WrapAuthError(err, "failed to create token request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+c.basicAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.WrapNetworkError(err, "failed to request token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.NewAuthError(fmt.Sprintf("token request failed: %s - %s", resp.Status, string(body)))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, errors.WrapAuthError(err, "failed to decode token response")
	}

	return &Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		Expiry:      time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:       tokenResp.Scope,
	}, nil
}

// GetAuthorizationURL returns the URL for user authorization
// Used for Authorization Code flow to access user data
func (c *Client) GetAuthorizationURL(scopes []string, state string) string {
	params := url.Values{}
	params.Set("client_id", c.ClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", c.RedirectURI)
	params.Set("scope", strings.Join(scopes, " "))
	params.Set("state", state)

	return SpotifyAuthorizeURL + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for an access token
func (c *Client) ExchangeCode(code string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", c.RedirectURI)

	req, err := http.NewRequest("POST", SpotifyTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, errors.WrapAuthError(err, "failed to create code exchange request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+c.basicAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.WrapNetworkError(err, "failed to exchange code")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.NewAuthError(fmt.Sprintf("code exchange failed: %s - %s", resp.Status, string(body)))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, errors.WrapAuthError(err, "failed to decode token response")
	}

	return &Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (c *Client) RefreshToken(refreshToken string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequest("POST", SpotifyTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, errors.WrapAuthError(err, "failed to create refresh request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+c.basicAuth())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.WrapNetworkError(err, "failed to refresh token")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.NewAuthError(fmt.Sprintf("token refresh failed: %s - %s", resp.Status, string(body)))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, errors.WrapAuthError(err, "failed to decode refresh response")
	}

	// If no new refresh token is provided, keep the old one
	newRefreshToken := tokenResp.RefreshToken
	if newRefreshToken == "" {
		newRefreshToken = refreshToken
	}

	return &Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: newRefreshToken,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Scope:        tokenResp.Scope,
	}, nil
}

// basicAuth returns the base64 encoded client credentials for Basic auth
func (c *Client) basicAuth() string {
	credentials := c.ClientID + ":" + c.ClientSecret
	return base64.StdEncoding.EncodeToString([]byte(credentials))
}