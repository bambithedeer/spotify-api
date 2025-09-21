package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/bambithedeer/spotify-api/internal/client"
	"github.com/bambithedeer/spotify-api/internal/errors"
)

// RequestBuilder builds HTTP requests for Spotify API
type RequestBuilder struct {
	client          *client.Client
	responseHandler *ResponseHandler
}

// NewRequestBuilder creates a new request builder
func NewRequestBuilder(client *client.Client) *RequestBuilder {
	return &RequestBuilder{
		client:          client,
		responseHandler: NewResponseHandler(),
	}
}

// QueryParams represents query parameters for API requests
type QueryParams map[string]interface{}

// ToURLValues converts QueryParams to url.Values
func (qp QueryParams) ToURLValues() url.Values {
	values := url.Values{}
	for key, value := range qp {
		switch v := value.(type) {
		case string:
			if v != "" {
				values.Set(key, v)
			}
		case int:
			values.Set(key, strconv.Itoa(v))
		case bool:
			values.Set(key, strconv.FormatBool(v))
		case []string:
			if len(v) > 0 {
				values.Set(key, strings.Join(v, ","))
			}
		case []int:
			if len(v) > 0 {
				strs := make([]string, len(v))
				for i, num := range v {
					strs[i] = strconv.Itoa(num)
				}
				values.Set(key, strings.Join(strs, ","))
			}
		default:
			// Convert to string as fallback
			values.Set(key, fmt.Sprintf("%v", v))
		}
	}
	return values
}

// PaginationOptions contains options for paginated requests
type PaginationOptions struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// Merge merges pagination options into query parameters
func (po *PaginationOptions) Merge(params QueryParams) QueryParams {
	if params == nil {
		params = make(QueryParams)
	}

	if po.Limit > 0 {
		params["limit"] = po.Limit
	}
	if po.Offset > 0 {
		params["offset"] = po.Offset
	}

	return params
}

// ValidateLimit validates the limit parameter against Spotify API constraints
func (po *PaginationOptions) ValidateLimit(min, max int) error {
	if po.Limit < min || po.Limit > max {
		return errors.NewValidationError(fmt.Sprintf("limit must be between %d and %d", min, max))
	}
	return nil
}

// Get performs a GET request
func (rb *RequestBuilder) Get(ctx context.Context, endpoint string, params QueryParams, result interface{}) error {
	url := rb.buildURL(endpoint, params)
	resp, err := rb.client.Get(ctx, url)
	if err != nil {
		return err
	}

	return rb.responseHandler.ParseResponse(resp, result)
}

// GetPaginated performs a GET request and returns pagination info
func (rb *RequestBuilder) GetPaginated(ctx context.Context, endpoint string, params QueryParams, result interface{}) (*PaginationInfo, error) {
	url := rb.buildURL(endpoint, params)
	resp, err := rb.client.Get(ctx, url)
	if err != nil {
		return nil, err
	}

	return rb.responseHandler.ParsePaginatedResponse(resp, result)
}

// Post performs a POST request with JSON body
func (rb *RequestBuilder) Post(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return errors.WrapValidationError(err, "failed to marshal request body")
		}
		bodyReader = strings.NewReader(string(jsonBody))
	}

	resp, err := rb.client.Post(ctx, endpoint, bodyReader)
	if err != nil {
		return err
	}

	return rb.responseHandler.ParseResponse(resp, result)
}

// Put performs a PUT request with JSON body
func (rb *RequestBuilder) Put(ctx context.Context, endpoint string, body interface{}, result interface{}) error {
	var bodyReader io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return errors.WrapValidationError(err, "failed to marshal request body")
		}
		bodyReader = strings.NewReader(string(jsonBody))
	}

	resp, err := rb.client.Put(ctx, endpoint, bodyReader)
	if err != nil {
		return err
	}

	return rb.responseHandler.ParseResponse(resp, result)
}

// Delete performs a DELETE request
func (rb *RequestBuilder) Delete(ctx context.Context, endpoint string, params QueryParams) error {
	url := rb.buildURL(endpoint, params)
	resp, err := rb.client.Delete(ctx, url)
	if err != nil {
		return err
	}

	return rb.responseHandler.ParseResponse(resp, nil)
}

// buildURL builds the complete URL with query parameters
func (rb *RequestBuilder) buildURL(endpoint string, params QueryParams) string {
	if params == nil || len(params) == 0 {
		return endpoint
	}

	// Add query parameters
	values := params.ToURLValues()
	if len(values) == 0 {
		return endpoint
	}

	separator := "?"
	if strings.Contains(endpoint, "?") {
		separator = "&"
	}

	return endpoint + separator + values.Encode()
}

// Batch represents a batch of operations
type Batch struct {
	operations []BatchOperation
}

// BatchOperation represents a single operation in a batch
type BatchOperation struct {
	Method   string
	Endpoint string
	Params   QueryParams
	Body     interface{}
}

// NewBatch creates a new batch
func NewBatch() *Batch {
	return &Batch{
		operations: make([]BatchOperation, 0),
	}
}

// AddGet adds a GET operation to the batch
func (b *Batch) AddGet(endpoint string, params QueryParams) *Batch {
	b.operations = append(b.operations, BatchOperation{
		Method:   "GET",
		Endpoint: endpoint,
		Params:   params,
	})
	return b
}

// AddPost adds a POST operation to the batch
func (b *Batch) AddPost(endpoint string, body interface{}) *Batch {
	b.operations = append(b.operations, BatchOperation{
		Method:   "POST",
		Endpoint: endpoint,
		Body:     body,
	})
	return b
}

// AddPut adds a PUT operation to the batch
func (b *Batch) AddPut(endpoint string, body interface{}) *Batch {
	b.operations = append(b.operations, BatchOperation{
		Method:   "PUT",
		Endpoint: endpoint,
		Body:     body,
	})
	return b
}

// AddDelete adds a DELETE operation to the batch
func (b *Batch) AddDelete(endpoint string, params QueryParams) *Batch {
	b.operations = append(b.operations, BatchOperation{
		Method:   "DELETE",
		Endpoint: endpoint,
		Params:   params,
	})
	return b
}

// Execute executes all operations in the batch
// Note: Spotify API doesn't support true batch requests, so this executes them sequentially
func (b *Batch) Execute(ctx context.Context, rb *RequestBuilder) ([]interface{}, []error) {
	results := make([]interface{}, len(b.operations))
	errors := make([]error, len(b.operations))

	for i, op := range b.operations {
		var result interface{}
		var err error

		switch op.Method {
		case "GET":
			err = rb.Get(ctx, op.Endpoint, op.Params, &result)
		case "POST":
			err = rb.Post(ctx, op.Endpoint, op.Body, &result)
		case "PUT":
			err = rb.Put(ctx, op.Endpoint, op.Body, &result)
		case "DELETE":
			err = rb.Delete(ctx, op.Endpoint, op.Params)
		}

		results[i] = result
		errors[i] = err

		// Stop on first error for now (could be configurable)
		if err != nil {
			break
		}
	}

	return results, errors
}