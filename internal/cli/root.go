package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/bambithedeer/spotify-api/internal/cli/config"
	"github.com/bambithedeer/spotify-api/internal/version"
)

var (
	cfgFile     string
	verbose     bool
	output      string
	configDir   string
	cacheDir    string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "spotify-cli",
	Short: "A command-line interface for Spotify Web API",
	Long: `spotify-cli is a command-line tool that provides access to Spotify's Web API.
You can search for music, manage your library, control playback, and more.

Before using this tool, you'll need to authenticate with Spotify using:
  spotify-cli auth login

Examples:
  spotify-cli search track "bohemian rhapsody"
  spotify-cli library tracks
  spotify-cli player play
  spotify-cli --help`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.spotify-cli/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "output format (text, json, yaml)")
	rootCmd.PersistentFlags().StringVar(&configDir, "config-dir", "", "config directory (default is $HOME/.spotify-cli)")
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "", "cache directory (default is $HOME/.spotify-cli/cache)")

	// Add subcommands
	rootCmd.AddCommand(newVersionCmd())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	// Set default directories
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir = filepath.Join(home, ".spotify-cli")
	}

	if cacheDir == "" {
		cacheDir = filepath.Join(configDir, "cache")
	}

	// Ensure directories exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Initialize config
	if cfgFile == "" {
		cfgFile = filepath.Join(configDir, "config.yaml")
	}

	return config.Init(cfgFile, verbose, output)
}

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print the version, build time, and git commit of spotify-cli",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("spotify-cli %s\n", version.Get().String())
			if verbose {
				fmt.Printf("Version: %s\n", version.Get().Version)
				fmt.Printf("Git Commit: %s\n", version.Get().GitCommit)
				fmt.Printf("Build Time: %s\n", version.Get().BuildTime)
				fmt.Printf("Go Version: %s\n", version.Get().GoVersion)
				fmt.Printf("Platform: %s\n", version.Get().Platform)
			}
		},
	}
}