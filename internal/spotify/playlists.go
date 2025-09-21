package spotify

import (
	"context"
	"fmt"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/bambithedeer/spotify-api/internal/models"
)

// PlaylistsService handles playlist-related operations
type PlaylistsService struct {
	client    *api.RequestBuilder
	validator *api.Validator
}

// NewPlaylistsService creates a new playlists service
func NewPlaylistsService(client *api.RequestBuilder) *PlaylistsService {
	return &PlaylistsService{
		client:    client,
		validator: api.NewValidator(),
	}
}

// GetPlaylist gets a playlist by ID
func (s *PlaylistsService) GetPlaylist(ctx context.Context, playlistID string, options *PlaylistOptions) (*models.Playlist, error) {
	if err := s.validator.ValidateSpotifyID(playlistID); err != nil {
		return nil, err
	}

	params := api.QueryParams{}
	if options != nil {
		if options.Market != "" {
			if err := s.validator.ValidateMarket(options.Market); err != nil {
				return nil, err
			}
			params["market"] = options.Market
		}

		if options.Fields != "" {
			params["fields"] = options.Fields
		}

		if len(options.AdditionalTypes) > 0 {
			if err := s.validateAdditionalTypes(options.AdditionalTypes); err != nil {
				return nil, err
			}
			params["additional_types"] = strings.Join(options.AdditionalTypes, ",")
		}
	}

	var playlist models.Playlist
	err := s.client.Get(ctx, fmt.Sprintf("/playlists/%s", playlistID), params, &playlist)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get playlist")
	}

	return &playlist, nil
}

// GetPlaylistTracks gets tracks for a playlist with pagination
func (s *PlaylistsService) GetPlaylistTracks(ctx context.Context, playlistID string, options *PlaylistTracksOptions) (*models.Paging[models.PlaylistTrack], *api.PaginationInfo, error) {
	if err := s.validator.ValidateSpotifyID(playlistID); err != nil {
		return nil, nil, err
	}

	params := api.QueryParams{}
	if options != nil {
		if options.Market != "" {
			if err := s.validator.ValidateMarket(options.Market); err != nil {
				return nil, nil, err
			}
			params["market"] = options.Market
		}

		if options.Fields != "" {
			params["fields"] = options.Fields
		}

		if options.Limit > 0 {
			if err := s.validator.ValidateLimit(options.Limit, 1, 100); err != nil {
				return nil, nil, err
			}
			params["limit"] = options.Limit
		}

		if options.Offset > 0 {
			if err := s.validator.ValidateOffset(options.Offset); err != nil {
				return nil, nil, err
			}
			params["offset"] = options.Offset
		}

		if len(options.AdditionalTypes) > 0 {
			if err := s.validateAdditionalTypes(options.AdditionalTypes); err != nil {
				return nil, nil, err
			}
			params["additional_types"] = strings.Join(options.AdditionalTypes, ",")
		}
	}

	var tracks models.Paging[models.PlaylistTrack]
	pagination, err := s.client.GetPaginated(ctx, fmt.Sprintf("/playlists/%s/tracks", playlistID), params, &tracks)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "failed to get playlist tracks")
	}

	return &tracks, pagination, nil
}

// GetUserPlaylists gets current user's playlists
func (s *PlaylistsService) GetUserPlaylists(ctx context.Context, options *api.PaginationOptions) (*models.Paging[models.Playlist], *api.PaginationInfo, error) {
	params := api.QueryParams{}
	if options != nil {
		params = options.Merge(params)
		if err := options.ValidateLimit(1, 50); err != nil {
			return nil, nil, err
		}
	}

	var playlists models.Paging[models.Playlist]
	pagination, err := s.client.GetPaginated(ctx, "/me/playlists", params, &playlists)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "failed to get user playlists")
	}

	return &playlists, pagination, nil
}

// GetUserPlaylistsByID gets playlists for a specific user
func (s *PlaylistsService) GetUserPlaylistsByID(ctx context.Context, userID string, options *api.PaginationOptions) (*models.Paging[models.Playlist], *api.PaginationInfo, error) {
	if userID == "" {
		return nil, nil, errors.NewValidationError("user ID cannot be empty")
	}

	params := api.QueryParams{}
	if options != nil {
		params = options.Merge(params)
		if err := options.ValidateLimit(1, 50); err != nil {
			return nil, nil, err
		}
	}

	var playlists models.Paging[models.Playlist]
	pagination, err := s.client.GetPaginated(ctx, fmt.Sprintf("/users/%s/playlists", userID), params, &playlists)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "failed to get user playlists")
	}

	return &playlists, pagination, nil
}

