package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/detect"
	"github.com/All-The-Vibes/ATV-StarterKit/pkg/installstate"
)

//go:embed all:templates
var templateFS embed.FS

// Component represents a single file or directory to scaffold.
type Component struct {
	Path      string // relative path in target directory
	Content   []byte // file content (empty for dirs)
	IsDir     bool
	MergeJSON bool // if true, merge with existing JSON instead of skipping
	HookType  int  // 1-6 matching Copilot lifecycle hooks
}

// BuildCatalog returns the full list of components for the given stack.
func BuildCatalog(stack detect.Stack) []Component {
	var catalog []Component

	// Directories first
	catalog = append(catalog, directories()...)
	catalog = append(catalog, documentationDirectories()...)

	// Hook 1: System Instructions
	catalog = append(catalog, systemInstructions(stack)...)

	// Hook 2: Setup Steps
	catalog = append(catalog, setupSteps(stack)...)

	// Hook 3: MCP Config
	catalog = append(catalog, mcpConfig()...)

	// Hook 4: Skills (from .github/skills/ in this repo)
	catalog = append(catalog, skills()...)

	// Hook 5: Agents (from .github/agents/ in this repo)
	catalog = append(catalog, agents(stack)...)

	// Hook 6: File Instructions
	catalog = append(catalog, fileInstructions(stack)...)

	// Observer hooks (copilot-hooks.json + scripts)
	catalog = append(catalog, observerHooks()...)

	// VS Code config
	catalog = append(catalog, vscodeConfig()...)

	return catalog
}

// BuildFilteredCatalog returns components filtered by the user's layer selections.
func BuildFilteredCatalog(stack detect.Stack, layers []string) []Component {
	return BuildFilteredCatalogForPacks([]installstate.StackPack{detect.StackPackForStack(stack)}, stack, layers)
}

// BuildFilteredCatalogForPacks returns components filtered by the user's layer selections
// and additive stack-pack selections. Primary stack controls singular root templates;
// selected packs control stack-specific file instructions and agents.
func BuildFilteredCatalogForPacks(packs []installstate.StackPack, primaryStack detect.Stack, layers []string) []Component {
	normalizedPacks, err := installstate.NormalizeStackPacks(packs)
	if err != nil || len(normalizedPacks) == 0 {
		normalizedPacks = []installstate.StackPack{installstate.StackPackGeneral}
	}
	primaryStack = detect.PrimaryStackForPacks(normalizedPacks, primaryStack)

	layerSet := make(map[string]bool)
	for _, l := range layers {
		layerSet[l] = true
	}

	var catalog []Component

	// Directories always included
	catalog = append(catalog, directories()...)
	if layerSet["docs-structure"] {
		catalog = append(catalog, documentationDirectories()...)
	}

	if layerSet["copilot-instructions"] {
		catalog = append(catalog, systemInstructions(primaryStack)...)
	}
	if layerSet["setup-steps"] {
		catalog = append(catalog, setupSteps(primaryStack)...)
	}
	if layerSet["mcp-servers"] {
		catalog = append(catalog, mcpConfig()...)
	}
	selectedSkillDirs := make(map[string]bool)
	if layerSet["core-skills"] {
		for _, dir := range coreSkillDirectories {
			selectedSkillDirs[dir] = true
		}
	}
	if layerSet["orchestrators"] {
		for _, dir := range orchestratorSkillDirectories {
			selectedSkillDirs[dir] = true
		}
	}
	if layerSet["easter-eggs"] {
		for _, dir := range easterEggSkillDirectories {
			selectedSkillDirs[dir] = true
		}
	}
	if len(selectedSkillDirs) > 0 {
		catalog = append(catalog, skillComponents(selectedSkillDirs)...)
	}
	if layerSet["universal-agents"] || layerSet["stack-agents"] {
		catalog = append(catalog, agentsForPacks(normalizedPacks, layerSet["universal-agents"], layerSet["stack-agents"])...)
	}
	if layerSet["easter-eggs"] {
		catalog = append(catalog, easterEggAgents()...)
	}
	if layerSet["file-instructions"] {
		catalog = append(catalog, fileInstructionsForPacks(normalizedPacks)...)
	}
	if layerSet["vscode-extensions"] {
		catalog = append(catalog, vscodeConfig()...)
	}
	// Observer hooks are included when core-skills are selected (learning pipeline)
	if layerSet["core-skills"] {
		catalog = append(catalog, observerHooks()...)
	}

	return catalog
}

