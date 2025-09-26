package musicbrainz

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()
	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}

	if client.rateLimiter == nil {
		t.Error("rateLimiter is nil")
	}

	if client.userAgent != UserAgent {
		t.Errorf("expected userAgent %s, got %s", UserAgent, client.userAgent)
	}

	// Clean up
	client.Close()
}

func TestSearchArtist(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()
	defer client.Close()

	// Test with a well-known artist
	resp, err := client.SearchArtist("Radiohead")
	if err != nil {
		t.Fatalf("SearchArtist failed: %v", err)
	}

	if resp == nil {
		t.Fatal("response is nil")
	}

	if len(resp.Artists) == 0 {
		t.Fatal("no artists returned")
	}

	// Check first result
	artist := resp.Artists[0]
	if artist.ID == "" {
		t.Error("artist ID is empty")
	}

	if artist.Name == "" {
		t.Error("artist name is empty")
	}

	t.Logf("Found artist: %s (ID: %s, Score: %d)", artist.Name, artist.ID, artist.Score)
}

func TestGetBestMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()
	defer client.Close()

	artist, err := client.GetBestMatch("The Beatles")
	if err != nil {
		t.Fatalf("GetBestMatch failed: %v", err)
	}

	if artist == nil {
		t.Fatal("artist is nil")
	}

	if artist.ID == "" {
		t.Error("artist ID is empty")
	}

	if artist.Name == "" {
		t.Error("artist name is empty")
	}

	t.Logf("Best match: %s (ID: %s)", artist.Name, artist.ID)
}

func TestGetArtistMBID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()
	defer client.Close()

	mbid, err := client.GetArtistMBID("Pink Floyd")
	if err != nil {
		t.Fatalf("GetArtistMBID failed: %v", err)
	}

	if mbid == "" {
		t.Fatal("MBID is empty")
	}

	// MusicBrainz IDs are UUIDs (36 characters)
	if len(mbid) != 36 {
		t.Errorf("expected MBID length 36, got %d", len(mbid))
	}

	t.Logf("Pink Floyd MBID: %s", mbid)
}

func TestGetBestMatchNotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := NewClient()
	defer client.Close()

	_, err := client.GetBestMatch("ThisArtistDefinitelyDoesNotExist123456789")
	if err == nil {
		t.Error("expected error for non-existent artist")
	}
}

func TestRateLimit(t *testing.T) {
	client := NewClient()
	defer client.Close()

	// Record start time
	start := time.Now()

	// Simulate two consecutive calls (should be rate limited)
	<-client.rateLimiter.C
	<-client.rateLimiter.C

	elapsed := time.Since(start)

	// Should take at least the rate limit duration
	if elapsed < RateLimit {
		t.Errorf("rate limiting not working properly: elapsed %v, expected at least %v", elapsed, RateLimit)
	}
}