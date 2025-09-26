package models

// Track represents a Spotify track
type Track struct {
	Album            *SimpleAlbum   `json:"album,omitempty"`
	Artists          []SimpleArtist `json:"artists"`
	AvailableMarkets []string       `json:"available_markets"`
	DiscNumber       int            `json:"disc_number"`
	DurationMs       int            `json:"duration_ms"`
	Explicit         bool           `json:"explicit"`
	ExternalIDs      ExternalIDs    `json:"external_ids"`
	ExternalURLs     ExternalURLs   `json:"external_urls"`
	Href             string         `json:"href"`
	ID               string         `json:"id"`
	IsPlayable       bool           `json:"is_playable,omitempty"`
	LinkedFrom       *TrackLink     `json:"linked_from,omitempty"`
	Restrictions     *Restrictions  `json:"restrictions,omitempty"`
	Name             string         `json:"name"`
	Popularity       int            `json:"popularity"`
	PreviewURL       string         `json:"preview_url"`
	TrackNumber      int            `json:"track_number"`
	Type             string         `json:"type"`
	URI              string         `json:"uri"`
	IsLocal          bool           `json:"is_local"`
}

// SimpleTrack represents a simplified track object
type SimpleTrack struct {
	Artists          []SimpleArtist `json:"artists"`
	AvailableMarkets []string       `json:"available_markets"`
	DiscNumber       int            `json:"disc_number"`
	DurationMs       int            `json:"duration_ms"`
	Explicit         bool           `json:"explicit"`
	ExternalURLs     ExternalURLs   `json:"external_urls"`
	Href             string         `json:"href"`
	ID               string         `json:"id"`
	IsPlayable       bool           `json:"is_playable,omitempty"`
	LinkedFrom       *TrackLink     `json:"linked_from,omitempty"`
	Restrictions     *Restrictions  `json:"restrictions,omitempty"`
	Name             string         `json:"name"`
	PreviewURL       string         `json:"preview_url"`
	TrackNumber      int            `json:"track_number"`
	Type             string         `json:"type"`
	URI              string         `json:"uri"`
	IsLocal          bool           `json:"is_local"`
}

// TrackLink represents a linked track
type TrackLink struct {
	ExternalURLs ExternalURLs `json:"external_urls"`
	Href         string       `json:"href"`
	ID           string       `json:"id"`
	Type         string       `json:"type"`
	URI          string       `json:"uri"`
}

// SavedTrack represents a track saved in user's library
type SavedTrack struct {
	AddedAt string `json:"added_at"`
	Track   Track  `json:"track"`
}

// PlaylistTrack represents a track in a playlist
type PlaylistTrack struct {
	AddedAt string      `json:"added_at"`
	AddedBy *User       `json:"added_by"`
	IsLocal bool        `json:"is_local"`
	Track   interface{} `json:"track"` // Can be Track or Episode
}

// AudioFeatures represents audio features for a track
type AudioFeatures struct {
	Acousticness     float64 `json:"acousticness"`
	AnalysisURL      string  `json:"analysis_url"`
	Danceability     float64 `json:"danceability"`
	DurationMs       int     `json:"duration_ms"`
	Energy           float64 `json:"energy"`
	ID               string  `json:"id"`
	Instrumentalness float64 `json:"instrumentalness"`
	Key              int     `json:"key"`
	Liveness         float64 `json:"liveness"`
	Loudness         float64 `json:"loudness"`
	Mode             int     `json:"mode"`
	Speechiness      float64 `json:"speechiness"`
	Tempo            float64 `json:"tempo"`
	TimeSignature    int     `json:"time_signature"`
	TrackHref        string  `json:"track_href"`
	Type             string  `json:"type"`
	URI              string  `json:"uri"`
	Valence          float64 `json:"valence"`
}