func directories() []Component {
	dirs := []string{
		".github/skills",
		".github/agents",
		".github/hooks/scripts",
		".vscode",
	}
	var comps []Component
	for _, d := range dirs {
		comps = append(comps, Component{Path: d, IsDir: true})
	}
	return comps
}

func documentationDirectories() []Component {
	dirs := []string{
		"docs/plans",
		"docs/brainstorms",
		"docs/solutions",
		".atv/instincts",
	}
	var comps []Component
	for _, d := range dirs {
		comps = append(comps, Component{Path: d, IsDir: true})
	}
	return comps
}

var coreSkillDirectories = []string{
	"brainstorming",
	"ce-brainstorm",
	"ce-compound",
	"ce-compound-refresh",
	"ce-ideate",
	"ce-plan",
	"ce-review",
	"ce-work",
	"deepen-plan",
	"document-review",
	"setup",
	// ATV Learning Pipeline
	"learn",
	"instincts",
	"evolve",
	"observe",
	// ATV Quality
	"unslop",
	// Behavioral Guidelines
	"karpathy-guidelines",
}

var orchestratorSkillDirectories = []string{
	"claude-permissions-optimizer",
	"feature-video",
	"lfg",
	"ralph-loop",
	"resolve_todo_parallel",
	"slfg",
	"test-browser",
}

var easterEggSkillDirectories = []string{
	"meme-iq",
}

var easterEggAgentFiles = map[string]bool{
	"meme-iq.agent.md": true,
}

func systemInstructions(stack detect.Stack) []Component {
	filename := fmt.Sprintf("templates/instructions/%s.md", stackName(stack))
	content := mustRead(filename)
	return []Component{
		{Path: ".github/copilot-instructions.md", Content: content, HookType: 1},
	}
}

func setupSteps(stack detect.Stack) []Component {
	filename := fmt.Sprintf("templates/setup-steps/%s.yml", stackName(stack))
	content := mustRead(filename)
	return []Component{
		{Path: ".github/copilot-setup-steps.yml", Content: content, HookType: 2},
	}
}

func mcpConfig() []Component {
	content := mustRead("templates/configs/copilot-mcp-config.json")
	return []Component{
		{Path: ".github/copilot-mcp-config.json", Content: content, HookType: 3, MergeJSON: true},
	}
}

func skills() []Component {
	return skillComponents(nil)
}

func skillComponents(selected map[string]bool) []Component {
	var comps []Component

	if err := fs.WalkDir(templateFS, "templates/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path == "templates/skills" {
			return nil
		}
		content, _ := templateFS.ReadFile(path)

		relPath := strings.TrimPrefix(path, "templates/skills/")
		if selected != nil {
			skillDir := strings.SplitN(relPath, "/", 2)[0]
			if !selected[skillDir] {
				return nil
			}
		}
		destPath := filepath.Join(".github", "skills", relPath)
		comps = append(comps, Component{Path: destPath, Content: content, HookType: 4})
		return nil
	}); err != nil {
		panic(fmt.Sprintf("failed to walk embedded skills templates: %v", err))
	}

	return comps
}

func agents(stack detect.Stack) []Component {
	return agentsForPacks([]installstate.StackPack{detect.StackPackForStack(stack)}, true, true)
}

