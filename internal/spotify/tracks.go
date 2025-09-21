package spotify

import (
	"context"
	"fmt"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/bambithedeer/spotify-api/internal/models"
)

// TracksService handles track-related operations
type TracksService struct {
	client    *api.RequestBuilder
	validator *api.Validator
}

// NewTracksService creates a new tracks service
func NewTracksService(client *api.RequestBuilder) *TracksService {
	return &TracksService{
		client:    client,
		validator: api.NewValidator(),
	}
}

// GetTrack gets a track by ID
func (s *TracksService) GetTrack(ctx context.Context, trackID string, market string) (*models.Track, error) {
	if err := s.validator.ValidateSpotifyID(trackID); err != nil {
		return nil, err
	}

	params := api.QueryParams{}
	if market != "" {
		if err := s.validator.ValidateMarket(market); err != nil {
			return nil, err
		}
		params["market"] = market
	}

	var track models.Track
	err := s.client.Get(ctx, fmt.Sprintf("/tracks/%s", trackID), params, &track)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get track")
	}

	return &track, nil
}

// GetTracks gets multiple tracks by their IDs
func (s *TracksService) GetTracks(ctx context.Context, trackIDs []string, market string) ([]models.Track, error) {
	if len(trackIDs) == 0 {
		return nil, errors.NewValidationError("track IDs cannot be empty")
	}

	if len(trackIDs) > 50 {
		return nil, errors.NewValidationError("cannot request more than 50 tracks at once")
	}

	// Validate and normalize IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(trackIDs)
	if err != nil {
		return nil, err
	}

	params := api.QueryParams{
		"ids": strings.Join(normalizedIDs, ","),
	}

	if market != "" {
		if err := s.validator.ValidateMarket(market); err != nil {
			return nil, err
		}
		params["market"] = market
	}

	var response struct {
		Tracks []models.Track `json:"tracks"`
	}

	err = s.client.Get(ctx, "/tracks", params, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get tracks")
	}

	return response.Tracks, nil
}

// GetTrackAudioFeatures gets audio features for a track
func (s *TracksService) GetTrackAudioFeatures(ctx context.Context, trackID string) (*models.AudioFeatures, error) {
	if err := s.validator.ValidateSpotifyID(trackID); err != nil {
		return nil, err
	}

	var audioFeatures models.AudioFeatures
	err := s.client.Get(ctx, fmt.Sprintf("/audio-features/%s", trackID), nil, &audioFeatures)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get track audio features")
	}

	return &audioFeatures, nil
}

// GetTracksAudioFeatures gets audio features for multiple tracks
func (s *TracksService) GetTracksAudioFeatures(ctx context.Context, trackIDs []string) ([]models.AudioFeatures, error) {
	if len(trackIDs) == 0 {
		return nil, errors.NewValidationError("track IDs cannot be empty")
	}

	if len(trackIDs) > 100 {
		return nil, errors.NewValidationError("cannot request more than 100 track audio features at once")
	}

	// Validate and normalize IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(trackIDs)
	if err != nil {
		return nil, err
	}

	params := api.QueryParams{
		"ids": strings.Join(normalizedIDs, ","),
	}

	var response struct {
		AudioFeatures []models.AudioFeatures `json:"audio_features"`
	}

	err = s.client.Get(ctx, "/audio-features", params, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get tracks audio features")
	}

	return response.AudioFeatures, nil
}

// GetTrackAudioAnalysis gets audio analysis for a track
func (s *TracksService) GetTrackAudioAnalysis(ctx context.Context, trackID string) (*models.AudioAnalysis, error) {
	if err := s.validator.ValidateSpotifyID(trackID); err != nil {
		return nil, err
	}

	var audioAnalysis models.AudioAnalysis
	err := s.client.Get(ctx, fmt.Sprintf("/audio-analysis/%s", trackID), nil, &audioAnalysis)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get track audio analysis")
	}

	return &audioAnalysis, nil
}

// GetRecommendations gets track recommendations based on seeds
func (s *TracksService) GetRecommendations(ctx context.Context, options *RecommendationOptions) (*models.Recommendations, error) {
	if options == nil {
		return nil, errors.NewValidationError("recommendation options cannot be nil")
	}

	if err := s.validateRecommendationOptions(options); err != nil {
		return nil, err
	}

	params := s.buildRecommendationParams(options)

	var recommendations models.Recommendations
	err := s.client.Get(ctx, "/recommendations", params, &recommendations)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get recommendations")
	}

	return &recommendations, nil
}

