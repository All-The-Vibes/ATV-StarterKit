package gstack

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDetectPrerequisites(t *testing.T) {
	prereqs := DetectPrerequisites()

	// Git should be available in any dev environment
	if !prereqs.HasGit {
		t.Log("git not found — this is expected in some CI environments")
	} else {
		if prereqs.GitVersion == "" || prereqs.GitVersion == "unknown" {
			t.Error("git detected but version is empty or unknown")
		}
	}

	// Bun and Node are optional
	t.Logf("Prerequisites: %s", prereqs.Summary())
}

func TestPrerequisitesSummary(t *testing.T) {
	p := Prerequisites{
		HasGit:      true,
		HasBun:      true,
		HasNode:     true,
		GitVersion:  "2.44.0",
		BunVersion:  "1.1.0",
		NodeVersion: "v20.0.0",
	}

	summary := p.Summary()
	if summary == "" {
		t.Error("summary should not be empty")
	}
	if !contains(summary, "git 2.44.0") {
		t.Errorf("summary missing git version: %s", summary)
	}
	if !contains(summary, "bun 1.1.0") {
		t.Errorf("summary missing bun version: %s", summary)
	}
}

func TestPrerequisitesSummaryMissing(t *testing.T) {
	p := Prerequisites{
		HasGit: false,
		HasBun: false,
	}

	summary := p.Summary()
	if !contains(summary, "git (missing)") {
		t.Errorf("summary should show git as missing: %s", summary)
	}
	if !contains(summary, "bun (missing)") {
		t.Errorf("summary should show bun as missing: %s", summary)
	}
}

func TestRuntimeAvailable(t *testing.T) {
	tests := []struct {
		name     string
		prereqs  Prerequisites
		expected bool
	}{
		{"bun only", Prerequisites{HasBun: true}, true},
		{"node only", Prerequisites{HasNode: true}, true},
		{"both", Prerequisites{HasBun: true, HasNode: true}, true},
		{"neither", Prerequisites{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.prereqs.RuntimeAvailable(); got != tt.expected {
				t.Errorf("RuntimeAvailable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCopyGeneratedSkills(t *testing.T) {
	tmpDir := t.TempDir()
	gstackDir := filepath.Join(tmpDir, ".gstack")
	skillsDir := filepath.Join(tmpDir, ".github", "skills")

	// Create fake generated .agents/skills/gstack-* dirs
	for _, name := range []string{"gstack-review", "gstack-qa", "gstack-ship"} {
		dir := filepath.Join(gstackDir, ".agents", "skills", name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	copied, dirs := copyGeneratedSkills(gstackDir, skillsDir)
	if !copied {
		t.Error("expected skills to be copied")
	}
	if len(dirs) != 3 {
		t.Errorf("expected 3 skill dirs, got %d: %v", len(dirs), dirs)
	}

	// Verify files exist in target
	for _, name := range []string{"gstack-review", "gstack-qa", "gstack-ship"} {
		skillPath := filepath.Join(skillsDir, name, "SKILL.md")
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			t.Errorf("expected %s to exist", skillPath)
		}
	}
}

func TestCopyFallbackRawSkills(t *testing.T) {
	tmpDir := t.TempDir()
	gstackDir := filepath.Join(tmpDir, ".gstack")
	skillsDir := filepath.Join(tmpDir, ".github", "skills")

	// Create fake raw skill dirs (no .agents/, simulating failed gen)
	for _, name := range []string{"review", "qa"} {
		dir := filepath.Join(gstackDir, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("# "+name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	copied, dirs := copyGeneratedSkills(gstackDir, skillsDir)
	if !copied {
		t.Error("expected fallback copy to work")
	}
	if len(dirs) != 2 {
		t.Errorf("expected 2 skill dirs, got %d: %v", len(dirs), dirs)
	}
	// Fallback should prefix with gstack-
	for _, d := range dirs {
		if len(d) < 7 || d[:7] != "gstack-" {
			t.Errorf("fallback dirs should be prefixed with gstack-, got %s", d)
		}
	}
}

func TestInstallIdempotent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fake existing .gstack/ install
	gstackDir := filepath.Join(tmpDir, ".gstack")
	if err := os.MkdirAll(gstackDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gstackDir, "SKILL.md"), []byte("# gstack"), 0644); err != nil {
		t.Fatal(err)
	}
	// Create .github/skills/ target
	if err := os.MkdirAll(filepath.Join(tmpDir, ".github", "skills"), 0755); err != nil {
		t.Fatal(err)
	}

	result := Install(tmpDir, ModeMarkdownOnly)
	if result.Error != nil {
		t.Fatalf("idempotent install should not error: %v", result.Error)
	}
	if result.Warning == "" {
		t.Error("idempotent install should produce a warning")
	}
	if result.Cloned {
		t.Error("should not clone when already installed")
	}
}

func TestFindBashSkipsWSL(t *testing.T) {
	bashPath := findBash()
	if bashPath == "" {
		t.Skip("no bash found on this system")
	}
	t.Logf("findBash() returned: %s", bashPath)

	// On Windows, findBash must NOT return WSL bash (System32) or WindowsApps bash
	if runtime.GOOS == "windows" {
		norm := strings.ToLower(filepath.Clean(bashPath))
		if strings.Contains(norm, "system32") {
			t.Errorf("findBash() returned WSL bash at %s — should prefer Git Bash", bashPath)
		}
		if strings.Contains(norm, "windowsapps") {
			t.Errorf("findBash() returned WindowsApps bash at %s — should prefer Git Bash", bashPath)
		}
	}
}

func TestPruneNonGitHubDirs(t *testing.T) {
	tmpDir := t.TempDir()
	gstackDir := filepath.Join(tmpDir, ".gstack")

	// Directories that should be REMOVED
	pruneDirs := []string{
		".cursor/skills", ".factory/skills", ".kiro/skills",
		".openclaw/skills", ".opencode/skills", ".slate/skills",
		"codex", "openclaw",
		"node_modules", ".git/objects", ".github/workflows",
		"extension", "hosts", "contrib", "supabase", "test", "scripts", "docs",
	}
	for _, d := range pruneDirs {
		if err := os.MkdirAll(filepath.Join(gstackDir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Directories that should SURVIVE
	keepDirs := []string{
		"office-hours", "review", "ship", "bin", "browse/dist", "lib",
		".agents/skills/gstack-review",
	}
	for _, d := range keepDirs {
		if err := os.MkdirAll(filepath.Join(gstackDir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Files that should survive
	if err := os.WriteFile(filepath.Join(gstackDir, "SKILL.md"), []byte("# gstack"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gstackDir, "ETHOS.md"), []byte("# ethos"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run the prune
	pruneNonGitHubDirs(gstackDir)

	// Verify pruned dirs are gone
	pruneRoots := []string{
		".cursor", ".factory", ".kiro", ".openclaw", ".opencode", ".slate",
		"codex", "openclaw", "node_modules", ".git", ".github",
		"extension", "hosts", "contrib", "supabase", "test", "scripts", "docs",
	}
	for _, d := range pruneRoots {
		path := filepath.Join(gstackDir, d)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("expected %s to be pruned, but it still exists", d)
		}
	}

	// Verify kept dirs survive
	for _, d := range keepDirs {
		path := filepath.Join(gstackDir, d)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to survive pruning, but it was removed", d)
		}
	}

	// Verify kept files survive
	for _, f := range []string{"SKILL.md", "ETHOS.md"} {
		path := filepath.Join(gstackDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected %s to survive pruning, but it was removed", f)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
