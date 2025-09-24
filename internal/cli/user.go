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
	userLimit     int
	userOffset    int
	userTimeRange string
	userFormat    string
)

// userCmd represents the user command
var userCmd = &cobra.Command{
	Use:   "user",
	Short: "User profile and preferences",
	Long: `Access user profiles, top tracks, top artists, and manage following.

Requires user authentication. Use 'auth login' to authenticate with user account first.
Client credentials authentication does not provide access to user data.`,
	Example: `  # Get current user profile
  spotify-cli user profile

  # Get a specific user's profile
  spotify-cli user profile <user-id>

  # Get your top tracks
  spotify-cli user top tracks

  # Get your top artists
  spotify-cli user top artists --time-range short_term

  # Get user's public playlists
  spotify-cli user playlists <user-id>

  # Follow artists
  spotify-cli user follow <artist-id> [artist-id...]

  # Check if following artists
  spotify-cli user following <artist-id> [artist-id...]`,
}

var userProfileCmd = &cobra.Command{
	Use:   "profile [user-id]",
	Short: "Get user profile",
	Long: `Get current user profile or a specific user's profile.

If no user ID is provided, returns your own profile.`,
	Args: cobra.MaximumNArgs(1),
	Example: `  spotify-cli user profile
  spotify-cli user profile spotify`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runUserCurrentProfile()
		}
		return runUserProfile(args[0])
	},
}

var userTopCmd = &cobra.Command{
	Use:   "top [tracks|artists]",
	Short: "Get your top tracks or artists",
	Long: `Get your most played tracks or artists based on listening history.

Time ranges:
- short_term: ~4 weeks
- medium_term: ~6 months (default)
- long_term: ~several years`,
	Args: cobra.ExactArgs(1),
	Example: `  spotify-cli user top tracks
  spotify-cli user top artists
  spotify-cli user top tracks --time-range short_term
  spotify-cli user top artists --limit 50`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUserTop(args[0])
	},
}

var userFollowCmd = &cobra.Command{
	Use:   "follow [artist-id...]",
	Short: "Follow artists",
	Long: `Follow one or more artists on Spotify.

You can provide multiple artist IDs to follow multiple artists at once (up to 50).`,
	Args: cobra.MinimumNArgs(1),
	Example: `  spotify-cli user follow 4Z8W4fKeB5YxbusRsdQVPb
  spotify-cli user follow artist1 artist2 artist3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUserFollow(args)
	},
}

var userUnfollowCmd = &cobra.Command{
	Use:   "unfollow [artist-id...]",
	Short: "Unfollow artists",
	Long: `Unfollow one or more artists on Spotify.

You can provide multiple artist IDs to unfollow multiple artists at once (up to 50).`,
	Args: cobra.MinimumNArgs(1),
	Example: `  spotify-cli user unfollow 4Z8W4fKeB5YxbusRsdQVPb
  spotify-cli user unfollow artist1 artist2 artist3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUserUnfollow(args)
	},
}

var userFollowingCmd = &cobra.Command{
	Use:   "following [artist-id...]",
	Short: "Check if following artists",
	Long: `Check if you are following one or more artists on Spotify.

You can check multiple artist IDs at once (up to 50).`,
	Args: cobra.MinimumNArgs(1),
	Example: `  spotify-cli user following 4Z8W4fKeB5YxbusRsdQVPb
  spotify-cli user following artist1 artist2 artist3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUserFollowing(args)
	},
}

var userPlaylistsCmd = &cobra.Command{
	Use:   "playlists [user-id]",
	Short: "Get user's public playlists",
	Long: `Get public playlists for a specific user.

If no user ID is provided, returns your own playlists (same as 'playlist list').
Only public playlists are returned when fetching another user's playlists.`,
	Args: cobra.MaximumNArgs(1),
	Example: `  spotify-cli user playlists
  spotify-cli user playlists spotify
  spotify-cli user playlists someuser --limit 50`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return runUserOwnPlaylists()
		}
		return runUserPlaylists(args[0])
	},
}

func init() {
	rootCmd.AddCommand(userCmd)
	userCmd.AddCommand(userProfileCmd)
	userCmd.AddCommand(userTopCmd)
	userCmd.AddCommand(userFollowCmd)
	userCmd.AddCommand(userUnfollowCmd)
	userCmd.AddCommand(userFollowingCmd)
	userCmd.AddCommand(userPlaylistsCmd)

	// Add flags to list commands
	for _, cmd := range []*cobra.Command{userTopCmd, userPlaylistsCmd} {
		cmd.Flags().IntVarP(&userLimit, "limit", "l", 20, "Number of results to return (1-50)")
		cmd.Flags().IntVarP(&userOffset, "offset", "", 0, "Offset for pagination")
		cmd.Flags().StringVarP(&userFormat, "format", "f", "table", "Output format (table, list, json, yaml)")
	}

	// Add time-range flag only to top command
	userTopCmd.Flags().StringVarP(&userTimeRange, "time-range", "t", "medium_term", "Time range (short_term, medium_term, long_term)")

	// Add format flag to profile command
	userProfileCmd.Flags().StringVarP(&userFormat, "format", "f", "table", "Output format (table, json, yaml)")
}

