package tui

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/github/atv-installer/pkg/detect"
)

// WizardResult holds the user's selections from the guided wizard.
type WizardResult struct {
	Stack      detect.Stack
	Components []string
}

// component layer keys
const (
	LayerCoreSkills        = "core-skills"
	LayerOrchestrators     = "orchestrators"
	LayerUniversalAgents   = "universal-agents"
	LayerStackAgents       = "stack-agents"
	LayerMCPServers        = "mcp-servers"
	LayerVSCodeExtensions  = "vscode-extensions"
	LayerCopilotInstructions = "copilot-instructions"
	LayerSetupSteps        = "setup-steps"
	LayerFileInstructions  = "file-instructions"
	LayerDocsStructure     = "docs-structure"
	LayerLocalConfig       = "local-config"
)

// AllLayers returns all available component layer keys.
func AllLayers() []string {
	return []string{
		LayerCoreSkills,
		LayerOrchestrators,
		LayerUniversalAgents,
		LayerStackAgents,
		LayerMCPServers,
		LayerVSCodeExtensions,
		LayerCopilotInstructions,
		LayerSetupSteps,
		LayerFileInstructions,
		LayerDocsStructure,
	}
}

// RunWizard runs the interactive guided mode TUI.
func RunWizard(detected detect.Environment) (*WizardResult, error) {
	result := &WizardResult{
		Stack: detected.Stack,
	}

	// Step 1: Confirm or override detected stack
	stackOptions := []huh.Option[string]{
		huh.NewOption("TypeScript", "typescript"),
		huh.NewOption("Python", "python"),
		huh.NewOption("Rails", "rails"),
		huh.NewOption("General", "general"),
	}

	var selectedStack string
	switch detected.Stack {
	case detect.StackTypeScript:
		selectedStack = "typescript"
	case detect.StackPython:
		selectedStack = "python"
	case detect.StackRails:
		selectedStack = "rails"
	default:
		selectedStack = "general"
	}

	// Step 2: Component selection
	selectedComponents := AllLayers() // all selected by default

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What's your primary stack?").
				Description(fmt.Sprintf("Auto-detected: %s (%s)", detected.Stack, detected.StackHint)).
				Options(stackOptions...).
				Value(&selectedStack),
		),
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Which components do you want?").
				Description("All are selected by default. Deselect any you don't need.").
				Options(
					huh.NewOption("Core pipeline skills (brainstorm, plan, work, review, compound)", LayerCoreSkills).Selected(true),
					huh.NewOption("Orchestrators (lfg, slfg)", LayerOrchestrators).Selected(true),
					huh.NewOption("Universal agents (security, performance, architecture, ...)", LayerUniversalAgents).Selected(true),
					huh.NewOption("Stack-specific agents (language reviewers)", LayerStackAgents).Selected(true),
					huh.NewOption("MCP servers (GitHub, Azure, Terraform, Context7)", LayerMCPServers).Selected(true),
					huh.NewOption("VS Code extensions.json", LayerVSCodeExtensions).Selected(true),
					huh.NewOption("Copilot instructions (.github/copilot-instructions.md)", LayerCopilotInstructions).Selected(true),
					huh.NewOption("Copilot setup steps (.github/copilot-setup-steps.yml)", LayerSetupSteps).Selected(true),
					huh.NewOption("File-scoped instructions (applyTo globs)", LayerFileInstructions).Selected(true),
					huh.NewOption("docs/ structure (plans, brainstorms, solutions)", LayerDocsStructure).Selected(true),
					huh.NewOption("Compound engineering local config", LayerLocalConfig),
				).
				Value(&selectedComponents),
		),
	).WithTheme(huh.ThemeCatppuccin())

	err := form.Run()
	if err != nil {
		return nil, fmt.Errorf("wizard cancelled: %w", err)
	}

	// Map selected stack string back to Stack type
	switch selectedStack {
	case "typescript":
		result.Stack = detect.StackTypeScript
	case "python":
		result.Stack = detect.StackPython
	case "rails":
		result.Stack = detect.StackRails
	default:
		result.Stack = detect.StackGeneral
	}

	result.Components = selectedComponents
	return result, nil
}