func agentsForPacks(packs []installstate.StackPack, includeUniversal bool, includeStackSpecific bool) []Component {
	selectedStacks := make(map[detect.Stack]bool)
	for _, pack := range packs {
		stack := detect.StackForPack(pack)
		if stack != detect.StackGeneral {
			selectedStacks[stack] = true
		}
	}

	var comps []Component

	if err := fs.WalkDir(templateFS, "templates/agents", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path == "templates/agents" {
			return nil
		}
		content, _ := templateFS.ReadFile(path)
		relPath := strings.TrimPrefix(path, "templates/agents/")
		destPath := filepath.Join(".github", "agents", relPath)

		// Skip easter-egg agents; they're included separately when easter-eggs layer is selected
		if easterEggAgentFiles[filepath.Base(relPath)] {
			return nil
		}

		if isStackSpecific(relPath) {
			if !includeStackSpecific {
				return nil
			}
			agentStack := stackAgents[filepath.Base(relPath)]
			if !selectedStacks[agentStack] {
				return nil
			}
		} else if !includeUniversal {
			return nil
		}

		comps = append(comps, Component{Path: destPath, Content: content, HookType: 5})
		return nil
	}); err != nil {
		panic(fmt.Sprintf("failed to walk embedded agents templates: %v", err))
	}

	return comps
}

func fileInstructions(stack detect.Stack) []Component {
	return fileInstructionsForPacks([]installstate.StackPack{detect.StackPackForStack(stack)})
}

func fileInstructionsForPacks(packs []installstate.StackPack) []Component {
	var comps []Component
	for _, pack := range packs {
		stack := detect.StackForPack(pack)
		sn := stackName(stack)
		if sn == "general" {
			continue // no file instructions for general stack
		}
		filename := fmt.Sprintf("templates/file-instructions/%s.instructions.md", sn)
		content := mustRead(filename)
		destName := fmt.Sprintf(".github/%s.instructions.md", sn)
		comps = append(comps, Component{Path: destName, Content: content, HookType: 6})
	}
	return comps
}

func vscodeConfig() []Component {
	content := mustRead("templates/configs/extensions.json")
	return []Component{
		{Path: ".vscode/extensions.json", Content: content, MergeJSON: true},
	}
}

// observerHooks returns the Copilot CLI hook configuration and observer scripts
// for the ATV continuous learning pipeline.
func observerHooks() []Component {
	hookConfig := mustRead("templates/hooks/copilot-hooks.json")
	observeScript := mustRead("templates/hooks/scripts/observe.js")

	return []Component{
		{Path: ".github/hooks/copilot-hooks.json", Content: hookConfig, MergeJSON: true},
		{Path: ".github/hooks/scripts/observe.js", Content: observeScript},
	}
}

func mustRead(path string) []byte {
	data, err := templateFS.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("embedded template not found: %s", path))
	}
	return data
}

func stackName(stack detect.Stack) string {
	switch stack {
	case detect.StackRails:
		return "rails"
	case detect.StackPython:
		return "python"
	case detect.StackTypeScript:
		return "typescript"
	default:
		return "general"
	}
}

// Stack-specific agent mapping
var stackAgents = map[string]detect.Stack{
	"kieran-rails-reviewer.agent.md":         detect.StackRails,
	"dhh-rails-reviewer.agent.md":            detect.StackRails,
	"julik-frontend-races-reviewer.agent.md": detect.StackRails,
	"data-integrity-guardian.agent.md":       detect.StackRails,
	"schema-drift-detector.agent.md":         detect.StackRails,
	"data-migration-expert.agent.md":         detect.StackRails,
	"deployment-verification-agent.agent.md": detect.StackRails,
	"lint.agent.md":                          detect.StackRails,
	"kieran-python-reviewer.agent.md":        detect.StackPython,
	"kieran-typescript-reviewer.agent.md":    detect.StackTypeScript,
}

func isStackSpecific(filename string) bool {
	_, ok := stackAgents[filepath.Base(filename)]
	return ok
}

// easterEggAgents returns agent components for easter-egg features.
func easterEggAgents() []Component {
	var comps []Component
	for filename := range easterEggAgentFiles {
		path := "templates/agents/" + filename
		content, err := templateFS.ReadFile(path)
		if err != nil {
			continue
		}
		destPath := filepath.Join(".github", "agents", filename)
		comps = append(comps, Component{Path: destPath, Content: content, HookType: 5})
	}
	return comps
}
