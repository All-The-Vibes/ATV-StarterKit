package plugingen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeSkill creates a minimal template skill dir with the given
// frontmatter description under a temp skills root.
func writeSkill(t *testing.T, skillsRoot, name, frontmatter string) {
	t.Helper()
	dir := filepath.Join(skillsRoot, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(frontmatter), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
}

func TestParseSkillFrontmatter_BareQuotedSingle(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"bare", "---\nname: land\ndescription: Session completion protocol\n---\nbody", "Session completion protocol"},
		{"double", "---\nname: ce-plan\ndescription: \"Transform 'plan this' into plans\"\n---\n", "Transform 'plan this' into plans"},
		{"single", "---\nname: cb\ndescription: 'Explore requirements, say ''brainstorm'' first'\n---\n", "Explore requirements, say 'brainstorm' first"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSkillDescription([]byte(tc.body))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got != tc.want {
				t.Errorf("description = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildRoutingCatalog_EmitsLinePerSkill(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "ce-plan", "---\nname: ce-plan\ndescription: Plan things\n---\n")
	writeSkill(t, root, "land", "---\nname: land\ndescription: Wrap up the session\n---\n")

	cat, err := buildRoutingCatalog(root)
	if err != nil {
		t.Fatalf("buildRoutingCatalog: %v", err)
	}
	if len(cat) != 2 {
		t.Fatalf("catalog has %d entries, want 2", len(cat))
	}
	// Sorted by name.
	if cat[0].Name != "ce-plan" || cat[1].Name != "land" {
		t.Errorf("order = %s,%s want ce-plan,land", cat[0].Name, cat[1].Name)
	}
	if cat[0].Description != "Plan things" {
		t.Errorf("desc = %q", cat[0].Description)
	}
}

func TestBuildRoutingCatalog_SkipsMalformed(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "good", "---\nname: good\ndescription: A good skill\n---\n")
	// No frontmatter at all — malformed, must be skipped, not abort.
	writeSkill(t, root, "nofront", "just a body, no frontmatter\n")
	// Frontmatter but no description — skipped.
	writeSkill(t, root, "nodesc", "---\nname: nodesc\n---\n")

	cat, err := buildRoutingCatalog(root)
	if err != nil {
		t.Fatalf("buildRoutingCatalog must not abort on malformed: %v", err)
	}
	if len(cat) != 1 || cat[0].Name != "good" {
		t.Fatalf("expected only [good], got %+v", cat)
	}
}

func TestRenderRoutingCatalog_LLMsTxtShape(t *testing.T) {
	cat := []RoutingEntry{
		{Name: "ce-plan", Description: "Plan things"},
		{Name: "land", Description: "Wrap up"},
	}
	out := renderRoutingCatalog(cat)
	// One line per skill, format: - [/name](skills/name/SKILL.md): description
	if !strings.Contains(out, "- [/ce-plan](skills/ce-plan/SKILL.md): Plan things") {
		t.Errorf("missing ce-plan line:\n%s", out)
	}
	if !strings.Contains(out, "- [/land](skills/land/SKILL.md): Wrap up") {
		t.Errorf("missing land line:\n%s", out)
	}
	lines := strings.Count(strings.TrimSpace(out), "\n") + 1
	if lines != 2 {
		t.Errorf("want 2 lines, got %d:\n%s", lines, out)
	}
}

func TestBuildRoutingCatalog_DedupesByName(t *testing.T) {
	// Simulate the triplication risk (C3): the SAME skill name discovered
	// twice must collapse to one canonical row. buildRoutingCatalog reads a
	// single skills root so dedup is inherent, but guard the invariant:
	// duplicate dir names cannot occur in one root, so we assert the helper
	// dedupeByName collapses collisions deterministically (first wins).
	in := []RoutingEntry{
		{Name: "lfg", Description: "canonical"},
		{Name: "lfg", Description: "duplicate"},
		{Name: "ce-plan", Description: "plan"},
	}
	out := dedupeByName(in)
	if len(out) != 2 {
		t.Fatalf("want 2 after dedup, got %d", len(out))
	}
	for _, e := range out {
		if e.Name == "lfg" && e.Description != "canonical" {
			t.Errorf("dedup kept wrong lfg: %q", e.Description)
		}
	}
}
