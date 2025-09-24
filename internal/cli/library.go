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
	"github.com/bambithedeer/spotify-api/internal/spotify"
	"github.com/spf13/cobra"
)

var (
	libraryLimit   int
	libraryOffset  int
	libraryMarket  string
	libraryFormat  string
)

// libraryCmd represents the library command
var libraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Manage your Spotify library",
	Long: `Manage your saved tracks, albums, and library content.

Requires user authentication. Use 'auth login' to authenticate with user account first.
Client credentials authentication does not provide access to user library data.`,
	Example: `  # List saved tracks
  spotify-cli library tracks

  # List saved albums
  spotify-cli library albums

  # List followed artists
  spotify-cli library follows

  # Save tracks to library
  spotify-cli library save track <track-id> [track-id...]

  # Remove albums from library
  spotify-cli library remove album <album-id> [album-id...]

  # Check if tracks are saved
  spotify-cli library check track <track-id> [track-id...]`,
}

var libraryTracksCmd = &cobra.Command{
	Use:   "tracks",
	Short: "List saved tracks",
	Long:  `List tracks saved in your Spotify library.`,
	Example: `  spotify-cli library tracks
  spotify-cli library tracks --limit 50
  spotify-cli library tracks --format list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLibraryTracks()
	},
}

var libraryAlbumsCmd = &cobra.Command{
	Use:   "albums",
	Short: "List saved albums",
	Long:  `List albums saved in your Spotify library.`,
	Example: `  spotify-cli library albums
  spotify-cli library albums --limit 20 --market US`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLibraryAlbums()
	},
}

var librarySaveCmd = &cobra.Command{
	Use:   "save [type] [id...]",
	Short: "Save tracks or albums to library",
	Long: `Save one or more tracks or albums to your Spotify library.

Type must be either 'track' or 'album'.
You can provide multiple IDs to save multiple items at once (up to 50).`,
	Args: cobra.MinimumNArgs(2),
	Example: `  spotify-cli library save track 4iV5W9uYEdYUVa79Axb7Rh
  spotify-cli library save album 1DFixLWuPkv3KT3TnV35m3
  spotify-cli library save track id1 id2 id3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLibrarySave(args[0], args[1:])
	},
}

var libraryRemoveCmd = &cobra.Command{
	Use:   "remove [type] [id...]",
	Short: "Remove tracks or albums from library",
	Long: `Remove one or more tracks or albums from your Spotify library.

Type must be either 'track' or 'album'.
You can provide multiple IDs to remove multiple items at once (up to 50).`,
	Args: cobra.MinimumNArgs(2),
	Example: `  spotify-cli library remove track 4iV5W9uYEdYUVa79Axb7Rh
  spotify-cli library remove album 1DFixLWuPkv3KT3TnV35m3
  spotify-cli library remove track id1 id2 id3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLibraryRemove(args[0], args[1:])
	},
}

var libraryCheckCmd = &cobra.Command{
	Use:   "check [type] [id...]",
	Short: "Check if tracks or albums are saved",
	Long: `Check whether one or more tracks or albums are saved in your library.

Type must be either 'track' or 'album'.
You can check multiple IDs at once (up to 50).`,
	Args: cobra.MinimumNArgs(2),
	Example: `  spotify-cli library check track 4iV5W9uYEdYUVa79Axb7Rh
  spotify-cli library check album 1DFixLWuPkv3KT3TnV35m3
  spotify-cli library check track id1 id2 id3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLibraryCheck(args[0], args[1:])
	},
}

var libraryFollowsCmd = &cobra.Command{
	Use:   "follows",
	Short: "List followed artists",
	Long: `List artists you are following on Spotify.

Note: Spotify's Web API currently only supports retrieving followed artists.
Following users is not supported by the API at this time.`,
	Example: `  spotify-cli library follows
  spotify-cli library follows --limit 50`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runLibraryFollows()
	},
}

