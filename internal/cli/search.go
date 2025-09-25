package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/cli/client"
	"github.com/bambithedeer/spotify-api/internal/cli/config"
	"github.com/bambithedeer/spotify-api/internal/cli/utils"
	"github.com/bambithedeer/spotify-api/internal/models"
	"github.com/spf13/cobra"
)

var (
	searchLimit   int
	searchOffset  int
	searchMarket  string
	searchFormat  string
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search Spotify catalog",
	Long: `Search for tracks, albums, artists, and playlists in the Spotify catalog.

Requires authentication with either user account or client credentials.
Use 'auth login' or 'auth client-credentials' to authenticate first.`,
	Example: `  # Search for tracks
  spotify-cli search track "bohemian rhapsody"
  spotify-cli search track "queen" --limit 10

  # Search for albums
  spotify-cli search album "a night at the opera"

  # Search for artists
  spotify-cli search artist "queen"

  # Search for playlists
  spotify-cli search playlist "rock classics"

  # Use different output formats
  spotify-cli search track "hello" --output json
  spotify-cli search album "abbey road" --format table`,
}

var searchTrackCmd = &cobra.Command{
	Use:   "track [query]",
	Short: "Search for tracks",
	Long: `Search for tracks in the Spotify catalog.

You can use Spotify's advanced search syntax:
  artist:queen          - Search by artist
  album:"a night"       - Search by album
  year:1975             - Search by year
  genre:rock            - Search by genre`,
	Args: cobra.ExactArgs(1),
	Example: `  spotify-cli search track "bohemian rhapsody"
  spotify-cli search track "artist:queen album:opera"
  spotify-cli search track "year:1970-1980 genre:rock"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearchTracks(args[0])
	},
}

var searchAlbumCmd = &cobra.Command{
	Use:   "album [query]",
	Short: "Search for albums",
	Long: `Search for albums in the Spotify catalog.

You can use Spotify's advanced search syntax:
  artist:queen          - Search by artist
  year:1975             - Search by year
  genre:rock            - Search by genre`,
	Args: cobra.ExactArgs(1),
	Example: `  spotify-cli search album "a night at the opera"
  spotify-cli search album "artist:queen year:1975"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearchAlbums(args[0])
	},
}

var searchArtistCmd = &cobra.Command{
	Use:   "artist [query]",
	Short: "Search for artists",
	Long: `Search for artists in the Spotify catalog.

You can use Spotify's advanced search syntax:
  genre:rock            - Search by genre
  year:1970-1980        - Search by active years`,
	Args: cobra.ExactArgs(1),
	Example: `  spotify-cli search artist "queen"
  spotify-cli search artist "genre:rock"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearchArtists(args[0])
	},
}

var searchPlaylistCmd = &cobra.Command{
	Use:   "playlist [query]",
	Short: "Search for playlists",
	Long:  `Search for playlists in the Spotify catalog.`,
	Args:  cobra.ExactArgs(1),
	Example: `  spotify-cli search playlist "rock classics"
  spotify-cli search playlist "workout"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearchPlaylists(args[0])
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.AddCommand(searchTrackCmd)
	searchCmd.AddCommand(searchAlbumCmd)
	searchCmd.AddCommand(searchArtistCmd)
	searchCmd.AddCommand(searchPlaylistCmd)

	// Add flags to all search commands
	for _, cmd := range []*cobra.Command{searchTrackCmd, searchAlbumCmd, searchArtistCmd, searchPlaylistCmd} {
		cmd.Flags().IntVarP(&searchLimit, "limit", "l", 20, "Number of results to return (1-50)")
		cmd.Flags().IntVarP(&searchOffset, "offset", "", 0, "Offset for pagination")
		cmd.Flags().StringVarP(&searchMarket, "market", "m", "", "Market/country code (e.g., US, GB)")
		cmd.Flags().StringVarP(&searchFormat, "format", "f", "table", "Output format (table, list, json, yaml)")
	}
}

func runSearchTracks(query string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' or 'spotify-cli auth client-credentials'")
	}

	// Create pagination options
	paginationOpts := &api.PaginationOptions{
		Limit:  searchLimit,
		Offset: searchOffset,
	}

	tracks, pagination, err := spotifyClient.Search.SearchTracks(GetCommandContext(), query, paginationOpts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	return outputSearchResults("tracks", tracks, pagination)
}

func runSearchAlbums(query string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' or 'spotify-cli auth client-credentials'")
	}

	// Create pagination options
	paginationOpts := &api.PaginationOptions{
		Limit:  searchLimit,
		Offset: searchOffset,
	}

	albums, pagination, err := spotifyClient.Search.SearchAlbums(GetCommandContext(), query, paginationOpts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	return outputSearchResults("albums", albums, pagination)
}

func runSearchArtists(query string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' or 'spotify-cli auth client-credentials'")
	}

	// Create pagination options
	paginationOpts := &api.PaginationOptions{
		Limit:  searchLimit,
		Offset: searchOffset,
	}

	artists, pagination, err := spotifyClient.Search.SearchArtists(GetCommandContext(), query, paginationOpts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	return outputSearchResults("artists", artists, pagination)
}

