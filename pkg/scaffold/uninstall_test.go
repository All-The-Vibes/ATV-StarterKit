package scaffold

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestUninstallRemovesATVDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create ATV-owned directories with content
	for _, dir := range atvDirectories {
		fullPath := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			t.Fatal(err)
		}
		// Put a file in each to make sure recursive removal works
		if err := os.WriteFile(filepath.Join(fullPath, "test.md"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	result := Uninstall(tmpDir, nil, false)

	if len(result.Removed) != len(atvDirectories) {
		t.Errorf("expected %d removals, got %d: %v", len(atvDirectories), len(result.Removed), result.Removed)
	}
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	// Verify directories are gone
	for _, dir := range atvDirectories {
		fullPath := filepath.Join(tmpDir, dir)
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed", dir)
		}
	}
}

func TestUninstallRemovesATVFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the files
	for _, file := range atvFiles {
		fullPath := filepath.Join(tmpDir, file)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("original content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	result := Uninstall(tmpDir, nil, false)

	if len(result.Removed) != len(atvFiles) {
		t.Errorf("expected %d file removals, got %d", len(atvFiles), len(result.Removed))
	}

	// Verify files are gone
	for _, file := range atvFiles {
		fullPath := filepath.Join(tmpDir, file)
		if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed", file)
		}
	}
}

func TestUninstallPreservesModifiedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	originalContent := []byte("original copilot instructions")
	modifiedContent := []byte("user customized this file")

	// Create copilot-instructions.md
	instrPath := filepath.Join(tmpDir, ".github", "copilot-instructions.md")
	if err := os.MkdirAll(filepath.Dir(instrPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Compute checksum of original
	h := sha256.Sum256(originalContent)
	originalChecksum := hex.EncodeToString(h[:])
	checksums := map[string]string{
		".github/copilot-instructions.md": originalChecksum,
	}

	// Write MODIFIED content (different from checksum)
	if err := os.WriteFile(instrPath, modifiedContent, 0644); err != nil {
		t.Fatal(err)
	}

	result := Uninstall(tmpDir, checksums, false)

	// File should be skipped (modified)
	if len(result.Skipped) != 1 {
		t.Errorf("expected 1 skipped file, got %d: %v", len(result.Skipped), result.Skipped)
	}

	// File should still exist
	if _, err := os.Stat(instrPath); os.IsNotExist(err) {
		t.Error("modified file should not have been removed")
	}
}

func TestUninstallForceRemovesModifiedFiles(t *testing.T) {
	tmpDir := t.TempDir()

	originalContent := []byte("original copilot instructions")
	modifiedContent := []byte("user customized this file")

	instrPath := filepath.Join(tmpDir, ".github", "copilot-instructions.md")
	if err := os.MkdirAll(filepath.Dir(instrPath), 0755); err != nil {
		t.Fatal(err)
	}

	h := sha256.Sum256(originalContent)
	originalChecksum := hex.EncodeToString(h[:])
	checksums := map[string]string{
		".github/copilot-instructions.md": originalChecksum,
	}

	if err := os.WriteFile(instrPath, modifiedContent, 0644); err != nil {
		t.Fatal(err)
	}

	result := Uninstall(tmpDir, checksums, true)

	// Force mode should remove even modified files
	if len(result.Skipped) != 0 {
		t.Errorf("force mode should skip nothing, got %v", result.Skipped)
	}

	if _, err := os.Stat(instrPath); !os.IsNotExist(err) {
		t.Error("force mode should have removed the modified file")
	}
}

func TestUninstallPreservesNonEmptyDocDirs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create docs/plans with user content
	plansDir := filepath.Join(tmpDir, "docs", "plans")
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plansDir, "my-plan.md"), []byte("important plan"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create docs/brainstorms empty
	brainstormsDir := filepath.Join(tmpDir, "docs", "brainstorms")
	if err := os.MkdirAll(brainstormsDir, 0755); err != nil {
		t.Fatal(err)
	}

	result := Uninstall(tmpDir, nil, false)

	// plans should be skipped (has content)
	found := false
	for _, s := range result.Skipped {
		if s == "docs/plans (has user content)" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected docs/plans to be skipped, result: %+v", result)
	}

	// brainstorms should be removed (empty)
	if _, err := os.Stat(brainstormsDir); !os.IsNotExist(err) {
		t.Error("empty docs/brainstorms should have been removed")
	}

	// plans should still exist
	if _, err := os.Stat(plansDir); os.IsNotExist(err) {
		t.Error("docs/plans with content should still exist")
	}
}

func TestUninstallMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create anything — everything should be "missing"
	result := Uninstall(tmpDir, nil, false)

	expectedMissing := len(atvDirectories) + len(atvFiles) + len(atvDocDirectories)
	if len(result.Missing) != expectedMissing {
		t.Errorf("expected %d missing items, got %d", expectedMissing, len(result.Missing))
	}
	if len(result.Removed) != 0 {
		t.Errorf("expected 0 removals on empty dir, got %d", len(result.Removed))
	}
}

func TestUninstallFullCycle(t *testing.T) {
	tmpDir := t.TempDir()

	// Simulate a full ATV install by creating everything
	// Directories
	for _, dir := range atvDirectories {
		fullPath := filepath.Join(tmpDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(fullPath, "content.md"), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Files
	for _, file := range atvFiles {
		fullPath := filepath.Join(tmpDir, file)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Doc directories (empty)
	for _, dir := range atvDocDirectories {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	// .vscode/extensions.json (should NOT be removed — not ATV-owned)
	if err := os.MkdirAll(filepath.Join(tmpDir, ".vscode"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".vscode", "extensions.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	result := Uninstall(tmpDir, nil, false)

	expectedRemoved := len(atvDirectories) + len(atvFiles) + len(atvDocDirectories)
	if len(result.Removed) != expectedRemoved {
		t.Errorf("expected %d total removals, got %d: %v", expectedRemoved, len(result.Removed), result.Removed)
	}
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	// .vscode should still exist (not ATV-owned)
	if _, err := os.Stat(filepath.Join(tmpDir, ".vscode", "extensions.json")); os.IsNotExist(err) {
		t.Error(".vscode/extensions.json should NOT be removed by uninstall")
	}

	// .github should be cleaned up if empty
	if _, err := os.Stat(filepath.Join(tmpDir, ".github")); !os.IsNotExist(err) {
		// Check if it's actually empty
		entries, _ := os.ReadDir(filepath.Join(tmpDir, ".github"))
		if len(entries) == 0 {
			t.Error(".github should have been cleaned up when empty")
		}
	}
}

func TestUninstallSummary(t *testing.T) {
	result := UninstallResult{
		Removed: []string{"a", "b"},
		Skipped: []string{"c"},
		Missing: []string{"d", "e", "f"},
	}

	summary := result.Summary()
	if summary != "2 removed, 1 skipped (user-modified), 3 already absent" {
		t.Errorf("unexpected summary: %s", summary)
	}
}
