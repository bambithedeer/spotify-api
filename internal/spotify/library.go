package spotify

import (
	"context"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/bambithedeer/spotify-api/internal/models"
)

// LibraryService handles user library operations
type LibraryService struct {
	client    *api.RequestBuilder
	validator *api.Validator
}

// NewLibraryService creates a new library service
func NewLibraryService(client *api.RequestBuilder) *LibraryService {
	return &LibraryService{
		client:    client,
		validator: api.NewValidator(),
	}
}

// GetSavedTracks gets the user's saved tracks
func (s *LibraryService) GetSavedTracks(ctx context.Context, options *api.PaginationOptions) (*models.Paging[models.SavedTrack], *api.PaginationInfo, error) {
	params := api.QueryParams{}
	if options != nil {
		params = options.Merge(params)
		if err := options.ValidateLimit(1, 50); err != nil {
			return nil, nil, err
		}
	}

	var tracks models.Paging[models.SavedTrack]
	pagination, err := s.client.GetPaginated(ctx, "/me/tracks", params, &tracks)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "failed to get saved tracks")
	}

	return &tracks, pagination, nil
}

// SaveTracks saves tracks to the user's library
func (s *LibraryService) SaveTracks(ctx context.Context, trackIDs []string) error {
	if len(trackIDs) == 0 {
		return errors.NewValidationError("track IDs cannot be empty")
	}

	if len(trackIDs) > 50 {
		return errors.NewValidationError("cannot save more than 50 tracks at once")
	}

	// Validate and normalize IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(trackIDs)
	if err != nil {
		return err
	}

	params := api.QueryParams{
		"ids": strings.Join(normalizedIDs, ","),
	}

	err = s.client.Put(ctx, "/me/tracks", params, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to save tracks")
	}

	return nil
}

// RemoveTracks removes tracks from the user's library
func (s *LibraryService) RemoveTracks(ctx context.Context, trackIDs []string) error {
	if len(trackIDs) == 0 {
		return errors.NewValidationError("track IDs cannot be empty")
	}

	if len(trackIDs) > 50 {
		return errors.NewValidationError("cannot remove more than 50 tracks at once")
	}

	// Validate and normalize IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(trackIDs)
	if err != nil {
		return err
	}

	params := api.QueryParams{
		"ids": strings.Join(normalizedIDs, ","),
	}

	err = s.client.Delete(ctx, "/me/tracks", params)
	if err != nil {
		return errors.WrapAPIError(err, "failed to remove tracks")
	}

	return nil
}

// CheckSavedTracks checks if tracks are saved in the user's library
func (s *LibraryService) CheckSavedTracks(ctx context.Context, trackIDs []string) ([]bool, error) {
	if len(trackIDs) == 0 {
		return nil, errors.NewValidationError("track IDs cannot be empty")
	}

	if len(trackIDs) > 50 {
		return nil, errors.NewValidationError("cannot check more than 50 tracks at once")
	}

	// Validate and normalize IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(trackIDs)
	if err != nil {
		return nil, err
	}

	params := api.QueryParams{
		"ids": strings.Join(normalizedIDs, ","),
	}

	var saved []bool
	err = s.client.Get(ctx, "/me/tracks/contains", params, &saved)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to check saved tracks")
	}

	return saved, nil
}

// GetSavedAlbums gets the user's saved albums
func (s *LibraryService) GetSavedAlbums(ctx context.Context, options *SavedAlbumsOptions) (*models.Paging[models.SavedAlbum], *api.PaginationInfo, error) {
	params := api.QueryParams{}
	if options != nil {
		if options.Market != "" {
			if err := s.validator.ValidateMarket(options.Market); err != nil {
				return nil, nil, err
			}
			params["market"] = options.Market
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

	var albums models.Paging[models.SavedAlbum]
	pagination, err := s.client.GetPaginated(ctx, "/me/albums", params, &albums)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "failed to get saved albums")
	}

	return &albums, pagination, nil
}

// SaveAlbums saves albums to the user's library
func (s *LibraryService) SaveAlbums(ctx context.Context, albumIDs []string) error {
	if len(albumIDs) == 0 {
		return errors.NewValidationError("album IDs cannot be empty")
	}

	if len(albumIDs) > 50 {
		return errors.NewValidationError("cannot save more than 50 albums at once")
	}

	// Validate and normalize IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(albumIDs)
	if err != nil {
		return err
	}

	params := api.QueryParams{
		"ids": strings.Join(normalizedIDs, ","),
	}

	err = s.client.Put(ctx, "/me/albums", params, nil)
	if err != nil {
		return errors.WrapAPIError(err, "failed to save albums")
	}

	return nil
}

// RemoveAlbums removes albums from the user's library
func (s *LibraryService) RemoveAlbums(ctx context.Context, albumIDs []string) error {
	if len(albumIDs) == 0 {
		return errors.NewValidationError("album IDs cannot be empty")
	}

	if len(albumIDs) > 50 {
		return errors.NewValidationError("cannot remove more than 50 albums at once")
	}

	// Validate and normalize IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(albumIDs)
	if err != nil {
		return err
	}

	params := api.QueryParams{
		"ids": strings.Join(normalizedIDs, ","),
	}

	err = s.client.Delete(ctx, "/me/albums", params)
	if err != nil {
		return errors.WrapAPIError(err, "failed to remove albums")
	}

	return nil
}

// CheckSavedAlbums checks if albums are saved in the user's library
func (s *LibraryService) CheckSavedAlbums(ctx context.Context, albumIDs []string) ([]bool, error) {
	if len(albumIDs) == 0 {
		return nil, errors.NewValidationError("album IDs cannot be empty")
	}

	if len(albumIDs) > 50 {
		return nil, errors.NewValidationError("cannot check more than 50 albums at once")
	}

	// Validate and normalize IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(albumIDs)
	if err != nil {
		return nil, err
	}

	params := api.QueryParams{
		"ids": strings.Join(normalizedIDs, ","),
	}

	var saved []bool
	err = s.client.Get(ctx, "/me/albums/contains", params, &saved)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to check saved albums")
	}

	return saved, nil
}

// SavedAlbumsOptions contains options for getting saved albums
type SavedAlbumsOptions struct {
	Market string `json:"market,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}