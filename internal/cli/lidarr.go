package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/client"
	"github.com/bambithedeer/spotify-api/internal/config"
	"github.com/bambithedeer/spotify-api/internal/integration"
	"github.com/bambithedeer/spotify-api/internal/lidarr"
	"github.com/bambithedeer/spotify-api/internal/logger"
	"github.com/bambithedeer/spotify-api/internal/models"
	"github.com/bambithedeer/spotify-api/internal/musicbrainz"
	"github.com/bambithedeer/spotify-api/internal/spotify"
	"github.com/spf13/cobra"
)

var lidarrCmd = &cobra.Command{
	Use:   "lidarr",
	Short: "Lidarr integration commands",
	Long:  "Commands for integrating Spotify data with Lidarr music management",
}

var lidarrAddArtistsCmd = &cobra.Command{
	Use:   "add-artists",
	Short: "Add artists to Lidarr from various sources",
	Long:  "Add artists to Lidarr by looking them up in MusicBrainz and using the MBID",
	RunE:  runLidarrAddArtists,
}

var lidarrImportPlaylistCmd = &cobra.Command{
	Use:   "import-from-playlist",
	Short: "Import artists from a Spotify playlist to Lidarr",
	Long:  "Extract unique artists from a Spotify playlist and add them to Lidarr",
	RunE:  runLidarrImportPlaylist,
}

var lidarrImportSavedCmd = &cobra.Command{
	Use:   "import-saved-artists",
	Short: "Import artists from saved tracks to Lidarr",
	Long:  "Extract unique artists from user's saved tracks and add them to Lidarr",
	RunE:  runLidarrImportSaved,
}

var lidarrTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test Lidarr connection and configuration",
	Long:  "Test connection to Lidarr and validate configuration settings",
	RunE:  runLidarrTest,
}

var lidarrConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Lidarr integration settings",
	Long:  "Interactive configuration of Lidarr connection and import settings",
	RunE:  runLidarrConfig,
}

func init() {
	// Add flags for add-artists command
	lidarrAddArtistsCmd.Flags().StringP("file", "f", "", "File containing list of artist names (one per line)")
	lidarrAddArtistsCmd.Flags().StringSliceP("artists", "a", nil, "Artist names to add (can be used multiple times)")
	lidarrAddArtistsCmd.Flags().BoolP("interactive", "i", false, "Interactive mode to enter artists manually")
	lidarrAddArtistsCmd.Flags().IntP("concurrency", "c", 3, "Number of concurrent requests (1-10)")

	// Add flags for import-from-playlist command
	lidarrImportPlaylistCmd.Flags().StringP("playlist-id", "p", "", "Spotify playlist ID")
	lidarrImportPlaylistCmd.Flags().IntP("limit", "l", 0, "Limit number of tracks to process (0 = all)")
	lidarrImportPlaylistCmd.Flags().IntP("concurrency", "c", 3, "Number of concurrent requests (1-10)")

	// Add flags for import-saved-artists command
	lidarrImportSavedCmd.Flags().IntP("limit", "l", 50, "Limit number of saved tracks to process")
	lidarrImportSavedCmd.Flags().IntP("concurrency", "c", 3, "Number of concurrent requests (1-10)")

	// Override Lidarr config via flags
	for _, cmd := range []*cobra.Command{lidarrAddArtistsCmd, lidarrImportPlaylistCmd, lidarrImportSavedCmd, lidarrTestCmd} {
		cmd.Flags().String("lidarr-url", "", "Lidarr URL (overrides config)")
		cmd.Flags().String("api-key", "", "Lidarr API key (overrides config)")
		cmd.Flags().String("root-folder", "", "Root folder path (overrides config)")
		cmd.Flags().Int("quality-profile", 0, "Quality profile ID (overrides config)")
		cmd.Flags().Int("metadata-profile", 0, "Metadata profile ID (overrides config)")
		cmd.Flags().Bool("monitor", true, "Monitor added artists")
		cmd.Flags().Bool("search", true, "Search for missing albums after adding")
	}

	// Add subcommands to lidarr command
	lidarrCmd.AddCommand(lidarrAddArtistsCmd)
	lidarrCmd.AddCommand(lidarrImportPlaylistCmd)
	lidarrCmd.AddCommand(lidarrImportSavedCmd)
	lidarrCmd.AddCommand(lidarrTestCmd)
	lidarrCmd.AddCommand(lidarrConfigCmd)

	// Add lidarr command to root
	rootCmd.AddCommand(lidarrCmd)
}

