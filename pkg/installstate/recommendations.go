package installstate

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// RepoState captures lightweight local facts used by deterministic recommendations.
type RepoState struct {
	// Compound engineering workflow
	BrainstormCount  int
	PlanCount        int
	SolutionCount    int
	HasUncheckedPlan bool
	HasCompletedPlan bool

	// Installed intelligence classification
	InstalledAgents        int
	InstalledSkills        int
	HasCopilotInstructions bool
	HasGstackStaging       bool
	HasAgentBrowserSkill   bool

	// Copilot context surface
	InstructionFileCount int  // .github/*.instructions.md
	PromptFileCount      int  // .github/prompts/*.prompt.md
	HasSetupSteps        bool // .github/copilot-setup-steps.yml

	// MCP + Extensions
	HasMCPConfig                 bool // .github/copilot-mcp-config.json
	MCPServerCount               int
	ExtensionRecommendationCount int // .vscode/extensions.json

	// gstack detail
	GstackSkillCount int  // skill dirs inside .gstack/
	HasGstackRuntime bool // .gstack/browse/dist/ exists

	// Memory
	MemoryFileCount int // .copilot-memory/ files

	// CE project config
	HasCELocalConfig   bool // compound-engineering.local.md exists
	CEReviewAgentCount int  // review_agents listed in frontmatter

	// Learning pipeline
	InstinctCount    int  // instincts in .atv/instincts/project.yaml
	HasInstincts     bool // .atv/instincts/project.yaml exists
	ObservationCount int  // lines in .atv/observations.jsonl
	HasObserverHooks bool // .github/hooks/copilot-hooks.json exists

	// User-global tool state (outside repo)
	HasGstackUserConfig      bool // ~/.gstack/ exists
	GstackSessionCount       int  // session dirs in ~/.gstack/
	HasAgentBrowserSessions  bool // ~/.agent-browser/sessions/ exists
	AgentBrowserSessionCount int  // session dirs in ~/.agent-browser/sessions/
}

// ScanRepoState inspects the local docs tree to find brainstorms, plans, and solutions.
func ScanRepoState(root string) RepoState {
	state := RepoState{}

	// Compound engineering workflow
	state.BrainstormCount = countMarkdownFiles(filepath.Join(root, "docs", "brainstorms"))
	state.PlanCount, state.HasUncheckedPlan, state.HasCompletedPlan = scanPlans(filepath.Join(root, "docs", "plans"))
	state.SolutionCount = countMarkdownFiles(filepath.Join(root, "docs", "solutions"))

	// Classify installed intelligence
	state.InstalledAgents = countFilesWithSuffix(filepath.Join(root, ".github", "agents"), ".agent.md")
	state.InstalledSkills = countSubdirs(filepath.Join(root, ".github", "skills"))
	state.HasCopilotInstructions = fileExists(filepath.Join(root, ".github", "copilot-instructions.md"))
	state.HasGstackStaging = dirExists(filepath.Join(root, ".gstack"))
	state.HasAgentBrowserSkill = dirExists(filepath.Join(root, ".github", "skills", "agent-browser"))

	// Copilot context surface
	state.InstructionFileCount = countFilesWithSuffix(filepath.Join(root, ".github"), ".instructions.md")
	state.PromptFileCount = countFilesWithSuffix(filepath.Join(root, ".github", "prompts"), ".prompt.md")
	state.HasSetupSteps = fileExists(filepath.Join(root, ".github", "copilot-setup-steps.yml"))

	// MCP + Extensions
	state.HasMCPConfig = fileExists(filepath.Join(root, ".github", "copilot-mcp-config.json"))
	state.MCPServerCount = countJSONKeys(filepath.Join(root, ".github", "copilot-mcp-config.json"), "servers")
	state.ExtensionRecommendationCount = countJSONArrayItems(filepath.Join(root, ".vscode", "extensions.json"), "recommendations")

	// gstack detail
	state.GstackSkillCount = countGstackSkills(filepath.Join(root, ".gstack"))
	state.HasGstackRuntime = dirExists(filepath.Join(root, ".gstack", "browse", "dist"))

	// Memory
	state.MemoryFileCount = countMarkdownFiles(filepath.Join(root, ".copilot-memory"))

	// CE project config
	ceLocalPath := filepath.Join(root, "compound-engineering.local.md")
	state.HasCELocalConfig = fileExists(ceLocalPath)
	if state.HasCELocalConfig {
		state.CEReviewAgentCount = countFrontmatterListItems(ceLocalPath, "review_agents")
	}

	// Learning pipeline
	instinctPath := filepath.Join(root, ".atv", "instincts", "project.yaml")
	state.HasInstincts = fileExists(instinctPath)
	if state.HasInstincts {
		state.InstinctCount = countYAMLListItems(instinctPath, "instincts")
	}
	observationsPath := filepath.Join(root, ".atv", "observations.jsonl")
	state.ObservationCount = countFileLines(observationsPath)
	state.HasObserverHooks = fileExists(filepath.Join(root, ".github", "hooks", "copilot-hooks.json"))

	// User-global tool state
	home, _ := os.UserHomeDir()
	if home != "" {
		gstackUserDir := filepath.Join(home, ".gstack")
		state.HasGstackUserConfig = dirExists(gstackUserDir)
		state.GstackSessionCount = countSubdirs(gstackUserDir)

		abSessionDir := filepath.Join(home, ".agent-browser", "sessions")
		state.HasAgentBrowserSessions = dirExists(abSessionDir)
		state.AgentBrowserSessionCount = countSubdirs(abSessionDir)
	}

	return state
}

