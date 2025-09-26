package integration

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bambithedeer/spotify-api/internal/lidarr"
	"github.com/bambithedeer/spotify-api/internal/logger"
	"github.com/bambithedeer/spotify-api/internal/musicbrainz"
)

// LidarrIntegration handles the integration between Spotify artists, MusicBrainz, and Lidarr
type LidarrIntegration struct {
	lidarrClient     *lidarr.Client
	musicbrainzClient *musicbrainz.Client
	config           *LidarrConfig
	logger           *logger.Logger
}

// LidarrConfig holds configuration for Lidarr integration
type LidarrConfig struct {
	RootFolderPath    string
	QualityProfileID  int
	MetadataProfileID int
	Monitor           bool
	SearchForMissing  bool
}

// ArtistResult represents the result of adding an artist to Lidarr
type ArtistResult struct {
	ArtistName   string
	SpotifyName  string
	MBID         string
	Success      bool
	Error        error
	LidarrArtist *lidarr.Artist
}

// BatchResult represents the results of a batch artist addition
type BatchResult struct {
	Total     int
	Successes int
	Failures  int
	Results   []ArtistResult
}

// NewLidarrIntegration creates a new Lidarr integration instance
func NewLidarrIntegration(lidarrClient *lidarr.Client, musicbrainzClient *musicbrainz.Client, config *LidarrConfig, log *logger.Logger) *LidarrIntegration {
	return &LidarrIntegration{
		lidarrClient:      lidarrClient,
		musicbrainzClient: musicbrainzClient,
		config:            config,
		logger:            log,
	}
}

// AddArtist adds a single artist to Lidarr
func (li *LidarrIntegration) AddArtist(artistName string) (*ArtistResult, error) {
	result := &ArtistResult{
		SpotifyName: artistName,
		Success:     false,
	}

	// Step 1: Look up artist in MusicBrainz
	li.logger.InfoWithFields("Looking up artist in MusicBrainz", logger.Fields{
		"artist": artistName,
	})
	mbArtist, err := li.musicbrainzClient.GetBestMatch(artistName)
	if err != nil {
		result.Error = fmt.Errorf("MusicBrainz lookup failed: %w", err)
		return result, result.Error
	}

	result.ArtistName = mbArtist.Name
	result.MBID = mbArtist.ID

	li.logger.InfoWithFields("Found MusicBrainz match", logger.Fields{
		"spotify_name": artistName,
		"mb_name":      mbArtist.Name,
		"mbid":         mbArtist.ID,
		"score":        mbArtist.Score,
	})

	// Step 2: Add artist to Lidarr using MBID
	li.logger.InfoWithFields("Adding artist to Lidarr", logger.Fields{"mbid": mbArtist.ID})
	lidarrArtist, err := li.lidarrClient.AddArtistByMBID(
		mbArtist.ID,
		li.config.RootFolderPath,
		li.config.QualityProfileID,
		li.config.MetadataProfileID,
		li.config.Monitor,
		li.config.SearchForMissing,
	)

	if err != nil {
		// Check if error is because artist already exists
		if strings.Contains(strings.ToLower(err.Error()), "already") {
			li.logger.WarnWithFields("Artist already exists in Lidarr", logger.Fields{"artist": mbArtist.Name})
			result.Error = fmt.Errorf("artist already exists: %s", mbArtist.Name)
			return result, result.Error
		}

		result.Error = fmt.Errorf("Lidarr add failed: %w", err)
		return result, result.Error
	}

	result.LidarrArtist = lidarrArtist
	result.Success = true

	li.logger.InfoWithFields("Successfully added artist to Lidarr", logger.Fields{
		"artist":    mbArtist.Name,
		"lidarr_id": lidarrArtist.ID,
	})

	return result, nil
}

