package spotify

import (
	"context"
	"fmt"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/bambithedeer/spotify-api/internal/models"
)

// AlbumsService handles album-related operations
type AlbumsService struct {
	client    *api.RequestBuilder
	validator *api.Validator
}

// NewAlbumsService creates a new albums service
func NewAlbumsService(client *api.RequestBuilder) *AlbumsService {
	return &AlbumsService{
		client:    client,
		validator: api.NewValidator(),
	}
}

// GetAlbum gets an album by ID
func (s *AlbumsService) GetAlbum(ctx context.Context, albumID string, market string) (*models.Album, error) {
	if err := s.validator.ValidateSpotifyID(albumID); err != nil {
		return nil, err
	}

	params := api.QueryParams{}
	if market != "" {
		if err := s.validator.ValidateMarket(market); err != nil {
			return nil, err
		}
		params["market"] = market
	}

	var album models.Album
	err := s.client.Get(ctx, fmt.Sprintf("/albums/%s", albumID), params, &album)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get album")
	}

	return &album, nil
}

// GetAlbums gets multiple albums by their IDs
func (s *AlbumsService) GetAlbums(ctx context.Context, albumIDs []string, market string) ([]models.Album, error) {
	if len(albumIDs) == 0 {
		return nil, errors.NewValidationError("album IDs cannot be empty")
	}

	if len(albumIDs) > 20 {
		return nil, errors.NewValidationError("cannot request more than 20 albums at once")
	}

	// Validate and normalize IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(albumIDs)
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
		Albums []models.Album `json:"albums"`
	}

	err = s.client.Get(ctx, "/albums", params, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get albums")
	}

	return response.Albums, nil
}

// GetAlbumTracks gets tracks for an album with pagination
func (s *AlbumsService) GetAlbumTracks(ctx context.Context, albumID string, options *api.PaginationOptions, market string) (*models.Paging[models.Track], *api.PaginationInfo, error) {
	if err := s.validator.ValidateSpotifyID(albumID); err != nil {
		return nil, nil, err
	}

	params := api.QueryParams{}
	if market != "" {
		if err := s.validator.ValidateMarket(market); err != nil {
			return nil, nil, err
		}
		params["market"] = market
	}

	if options != nil {
		params = options.Merge(params)
		if err := options.ValidateLimit(1, 50); err != nil {
			return nil, nil, err
		}
	}

	var tracks models.Paging[models.Track]
	pagination, err := s.client.GetPaginated(ctx, fmt.Sprintf("/albums/%s/tracks", albumID), params, &tracks)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "failed to get album tracks")
	}

	return &tracks, pagination, nil
}

// GetNewReleases gets new album releases
func (s *AlbumsService) GetNewReleases(ctx context.Context, options *NewReleasesOptions) (*models.Paging[models.Album], *api.PaginationInfo, error) {
	params := api.QueryParams{}

	if options != nil {
		if options.Country != "" {
			if err := s.validator.ValidateMarket(options.Country); err != nil {
				return nil, nil, err
			}
			params["country"] = options.Country
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

	var response struct {
		Albums models.Paging[models.Album] `json:"albums"`
	}

	pagination, err := s.client.GetPaginated(ctx, "/browse/new-releases", params, &response)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "failed to get new releases")
	}

	return &response.Albums, pagination, nil
}

// GetAlbumsByArtist gets albums for a specific artist
func (s *AlbumsService) GetAlbumsByArtist(ctx context.Context, artistID string, options *ArtistAlbumsOptions) (*models.Paging[models.Album], *api.PaginationInfo, error) {
	if err := s.validator.ValidateSpotifyID(artistID); err != nil {
		return nil, nil, err
	}

	params := api.QueryParams{}

	if options != nil {
		if options.IncludeGroups != nil && len(options.IncludeGroups) > 0 {
			if err := s.validateIncludeGroups(options.IncludeGroups); err != nil {
				return nil, nil, err
			}
			params["include_groups"] = strings.Join(options.IncludeGroups, ",")
		}

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

	var albums models.Paging[models.Album]
	pagination, err := s.client.GetPaginated(ctx, fmt.Sprintf("/artists/%s/albums", artistID), params, &albums)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "failed to get artist albums")
	}

	return &albums, pagination, nil
}

// NewReleasesOptions contains options for getting new releases
type NewReleasesOptions struct {
	Country string `json:"country,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Offset  int    `json:"offset,omitempty"`
}

// ArtistAlbumsOptions contains options for getting artist albums
type ArtistAlbumsOptions struct {
	IncludeGroups []string `json:"include_groups,omitempty"`
	Market        string   `json:"market,omitempty"`
	Limit         int      `json:"limit,omitempty"`
	Offset        int      `json:"offset,omitempty"`
}

// validateIncludeGroups validates album include groups
func (s *AlbumsService) validateIncludeGroups(groups []string) error {
	validGroups := map[string]bool{
		"album":        true,
		"single":       true,
		"appears_on":   true,
		"compilation":  true,
	}

	for _, group := range groups {
		if !validGroups[group] {
			return errors.NewValidationError(fmt.Sprintf("invalid include group: %s", group))
		}
	}

	return nil
}