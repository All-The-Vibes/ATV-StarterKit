package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/detect"
	"github.com/All-The-Vibes/ATV-StarterKit/pkg/output"
	"github.com/All-The-Vibes/ATV-StarterKit/pkg/scaffold"
	"github.com/All-The-Vibes/ATV-StarterKit/pkg/tui"
	"github.com/spf13/cobra"
)

const installerModulePath = "module github.com/All-The-Vibes/ATV-StarterKit"

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

	// Update plan checkboxes only when running inside the installer repository.
	if isInstallerRepo(targetDir) {
		printer.Info("Plan directory detected. Update plan checkboxes manually.")
	}

	return nil
}

func isInstallerRepo(dir string) bool {
	goModPath := filepath.Join(dir, "go.mod")
	goMod, err := os.ReadFile(goModPath)
	if err != nil {
		return false
	}

	templatesPath := filepath.Join(dir, "pkg", "scaffold", "templates")
	if info, err := os.Stat(templatesPath); err != nil || !info.IsDir() {
		return false
	}

	return bytes.Contains(goMod, []byte(installerModulePath))
}
