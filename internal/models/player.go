package models

// Device represents a Spotify Connect device
type Device struct {
	ID               string `json:"id"`
	IsActive         bool   `json:"is_active"`
	IsPrivateSession bool   `json:"is_private_session"`
	IsRestricted     bool   `json:"is_restricted"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	VolumePercent    int    `json:"volume_percent"`
	SupportsVolume   bool   `json:"supports_volume"`
}

// PlaybackState represents the current playback state
type PlaybackState struct {
	Device               Device      `json:"device"`
	RepeatState          string      `json:"repeat_state"`
	ShuffleState         bool        `json:"shuffle_state"`
	Context              *Context    `json:"context"`
	Timestamp            int64       `json:"timestamp"`
	ProgressMs           int         `json:"progress_ms"`
	IsPlaying            bool        `json:"is_playing"`
	Item                 interface{} `json:"item"` // Can be Track or Episode
	CurrentlyPlayingType string      `json:"currently_playing_type"`
	Actions              Actions     `json:"actions"`
}

// CurrentlyPlaying represents currently playing track/episode
type CurrentlyPlaying struct {
	Context              *Context    `json:"context"`
	Timestamp            int64       `json:"timestamp"`
	ProgressMs           int         `json:"progress_ms"`
	IsPlaying            bool        `json:"is_playing"`
	Item                 interface{} `json:"item"` // Can be Track or Episode
	CurrentlyPlayingType string      `json:"currently_playing_type"`
	Actions              Actions     `json:"actions"`
}

// Context represents playback context
type Context struct {
	Type         string            `json:"type"`
	Href         string            `json:"href"`
	ExternalURLs ExternalURLs      `json:"external_urls"`
	URI          string            `json:"uri"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// Actions represents available playback actions
type Actions struct {
	InterruptingPlayback  bool `json:"interrupting_playback"`
	Pausing               bool `json:"pausing"`
	Resuming              bool `json:"resuming"`
	Seeking               bool `json:"seeking"`
	SkippingNext          bool `json:"skipping_next"`
	SkippingPrev          bool `json:"skipping_prev"`
	TogglingRepeatContext bool `json:"toggling_repeat_context"`
	TogglingRepeatTrack   bool `json:"toggling_repeat_track"`
	TogglingShuffle       bool `json:"toggling_shuffle"`
	TransferringPlayback  bool `json:"transferring_playback"`
}

// PlayHistory represents recently played tracks
type PlayHistory struct {
	Track     Track   `json:"track"`
	PlayedAt  string  `json:"played_at"`
	Context   Context `json:"context"`
}

// Queue represents the user's queue
type Queue struct {
	CurrentlyPlaying interface{}   `json:"currently_playing"` // Can be Track or Episode
	Queue            []interface{} `json:"queue"`              // Can be Track or Episode
}

// DevicesResponse represents the response from the devices endpoint
type DevicesResponse struct {
	Devices []Device `json:"devices"`
}