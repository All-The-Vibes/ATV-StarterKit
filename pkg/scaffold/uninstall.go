package scaffold

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UninstallResult reports what happened during uninstallation.
type UninstallResult struct {
	Removed  []string // paths that were removed
	Skipped  []string // paths skipped (modified by user)
	Missing  []string // paths that didn't exist
	Errors   []string // paths that failed to remove
}

// Summary returns a human-readable summary.
func (r UninstallResult) Summary() string {
	parts := make([]string, 0, 4)
	if len(r.Removed) > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", len(r.Removed)))
	}
	if len(r.Skipped) > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped (user-modified)", len(r.Skipped)))
	}
	if len(r.Missing) > 0 {
		parts = append(parts, fmt.Sprintf("%d already absent", len(r.Missing)))
	}
	if len(r.Errors) > 0 {
		parts = append(parts, fmt.Sprintf("%d errors", len(r.Errors)))
	}
	if len(parts) == 0 {
		return "nothing to uninstall"
	}
	return strings.Join(parts, ", ")
}

// atvDirectories are the top-level directories that ATV owns entirely.
// These are safe to remove recursively without file-level checks.
var atvDirectories = []string{
	".github/skills",
	".github/agents",
	".github/hooks",
	".gstack",
	".atv",
}

// atvFiles are individual files that ATV creates.
// These need checksum verification before removal (user may have customized them).
var atvFiles = []string{
	".github/copilot-instructions.md",
	".github/copilot-setup-steps.yml",
	".github/copilot-mcp-config.json",
	".github/rails.instructions.md",
	".github/python.instructions.md",
	".github/typescript.instructions.md",
}

// atvDocDirectories are documentation directories ATV creates.
// Only removed if empty (user may have put real docs there).
var atvDocDirectories = []string{
	"docs/plans",
	"docs/brainstorms",
	"docs/solutions",
}

// Uninstall removes all ATV-installed components from the target directory.
// If checksums are provided (from the install manifest), files are only removed
// if they haven't been modified. Pass nil to force-remove everything.
func Uninstall(targetDir string, checksums map[string]string, force bool) UninstallResult {
	var result UninstallResult

	// 1. Remove ATV-owned directories (always safe — these are entirely ours)
	for _, dir := range atvDirectories {
		fullPath := filepath.Join(targetDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			result.Missing = append(result.Missing, dir)
			continue
		}
		if err := os.RemoveAll(fullPath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", dir, err))
		} else {
			result.Removed = append(result.Removed, dir)
		}
	}

	// 2. Remove individual ATV files (check checksums unless force)
	for _, file := range atvFiles {
		fullPath := filepath.Join(targetDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			result.Missing = append(result.Missing, file)
			continue
		}

		if !force && checksums != nil {
			if originalChecksum, ok := checksums[file]; ok {
				currentChecksum := fileChecksum(fullPath)
				if currentChecksum != "" && currentChecksum != originalChecksum {
					result.Skipped = append(result.Skipped, file)
					continue
				}
			}
		}

		if err := os.Remove(fullPath); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", file, err))
		} else {
			result.Removed = append(result.Removed, file)
		}
	}

	// 3. Remove empty doc directories (don't remove if user added real content)
	for _, dir := range atvDocDirectories {
		fullPath := filepath.Join(targetDir, dir)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			result.Missing = append(result.Missing, dir)
			continue
		}
		if isDirEmpty(fullPath) || force {
			if err := os.RemoveAll(fullPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", dir, err))
			} else {
				result.Removed = append(result.Removed, dir)
			}
		} else {
			result.Skipped = append(result.Skipped, dir+" (has user content)")
		}
	}

	// 4. Clean up empty parent directories
	cleanEmptyDir(filepath.Join(targetDir, "docs"))
	cleanEmptyDir(filepath.Join(targetDir, ".github"))

	return result
}

func fileChecksum(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func isDirEmpty(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	return len(entries) == 0
}

func cleanEmptyDir(path string) {
	if isDirEmpty(path) {
		os.Remove(path)
	}
}
