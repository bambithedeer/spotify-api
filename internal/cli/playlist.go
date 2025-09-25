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
	playlistLimit  int
	playlistOffset int
	playlistFormat string
	playlistPublic bool
	playlistDesc   string
)

// playlistCmd represents the playlist command
var playlistCmd = &cobra.Command{
	Use:   "playlist",
	Short: "Manage your playlists",
	Long: `Create, list, and manage your Spotify playlists.

Requires user authentication. Use 'auth login' to authenticate with user account first.
Client credentials authentication does not provide access to user playlists.`,
	Example: `  # List your playlists
  spotify-cli playlist list

  # Get playlist details
  spotify-cli playlist get <playlist-id>

  # List tracks in a playlist
  spotify-cli playlist tracks <playlist-id>

  # Create a new playlist
  spotify-cli playlist create "My Playlist" --description "My awesome playlist"

  # Add tracks to playlist
  spotify-cli playlist add <playlist-id> <track-id> [track-id...]

  # Remove tracks from playlist
  spotify-cli playlist remove <playlist-id> <track-id> [track-id...]`,
}

var playlistListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your playlists",
	Long:  `List all playlists owned by the current user.`,
	Example: `  spotify-cli playlist list
  spotify-cli playlist list --limit 50`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlaylistList()
	},
}

var playlistGetCmd = &cobra.Command{
	Use:     "get [playlist-id]",
	Short:   "Get playlist details",
	Long:    `Get detailed information about a specific playlist.`,
	Args:    cobra.ExactArgs(1),
	Example: `  spotify-cli playlist get 37i9dQZF1DXcBWIGoYBM5M`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlaylistGet(args[0])
	},
}

var playlistCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new playlist",
	Long:  `Create a new empty playlist.`,
	Args:  cobra.ExactArgs(1),
	Example: `  spotify-cli playlist create "My Playlist"
  spotify-cli playlist create "Rock Hits" --description "Best rock songs" --public`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlaylistCreate(args[0])
	},
}

var playlistAddCmd = &cobra.Command{
	Use:   "add [playlist-id] [track-id...]",
	Short: "Add tracks to playlist",
	Long: `Add one or more tracks to a playlist.

You can provide multiple track IDs to add multiple tracks at once (up to 100).`,
	Args: cobra.MinimumNArgs(2),
	Example: `  spotify-cli playlist add 37i9dQZF1DXcBWIGoYBM5M 4iV5W9uYEdYUVa79Axb7Rh
  spotify-cli playlist add playlist-id track1 track2 track3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlaylistAdd(args[0], args[1:])
	},
}

var playlistRemoveCmd = &cobra.Command{
	Use:   "remove [playlist-id] [track-id...]",
	Short: "Remove tracks from playlist",
	Long: `Remove one or more tracks from a playlist.

You can provide multiple track IDs to remove multiple tracks at once (up to 100).`,
	Args: cobra.MinimumNArgs(2),
	Example: `  spotify-cli playlist remove 37i9dQZF1DXcBWIGoYBM5M 4iV5W9uYEdYUVa79Axb7Rh
  spotify-cli playlist remove playlist-id track1 track2 track3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlaylistRemove(args[0], args[1:])
	},
}

