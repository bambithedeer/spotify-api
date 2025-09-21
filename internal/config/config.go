package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Spotify SpotifyConfig `yaml:"spotify"`
	Lidarr  LidarrConfig  `yaml:"lidarr"`
	Logging LoggingConfig `yaml:"logging"`
}

type SpotifyConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
	Scopes       []string `yaml:"scopes"`
}

type LidarrConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// Default configuration values
func DefaultConfig() *Config {
	return &Config{
		Spotify: SpotifyConfig{
			RedirectURI: "http://localhost:8080/callback",
			Scopes: []string{
				"user-read-private",
				"user-read-email",
				"user-library-read",
				"user-library-modify",
				"playlist-read-private",
				"playlist-modify-private",
				"playlist-modify-public",
				"user-top-read",
				"user-read-playback-state",
				"user-modify-playback-state",
				"user-read-recently-played",
			},
		},
		Lidarr: LidarrConfig{
			URL: "http://localhost:8686",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
			Output: "stdout",
		},
	}
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	config := DefaultConfig()

	// Load .env file if it exists (this will set environment variables)
	loadDotEnv()

	// Try to load from config file
	if err := loadFromFile(config); err != nil {
		// Config file is optional, so we only return error if it exists but is invalid
		if !os.IsNotExist(err) {
			return nil, errors.WrapConfigError(err, "failed to load config file")
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	// Validate required fields
	if err := validate(config); err != nil {
		return nil, errors.WrapConfigError(err, "config validation failed")
	}

	return config, nil
}

// loadDotEnv loads .env file if it exists
func loadDotEnv() {
	possibleEnvFiles := []string{
		".env",
		".env.local",
		filepath.Join(os.Getenv("HOME"), ".spotify-cli.env"),
	}

	for _, envFile := range possibleEnvFiles {
		if _, err := os.Stat(envFile); err == nil {
			_ = godotenv.Load(envFile) // Ignore errors as .env is optional
			break // Load only the first found .env file
		}
	}
}

// loadFromFile attempts to load configuration from various possible locations
func loadFromFile(config *Config) error {
	possiblePaths := []string{
		"spotify-cli.yaml",
		"spotify-cli.yml",
		"config.yaml",
		"config.yml",
		filepath.Join(os.Getenv("HOME"), ".config", "spotify-cli", "config.yaml"),
		filepath.Join(os.Getenv("HOME"), ".spotify-cli.yaml"),
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return loadConfigFile(path, config)
		}
	}

	return os.ErrNotExist
}

// loadConfigFile loads configuration from a specific file
func loadConfigFile(path string, config *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// Spotify configuration
	if val := os.Getenv("SPOTIFY_CLIENT_ID"); val != "" {
		config.Spotify.ClientID = val
	}
	if val := os.Getenv("SPOTIFY_CLIENT_SECRET"); val != "" {
		config.Spotify.ClientSecret = val
	}
	if val := os.Getenv("SPOTIFY_REDIRECT_URI"); val != "" {
		config.Spotify.RedirectURI = val
	}
	if val := os.Getenv("SPOTIFY_SCOPES"); val != "" {
		config.Spotify.Scopes = strings.Split(val, ",")
		// Trim whitespace
		for i, scope := range config.Spotify.Scopes {
			config.Spotify.Scopes[i] = strings.TrimSpace(scope)
		}
	}

	// Lidarr configuration
	if val := os.Getenv("LIDARR_URL"); val != "" {
		config.Lidarr.URL = val
	}
	if val := os.Getenv("LIDARR_API_KEY"); val != "" {
		config.Lidarr.APIKey = val
	}

	// Logging configuration
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		config.Logging.Level = val
	}
	if val := os.Getenv("LOG_FORMAT"); val != "" {
		config.Logging.Format = val
	}
	if val := os.Getenv("LOG_OUTPUT"); val != "" {
		config.Logging.Output = val
	}
}

// validate checks that required configuration is present
func validate(config *Config) error {
	if config.Spotify.ClientID == "" {
		return errors.NewValidationError("spotify client_id is required (set SPOTIFY_CLIENT_ID environment variable)")
	}
	if config.Spotify.ClientSecret == "" {
		return errors.NewValidationError("spotify client_secret is required (set SPOTIFY_CLIENT_SECRET environment variable)")
	}

	// Validate log level
	validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	levelValid := false
	for _, level := range validLevels {
		if strings.ToLower(config.Logging.Level) == level {
			levelValid = true
			break
		}
	}
	if !levelValid {
		return errors.NewValidationError(fmt.Sprintf("invalid log level: %s (valid: %s)", config.Logging.Level, strings.Join(validLevels, ", ")))
	}

	// Validate log format
	if config.Logging.Format != "text" && config.Logging.Format != "json" {
		return errors.NewValidationError(fmt.Sprintf("invalid log format: %s (valid: text, json)", config.Logging.Format))
	}

	return nil
}

// Save saves the current configuration to a file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return errors.WrapFileError(err, "failed to marshal config")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.WrapFileError(err, "failed to create config directory")
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return errors.WrapFileError(err, "failed to write config file")
	}

	return nil
}