func runUserCurrentProfile() error {
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
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access your profile")
	}

	user, err := spotifyClient.Users.GetCurrentUser(GetCommandContext())
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	return outputUser(user, "Your Profile")
}

func runUserProfile(userID string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' or 'spotify-cli auth client-credentials'")
	}

	user, err := spotifyClient.Users.GetUser(GetCommandContext(), userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	return outputUser(user, fmt.Sprintf("User Profile: %s", userID))
}

func runUserTop(topType string) error {
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
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access your top content")
	}

	// Validate top type
	if topType != "tracks" && topType != "artists" {
		return fmt.Errorf("invalid top type '%s'. Must be 'tracks' or 'artists'", topType)
	}

	// Create options
	options := &spotify.TopItemsOptions{
		TimeRange: userTimeRange,
		Limit:     userLimit,
		Offset:    userOffset,
	}

	switch topType {
	case "tracks":
		tracks, pagination, err := spotifyClient.Users.GetTopTracks(GetCommandContext(), options)
		if err != nil {
			return fmt.Errorf("failed to get top tracks: %w", err)
		}
		return outputTopTracks(tracks, pagination)

	case "artists":
		artists, pagination, err := spotifyClient.Users.GetTopArtists(GetCommandContext(), options)
		if err != nil {
			return fmt.Errorf("failed to get top artists: %w", err)
		}
		return outputTopArtists(artists, pagination)

	default:
		return fmt.Errorf("unsupported top type: %s", topType)
	}
}

func runUserFollow(artistIDs []string) error {
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
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to follow artists")
	}

	if len(artistIDs) > 50 {
		return fmt.Errorf("cannot follow more than 50 artists at once")
	}

	err = spotifyClient.Users.FollowArtists(GetCommandContext(), artistIDs)
	if err != nil {
		return fmt.Errorf("failed to follow artists: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("Successfully followed %d artist(s)", len(artistIDs)))
	return nil
}

func runUserUnfollow(artistIDs []string) error {
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
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to unfollow artists")
	}

	if len(artistIDs) > 50 {
		return fmt.Errorf("cannot unfollow more than 50 artists at once")
	}

	err = spotifyClient.Users.UnfollowArtists(GetCommandContext(), artistIDs)
	if err != nil {
		return fmt.Errorf("failed to unfollow artists: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("Successfully unfollowed %d artist(s)", len(artistIDs)))
	return nil
}

func runUserFollowing(artistIDs []string) error {
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
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to check following status")
	}

	if len(artistIDs) > 50 {
		return fmt.Errorf("cannot check more than 50 artists at once")
	}

	following, err := spotifyClient.Users.CheckFollowingArtists(GetCommandContext(), artistIDs)
	if err != nil {
		return fmt.Errorf("failed to check following artists: %w", err)
	}

	return outputFollowingResults(artistIDs, following)
}

func runUserOwnPlaylists() error {
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
		Limit:  userLimit,
		Offset: userOffset,
	}

	playlists, pagination, err := spotifyClient.Playlists.GetUserPlaylists(GetCommandContext(), paginationOpts)
	if err != nil {
		return fmt.Errorf("failed to get your playlists: %w", err)
	}

	return outputUserPlaylists(playlists, pagination, "Your Playlists")
}

func runUserPlaylists(userID string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' or 'spotify-cli auth client-credentials'")
	}

	// Create pagination options
	paginationOpts := &api.PaginationOptions{
		Limit:  userLimit,
		Offset: userOffset,
	}

	playlists, pagination, err := spotifyClient.Playlists.GetUserPlaylistsByID(GetCommandContext(), userID, paginationOpts)
	if err != nil {
		return fmt.Errorf("failed to get playlists for user '%s': %w", userID, err)
	}

	return outputUserPlaylists(playlists, pagination, fmt.Sprintf("Public Playlists for %s", userID))
}