// AddArtistsBatch adds multiple artists to Lidarr with concurrent processing
func (li *LidarrIntegration) AddArtistsBatch(artistNames []string, maxConcurrency int) *BatchResult {
	if maxConcurrency <= 0 {
		maxConcurrency = 3 // Default to 3 concurrent requests to be respectful to APIs
	}

	result := &BatchResult{
		Total:   len(artistNames),
		Results: make([]ArtistResult, 0, len(artistNames)),
	}

	// Create worker pool
	jobs := make(chan string, len(artistNames))
	results := make(chan ArtistResult, len(artistNames))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for artistName := range jobs {
				artistResult, err := li.AddArtist(artistName)
				if err != nil {
					// Error is already stored in artistResult
					results <- *artistResult
				} else {
					results <- *artistResult
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, artistName := range artistNames {
			jobs <- artistName
		}
		close(jobs)
	}()

	// Close results channel after all workers complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for artistResult := range results {
		result.Results = append(result.Results, artistResult)
		if artistResult.Success {
			result.Successes++
		} else {
			result.Failures++
		}
	}

	li.logger.InfoWithFields("Batch operation completed", logger.Fields{
		"total":     result.Total,
		"successes": result.Successes,
		"failures":  result.Failures,
	})

	return result
}

// ValidateConfig validates the Lidarr configuration
func (li *LidarrIntegration) ValidateConfig() error {
	// Test Lidarr connection
	if err := li.lidarrClient.TestConnection(); err != nil {
		return fmt.Errorf("Lidarr connection failed: %w", err)
	}

	// Validate root folder exists
	folders, err := li.lidarrClient.GetRootFolders()
	if err != nil {
		return fmt.Errorf("failed to get root folders: %w", err)
	}

	rootFolderExists := false
	for _, folder := range folders {
		if folder.Path == li.config.RootFolderPath {
			rootFolderExists = true
			if !folder.Accessible {
				return fmt.Errorf("root folder not accessible: %s", li.config.RootFolderPath)
			}
			break
		}
	}

	if !rootFolderExists {
		return fmt.Errorf("root folder not found: %s", li.config.RootFolderPath)
	}

	// Validate quality profile
	qualityProfiles, err := li.lidarrClient.GetQualityProfiles()
	if err != nil {
		return fmt.Errorf("failed to get quality profiles: %w", err)
	}

	qualityProfileExists := false
	for _, profile := range qualityProfiles {
		if profile.ID == li.config.QualityProfileID {
			qualityProfileExists = true
			break
		}
	}

	if !qualityProfileExists {
		return fmt.Errorf("quality profile not found: %d", li.config.QualityProfileID)
	}

	// Validate metadata profile
	metadataProfiles, err := li.lidarrClient.GetMetadataProfiles()
	if err != nil {
		return fmt.Errorf("failed to get metadata profiles: %w", err)
	}

	metadataProfileExists := false
	for _, profile := range metadataProfiles {
		if profile.ID == li.config.MetadataProfileID {
			metadataProfileExists = true
			break
		}
	}

	if !metadataProfileExists {
		return fmt.Errorf("metadata profile not found: %d", li.config.MetadataProfileID)
	}

	li.logger.Info("Lidarr configuration validated successfully")
	return nil
}

// GetAvailableProfiles returns available quality and metadata profiles
func (li *LidarrIntegration) GetAvailableProfiles() ([]lidarr.QualityProfile, []lidarr.MetadataProfile, []lidarr.RootFolder, error) {
	qualityProfiles, err := li.lidarrClient.GetQualityProfiles()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get quality profiles: %w", err)
	}

	metadataProfiles, err := li.lidarrClient.GetMetadataProfiles()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get metadata profiles: %w", err)
	}

	rootFolders, err := li.lidarrClient.GetRootFolders()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get root folders: %w", err)
	}

	return qualityProfiles, metadataProfiles, rootFolders, nil
}

// Close cleans up resources
func (li *LidarrIntegration) Close() {
	if li.musicbrainzClient != nil {
		li.musicbrainzClient.Close()
	}
}