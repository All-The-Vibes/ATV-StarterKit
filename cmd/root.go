package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "atv-installer",
	Short: "ATV Starter Kit — Agentic Tool & Workflow installer",
	Long:  "Scaffold a complete GitHub Copilot agentic coding environment into any directory.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