// BuildRecommendations computes a small deterministic set of next moves from local facts.
func BuildRecommendations(root string, manifest InstallManifest) []Recommendation {
	state := ScanRepoState(root)
	var recommendations []Recommendation

	if outcome := firstProblemOutcome(manifest.Outcomes); outcome != nil {
		reason := outcome.Reason
		if reason == "" {
			reason = outcome.Detail
		}
		recommendations = append(recommendations, Recommendation{
			ID:       "fix-install-issues",
			Title:    "Fix installer warnings before relying on every capability",
			Reason:   reason,
			Priority: 100,
		})
	}

	switch {
	case state.BrainstormCount == 0:
		recommendations = append(recommendations, Recommendation{
			ID:       "start-brainstorm",
			Title:    `Start with /ce-brainstorm to shape the first feature`,
			Reason:   "No brainstorms were found in docs/brainstorms yet.",
			Priority: 90,
		})
	case state.PlanCount == 0:
		recommendations = append(recommendations, Recommendation{
			ID:       "turn-brainstorm-into-plan",
			Title:    `Turn the brainstorm into a plan with /ce-plan`,
			Reason:   "Brainstorms exist, but no plan files were found in docs/plans.",
			Priority: 85,
		})
	case state.HasUncheckedPlan:
		recommendations = append(recommendations, Recommendation{
			ID:       "execute-active-plan",
			Title:    `Continue the active plan with /ce-work`,
			Reason:   "At least one plan still has unchecked items.",
			Priority: 80,
		})
	case state.HasCompletedPlan && state.SolutionCount == 0:
		recommendations = append(recommendations, Recommendation{
			ID:       "compound-learnings",
			Title:    `Capture what shipped with /ce-compound`,
			Reason:   "Completed plans exist, but docs/solutions is still empty.",
			Priority: 70,
		})
	}

	if len(manifest.Requested.GstackDirs) > 0 && stepUsable(manifest.Outcomes, "gstack") {
		recommendations = append(recommendations, Recommendation{
			ID:       "start-gstack-sprint",
			Title:    `Use /gstack-office-hours for a deeper sprint kickoff`,
			Reason:   "gstack skills were requested and synced successfully enough to use.",
			Priority: 55,
		})
	}

	if manifest.Requested.IncludeAgentBrowser && stepUsable(manifest.Outcomes, "agent-browser") {
		recommendations = append(recommendations, Recommendation{
			ID:       "browser-check",
			Title:    "Open the app in a real browser with agent-browser",
			Reason:   "Browser automation tooling was installed or partially prepared.",
			Priority: 45,
		})
	}

	// Copilot context recommendations
	if !state.HasCopilotInstructions {
		recommendations = append(recommendations, Recommendation{
			ID:       "configure-copilot-instructions",
			Title:    "Add copilot-instructions.md to personalize Copilot",
			Reason:   "No .github/copilot-instructions.md found.",
			Priority: 85,
		})
	}
	if state.HasGstackStaging && !state.HasGstackRuntime {
		recommendations = append(recommendations, Recommendation{
			ID:       "fix-gstack-runtime",
			Title:    "Build the gstack runtime for full agent-browser support",
			Reason:   ".gstack/ exists but the browse binary was not built.",
			Priority: 75,
		})
	}
	if !state.HasMCPConfig {
		recommendations = append(recommendations, Recommendation{
			ID:       "setup-mcp-servers",
			Title:    "Configure MCP servers for external tool access",
			Reason:   "No .github/copilot-mcp-config.json found.",
			Priority: 65,
		})
	}
	if state.InstructionFileCount == 0 {
		recommendations = append(recommendations, Recommendation{
			ID:       "add-file-instructions",
			Title:    "Add file-level instructions for language-specific guidance",
			Reason:   "No .instructions.md files found in .github/.",
			Priority: 60,
		})
	}
	if state.HasObserverHooks && !state.HasInstincts && state.ObservationCount > 10 {
		recommendations = append(recommendations, Recommendation{
			ID:       "run-learn",
			Title:    "Run /learn to extract patterns from accumulated observations",
			Reason:   fmt.Sprintf("%d observations captured but no instincts created yet.", state.ObservationCount),
			Priority: 65,
		})
	}
	if state.HasInstincts && state.InstinctCount > 0 {
		recommendations = append(recommendations, Recommendation{
			ID:       "check-instincts",
			Title:    "Run /instincts to review learned patterns and check for evolution candidates",
			Reason:   fmt.Sprintf("%d instincts learned — some may be ready to evolve into skills.", state.InstinctCount),
			Priority: 40,
		})
	}

	slices.SortStableFunc(recommendations, func(a, b Recommendation) int {
		if a.Priority == b.Priority {
			return strings.Compare(a.ID, b.ID)
		}
		if a.Priority > b.Priority {
			return -1
		}
		return 1
	})

	if len(recommendations) > 5 {
		return recommendations[:5]
	}
	return recommendations
}

func countMarkdownFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		count++
	}
	return count
}

func scanPlans(dir string) (count int, hasUnchecked bool, hasCompleted bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, false, false
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		count++
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		text := string(content)
		if strings.Contains(text, "- [ ]") {
			hasUnchecked = true
		}
		if strings.Contains(text, "status: completed") {
			hasCompleted = true
		}
	}
	return count, hasUnchecked, hasCompleted
}

func firstProblemOutcome(outcomes []InstallOutcome) *InstallOutcome {
	for i := range outcomes {
		if outcomes[i].Status == InstallStepFailed || outcomes[i].Status == InstallStepWarning {
			return &outcomes[i]
		}
	}
	return nil
}

func stepUsable(outcomes []InstallOutcome, contains string) bool {
	for _, outcome := range outcomes {
		if !strings.Contains(strings.ToLower(outcome.Step), strings.ToLower(contains)) {
			continue
		}
		return outcome.Status != InstallStepFailed
	}
	return false
}

// WalkMarkdownFiles counts markdown files recursively.
func WalkMarkdownFiles(root string) int {
	count := 0
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		count++
		return nil
	})
	return count
}

func countSubdirs(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			count++
		}
	}
	return count
}

func countFilesWithSuffix(dir string, suffix string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), suffix) {
			count++
		}
	}
	return count
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// CountJSONKeyCount counts the number of keys in a top-level JSON object field.
// Used to count MCP servers in copilot-mcp-config.json.
func CountJSONKeyCount(path, field string) int {
	return countJSONKeys(path, field)
}

