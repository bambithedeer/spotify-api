package cli

import (
	"fmt"
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
	browseLimit   int
	browseOffset  int
	browseCountry string
)

// browseCmd represents the browse command
var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "Browse Spotify content",
	Long: `Browse featured content, new releases, and recommendations from Spotify.

Requires authentication with either user account or client credentials.
Use 'auth login' or 'auth client-credentials' to authenticate first.`,
	Example: `  # Browse new album releases
  spotify-cli browse new-releases

  # Browse featured playlists
  spotify-cli browse featured-playlists

  # Browse with specific country/market
  spotify-cli browse new-releases --country US`,
}

var newReleasesCmd = &cobra.Command{
	Use:   "new-releases",
	Short: "Browse new album releases",
	Long:  `Browse the latest album releases available on Spotify.`,
	Example: `  spotify-cli browse new-releases
  spotify-cli browse new-releases --limit 10 --country US`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBrowseNewReleases()
	},
}

var featuredPlaylistsCmd = &cobra.Command{
	Use:   "featured-playlists",
	Short: "Browse featured playlists",
	Long:  `Browse Spotify's featured playlists and editorial recommendations.`,
	Example: `  spotify-cli browse featured-playlists
  spotify-cli browse featured-playlists --limit 20 --country GB`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBrowseFeaturedPlaylists()
	},
}

func init() {
	rootCmd.AddCommand(browseCmd)
	browseCmd.AddCommand(newReleasesCmd)
	browseCmd.AddCommand(featuredPlaylistsCmd)

	// Add flags to browse commands
	for _, cmd := range []*cobra.Command{newReleasesCmd, featuredPlaylistsCmd} {
		cmd.Flags().IntVarP(&browseLimit, "limit", "l", 20, "Number of results to return (1-50)")
		cmd.Flags().IntVarP(&browseOffset, "offset", "", 0, "Offset for pagination")
		cmd.Flags().StringVarP(&browseCountry, "country", "c", "", "Country/market code (e.g., US, GB)")
	}
}

func runBrowseNewReleases() error {
	spotifyClient, err := client.NewSpotifyClient()
	if err != nil {
		return fmt.Errorf("failed to create Spotify client: %w", err)
	}

	if !spotifyClient.IsAuthenticated() {
		return fmt.Errorf("authentication required. Run 'spotify-cli auth login' or 'spotify-cli auth client-credentials'")
	}

	// Create options
	options := &spotify.NewReleasesOptions{
		Country: browseCountry,
		Limit:   browseLimit,
		Offset:  browseOffset,
	}

	albums, pagination, err := spotifyClient.Albums.GetNewReleases(GetCommandContext(), options)
	if err != nil {
		return fmt.Errorf("failed to get new releases: %w", err)
	}

	return outputBrowseResults("new releases", albums, pagination)
}

func runBrowseFeaturedPlaylists() error {
	// Note: This would require implementing the browse endpoints in the API client
	// For now, return a message indicating this is not yet implemented
	fmt.Println("Featured playlists browsing is not yet implemented.")
	fmt.Println("This feature requires the Spotify Browse API endpoints to be implemented.")
	return nil
}

func outputBrowseResults(browseType string, results interface{}, pagination *api.PaginationInfo) error {
	cfg := config.Get()

	// Check output format
	outputFormat := "table"
	if cfg.DefaultOutput == "json" || cfg.DefaultOutput == "yaml" {
		outputFormat = cfg.DefaultOutput
	}

	// For structured output, return the data directly
	if outputFormat == "json" || outputFormat == "yaml" {
		return utils.Output(map[string]interface{}{
			"results":    results,
			"pagination": pagination,
			"browse_info": map[string]interface{}{
				"type":    browseType,
				"limit":   browseLimit,
				"offset":  browseOffset,
				"country": browseCountry,
			},
		})
	}

	// Text-based output
	switch v := results.(type) {
	case *models.Paging[models.Album]:
		return outputNewReleasesTable(v, pagination)
	default:
		return fmt.Errorf("unsupported result type")
	}
}

func outputNewReleasesTable(albums *models.Paging[models.Album], pagination *api.PaginationInfo) error {
	if len(albums.Items) == 0 {
		fmt.Println("No new releases found.")
		return nil
	}

	// Print header
	fmt.Printf("New Releases - Found %d albums", albums.Total)
	if pagination != nil {
		fmt.Printf(" (showing %d-%d)", pagination.Offset+1, pagination.Offset+len(albums.Items))
	}
	fmt.Println()
	fmt.Println()

	// Table format
	fmt.Printf("%-22s %-30s %-25s %-12s %s\n", "ID", "ALBUM", "ARTIST", "RELEASED", "TYPE")
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

		albumType := album.AlbumType
		if albumType == "" {
			albumType = "album"
		}

		fmt.Printf("%-22s %-30s %-25s %-12s %s\n",
			album.ID,
			truncateString(album.Name, 28),
			truncateString(artists, 23),
			released,
			albumType)
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