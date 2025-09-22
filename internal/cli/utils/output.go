package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bambithedeer/spotify-api/internal/cli/config"
	"gopkg.in/yaml.v3"
)

// OutputFormat represents the output format
type OutputFormat string

const (
	OutputTextFormat OutputFormat = "text"
	OutputJSONFormat OutputFormat = "json"
	OutputYAMLFormat OutputFormat = "yaml"
)

// Output writes data in the specified format
func Output(data interface{}) error {
	cfg := config.Get()
	format := OutputFormat(cfg.DefaultOutput)

	switch format {
	case OutputJSONFormat:
		return OutputJSON(data)
	case OutputYAMLFormat:
		return OutputYAML(data)
	default:
		return OutputText(data)
	}
}

// OutputJSON writes data as JSON
func OutputJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// OutputYAML writes data as YAML
func OutputYAML(data interface{}) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(data)
}

// OutputText writes data as formatted text
func OutputText(data interface{}) error {
	fmt.Println(data)
	return nil
}

// PrintError prints an error message to stderr
func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

// PrintVerbose prints a verbose message if verbose mode is enabled
func PrintVerbose(format string, args ...interface{}) {
	cfg := config.Get()
	if cfg.Verbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] "+format+"\n", args...)
	}
}

// PrintSuccess prints a success message
func PrintSuccess(format string, args ...interface{}) {
	cfg := config.Get()
	if cfg.ColorOutput {
		fmt.Printf("\033[32m✓\033[0m "+format+"\n", args...)
	} else {
		fmt.Printf("✓ "+format+"\n", args...)
	}
}

// PrintWarning prints a warning message
func PrintWarning(format string, args ...interface{}) {
	cfg := config.Get()
	if cfg.ColorOutput {
		fmt.Printf("\033[33m⚠\033[0m "+format+"\n", args...)
	} else {
		fmt.Printf("⚠ "+format+"\n", args...)
	}
}
