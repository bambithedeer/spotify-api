package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGet(t *testing.T) {
	info := Get()

	// Test that we get some version info
	if info.Version == "" {
		t.Error("Expected version to be set")
	}

	if info.GoVersion == "" {
		t.Error("Expected Go version to be set")
	}

	if info.Platform == "" {
		t.Error("Expected platform to be set")
	}

	// Test that Go version matches runtime
	if info.GoVersion != runtime.Version() {
		t.Errorf("Expected Go version %s, got %s", runtime.Version(), info.GoVersion)
	}

	// Test platform format
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != expectedPlatform {
		t.Errorf("Expected platform %s, got %s", expectedPlatform, info.Platform)
	}
}

func TestString(t *testing.T) {
	info := Get()

	// Test with default values
	str := info.String()
	if str != info.Version {
		t.Errorf("Expected string representation to be %s, got %s", info.Version, str)
	}

	// Test with custom git commit
	info.GitCommit = "1234567890abcdef"
	str = info.String()
	expected := info.Version + " (1234567)"
	if str != expected {
		t.Errorf("Expected string representation to be %s, got %s", expected, str)
	}

	// Test with short git commit
	info.GitCommit = "123"
	str = info.String()
	if str != info.Version {
		t.Errorf("Expected string representation to be %s for short commit, got %s", info.Version, str)
	}

	// Test with "unknown" git commit
	info.GitCommit = "unknown"
	str = info.String()
	if str != info.Version {
		t.Errorf("Expected string representation to be %s for unknown commit, got %s", info.Version, str)
	}
}

func TestVersionValues(t *testing.T) {
	info := Get()

	// Test default values
	if info.Version != "dev" {
		t.Errorf("Expected default version 'dev', got %s", info.Version)
	}

	if info.GitCommit != "unknown" {
		t.Errorf("Expected default git commit 'unknown', got %s", info.GitCommit)
	}

	if info.BuildTime != "unknown" {
		t.Errorf("Expected default build time 'unknown', got %s", info.BuildTime)
	}

	// Test that Go version is properly formatted
	if !strings.HasPrefix(info.GoVersion, "go") {
		t.Errorf("Expected Go version to start with 'go', got %s", info.GoVersion)
	}

	// Test that platform contains OS and arch
	parts := strings.Split(info.Platform, "/")
	if len(parts) != 2 {
		t.Errorf("Expected platform format 'os/arch', got %s", info.Platform)
	}

	if parts[0] == "" || parts[1] == "" {
		t.Errorf("Expected non-empty OS and arch in platform %s", info.Platform)
	}
}