func init() {
	rootCmd.AddCommand(libraryCmd)
	libraryCmd.AddCommand(libraryTracksCmd)
	libraryCmd.AddCommand(libraryAlbumsCmd)
	libraryCmd.AddCommand(librarySaveCmd)
	libraryCmd.AddCommand(libraryRemoveCmd)
	libraryCmd.AddCommand(libraryCheckCmd)
	libraryCmd.AddCommand(libraryFollowsCmd)

	// Add flags to list commands
	for _, cmd := range []*cobra.Command{libraryTracksCmd, libraryAlbumsCmd, libraryFollowsCmd} {
		cmd.Flags().IntVarP(&libraryLimit, "limit", "l", 20, "Number of results to return (1-50)")
		cmd.Flags().IntVarP(&libraryOffset, "offset", "", 0, "Offset for pagination")
		cmd.Flags().StringVarP(&libraryMarket, "market", "m", "", "Market/country code (e.g., US, GB)")
		cmd.Flags().StringVarP(&libraryFormat, "format", "f", "table", "Output format (table, list, json, yaml)")
	}
}

func runLibraryTracks() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	// Check if we're using client credentials (which don't have user scope access)
	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access your personal library")
	}

	// Create pagination options
	paginationOpts := &api.PaginationOptions{
		Limit:  libraryLimit,
		Offset: libraryOffset,
	}

	tracks, pagination, err := spotifyClient.Library.GetSavedTracks(GetCommandContext(), paginationOpts)
	if err != nil {
		return fmt.Errorf("failed to get saved tracks: %w", err)
	}

	return outputLibraryResults("saved tracks", tracks, pagination)
}

func runLibraryAlbums() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	// Create options
	options := &spotify.SavedAlbumsOptions{
		Market: libraryMarket,
		Limit:  libraryLimit,
		Offset: libraryOffset,
	}

	albums, pagination, err := spotifyClient.Library.GetSavedAlbums(GetCommandContext(), options)
	if err != nil {
		return fmt.Errorf("failed to get saved albums: %w", err)
	}

	return outputLibraryResults("saved albums", albums, pagination)
}

func runLibrarySave(itemType string, ids []string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	if len(ids) > 50 {
		return fmt.Errorf("cannot save more than 50 items at once")
	}

	switch itemType {
	case "track", "tracks":
		err = spotifyClient.Library.SaveTracks(GetCommandContext(), ids)
		if err != nil {
			return fmt.Errorf("failed to save tracks: %w", err)
		}
		utils.PrintSuccess(fmt.Sprintf("Successfully saved %d track(s) to library", len(ids)))

	case "album", "albums":
		err = spotifyClient.Library.SaveAlbums(GetCommandContext(), ids)
		if err != nil {
			return fmt.Errorf("failed to save albums: %w", err)
		}
		utils.PrintSuccess(fmt.Sprintf("Successfully saved %d album(s) to library", len(ids)))

	default:
		return fmt.Errorf("invalid type '%s'. Must be 'track' or 'album'", itemType)
	}

	return nil
}

func runLibraryRemove(itemType string, ids []string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	if len(ids) > 50 {
		return fmt.Errorf("cannot remove more than 50 items at once")
	}

	switch itemType {
	case "track", "tracks":
		err = spotifyClient.Library.RemoveTracks(GetCommandContext(), ids)
		if err != nil {
			return fmt.Errorf("failed to remove tracks: %w", err)
		}
		utils.PrintSuccess(fmt.Sprintf("Successfully removed %d track(s) from library", len(ids)))

	case "album", "albums":
		err = spotifyClient.Library.RemoveAlbums(GetCommandContext(), ids)
		if err != nil {
			return fmt.Errorf("failed to remove albums: %w", err)
		}
		utils.PrintSuccess(fmt.Sprintf("Successfully removed %d album(s) from library", len(ids)))

	default:
		return fmt.Errorf("invalid type '%s'. Must be 'track' or 'album'", itemType)
	}

	return nil
}