// RecommendationOptions contains options for getting recommendations
type RecommendationOptions struct {
	SeedArtists  []string                   `json:"seed_artists,omitempty"`
	SeedGenres   []string                   `json:"seed_genres,omitempty"`
	SeedTracks   []string                   `json:"seed_tracks,omitempty"`
	Limit        int                        `json:"limit,omitempty"`
	Market       string                     `json:"market,omitempty"`
	AudioFeatures map[string]interface{}    `json:"audio_features,omitempty"`
}

// validateRecommendationOptions validates recommendation options
func (s *TracksService) validateRecommendationOptions(options *RecommendationOptions) error {
	// At least one seed is required
	totalSeeds := len(options.SeedArtists) + len(options.SeedGenres) + len(options.SeedTracks)
	if totalSeeds == 0 {
		return errors.NewValidationError("at least one seed (artist, genre, or track) is required")
	}

	// Maximum 5 seeds total
	if totalSeeds > 5 {
		return errors.NewValidationError("maximum 5 seeds allowed in total")
	}

	// Validate artist seed IDs
	if len(options.SeedArtists) > 0 {
		for _, artistID := range options.SeedArtists {
			if err := s.validator.ValidateSpotifyID(artistID); err != nil {
				return errors.WrapValidationError(err, "invalid seed artist ID")
			}
		}
	}

	// Validate track seed IDs
	if len(options.SeedTracks) > 0 {
		for _, trackID := range options.SeedTracks {
			if err := s.validator.ValidateSpotifyID(trackID); err != nil {
				return errors.WrapValidationError(err, "invalid seed track ID")
			}
		}
	}

	// Validate genre seeds (basic validation - genres are predefined by Spotify)
	if len(options.SeedGenres) > 0 {
		for _, genre := range options.SeedGenres {
			if genre == "" {
				return errors.NewValidationError("genre seed cannot be empty")
			}
		}
	}

	// Validate limit
	if options.Limit > 0 {
		if err := s.validator.ValidateLimit(options.Limit, 1, 100); err != nil {
			return err
		}
	}

	// Validate market
	if options.Market != "" {
		if err := s.validator.ValidateMarket(options.Market); err != nil {
			return err
		}
	}

	return nil
}

// buildRecommendationParams builds query parameters for recommendations
func (s *TracksService) buildRecommendationParams(options *RecommendationOptions) api.QueryParams {
	params := api.QueryParams{}

	if len(options.SeedArtists) > 0 {
		params["seed_artists"] = strings.Join(options.SeedArtists, ",")
	}

	if len(options.SeedGenres) > 0 {
		params["seed_genres"] = strings.Join(options.SeedGenres, ",")
	}

	if len(options.SeedTracks) > 0 {
		params["seed_tracks"] = strings.Join(options.SeedTracks, ",")
	}

	if options.Limit > 0 {
		params["limit"] = options.Limit
	}

	if options.Market != "" {
		params["market"] = options.Market
	}

	// Add audio features tuning parameters
	if options.AudioFeatures != nil {
		validParams := map[string]bool{
			"min_acousticness":     true, "max_acousticness":     true, "target_acousticness":     true,
			"min_danceability":     true, "max_danceability":     true, "target_danceability":     true,
			"min_duration_ms":      true, "max_duration_ms":      true, "target_duration_ms":      true,
			"min_energy":           true, "max_energy":           true, "target_energy":           true,
			"min_instrumentalness": true, "max_instrumentalness": true, "target_instrumentalness": true,
			"min_key":              true, "max_key":              true, "target_key":              true,
			"min_liveness":         true, "max_liveness":         true, "target_liveness":         true,
			"min_loudness":         true, "max_loudness":         true, "target_loudness":         true,
			"min_mode":             true, "max_mode":             true, "target_mode":             true,
			"min_popularity":       true, "max_popularity":       true, "target_popularity":       true,
			"min_speechiness":      true, "max_speechiness":      true, "target_speechiness":      true,
			"min_tempo":            true, "max_tempo":            true, "target_tempo":            true,
			"min_time_signature":   true, "max_time_signature":   true, "target_time_signature":   true,
			"min_valence":          true, "max_valence":          true, "target_valence":          true,
		}

		for key, value := range options.AudioFeatures {
			if validParams[key] {
				params[key] = value
			}
		}
	}

	return params
}