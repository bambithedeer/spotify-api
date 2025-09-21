package models

import (
	"encoding/json"
	"testing"
)

func TestTrackUnmarshal(t *testing.T) {
	trackJSON := `{
		"album": {
			"album_type": "album",
			"artists": [{"id": "artist1", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:artist1", "href": "https://api.spotify.com/v1/artists/artist1", "external_urls": {"spotify": "https://open.spotify.com/artist/artist1"}}],
			"available_markets": ["US", "CA"],
			"external_urls": {"spotify": "https://open.spotify.com/album/album1"},
			"href": "https://api.spotify.com/v1/albums/album1",
			"id": "album1",
			"images": [{"url": "https://example.com/image.jpg", "height": 640, "width": 640}],
			"name": "Test Album",
			"release_date": "2023-01-01",
			"release_date_precision": "day",
			"type": "album",
			"uri": "spotify:album:album1",
			"total_tracks": 10
		},
		"artists": [{"id": "artist1", "name": "Test Artist", "type": "artist", "uri": "spotify:artist:artist1", "href": "https://api.spotify.com/v1/artists/artist1", "external_urls": {"spotify": "https://open.spotify.com/artist/artist1"}}],
		"available_markets": ["US", "CA"],
		"disc_number": 1,
		"duration_ms": 180000,
		"explicit": false,
		"external_ids": {"isrc": "TEST123456789"},
		"external_urls": {"spotify": "https://open.spotify.com/track/track1"},
		"href": "https://api.spotify.com/v1/tracks/track1",
		"id": "track1",
		"name": "Test Track",
		"popularity": 75,
		"preview_url": "https://example.com/preview.mp3",
		"track_number": 1,
		"type": "track",
		"uri": "spotify:track:track1",
		"is_local": false
	}`

	var track Track
	err := json.Unmarshal([]byte(trackJSON), &track)
	if err != nil {
		t.Fatalf("Failed to unmarshal track: %v", err)
	}

	if track.ID != "track1" {
		t.Errorf("Expected track ID 'track1', got %s", track.ID)
	}

	if track.Name != "Test Track" {
		t.Errorf("Expected track name 'Test Track', got %s", track.Name)
	}

	if track.DurationMs != 180000 {
		t.Errorf("Expected duration 180000ms, got %d", track.DurationMs)
	}

	if track.Album == nil {
		t.Fatal("Expected album to be present")
	}

	if track.Album.Name != "Test Album" {
		t.Errorf("Expected album name 'Test Album', got %s", track.Album.Name)
	}
}

func TestArtistUnmarshal(t *testing.T) {
	artistJSON := `{
		"external_urls": {"spotify": "https://open.spotify.com/artist/artist1"},
		"followers": {"href": null, "total": 1000000},
		"genres": ["pop", "rock"],
		"href": "https://api.spotify.com/v1/artists/artist1",
		"id": "artist1",
		"images": [{"url": "https://example.com/artist.jpg", "height": 640, "width": 640}],
		"name": "Test Artist",
		"popularity": 85,
		"type": "artist",
		"uri": "spotify:artist:artist1"
	}`

	var artist Artist
	err := json.Unmarshal([]byte(artistJSON), &artist)
	if err != nil {
		t.Fatalf("Failed to unmarshal artist: %v", err)
	}

	if artist.ID != "artist1" {
		t.Errorf("Expected artist ID 'artist1', got %s", artist.ID)
	}

	if artist.Name != "Test Artist" {
		t.Errorf("Expected artist name 'Test Artist', got %s", artist.Name)
	}

	if artist.Popularity != 85 {
		t.Errorf("Expected popularity 85, got %d", artist.Popularity)
	}

	if len(artist.Genres) != 2 {
		t.Errorf("Expected 2 genres, got %d", len(artist.Genres))
	}

	expectedGenres := []string{"pop", "rock"}
	for i, genre := range artist.Genres {
		if genre != expectedGenres[i] {
			t.Errorf("Expected genre '%s', got '%s'", expectedGenres[i], genre)
		}
	}
}

func TestPlaylistUnmarshal(t *testing.T) {
	playlistJSON := `{
		"collaborative": false,
		"description": "A test playlist",
		"external_urls": {"spotify": "https://open.spotify.com/playlist/playlist1"},
		"followers": {"href": null, "total": 100},
		"href": "https://api.spotify.com/v1/playlists/playlist1",
		"id": "playlist1",
		"images": [{"url": "https://example.com/playlist.jpg", "height": 640, "width": 640}],
		"name": "Test Playlist",
		"owner": {
			"display_name": "Test User",
			"external_urls": {"spotify": "https://open.spotify.com/user/user1"},
			"followers": {"href": null, "total": 50},
			"href": "https://api.spotify.com/v1/users/user1",
			"id": "user1",
			"images": [],
			"type": "user",
			"uri": "spotify:user:user1"
		},
		"public": true,
		"snapshot_id": "snapshot123",
		"tracks": {
			"href": "https://api.spotify.com/v1/playlists/playlist1/tracks",
			"items": [],
			"limit": 100,
			"next": null,
			"offset": 0,
			"previous": null,
			"total": 0
		},
		"type": "playlist",
		"uri": "spotify:playlist:playlist1"
	}`

	var playlist Playlist
	err := json.Unmarshal([]byte(playlistJSON), &playlist)
	if err != nil {
		t.Fatalf("Failed to unmarshal playlist: %v", err)
	}

	if playlist.ID != "playlist1" {
		t.Errorf("Expected playlist ID 'playlist1', got %s", playlist.ID)
	}

	if playlist.Name != "Test Playlist" {
		t.Errorf("Expected playlist name 'Test Playlist', got %s", playlist.Name)
	}

	if playlist.Owner.ID != "user1" {
		t.Errorf("Expected owner ID 'user1', got %s", playlist.Owner.ID)
	}

	if !playlist.Public {
		t.Error("Expected playlist to be public")
	}

	if playlist.Collaborative {
		t.Error("Expected playlist to not be collaborative")
	}
}

