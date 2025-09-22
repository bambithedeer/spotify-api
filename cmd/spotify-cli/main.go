package main

import (
	"os"

	"github.com/bambithedeer/spotify-api/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}