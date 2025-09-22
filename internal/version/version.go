package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time via ldflags
var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

// Info contains version information
type Info struct {
	Version   string `json:"version" yaml:"version"`
	GitCommit string `json:"gitCommit" yaml:"gitCommit"`
	BuildTime string `json:"buildTime" yaml:"buildTime"`
	GoVersion string `json:"goVersion" yaml:"goVersion"`
	Platform  string `json:"platform" yaml:"platform"`
}

// String returns the version as a string
func (i Info) String() string {
	if i.GitCommit != "unknown" && len(i.GitCommit) > 7 {
		return fmt.Sprintf("%s (%s)", i.Version, i.GitCommit[:7])
	}
	return i.Version
}

// Get returns the version information
func Get() Info {
	return Info{
		Version:   version,
		GitCommit: gitCommit,
		BuildTime: buildTime,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}