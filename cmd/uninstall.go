package cmd

import (
	"fmt"
	"os"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/installstate"
	"github.com/All-The-Vibes/ATV-StarterKit/pkg/output"
	"github.com/All-The-Vibes/ATV-StarterKit/pkg/scaffold"
	"github.com/spf13/cobra"
)

var forceUninstall bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove all ATV-installed files from current directory",
	Long: `Remove all ATV Starter Kit files scaffolded by 'atv-installer init'.

Files that you've modified since installation are preserved by default.
Use --force to remove everything regardless of modifications.

Removes:
  .github/skills/       ATV and gstack skill directories
  .github/agents/       Agent definition files
  .github/hooks/        Copilot hooks and observer scripts
  .github/copilot-*     System instructions, setup steps, MCP config
  .github/*.instructions.md  File-scoped instructions
  .gstack/              gstack staging directory
  .atv/                 Install manifest and instincts
  docs/plans|brainstorms|solutions/  (only if empty)`,
	RunE: runUninstall,
}

func init() {
	uninstallCmd.Flags().BoolVar(&forceUninstall, "force", false, "Remove all files even if modified since installation")
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	printer := output.NewPrinter()
	targetDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Try to load install manifest for checksums
	var checksums map[string]string
	manifest, err := installstate.ReadManifest(targetDir)
	if err == nil && manifest.FileChecksums != nil {
		checksums = manifest.FileChecksums
		printer.Info("Found install manifest — will preserve user-modified files")
	} else {
		printer.Info("No install manifest found — removing all known ATV files")
	}

	if forceUninstall {
		printer.Info("Force mode — removing everything regardless of modifications")
		checksums = nil
	}

	result := scaffold.Uninstall(targetDir, checksums, forceUninstall)

	// Print results
	fmt.Println()
	if len(result.Removed) > 0 {
		printer.Info(fmt.Sprintf("Removed %d items:", len(result.Removed)))
		for _, path := range result.Removed {
			fmt.Printf("  - %s\n", path)
		}
	}
	if len(result.Skipped) > 0 {
		printer.Info(fmt.Sprintf("Preserved %d user-modified items:", len(result.Skipped)))
		for _, path := range result.Skipped {
			fmt.Printf("  - %s\n", path)
		}
	}
	if len(result.Errors) > 0 {
		printer.Info(fmt.Sprintf("Failed to remove %d items:", len(result.Errors)))
		for _, errStr := range result.Errors {
			fmt.Printf("  - %s\n", errStr)
		}
	}

	fmt.Println()
	printer.Info(fmt.Sprintf("Uninstall complete: %s", result.Summary()))

	if len(result.Skipped) > 0 {
		fmt.Println()
		printer.Info("Tip: use --force to remove everything including modified files")
	}

	return nil
}