func createLidarrIntegration(cmd *cobra.Command) (*integration.LidarrIntegration, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with flags if provided
	if url, _ := cmd.Flags().GetString("lidarr-url"); url != "" {
		cfg.Lidarr.URL = url
	}
	if apiKey, _ := cmd.Flags().GetString("api-key"); apiKey != "" {
		cfg.Lidarr.APIKey = apiKey
	}
	if rootFolder, _ := cmd.Flags().GetString("root-folder"); rootFolder != "" {
		cfg.Lidarr.RootFolderPath = rootFolder
	}
	if qualityProfile, _ := cmd.Flags().GetInt("quality-profile"); qualityProfile > 0 {
		cfg.Lidarr.QualityProfileID = qualityProfile
	}
	if metadataProfile, _ := cmd.Flags().GetInt("metadata-profile"); metadataProfile > 0 {
		cfg.Lidarr.MetadataProfileID = metadataProfile
	}
	if monitor, _ := cmd.Flags().GetBool("monitor"); cmd.Flags().Changed("monitor") {
		cfg.Lidarr.Monitor = monitor
	}
	if search, _ := cmd.Flags().GetBool("search"); cmd.Flags().Changed("search") {
		cfg.Lidarr.SearchForMissing = search
	}

	// Create clients
	lidarrClient := lidarr.NewClient(lidarr.Config{
		BaseURL: cfg.Lidarr.URL,
		APIKey:  cfg.Lidarr.APIKey,
	})

	mbClient := musicbrainz.NewClient()

	log := logger.NewLogger(&logger.Config{
		Level:  cfg.Logging.Level,
		Format: cfg.Logging.Format,
		Output: cfg.Logging.Output,
	})

	integrationConfig := &integration.LidarrConfig{
		RootFolderPath:    cfg.Lidarr.RootFolderPath,
		QualityProfileID:  cfg.Lidarr.QualityProfileID,
		MetadataProfileID: cfg.Lidarr.MetadataProfileID,
		Monitor:           cfg.Lidarr.Monitor,
		SearchForMissing:  cfg.Lidarr.SearchForMissing,
	}

	return integration.NewLidarrIntegration(lidarrClient, mbClient, integrationConfig, log), nil
}

