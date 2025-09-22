package utils

import (
	"testing"

	"github.com/bambithedeer/spotify-api/internal/models"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		ms       int
		expected string
	}{
		{0, "0:00"},
		{1000, "0:01"},
		{60000, "1:00"},
		{90000, "1:30"},
		{3661000, "61:01"},
		{180000, "3:00"},
		{240000, "4:00"},
	}

	for _, test := range tests {
		result := FormatDuration(test.ms)
		if result != test.expected {
			t.Errorf("FormatDuration(%d) = %s, expected %s", test.ms, result, test.expected)
		}
	}
}

func TestFormatArtists(t *testing.T) {
	artists := []models.Artist{
		{Name: "Queen"},
		{Name: "David Bowie"},
		{Name: "Elton John"},
	}

	result := FormatArtists(artists)
	expected := "Queen, David Bowie, Elton John"

	if result != expected {
		t.Errorf("FormatArtists() = %s, expected %s", result, expected)
	}

	// Test empty slice
	result = FormatArtists([]models.Artist{})
	if result != "" {
		t.Errorf("FormatArtists([]) = %s, expected empty string", result)
	}

	// Test single artist
	result = FormatArtists([]models.Artist{{Name: "Queen"}})
	if result != "Queen" {
		t.Errorf("FormatArtists([Queen]) = %s, expected 'Queen'", result)
	}
}

func TestFormatSimpleArtists(t *testing.T) {
	artists := []models.SimpleArtist{
		{Name: "Queen"},
		{Name: "David Bowie"},
	}

	result := FormatSimpleArtists(artists)
	expected := "Queen, David Bowie"

	if result != expected {
		t.Errorf("FormatSimpleArtists() = %s, expected %s", result, expected)
	}
}

func TestParseLimit(t *testing.T) {
	tests := []struct {
		limitStr    string
		defaultVal  int
		maxLimit    int
		expected    int
		shouldError bool
	}{
		{"", 20, 50, 20, false},      // Default value
		{"10", 20, 50, 10, false},    // Valid limit
		{"50", 20, 50, 50, false},    // Max limit
		{"0", 20, 50, 0, true},       // Below minimum
		{"51", 20, 50, 0, true},      // Above maximum
		{"invalid", 20, 50, 0, true}, // Invalid format
		{"-5", 20, 50, 0, true},      // Negative
	}

	for _, test := range tests {
		result, err := ParseLimit(test.limitStr, test.defaultVal, test.maxLimit)

		if test.shouldError {
			if err == nil {
				t.Errorf("ParseLimit(%s) expected error, got nil", test.limitStr)
			}
		} else {
			if err != nil {
				t.Errorf("ParseLimit(%s) unexpected error: %v", test.limitStr, err)
			}
			if result != test.expected {
				t.Errorf("ParseLimit(%s) = %d, expected %d", test.limitStr, result, test.expected)
			}
		}
	}
}

func TestParseOffset(t *testing.T) {
	tests := []struct {
		offsetStr   string
		expected    int
		shouldError bool
	}{
		{"", 0, false},       // Default value
		{"10", 10, false},    // Valid offset
		{"0", 0, false},      // Zero offset
		{"-1", 0, true},      // Negative offset
		{"invalid", 0, true}, // Invalid format
	}

	for _, test := range tests {
		result, err := ParseOffset(test.offsetStr)

		if test.shouldError {
			if err == nil {
				t.Errorf("ParseOffset(%s) expected error, got nil", test.offsetStr)
			}
		} else {
			if err != nil {
				t.Errorf("ParseOffset(%s) unexpected error: %v", test.offsetStr, err)
			}
			if result != test.expected {
				t.Errorf("ParseOffset(%s) = %d, expected %d", test.offsetStr, result, test.expected)
			}
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"Hello World", 20, "Hello World"},               // Shorter than max
		{"Hello World", 11, "Hello World"},               // Equal to max
		{"Hello World", 10, "Hello W..."},                // Longer than max
		{"Hi", 5, "Hi"},                                  // Much shorter
		{"", 5, ""},                                      // Empty string
		{"This is a very long string", 10, "This is..."}, // Much longer
	}

	for _, test := range tests {
		result := TruncateString(test.input, test.maxLen)
		if result != test.expected {
			t.Errorf("TruncateString(%q, %d) = %q, expected %q", test.input, test.maxLen, result, test.expected)
		}
	}
}

func TestFormatTable(t *testing.T) {
	headers := []string{"Name", "Artist", "Duration"}
	rows := [][]string{
		{"Bohemian Rhapsody", "Queen", "5:55"},
		{"Stairway to Heaven", "Led Zeppelin", "8:02"},
		{"Hotel California", "Eagles", "6:30"},
	}

	result := FormatTable(headers, rows)

	// Check that the result contains the headers
	if !containsString(result, "Name") {
		t.Error("Expected table to contain 'Name' header")
	}

	if !containsString(result, "Artist") {
		t.Error("Expected table to contain 'Artist' header")
	}

	if !containsString(result, "Duration") {
		t.Error("Expected table to contain 'Duration' header")
	}

	// Check that the result contains the data
	if !containsString(result, "Bohemian Rhapsody") {
		t.Error("Expected table to contain 'Bohemian Rhapsody'")
	}

	if !containsString(result, "Queen") {
		t.Error("Expected table to contain 'Queen'")
	}

	// Test empty rows
	emptyResult := FormatTable(headers, [][]string{})
	if emptyResult != "" {
		t.Errorf("Expected empty table for no rows, got: %s", emptyResult)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsStringHelper(s, substr)))
}

func containsStringHelper(s, substr string) bool {
	for i := 1; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
