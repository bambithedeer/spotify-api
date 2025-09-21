package models

// SearchResults represents search results from Spotify API
type SearchResults struct {
	Tracks    *Paging[Track]         `json:"tracks,omitempty"`
	Artists   *Paging[Artist]        `json:"artists,omitempty"`
	Albums    *Paging[SimpleAlbum]   `json:"albums,omitempty"`
	Playlists *Paging[SimplePlaylist] `json:"playlists,omitempty"`
	Shows     *Paging[Show]          `json:"shows,omitempty"`
	Episodes  *Paging[Episode]       `json:"episodes,omitempty"`
	Audiobooks *Paging[Audiobook]    `json:"audiobooks,omitempty"`
}

// SearchType represents the type of content to search for
type SearchType string

const (
	SearchTypeTrack     SearchType = "track"
	SearchTypeAlbum     SearchType = "album"
	SearchTypeArtist    SearchType = "artist"
	SearchTypePlaylist  SearchType = "playlist"
	SearchTypeShow      SearchType = "show"
	SearchTypeEpisode   SearchType = "episode"
	SearchTypeAudiobook SearchType = "audiobook"
)

// Show represents a Spotify podcast show
type Show struct {
	AvailableMarkets []string     `json:"available_markets"`
	Copyrights       []Copyright  `json:"copyrights"`
	Description      string       `json:"description"`
	HTMLDescription  string       `json:"html_description"`
	Explicit         bool         `json:"explicit"`
	ExternalURLs     ExternalURLs `json:"external_urls"`
	Href             string       `json:"href"`
	ID               string       `json:"id"`
	Images           []Image      `json:"images"`
	IsExternallyHosted bool       `json:"is_externally_hosted"`
	Languages        []string     `json:"languages"`
	MediaType        string       `json:"media_type"`
	Name             string       `json:"name"`
	Publisher        string       `json:"publisher"`
	Type             string       `json:"type"`
	URI              string       `json:"uri"`
	TotalEpisodes    int          `json:"total_episodes"`
	Episodes         *Paging[Episode] `json:"episodes,omitempty"`
}

// Episode represents a podcast episode
type Episode struct {
	AudioPreviewURL  string       `json:"audio_preview_url"`
	Description      string       `json:"description"`
	HTMLDescription  string       `json:"html_description"`
	DurationMs       int          `json:"duration_ms"`
	Explicit         bool         `json:"explicit"`
	ExternalURLs     ExternalURLs `json:"external_urls"`
	Href             string       `json:"href"`
	ID               string       `json:"id"`
	Images           []Image      `json:"images"`
	IsExternallyHosted bool       `json:"is_externally_hosted"`
	IsPlayable       bool         `json:"is_playable"`
	Language         string       `json:"language"`
	Languages        []string     `json:"languages"`
	Name             string       `json:"name"`
	ReleaseDate      string       `json:"release_date"`
	ReleaseDatePrecision DatePrecision `json:"release_date_precision"`
	ResumePoint      *ResumePoint `json:"resume_point"`
	Type             string       `json:"type"`
	URI              string       `json:"uri"`
	Restrictions     *Restrictions `json:"restrictions,omitempty"`
	Show             *Show        `json:"show,omitempty"`
}

// ResumePoint represents resume point for episodes
type ResumePoint struct {
	FullyPlayed      bool `json:"fully_played"`
	ResumePositionMs int  `json:"resume_position_ms"`
}

// Audiobook represents a Spotify audiobook
type Audiobook struct {
	Authors          []Author     `json:"authors"`
	AvailableMarkets []string     `json:"available_markets"`
	Copyrights       []Copyright  `json:"copyrights"`
	Description      string       `json:"description"`
	HTMLDescription  string       `json:"html_description"`
	Edition          string       `json:"edition"`
	Explicit         bool         `json:"explicit"`
	ExternalURLs     ExternalURLs `json:"external_urls"`
	Href             string       `json:"href"`
	ID               string       `json:"id"`
	Images           []Image      `json:"images"`
	Languages        []string     `json:"languages"`
	MediaType        string       `json:"media_type"`
	Name             string       `json:"name"`
	Narrators        []Narrator   `json:"narrators"`
	Publisher        string       `json:"publisher"`
	Type             string       `json:"type"`
	URI              string       `json:"uri"`
	TotalChapters    int          `json:"total_chapters"`
	Chapters         *Paging[Chapter] `json:"chapters,omitempty"`
}

// Author represents an audiobook author
type Author struct {
	Name string `json:"name"`
}

// Narrator represents an audiobook narrator
type Narrator struct {
	Name string `json:"name"`
}

// Chapter represents an audiobook chapter
type Chapter struct {
	AudioPreviewURL  string        `json:"audio_preview_url"`
	AvailableMarkets []string      `json:"available_markets"`
	ChapterNumber    int           `json:"chapter_number"`
	Description      string        `json:"description"`
	HTMLDescription  string        `json:"html_description"`
	DurationMs       int           `json:"duration_ms"`
	Explicit         bool          `json:"explicit"`
	ExternalURLs     ExternalURLs  `json:"external_urls"`
	Href             string        `json:"href"`
	ID               string        `json:"id"`
	Images           []Image       `json:"images"`
	IsPlayable       bool          `json:"is_playable"`
	Languages        []string      `json:"languages"`
	Name             string        `json:"name"`
	ReleaseDate      string        `json:"release_date"`
	ReleaseDatePrecision DatePrecision `json:"release_date_precision"`
	ResumePoint      *ResumePoint  `json:"resume_point"`
	Type             string        `json:"type"`
	URI              string        `json:"uri"`
	Restrictions     *Restrictions `json:"restrictions,omitempty"`
	Audiobook        *Audiobook    `json:"audiobook,omitempty"`
}