func createSpotifyClientWithServices(cfg *config.Config) (*spotify.PlaylistsService, *spotify.LibraryService, error) {
	// Create the underlying client
	spotifyClient := client.NewClient(cfg.Spotify.ClientID, cfg.Spotify.ClientSecret, cfg.Spotify.RedirectURI)

	// For now, use client credentials flow
	// TODO: In future, handle user authentication for playlists/library access
	if err := spotifyClient.AuthenticateClientCredentials(); err != nil {
		return nil, nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// Create request builder and services
	requestBuilder := api.NewRequestBuilder(spotifyClient)
	playlistsService := spotify.NewPlaylistsService(requestBuilder)
	libraryService := spotify.NewLibraryService(requestBuilder)

	return playlistsService, libraryService, nil
}

func runLidarrAddArtists(cmd *cobra.Command, args []string) error {
	integration, err := createLidarrIntegration(cmd)
	if err != nil {
		return err
	}
	defer integration.Close()

	// Validate configuration
	if err := integration.ValidateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	var artistNames []string

	// Get artists from various sources
	if file, _ := cmd.Flags().GetString("file"); file != "" {
		names, err := readArtistsFromFile(file)
		if err != nil {
			return fmt.Errorf("failed to read artists from file: %w", err)
		}
		artistNames = append(artistNames, names...)
	}

	if artists, _ := cmd.Flags().GetStringSlice("artists"); len(artists) > 0 {
		artistNames = append(artistNames, artists...)
	}

	if interactive, _ := cmd.Flags().GetBool("interactive"); interactive {
		names, err := readArtistsInteractively()
		if err != nil {
			return fmt.Errorf("failed to read artists interactively: %w", err)
		}
		artistNames = append(artistNames, names...)
	}

	if len(artistNames) == 0 {
		return fmt.Errorf("no artists specified. Use --file, --artists, or --interactive")
	}

	// Remove duplicates
	artistNames = removeDuplicates(artistNames)

	fmt.Printf("Adding %d artists to Lidarr...\n", len(artistNames))

	concurrency, _ := cmd.Flags().GetInt("concurrency")
	if concurrency < 1 || concurrency > 10 {
		concurrency = 3
	}

	// Process artists
	result := integration.AddArtistsBatch(artistNames, concurrency)

	// Print results
	printBatchResults(result)

	if result.Failures > 0 {
		return fmt.Errorf("%d artists failed to add", result.Failures)
	}

	return nil
}

func runLidarrImportPlaylist(cmd *cobra.Command, args []string) error {
	playlistID, _ := cmd.Flags().GetString("playlist-id")
	if playlistID == "" {
		return fmt.Errorf("playlist ID is required (use --playlist-id)")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create Spotify services
	playlistsService, _, err := createSpotifyClientWithServices(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	fmt.Printf("Fetching playlist tracks from Spotify...\n")

	// Get playlist tracks with pagination
	ctx := context.Background()
	limit, _ := cmd.Flags().GetInt("limit")

	var allTracks []models.PlaylistTrack
	offset := 0
	pageLimit := 50 // Spotify API max per request

	for {
		options := &spotify.PlaylistTracksOptions{
			Limit:  pageLimit,
			Offset: offset,
		}

		paging, _, err := playlistsService.GetPlaylistTracks(ctx, playlistID, options)
		if err != nil {
			return fmt.Errorf("failed to get playlist tracks: %w", err)
		}


		allTracks = append(allTracks, paging.Items...)

		// Check if we've hit our limit or reached the end
		if limit > 0 && len(allTracks) >= limit {
			allTracks = allTracks[:limit]
			break
		}

		if paging.Next == "" || len(paging.Items) == 0 {
			break
		}

		offset += pageLimit
	}

	// Extract unique artists
	artistSet := make(map[string]bool)
	for _, playlistTrack := range allTracks {
		// Handle JSON track data (map[string]interface{})
		if trackMap, ok := playlistTrack.Track.(map[string]interface{}); ok {
			if artistsData, exists := trackMap["artists"]; exists {
				if artistsSlice, ok := artistsData.([]interface{}); ok {
					for _, artistData := range artistsSlice {
						if artistMap, ok := artistData.(map[string]interface{}); ok {
							if name, ok := artistMap["name"].(string); ok && name != "" {
								artistSet[name] = true
							}
						}
					}
				}
			}
		}
	}

	var artistNames []string
	for artistName := range artistSet {
		artistNames = append(artistNames, artistName)
	}

	fmt.Printf("Found %d unique artists in playlist from %d tracks\n", len(artistNames), len(allTracks))

	if len(artistNames) == 0 {
		return fmt.Errorf("no artists found in playlist")
	}

	// Create Lidarr integration
	integration, err := createLidarrIntegration(cmd)
	if err != nil {
		return err
	}
	defer integration.Close()

	// Validate configuration
	if err := integration.ValidateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	concurrency, _ := cmd.Flags().GetInt("concurrency")
	if concurrency < 1 || concurrency > 10 {
		concurrency = 3
	}

	// Process artists
	result := integration.AddArtistsBatch(artistNames, concurrency)

	// Print results
	printBatchResults(result)

	if result.Failures > 0 {
		return fmt.Errorf("%d artists failed to add", result.Failures)
	}

	return nil
}

func runLidarrImportSaved(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create Spotify services
	_, libraryService, err := createSpotifyClientWithServices(cfg)
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	fmt.Printf("Fetching saved tracks from Spotify...\n")

	// Get saved tracks with pagination
	ctx := context.Background()
	limit, _ := cmd.Flags().GetInt("limit")

	var allSavedTracks []models.SavedTrack
	offset := 0
	pageLimit := 50 // Spotify API max per request

	for {
		options := &api.PaginationOptions{
			Limit:  pageLimit,
			Offset: offset,
		}

		paging, _, err := libraryService.GetSavedTracks(ctx, options)
		if err != nil {
			return fmt.Errorf("failed to get saved tracks: %w", err)
		}

		allSavedTracks = append(allSavedTracks, paging.Items...)

		// Check if we've hit our limit or reached the end
		if limit > 0 && len(allSavedTracks) >= limit {
			allSavedTracks = allSavedTracks[:limit]
			break
		}

		if paging.Next == "" || len(paging.Items) == 0 {
			break
		}

		offset += pageLimit
	}

	// Extract unique artists
	artistSet := make(map[string]bool)
	for _, savedTrack := range allSavedTracks {
		for _, artist := range savedTrack.Track.Artists {
			if artist.Name != "" {
				artistSet[artist.Name] = true
			}
		}
	}

	var artistNames []string
	for artistName := range artistSet {
		artistNames = append(artistNames, artistName)
	}

	fmt.Printf("Found %d unique artists in saved tracks from %d tracks\n", len(artistNames), len(allSavedTracks))

	if len(artistNames) == 0 {
		return fmt.Errorf("no artists found in saved tracks")
	}

	// Create Lidarr integration
	integration, err := createLidarrIntegration(cmd)
	if err != nil {
		return err
	}
	defer integration.Close()

	// Validate configuration
	if err := integration.ValidateConfig(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	concurrency, _ := cmd.Flags().GetInt("concurrency")
	if concurrency < 1 || concurrency > 10 {
		concurrency = 3
	}

	// Process artists
	result := integration.AddArtistsBatch(artistNames, concurrency)

	// Print results
	printBatchResults(result)

	if result.Failures > 0 {
		return fmt.Errorf("%d artists failed to add", result.Failures)
	}

	return nil
}

func runLidarrTest(cmd *cobra.Command, args []string) error {
	integration, err := createLidarrIntegration(cmd)
	if err != nil {
		return err
	}
	defer integration.Close()

	fmt.Println("Testing Lidarr connection...")

	// Test basic connection
	if err := integration.ValidateConfig(); err != nil {
		fmt.Printf("❌ Configuration validation failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Lidarr connection successful")

	// Get and display available profiles and folders
	qualityProfiles, metadataProfiles, rootFolders, err := integration.GetAvailableProfiles()
	if err != nil {
		return fmt.Errorf("failed to get profiles: %w", err)
	}

	fmt.Println("\n📁 Available Root Folders:")
	for _, folder := range rootFolders {
		status := "✅"
		if !folder.Accessible {
			status = "❌"
		}
		fmt.Printf("  %s %s (ID: %d)\n", status, folder.Path, folder.ID)
	}

	fmt.Println("\n🎵 Available Quality Profiles:")
	for _, profile := range qualityProfiles {
		fmt.Printf("  - %s (ID: %d)\n", profile.Name, profile.ID)
	}

	fmt.Println("\n📊 Available Metadata Profiles:")
	for _, profile := range metadataProfiles {
		fmt.Printf("  - %s (ID: %d)\n", profile.Name, profile.ID)
	}

	return nil
}

func runLidarrConfig(cmd *cobra.Command, args []string) error {
	fmt.Println("Interactive Lidarr Configuration")
	fmt.Println("================================")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)

	// URL
	fmt.Printf("Lidarr URL [%s]: ", cfg.Lidarr.URL)
	if input, err := reader.ReadString('\n'); err == nil {
		input = strings.TrimSpace(input)
		if input != "" {
			cfg.Lidarr.URL = input
		}
	}

	// API Key
	fmt.Printf("API Key [%s]: ", maskString(cfg.Lidarr.APIKey))
	if input, err := reader.ReadString('\n'); err == nil {
		input = strings.TrimSpace(input)
		if input != "" {
			cfg.Lidarr.APIKey = input
		}
	}

	// Test connection
	fmt.Println("\nTesting connection...")
	lidarrClient := lidarr.NewClient(lidarr.Config{
		BaseURL: cfg.Lidarr.URL,
		APIKey:  cfg.Lidarr.APIKey,
	})

	if err := lidarrClient.TestConnection(); err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Connection successful")

	// Get available options
	qualityProfiles, err := lidarrClient.GetQualityProfiles()
	if err != nil {
		return fmt.Errorf("failed to get quality profiles: %w", err)
	}

	metadataProfiles, err := lidarrClient.GetMetadataProfiles()
	if err != nil {
		return fmt.Errorf("failed to get metadata profiles: %w", err)
	}

	rootFolders, err := lidarrClient.GetRootFolders()
	if err != nil {
		return fmt.Errorf("failed to get root folders: %w", err)
	}

	// Root folder selection
	fmt.Println("\nAvailable Root Folders:")
	for i, folder := range rootFolders {
		fmt.Printf("  %d. %s\n", i+1, folder.Path)
	}
	fmt.Printf("Select root folder [1-%d]: ", len(rootFolders))
	if input, err := reader.ReadString('\n'); err == nil {
		if idx, err := strconv.Atoi(strings.TrimSpace(input)); err == nil && idx > 0 && idx <= len(rootFolders) {
			cfg.Lidarr.RootFolderPath = rootFolders[idx-1].Path
		}
	}

	// Quality profile selection
	fmt.Println("\nAvailable Quality Profiles:")
	for i, profile := range qualityProfiles {
		fmt.Printf("  %d. %s\n", i+1, profile.Name)
	}
	fmt.Printf("Select quality profile [1-%d]: ", len(qualityProfiles))
	if input, err := reader.ReadString('\n'); err == nil {
		if idx, err := strconv.Atoi(strings.TrimSpace(input)); err == nil && idx > 0 && idx <= len(qualityProfiles) {
			cfg.Lidarr.QualityProfileID = qualityProfiles[idx-1].ID
		}
	}

	// Metadata profile selection
	fmt.Println("\nAvailable Metadata Profiles:")
	for i, profile := range metadataProfiles {
		fmt.Printf("  %d. %s\n", i+1, profile.Name)
	}
	fmt.Printf("Select metadata profile [1-%d]: ", len(metadataProfiles))
	if input, err := reader.ReadString('\n'); err == nil {
		if idx, err := strconv.Atoi(strings.TrimSpace(input)); err == nil && idx > 0 && idx <= len(metadataProfiles) {
			cfg.Lidarr.MetadataProfileID = metadataProfiles[idx-1].ID
		}
	}

	// Monitor option
	fmt.Printf("Monitor new artists? [Y/n]: ")
	if input, err := reader.ReadString('\n'); err == nil {
		input = strings.ToLower(strings.TrimSpace(input))
		cfg.Lidarr.Monitor = input != "n" && input != "no"
	}

	// Search for missing
	fmt.Printf("Search for missing albums? [Y/n]: ")
	if input, err := reader.ReadString('\n'); err == nil {
		input = strings.ToLower(strings.TrimSpace(input))
		cfg.Lidarr.SearchForMissing = input != "n" && input != "no"
	}

	// Save configuration to standard location
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "spotify-cli")
	configPath := filepath.Join(configDir, "config.yaml")

	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("\n✅ Configuration saved to %s\n", configPath)
	return nil
}

// Helper functions

func readArtistsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var artists []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			artists = append(artists, line)
		}
	}

	return artists, scanner.Err()
}

