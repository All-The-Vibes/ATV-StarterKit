package plugingen

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// routingFixture is one parsed line of testdata/routing-fixtures.txt.
type routingFixture struct {
	prompt string
	target string
}

func loadRoutingFixtures(t *testing.T) []routingFixture {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "routing-fixtures.txt"))
	if err != nil {
		t.Fatalf("open fixtures: %v", err)
	}
	defer f.Close()

	var out []routingFixture
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		if len(parts) != 2 {
			t.Fatalf("malformed fixture line: %q", line)
		}
		out = append(out, routingFixture{
			prompt: strings.TrimSpace(parts[0]),
			target: strings.TrimSpace(parts[1]),
		})
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan fixtures: %v", err)
	}
	return out
}

// TestRoutingFixtures_TargetsResolveToRealSkills is the Layer-1 deterministic
// guard: every fixture whose expected target names a skill (directly, via
// emit:, or via force:) must resolve to a skill that actually exists in the
// generated catalog. This catches a renamed/removed skill that would silently
// break routing, without asserting any live-model behavior.
func TestRoutingFixtures_TargetsResolveToRealSkills(t *testing.T) {
	skillsRoot := filepath.Join(repoRoot(t), "pkg", "scaffold", "templates", "skills")
	cat, err := BuildRoutingCatalog(skillsRoot)
	if err != nil {
		t.Fatalf("BuildRoutingCatalog: %v", err)
	}
	known := make(map[string]bool, len(cat))
	for _, e := range cat {
		known[e.Name] = true
	}

	for _, fx := range loadRoutingFixtures(t) {
		skill := targetSkill(fx.target)
		if skill == "" {
			continue // no-match / control command: no skill to resolve
		}
		if !known[skill] {
			t.Errorf("fixture %q expects skill %q which is not in the catalog", fx.prompt, skill)
		}
	}
}

// targetSkill extracts the skill name from a fixture target, or "" if the
// target is a control command or no-match (no skill to resolve).
func targetSkill(target string) string {
	switch {
	case target == "no-match":
		return ""
	case strings.HasPrefix(target, "control:"):
		return ""
	case strings.HasPrefix(target, "emit:"):
		return strings.TrimPrefix(target, "emit:")
	case strings.HasPrefix(target, "force:"):
		return strings.TrimPrefix(target, "force:")
	default:
		return target
	}
}

// TestRoutingFixtures_CoverRequiredEdgeCases asserts the required GATE-3 edge
// cases are represented in the fixture set: no-match floor, a PROACTIVE toggle
// (control), a force-skill case, and a build emit case. Without these the live
// smoke run wouldn't exercise the router's safety behaviors.
func TestRoutingFixtures_CoverRequiredEdgeCases(t *testing.T) {
	fixtures := loadRoutingFixtures(t)
	var hasNoMatch, hasControl, hasForce, hasEmit, hasCollision bool
	seenReviewPrompts := 0
	for _, fx := range fixtures {
		switch {
		case fx.target == "no-match":
			hasNoMatch = true
		case strings.HasPrefix(fx.target, "control:"):
			hasControl = true
		case strings.HasPrefix(fx.target, "force:"):
			hasForce = true
		case strings.HasPrefix(fx.target, "emit:"):
			hasEmit = true
		}
		if strings.Contains(fx.prompt, "review") {
			seenReviewPrompts++
		}
	}
	// A "collision" case = more than one prompt containing the same ambiguous
	// word ("review") routing into the catalog.
	hasCollision = seenReviewPrompts >= 2

	for name, ok := range map[string]bool{
		"no-match floor":        hasNoMatch,
		"control (PROACTIVE)":   hasControl,
		"force-skill syntax":    hasForce,
		"build emit":            hasEmit,
		"collision (ambiguous)": hasCollision,
	} {
		if !ok {
			t.Errorf("routing fixtures missing required edge case: %s", name)
		}
	}
}

// TestRoutingFixtures_HaveMinimumCoverage ensures the fixture set stays broad
// enough to be a meaningful smoke suite (plan: ~20 prompts).
func TestRoutingFixtures_HaveMinimumCoverage(t *testing.T) {
	fixtures := loadRoutingFixtures(t)
	if len(fixtures) < 20 {
		t.Errorf("only %d routing fixtures; plan calls for ~20 covering the catalog", len(fixtures))
	}
}