var playlistTracksCmd = &cobra.Command{
	Use:   "tracks [playlist-id]",
	Short: "List tracks in a playlist",
	Long: `List all tracks in a specific playlist.

Shows track details including ID, name, artist, album, and duration.
Works with both your own playlists and public playlists from other users.`,
	Args: cobra.ExactArgs(1),
	Example: `  spotify-cli playlist tracks 37i9dQZF1DXcBWIGoYBM5M
  spotify-cli playlist tracks 6pHeFS94QibtA0qCcAO2Iv --limit 50
  spotify-cli playlist tracks playlist-id --format list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlaylistTracks(args[0])
	},
}

func init() {
	rootCmd.AddCommand(playlistCmd)
	playlistCmd.AddCommand(playlistListCmd)
	playlistCmd.AddCommand(playlistGetCmd)
	playlistCmd.AddCommand(playlistCreateCmd)
	playlistCmd.AddCommand(playlistAddCmd)
	playlistCmd.AddCommand(playlistRemoveCmd)
	playlistCmd.AddCommand(playlistTracksCmd)

	// Add flags to list commands
	for _, cmd := range []*cobra.Command{playlistListCmd, playlistGetCmd, playlistTracksCmd} {
		cmd.Flags().IntVarP(&playlistLimit, "limit", "l", 20, "Number of results to return (1-50)")
		cmd.Flags().IntVarP(&playlistOffset, "offset", "", 0, "Offset for pagination")
		cmd.Flags().StringVarP(&playlistFormat, "format", "f", "table", "Output format (table, list, json, yaml)")
	}

	// Create playlist flags
	playlistCreateCmd.Flags().StringVarP(&playlistDesc, "description", "d", "", "Playlist description")
	playlistCreateCmd.Flags().BoolVarP(&playlistPublic, "public", "p", false, "Make playlist public")
}

func runPlaylistList() error {
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
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access your playlists")
	}

	// Create pagination options
	paginationOpts := &api.PaginationOptions{
		Limit:  playlistLimit,
		Offset: playlistOffset,
	}

	playlists, pagination, err := spotifyClient.Playlists.GetUserPlaylists(GetCommandContext(), paginationOpts)
	if err != nil {
		return fmt.Errorf("failed to get playlists: %w", err)
	}

	return outputPlaylistResults("your playlists", playlists, pagination)
}

func runPlaylistGet(playlistID string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	playlist, err := spotifyClient.Playlists.GetPlaylist(GetCommandContext(), playlistID, nil)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}

	return outputSinglePlaylist(playlist)
}

func runPlaylistCreate(name string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	// Get current user to create playlist
	user, err := spotifyClient.Users.GetCurrentUser(GetCommandContext())
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// Create playlist request
	request := &spotify.CreatePlaylistRequest{
		Name:        name,
		Description: playlistDesc,
		Public:      &playlistPublic,
	}

	playlist, err := spotifyClient.Playlists.CreatePlaylist(GetCommandContext(), user.ID, request)
	if err != nil {
		return fmt.Errorf("failed to create playlist: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("Created playlist: %s", playlist.Name))
	fmt.Printf("Playlist ID: %s\n", playlist.ID)
	fmt.Printf("Public: %t\n", playlist.Public)
	if playlist.Description != "" {
		fmt.Printf("Description: %s\n", playlist.Description)
	}

	return nil
}

func runPlaylistAdd(playlistID string, trackIDs []string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	if len(trackIDs) > 100 {
		return fmt.Errorf("cannot add more than 100 tracks at once")
	}

	// Convert track IDs to URIs
	trackURIs := make([]string, len(trackIDs))
	for i, id := range trackIDs {
		trackURIs[i] = fmt.Sprintf("spotify:track:%s", id)
	}

	request := &spotify.AddTracksRequest{
		URIs: trackURIs,
	}

	_, err = spotifyClient.Playlists.AddTracksToPlaylist(GetCommandContext(), playlistID, request)
	if err != nil {
		return fmt.Errorf("failed to add tracks to playlist: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("Successfully added %d track(s) to playlist", len(trackIDs)))
	return nil
}

func runPlaylistRemove(playlistID string, trackIDs []string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	if len(trackIDs) > 100 {
		return fmt.Errorf("cannot remove more than 100 tracks at once")
	}

	// Convert track IDs to track removal objects
	tracks := make([]spotify.TrackToRemove, len(trackIDs))
	for i, id := range trackIDs {
		tracks[i] = spotify.TrackToRemove{
			URI: fmt.Sprintf("spotify:track:%s", id),
		}
	}

	request := &spotify.RemoveTracksRequest{
		Tracks: tracks,
	}

	_, err = spotifyClient.Playlists.RemoveTracksFromPlaylist(GetCommandContext(), playlistID, request)
	if err != nil {
		return fmt.Errorf("failed to remove tracks from playlist: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("Successfully removed %d track(s) from playlist", len(trackIDs)))
	return nil
}

func outputPlaylistResults(playlistType string, results interface{}, pagination *api.PaginationInfo) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := playlistFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output, return the data directly
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(map[string]interface{}{
			"results":    results,
			"pagination": pagination,
			"playlist_info": map[string]interface{}{
				"type":   playlistType,
				"limit":  playlistLimit,
				"offset": playlistOffset,
			},
		})
	}

	// Text-based output
	switch v := results.(type) {
	case *models.Paging[models.Playlist]:
		return outputUserPlaylistsTable(v, pagination)
	default:
		return fmt.Errorf("unsupported result type")
	}
}

func outputUserPlaylistsTable(playlists *models.Paging[models.Playlist], pagination *api.PaginationInfo) error {
	if len(playlists.Items) == 0 {
		fmt.Println("No playlists found.")
		return nil
	}

	// Print header
	fmt.Printf("Your Playlists - %d total", playlists.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(playlists.Items))
	}
	fmt.Println()
	fmt.Println()

	if playlistFormat == "list" {
		for i, playlist := range playlists.Items {
			fmt.Printf("%d. %s\n", i+1, playlist.Name)
			if playlist.Owner.DisplayName != "" {
				fmt.Printf("   by %s\n", playlist.Owner.DisplayName)
			}
			if playlist.Description != "" {
				fmt.Printf("   %s\n", truncateString(playlist.Description, 80))
			}
			if playlist.Tracks.Total > 0 {
				fmt.Printf("   %d tracks\n", playlist.Tracks.Total)
			}
			fmt.Printf("   %s\n", map[bool]string{true: "Public", false: "Private"}[playlist.Public])
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-22s %-30s %-18s %-25s %-6s %s\n", "ID", "PLAYLIST", "OWNER", "DESCRIPTION", "TRACKS", "PUBLIC")
		fmt.Println(strings.Repeat("-", 125))

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

			public := "No"
			if playlist.Public {
				public = "Yes"
			}

			fmt.Printf("%-22s %-30s %-18s %-25s %-6s %s\n",
				playlist.ID,
				truncateString(playlist.Name, 28),
				truncateString(owner, 16),
				truncateString(description, 23),
				tracks,
				public)
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

func outputSinglePlaylist(playlist *models.Playlist) error {
	cfg := config.Get()

	// For structured output
	if cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml" {
		return utils.Output(playlist)
	}

	// Text output
	fmt.Printf("Playlist: %s\n", playlist.Name)
	fmt.Printf("ID: %s\n", playlist.ID)
	if playlist.Owner.DisplayName != "" {
		fmt.Printf("Owner: %s\n", playlist.Owner.DisplayName)
	} else if playlist.Owner.ID != "" {
		fmt.Printf("Owner: %s\n", playlist.Owner.ID)
	}
	if playlist.Description != "" {
		fmt.Printf("Description: %s\n", playlist.Description)
	}
	fmt.Printf("Tracks: %d\n", playlist.Tracks.Total)
	fmt.Printf("Public: %t\n", playlist.Public)
	fmt.Printf("Collaborative: %t\n", playlist.Collaborative)
	if playlist.ExternalURLs.Spotify != "" {
		fmt.Printf("Spotify URL: %s\n", playlist.ExternalURLs.Spotify)
	}

	return nil
}

func runPlaylistTracks(playlistID string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' or 'spotify-cli auth client-credentials'")
	}

	// Create playlist tracks options
	options := &spotify.PlaylistTracksOptions{
		Limit:  playlistLimit,
		Offset: playlistOffset,
	}

	tracks, pagination, err := spotifyClient.Playlists.GetPlaylistTracks(GetCommandContext(), playlistID, options)
	if err != nil {
		return fmt.Errorf("failed to get playlist tracks: %w", err)
	}

	return outputPlaylistTracks(playlistID, tracks, pagination)
}

func outputPlaylistTracks(playlistID string, tracks *models.Paging[models.PlaylistTrack], pagination *api.PaginationInfo) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := playlistFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output, return the data directly
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(map[string]interface{}{
			"playlist_id": playlistID,
			"results":     tracks,
			"pagination":  pagination,
		})
	}

	// Text-based output
	if len(tracks.Items) == 0 {
		fmt.Println("No tracks found in this playlist.")
		return nil
	}

	// Print header
	fmt.Printf("Playlist Tracks - %d total", tracks.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(tracks.Items))
	}
	fmt.Println()
	fmt.Println()

	if playlistFormat == "list" {
		for i, playlistTrack := range tracks.Items {
			if playlistTrack.Track == nil {
				fmt.Printf("%d. [Unavailable Track]\n", i+1)
				continue
			}

			// Type assert to Track
			track, ok := playlistTrack.Track.(map[string]interface{})
			if !ok {
				fmt.Printf("%d. [Invalid Track Data]\n", i+1)
				continue
			}

			name, _ := track["name"].(string)
			fmt.Printf("%d. %s\n", i+1, name)

			if id, ok := track["id"].(string); ok {
				fmt.Printf("   ID: %s\n", id)
			}

			if artistsInterface, ok := track["artists"].([]interface{}); ok && len(artistsInterface) > 0 {
				artistNames := make([]string, 0, len(artistsInterface))
				for _, artistInterface := range artistsInterface {
					if artist, ok := artistInterface.(map[string]interface{}); ok {
						if artistName, ok := artist["name"].(string); ok {
							artistNames = append(artistNames, artistName)
						}
					}
				}
				if len(artistNames) > 0 {
					fmt.Printf("   by %s\n", strings.Join(artistNames, ", "))
				}
			}

			if albumInterface, ok := track["album"].(map[string]interface{}); ok {
				if albumName, ok := albumInterface["name"].(string); ok && albumName != "" {
					fmt.Printf("   from %s\n", albumName)
				}
			}

			if durationMs, ok := track["duration_ms"].(float64); ok && durationMs > 0 {
				fmt.Printf("   Duration: %s\n", formatTrackDuration(int(durationMs)))
			}
			if playlistTrack.AddedAt != "" {
				fmt.Printf("   Added: %s\n", formatDate(playlistTrack.AddedAt))
			}
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-22s %-40s %-25s %-25s %-8s %s\n", "ID", "TRACK", "ARTIST", "ALBUM", "DURATION", "ADDED")
		fmt.Println(strings.Repeat("-", 145))

		for _, playlistTrack := range tracks.Items {
			if playlistTrack.Track == nil {
				fmt.Printf("%-22s %-40s %-25s %-25s %-8s %s\n",
					"—", "[Unavailable Track]", "—", "—", "—", "—")
				continue
			}

			// Type assert to Track
			track, ok := playlistTrack.Track.(map[string]interface{})
			if !ok {
				fmt.Printf("%-22s %-40s %-25s %-25s %-8s %s\n",
					"—", "[Invalid Track Data]", "—", "—", "—", "—")
				continue
			}

			// Extract track data
			trackID, _ := track["id"].(string)
			trackName, _ := track["name"].(string)

			artists := "Unknown Artist"
			if artistsInterface, ok := track["artists"].([]interface{}); ok && len(artistsInterface) > 0 {
				artistNames := make([]string, 0, len(artistsInterface))
				for _, artistInterface := range artistsInterface {
					if artist, ok := artistInterface.(map[string]interface{}); ok {
						if artistName, ok := artist["name"].(string); ok {
							artistNames = append(artistNames, artistName)
						}
					}
				}
				if len(artistNames) > 0 {
					artists = strings.Join(artistNames, ", ")
				}
			}

			album := "Unknown Album"
			if albumInterface, ok := track["album"].(map[string]interface{}); ok {
				if albumName, ok := albumInterface["name"].(string); ok && albumName != "" {
					album = albumName
				}
			}

			duration := "—"
			if durationMs, ok := track["duration_ms"].(float64); ok && durationMs > 0 {
				duration = formatTrackDuration(int(durationMs))
			}

			added := "—"
			if playlistTrack.AddedAt != "" {
				added = formatDate(playlistTrack.AddedAt)
			}

			fmt.Printf("%-22s %-40s %-25s %-25s %-8s %s\n",
				trackID,
				truncateString(trackName, 38),
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
