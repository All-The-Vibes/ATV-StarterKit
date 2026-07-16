package plugingen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestInvestigateSkill_BodyHasMethodology asserts the /investigate skill body
// carries the systematic-debugging methodology it was ported for: the Iron Law
// (no fix before root cause) and all five phases. A content-thin copy would
// route users to a skill that doesn't actually enforce repro-first debugging.
func TestInvestigateSkill_BodyHasMethodology(t *testing.T) {
	root := repoRoot(t)
	for _, rel := range []string{
		filepath.Join("pkg", "scaffold", "templates", "skills", "investigate", "SKILL.md"),
		filepath.Join(".github", "skills", "investigate", "SKILL.md"),
	} {
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		text := string(body)
		for _, want := range []string{
			"Iron Law",
			"NO FIX WITHOUT ROOT-CAUSE",
			"Phase 1: Root-Cause Investigation",
			"Phase 2: Pattern Analysis",
			"Phase 3: Hypothesis Testing",
			"Phase 4: Implementation",
			"Phase 5: Verification & Report",
			"DEBUG REPORT",
		} {
			if !strings.Contains(text, want) {
				t.Errorf("%s missing required methodology marker %q", rel, want)
			}
		}
	}
}

// TestAtvRouter_BugRouteIsInvestigate asserts both /atv router copies route
// bugs to /investigate and no longer carry the provisional /ce-work note. This
// guards the T7 route flip against a regression in either the template or the
// dogfood mirror.
func TestAtvRouter_BugRouteIsInvestigate(t *testing.T) {
	root := repoRoot(t)
	for _, rel := range []string{
		filepath.Join("pkg", "scaffold", "templates", "skills", "atv", "SKILL.md"),
		filepath.Join(".github", "skills", "atv", "SKILL.md"),
	} {
		body, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		text := string(body)
		if !strings.Contains(text, "/investigate") {
			t.Errorf("%s: router should route bugs to /investigate", rel)
		}
		if strings.Contains(text, "provisional") {
			t.Errorf("%s: provisional bug-route note should be removed", rel)
		}
	}
}
