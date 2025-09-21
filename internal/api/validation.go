package api

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/errors"
)

// Validator handles validation of API requests and parameters
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// SpotifyID validation patterns
var (
	spotifyIDPattern = regexp.MustCompile(`^[0-9A-Za-z]{22}$`)
	spotifyURIPattern = regexp.MustCompile(`^spotify:([a-z]+):([0-9A-Za-z]{22})$`)
)

// ValidateSpotifyID validates a Spotify ID format
func (v *Validator) ValidateSpotifyID(id string) error {
	if id == "" {
		return errors.NewValidationError("spotify ID cannot be empty")
	}

	if !spotifyIDPattern.MatchString(id) {
		return errors.NewValidationError(fmt.Sprintf("invalid Spotify ID format: %s", id))
	}

	return nil
}

// ValidateSpotifyIDs validates multiple Spotify IDs
func (v *Validator) ValidateSpotifyIDs(ids []string) error {
	if len(ids) == 0 {
		return errors.NewValidationError("at least one Spotify ID is required")
	}

	for i, id := range ids {
		if err := v.ValidateSpotifyID(id); err != nil {
			return errors.NewValidationError(fmt.Sprintf("invalid ID at position %d: %v", i, err))
		}
	}

	return nil
}

// ValidateSpotifyURI validates a Spotify URI format
func (v *Validator) ValidateSpotifyURI(uri string) error {
	if uri == "" {
		return errors.NewValidationError("spotify URI cannot be empty")
	}

	if !spotifyURIPattern.MatchString(uri) {
		return errors.NewValidationError(fmt.Sprintf("invalid Spotify URI format: %s", uri))
	}

	return nil
}

// ExtractIDFromURI extracts the ID from a Spotify URI
func (v *Validator) ExtractIDFromURI(uri string) (string, error) {
	matches := spotifyURIPattern.FindStringSubmatch(uri)
	if len(matches) != 3 {
		return "", errors.NewValidationError(fmt.Sprintf("invalid Spotify URI format: %s", uri))
	}

	return matches[2], nil
}

// ValidateMarket validates an ISO 3166-1 alpha-2 country code
func (v *Validator) ValidateMarket(market string) error {
	if market == "" {
		return nil // Market is optional in most cases
	}

	if market == "from_token" {
		return nil // Special value for user's market
	}

	if len(market) != 2 {
		return errors.NewValidationError("market must be a 2-letter ISO 3166-1 alpha-2 country code")
	}

	// Convert to uppercase for consistency
	market = strings.ToUpper(market)

	// Basic validation - in a real implementation you might want a comprehensive list
	if !regexp.MustCompile(`^[A-Z]{2}$`).MatchString(market) {
		return errors.NewValidationError("market must contain only letters")
	}

	return nil
}

// ValidateLimit validates limit parameters for pagination
func (v *Validator) ValidateLimit(limit, min, max int) error {
	if limit < min {
		return errors.NewValidationError(fmt.Sprintf("limit must be at least %d", min))
	}

	if limit > max {
		return errors.NewValidationError(fmt.Sprintf("limit must be at most %d", max))
	}

	return nil
}

// ValidateOffset validates offset parameters for pagination
func (v *Validator) ValidateOffset(offset int) error {
	if offset < 0 {
		return errors.NewValidationError("offset must be non-negative")
	}

	return nil
}

// ValidateTimeRange validates time range parameters
func (v *Validator) ValidateTimeRange(timeRange string) error {
	validRanges := []string{"short_term", "medium_term", "long_term"}

	if timeRange == "" {
		return nil // Time range is optional
	}

	for _, valid := range validRanges {
		if timeRange == valid {
			return nil
		}
	}

	return errors.NewValidationError(fmt.Sprintf("time_range must be one of: %s", strings.Join(validRanges, ", ")))
}

// ValidateSearchQuery validates search query parameters
func (v *Validator) ValidateSearchQuery(query string) error {
	if strings.TrimSpace(query) == "" {
		return errors.NewValidationError("search query cannot be empty")
	}

	if len(query) > 500 {
		return errors.NewValidationError("search query cannot exceed 500 characters")
	}

	return nil
}

// ValidateSearchTypes validates search type parameters
func (v *Validator) ValidateSearchTypes(types []string) error {
	if len(types) == 0 {
		return errors.NewValidationError("at least one search type is required")
	}

	validTypes := map[string]bool{
		"album":     true,
		"artist":    true,
		"playlist":  true,
		"track":     true,
		"show":      true,
		"episode":   true,
		"audiobook": true,
	}

	for _, searchType := range types {
		if !validTypes[searchType] {
			return errors.NewValidationError(fmt.Sprintf("invalid search type: %s", searchType))
		}
	}

	return nil
}

// ValidateURL validates URL format
func (v *Validator) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return errors.NewValidationError("URL cannot be empty")
	}

	_, err := url.Parse(urlStr)
	if err != nil {
		return errors.WrapValidationError(err, "invalid URL format")
	}

	return nil
}

// ValidatePlaylistName validates playlist name
func (v *Validator) ValidatePlaylistName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.NewValidationError("playlist name cannot be empty")
	}

	if len(name) > 100 {
		return errors.NewValidationError("playlist name cannot exceed 100 characters")
	}

	return nil
}

// ValidatePlaylistDescription validates playlist description
func (v *Validator) ValidatePlaylistDescription(description string) error {
	if len(description) > 300 {
		return errors.NewValidationError("playlist description cannot exceed 300 characters")
	}

	return nil
}

// ValidateVolumePercent validates volume percentage (0-100)
func (v *Validator) ValidateVolumePercent(volume int) error {
	if volume < 0 || volume > 100 {
		return errors.NewValidationError("volume must be between 0 and 100")
	}

	return nil
}

// ValidatePosition validates position in playlist/queue
func (v *Validator) ValidatePosition(position int) error {
	if position < 0 {
		return errors.NewValidationError("position must be non-negative")
	}

	return nil
}

// ValidatePositionMs validates position in milliseconds
func (v *Validator) ValidatePositionMs(positionMs int) error {
	if positionMs < 0 {
		return errors.NewValidationError("position in milliseconds must be non-negative")
	}

	return nil
}

// NormalizeAndValidateIDs normalizes and validates a list of IDs or URIs
func (v *Validator) NormalizeAndValidateIDs(input []string) ([]string, error) {
	if len(input) == 0 {
		return nil, errors.NewValidationError("at least one ID or URI is required")
	}

	normalized := make([]string, len(input))

	for i, item := range input {
		// Try to extract ID from URI first
		if strings.HasPrefix(item, "spotify:") {
			id, err := v.ExtractIDFromURI(item)
			if err != nil {
				return nil, err
			}
			normalized[i] = id
		} else {
			// Validate as ID
			if err := v.ValidateSpotifyID(item); err != nil {
				return nil, err
			}
			normalized[i] = item
		}
	}

	return normalized, nil
}