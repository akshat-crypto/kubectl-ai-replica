package main

import (
	"fmt"
	"os"

	"github.com/mcp-servers/cli/internal/cli"
	"github.com/sirupsen/logrus"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func main() {
	// Initialize logger
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Create CLI application
	app := cli.NewApp(Version, Commit, Date)

	// Execute the CLI
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
