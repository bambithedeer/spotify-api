package spotify

import (
	"context"
	"net/url"
	"strconv"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/bambithedeer/spotify-api/internal/models"
)

// PlayerService handles playback control operations
type PlayerService struct {
	client    *api.RequestBuilder
	validator *api.Validator
}

// NewPlayerService creates a new player service
func NewPlayerService(client *api.RequestBuilder) *PlayerService {
	return &PlayerService{
		client:    client,
		validator: api.NewValidator(),
	}
}

// GetPlaybackState gets the current playback state
func (s *PlayerService) GetPlaybackState(ctx context.Context, market string) (*models.PlaybackState, error) {
	params := api.QueryParams{}
	if market != "" {
		if err := s.validator.ValidateMarket(market); err != nil {
			return nil, err
		}
		params["market"] = market
	}

	var state models.PlaybackState
	err := s.client.Get(ctx, "/me/player", params, &state)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get playback state")
	}

	return &state, nil
}

// GetCurrentlyPlaying gets information about the user's current playing track
func (s *PlayerService) GetCurrentlyPlaying(ctx context.Context, options *CurrentlyPlayingOptions) (*models.CurrentlyPlaying, error) {
	params := api.QueryParams{}
	if options != nil {
		if options.Market != "" {
			if err := s.validator.ValidateMarket(options.Market); err != nil {
				return nil, err
			}
			params["market"] = options.Market
		}

		if len(options.AdditionalTypes) > 0 {
			if err := s.validateAdditionalTypes(options.AdditionalTypes); err != nil {
				return nil, err
			}
			params["additional_types"] = options.AdditionalTypes
		}
	}

	var playing models.CurrentlyPlaying
	err := s.client.Get(ctx, "/me/player/currently-playing", params, &playing)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get currently playing")
	}

	return &playing, nil
}

// GetDevices gets the user's available devices
func (s *PlayerService) GetDevices(ctx context.Context) (*models.DevicesResponse, error) {
	var devices models.DevicesResponse
	err := s.client.Get(ctx, "/me/player/devices", nil, &devices)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get devices")
	}

	return &devices, nil
}

// Play starts or resumes playback
func (s *PlayerService) Play(ctx context.Context, options *PlayOptions) error {
	params := api.QueryParams{}
	if options != nil && options.DeviceID != "" {
		params["device_id"] = options.DeviceID
	}

	var body interface{}
	if options != nil {
		body = options
	}

	endpoint := "/me/player/play"
	if options != nil && options.DeviceID != "" {
		endpoint += "?device_id=" + options.DeviceID
	}

	err := s.client.Put(ctx, endpoint, body, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to start playback")
	}

	return nil
}

// Pause pauses playback
func (s *PlayerService) Pause(ctx context.Context, deviceID string) error {
	endpoint := "/me/player/pause"
	if deviceID != "" {
		endpoint += "?device_id=" + deviceID
	}

	err := s.client.Put(ctx, endpoint, nil, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to pause playback")
	}

	return nil
}

// Next skips to next track
func (s *PlayerService) Next(ctx context.Context, deviceID string) error {
	endpoint := "/me/player/next"
	if deviceID != "" {
		endpoint += "?device_id=" + deviceID
	}

	err := s.client.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to skip to next track")
	}

	return nil
}

// Previous skips to previous track
func (s *PlayerService) Previous(ctx context.Context, deviceID string) error {
	endpoint := "/me/player/previous"
	if deviceID != "" {
		endpoint += "?device_id=" + deviceID
	}

	err := s.client.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to skip to previous track")
	}

	return nil
}

// Seek seeks to position in currently playing track
func (s *PlayerService) Seek(ctx context.Context, positionMs int, deviceID string) error {
	if positionMs < 0 {
		return errors.NewValidationError("position cannot be negative")
	}

	endpoint := "/me/player/seek?position_ms=" + strconv.Itoa(positionMs)
	if deviceID != "" {
		endpoint += "&device_id=" + deviceID
	}

	err := s.client.Put(ctx, endpoint, nil, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to seek")
	}

	return nil
}

// SetRepeat sets repeat mode
func (s *PlayerService) SetRepeat(ctx context.Context, state string, deviceID string) error {
	if err := s.validateRepeatState(state); err != nil {
		return err
	}

	endpoint := "/me/player/repeat?state=" + state
	if deviceID != "" {
		endpoint += "&device_id=" + deviceID
	}

	err := s.client.Put(ctx, endpoint, nil, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to set repeat mode")
	}

	return nil
}

// SetShuffle sets shuffle mode
func (s *PlayerService) SetShuffle(ctx context.Context, state bool, deviceID string) error {
	stateStr := "false"
	if state {
		stateStr = "true"
	}

	endpoint := "/me/player/shuffle?state=" + stateStr
	if deviceID != "" {
		endpoint += "&device_id=" + deviceID
	}

	err := s.client.Put(ctx, endpoint, nil, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to set shuffle mode")
	}

	return nil
}