func runSearchPlaylists(query string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' or 'spotify-cli auth client-credentials'")
	}

	// Create pagination options
	paginationOpts := &api.PaginationOptions{
		Limit:  searchLimit,
		Offset: searchOffset,
	}

	playlists, pagination, err := spotifyClient.Search.SearchPlaylists(GetCommandContext(), query, paginationOpts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	return outputSearchResults("playlists", playlists, pagination)
}

func outputSearchResults(searchType string, results interface{}, pagination *api.PaginationInfo) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := searchFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output, return the data directly
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(map[string]interface{}{
			"results":    results,
			"pagination": pagination,
			"query_info": map[string]interface{}{
				"type":   searchType,
				"limit":  searchLimit,
				"offset": searchOffset,
				"market": searchMarket,
			},
		})
	}

	// Text-based output
	switch v := results.(type) {
	case *models.Paging[models.Track]:
		return outputTracksTable(v, pagination)
	case *models.Paging[models.Album]:
		return outputAlbumsTable(v, pagination)
	case *models.Paging[models.Artist]:
		return outputArtistsTable(v, pagination)
	case *models.Paging[models.Playlist]:
		return outputPlaylistsTable(v, pagination)
	default:
		return fmt.Errorf("unsupported result type")
	}
}

func outputTracksTable(tracks *models.Paging[models.Track], pagination *api.PaginationInfo) error {
	if len(tracks.Items) == 0 {
		fmt.Println("No tracks found.")
		return nil
	}

	// Print header
	fmt.Printf("Found %d tracks", tracks.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(tracks.Items))
	}
	fmt.Println()
	fmt.Println()

	// Print results based on format
	if searchFormat == "list" {
		for i, track := range tracks.Items {
			fmt.Printf("%d. %s\n", i+1, track.Name)
			fmt.Printf("   ID: %s\n", track.ID)
			if len(track.Artists) > 0 {
				artists := make([]string, len(track.Artists))
				for j, artist := range track.Artists {
					artists[j] = artist.Name
				}
				fmt.Printf("   by %s\n", strings.Join(artists, ", "))
			}
			if track.Album.Name != "" {
				fmt.Printf("   from %s\n", track.Album.Name)
			}
			if track.DurationMs > 0 {
				duration := formatTrackDuration(track.DurationMs)
				fmt.Printf("   ⏱ %s\n", duration)
			}
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-22s %-40s %-25s %-25s %s\n", "ID", "TRACK", "ARTIST", "ALBUM", "DURATION")
		fmt.Println(strings.Repeat("-", 130))

		for _, track := range tracks.Items {
			artists := "Unknown Artist"
			if len(track.Artists) > 0 {
				artistNames := make([]string, len(track.Artists))
				for i, artist := range track.Artists {
					artistNames[i] = artist.Name
				}
				artists = strings.Join(artistNames, ", ")
			}

			album := "Unknown Album"
			if track.Album.Name != "" {
				album = track.Album.Name
			}

			duration := ""
			if track.DurationMs > 0 {
				duration = formatTrackDuration(track.DurationMs)
			}

			fmt.Printf("%-22s %-40s %-25s %-25s %s\n",
				track.ID,
				truncateString(track.Name, 38),
				truncateString(artists, 23),
				truncateString(album, 23),
				duration)
		}
	}

	// Show pagination info
	if pagination != nil && pagination.HasNext() {
		fmt.Println()
		nextOffset := pagination.GetNextOffset()
		if nextOffset > 0 {
			fmt.Printf("Use --offset %d for next page\n", nextOffset)
		}
	}

	return nil
}

func outputAlbumsTable(albums *models.Paging[models.Album], pagination *api.PaginationInfo) error {
	if len(albums.Items) == 0 {
		fmt.Println("No albums found.")
		return nil
	}

	// Print header
	fmt.Printf("Found %d albums", albums.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(albums.Items))
	}
	fmt.Println()
	fmt.Println()

	if searchFormat == "list" {
		for i, album := range albums.Items {
			fmt.Printf("%d. %s\n", i+1, album.Name)
			fmt.Printf("   ID: %s\n", album.ID)
			if len(album.Artists) > 0 {
				artists := make([]string, len(album.Artists))
				for j, artist := range album.Artists {
					artists[j] = artist.Name
				}
				fmt.Printf("   by %s\n", strings.Join(artists, ", "))
			}
			if album.ReleaseDatePrecision.DateStr != "" {
				fmt.Printf("   released %s\n", album.ReleaseDatePrecision.DateStr)
			}
			if album.TotalTracks > 0 {
				fmt.Printf("   %d tracks\n", album.TotalTracks)
			}
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-22s %-30s %-25s %-12s %s\n", "ID", "ALBUM", "ARTIST", "RELEASED", "TRACKS")
		fmt.Println(strings.Repeat("-", 110))

		for _, album := range albums.Items {
			artists := "Unknown Artist"
			if len(album.Artists) > 0 {
				artistNames := make([]string, len(album.Artists))
				for i, artist := range album.Artists {
					artistNames[i] = artist.Name
				}
				artists = strings.Join(artistNames, ", ")
			}

			released := album.ReleaseDatePrecision.DateStr
			if len(released) > 10 {
				released = released[:10] // Just the date part
			}

			tracks := ""
			if album.TotalTracks > 0 {
				tracks = strconv.Itoa(album.TotalTracks)
			}

			fmt.Printf("%-22s %-30s %-25s %-12s %s\n",
				album.ID,
				truncateString(album.Name, 28),
				truncateString(artists, 23),
				released,
				tracks)
		}
	}

	// Show pagination info
	if pagination != nil && pagination.HasNext() {
		fmt.Println()
		nextOffset := pagination.GetNextOffset()
		if nextOffset > 0 {
			fmt.Printf("Use --offset %d for next page\n", nextOffset)
		}
	}

	return nil
}

