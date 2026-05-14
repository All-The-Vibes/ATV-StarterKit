package tui

import (
	"fmt"
	"strings"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/gstack"
)

// CategorySkill represents a single skill entry in the TUI, combining ATV and gstack sources.
type CategorySkill struct {
	Label       string // display label
	Key         string // selection key (layer key or gstack skill dir)
	Source      string // "atv" or "gstack"
	IsGstack    bool
	RequiresBun bool
}

// CategoryGroup holds the skills for one functional category in the TUI.
type CategoryGroup struct {
	Category    string
	Label       string
	Description string
	Skills      []CategorySkill
}

// ATVSkillsByCategory maps ATV layer keys to functional categories.
var atvCategoryMapping = map[string][]CategorySkill{
	gstack.CategoryPlanning: {
		{Label: "Brainstorming — explore what to build", Key: "core-skills:brainstorming", Source: "atv"},
		{Label: "CE Ideate — structured idea exploration", Key: "core-skills:ce-ideate", Source: "atv"},
		{Label: "Plan — turn ideas into an implementation plan", Key: "core-skills:ce-plan", Source: "atv"},
		{Label: "Deepen Plan — parallel research to harden the plan", Key: "core-skills:deepen-plan", Source: "atv"},
	},
	gstack.CategoryReview: {
		{Label: "CE Review — multi-agent code review", Key: "core-skills:ce-review", Source: "atv"},
	},
	gstack.CategorySecurity: {
		{Label: "ATV Security — agentic config audit + OWASP Top 10 + STRIDE source-code review", Key: "core-skills:atv-security", Source: "atv"},
	},
	gstack.CategoryShipping: {
		{Label: "Takeoff — backlog briefing at session start", Key: "core-skills:takeoff", Source: "atv"},
		{Label: "CE Work — execute plans with quality checks", Key: "core-skills:ce-work", Source: "atv"},
		{Label: "LFG — full autonomous pipeline", Key: "orchestrators:lfg", Source: "atv"},
		{Label: "SLFG — swarm mode parallel execution", Key: "orchestrators:slfg", Source: "atv"},
		{Label: "CE Compound — document solutions", Key: "core-skills:ce-compound", Source: "atv"},
		{Label: "CE Compound Refresh — refresh documented solutions", Key: "core-skills:ce-compound-refresh", Source: "atv"},
		{Label: "Land — commit, push, and open a PR at session end", Key: "core-skills:land", Source: "atv"},
	},
	gstack.CategoryQATesting: {
		{Label: "agent-browser — real browser automation with screenshots and form fills", Key: "agent-browser", Source: "atv"},
	},
	gstack.CategoryGuidelines: {
		{Label: "Karpathy Guidelines — think before coding, simplicity, surgical changes, goal-driven execution", Key: "core-skills:karpathy-guidelines", Source: "atv"},
		{Label: "Autoresearch — autonomous experiment loop: define goal + metric, agent iterates", Key: "core-skills:autoresearch", Source: "atv"},
	},
	gstack.CategoryMaintenance: {
		{Label: "ATV Doctor — diagnose installation health, version drift, missing prereqs", Key: "core-skills:atv-doctor", Source: "atv"},
		{Label: "ATV Update — update marketplace plugins; advisory for project scaffold", Key: "core-skills:atv-update", Source: "atv"},
	},
	gstack.CategoryEasterEgg: {
		{Label: "memeIQ — AI-powered meme generation toolkit", Key: "easter-eggs:meme-iq", Source: "atv"},
	},
}

// BuildCategoryGroups creates the full list of TUI category groups,
// mixing ATV and gstack skills organized by function.
func BuildCategoryGroups(prereqs gstack.Prerequisites) []CategoryGroup {
	gstackByCategory := gstack.SkillsByCategory()
	var groups []CategoryGroup

	for _, cat := range gstack.AllCategories() {
		group := CategoryGroup{
			Category:    cat,
			Label:       gstack.CategoryLabel(cat),
			Description: categoryDescription(cat, prereqs),
		}

		// Add ATV skills for this category first (ATV overrides)
		if atvSkills, ok := atvCategoryMapping[cat]; ok {
			group.Skills = append(group.Skills, atvSkills...)
		}

		// Add gstack skills for this category
		if gSkills, ok := gstackByCategory[cat]; ok {
			for _, gs := range gSkills {
				skill := CategorySkill{
					Label:       fmt.Sprintf("%s — %s", gs.Name, gs.Description),
					Key:         fmt.Sprintf("gstack:%s", gs.Dir),
					Source:      "gstack",
					IsGstack:    true,
					RequiresBun: gs.RequiresRuntime,
				}
				group.Skills = append(group.Skills, skill)
			}
		}

		if len(group.Skills) > 0 {
			groups = append(groups, group)
		}
	}

	return groups
}

