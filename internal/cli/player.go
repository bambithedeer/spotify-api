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
	playerDeviceID   string
	playerVolume     int
	playerPosition   int
	playerRepeat     string
	playerShuffle    bool
	playerLimit      int
	playerFormat     string
	playerURI        string
	playerURIs       []string
	playerContext    string
)

// playerCmd represents the player command
var playerCmd = &cobra.Command{
	Use:   "player",
	Short: "Control playback and manage player state",
	Long: `Control Spotify playback including play, pause, skip, volume, and more.

Requires user authentication and an active Spotify device. Use 'auth login' to authenticate with user account first.
Client credentials authentication does not provide access to playback control.`,
	Example: `  # Get current playback state
  spotify-cli player status

  # Control playback
  spotify-cli player play
  spotify-cli player pause
  spotify-cli player next
  spotify-cli player previous

  # Control volume
  spotify-cli player volume 75

  # Control shuffle and repeat
  spotify-cli player shuffle on
  spotify-cli player repeat track`,
}

var playerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get current playback state",
	Long:  `Get detailed information about the current playback state including track, device, and playback settings.`,
	Example: `  spotify-cli player status
  spotify-cli player status --format json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerStatus()
	},
}

var playerCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Get currently playing track",
	Long:  `Get information about the currently playing track.`,
	Example: `  spotify-cli player current`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerCurrent()
	},
}

var playerDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List available devices",
	Long:  `List all devices available for playback control.`,
	Example: `  spotify-cli player devices`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerDevices()
	},
}

var playerPlayCmd = &cobra.Command{
	Use:   "play [uri...] or [query]",
	Short: "Start or resume playback",
	Long: `Start or resume playback. You can specify URIs, IDs, or search queries.

