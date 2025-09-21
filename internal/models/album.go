package models

// Album represents a Spotify album
type Album struct {
	AlbumType            string               `json:"album_type"`
	TotalTracks          int                  `json:"total_tracks"`
	AvailableMarkets     []string             `json:"available_markets"`
	ExternalURLs         ExternalURLs         `json:"external_urls"`
	Href                 string               `json:"href"`
	ID                   string               `json:"id"`
	Images               []Image              `json:"images"`
	Name                 string               `json:"name"`
	ReleaseDatePrecision ReleaseDatePrecision `json:",inline"`
	Restrictions         *Restrictions        `json:"restrictions,omitempty"`
	Type                 string               `json:"type"`
	URI                  string               `json:"uri"`
	Copyrights           []Copyright          `json:"copyrights"`
	ExternalIDs          ExternalIDs          `json:"external_ids"`
	Genres               []string             `json:"genres"`
	Label                string               `json:"label"`
	Popularity           int                  `json:"popularity"`
	Artists              []SimpleArtist       `json:"artists"`
	Tracks               Paging[SimpleTrack]  `json:"tracks"`
}

// SimpleAlbum represents a simplified album object
type SimpleAlbum struct {
	AlbumGroup           string               `json:"album_group,omitempty"`
	AlbumType            string               `json:"album_type"`
	Artists              []SimpleArtist       `json:"artists"`
	AvailableMarkets     []string             `json:"available_markets"`
	ExternalURLs         ExternalURLs         `json:"external_urls"`
	Href                 string               `json:"href"`
	ID                   string               `json:"id"`
	Images               []Image              `json:"images"`
	Name                 string               `json:"name"`
	ReleaseDatePrecision ReleaseDatePrecision `json:",inline"`
	Restrictions         *Restrictions        `json:"restrictions,omitempty"`
	Type                 string               `json:"type"`
	URI                  string               `json:"uri"`
	TotalTracks          int                  `json:"total_tracks"`
}

// SavedAlbum represents an album saved in user's library
type SavedAlbum struct {
	AddedAt string `json:"added_at"`
	Album   Album  `json:"album"`
}

// NewReleases represents new album releases
type NewReleases struct {
	Albums Paging[SimpleAlbum] `json:"albums"`
}