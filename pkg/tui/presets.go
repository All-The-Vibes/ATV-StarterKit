package tui

import (
	"fmt"
	"strings"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/gstack"
)

// Preset defines a curated skill selection level.
type Preset struct {
	Key                 string
	Name                string
	Emoji               string
	Description         string
	Detail              string
	InstallFootprint    string
	CapabilitySummary   string
	PrerequisiteSummary string
	DowngradeSummary    string
	ATVLayers           []string
	GstackDirs          []string
	EnableGstackRuntime bool
	IncludeAgentBrowser bool
	NeedsBun            bool
}

// StarterPreset includes core ATV skills only — no network calls, instant install.
var StarterPreset = Preset{
	Key:                 "starter",
	Name:                "Starter",
	Emoji:               "⚡",
	Description:         "repo-only scaffold",
	Detail:              "Core ATV planning, execution, review, and docs flow. No gstack sync or browser tooling.",
	InstallFootprint:    "lightest install",
	CapabilitySummary:   "ATV core skills, orchestrators, agents, MCP config, and docs structure",
	PrerequisiteSummary: "No extra runtime tools required beyond writing files into the repo",
	DowngradeSummary:    "No downgrade path needed; this preset is already local-only",
	ATVLayers: []string{
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
	},
	EnableGstackRuntime: false,
	GstackDirs:          nil,
	IncludeAgentBrowser: false,
}

// ProPreset adds gstack text-only sprint skills — no browser QA.
var ProPreset = Preset{
	Key:                 "pro",
	Name:                "Pro",
	Emoji:               "🚀",
	Description:         "text-first gstack upgrade",
	Detail:              "Adds markdown-first gstack planning, review, shipping, safety, debugging, and retro skills without browser QA.",
	InstallFootprint:    "adds gstack repo + skill sync",
	CapabilitySummary:   "Everything in Starter plus richer gstack workflows for planning, review, shipping, safety, and retros",
	PrerequisiteSummary: "Git is needed to clone gstack; Bun improves doc generation, but runtime build stays off by default",
	DowngradeSummary:    "If Bun is missing, skills still sync as markdown docs and browser/runtime workflows remain off",
	ATVLayers: []string{
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
		// Pro adds the workflow-helper layers so users get git/PR helpers,
		// onboarding/skill-creator/reproduce-bug, and the todo-CLI suite.
		LayerDevTools,
	},
	GstackDirs: []string{
		"office-hours", "plan-ceo-review", "plan-eng-review", "plan-design-review",
		"design-consultation", "autoplan",
		"review", "design-review", "design-shotgun", "codex",
		"cso",
		"ship", "land-and-deploy", "canary", "document-release",
		"careful", "freeze", "guard", "unfreeze",
		"investigate",
		"retro",
		"learn",
	},
	EnableGstackRuntime: false,
	IncludeAgentBrowser: false,
}

// FullPreset installs everything — all gstack skills including browser QA.
var FullPreset = Preset{
	Key:                 "full",
	Name:                "Full",
	Emoji:               "🔥",
	Description:         "browser QA + runtime setup",
	Detail:              "Adds all gstack skills plus agent-browser and Chrome-friendly QA helpers.",
	InstallFootprint:    "largest install",
	CapabilitySummary:   "Everything in Pro plus runtime-enabled gstack QA/browser workflows and agent-browser setup",
	PrerequisiteSummary: "Git clones gstack, Bun enables runtime gstack workflows, and npm/Node helps install agent-browser + Chrome",
	DowngradeSummary:    "Without Bun, gstack falls back to markdown-only sync; agent-browser still attempts install separately",
	ATVLayers: []string{
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
		// Full pulls in every shippable skill layer.
		LayerDevTools,
		LayerStyleSkills,
		LayerMediaSkills,
	},
	EnableGstackRuntime: true,
	GstackDirs:          allGstackDirs(),
	IncludeAgentBrowser: true,
	NeedsBun:            true,
}

// ShouldEnableGstackRuntime returns whether the guided flow should enable gstack runtime work.
func (p Preset) ShouldEnableGstackRuntime(selectedDirs []string, prereqs gstack.Prerequisites) bool {
	return p.EnableGstackRuntime && prereqs.HasBun && len(selectedDirs) > 0
}

// PreviewLabel returns a one-line summary used in the preset picker.
func (p Preset) PreviewLabel(prereqs gstack.Prerequisites) string {
	parts := []string{p.InstallFootprint, p.Description, p.prerequisiteBadge(prereqs)}
	return fmt.Sprintf("%s %s — %s", p.Emoji, p.Name, strings.Join(parts, " · "))
}

// SelectionSummary returns a richer explanation used after a preset is chosen.
func (p Preset) SelectionSummary(prereqs gstack.Prerequisites) string {
	lines := []string{
		fmt.Sprintf("%s %s", p.Emoji, p.Name),
		fmt.Sprintf("What it adds: %s.", p.CapabilitySummary),
		fmt.Sprintf("Install footprint: %s.", p.InstallFootprint),
		fmt.Sprintf("Prerequisites: %s.", p.PrerequisiteSummary),
		fmt.Sprintf("If tools are missing: %s.", p.DowngradeSummary),
		fmt.Sprintf("Detected now: %s.", p.prerequisiteBadge(prereqs)),
	}
	return strings.Join(lines, "\n")
}

func presetSelectionDescription(prereqs gstack.Prerequisites) string {
	lines := []string{
		"Presets are still the happy path. Pick one unless you already know you want to add or remove individual capabilities.",
		fmt.Sprintf("Detected prerequisites: %s.", prereqs.Summary()),
		"Starter stays repo-local, Pro adds text-first gstack workflows, and Full adds browser QA plus runtime setup when tools are available.",
	}
	return strings.Join(lines, "\n")
}

func (p Preset) prerequisiteBadge(prereqs gstack.Prerequisites) string {
	switch p.Key {
	case StarterPreset.Key:
		return "no runtime prerequisites"
	case ProPreset.Key:
		if !prereqs.HasGit {
			return "git missing → gstack sync blocked"
		}
		if prereqs.HasBun {
			return "git ready · Bun available"
		}
		return "git ready · Bun optional"
	case FullPreset.Key:
		if !prereqs.HasGit {
			return "git missing → gstack sync blocked"
		}
		if prereqs.HasBun {
			return "git + Bun ready"
		}
		return "Bun missing → docs-only gstack fallback"
	default:
		return prereqs.Summary()
	}
}

// AllPresets returns the ordered list of presets.
func AllPresets() []Preset {
	return []Preset{StarterPreset, ProPreset, FullPreset}
}

func allGstackDirs() []string {
	var dirs []string
	for _, s := range gstack.AllSkills() {
		dirs = append(dirs, s.Dir)
	}
	return dirs
}