// CreatePlaylist creates a new playlist for the current user
func (s *PlaylistsService) CreatePlaylist(ctx context.Context, userID string, request *CreatePlaylistRequest) (*models.Playlist, error) {
	if userID == "" {
		return nil, errors.NewValidationError("user ID cannot be empty")
	}

	if request == nil {
		return nil, errors.NewValidationError("create playlist request cannot be nil")
	}

	if err := s.validateCreatePlaylistRequest(request); err != nil {
		return nil, err
	}

	var playlist models.Playlist
	err := s.client.Post(ctx, fmt.Sprintf("/users/%s/playlists", userID), request, &playlist)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to create playlist")
	}

	return &playlist, nil
}

// UpdatePlaylist updates playlist details
func (s *PlaylistsService) UpdatePlaylist(ctx context.Context, playlistID string, request *UpdatePlaylistRequest) error {
	if err := s.validator.ValidateSpotifyID(playlistID); err != nil {
		return err
	}

	if request == nil {
		return errors.NewValidationError("update playlist request cannot be nil")
	}

	if err := s.validateUpdatePlaylistRequest(request); err != nil {
		return err
	}

	err := s.client.Put(ctx, fmt.Sprintf("/playlists/%s", playlistID), request, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to update playlist")
	}

	return nil
}

// AddTracksToPlaylist adds tracks to a playlist
func (s *PlaylistsService) AddTracksToPlaylist(ctx context.Context, playlistID string, request *AddTracksRequest) (*models.SnapshotResponse, error) {
	if err := s.validator.ValidateSpotifyID(playlistID); err != nil {
		return nil, err
	}

	if request == nil {
		return nil, errors.NewValidationError("add tracks request cannot be nil")
	}

	if err := s.validateAddTracksRequest(request); err != nil {
		return nil, err
	}

	var response models.SnapshotResponse
	err := s.client.Post(ctx, fmt.Sprintf("/playlists/%s/tracks", playlistID), request, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to add tracks to playlist")
	}

	return &response, nil
}

// RemoveTracksFromPlaylist removes tracks from a playlist
func (s *PlaylistsService) RemoveTracksFromPlaylist(ctx context.Context, playlistID string, request *RemoveTracksRequest) (*models.SnapshotResponse, error) {
	if err := s.validator.ValidateSpotifyID(playlistID); err != nil {
		return nil, err
	}

	if request == nil {
		return nil, errors.NewValidationError("remove tracks request cannot be nil")
	}

	if err := s.validateRemoveTracksRequest(request); err != nil {
		return nil, err
	}

	var response models.SnapshotResponse
	err := s.client.DeleteWithBody(ctx, fmt.Sprintf("/playlists/%s/tracks", playlistID), request, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to remove tracks from playlist")
	}

	return &response, nil
}

// ReorderPlaylistTracks reorders tracks in a playlist
func (s *PlaylistsService) ReorderPlaylistTracks(ctx context.Context, playlistID string, request *ReorderTracksRequest) (*models.SnapshotResponse, error) {
	if err := s.validator.ValidateSpotifyID(playlistID); err != nil {
		return nil, err
	}

	if request == nil {
		return nil, errors.NewValidationError("reorder tracks request cannot be nil")
	}

	if err := s.validateReorderTracksRequest(request); err != nil {
		return nil, err
	}

	var response models.SnapshotResponse
	err := s.client.Put(ctx, fmt.Sprintf("/playlists/%s/tracks", playlistID), request, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to reorder playlist tracks")
	}

	return &response, nil
}

// ReplacePlaylistTracks replaces all tracks in a playlist
func (s *PlaylistsService) ReplacePlaylistTracks(ctx context.Context, playlistID string, trackURIs []string) (*models.SnapshotResponse, error) {
	if err := s.validator.ValidateSpotifyID(playlistID); err != nil {
		return nil, err
	}

	if len(trackURIs) > 100 {
		return nil, errors.NewValidationError("cannot replace with more than 100 tracks at once")
	}

	// Validate track URIs
	for i, uri := range trackURIs {
		if err := s.validator.ValidateSpotifyURI(uri); err != nil {
			return nil, errors.WrapValidationError(err, fmt.Sprintf("invalid track URI at position %d", i))
		}
	}

	request := map[string]interface{}{
		"uris": trackURIs,
	}

	var response models.SnapshotResponse
	err := s.client.Put(ctx, fmt.Sprintf("/playlists/%s/tracks", playlistID), request, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to replace playlist tracks")
	}

	return &response, nil
}

// Request and response types

// PlaylistOptions contains options for getting a playlist
type PlaylistOptions struct {
	Market          string   `json:"market,omitempty"`
	Fields          string   `json:"fields,omitempty"`
	AdditionalTypes []string `json:"additional_types,omitempty"`
}

