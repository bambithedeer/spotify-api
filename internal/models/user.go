package models

// User represents a Spotify user
type User struct {
	Country         string       `json:"country,omitempty"`
	DisplayName     string       `json:"display_name"`
	Email           string       `json:"email,omitempty"`
	ExplicitContent struct {
		FilterEnabled bool `json:"filter_enabled"`
		FilterLocked  bool `json:"filter_locked"`
	} `json:"explicit_content,omitempty"`
	ExternalURLs ExternalURLs `json:"external_urls"`
	Followers    Followers    `json:"followers"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Images       []Image      `json:"images"`
	Product      string       `json:"product,omitempty"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}

// PrivateUser represents current user with private information
type PrivateUser struct {
	User
	Birthdate string `json:"birthdate,omitempty"`
}

// PublicUser represents a public user profile
type PublicUser struct {
	DisplayName  string       `json:"display_name"`
	ExternalURLs ExternalURLs `json:"external_urls"`
	Followers    Followers    `json:"followers"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Images       []Image      `json:"images"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}