// AudioAnalysis represents detailed audio analysis
type AudioAnalysis struct {
	Meta     AudioAnalysisMeta      `json:"meta"`
	Track    AudioAnalysisTrack     `json:"track"`
	Bars     []AudioAnalysisSegment `json:"bars"`
	Beats    []AudioAnalysisSegment `json:"beats"`
	Sections []AudioAnalysisSection `json:"sections"`
	Segments []AudioAnalysisSegment `json:"segments"`
	Tatums   []AudioAnalysisSegment `json:"tatums"`
}

// AudioAnalysisMeta represents metadata for audio analysis
type AudioAnalysisMeta struct {
	AnalyzerVersion string  `json:"analyzer_version"`
	Platform        string  `json:"platform"`
	DetailedStatus  string  `json:"detailed_status"`
	StatusCode      int     `json:"status_code"`
	Timestamp       int64   `json:"timestamp"`
	AnalysisTime    float64 `json:"analysis_time"`
	InputTotalTime  float64 `json:"input_total_time"`
}

// AudioAnalysisTrack represents track-level audio analysis
type AudioAnalysisTrack struct {
	NumSamples              int     `json:"num_samples"`
	Duration                float64 `json:"duration"`
	SampleMd5               string  `json:"sample_md5"`
	OffsetSeconds           int     `json:"offset_seconds"`
	WindowSeconds           int     `json:"window_seconds"`
	AnalysisSampleRate      int     `json:"analysis_sample_rate"`
	AnalysisChannels        int     `json:"analysis_channels"`
	EndOfFadeIn             float64 `json:"end_of_fade_in"`
	StartOfFadeOut          float64 `json:"start_of_fade_out"`
	Loudness                float64 `json:"loudness"`
	Tempo                   float64 `json:"tempo"`
	TempoConfidence         float64 `json:"tempo_confidence"`
	TimeSignature           int     `json:"time_signature"`
	TimeSignatureConfidence float64 `json:"time_signature_confidence"`
	Key                     int     `json:"key"`
	KeyConfidence           float64 `json:"key_confidence"`
	Mode                    int     `json:"mode"`
	ModeConfidence          float64 `json:"mode_confidence"`
	Codestring              string  `json:"codestring"`
	CodeVersion             float64 `json:"code_version"`
	Echoprintstring         string  `json:"echoprintstring"`
	EchoprintVersion        float64 `json:"echoprint_version"`
	Synchstring             string  `json:"synchstring"`
	SynchVersion            float64 `json:"synch_version"`
	Rhythmstring            string  `json:"rhythmstring"`
	RhythmVersion           float64 `json:"rhythm_version"`
}

// AudioAnalysisSection represents a section in audio analysis
type AudioAnalysisSection struct {
	Start                   float64 `json:"start"`
	Duration                float64 `json:"duration"`
	Confidence              float64 `json:"confidence"`
	Loudness                float64 `json:"loudness"`
	Tempo                   float64 `json:"tempo"`
	TempoConfidence         float64 `json:"tempo_confidence"`
	Key                     int     `json:"key"`
	KeyConfidence           float64 `json:"key_confidence"`
	Mode                    int     `json:"mode"`
	ModeConfidence          float64 `json:"mode_confidence"`
	TimeSignature           int     `json:"time_signature"`
	TimeSignatureConfidence float64 `json:"time_signature_confidence"`
}

// AudioAnalysisSegment represents a segment in audio analysis
type AudioAnalysisSegment struct {
	Start           float64   `json:"start"`
	Duration        float64   `json:"duration"`
	Confidence      float64   `json:"confidence"`
	LoudnessStart   float64   `json:"loudness_start,omitempty"`
	LoudnessMaxTime float64   `json:"loudness_max_time,omitempty"`
	LoudnessMax     float64   `json:"loudness_max,omitempty"`
	LoudnessEnd     float64   `json:"loudness_end,omitempty"`
	Pitches         []float64 `json:"pitches,omitempty"`
	Timbre          []float64 `json:"timbre,omitempty"`
}
