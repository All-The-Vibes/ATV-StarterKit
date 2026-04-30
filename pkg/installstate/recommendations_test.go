package installstate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildRecommendationsForEmptyRepo(t *testing.T) {
	root := t.TempDir()
	recs := BuildRecommendations(root, InstallManifest{})
	if len(recs) == 0 || recs[0].ID != "start-brainstorm" {
		t.Fatalf("expected start-brainstorm recommendation first, got %+v", recs)
	}
}

func TestBuildRecommendationsPrefersActivePlanWork(t *testing.T) {
	root := t.TempDir()
	mustWriteMarkdown(t, filepath.Join(root, "docs", "brainstorms", "idea.md"), "# Idea")
	mustWriteMarkdown(t, filepath.Join(root, "docs", "plans", "work.md"), "status: active\n- [ ] do the work\n")
	// Satisfy copilot-instructions rec so it doesn't outrank the plan rec.
	mustWriteMarkdown(t, filepath.Join(root, ".github", "copilot-instructions.md"), "# Instructions")

	manifest := InstallManifest{
		Requested: RequestedState{GstackDirs: []string{"review"}},
		Outcomes:  []InstallOutcome{{Step: "Syncing gstack skills", Status: InstallStepDone}},
	}

	recs := BuildRecommendations(root, manifest)
	if len(recs) == 0 || recs[0].ID != "execute-active-plan" {
		t.Fatalf("expected execute-active-plan recommendation first, got %+v", recs)
	}
	// Verify gstack recommendation appears.
	found := false
	for _, r := range recs {
		if r.ID == "start-gstack-sprint" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected start-gstack-sprint in recommendations, got %+v", recs)
	}
}

func TestBuildRecommendationsPrioritizesInstallProblems(t *testing.T) {
	root := t.TempDir()
	mustWriteMarkdown(t, filepath.Join(root, "docs", "brainstorms", "idea.md"), "# Idea")

	manifest := InstallManifest{
		Outcomes: []InstallOutcome{{Step: "Installing agent-browser", Status: InstallStepWarning, Reason: "npm not found"}},
	}

	recs := BuildRecommendations(root, manifest)
	if len(recs) == 0 || recs[0].ID != "fix-install-issues" {
		t.Fatalf("expected fix-install-issues first, got %+v", recs)
	}
}

func TestBuildRecommendationsIsDeterministicAndCapped(t *testing.T) {
	root := t.TempDir()
	mustWriteMarkdown(t, filepath.Join(root, "docs", "brainstorms", "idea.md"), "# Idea")
	mustWriteMarkdown(t, filepath.Join(root, "docs", "plans", "work.md"), "status: active\n- [ ] keep going\n")

	manifest := InstallManifest{
		Requested: RequestedState{
			GstackDirs:          []string{"review"},
			IncludeAgentBrowser: true,
		},
		Outcomes: []InstallOutcome{
			{Step: "Installing agent-browser", Status: InstallStepWarning, Reason: "npm not found"},
			{Step: "Syncing gstack skills", Status: InstallStepDone},
		},
	}

	first := BuildRecommendations(root, manifest)
	second := BuildRecommendations(root, manifest)

	if len(first) != len(second) {
		t.Fatalf("recommendation lengths differ: %d vs %d", len(first), len(second))
	}
	if len(first) != 5 {
		t.Fatalf("expected recommendations to be capped at 5, got %d: %+v", len(first), first)
	}
	for i := range first {
		if first[i].ID != second[i].ID {
			t.Fatalf("recommendation order differs at %d: %+v vs %+v", i, first, second)
		}
	}
	// fix-install-issues (P100) should always be first.
	if first[0].ID != "fix-install-issues" {
		t.Fatalf("expected fix-install-issues first, got %+v", first)
	}
}

func mustWriteMarkdown(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestBuildRecommendations_NoAddPromptsRecommendation guards the removal of
// the legacy "add-prompts" recommendation. The installer now ships shim
// .prompt.md files by default, so prompting users to create one is misleading.
func TestBuildRecommendations_NoAddPromptsRecommendation(t *testing.T) {
root := t.TempDir()
recs := BuildRecommendations(root, InstallManifest{})
for _, r := range recs {
if r.ID == "add-prompts" {
t.Fatalf("legacy add-prompts recommendation reappeared in BuildRecommendations output: %+v", r)
}
}
}
