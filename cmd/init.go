package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/github/atv-installer/pkg/detect"
	"github.com/github/atv-installer/pkg/output"
	"github.com/github/atv-installer/pkg/scaffold"
	"github.com/github/atv-installer/pkg/tui"
	"github.com/spf13/cobra"
)

var guided bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize ATV Starter Kit in current directory",
	Long: `Scaffold a complete agentic coding environment with all 6 Copilot lifecycle hooks.

Default: auto-detects your stack and installs everything (zero questions).
Use --guided for interactive mode with component selection.`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&guided, "guided", false, "Interactive mode with component selection")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	printer := output.NewPrinter()
	targetDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Show banner
	printer.PrintBanner()

	// Phase 1: Detect environment
	env := detect.DetectEnvironment(targetDir)
	printer.PrintDetection(env)

	// Phase 2: Determine components
	var catalog []scaffold.Component

	if guided {
		// Interactive TUI wizard
		result, err := tui.RunWizard(env)
		if err != nil {
			return err
		}
		catalog = scaffold.BuildFilteredCatalog(result.Stack, result.Components)
	} else {
		// One-click mode — install everything for detected stack
		catalog = scaffold.BuildCatalog(env.Stack)
	}

	// Phase 3: Write files
	results := scaffold.WriteAll(targetDir, catalog)

	// Phase 4: Print summary
	printer.PrintResults(results)
	printer.PrintNextSteps(env.Stack)

	// Update plan checkboxes if running in-repo
	planPath := filepath.Join(targetDir, "docs", "plans")
	if _, err := os.Stat(planPath); err == nil {
		// Plan directory exists — this is the atv-installer repo itself
		printer.Info("Plan directory detected. Update plan checkboxes manually.")
	}

	return nil
}
