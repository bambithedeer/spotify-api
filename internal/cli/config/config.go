package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	// Spotify API Configuration
	ClientID     string `yaml:"client_id" json:"client_id"`
	ClientSecret string `yaml:"client_secret" json:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri" json:"redirect_uri"`

	// Authentication
	AccessToken  string `yaml:"access_token,omitempty" json:"access_token,omitempty"`
	RefreshToken string `yaml:"refresh_token,omitempty" json:"refresh_token,omitempty"`
	TokenType    string `yaml:"token_type,omitempty" json:"token_type,omitempty"`
	ExpiresAt    string `yaml:"expires_at,omitempty" json:"expires_at,omitempty"`

	// CLI Settings
	DefaultOutput string `yaml:"default_output" json:"default_output"`
	Verbose       bool   `yaml:"verbose" json:"verbose"`
	ColorOutput   bool   `yaml:"color_output" json:"color_output"`

	// Cache Settings
	CacheEnabled bool   `yaml:"cache_enabled" json:"cache_enabled"`
	CacheTTL     string `yaml:"cache_ttl" json:"cache_ttl"`
}

var (
	current    *Config
	configFile string
	verbose    bool
	output     string
)

// Default returns a default configuration
func Default() *Config {
	// Try to load .env file (ignore errors if it doesn't exist)
	godotenv.Load()

	config := &Config{
		RedirectURI:   "http://127.0.0.1:4000",
		DefaultOutput: "text",
		Verbose:       false,
		ColorOutput:   true,
		CacheEnabled:  true,
		CacheTTL:      "1h",
	}

	// Override with environment variables if present
	if clientID := os.Getenv("SPOTIFY_CLIENT_ID"); clientID != "" {
		config.ClientID = clientID
	}
	if clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET"); clientSecret != "" {
		config.ClientSecret = clientSecret
	}
	if redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI"); redirectURI != "" {
		config.RedirectURI = redirectURI
	}

	return config
}

// Init initializes the configuration system
func Init(cfgFile string, verboseFlag bool, outputFlag string) error {
	configFile = cfgFile
	verbose = verboseFlag
	output = outputFlag

	// Load configuration
	config, err := load()
	if err != nil {
		return err
	}

	// Override with command line flags
	if verboseFlag {
		config.Verbose = true
	}
	if outputFlag != "" {
		config.DefaultOutput = outputFlag
	}

	current = config
	return nil
}

// Get returns the current configuration
func Get() *Config {
	if current == nil {
		current = Default()
	}
	return current
}

// Save saves the current configuration to file
func Save() error {
	if current == nil || configFile == "" {
		return fmt.Errorf("configuration not initialized")
	}

	// Ensure directory exists
	dir := filepath.Dir(configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(current)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Configuration saved to %s\n", configFile)
	}

	return nil
}

// SetCredentials sets the Spotify API credentials
func SetCredentials(clientID, clientSecret, redirectURI string) {
	config := Get()
	config.ClientID = clientID
	config.ClientSecret = clientSecret
	if redirectURI != "" {
		config.RedirectURI = redirectURI
	}
}

// SetTokens sets the authentication tokens
func SetTokens(accessToken, refreshToken, tokenType, expiresAt string) {
	config := Get()
	config.AccessToken = accessToken
	config.RefreshToken = refreshToken
	config.TokenType = tokenType
	config.ExpiresAt = expiresAt
}

// ClearTokens clears the authentication tokens
func ClearTokens() {
	config := Get()
	config.AccessToken = ""
	config.RefreshToken = ""
	config.TokenType = ""
	config.ExpiresAt = ""
}

// IsAuthenticated returns true if the user is authenticated with a valid token
func IsAuthenticated() bool {
	config := Get()

	// Check if access token exists
	if config.AccessToken == "" {
		return false
	}

	// Check if token is expired
	if config.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, config.ExpiresAt)
		if err != nil {
			// Invalid expiry format, consider token invalid
			return false
		}

		// Token is expired if current time is after expiry time
		if time.Now().After(expiresAt) {
			return false
		}
	}

	return true
}

// HasCredentials returns true if API credentials are configured
func HasCredentials() bool {
	config := Get()
	return config.ClientID != "" && config.ClientSecret != ""
}

// load loads configuration from file
func load() (*Config, error) {
	config := Default()

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Config file doesn't exist, return default config
		return config, nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// GetConfigFile returns the path to the config file
func GetConfigFile() string {
	return configFile
}

// IsTokenExpired returns true if the current token is expired
func IsTokenExpired() bool {
	config := Get()

	if config.ExpiresAt == "" {
		return false // No expiry info, assume valid
	}

	expiresAt, err := time.Parse(time.RFC3339, config.ExpiresAt)
	if err != nil {
		return true // Invalid format, consider expired
	}

	return time.Now().After(expiresAt)
}

// IsTokenExpiringSoon returns true if token expires within the next 5 minutes
func IsTokenExpiringSoon() bool {
	config := Get()

	if config.ExpiresAt == "" {
		return false // No expiry info
	}

	expiresAt, err := time.Parse(time.RFC3339, config.ExpiresAt)
	if err != nil {
		return true // Invalid format
	}

	return time.Now().Add(5 * time.Minute).After(expiresAt)
}

// Reset clears the current configuration (useful for testing)
func Reset() {
	current = nil
}
