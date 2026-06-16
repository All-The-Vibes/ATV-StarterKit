package scaffold

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// known_pitfalls_test.go keeps .github/known-pitfalls.md — the consolidated
// DO-NOT register for ATV contributors — honest. The register documents the
// pitfalls the codebase already guards against, and tags each machine-enforced
// entry with an HTML-comment marker of the form:
//
//	<!-- enforced-by: pkg/scaffold/skillspec_test.go -->
//	<!-- enforced-by: cmd/plugingen/main.go, .github/workflows/ci.yml -->
//
// This test parses those markers and asserts that every referenced guard still
// exists on disk. The failure mode it prevents: someone renames or deletes a
// test/workflow that a pitfall entry claims enforces it, leaving the register
// advertising a check that no longer runs. That is the same staleness guard
// ATV's own allowlists use (see no_claude_refs_test.go and parity_test.go).

// enforcedByMarker matches a single `<!-- enforced-by: a, b, c -->` comment and
// captures the comma-separated payload.
var enforcedByMarker = regexp.MustCompile(`(?i)<!--\s*enforced-by:\s*(.+?)\s*-->`)

// knownPitfallsPath returns the repo-relative location of the register.
const knownPitfallsRel = ".github/known-pitfalls.md"

func TestKnownPitfallsRegisterExists(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, filepath.FromSlash(knownPitfallsRel))

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s is missing or unreadable: %v\n"+
			"The consolidated DO-NOT register is a required contributor artifact.", knownPitfallsRel, err)
	}

	content := string(raw)

	// Substance check (an invariant, not a snapshot): the register must carry
	// at least one imperative entry. We assert presence, never an exact count,
	// so legitimately adding or removing entries never fails this test.
	if !strings.Contains(content, "### DO NOT") {
		t.Errorf("%s contains no `### DO NOT` entries — it must hold at least one imperative pitfall.", knownPitfallsRel)
	}
}

// TestKnownPitfallsEnforcedByReferencesResolve asserts that every path named in
// an `enforced-by:` marker exists, so the register cannot advertise a guard
// that has been moved or deleted.
func TestKnownPitfallsEnforcedByReferencesResolve(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, filepath.FromSlash(knownPitfallsRel))

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s is missing or unreadable: %v", knownPitfallsRel, err)
	}

	matches := enforcedByMarker.FindAllStringSubmatch(string(raw), -1)
	if len(matches) == 0 {
		t.Fatalf("%s declares no `<!-- enforced-by: ... -->` markers — "+
			"every machine-enforced entry must name its guard so this test can keep it honest.", knownPitfallsRel)
	}

	seen := make(map[string]bool)
	for _, m := range matches {
		for _, ref := range strings.Split(m[1], ",") {
			ref = strings.TrimSpace(ref)
			if ref == "" {
				t.Errorf("empty path in an `enforced-by:` marker: %q", m[0])
				continue
			}
			// Skip documentation placeholders like `<repo-relative-path>`.
			// A real repo path never contains angle brackets, so this lets
			// the "how to add a pitfall" example show the marker syntax
			// verbatim without being treated as a live reference.
			if strings.ContainsAny(ref, "<>") {
				continue
			}
			if strings.HasPrefix(ref, "/") || strings.Contains(ref, "..") {
				t.Errorf("enforced-by reference %q must be a clean repo-relative path (no leading slash, no `..`)", ref)
				continue
			}
			if seen[ref] {
				continue
			}
			seen[ref] = true

			abs := filepath.Join(root, filepath.FromSlash(ref))
			if _, err := os.Stat(abs); err != nil {
				t.Errorf("enforced-by reference %q does not exist: %v\n"+
					"Update %s to point at the guard's new location, or remove the stale marker.", ref, err, knownPitfallsRel)
			}
		}
	}
}
