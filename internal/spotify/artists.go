package spotify

import (
	"context"
	"fmt"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/bambithedeer/spotify-api/internal/models"
)

// ArtistsService handles artist-related operations
type ArtistsService struct {
	client    *api.RequestBuilder
	validator *api.Validator
}

// NewArtistsService creates a new artists service
func NewArtistsService(client *api.RequestBuilder) *ArtistsService {
	return &ArtistsService{
		client:    client,
		validator: api.NewValidator(),
	}
}

// GetArtist gets an artist by ID
func (s *ArtistsService) GetArtist(ctx context.Context, artistID string) (*models.Artist, error) {
	if err := s.validator.ValidateSpotifyID(artistID); err != nil {
		return nil, err
	}

	var artist models.Artist
	err := s.client.Get(ctx, fmt.Sprintf("/artists/%s", artistID), nil, &artist)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get artist")
	}

	return &artist, nil
}

// GetArtists gets multiple artists by their IDs
func (s *ArtistsService) GetArtists(ctx context.Context, artistIDs []string) ([]models.Artist, error) {
	if len(artistIDs) == 0 {
		return nil, errors.NewValidationError("artist IDs cannot be empty")
	}

	if len(artistIDs) > 50 {
		return nil, errors.NewValidationError("cannot request more than 50 artists at once")
	}

	// Validate and normalize IDs
	normalizedIDs, err := s.validator.NormalizeAndValidateIDs(artistIDs)
	if err != nil {
		return nil, err
	}

	params := api.QueryParams{
		"ids": strings.Join(normalizedIDs, ","),
	}

	var response struct {
		Artists []models.Artist `json:"artists"`
	}

	err = s.client.Get(ctx, "/artists", params, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get artists")
	}

	return response.Artists, nil
}

// GetArtistAlbums gets albums for an artist with filtering options
func (s *ArtistsService) GetArtistAlbums(ctx context.Context, artistID string, options *ArtistAlbumsOptions) (*models.Paging[models.Album], *api.PaginationInfo, error) {
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

// GetArtistTopTracks gets an artist's top tracks by market
func (s *ArtistsService) GetArtistTopTracks(ctx context.Context, artistID string, market string) ([]models.Track, error) {
	if err := s.validator.ValidateSpotifyID(artistID); err != nil {
		return nil, err
	}

	params := api.QueryParams{}
	if market != "" {
		if err := s.validator.ValidateMarket(market); err != nil {
			return nil, err
		}
		params["market"] = market
	} else {
		// Market is required for top tracks endpoint
		return nil, errors.NewValidationError("market parameter is required for top tracks")
	}

	var response struct {
		Tracks []models.Track `json:"tracks"`
	}

	err := s.client.Get(ctx, fmt.Sprintf("/artists/%s/top-tracks", artistID), params, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get artist top tracks")
	}

	return response.Tracks, nil
}

// GetRelatedArtists gets artists related to a given artist
func (s *ArtistsService) GetRelatedArtists(ctx context.Context, artistID string) ([]models.Artist, error) {
	if err := s.validator.ValidateSpotifyID(artistID); err != nil {
		return nil, err
	}

	var response struct {
		Artists []models.Artist `json:"artists"`
	}

	err := s.client.Get(ctx, fmt.Sprintf("/artists/%s/related-artists", artistID), nil, &response)
	if err != nil {
		return nil, errors.WrapAPIError(err, "failed to get related artists")
	}

	return response.Artists, nil
}

// validateIncludeGroups validates album include groups
func (s *ArtistsService) validateIncludeGroups(groups []string) error {
	validGroups := map[string]bool{
		"album":       true,
		"single":      true,
		"appears_on":  true,
		"compilation": true,
	}

	for _, group := range groups {
		if !validGroups[group] {
			return errors.NewValidationError(fmt.Sprintf("invalid include group: %s", group))
		}
	}

	return nil
}