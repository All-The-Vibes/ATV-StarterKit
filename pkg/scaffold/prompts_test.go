package scaffold

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/detect"
)

// TestBuildPromptShim_HappyPath verifies that a generated shim contains the
// VS Code Copilot Chat front-matter, a description that names the skill, and
// a body that delegates to the canonical SKILL.md path.
func TestBuildPromptShim_HappyPath(t *testing.T) {
	got := string(BuildPromptShim("ce-plan"))

	mustContain := []string{
		"mode: agent",
		"ce-plan",
		".github/skills/ce-plan/SKILL.md",
	}
	for _, sub := range mustContain {
		if !strings.Contains(got, sub) {
			t.Errorf("BuildPromptShim(\"ce-plan\") missing %q\n--- got ---\n%s", sub, got)
		}
	}

	if !strings.HasPrefix(got, "---\n") {
		t.Errorf("BuildPromptShim output should start with YAML front-matter delimiter, got prefix %q", firstLine(got))
	}

	// Front-matter must contain a description: line; VS Code surfaces it in the picker.
	if !strings.Contains(got, "\ndescription:") {
		t.Errorf("BuildPromptShim output missing `description:` line in front-matter\n--- got ---\n%s", got)
	}
}

// TestBuildPromptShim_HyphenatedSkillName ensures hyphens in skill directory
// names round-trip through both the front-matter description and the skill
// link target without escaping or quoting damage.
func TestBuildPromptShim_HyphenatedSkillName(t *testing.T) {
	got := string(BuildPromptShim("ce-brainstorm"))

	if !strings.Contains(got, ".github/skills/ce-brainstorm/SKILL.md") {
		t.Errorf("hyphenated skill name not preserved in link target\n--- got ---\n%s", got)
	}
	if !strings.Contains(got, "ce-brainstorm") {
		t.Errorf("hyphenated skill name missing from output\n--- got ---\n%s", got)
	}
}

// TestPromptShimAllowListIsSubsetOfShippedSkills protects the invariant that
// every shim entry corresponds to a skill the installer actually ships.
// A shim referencing a non-existent skill would generate a dead link.
func TestPromptShimAllowListIsSubsetOfShippedSkills(t *testing.T) {
	shipped := make(map[string]bool)
	for _, n := range coreSkillDirectories {
		shipped[n] = true
	}
	for _, n := range orchestratorSkillDirectories {
		shipped[n] = true
	}
	for _, n := range easterEggSkillDirectories {
		shipped[n] = true
	}

	for _, name := range promptShimSkillDirectories {
		if !shipped[name] {
			t.Errorf("promptShimSkillDirectories entry %q is not present in any catalog skill slice "+
				"(coreSkillDirectories, orchestratorSkillDirectories, easterEggSkillDirectories)", name)
		}
	}
}

// TestPromptShimExclusionsCoverAllUnshippedSkills enforces that every shipped
// skill is either an allow-listed shim or explicitly listed as non-user-facing.
// Adding a new skill without making this decision will fail the test.
func TestPromptShimExclusionsCoverAllUnshippedSkills(t *testing.T) {
	allow := make(map[string]bool, len(promptShimSkillDirectories))
	for _, n := range promptShimSkillDirectories {
		allow[n] = true
	}
	exclude := make(map[string]bool, len(nonUserFacingSkills))
	for _, n := range nonUserFacingSkills {
		exclude[n] = true
	}

	all := append(append([]string{}, coreSkillDirectories...), orchestratorSkillDirectories...)
	all = append(all, easterEggSkillDirectories...)

	var unclassified []string
	for _, name := range all {
		if !allow[name] && !exclude[name] {
			unclassified = append(unclassified, name)
		}
	}
	if len(unclassified) > 0 {
		t.Fatalf("skills not classified as either user-facing (promptShimSkillDirectories) or "+
			"non-user-facing (nonUserFacingSkills) in pkg/scaffold/prompts.go: %v", unclassified)
	}

	// The two sets must be disjoint — a skill cannot be both shimmed and excluded.
	for name := range allow {
		if exclude[name] {
			t.Errorf("skill %q appears in both promptShimSkillDirectories and nonUserFacingSkills", name)
		}
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// --- Unit 2: catalog wiring tests ---

func TestBuildCatalog_IncludesAllPromptShims(t *testing.T) {
	comps := BuildCatalog(detectStackForTest())

	got := promptShimPaths(comps)
	for _, name := range promptShimSkillDirectories {
		want := ".github/prompts/" + name + ".prompt.md"
		if !got[want] {
			t.Errorf("BuildCatalog missing prompt component %q", want)
		}
	}
}

func TestBuildCatalog_PromptShimContentMatchesBuilder(t *testing.T) {
	comps := BuildCatalog(detectStackForTest())
	target := ".github/prompts/ce-plan.prompt.md"
	for _, c := range comps {
		if filepathToSlash(c.Path) == target {
			want := string(BuildPromptShim("ce-plan"))
			if string(c.Content) != want {
				t.Errorf("catalog component %q content drifted from BuildPromptShim output", target)
			}
			return
		}
	}
	t.Fatalf("catalog component %q not found", target)
}

func TestBuildFilteredCatalog_CoreSkillsOnlyIncludesCoreShims(t *testing.T) {
	comps := BuildFilteredCatalog(detectStackForTest(), []string{"core-skills"})
	got := promptShimPaths(comps)

	if !got[".github/prompts/ce-plan.prompt.md"] {
		t.Errorf("core-skills layer missing ce-plan.prompt.md")
	}
	if got[".github/prompts/lfg.prompt.md"] {
		t.Errorf("core-skills layer should NOT include orchestrator shim lfg.prompt.md")
	}
}

func TestBuildFilteredCatalog_OrchestratorsOnlyIncludesOrchestratorShims(t *testing.T) {
	comps := BuildFilteredCatalog(detectStackForTest(), []string{"orchestrators"})
	got := promptShimPaths(comps)

	if !got[".github/prompts/lfg.prompt.md"] {
		t.Errorf("orchestrators layer missing lfg.prompt.md")
	}
	if got[".github/prompts/ce-plan.prompt.md"] {
		t.Errorf("orchestrators layer should NOT include core-skill shim ce-plan.prompt.md")
	}
}

func TestBuildFilteredCatalog_NoLayersEmitsZeroShims(t *testing.T) {
	comps := BuildFilteredCatalog(detectStackForTest(), []string{})
	for _, c := range comps {
		if isPromptShimPath(c.Path) {
			t.Errorf("expected zero prompt shims with no layers selected, got %q", c.Path)
		}
	}
}

func TestBuildCatalog_IncludesPromptsDirectory(t *testing.T) {
	comps := BuildCatalog(detectStackForTest())
	for _, c := range comps {
		if c.IsDir && filepathToSlash(c.Path) == ".github/prompts" {
			return
		}
	}
	t.Errorf(".github/prompts directory component missing from BuildCatalog output")
}

func detectStackForTest() detect.Stack { return detect.StackGeneral }

func filepathToSlash(p string) string { return filepath.ToSlash(p) }

func isPromptShimPath(p string) bool {
	s := filepathToSlash(p)
	return strings.HasPrefix(s, ".github/prompts/") && strings.HasSuffix(s, ".prompt.md")
}

func promptShimPaths(comps []Component) map[string]bool {
	out := map[string]bool{}
	for _, c := range comps {
		if isPromptShimPath(c.Path) {
			out[filepathToSlash(c.Path)] = true
		}
	}
	return out
}
