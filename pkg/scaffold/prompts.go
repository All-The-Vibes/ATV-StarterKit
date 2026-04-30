package scaffold

//go:generate go run ../../cmd/promptgen

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"
)

// promptShimSkillDirectories enumerates skills that should be discoverable as
// VS Code Copilot Chat slash commands via .github/prompts/<name>.prompt.md
// shims. Sub-skills, behavioral guides, and helper-only skills are excluded —
// see nonUserFacingSkills for the corresponding deny-list. The two together
// must cover every entry in coreSkillDirectories ∪ orchestratorSkillDirectories
// ∪ easterEggSkillDirectories (enforced by parity tests).
var promptShimSkillDirectories = []string{
	// Compound Engineering core workflow
	"ce-brainstorm",
	"ce-plan",
	"ce-work",
	"ce-review",
	"ce-compound",
	"ce-ideate",
	// Session lifecycle
	"takeoff",
	"land",
	// ATV Learning Pipeline
	"learn",
	"instincts",
	"evolve",
	"observe",
	// ATV Quality
	"unslop",
	// Experimentation
	"autoresearch",
	// Maintenance
	"atv-doctor",
	"atv-update",
	// Security
	"atv-security",
	// Orchestrators
	"lfg",
}

// nonUserFacingSkills lists shipped skills that are intentionally NOT exposed
// as VS Code slash commands. Each entry must include a justification comment
// so the next maintainer understands the choice. This list plus
// promptShimSkillDirectories must cover every shipped skill (parity-tested).
var nonUserFacingSkills = []string{
	// Sub-skill invoked by ce-brainstorm; not a top-level command.
	"brainstorming",
	// Sub-skill invoked by ce-compound; not a top-level command.
	"ce-compound-refresh",
	// Sub-skill invoked by ce-plan; not a top-level command.
	"deepen-plan",
	// Sub-skill invoked by ce-review and the document-review pipeline.
	"document-review",
	// Bootstrap helper invoked during install, not by users at runtime.
	"setup",
	// Behavioral guideline reference text, not an invocable command.
	"karpathy-guidelines",
	// Orchestrator sub-skills — internal building blocks of /lfg.
	"feature-video",
	"ralph-loop",
	"resolve_todo_parallel",
	"slfg",
	"test-browser",
	// Easter-egg skill — discoverable via the easter-eggs layer, not the chat picker.
	"meme-iq",
}

// promptShimTemplate is the canonical body of a generated .prompt.md shim.
// VS Code Copilot Chat reads `mode` and `description` from the YAML
// front-matter and surfaces `description` in the slash-command picker.
// The body is a thin delegation to the SKILL.md so SKILL.md remains the
// single source of truth.
const promptShimTemplate = `---
mode: agent
description: Run the {{.Name}} skill
---

Invoke the ` + "`{{.Name}}`" + ` skill defined at ` + "`.github/skills/{{.Name}}/SKILL.md`" + ` and follow its instructions.

Forward any arguments or context the user provided in this message verbatim to the skill.
`

// BuildPromptShim renders the .prompt.md shim body for a single skill name.
// The output is byte-for-byte deterministic so it can be parity-checked
// against dogfooded copies under .github/prompts/.
func BuildPromptShim(skillName string) []byte {
	tmpl, err := template.New("prompt-shim").Parse(promptShimTemplate)
	if err != nil {
		panic(fmt.Sprintf("prompt shim template invalid: %v", err))
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct{ Name string }{Name: skillName}); err != nil {
		panic(fmt.Sprintf("prompt shim render failed for %q: %v", skillName, err))
	}
	return buf.Bytes()
}

// PromptShimPlan returns the ordered list of skill names that receive prompt
// shims. Exposed for the cmd/promptgen tool which regenerates the dogfooded
// .github/prompts/ files from BuildPromptShim.
func PromptShimPlan() []string {
	out := make([]string, len(promptShimSkillDirectories))
	copy(out, promptShimSkillDirectories)
	return out
}

// promptShims returns components for every allow-listed skill (full install).
func promptShims() []Component {
	return promptShimComponents(nil)
}

// selectedPromptShimsForLayers returns the set of allow-listed skill names
// whose layer is enabled. A nil/empty result means no prompt shims should be
// emitted (e.g., user deselected both core-skills and orchestrators).
func selectedPromptShimsForLayers(layerSet map[string]bool) map[string]bool {
	if len(layerSet) == 0 {
		return nil
	}
	enabledSkills := make(map[string]bool)
	if layerSet["core-skills"] {
		for _, n := range coreSkillDirectories {
			enabledSkills[n] = true
		}
	}
	if layerSet["orchestrators"] {
		for _, n := range orchestratorSkillDirectories {
			enabledSkills[n] = true
		}
	}
	if layerSet["easter-eggs"] {
		for _, n := range easterEggSkillDirectories {
			enabledSkills[n] = true
		}
	}
	if len(enabledSkills) == 0 {
		return nil
	}
	out := make(map[string]bool)
	for _, name := range promptShimSkillDirectories {
		if enabledSkills[name] {
			out[name] = true
		}
	}
	return out
}

// promptShimComponents returns shim components for the subset of allow-listed
// skills present in `selected`. A nil `selected` means "include all".
func promptShimComponents(selected map[string]bool) []Component {
	var comps []Component
	for _, name := range promptShimSkillDirectories {
		if selected != nil && !selected[name] {
			continue
		}
		dest := filepath.Join(".github", "prompts", name+".prompt.md")
		comps = append(comps, Component{
			Path:     dest,
			Content:  BuildPromptShim(name),
			HookType: HookPromptShims,
		})
	}
	return comps
}