func countJSONKeys(path, field string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return 0
	}
	raw, ok := obj[field]
	if !ok {
		return 0
	}
	var nested map[string]json.RawMessage
	if err := json.Unmarshal(raw, &nested); err != nil {
		return 0
	}
	return len(nested)
}

// countJSONArrayItems counts items in a top-level JSON array field.
// Used to count VS Code extension recommendations.
func countJSONArrayItems(path, field string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return 0
	}
	raw, ok := obj[field]
	if !ok {
		return 0
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return 0
	}
	return len(items)
}

// countGstackSkills counts skill directories inside .gstack/ that contain a SKILL.md.
func countGstackSkills(gstackDir string) int {
	entries, err := os.ReadDir(gstackDir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() && fileExists(filepath.Join(gstackDir, entry.Name(), "SKILL.md")) {
			count++
		}
	}
	return count
}

// ListInstructionFiles returns file-level instruction filenames from .github/.
func ListInstructionFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".instructions.md") {
			names = append(names, entry.Name())
		}
	}
	return names
}

// ListPromptFiles returns prompt filenames from .github/prompts/.
func ListPromptFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".prompt.md") {
			names = append(names, entry.Name())
		}
	}
	return names
}

// ListGstackSkillNames returns skill directory names from .gstack/.
func ListGstackSkillNames(gstackDir string) []string {
	entries, err := os.ReadDir(gstackDir)
	if err != nil {
		return nil
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() && fileExists(filepath.Join(gstackDir, entry.Name(), "SKILL.md")) {
			names = append(names, entry.Name())
		}
	}
	return names
}

// ListMCPServerNames returns MCP server names from a copilot-mcp-config.json file.
func ListMCPServerNames(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil
	}
	raw, ok := obj["servers"]
	if !ok {
		return nil
	}
	var servers map[string]json.RawMessage
	if err := json.Unmarshal(raw, &servers); err != nil {
		return nil
	}
	names := make([]string, 0, len(servers))
	for name := range servers {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// countFrontmatterListItems counts YAML list items under a key in markdown frontmatter.
// Looks for lines like "review_agents:" followed by "  - agent-name" items.
func countFrontmatterListItems(path, key string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	inFrontmatter := false
	foundKey := false
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if inFrontmatter {
				break // end of frontmatter
			}
			inFrontmatter = true
			continue
		}
		if !inFrontmatter {
			continue
		}
		if trimmed == key+":" || strings.HasPrefix(trimmed, key+":") {
			foundKey = true
			continue
		}
		if foundKey {
			if strings.HasPrefix(line, "  -") || strings.HasPrefix(line, "  - ") {
				count++
			} else if !strings.HasPrefix(line, "  ") && trimmed != "" {
				break // new key
			}
		}
	}
	return count
}

// ListExtensionRecommendations returns extension IDs from .vscode/extensions.json.
func ListExtensionRecommendations(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil
	}
	raw, ok := obj["recommendations"]
	if !ok {
		return nil
	}
	var items []string
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil
	}
	return items
}

// ListMemoryFiles returns names of files in .copilot-memory/ (repo memory).
func ListMemoryFiles(root string) []string {
	memDir := filepath.Join(root, ".copilot-memory")
	var names []string
	_ = filepath.WalkDir(memDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(memDir, path)
		names = append(names, filepath.ToSlash(rel))
		return nil
	})
	return names
}

// countYAMLListItems counts top-level list items under a key in a YAML file.
// Looks for "key:" followed by lines starting with "  - ".
func countYAMLListItems(path, key string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	foundKey := false
	count := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == key+":" || strings.HasPrefix(trimmed, key+":") {
			foundKey = true
			continue
		}
		if foundKey {
			if strings.HasPrefix(line, "  -") || strings.HasPrefix(line, "- ") {
				count++
			} else if trimmed != "" && !strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
				break // new top-level key
			}
		}
	}
	return count
}

// countFileLines counts non-empty lines in a file. Returns 0 if file doesn't exist.
func countFileLines(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}
