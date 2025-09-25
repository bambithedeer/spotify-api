package spotify

import (
	"context"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/bambithedeer/spotify-api/internal/models"
)

// UsersService handles user profile and following operations
type UsersService struct {
	client    *api.RequestBuilder
	validator *api.Validator
}

// NewUsersService creates a new users service
func NewUsersService(client *api.RequestBuilder) *UsersService {
	return &UsersService{
		client:    client,
		validator: api.NewValidator(),
	}
}

// GetCurrentUser gets the current user's profile
func (s *UsersService) GetCurrentUser(ctx context.Context) (*models.User, error) {
	var user models.User
	err := s.client.Get(ctx, "/me", nil, &user)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get current user")
	}

	return &user, nil
}

// GetUser gets a user's profile by ID
func (s *UsersService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	if userID == "" {
		return nil, errors.NewValidationError("user ID cannot be empty")
	}

	var user models.User
	err := s.client.Get(ctx, "/users/"+userID, nil, &user)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get user")
	}

	return &user, nil
}

// GetFollowedArtists gets the current user's followed artists
func (s *UsersService) GetFollowedArtists(ctx context.Context, options *FollowedArtistsOptions) (*models.CursorPaging[models.Artist], error) {
	params := api.QueryParams{
		"type": "artist", // Required for followed artists
	}

	if options != nil {
		if options.Limit > 0 {
			if err := s.validator.ValidateLimit(options.Limit, 1, 50); err != nil {
				return nil, err
			}
			params["limit"] = options.Limit
		}

		if options.After != "" {
			params["after"] = options.After
		}
	}

	var response struct {
		Artists models.CursorPaging[models.Artist] `json:"artists"`
	}

	err := s.client.Get(ctx, "/me/following", params, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get followed artists")
	}

	return &response.Artists, nil
}


// FollowArtists follows one or more artists
func (s *UsersService) FollowArtists(ctx context.Context, artistIDs []string) error {
	if len(artistIDs) == 0 {
		return errors.NewValidationError("artist IDs cannot be empty")
	}

	if len(artistIDs) > 50 {
		return errors.NewValidationError("cannot follow more than 50 artists at once")
	}

	// Validate artist IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(artistIDs)
	if err != nil {
		return err
	}

	params := api.QueryParams{
		"type": "artist",
		"ids":  strings.Join(normalizedIDs, ","),
	}

	err = s.client.Put(ctx, "/me/following", params, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to follow artists")
	}

	return nil
}

// UnfollowArtists unfollows one or more artists
func (s *UsersService) UnfollowArtists(ctx context.Context, artistIDs []string) error {
	if len(artistIDs) == 0 {
		return errors.NewValidationError("artist IDs cannot be empty")
	}

	if len(artistIDs) > 50 {
		return errors.NewValidationError("cannot unfollow more than 50 artists at once")
	}

	// Validate artist IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(artistIDs)
	if err != nil {
		return err
	}

	params := api.QueryParams{
		"type": "artist",
		"ids":  strings.Join(normalizedIDs, ","),
	}

	err = s.client.Delete(ctx, "/me/following", params)
	if err != nil {
		return errors.WrapAPIError(err, "failed to unfollow artists")
	}

	return nil
}

// CheckFollowingArtists checks if the current user follows one or more artists
func (s *UsersService) CheckFollowingArtists(ctx context.Context, artistIDs []string) ([]bool, error) {
	if len(artistIDs) == 0 {
		return nil, errors.NewValidationError("artist IDs cannot be empty")
	}

	if len(artistIDs) > 50 {
		return nil, errors.NewValidationError("cannot check more than 50 artists at once")
	}

	// Validate artist IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(artistIDs)
	if err != nil {
		return nil, err
	}

	params := api.QueryParams{
		"type": "artist",
		"ids":  strings.Join(normalizedIDs, ","),
	}

	var following []bool
	err = s.client.Get(ctx, "/me/following/contains", params, &following)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to check following artists")
	}

	return following, nil
}

// GetTopArtists gets the current user's top artists
func (s *UsersService) GetTopArtists(ctx context.Context, options *TopItemsOptions) (*models.Paging[models.Artist], *api.PaginationInfo, error) {
	params := api.QueryParams{}

	if options != nil {
		if options.TimeRange != "" {
			if err := s.validateTimeRange(options.TimeRange); err != nil {
				return nil, nil, err
			}
			params["time_range"] = options.TimeRange
		}

		if options.Limit > 0 {
			if err := s.validator.ValidateLimit(options.Limit, 1, 50); err != nil {
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
	}

	var artists models.Paging[models.Artist]
	pagination, err := s.client.GetPaginated(ctx, "/me/top/artists", params, &artists)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "failed to get top artists")
	}

	return &artists, pagination, nil
}

// GetTopTracks gets the current user's top tracks
func (s *UsersService) GetTopTracks(ctx context.Context, options *TopItemsOptions) (*models.Paging[models.Track], *api.PaginationInfo, error) {
	params := api.QueryParams{}

	if options != nil {
		if options.TimeRange != "" {
			if err := s.validateTimeRange(options.TimeRange); err != nil {
				return nil, nil, err
			}
			params["time_range"] = options.TimeRange
		}

		if options.Limit > 0 {
			if err := s.validator.ValidateLimit(options.Limit, 1, 50); err != nil {
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
	}

	var tracks models.Paging[models.Track]
	pagination, err := s.client.GetPaginated(ctx, "/me/top/tracks", params, &tracks)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "failed to get top tracks")
	}

	return &tracks, pagination, nil
}

// FollowedArtistsOptions contains options for getting followed artists
type FollowedArtistsOptions struct {
	Limit int    `json:"limit,omitempty"`
	After string `json:"after,omitempty"`
}


// TopItemsOptions contains options for getting top artists and tracks
type TopItemsOptions struct {
	TimeRange string `json:"time_range,omitempty"` // short_term, medium_term, long_term
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
}

// validateTimeRange validates the time range parameter
func (s *UsersService) validateTimeRange(timeRange string) error {
	validRanges := map[string]bool{
		"short_term":  true, // ~4 weeks
		"medium_term": true, // ~6 months
		"long_term":   true, // ~several years
	}

	if !validRanges[timeRange] {
		return errors.NewValidationError("invalid time range. Must be one of: short_term, medium_term, long_term")
	}

	return nil
}