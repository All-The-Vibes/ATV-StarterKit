package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/github/atv-installer/pkg/detect"
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

	// VS Code config
	catalog = append(catalog, vscodeConfig()...)

	return catalog
}

// BuildFilteredCatalog returns components filtered by the user's layer selections.
func BuildFilteredCatalog(stack detect.Stack, layers []string) []Component {
	layerSet := make(map[string]bool)
	for _, l := range layers {
		layerSet[l] = true
	}

	var catalog []Component

	// Directories always included
	catalog = append(catalog, directories()...)

	if layerSet["copilot-instructions"] {
		catalog = append(catalog, systemInstructions(stack)...)
	}
	if layerSet["setup-steps"] {
		catalog = append(catalog, setupSteps(stack)...)
	}
	if layerSet["mcp-servers"] {
		catalog = append(catalog, mcpConfig()...)
	}
	if layerSet["core-skills"] || layerSet["orchestrators"] {
		catalog = append(catalog, skills()...)
	}
	if layerSet["universal-agents"] || layerSet["stack-agents"] {
		catalog = append(catalog, agents(stack)...)
	}
	if layerSet["file-instructions"] {
		catalog = append(catalog, fileInstructions(stack)...)
	}
	if layerSet["vscode-extensions"] {
		catalog = append(catalog, vscodeConfig()...)
	}

	return catalog
}

func directories() []Component {
	dirs := []string{
		".github/skills",
		".github/agents",
		".vscode",
		"docs/plans",
		"docs/brainstorms",
		"docs/solutions",
	}
	var comps []Component
	for _, d := range dirs {
		comps = append(comps, Component{Path: d, IsDir: true})
	}
	return comps
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
	// Walk the templates/skills directory for embedded skill content
	var comps []Component
	fs.WalkDir(templateFS, "templates/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path == "templates/skills" {
			return nil
		}
		content, _ := templateFS.ReadFile(path)
		// Convert templates/skills/ce-brainstorm/SKILL.md → .github/skills/ce-brainstorm/SKILL.md
		relPath := strings.TrimPrefix(path, "templates/skills/")
		destPath := filepath.Join(".github", "skills", relPath)
		comps = append(comps, Component{Path: destPath, Content: content, HookType: 4})
		return nil
	})
	return comps
}

func agents(stack detect.Stack) []Component {
	// Walk templates/agents for embedded agent content
	var comps []Component
	fs.WalkDir(templateFS, "templates/agents", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path == "templates/agents" {
			return nil
		}
		content, _ := templateFS.ReadFile(path)
		relPath := strings.TrimPrefix(path, "templates/agents/")
		destPath := filepath.Join(".github", "agents", relPath)

		// Check if this is a stack-specific agent
		if isStackSpecific(relPath) && !isForStack(relPath, stack) {
			return nil // skip agents not for this stack
		}

		comps = append(comps, Component{Path: destPath, Content: content, HookType: 5})
		return nil
	})
	return comps
}

func fileInstructions(stack detect.Stack) []Component {
	sn := stackName(stack)
	if sn == "general" {
		return nil // no file instructions for general stack
	}
	filename := fmt.Sprintf("templates/file-instructions/%s.instructions.md", sn)
	content := mustRead(filename)
	destName := fmt.Sprintf(".github/%s.instructions.md", sn)
	return []Component{
		{Path: destName, Content: content, HookType: 6},
	}
}

func vscodeConfig() []Component {
	content := mustRead("templates/configs/extensions.json")
	return []Component{
		{Path: ".vscode/extensions.json", Content: content, MergeJSON: true},
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
	"kieran-rails-reviewer.agent.md":              detect.StackRails,
	"dhh-rails-reviewer.agent.md":                 detect.StackRails,
	"julik-frontend-races-reviewer.agent.md":      detect.StackRails,
	"data-integrity-guardian.agent.md":             detect.StackRails,
	"schema-drift-detector.agent.md":              detect.StackRails,
	"data-migration-expert.agent.md":              detect.StackRails,
	"deployment-verification-agent.agent.md":      detect.StackRails,
	"lint.agent.md":                               detect.StackRails,
	"kieran-python-reviewer.agent.md":             detect.StackPython,
	"kieran-typescript-reviewer.agent.md":         detect.StackTypeScript,
}

func isStackSpecific(filename string) bool {
	_, ok := stackAgents[filepath.Base(filename)]
	return ok
}

func isForStack(filename string, stack detect.Stack) bool {
	agentStack, ok := stackAgents[filepath.Base(filename)]
	if !ok {
		return true // not stack-specific, include for all
	}
	return agentStack == stack
}