// SetVolume sets playback volume
func (s *PlayerService) SetVolume(ctx context.Context, volumePercent int, deviceID string) error {
	if volumePercent < 0 || volumePercent > 100 {
		return errors.NewValidationError("volume must be between 0 and 100")
	}

	endpoint := "/me/player/volume?volume_percent=" + strconv.Itoa(volumePercent)
	if deviceID != "" {
		endpoint += "&device_id=" + deviceID
	}

	err := s.client.Put(ctx, endpoint, nil, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to set volume")
	}

	return nil
}

// TransferPlayback transfers playback to a new device
func (s *PlayerService) TransferPlayback(ctx context.Context, request *TransferPlaybackRequest) error {
	if request == nil {
		return errors.NewValidationError("transfer playback request cannot be nil")
	}

	if len(request.DeviceIDs) == 0 {
		return errors.NewValidationError("device IDs cannot be empty")
	}

	if len(request.DeviceIDs) > 1 {
		return errors.NewValidationError("only one device ID is supported for transfer")
	}

	err := s.client.Put(ctx, "/me/player", request, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to transfer playback")
	}

	return nil
}

// AddToQueue adds an item to the user's playback queue
func (s *PlayerService) AddToQueue(ctx context.Context, uri string, deviceID string) error {
	if uri == "" {
		return errors.NewValidationError("URI cannot be empty")
	}

	if err := s.validator.ValidateSpotifyURI(uri); err != nil {
		return err
	}

	endpoint := "/me/player/queue?uri=" + url.QueryEscape(uri)
	if deviceID != "" {
		endpoint += "&device_id=" + deviceID
	}

	err := s.client.Post(ctx, endpoint, nil, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to add to queue")
	}

	return nil
}

// GetRecentlyPlayed gets tracks from the user's recently played tracks
func (s *PlayerService) GetRecentlyPlayed(ctx context.Context, options *RecentlyPlayedOptions) (*models.CursorPaging[models.PlayHistory], error) {
	params := api.QueryParams{}

	if options != nil {
		if options.Limit > 0 {
			if err := s.validator.ValidateLimit(options.Limit, 1, 50); err != nil {
				return nil, err
			}
			params["limit"] = options.Limit
		}

		if options.After > 0 {
			params["after"] = strconv.FormatInt(options.After, 10)
		}

		if options.Before > 0 {
			params["before"] = strconv.FormatInt(options.Before, 10)
		}
	}

	var playHistory models.CursorPaging[models.PlayHistory]
	err := s.client.Get(ctx, "/me/player/recently-played", params, &playHistory)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get recently played")
	}

	return &playHistory, nil
}

// Request and response types

// CurrentlyPlayingOptions contains options for getting currently playing track
type CurrentlyPlayingOptions struct {
	Market          string   `json:"market,omitempty"`
	AdditionalTypes []string `json:"additional_types,omitempty"`
}

// PlayOptions contains options for starting playback
type PlayOptions struct {
	DeviceID        string   `json:"-"` // Passed as query param, not in body
	ContextURI      string   `json:"context_uri,omitempty"`
	URIs            []string `json:"uris,omitempty"`
	Offset          *Offset  `json:"offset,omitempty"`
	PositionMs      int      `json:"position_ms,omitempty"`
}

// Offset represents playback offset
type Offset struct {
	Position int    `json:"position,omitempty"`
	URI      string `json:"uri,omitempty"`
}

// TransferPlaybackRequest represents a request to transfer playback
type TransferPlaybackRequest struct {
	DeviceIDs []string `json:"device_ids"`
	Play      *bool    `json:"play,omitempty"`
}

// RecentlyPlayedOptions contains options for getting recently played tracks
type RecentlyPlayedOptions struct {
	Limit  int   `json:"limit,omitempty"`
	After  int64 `json:"after,omitempty"`  // Unix timestamp in milliseconds
	Before int64 `json:"before,omitempty"` // Unix timestamp in milliseconds
}

// Validation methods

func (s *PlayerService) validateAdditionalTypes(types []string) error {
	validTypes := map[string]bool{
		"track":   true,
		"episode": true,
	}

	for _, t := range types {
		if !validTypes[t] {
			return errors.NewValidationError("invalid additional type: " + t)
		}
	}

	return nil
}

func (s *PlayerService) validateRepeatState(state string) error {
	validStates := map[string]bool{
		"track":   true,
		"context": true,
		"off":     true,
	}

	if !validStates[state] {
		return errors.NewValidationError("invalid repeat state. Must be one of: track, context, off")
	}

	return nil
}