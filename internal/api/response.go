package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/bambithedeer/spotify-api/internal/errors"
	"github.com/bambithedeer/spotify-api/internal/models"
)

// ResponseHandler handles HTTP responses from Spotify API
type ResponseHandler struct{}

// NewResponseHandler creates a new response handler
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// ParseResponse parses a JSON response into the provided struct
func (rh *ResponseHandler) ParseResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	// Check for successful status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return rh.handleErrorResponse(resp)
	}

	// Parse JSON response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.WrapNetworkError(err, "failed to read response body")
	}

	if len(body) == 0 {
		// Empty response is valid for some endpoints (like DELETE)
		return nil
	}

	// If v is nil, don't attempt to unmarshal (for endpoints that don't return data)
	if v == nil {
		return nil
	}

	if err := json.Unmarshal(body, v); err != nil {
		return errors.WrapAPIError(err, fmt.Sprintf("failed to unmarshal response: %s", string(body)))
	}

	return nil
}

// ParsePaginatedResponse parses a paginated response
func (rh *ResponseHandler) ParsePaginatedResponse(resp *http.Response, v interface{}) (*PaginationInfo, error) {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, rh.handleErrorResponse(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WrapNetworkError(err, "failed to read response body")
	}

	if err := json.Unmarshal(body, v); err != nil {
		return nil, errors.WrapAPIError(err, "failed to unmarshal paginated response")
	}

	// Extract pagination info from Link header or response body
	pagination := &PaginationInfo{}

	// Try to extract from response body if it's a paging object
	var temp map[string]interface{}
	if err := json.Unmarshal(body, &temp); err == nil {
		if href, ok := temp["href"].(string); ok {
			pagination.Current = href
		}
		if next, ok := temp["next"].(string); ok {
			pagination.Next = next
		}
		if prev, ok := temp["previous"].(string); ok {
			pagination.Previous = prev
		}
		if total, ok := temp["total"].(float64); ok {
			pagination.Total = int(total)
		}
		if limit, ok := temp["limit"].(float64); ok {
			pagination.Limit = int(limit)
		}
		if offset, ok := temp["offset"].(float64); ok {
			pagination.Offset = int(offset)
		}
	}

	return pagination, nil
}

// handleErrorResponse handles error responses from the API
func (rh *ResponseHandler) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("HTTP %d: failed to read error response", resp.StatusCode))
	}

	// Try to parse Spotify error format
	var errorResp models.ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err == nil {
		return errors.NewAPIError(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, errorResp.Error.Message))
	}

	// Fallback to generic error message
	return errors.NewAPIError(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)))
}

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	Current  string `json:"current"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Total    int    `json:"total"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// HasNext returns true if there is a next page
func (p *PaginationInfo) HasNext() bool {
	return p.Next != ""
}

// HasPrevious returns true if there is a previous page
func (p *PaginationInfo) HasPrevious() bool {
	return p.Previous != ""
}

// GetNextOffset returns the offset for the next page
func (p *PaginationInfo) GetNextOffset() int {
	if !p.HasNext() {
		return -1
	}

	u, err := url.Parse(p.Next)
	if err != nil {
		return -1
	}

	offsetStr := u.Query().Get("offset")
	if offsetStr == "" {
		return -1
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return -1
	}

	return offset
}

// GetPreviousOffset returns the offset for the previous page
func (p *PaginationInfo) GetPreviousOffset() int {
	if !p.HasPrevious() {
		return -1
	}

	u, err := url.Parse(p.Previous)
	if err != nil {
		return -1
	}

	offsetStr := u.Query().Get("offset")
	if offsetStr == "" {
		return -1
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return -1
	}

	return offset
}