func TestPagingUnmarshal(t *testing.T) {
	pagingJSON := `{
		"href": "https://api.spotify.com/v1/tracks",
		"items": [
			{"id": "track1", "name": "Track 1", "type": "track", "uri": "spotify:track:track1", "href": "https://api.spotify.com/v1/tracks/track1", "external_urls": {"spotify": "https://open.spotify.com/track/track1"}, "artists": [], "available_markets": [], "disc_number": 1, "duration_ms": 180000, "explicit": false, "external_ids": {}, "popularity": 50, "preview_url": null, "track_number": 1, "is_local": false},
			{"id": "track2", "name": "Track 2", "type": "track", "uri": "spotify:track:track2", "href": "https://api.spotify.com/v1/tracks/track2", "external_urls": {"spotify": "https://open.spotify.com/track/track2"}, "artists": [], "available_markets": [], "disc_number": 1, "duration_ms": 200000, "explicit": false, "external_ids": {}, "popularity": 60, "preview_url": null, "track_number": 2, "is_local": false}
		],
		"limit": 20,
		"next": "https://api.spotify.com/v1/tracks?offset=20&limit=20",
		"offset": 0,
		"previous": null,
		"total": 100
	}`

	var paging Paging[Track]
	err := json.Unmarshal([]byte(pagingJSON), &paging)
	if err != nil {
		t.Fatalf("Failed to unmarshal paging: %v", err)
	}

	if paging.Total != 100 {
		t.Errorf("Expected total 100, got %d", paging.Total)
	}

	if paging.Limit != 20 {
		t.Errorf("Expected limit 20, got %d", paging.Limit)
	}

	if paging.Offset != 0 {
		t.Errorf("Expected offset 0, got %d", paging.Offset)
	}

	if len(paging.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(paging.Items))
	}

	if paging.Items[0].ID != "track1" {
		t.Errorf("Expected first track ID 'track1', got %s", paging.Items[0].ID)
	}

	if paging.Items[1].ID != "track2" {
		t.Errorf("Expected second track ID 'track2', got %s", paging.Items[1].ID)
	}
}

func TestAudioFeaturesUnmarshal(t *testing.T) {
	audioFeaturesJSON := `{
		"danceability": 0.735,
		"energy": 0.578,
		"key": 5,
		"loudness": -11.840,
		"mode": 0,
		"speechiness": 0.0461,
		"acousticness": 0.514,
		"instrumentalness": 0.0902,
		"liveness": 0.159,
		"valence": 0.624,
		"tempo": 98.002,
		"type": "audio_features",
		"id": "track1",
		"uri": "spotify:track:track1",
		"track_href": "https://api.spotify.com/v1/tracks/track1",
		"analysis_url": "https://api.spotify.com/v1/audio-analysis/track1",
		"duration_ms": 255349,
		"time_signature": 4
	}`

	var features AudioFeatures
	err := json.Unmarshal([]byte(audioFeaturesJSON), &features)
	if err != nil {
		t.Fatalf("Failed to unmarshal audio features: %v", err)
	}

	if features.ID != "track1" {
		t.Errorf("Expected ID 'track1', got %s", features.ID)
	}

	if features.Danceability != 0.735 {
		t.Errorf("Expected danceability 0.735, got %f", features.Danceability)
	}

	if features.Energy != 0.578 {
		t.Errorf("Expected energy 0.578, got %f", features.Energy)
	}

	if features.Tempo != 98.002 {
		t.Errorf("Expected tempo 98.002, got %f", features.Tempo)
	}

	if features.TimeSignature != 4 {
		t.Errorf("Expected time signature 4, got %d", features.TimeSignature)
	}
}

func TestDatePrecisionHandling(t *testing.T) {
	tests := []struct {
		name      string
		precision DatePrecision
		expected  DatePrecision
	}{
		{"Year precision", DatePrecisionYear, DatePrecisionYear},
		{"Month precision", DatePrecisionMonth, DatePrecisionMonth},
		{"Day precision", DatePrecisionDay, DatePrecisionDay},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.precision != tt.expected {
				t.Errorf("Expected precision %s, got %s", tt.expected, tt.precision)
			}
		})
	}
}

func TestSearchTypesConstants(t *testing.T) {
	expectedTypes := map[SearchType]string{
		SearchTypeTrack:     "track",
		SearchTypeAlbum:     "album",
		SearchTypeArtist:    "artist",
		SearchTypePlaylist:  "playlist",
		SearchTypeShow:      "show",
		SearchTypeEpisode:   "episode",
		SearchTypeAudiobook: "audiobook",
	}

	for searchType, expected := range expectedTypes {
		if string(searchType) != expected {
			t.Errorf("Expected search type %s, got %s", expected, string(searchType))
		}
	}
}