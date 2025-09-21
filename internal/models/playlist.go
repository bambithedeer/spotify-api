package models

// Playlist represents a Spotify playlist
type Playlist struct {
	Collaborative bool                    `json:"collaborative"`
	Description   string                  `json:"description"`
	ExternalURLs  ExternalURLs            `json:"external_urls"`
	Followers     Followers               `json:"followers"`
	Href          string                  `json:"href"`
	ID            string                  `json:"id"`
	Images        []Image                 `json:"images"`
	Name          string                  `json:"name"`
	Owner         User                    `json:"owner"`
	Public        bool                    `json:"public"`
	SnapshotID    string                  `json:"snapshot_id"`
	Tracks        Paging[PlaylistTrack]   `json:"tracks"`
	Type          string                  `json:"type"`
	URI           string                  `json:"uri"`
}

// SimplePlaylist represents a simplified playlist object
type SimplePlaylist struct {
	Collaborative bool         `json:"collaborative"`
	Description   string       `json:"description"`
	ExternalURLs  ExternalURLs `json:"external_urls"`
	Href          string       `json:"href"`
	ID            string       `json:"id"`
	Images        []Image      `json:"images"`
	Name          string       `json:"name"`
	Owner         User         `json:"owner"`
	Public        bool         `json:"public"`
	SnapshotID    string       `json:"snapshot_id"`
	Tracks        struct {
		Href  string `json:"href"`
		Total int    `json:"total"`
	} `json:"tracks"`
	Type string `json:"type"`
	URI  string `json:"uri"`
}

// FeaturedPlaylists represents featured playlists response
type FeaturedPlaylists struct {
	Message   string                  `json:"message"`
	Playlists Paging[SimplePlaylist]  `json:"playlists"`
}

// CategoryPlaylists represents playlists for a category
type CategoryPlaylists struct {
	Playlists Paging[SimplePlaylist] `json:"playlists"`
}

// PlaylistSnapshot represents a playlist snapshot after modification
type PlaylistSnapshot struct {
	SnapshotID string `json:"snapshot_id"`
}