func outputUser(user *models.User, title string) error {
	cfg := config.Get()

	// For structured output
	if cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml" || userFormat == "json" || userFormat == "yaml" {
		return utils.Output(user)
	}

	// Text output
	fmt.Printf("%s\n", title)
	fmt.Println(strings.Repeat("-", len(title)))
	fmt.Printf("Display Name: %s\n", user.DisplayName)
	fmt.Printf("User ID: %s\n", user.ID)
	if user.Email != "" {
		fmt.Printf("Email: %s\n", user.Email)
	}
	if user.Country != "" {
		fmt.Printf("Country: %s\n", user.Country)
	}
	if user.Product != "" {
		fmt.Printf("Subscription: %s\n", user.Product)
	}
	fmt.Printf("Followers: %d\n", user.Followers.Total)
	if user.ExternalURLs.Spotify != "" {
		fmt.Printf("Spotify URL: %s\n", user.ExternalURLs.Spotify)
	}

	return nil
}

func outputTopTracks(tracks *models.Paging[models.Track], pagination *api.PaginationInfo) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := userFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output, return the data directly
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(map[string]interface{}{
			"results":    tracks,
			"pagination": pagination,
			"time_range": userTimeRange,
		})
	}

	// Text-based output
	if len(tracks.Items) == 0 {
		fmt.Println("No top tracks found.")
		return nil
	}

	// Print header
	fmt.Printf("Your Top Tracks (%s) - Found %d tracks", userTimeRange, tracks.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(tracks.Items))
	}
	fmt.Println()
	fmt.Println()

	if userFormat == "list" {
		for i, track := range tracks.Items {
			fmt.Printf("%d. %s\n", i+1, track.Name)
			if len(track.Artists) > 0 {
				artistNames := make([]string, len(track.Artists))
				for j, artist := range track.Artists {
					artistNames[j] = artist.Name
				}
				fmt.Printf("   by %s\n", strings.Join(artistNames, ", "))
			}
			if track.Album.Name != "" {
				fmt.Printf("   from %s\n", track.Album.Name)
			}
			fmt.Printf("   ID: %s\n", track.ID)
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-22s %-40s %-25s %-25s %s\n", "ID", "TRACK", "ARTIST", "ALBUM", "DURATION")
		fmt.Println(strings.Repeat("-", 135))

		for _, track := range tracks.Items {
			artists := "Unknown Artist"
			if len(track.Artists) > 0 {
				artistNames := make([]string, len(track.Artists))
				for i, artist := range track.Artists {
					artistNames[i] = artist.Name
				}
				artists = strings.Join(artistNames, ", ")
			}

			album := track.Album.Name
			if album == "" {
				album = "Unknown Album"
			}

			duration := utils.FormatDuration(track.DurationMs)

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

func outputTopArtists(artists *models.Paging[models.Artist], pagination *api.PaginationInfo) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := userFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output, return the data directly
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(map[string]interface{}{
			"results":    artists,
			"pagination": pagination,
			"time_range": userTimeRange,
		})
	}

	// Text-based output
	if len(artists.Items) == 0 {
		fmt.Println("No top artists found.")
		return nil
	}

	// Print header
	fmt.Printf("Your Top Artists (%s) - Found %d artists", userTimeRange, artists.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(artists.Items))
	}
	fmt.Println()
	fmt.Println()

	if userFormat == "list" {
		for i, artist := range artists.Items {
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

		for _, artist := range artists.Items {
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

func outputFollowingResults(artistIDs []string, following []bool) error {
	cfg := config.Get()

	// For structured output
	if cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml" {
		results := make([]map[string]interface{}, len(artistIDs))
		for i, id := range artistIDs {
			results[i] = map[string]interface{}{
				"artist_id": id,
				"following": following[i],
			}
		}
		return utils.Output(map[string]interface{}{
			"results": results,
		})
	}

	// Text output
	fmt.Printf("Following Check Results - %d artist%s\n", len(artistIDs), pluralize(len(artistIDs)))
	fmt.Println(strings.Repeat("-", 60))

	for i, id := range artistIDs {
		status := "❌ Not following"
		if following[i] {
			status = "✅ Following"
		}
		fmt.Printf("%-40s %s\n", truncateString(id, 38), status)
	}

	// Summary
	followingCount := 0
	for _, isFollowing := range following {
		if isFollowing {
			followingCount++
		}
	}
	fmt.Printf("\nSummary: Following %d/%d artist%s\n",
		followingCount, len(artistIDs), pluralize(len(artistIDs)))

	return nil
}

func outputUserPlaylists(playlists *models.Paging[models.Playlist], pagination *api.PaginationInfo, title string) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := userFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output, return the data directly
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(map[string]interface{}{
			"results":    playlists,
			"pagination": pagination,
		})
	}

	// Text-based output
	if len(playlists.Items) == 0 {
		fmt.Printf("No playlists found.\n")
		return nil
	}

	// Print header
	fmt.Printf("%s - %d total", title, playlists.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(playlists.Items))
	}
	fmt.Println()
	fmt.Println()

	if userFormat == "list" {
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
				description = "—"
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