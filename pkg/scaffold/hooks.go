package scaffold

// HookType represents each of the 6 Copilot lifecycle hooks.
type HookType int

const (
	HookSystemInstructions HookType = 1 // .github/copilot-instructions.md
	HookSetupSteps         HookType = 2 // .github/copilot-setup-steps.yml
	HookMCPServers         HookType = 3 // .github/copilot-mcp-config.json
	HookSkills             HookType = 4 // .github/skills/*/SKILL.md
	HookAgents             HookType = 5 // .github/agents/*.agent.md
	HookFileInstructions   HookType = 6 // .github/*.instructions.md
)

// HookName returns the human-readable name for a hook type.
func HookName(h HookType) string {
	switch h {
	case HookSystemInstructions:
		return "System Instructions"
	case HookSetupSteps:
		return "Setup Steps"
	case HookMCPServers:
		return "MCP Servers"
	case HookSkills:
		return "Skills"
	case HookAgents:
		return "Agents"
	case HookFileInstructions:
		return "File Instructions"
	default:
		return "Unknown"
	}
}