// PlaylistTracksOptions contains options for getting playlist tracks
type PlaylistTracksOptions struct {
	Market          string   `json:"market,omitempty"`
	Fields          string   `json:"fields,omitempty"`
	Limit           int      `json:"limit,omitempty"`
	Offset          int      `json:"offset,omitempty"`
	AdditionalTypes []string `json:"additional_types,omitempty"`
}

// CreatePlaylistRequest represents a request to create a playlist
type CreatePlaylistRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description,omitempty"`
	Public        *bool  `json:"public,omitempty"`
	Collaborative *bool  `json:"collaborative,omitempty"`
}

// UpdatePlaylistRequest represents a request to update playlist details
type UpdatePlaylistRequest struct {
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	Public        *bool   `json:"public,omitempty"`
	Collaborative *bool   `json:"collaborative,omitempty"`
}

// AddTracksRequest represents a request to add tracks to a playlist
type AddTracksRequest struct {
	URIs     []string `json:"uris"`
	Position *int     `json:"position,omitempty"`
}

// RemoveTracksRequest represents a request to remove tracks from a playlist
type RemoveTracksRequest struct {
	Tracks     []TrackToRemove `json:"tracks"`
	SnapshotID *string         `json:"snapshot_id,omitempty"`
}

// TrackToRemove represents a track to remove from a playlist
type TrackToRemove struct {
	URI       string `json:"uri"`
	Positions []int  `json:"positions,omitempty"`
}

// ReorderTracksRequest represents a request to reorder tracks in a playlist
type ReorderTracksRequest struct {
	RangeStart   int     `json:"range_start"`
	InsertBefore int     `json:"insert_before"`
	RangeLength  *int    `json:"range_length,omitempty"`
	SnapshotID   *string `json:"snapshot_id,omitempty"`
}

// Validation methods

func (s *PlaylistsService) validateAdditionalTypes(types []string) error {
	validTypes := map[string]bool{
		"track":   true,
		"episode": true,
	}

	for _, t := range types {
		if !validTypes[t] {
			return errors.NewValidationError(fmt.Sprintf("invalid additional type: %s", t))
		}
	}

	return nil
}

func (s *PlaylistsService) validateCreatePlaylistRequest(request *CreatePlaylistRequest) error {
	if err := s.validator.ValidatePlaylistName(request.Name); err != nil {
		return err
	}

	if len(request.Description) > 300 {
		return errors.NewValidationError("playlist description cannot exceed 300 characters")
	}

	return nil
}

func (s *PlaylistsService) validateUpdatePlaylistRequest(request *UpdatePlaylistRequest) error {
	if request.Name != nil {
		if err := s.validator.ValidatePlaylistName(*request.Name); err != nil {
			return err
		}
	}

	if request.Description != nil && len(*request.Description) > 300 {
		return errors.NewValidationError("playlist description cannot exceed 300 characters")
	}

	return nil
}

func (s *PlaylistsService) validateAddTracksRequest(request *AddTracksRequest) error {
	if len(request.URIs) == 0 {
		return errors.NewValidationError("track URIs cannot be empty")
	}

	if len(request.URIs) > 100 {
		return errors.NewValidationError("cannot add more than 100 tracks at once")
	}

	for i, uri := range request.URIs {
		if err := s.validator.ValidateSpotifyURI(uri); err != nil {
			return errors.WrapValidationError(err, fmt.Sprintf("invalid track URI at position %d", i))
		}
	}

	if request.Position != nil && *request.Position < 0 {
		return errors.NewValidationError("position cannot be negative")
	}

	return nil
}

func (s *PlaylistsService) validateRemoveTracksRequest(request *RemoveTracksRequest) error {
	if len(request.Tracks) == 0 {
		return errors.NewValidationError("tracks to remove cannot be empty")
	}

	if len(request.Tracks) > 100 {
		return errors.NewValidationError("cannot remove more than 100 tracks at once")
	}

	for i, track := range request.Tracks {
		if err := s.validator.ValidateSpotifyURI(track.URI); err != nil {
			return errors.WrapValidationError(err, fmt.Sprintf("invalid track URI at position %d", i))
		}

		for _, pos := range track.Positions {
			if pos < 0 {
				return errors.NewValidationError(fmt.Sprintf("position cannot be negative at track %d", i))
			}
		}
	}

	return nil
}

func (s *PlaylistsService) validateReorderTracksRequest(request *ReorderTracksRequest) error {
	if request.RangeStart < 0 {
		return errors.NewValidationError("range_start cannot be negative")
	}

	if request.InsertBefore < 0 {
		return errors.NewValidationError("insert_before cannot be negative")
	}

	if request.RangeLength != nil && *request.RangeLength <= 0 {
		return errors.NewValidationError("range_length must be positive")
	}

	return nil
}