package spotify

import (
	"context"
	"fmt"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/api"
	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/bambithedeer/spotify-api/internal/models"
)

// SearchService handles search operations
type SearchService struct {
	client    *api.RequestBuilder
	validator *api.Validator
}

// NewSearchService creates a new search service
func NewSearchService(client *api.RequestBuilder) *SearchService {
	return &SearchService{
		client:    client,
		validator: api.NewValidator(),
	}
}

// SearchOptions contains options for search requests
type SearchOptions struct {
	Query      string   `json:"q"`
	Types      []string `json:"type"`
	Market     string   `json:"market,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Offset     int      `json:"offset,omitempty"`
	IncludeExternal string `json:"include_external,omitempty"`
}

// SearchResult contains search results for all types
type SearchResult struct {
	Tracks    *models.Paging[models.Track]    `json:"tracks,omitempty"`
	Albums    *models.Paging[models.Album]    `json:"albums,omitempty"`
	Artists   *models.Paging[models.Artist]   `json:"artists,omitempty"`
	Playlists *models.Paging[models.Playlist] `json:"playlists,omitempty"`
	Shows     *models.Paging[models.Show]     `json:"shows,omitempty"`
	Episodes  *models.Paging[models.Episode]  `json:"episodes,omitempty"`
	Audiobooks *models.Paging[models.Audiobook] `json:"audiobooks,omitempty"`
}

// Search performs a general search across multiple types
func (s *SearchService) Search(ctx context.Context, options *SearchOptions) (*SearchResult, error) {
	if err := s.validateSearchOptions(options); err != nil {
		return nil, err
	}

	params := s.buildSearchParams(options)

	var result SearchResult
	err := s.client.Get(ctx, "/search", params, &result)
	if err != nil {
		return nil, errors.WrapAPIError(err, "search request failed")
	}

	return &result, nil
}

// SearchTracks searches for tracks only
func (s *SearchService) SearchTracks(ctx context.Context, query string, options *api.PaginationOptions) (*models.Paging[models.Track], *api.PaginationInfo, error) {
	if err := s.validator.ValidateSearchQuery(query); err != nil {
		return nil, nil, err
	}

	params := api.QueryParams{
		"q":    query,
		"type": "track",
	}

	if options != nil {
		params = options.Merge(params)
		if err := options.ValidateLimit(1, 50); err != nil {
			return nil, nil, err
		}
	}

	var result struct {
		Tracks models.Paging[models.Track] `json:"tracks"`
	}

	pagination, err := s.client.GetPaginated(ctx, "/search", params, &result)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "track search failed")
	}

	return &result.Tracks, pagination, nil
}

// SearchAlbums searches for albums only
func (s *SearchService) SearchAlbums(ctx context.Context, query string, options *api.PaginationOptions) (*models.Paging[models.Album], *api.PaginationInfo, error) {
	if err := s.validator.ValidateSearchQuery(query); err != nil {
		return nil, nil, err
	}

	params := api.QueryParams{
		"q":    query,
		"type": "album",
	}

	if options != nil {
		params = options.Merge(params)
		if err := options.ValidateLimit(1, 50); err != nil {
			return nil, nil, err
		}
	}

	var result struct {
		Albums models.Paging[models.Album] `json:"albums"`
	}

	pagination, err := s.client.GetPaginated(ctx, "/search", params, &result)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "album search failed")
	}

	return &result.Albums, pagination, nil
}

// SearchArtists searches for artists only
func (s *SearchService) SearchArtists(ctx context.Context, query string, options *api.PaginationOptions) (*models.Paging[models.Artist], *api.PaginationInfo, error) {
	if err := s.validator.ValidateSearchQuery(query); err != nil {
		return nil, nil, err
	}

	params := api.QueryParams{
		"q":    query,
		"type": "artist",
	}

	if options != nil {
		params = options.Merge(params)
		if err := options.ValidateLimit(1, 50); err != nil {
			return nil, nil, err
		}
	}

	var result struct {
		Artists models.Paging[models.Artist] `json:"artists"`
	}

	pagination, err := s.client.GetPaginated(ctx, "/search", params, &result)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "artist search failed")
	}

	return &result.Artists, pagination, nil
}

// SearchPlaylists searches for playlists only
func (s *SearchService) SearchPlaylists(ctx context.Context, query string, options *api.PaginationOptions) (*models.Paging[models.Playlist], *api.PaginationInfo, error) {
	if err := s.validator.ValidateSearchQuery(query); err != nil {
		return nil, nil, err
	}

	params := api.QueryParams{
		"q":    query,
		"type": "playlist",
	}

	if options != nil {
		params = options.Merge(params)
		if err := options.ValidateLimit(1, 50); err != nil {
			return nil, nil, err
		}
	}

	var result struct {
		Playlists models.Paging[models.Playlist] `json:"playlists"`
	}

	pagination, err := s.client.GetPaginated(ctx, "/search", params, &result)
	if err != nil {
		return nil, nil, errors.WrapAPIError(err, "playlist search failed")
	}

	return &result.Playlists, pagination, nil
}

// validateSearchOptions validates search options
func (s *SearchService) validateSearchOptions(options *SearchOptions) error {
	if err := s.validator.ValidateSearchQuery(options.Query); err != nil {
		return err
	}

	if err := s.validator.ValidateSearchTypes(options.Types); err != nil {
		return err
	}

	if options.Market != "" {
		if err := s.validator.ValidateMarket(options.Market); err != nil {
			return err
		}
	}

	if options.Limit > 0 {
		if err := s.validator.ValidateLimit(options.Limit, 1, 50); err != nil {
			return err
		}
	}

	if options.Offset > 0 {
		if err := s.validator.ValidateOffset(options.Offset); err != nil {
			return err
		}
	}

	return nil
}

// buildSearchParams builds query parameters for search
func (s *SearchService) buildSearchParams(options *SearchOptions) api.QueryParams {
	params := api.QueryParams{
		"q":    options.Query,
		"type": strings.Join(options.Types, ","),
	}

	if options.Market != "" {
		params["market"] = options.Market
	}

	if options.Limit > 0 {
		params["limit"] = options.Limit
	}

	if options.Offset > 0 {
		params["offset"] = options.Offset
	}

	if options.IncludeExternal != "" {
		params["include_external"] = options.IncludeExternal
	}

	return params
}

// SearchFilter provides a fluent interface for building search queries
type SearchFilter struct {
	query string
}

// NewSearchFilter creates a new search filter
func NewSearchFilter(query string) *SearchFilter {
	return &SearchFilter{query: query}
}

// Artist adds an artist filter
func (f *SearchFilter) Artist(artist string) *SearchFilter {
	f.query += fmt.Sprintf(" artist:%s", artist)
	return f
}

// Album adds an album filter
func (f *SearchFilter) Album(album string) *SearchFilter {
	f.query += fmt.Sprintf(" album:%s", album)
	return f
}

// Track adds a track filter
func (f *SearchFilter) Track(track string) *SearchFilter {
	f.query += fmt.Sprintf(" track:%s", track)
	return f
}

// Year adds a year filter
func (f *SearchFilter) Year(year int) *SearchFilter {
	f.query += fmt.Sprintf(" year:%d", year)
	return f
}

// YearRange adds a year range filter
func (f *SearchFilter) YearRange(startYear, endYear int) *SearchFilter {
	f.query += fmt.Sprintf(" year:%d-%d", startYear, endYear)
	return f
}

// Genre adds a genre filter
func (f *SearchFilter) Genre(genre string) *SearchFilter {
	f.query += fmt.Sprintf(" genre:%s", genre)
	return f
}

// IsNew filters for new releases
func (f *SearchFilter) IsNew() *SearchFilter {
	f.query += " tag:new"
	return f
}

// IsHipster filters for hipster content
func (f *SearchFilter) IsHipster() *SearchFilter {
	f.query += " tag:hipster"
	return f
}

// String returns the complete search query
func (f *SearchFilter) String() string {
	return strings.TrimSpace(f.query)
}