func readArtistsInteractively() ([]string, error) {
	fmt.Println("Enter artist names (one per line, empty line to finish):")

	reader := bufio.NewReader(os.Stdin)
	var artists []string

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		artists = append(artists, line)
	}

	return artists, nil
}

func removeDuplicates(artists []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, artist := range artists {
		if !seen[artist] {
			seen[artist] = true
			result = append(result, artist)
		}
	}

	return result
}

func printBatchResults(result *integration.BatchResult) {
	fmt.Printf("\n📊 Results Summary:\n")
	fmt.Printf("  Total: %d\n", result.Total)
	fmt.Printf("  ✅ Successes: %d\n", result.Successes)
	fmt.Printf("  ❌ Failures: %d\n", result.Failures)

	if result.Failures > 0 {
		fmt.Println("\n❌ Failed Artists:")
		for _, artistResult := range result.Results {
			if !artistResult.Success {
				fmt.Printf("  - %s: %v\n", artistResult.SpotifyName, artistResult.Error)
			}
		}
	}

	if result.Successes > 0 {
		fmt.Println("\n✅ Successfully Added:")
		for _, artistResult := range result.Results {
			if artistResult.Success {
				fmt.Printf("  - %s → %s (MBID: %s)\n", artistResult.SpotifyName, artistResult.ArtistName, artistResult.MBID)
			}
		}
	}
}
