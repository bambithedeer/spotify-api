package models

import "time"

// ExternalURLs represents external URLs for Spotify entities
type ExternalURLs struct {
	Spotify string `json:"spotify"`
}

// ExternalIDs represents external IDs for tracks
type ExternalIDs struct {
	ISRC string `json:"isrc,omitempty"`
	EAN  string `json:"ean,omitempty"`
	UPC  string `json:"upc,omitempty"`
}

// Image represents an image with different sizes
type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// Followers represents follower information
type Followers struct {
	Href  string `json:"href"`
	Total int    `json:"total"`
}

// Copyright represents copyright information
type Copyright struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

// Restrictions represents market restrictions
type Restrictions struct {
	Reason string `json:"reason"`
}

// Paging represents a paged response from Spotify API
type Paging[T any] struct {
	Href     string `json:"href"`
	Items    []T    `json:"items"`
	Limit    int    `json:"limit"`
	Next     string `json:"next"`
	Offset   int    `json:"offset"`
	Previous string `json:"previous"`
	Total    int    `json:"total"`
}

// CursorPaging represents cursor-based pagination
type CursorPaging[T any] struct {
	Href    string          `json:"href"`
	Items   []T             `json:"items"`
	Limit   int             `json:"limit"`
	Next    string          `json:"next"`
	Cursors CursorPagingObj `json:"cursors"`
	Total   int             `json:"total,omitempty"`
}

// CursorPagingObj represents cursors for pagination
type CursorPagingObj struct {
	After  string `json:"after"`
	Before string `json:"before,omitempty"`
}

// ErrorResponse represents an error response from Spotify API
type ErrorResponse struct {
	Error ErrorObject `json:"error"`
}

// ErrorObject represents detailed error information
type ErrorObject struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// DatePrecision represents the precision of release dates
type DatePrecision string

const (
	DatePrecisionYear  DatePrecision = "year"
	DatePrecisionMonth DatePrecision = "month"
	DatePrecisionDay   DatePrecision = "day"
)

// ReleaseDatePrecision represents a release date with precision
type ReleaseDatePrecision struct {
	Date      time.Time     `json:"-"`
	DateStr   string        `json:"release_date"`
	Precision DatePrecision `json:"release_date_precision"`
}