You can provide:
- Track URIs: spotify:track:4iV5W9uYEdYUVa79Axb7Rh
- Track IDs: 4iV5W9uYEdYUVa79Axb7Rh
- Album/Playlist URIs: spotify:album:4aawyAB9vmqN3uQ7FjRGTy
- Context URI with --context flag
- Search queries: artist:"queen", track:"bohemian rhapsody", album:"greatest hits"
- Saved content: saved:tracks, saved:albums, my:playlists, followed:artists`,
	Example: `  # Resume playback
  spotify-cli player play

  # Play specific tracks
  spotify-cli player play spotify:track:4iV5W9uYEdYUVa79Axb7Rh
  spotify-cli player play 4iV5W9uYEdYUVa79Axb7Rh

  # Play from context (album/playlist)
  spotify-cli player play --context spotify:album:4aawyAB9vmqN3uQ7FjRGTy

  # Search and play
  spotify-cli player play artist:"queen"
  spotify-cli player play track:"bohemian rhapsody"
  spotify-cli player play album:"greatest hits"

  # Play from your saved content
  spotify-cli player play saved:tracks
  spotify-cli player play saved:albums
  spotify-cli player play my:playlists
  spotify-cli player play followed:artists`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerPlay(args)
	},
}

var playerPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause playback",
	Long:  `Pause the currently playing track.`,
	Example: `  spotify-cli player pause`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerPause()
	},
}

var playerNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Skip to next track",
	Long:  `Skip to the next track in the queue.`,
	Example: `  spotify-cli player next`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerNext()
	},
}

var playerPreviousCmd = &cobra.Command{
	Use:   "previous",
	Short: "Skip to previous track",
	Long:  `Skip to the previous track.`,
	Example: `  spotify-cli player previous`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerPrevious()
	},
}

var playerVolumeCmd = &cobra.Command{
	Use:   "volume [0-100]",
	Short: "Set playback volume",
	Long:  `Set the playback volume (0-100).`,
	Args:  cobra.ExactArgs(1),
	Example: `  spotify-cli player volume 75`,
	RunE: func(cmd *cobra.Command, args []string) error {
		volume, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid volume: %s", args[0])
		}
		return runPlayerVolume(volume)
	},
}

var playerShuffleCmd = &cobra.Command{
	Use:   "shuffle [on|off]",
	Short: "Set shuffle mode",
	Long:  `Enable or disable shuffle mode.`,
	Args:  cobra.ExactArgs(1),
	Example: `  spotify-cli player shuffle on
  spotify-cli player shuffle off`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerShuffle(args[0])
	},
}

var playerRepeatCmd = &cobra.Command{
	Use:   "repeat [track|context|off]",
	Short: "Set repeat mode",
	Long:  `Set repeat mode: track (repeat current track), context (repeat album/playlist), or off.`,
	Args:  cobra.ExactArgs(1),
	Example: `  spotify-cli player repeat track
  spotify-cli player repeat context
  spotify-cli player repeat off`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerRepeat(args[0])
	},
}

var playerSeekCmd = &cobra.Command{
	Use:   "seek [position]",
	Short: "Seek to position in track",
	Long:  `Seek to a specific position in the currently playing track. Position can be in seconds or MM:SS format.`,
	Args:  cobra.ExactArgs(1),
	Example: `  spotify-cli player seek 120
  spotify-cli player seek 2:30`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerSeek(args[0])
	},
}

var playerQueueCmd = &cobra.Command{
	Use:   "queue [uri]",
	Short: "Add track to queue",
	Long:  `Add a track to the playback queue.`,
	Args:  cobra.ExactArgs(1),
	Example: `  spotify-cli player queue spotify:track:4iV5W9uYEdYUVa79Axb7Rh`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerQueue(args[0])
	},
}

var playerRecentCmd = &cobra.Command{
	Use:   "recent",
	Short: "Get recently played tracks",
	Long:  `Get tracks from recently played history.`,
	Example: `  spotify-cli player recent
  spotify-cli player recent --limit 10`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPlayerRecent()
	},
}

func init() {
	rootCmd.AddCommand(playerCmd)
	playerCmd.AddCommand(playerStatusCmd)
	playerCmd.AddCommand(playerCurrentCmd)
	playerCmd.AddCommand(playerDevicesCmd)
	playerCmd.AddCommand(playerPlayCmd)
	playerCmd.AddCommand(playerPauseCmd)
	playerCmd.AddCommand(playerNextCmd)
	playerCmd.AddCommand(playerPreviousCmd)
	playerCmd.AddCommand(playerVolumeCmd)
	playerCmd.AddCommand(playerShuffleCmd)
	playerCmd.AddCommand(playerRepeatCmd)
	playerCmd.AddCommand(playerSeekCmd)
	playerCmd.AddCommand(playerQueueCmd)
	playerCmd.AddCommand(playerRecentCmd)

	// Global flags for all player commands
	for _, cmd := range []*cobra.Command{
		playerStatusCmd, playerCurrentCmd, playerDevicesCmd, playerPlayCmd,
		playerPauseCmd, playerNextCmd, playerPreviousCmd, playerVolumeCmd,
		playerShuffleCmd, playerRepeatCmd, playerSeekCmd, playerQueueCmd,
	} {
		cmd.Flags().StringVarP(&playerDeviceID, "device", "d", "", "Target device ID")
	}

	// Format flags for display commands
	for _, cmd := range []*cobra.Command{playerStatusCmd, playerCurrentCmd, playerDevicesCmd, playerRecentCmd} {
		cmd.Flags().StringVarP(&playerFormat, "format", "f", "table", "Output format (table, list, json, yaml)")
	}

	// Play command specific flags
	playerPlayCmd.Flags().StringVarP(&playerContext, "context", "c", "", "Context URI (album, playlist, etc.)")
	playerPlayCmd.Flags().IntVarP(&playerPosition, "position", "p", 0, "Start position in milliseconds")

	// Recent tracks flags
	playerRecentCmd.Flags().IntVarP(&playerLimit, "limit", "l", 20, "Number of results to return (1-50)")
}

func runPlayerStatus() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	state, err := spotifyClient.Player.GetPlaybackState(GetCommandContext(), "")
	if err != nil {
		return fmt.Errorf("failed to get playback state: %w", err)
	}

	return outputPlaybackState(state)
}

func runPlayerCurrent() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	playing, err := spotifyClient.Player.GetCurrentlyPlaying(GetCommandContext(), nil)
	if err != nil {
		return fmt.Errorf("failed to get currently playing: %w", err)
	}

	return outputCurrentlyPlaying(playing)
}

func runPlayerDevices() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	devices, err := spotifyClient.Player.GetDevices(GetCommandContext())
	if err != nil {
		return fmt.Errorf("failed to get devices: %w", err)
	}

	return outputDevices(devices)
}

func runPlayerPlay(uris []string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	options := &spotify.PlayOptions{
		DeviceID:   playerDeviceID,
		PositionMs: playerPosition,
	}

	if playerContext != "" {
		options.ContextURI = playerContext
	} else if len(uris) > 0 {
		// Check if this is a search query
		query := strings.Join(uris, " ")
		if isSearchQuery(query) {
			searchResults, err := handlePlayerSearchQuery(spotifyClient, query)
			if err != nil {
				return fmt.Errorf("search failed: %w", err)
			}
			options.URIs = searchResults
		} else if len(uris) == 1 {
			// Single URI/ID - check if it's a context (playlist, album, artist) or track
			uri := uris[0]
			if strings.HasPrefix(uri, "spotify:") {
				// Already a URI - check type
				if strings.Contains(uri, ":playlist:") || strings.Contains(uri, ":album:") || strings.Contains(uri, ":artist:") {
					options.ContextURI = uri
				} else {
					options.URIs = []string{uri}
				}
			} else if len(uri) == 22 {
				// 22-character ID - try to determine type by checking if it's a known playlist/album
				contextURI, err := tryAsContextURI(spotifyClient, uri)
				if err == nil && contextURI != "" {
					options.ContextURI = contextURI
				} else {
					// Default to track
					options.URIs = []string{fmt.Sprintf("spotify:track:%s", uri)}
				}
			} else {
				options.URIs = []string{uri} // Let API handle error if invalid
			}
		} else {
			// Multiple URIs - convert IDs to track URIs
			spotifyURIs := make([]string, len(uris))
			for i, uri := range uris {
				if strings.HasPrefix(uri, "spotify:") {
					spotifyURIs[i] = uri
				} else if len(uri) == 22 {
					spotifyURIs[i] = fmt.Sprintf("spotify:track:%s", uri)
				} else {
					spotifyURIs[i] = uri // Let API handle error if invalid
				}
			}
			options.URIs = spotifyURIs
		}
	}

	err = spotifyClient.Player.Play(GetCommandContext(), options)
	if err != nil {
		return fmt.Errorf("failed to start playback: %w", err)
	}

	if options.ContextURI != "" {
		// Extract type from context URI
		contextType := "content"
		if strings.Contains(options.ContextURI, ":playlist:") {
			contextType = "playlist"
		} else if strings.Contains(options.ContextURI, ":album:") {
			contextType = "album"
		} else if strings.Contains(options.ContextURI, ":artist:") {
			contextType = "artist"
		}
		utils.PrintSuccess(fmt.Sprintf("Started playback of %s", contextType))
	} else if len(options.URIs) > 0 {
		utils.PrintSuccess(fmt.Sprintf("Started playback of %d track(s)", len(options.URIs)))
	} else if playerContext != "" {
		utils.PrintSuccess("Started playback from context")
	} else {
		utils.PrintSuccess("Resumed playback")
	}

	return nil
}

func runPlayerPause() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	err = spotifyClient.Player.Pause(GetCommandContext(), playerDeviceID)
	if err != nil {
		return fmt.Errorf("failed to pause playback: %w", err)
	}

	utils.PrintSuccess("Paused playback")
	return nil
}

func runPlayerNext() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	err = spotifyClient.Player.Next(GetCommandContext(), playerDeviceID)
	if err != nil {
		return fmt.Errorf("failed to skip to next track: %w", err)
	}

	utils.PrintSuccess("Skipped to next track")
	return nil
}

func runPlayerPrevious() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	err = spotifyClient.Player.Previous(GetCommandContext(), playerDeviceID)
	if err != nil {
		return fmt.Errorf("failed to skip to previous track: %w", err)
	}

	utils.PrintSuccess("Skipped to previous track")
	return nil
}

func runPlayerVolume(volume int) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	err = spotifyClient.Player.SetVolume(GetCommandContext(), volume, playerDeviceID)
	if err != nil {
		return fmt.Errorf("failed to set volume: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("Set volume to %d%%", volume))
	return nil
}

func runPlayerShuffle(state string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	var shuffle bool
	switch strings.ToLower(state) {
	case "on", "true", "1":
		shuffle = true
	case "off", "false", "0":
		shuffle = false
	default:
		return fmt.Errorf("invalid shuffle state: %s (use 'on' or 'off')", state)
	}

	err = spotifyClient.Player.SetShuffle(GetCommandContext(), shuffle, playerDeviceID)
	if err != nil {
		return fmt.Errorf("failed to set shuffle: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("Set shuffle %s", map[bool]string{true: "on", false: "off"}[shuffle]))
	return nil
}

func runPlayerRepeat(state string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	err = spotifyClient.Player.SetRepeat(GetCommandContext(), strings.ToLower(state), playerDeviceID)
	if err != nil {
		return fmt.Errorf("failed to set repeat: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("Set repeat mode to %s", state))
	return nil
}

func runPlayerSeek(position string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	positionMs, err := parsePosition(position)
	if err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	err = spotifyClient.Player.Seek(GetCommandContext(), positionMs, playerDeviceID)
	if err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	utils.PrintSuccess(fmt.Sprintf("Seeked to %s", position))
	return nil
}

func runPlayerQueue(uri string) error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	err = spotifyClient.Player.AddToQueue(GetCommandContext(), uri, playerDeviceID)
	if err != nil {
		return fmt.Errorf("failed to add to queue: %w", err)
	}

	utils.PrintSuccess("Added track to queue")
	return nil
}

func runPlayerRecent() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' for user account access")
	}

	cfg := config.Get()
	if cfg.RefreshToken == "" {
		return fmt.Errorf("user authentication required. Client credentials only provide access to public data. Run 'spotify-cli auth login' to access playback control")
	}

	options := &spotify.RecentlyPlayedOptions{
		Limit: playerLimit,
	}

	playHistory, err := spotifyClient.Player.GetRecentlyPlayed(GetCommandContext(), options)
	if err != nil {
		return fmt.Errorf("failed to get recently played: %w", err)
	}

	return outputRecentlyPlayed(playHistory)
}

func outputPlaybackState(state *models.PlaybackState) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := playerFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(state)
	}

	// Text output
	if state.Item == nil {
		fmt.Println("No track currently playing")
		return nil
	}

	// Convert interface{} to Track using type assertion on map
	if itemMap, ok := state.Item.(map[string]interface{}); ok {
		if trackName, exists := itemMap["name"].(string); exists {
			fmt.Printf("Now Playing: %s\n", trackName)

			// Extract artist information
			if artistsData, exists := itemMap["artists"].([]interface{}); exists && len(artistsData) > 0 {
				artists := make([]string, 0, len(artistsData))
				for _, artistData := range artistsData {
					if artistMap, ok := artistData.(map[string]interface{}); ok {
						if artistName, ok := artistMap["name"].(string); ok {
							artists = append(artists, artistName)
						}
					}
				}
				if len(artists) > 0 {
					fmt.Printf("Artist(s): %s\n", strings.Join(artists, ", "))
				}
			}

			// Extract album information
			if albumData, exists := itemMap["album"].(map[string]interface{}); exists {
				if albumName, ok := albumData["name"].(string); ok {
					fmt.Printf("Album: %s\n", albumName)
				}
			}

			// Extract duration
			if durationMs, exists := itemMap["duration_ms"].(float64); exists {
				fmt.Printf("Progress: %s / %s\n",
					formatPlayerDuration(state.ProgressMs),
					formatPlayerDuration(int(durationMs)))
			}
		}
	} else {
		fmt.Printf("Currently playing: %s\n", state.CurrentlyPlayingType)
	}

	fmt.Printf("Playing: %t\n", state.IsPlaying)
	fmt.Printf("Shuffle: %t\n", state.ShuffleState)
	fmt.Printf("Repeat: %s\n", state.RepeatState)
	fmt.Printf("Volume: %d%%\n", state.Device.VolumePercent)
	if state.Device.Name != "" {
		fmt.Printf("Device: %s (%s)\n", state.Device.Name, state.Device.Type)
	}

	return nil
}

func outputCurrentlyPlaying(playing *models.CurrentlyPlaying) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := playerFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(playing)
	}

	// Text output
	if playing.Item == nil {
		fmt.Println("No track currently playing")
		return nil
	}

	// Convert interface{} to Track using type assertion on map
	if itemMap, ok := playing.Item.(map[string]interface{}); ok {
		if trackName, exists := itemMap["name"].(string); exists {
			fmt.Printf("Currently Playing: %s\n", trackName)

			// Extract artist information
			if artistsData, exists := itemMap["artists"].([]interface{}); exists && len(artistsData) > 0 {
				artists := make([]string, 0, len(artistsData))
				for _, artistData := range artistsData {
					if artistMap, ok := artistData.(map[string]interface{}); ok {
						if artistName, ok := artistMap["name"].(string); ok {
							artists = append(artists, artistName)
						}
					}
				}
				if len(artists) > 0 {
					fmt.Printf("Artist(s): %s\n", strings.Join(artists, ", "))
				}
			}

			// Extract album information
			if albumData, exists := itemMap["album"].(map[string]interface{}); exists {
				if albumName, ok := albumData["name"].(string); ok {
					fmt.Printf("Album: %s\n", albumName)
				}
			}

			fmt.Printf("Playing: %t\n", playing.IsPlaying)

			// Extract duration
			if durationMs, exists := itemMap["duration_ms"].(float64); exists {
				fmt.Printf("Progress: %s / %s\n",
					formatPlayerDuration(playing.ProgressMs),
					formatPlayerDuration(int(durationMs)))
			}
		}
	} else {
		fmt.Printf("Currently playing: %s\n", playing.CurrentlyPlayingType)
		fmt.Printf("Playing: %t\n", playing.IsPlaying)
	}

	return nil
}

func outputDevices(devices *models.DevicesResponse) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := playerFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(devices)
	}

	// Text output
	if len(devices.Devices) == 0 {
		fmt.Println("No devices available")
		return nil
	}

	fmt.Printf("Available Devices (%d)\n\n", len(devices.Devices))

	if playerFormat == "list" {
		for i, device := range devices.Devices {
			fmt.Printf("%d. %s\n", i+1, device.Name)
			fmt.Printf("   Type: %s\n", device.Type)
			fmt.Printf("   Active: %t\n", device.IsActive)
			fmt.Printf("   Volume: %d%%\n", device.VolumePercent)
			fmt.Printf("   ID: %s\n", device.ID)
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-30s %-15s %-8s %-8s %s\n", "NAME", "TYPE", "ACTIVE", "VOLUME", "ID")
		fmt.Println(strings.Repeat("-", 80))

		for _, device := range devices.Devices {
			active := "No"
			if device.IsActive {
				active = "Yes"
			}

			fmt.Printf("%-30s %-15s %-8s %-8s %s\n",
				truncateString(device.Name, 28),
				device.Type,
				active,
				fmt.Sprintf("%d%%", device.VolumePercent),
				device.ID[:min(len(device.ID), 20)])
		}
	}

	return nil
}

func outputRecentlyPlayed(playHistory *models.CursorPaging[models.PlayHistory]) error {
	cfg := config.Get()

	// Check output format priority: flag > global config > default
	outputFormat := playerFormat
	if outputFormat == "table" && (cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml") {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(playHistory)
	}

	// Text output
	if len(playHistory.Items) == 0 {
		fmt.Println("No recently played tracks found.")
		return nil
	}

	fmt.Printf("Recently Played Tracks (%d)\n\n", len(playHistory.Items))

	if playerFormat == "list" {
		for i, item := range playHistory.Items {
			fmt.Printf("%d. %s\n", i+1, item.Track.Name)
			if len(item.Track.Artists) > 0 {
				artists := make([]string, len(item.Track.Artists))
				for j, artist := range item.Track.Artists {
					artists[j] = artist.Name
				}
				fmt.Printf("   by %s\n", strings.Join(artists, ", "))
			}
			if item.Track.Album != nil {
				fmt.Printf("   from %s\n", item.Track.Album.Name)
			}
			fmt.Printf("   played at %s\n", item.PlayedAt)
			fmt.Println()
		}
	} else {
		// Table format
		fmt.Printf("%-40s %-30s %-25s %s\n", "TRACK", "ARTIST", "ALBUM", "PLAYED AT")
		fmt.Println(strings.Repeat("-", 120))

		for _, item := range playHistory.Items {
			artists := ""
			if len(item.Track.Artists) > 0 {
				artistNames := make([]string, len(item.Track.Artists))
				for i, artist := range item.Track.Artists {
					artistNames[i] = artist.Name
				}
				artists = strings.Join(artistNames, ", ")
			}

			album := ""
			if item.Track.Album != nil {
				album = item.Track.Album.Name
			}

			fmt.Printf("%-40s %-30s %-25s %s\n",
				truncateString(item.Track.Name, 38),
				truncateString(artists, 28),
				truncateString(album, 23),
				item.PlayedAt)
		}
	}

	return nil
}

// Utility functions

func parsePosition(position string) (int, error) {
	// Try parsing as seconds first
	if seconds, err := strconv.Atoi(position); err == nil {
		return seconds * 1000, nil
	}

	// Try parsing as MM:SS format
	if strings.Contains(position, ":") {
		parts := strings.Split(position, ":")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid time format, use seconds or MM:SS")
		}

		minutes, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %s", parts[0])
		}

		seconds, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %s", parts[1])
		}

		return (minutes*60 + seconds) * 1000, nil
	}

	return 0, fmt.Errorf("invalid position format")
}

func formatPlayerDuration(ms int) string {
	seconds := ms / 1000
	minutes := seconds / 60
	seconds = seconds % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Helper functions for search query support

func isSearchQuery(query string) bool {
	// Check if query contains search operators like artist:, track:, album:, playlist:
	searchOperators := []string{"artist:", "track:", "album:", "playlist:", "genre:", "year:", "saved:", "my:", "followed:"}
	queryLower := strings.ToLower(query)
	for _, operator := range searchOperators {
		if strings.Contains(queryLower, operator) {
			return true
		}
	}
	return false
}

func handlePlayerSearchQuery(spotifyClient *client.SpotifyClient, query string) ([]string, error) {
	// Determine search type and execute search
	queryLower := strings.ToLower(query)

	// Default pagination options for search
	paginationOpts := &api.PaginationOptions{
		Limit:  10,
		Offset: 0,
	}

	if strings.Contains(queryLower, "saved:") {
		// Handle saved content
		return handleSavedContentSearch(spotifyClient, query, paginationOpts)
	} else if strings.Contains(queryLower, "my:") {
		// Handle user's own content
		return handleMyContentSearch(spotifyClient, query, paginationOpts)
	} else if strings.Contains(queryLower, "followed:") {
		// Handle followed artists
		return handleFollowedContentSearch(spotifyClient, query, paginationOpts)
	} else if strings.Contains(queryLower, "artist:") {
		// Search for artist's top tracks
		return handleArtistSearch(spotifyClient, query, paginationOpts)
	} else if strings.Contains(queryLower, "album:") {
		// Search for album and return its context URI
		return handleAlbumSearch(spotifyClient, query, paginationOpts)
	} else if strings.Contains(queryLower, "playlist:") {
		// Search for playlist and return its context URI
		return handlePlaylistSearch(spotifyClient, query, paginationOpts)
	} else {
		// Default to track search
		return handleTrackSearch(spotifyClient, query, paginationOpts)
	}
}

func handleTrackSearch(spotifyClient *client.SpotifyClient, query string, opts *api.PaginationOptions) ([]string, error) {
	tracks, _, err := spotifyClient.Search.SearchTracks(GetCommandContext(), query, opts)
	if err != nil {
		return nil, err
	}

	if len(tracks.Items) == 0 {
		return nil, fmt.Errorf("no tracks found for query: %s", query)
	}

	// Return URIs for first few tracks
	uris := make([]string, 0, min(5, len(tracks.Items)))
	for i := 0; i < min(5, len(tracks.Items)); i++ {
		uris = append(uris, tracks.Items[i].URI)
	}

	fmt.Printf("Playing %d track(s) from search: %s\n", len(uris), query)
	return uris, nil
}

func handleArtistSearch(spotifyClient *client.SpotifyClient, query string, opts *api.PaginationOptions) ([]string, error) {
	artists, _, err := spotifyClient.Search.SearchArtists(GetCommandContext(), query, opts)
	if err != nil {
		return nil, err
	}

	if len(artists.Items) == 0 {
		return nil, fmt.Errorf("no artists found for query: %s", query)
	}

	// Get top tracks for the first artist
	artistID := artists.Items[0].ID
	topTracks, err := spotifyClient.Artists.GetArtistTopTracks(GetCommandContext(), artistID, "US")
	if err != nil {
		return nil, fmt.Errorf("failed to get top tracks for artist: %w", err)
	}

	if len(topTracks) == 0 {
		return nil, fmt.Errorf("no top tracks found for artist: %s", artists.Items[0].Name)
	}

	// Return URIs for top tracks
	uris := make([]string, 0, min(10, len(topTracks)))
	for i := 0; i < min(10, len(topTracks)); i++ {
		uris = append(uris, topTracks[i].URI)
	}

	fmt.Printf("Playing %d top track(s) by %s\n", len(uris), artists.Items[0].Name)
	return uris, nil
}

func handleAlbumSearch(spotifyClient *client.SpotifyClient, query string, opts *api.PaginationOptions) ([]string, error) {
	albums, _, err := spotifyClient.Search.SearchAlbums(GetCommandContext(), query, opts)
	if err != nil {
		return nil, err
	}

	if len(albums.Items) == 0 {
		return nil, fmt.Errorf("no albums found for query: %s", query)
	}

	// Return the album URI as context (will play the whole album)
	album := albums.Items[0]
	fmt.Printf("Playing album: %s by %s\n", album.Name, album.Artists[0].Name)

	// For albums, we need to get the tracks and return their URIs
	tracks, _, err := spotifyClient.Albums.GetAlbumTracks(GetCommandContext(), album.ID, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get album tracks: %w", err)
	}

	if len(tracks.Items) == 0 {
		return nil, fmt.Errorf("no tracks found in album: %s", album.Name)
	}

	// Return URIs for all album tracks
	uris := make([]string, len(tracks.Items))
	for i, track := range tracks.Items {
		uris[i] = track.URI
	}

	return uris, nil
}

func handlePlaylistSearch(spotifyClient *client.SpotifyClient, query string, opts *api.PaginationOptions) ([]string, error) {
	playlists, _, err := spotifyClient.Search.SearchPlaylists(GetCommandContext(), query, opts)
	if err != nil {
		return nil, err
	}

	if len(playlists.Items) == 0 {
		return nil, fmt.Errorf("no playlists found for query: %s", query)
	}

	playlist := playlists.Items[0]
	fmt.Printf("Playing playlist: %s by %s\n", playlist.Name, playlist.Owner.DisplayName)

	// For playlists, get the tracks
	tracks, _, err := spotifyClient.Playlists.GetPlaylistTracks(GetCommandContext(), playlist.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist tracks: %w", err)
	}

	if len(tracks.Items) == 0 {
		return nil, fmt.Errorf("no tracks found in playlist: %s", playlist.Name)
	}

	// Return URIs for playlist tracks (up to 50 to avoid overwhelming)
	uris := make([]string, 0, min(50, len(tracks.Items)))
	for i := 0; i < min(50, len(tracks.Items)); i++ {
		if track, ok := tracks.Items[i].Track.(*models.Track); ok {
			uris = append(uris, track.URI)
		}
	}

	if len(uris) == 0 {
		return nil, fmt.Errorf("no playable tracks found in playlist: %s", playlist.Name)
	}

	return uris, nil
}

// Handler functions for saved/followed content

func handleSavedContentSearch(spotifyClient *client.SpotifyClient, query string, opts *api.PaginationOptions) ([]string, error) {
	queryLower := strings.ToLower(query)

	if strings.Contains(queryLower, "saved:tracks") {
		return handleSavedTracks(spotifyClient, opts)
	} else if strings.Contains(queryLower, "saved:albums") {
		return handleSavedAlbums(spotifyClient, opts)
	}

	return nil, fmt.Errorf("unsupported saved content type. Use 'saved:tracks' or 'saved:albums'")
}

func handleMyContentSearch(spotifyClient *client.SpotifyClient, query string, opts *api.PaginationOptions) ([]string, error) {
	queryLower := strings.ToLower(query)

	if strings.Contains(queryLower, "my:playlists") {
		return handleMyPlaylists(spotifyClient, opts)
	}

	return nil, fmt.Errorf("unsupported my content type. Use 'my:playlists'")
}

func handleFollowedContentSearch(spotifyClient *client.SpotifyClient, query string, opts *api.PaginationOptions) ([]string, error) {
	queryLower := strings.ToLower(query)

	if strings.Contains(queryLower, "followed:artists") {
		return handleFollowedArtists(spotifyClient, opts)
	}

	return nil, fmt.Errorf("unsupported followed content type. Use 'followed:artists'")
}

func handleSavedTracks(spotifyClient *client.SpotifyClient, opts *api.PaginationOptions) ([]string, error) {
	// Get user's saved tracks
	savedTracks, _, err := spotifyClient.Library.GetSavedTracks(GetCommandContext(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get saved tracks: %w", err)
	}

	if len(savedTracks.Items) == 0 {
		return nil, fmt.Errorf("no saved tracks found")
	}

	// Return URIs for saved tracks
	uris := make([]string, 0, len(savedTracks.Items))
	for _, item := range savedTracks.Items {
		uris = append(uris, item.Track.URI)
	}

	fmt.Printf("Playing %d saved track(s)\n", len(uris))
	return uris, nil
}

func handleSavedAlbums(spotifyClient *client.SpotifyClient, opts *api.PaginationOptions) ([]string, error) {
	// Get user's saved albums
	savedAlbumsOpts := &spotify.SavedAlbumsOptions{
		Limit:  opts.Limit,
		Offset: opts.Offset,
	}
	savedAlbums, _, err := spotifyClient.Library.GetSavedAlbums(GetCommandContext(), savedAlbumsOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get saved albums: %w", err)
	}

	if len(savedAlbums.Items) == 0 {
		return nil, fmt.Errorf("no saved albums found")
	}

	// Get tracks from a few saved albums (to avoid overwhelming)
	var allURIs []string
	maxAlbums := min(3, len(savedAlbums.Items))

	for i := 0; i < maxAlbums; i++ {
		album := savedAlbums.Items[i].Album
		tracks, _, err := spotifyClient.Albums.GetAlbumTracks(GetCommandContext(), album.ID, nil, "")
		if err != nil {
			continue // Skip this album if we can't get tracks
		}

		for _, track := range tracks.Items {
			allURIs = append(allURIs, track.URI)
		}
	}

	if len(allURIs) == 0 {
		return nil, fmt.Errorf("no tracks found in saved albums")
	}

	fmt.Printf("Playing tracks from %d saved album(s) (%d tracks total)\n", maxAlbums, len(allURIs))
	return allURIs, nil
}

func handleMyPlaylists(spotifyClient *client.SpotifyClient, opts *api.PaginationOptions) ([]string, error) {
	// Get user's playlists
	playlists, _, err := spotifyClient.Playlists.GetUserPlaylists(GetCommandContext(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get user playlists: %w", err)
	}

	if len(playlists.Items) == 0 {
		return nil, fmt.Errorf("no playlists found")
	}

	// Get tracks from the first few playlists
	var allURIs []string
	maxPlaylists := min(2, len(playlists.Items))

	for i := 0; i < maxPlaylists; i++ {
		playlist := playlists.Items[i]
		tracks, _, err := spotifyClient.Playlists.GetPlaylistTracks(GetCommandContext(), playlist.ID, nil)
		if err != nil {
			continue // Skip this playlist if we can't get tracks
		}

		for j, item := range tracks.Items {
			if j >= 25 { // Limit tracks per playlist
				break
			}
			if track, ok := item.Track.(*models.Track); ok {
				allURIs = append(allURIs, track.URI)
			}
		}
	}

	if len(allURIs) == 0 {
		return nil, fmt.Errorf("no tracks found in playlists")
	}

	fmt.Printf("Playing tracks from %d playlist(s) (%d tracks total)\n", maxPlaylists, len(allURIs))
	return allURIs, nil
}

func handleFollowedArtists(spotifyClient *client.SpotifyClient, opts *api.PaginationOptions) ([]string, error) {
	// Get user's followed artists
	followedOpts := &spotify.FollowedArtistsOptions{
		Limit: opts.Limit,
	}
	followedArtists, err := spotifyClient.Users.GetFollowedArtists(GetCommandContext(), followedOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get followed artists: %w", err)
	}

	if len(followedArtists.Items) == 0 {
		return nil, fmt.Errorf("no followed artists found")
	}

	// Get top tracks from a few followed artists
	var allURIs []string
	maxArtists := min(3, len(followedArtists.Items))

	for i := 0; i < maxArtists; i++ {
		artist := followedArtists.Items[i]
		topTracks, err := spotifyClient.Artists.GetArtistTopTracks(GetCommandContext(), artist.ID, "US")
		if err != nil {
			continue // Skip this artist if we can't get top tracks
		}

		// Add top 3 tracks from each artist
		maxTracks := min(3, len(topTracks))
		for j := 0; j < maxTracks; j++ {
			allURIs = append(allURIs, topTracks[j].URI)
		}
	}

	if len(allURIs) == 0 {
		return nil, fmt.Errorf("no tracks found from followed artists")
	}

	fmt.Printf("Playing top tracks from %d followed artist(s) (%d tracks total)\n", maxArtists, len(allURIs))
	return allURIs, nil
}

// tryAsContextURI attempts to determine if an ID is a playlist, album, or artist
// Returns the appropriate context URI if successful, empty string if it's likely a track
func tryAsContextURI(client *client.SpotifyClient, id string) (string, error) {
	// Try playlist first (most common use case)
	_, err := client.Playlists.GetPlaylist(GetCommandContext(), id, nil)
	if err == nil {
		return fmt.Sprintf("spotify:playlist:%s", id), nil
	}

	// Try album
	_, err = client.Albums.GetAlbum(GetCommandContext(), id, "US")
	if err == nil {
		return fmt.Sprintf("spotify:album:%s", id), nil
	}

	// Try artist (less common for direct playback, but possible)
	_, err = client.Artists.GetArtist(GetCommandContext(), id)
	if err == nil {
		return fmt.Sprintf("spotify:artist:%s", id), nil
	}

	// If none of the above worked, it's likely a track or invalid ID
	return "", fmt.Errorf("unable to determine context type for ID: %s", id)
}