func outputArtistsTable(artists *models.Paging[models.Artist], pagination *api.PaginationInfo) error {
	if len(artists.Items) == 0 {
		fmt.Println("No artists found.")
		return nil
	}

	// Print header
	fmt.Printf("Found %d artists", artists.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(artists.Items))
	}
	fmt.Println()
	fmt.Println()

	if searchFormat == "list" {
		for i, artist := range artists.Items {
			fmt.Printf("%d. %s\n", i+1, artist.Name)
			fmt.Printf("   ID: %s\n", artist.ID)
			if len(artist.Genres) > 0 {
				fmt.Printf("   genres: %s\n", strings.Join(artist.Genres, ", "))
			}
			if artist.Popularity > 0 {
				fmt.Printf("   popularity: %d/100\n", artist.Popularity)
			}
			if artist.Followers.Total > 0 {
				fmt.Printf("   followers: %s\n", formatNumber(artist.Followers.Total))
			}
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-22s %-30s %-25s %-12s %s\n", "ID", "ARTIST", "GENRES", "POPULARITY", "FOLLOWERS")
		fmt.Println(strings.Repeat("-", 110))

		for _, artist := range artists.Items {
			genres := strings.Join(artist.Genres, ", ")
			if len(genres) == 0 {
				genres = "-"
			}

			popularity := ""
			if artist.Popularity > 0 {
				popularity = fmt.Sprintf("%d/100", artist.Popularity)
			}

			followers := ""
			if artist.Followers.Total > 0 {
				followers = formatNumber(artist.Followers.Total)
			}

			fmt.Printf("%-22s %-30s %-25s %-12s %s\n",
				artist.ID,
				truncateString(artist.Name, 28),
				truncateString(genres, 23),
				popularity,
				followers)
		}
	}

	// Show pagination info
	if pagination != nil && pagination.HasNext() {
		fmt.Println()
		nextOffset := pagination.GetNextOffset()
		if nextOffset > 0 {
			fmt.Printf("Use --offset %d for next page\n", nextOffset)
		}
	}

	return nil
}

func outputPlaylistsTable(playlists *models.Paging[models.Playlist], pagination *api.PaginationInfo) error {
	if len(playlists.Items) == 0 {
		fmt.Println("No playlists found.")
		return nil
	}

	// Print header
	fmt.Printf("Found %d playlists", playlists.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(playlists.Items))
	}
	fmt.Println()
	fmt.Println()

	if searchFormat == "list" {
		for i, playlist := range playlists.Items {
			fmt.Printf("%d. %s\n", i+1, playlist.Name)
			fmt.Printf("   ID: %s\n", playlist.ID)
			if playlist.Owner.DisplayName != "" {
				fmt.Printf("   by %s\n", playlist.Owner.DisplayName)
			}
			if playlist.Description != "" {
				fmt.Printf("   %s\n", truncateString(playlist.Description, 80))
			}
			if playlist.Tracks.Total > 0 {
				fmt.Printf("   %d tracks\n", playlist.Tracks.Total)
			}
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-22s %-30s %-18s %-25s %s\n", "ID", "PLAYLIST", "OWNER", "DESCRIPTION", "TRACKS")
		fmt.Println(strings.Repeat("-", 115))

		for _, playlist := range playlists.Items {
			owner := playlist.Owner.DisplayName
			if owner == "" && playlist.Owner.ID != "" {
				owner = playlist.Owner.ID
			}

			description := playlist.Description
			if description == "" {
				description = "-"
			}

			tracks := ""
			if playlist.Tracks.Total > 0 {
				tracks = strconv.Itoa(playlist.Tracks.Total)
			}

			fmt.Printf("%-22s %-30s %-18s %-25s %s\n",
				playlist.ID,
				truncateString(playlist.Name, 28),
				truncateString(owner, 16),
				truncateString(description, 23),
				tracks)
		}
	}

	// Show pagination info
	if pagination != nil && pagination.HasNext() {
		fmt.Println()
		nextOffset := pagination.GetNextOffset()
		if nextOffset > 0 {
			fmt.Printf("Use --offset %d for next page\n", nextOffset)
		}
	}

	return nil
}

// Helper functions

func formatTrackDuration(durationMs int) string {
	seconds := durationMs / 1000
	minutes := seconds / 60
	remainingSeconds := seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
}

func formatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return strconv.Itoa(n)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}