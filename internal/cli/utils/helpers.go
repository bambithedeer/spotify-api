package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bambithedeer/spotify-api/internal/models"
)

// FormatDuration formats a duration in milliseconds to MM:SS format
func FormatDuration(ms int) string {
	duration := time.Duration(ms) * time.Millisecond
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// FormatArtists formats a list of artists as a comma-separated string
func FormatArtists(artists []models.Artist) string {
	names := make([]string, len(artists))
	for i, artist := range artists {
		names[i] = artist.Name
	}
	return strings.Join(names, ", ")
}

// FormatSimpleArtists formats a list of simple artists as a comma-separated string
func FormatSimpleArtists(artists []models.SimpleArtist) string {
	names := make([]string, len(artists))
	for i, artist := range artists {
		names[i] = artist.Name
	}
	return strings.Join(names, ", ")
}

// FormatTrack formats a track for display
func FormatTrack(track models.Track) string {
	artists := FormatSimpleArtists(track.Artists)
	duration := FormatDuration(track.DurationMs)
	return fmt.Sprintf("%s - %s (%s)", artists, track.Name, duration)
}

// FormatAlbum formats an album for display
func FormatAlbum(album models.Album) string {
	artists := FormatSimpleArtists(album.Artists)
	return fmt.Sprintf("%s - %s (%d tracks)", artists, album.Name, album.TotalTracks)
}

// FormatArtist formats an artist for display
func FormatArtist(artist models.Artist) string {
	genres := strings.Join(artist.Genres, ", ")
	if genres == "" {
		genres = "No genres"
	}
	return fmt.Sprintf("%s (Popularity: %d, Genres: %s)", artist.Name, artist.Popularity, genres)
}

// FormatPlaylist formats a playlist for display
func FormatPlaylist(playlist models.Playlist) string {
	return fmt.Sprintf("%s by %s (%d tracks)", playlist.Name, playlist.Owner.DisplayName, playlist.Tracks.Total)
}

// ParseLimit parses a limit string and validates it
func ParseLimit(limitStr string, defaultLimit, maxLimit int) (int, error) {
	if limitStr == "" {
		return defaultLimit, nil
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 0, fmt.Errorf("invalid limit: %s", limitStr)
	}

	if limit < 1 || limit > maxLimit {
		return 0, fmt.Errorf("limit must be between 1 and %d", maxLimit)
	}

	return limit, nil
}

// ParseOffset parses an offset string and validates it
func ParseOffset(offsetStr string) (int, error) {
	if offsetStr == "" {
		return 0, nil
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return 0, fmt.Errorf("invalid offset: %s", offsetStr)
	}

	if offset < 0 {
		return 0, fmt.Errorf("offset cannot be negative")
	}

	return offset, nil
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// FormatTable formats data as a simple table
func FormatTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var result strings.Builder

	// Print headers
	for i, header := range headers {
		result.WriteString(fmt.Sprintf("%-*s", colWidths[i]+2, header))
	}
	result.WriteString("\n")

	// Print separator
	for i := range headers {
		result.WriteString(strings.Repeat("-", colWidths[i]+2))
	}
	result.WriteString("\n")

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) {
				result.WriteString(fmt.Sprintf("%-*s", colWidths[i]+2, cell))
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}