func categoryDescription(cat string, prereqs gstack.Prerequisites) string {
	switch cat {
	case gstack.CategoryPlanning:
		return "Shape the work before coding: brainstorms, plans, architecture review, and design direction."
	case gstack.CategoryReview:
		return "Catch issues early with code, design, and independent review passes."
	case gstack.CategoryQATesting:
		if prereqs.HasBun {
			return "Browser and performance workflows. Bun is available, so runtime QA flows can be preselected and built."
		}
		return "Browser and performance workflows. Bun is missing, so runtime-heavy QA stays docs-only until Bun is installed."
	case gstack.CategorySecurity:
		return "Threat modeling and security review guardrails for higher-risk work."
	case gstack.CategoryShipping:
		return "PR, release, deploy, and post-deploy follow-through once the work is ready to ship."
	case gstack.CategorySafety:
		return "Reduce destructive mistakes during risky debugging or production work."
	case gstack.CategoryDebugging:
		return "Systematic investigation patterns that push the agent toward root cause before fixes."
	case gstack.CategoryRetrospective:
		return "Capture learnings after shipping so the repo compounds team knowledge over time."
	case gstack.CategoryGuidelines:
		return "Behavioral guidelines that shape how Copilot approaches work: assumptions, simplicity, change scope, and verification."
	case gstack.CategoryMaintenance:
		return "Keep your ATV install healthy and up to date — diagnostic and update workflows for both project-scaffold and marketplace install paths."
	case gstack.CategoryEasterEgg:
		return "Hidden gems and fun extras. Because every good toolkit deserves a few surprises."
	default:
		return "Additional workflow capabilities for this repo."
	}
}

func infraSelectionDescription() string {
	return "Repo files and configuration scaffolding. Everything starts selected; deselect only the artifacts you know you do not want written into the repo."
}

// InfraLayers returns the non-skill infrastructure layer options (unchanged from original).
var InfraLayers = []CategorySkill{
	{Label: "MCP server config — GitHub, Azure, Terraform, Context7 connectors", Key: LayerMCPServers, Source: "atv"},
	{Label: "VS Code recommendations — .vscode/extensions.json", Key: LayerVSCodeExtensions, Source: "atv"},
	{Label: "Copilot system instructions — .github/copilot-instructions.md", Key: LayerCopilotInstructions, Source: "atv"},
	{Label: "Copilot setup steps — .github/copilot-setup-steps.yml", Key: LayerSetupSteps, Source: "atv"},
	{Label: "File-scoped instructions — applyTo rules for stack/tool guidance", Key: LayerFileInstructions, Source: "atv"},
	{Label: "Documentation folders — docs/plans, brainstorms, and solutions", Key: LayerDocsStructure, Source: "atv"},
	{Label: "Universal reviewer agents — security, performance, architecture", Key: LayerUniversalAgents, Source: "atv"},
	{Label: "Stack reviewers — language/framework specific agents", Key: LayerStackAgents, Source: "atv"},
}

// ParseSelections splits the selected keys into ATV layers and gstack skill dirs.
func ParseSelections(selected []string) (atvLayers []string, gstackDirs []string) {
	for _, key := range selected {
		if strings.HasPrefix(key, "gstack:") {
			gstackDirs = append(gstackDirs, strings.TrimPrefix(key, "gstack:"))
		} else if strings.Contains(key, ":") {
			// ATV category:skill format — extract the layer
			parts := strings.SplitN(key, ":", 2)
			atvLayers = append(atvLayers, parts[0])
		} else {
			// Plain layer key (infra layers)
			atvLayers = append(atvLayers, key)
		}
	}

	// Deduplicate ATV layers
	seen := make(map[string]bool)
	var deduped []string
	for _, l := range atvLayers {
		if !seen[l] {
			seen[l] = true
			deduped = append(deduped, l)
		}
	}
	return deduped, gstackDirs
}