func runLibraryCheck(itemType string, ids []string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	if len(ids) > 50 {
		return fmt.Errorf("cannot check more than 50 items at once")
	}

	var saved []bool
	var checkType string

	switch itemType {
	case "track", "tracks":
		saved, err = spotifyClient.Library.CheckSavedTracks(GetCommandContext(), ids)
		checkType = "track"
	case "album", "albums":
		saved, err = spotifyClient.Library.CheckSavedAlbums(GetCommandContext(), ids)
		checkType = "album"
	default:
		return fmt.Errorf("invalid type '%s'. Must be 'track' or 'album'", itemType)
	}

	if err != nil {
		return fmt.Errorf("failed to check saved %ss: %w", checkType, err)
	}

	return outputLibraryCheckResults(checkType, ids, saved)
}

func outputLibraryResults(libraryType string, results interface{}, pagination *api.PaginationInfo) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := libraryFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output, return the data directly
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(map[string]interface{}{
			"results":    results,
			"pagination": pagination,
			"library_info": map[string]interface{}{
				"type":   libraryType,
				"limit":  libraryLimit,
				"offset": libraryOffset,
				"market": libraryMarket,
			},
		})
	}

	// Text-based output
	switch v := results.(type) {
	case *models.Paging[models.SavedTrack]:
		return outputSavedTracksTable(v, pagination)
	case *models.Paging[models.SavedAlbum]:
		return outputSavedAlbumsTable(v, pagination)
	default:
		return fmt.Errorf("unsupported result type")
	}
}

