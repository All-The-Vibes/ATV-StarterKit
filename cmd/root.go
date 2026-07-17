package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	appVersion = "dev"
	appCommit  = "none"
)

// SetVersion is called from main to inject build-time version info.
func SetVersion(version, commit string) {
	appVersion = version
	appCommit = commit
}

var rootCmd = &cobra.Command{
	Use:     "atv-starterkit",
	Short:   "ATV Starter Kit — All The Vibes 2.0",
	Long:    "Scaffold a complete GitHub Copilot agentic coding environment into any directory.",
	Version: "dev",
}

func Execute() {
	rootCmd.Version = appVersion
	rootCmd.SetVersionTemplate(fmt.Sprintf("ATV Starter Kit v%s (%s)\n", appVersion, appCommit))
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
