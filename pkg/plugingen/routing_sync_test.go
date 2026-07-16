package plugingen

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRoutingCatalog_CommittedLLMsTxtIsFresh guards against drift between the
// committed pkg/scaffold/templates/skills/atv/llms.txt catalog and what the
// builder produces from the current set of template skills. When a skill is
// added, removed, or its frontmatter description changes, this test fails
// until llms.txt is regenerated — the same posture as `plugingen -check`.
func TestRoutingCatalog_CommittedLLMsTxtIsFresh(t *testing.T) {
	root := repoRoot(t)
	skillsRoot := filepath.Join(root, "pkg", "scaffold", "templates", "skills")

	cat, err := BuildRoutingCatalog(skillsRoot)
	if err != nil {
		t.Fatalf("BuildRoutingCatalog: %v", err)
	}
	want := RenderRoutingCatalog(cat)

	committedPath := filepath.Join(skillsRoot, "atv", "llms.txt")
	gotBytes, err := os.ReadFile(committedPath)
	if err != nil {
		t.Fatalf("read committed llms.txt: %v", err)
	}
	got := normalizeLineEndings(string(gotBytes))

	if got != want {
		t.Errorf("llms.txt is stale — regenerate it.\n"+
			"Run: go run ./cmd/plugingen (or the catalog regen step)\n"+
			"committed %d bytes, want %d bytes", len(got), len(want))
	}
}

// TestRoutingCatalog_IncludesKeyTargets asserts the catalog contains the
// routing targets the /atv router depends on, so a rename/removal that would
// break routing surfaces here.
func TestRoutingCatalog_IncludesKeyTargets(t *testing.T) {
	root := repoRoot(t)
	skillsRoot := filepath.Join(root, "pkg", "scaffold", "templates", "skills")
	cat, err := BuildRoutingCatalog(skillsRoot)
	if err != nil {
		t.Fatalf("BuildRoutingCatalog: %v", err)
	}
	have := make(map[string]bool, len(cat))
	for _, e := range cat {
		have[e.Name] = true
	}
	for _, want := range []string{"atv", "lfg", "slfg", "ce-plan", "ce-review", "test-browser", "atv-security"} {
		if !have[want] {
			t.Errorf("routing catalog missing key target %q", want)
		}
	}
}

// TestRoutingCatalog_ShipsInEveryPluginCopy verifies the generator copies
// llms.txt (a SKILL.md sidecar) into every generated plugin that includes the
// atv skill, so the router's menu is present wherever the skill is installed
// (T11 install topology).
func TestRoutingCatalog_ShipsInEveryPluginCopy(t *testing.T) {
	tmp := regenerateInto(t)
	for _, plugin := range []string{"atv-skill-atv", "atv-pack-shipping", "atv-everything"} {
		p := filepath.Join(tmp, "plugins", plugin, "skills", "atv", "llms.txt")
		if _, err := os.Stat(p); err != nil {
			t.Errorf("llms.txt missing from generated %s: %v", plugin, err)
		}
	}
}
