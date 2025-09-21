package api

import (
	"strings"
	"testing"
)

func TestValidator_ValidateSpotifyID(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"valid ID", "4iV5W9uYEdYUVa79Axb7Rh", false},
		{"empty ID", "", true},
		{"too short", "4iV5W9uYEdYUVa79Axb7R", true},
		{"too long", "4iV5W9uYEdYUVa79Axb7Rh1", true},
		{"invalid characters", "4iV5W9uYEdYUVa79Axb7R!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSpotifyID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSpotifyID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateSpotifyIDs(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		ids     []string
		wantErr bool
	}{
		{"valid IDs", []string{"4iV5W9uYEdYUVa79Axb7Rh", "1301WleyT98MSxVHPZCA6M"}, false},
		{"empty slice", []string{}, true},
		{"mixed valid/invalid", []string{"4iV5W9uYEdYUVa79Axb7Rh", "invalid"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSpotifyIDs(tt.ids)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSpotifyIDs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateSpotifyURI(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		uri     string
		wantErr bool
	}{
		{"valid track URI", "spotify:track:4iV5W9uYEdYUVa79Axb7Rh", false},
		{"valid artist URI", "spotify:artist:4iV5W9uYEdYUVa79Axb7Rh", false},
		{"valid album URI", "spotify:album:4iV5W9uYEdYUVa79Axb7Rh", false},
		{"empty URI", "", true},
		{"invalid format", "invalid:track:4iV5W9uYEdYUVa79Axb7Rh", true},
		{"missing ID", "spotify:track:", true},
		{"invalid ID in URI", "spotify:track:invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSpotifyURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSpotifyURI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ExtractIDFromURI(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		uri      string
		expected string
		wantErr  bool
	}{
		{"valid track URI", "spotify:track:4iV5W9uYEdYUVa79Axb7Rh", "4iV5W9uYEdYUVa79Axb7Rh", false},
		{"valid artist URI", "spotify:artist:1301WleyT98MSxVHPZCA6M", "1301WleyT98MSxVHPZCA6M", false},
		{"invalid URI", "invalid:uri", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := validator.ExtractIDFromURI(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractIDFromURI() error = %v, wantErr %v", err, tt.wantErr)
			}
			if id != tt.expected {
				t.Errorf("ExtractIDFromURI() id = %v, expected %v", id, tt.expected)
			}
		})
	}
}

func TestValidator_ValidateMarket(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		market  string
		wantErr bool
	}{
		{"valid market", "US", false},
		{"from_token", "from_token", false},
		{"empty market", "", false}, // Optional
		{"invalid length", "USA", true},
		{"invalid characters", "U1", true},
		{"lowercase", "us", false}, // Should be accepted and normalized
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMarket(tt.market)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMarket() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateLimit(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		limit   int
		min     int
		max     int
		wantErr bool
	}{
		{"valid limit", 20, 1, 50, false},
		{"minimum limit", 1, 1, 50, false},
		{"maximum limit", 50, 1, 50, false},
		{"below minimum", 0, 1, 50, true},
		{"above maximum", 51, 1, 50, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateLimit(tt.limit, tt.min, tt.max)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLimit() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateOffset(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		offset  int
		wantErr bool
	}{
		{"valid offset", 0, false},
		{"positive offset", 100, false},
		{"negative offset", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOffset(tt.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOffset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateTimeRange(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name      string
		timeRange string
		wantErr   bool
	}{
		{"short_term", "short_term", false},
		{"medium_term", "medium_term", false},
		{"long_term", "long_term", false},
		{"empty", "", false}, // Optional
		{"invalid", "invalid_term", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTimeRange(tt.timeRange)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTimeRange() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateSearchQuery(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{"valid query", "test search", false},
		{"empty query", "", true},
		{"whitespace only", "   ", true},
		{"long query", strings.Repeat("a", 500), false},
		{"too long query", strings.Repeat("a", 501), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSearchQuery(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSearchQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateSearchTypes(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		types   []string
		wantErr bool
	}{
		{"valid types", []string{"track", "artist"}, false},
		{"all valid types", []string{"album", "artist", "playlist", "track", "show", "episode", "audiobook"}, false},
		{"empty types", []string{}, true},
		{"invalid type", []string{"track", "invalid"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSearchTypes(tt.types)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSearchTypes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidatePlaylistName(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		playlist string
		wantErr  bool
	}{
		{"valid name", "My Playlist", false},
		{"empty name", "", true},
		{"whitespace only", "   ", true},
		{"max length", strings.Repeat("a", 100), false},
		{"too long", strings.Repeat("a", 101), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePlaylistName(tt.playlist)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePlaylistName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateVolumePercent(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		volume  int
		wantErr bool
	}{
		{"minimum volume", 0, false},
		{"maximum volume", 100, false},
		{"mid volume", 50, false},
		{"below minimum", -1, true},
		{"above maximum", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateVolumePercent(tt.volume)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVolumePercent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_NormalizeAndValidateIDs(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		input    []string
		expected []string
		wantErr  bool
	}{
		{
			"valid IDs",
			[]string{"4iV5W9uYEdYUVa79Axb7Rh", "1301WleyT98MSxVHPZCA6M"},
			[]string{"4iV5W9uYEdYUVa79Axb7Rh", "1301WleyT98MSxVHPZCA6M"},
			false,
		},
		{
			"valid URIs",
			[]string{"spotify:track:4iV5W9uYEdYUVa79Axb7Rh", "spotify:artist:1301WleyT98MSxVHPZCA6M"},
			[]string{"4iV5W9uYEdYUVa79Axb7Rh", "1301WleyT98MSxVHPZCA6M"},
			false,
		},
		{
			"mixed IDs and URIs",
			[]string{"4iV5W9uYEdYUVa79Axb7Rh", "spotify:artist:1301WleyT98MSxVHPZCA6M"},
			[]string{"4iV5W9uYEdYUVa79Axb7Rh", "1301WleyT98MSxVHPZCA6M"},
			false,
		},
		{
			"empty input",
			[]string{},
			nil,
			true,
		},
		{
			"invalid ID",
			[]string{"invalid"},
			nil,
			true,
		},
		{
			"invalid URI",
			[]string{"spotify:track:invalid"},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.NormalizeAndValidateIDs(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeAndValidateIDs() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(result) != len(tt.expected) {
					t.Errorf("NormalizeAndValidateIDs() length = %v, expected %v", len(result), len(tt.expected))
				}
				for i, id := range result {
					if id != tt.expected[i] {
						t.Errorf("NormalizeAndValidateIDs() result[%d] = %v, expected %v", i, id, tt.expected[i])
					}
				}
			}
		})
	}
}