func outputSavedTracksTable(savedTracks *models.Paging[models.SavedTrack], pagination *api.PaginationInfo) error {
	if len(savedTracks.Items) == 0 {
		fmt.Println("No saved tracks found.")
		return nil
	}

	// Print header
	fmt.Printf("Your Saved Tracks - %d total", savedTracks.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(savedTracks.Items))
	}
	fmt.Println()
	fmt.Println()

	if libraryFormat == "list" {
		for i, savedTrack := range savedTracks.Items {
			track := savedTrack.Track
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
			if savedTrack.AddedAt != "" {
				fmt.Printf("   📅 Added %s\n", formatDate(savedTrack.AddedAt))
			}
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-22s %-40s %-25s %-25s %-8s %s\n", "ID", "TRACK", "ARTIST", "ALBUM", "DURATION", "ADDED")
		fmt.Println(strings.Repeat("-", 140))

		for _, savedTrack := range savedTracks.Items {
			track := savedTrack.Track
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

			added := ""
			if savedTrack.AddedAt != "" {
				added = formatDate(savedTrack.AddedAt)
			}

			fmt.Printf("%-22s %-40s %-25s %-25s %-8s %s\n",
				track.ID,
				truncateString(track.Name, 38),
				truncateString(artists, 23),
				truncateString(album, 23),
				duration,
				added)
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

func outputSavedAlbumsTable(savedAlbums *models.Paging[models.SavedAlbum], pagination *api.PaginationInfo) error {
	if len(savedAlbums.Items) == 0 {
		fmt.Println("No saved albums found.")
		return nil
	}

	// Print header
	fmt.Printf("Your Saved Albums - %d total", savedAlbums.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(savedAlbums.Items))
	}
	fmt.Println()
	fmt.Println()

	if libraryFormat == "list" {
		for i, savedAlbum := range savedAlbums.Items {
			album := savedAlbum.Album
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
			if savedAlbum.AddedAt != "" {
				fmt.Printf("   📅 Added %s\n", formatDate(savedAlbum.AddedAt))
			}
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-22s %-30s %-25s %-12s %-6s %s\n", "ID", "ALBUM", "ARTIST", "RELEASED", "TRACKS", "ADDED")
		fmt.Println(strings.Repeat("-", 120))

		for _, savedAlbum := range savedAlbums.Items {
			album := savedAlbum.Album
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

			added := ""
			if savedAlbum.AddedAt != "" {
				added = formatDate(savedAlbum.AddedAt)
			}

			fmt.Printf("%-22s %-30s %-25s %-12s %-6s %s\n",
				album.ID,
				truncateString(album.Name, 28),
				truncateString(artists, 23),
				released,
				tracks,
				added)
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

func outputLibraryCheckResults(itemType string, ids []string, saved []bool) error {
	cfg := config.Get()

	// For structured output
	if cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml" {
		results := make([]map[string]interface{}, len(ids))
		for i, id := range ids {
			results[i] = map[string]interface{}{
				"id":    id,
				"saved": saved[i],
			}
		}
		return utils.Output(map[string]interface{}{
			"type":    itemType,
			"results": results,
		})
	}

	// Text output
	fmt.Printf("Library Check Results - %s%s\n", itemType, pluralize(len(ids)))
	fmt.Println(strings.Repeat("-", 60))

	for i, id := range ids {
		status := "❌ Not saved"
		if saved[i] {
			status = "✅ Saved"
		}
		fmt.Printf("%-40s %s\n", truncateString(id, 38), status)
	}

	// Summary
	savedCount := 0
	for _, isSaved := range saved {
		if isSaved {
			savedCount++
		}
	}
	fmt.Printf("\nSummary: %d/%d %s%s saved in library\n",
		savedCount, len(ids), itemType, pluralize(len(ids)))

	return nil
}

// Helper functions

func formatDate(dateStr string) string {
	// Simple date formatting - just take the date part if it's ISO format
	if len(dateStr) >= 10 {
		return dateStr[:10]
	}
	return dateStr
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func runLibraryFollows() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	// Check if we're using client credentials (which don't have user scope access)
	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access your followed artists")
	}

	// Create options for getting followed artists
	options := &spotify.FollowedArtistsOptions{
		Limit: libraryLimit,
	}

	followedArtists, err := spotifyClient.Users.GetFollowedArtists(GetCommandContext(), options)
	if err != nil {
		return fmt.Errorf("failed to get followed artists: %w", err)
	}

	return outputFollowedArtists(followedArtists)
}

func outputFollowedArtists(followedArtists *models.CursorPaging[models.Artist]) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := libraryFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output, return the data directly
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(map[string]interface{}{
			"followed_artists": followedArtists,
		})
	}

	// Text-based output
	if len(followedArtists.Items) == 0 {
		fmt.Println("You are not following any artists.")
		return nil
	}

	// Print header
	fmt.Printf("Followed Artists - %d total\n", len(followedArtists.Items))
	fmt.Println()

	if libraryFormat == "list" {
		for i, artist := range followedArtists.Items {
			fmt.Printf("%d. %s\n", i+1, artist.Name)
			fmt.Printf("   ID: %s\n", artist.ID)
			if artist.Followers.Total > 0 {
				fmt.Printf("   %d followers\n", artist.Followers.Total)
			}
			if len(artist.Genres) > 0 {
				fmt.Printf("   Genres: %s\n", strings.Join(artist.Genres, ", "))
			}
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-22s %-30s %-15s %s\n", "ID", "ARTIST", "FOLLOWERS", "GENRES")
		fmt.Println(strings.Repeat("-", 90))

		for _, artist := range followedArtists.Items {
			followers := ""
			if artist.Followers.Total > 0 {
				followers = strconv.Itoa(artist.Followers.Total)
			}

			genres := "—"
			if len(artist.Genres) > 0 {
				genres = strings.Join(artist.Genres, ", ")
			}

			fmt.Printf("%-22s %-30s %-15s %s\n",
				artist.ID,
				truncateString(artist.Name, 28),
				followers,
				truncateString(genres, 35))
		}
	}

	// Show cursor info for next page if available
	if followedArtists.Next != "" {
		fmt.Println()
		fmt.Println("Use cursor pagination for more results (feature not yet implemented in CLI)")
	}

	return nil
}

