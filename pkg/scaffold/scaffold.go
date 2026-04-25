package scaffold

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/installstate"
)

// WriteStatus indicates what happened when writing a file.
type WriteStatus int

const (
	StatusCreated WriteStatus = iota
	StatusSkipped
	StatusMerged
	StatusDirCreated
	StatusFailed
)

// WriteResult records what happened with a single file.
type WriteResult struct {
	Path   string
	Status WriteStatus
	Error  string
}

// WriteSummary is a coarse aggregate of scaffold write results.
type WriteSummary struct {
	Created     int
	Skipped     int
	Merged      int
	Directories int
	Failed      int
}

// WriteAll writes all catalog components to the target directory.
func WriteAll(targetDir string, catalog []Component) []WriteResult {
	var results []WriteResult

	for _, comp := range catalog {
		destPath := filepath.Join(targetDir, comp.Path)

		if comp.IsDir {
			result := ensureDir(destPath, comp.Path)
			results = append(results, result)
			continue
		}

		// Ensure parent directory exists
		parentDir := filepath.Dir(destPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "  ❌ Failed to create parent dir for %s: %v\n", comp.Path, err)
			results = append(results, WriteResult{Path: comp.Path, Status: StatusSkipped})
			continue
		}

		if comp.MergeJSON {
			result := writeOrMergeJSON(destPath, comp.Path, comp.Content)
			results = append(results, result)
			continue
		}

		mode := os.FileMode(0644)
		if isExecutableScript(comp.Path) {
			mode = 0755
		}
		result := writeIfNotExists(destPath, comp.Path, comp.Content, mode)
		results = append(results, result)
	}

	return results
}

// isExecutableScript returns true for paths whose embedded content is meant
// to run as an executable (shell or python script under a skill's scripts/
// directory). go:embed strips Unix mode bits, so the install pipeline must
// re-apply 0755 for these files explicitly.
func isExecutableScript(relPath string) bool {
	switch filepath.Ext(relPath) {
	case ".sh", ".py", ".rb", ".bash", ".zsh":
		return true
	}
	return false
}

// SummarizeResults returns aggregate counts for a scaffold write batch.
func SummarizeResults(results []WriteResult) WriteSummary {
	var summary WriteSummary
	for _, result := range results {
		switch result.Status {
		case StatusCreated:
			summary.Created++
		case StatusSkipped:
			summary.Skipped++
		case StatusMerged:
			summary.Merged++
		case StatusDirCreated:
			summary.Directories++
		case StatusFailed:
			summary.Failed++
		}
	}
	return summary
}

// Detail returns a human-readable summary of the write batch.
func (s WriteSummary) Detail() string {
	parts := make([]string, 0, 5)
	if s.Created > 0 {
		parts = append(parts, fmt.Sprintf("%d files created", s.Created))
	}
	if s.Directories > 0 {
		parts = append(parts, fmt.Sprintf("%d directories created", s.Directories))
	}
	if s.Merged > 0 {
		parts = append(parts, fmt.Sprintf("%d JSON configs merged", s.Merged))
	}
	if s.Skipped > 0 {
		parts = append(parts, fmt.Sprintf("%d existing paths skipped", s.Skipped))
	}
	if s.Failed > 0 {
		parts = append(parts, fmt.Sprintf("%d writes failed", s.Failed))
	}
	if len(parts) == 0 {
		return "no file changes"
	}
	return strings.Join(parts, ", ")
}

// FailureReason returns a short failure summary when writes fail.
func (s WriteSummary) FailureReason() string {
	if s.Failed == 0 {
		return ""
	}
	return fmt.Sprintf("%d scaffold writes failed", s.Failed)
}

// Successful reports whether no write operations failed.
func (s WriteSummary) Successful() bool {
	return s.Failed == 0
}

func ensureDir(fullPath, relPath string) WriteResult {
	if _, err := os.Stat(fullPath); err == nil {
		return WriteResult{Path: relPath, Status: StatusSkipped}
	}
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Failed to create dir %s: %v\n", relPath, err)
		return WriteResult{Path: relPath, Status: StatusFailed, Error: err.Error()}
	}
	return WriteResult{Path: relPath, Status: StatusDirCreated}
}

func writeIfNotExists(fullPath, relPath string, content []byte, mode os.FileMode) WriteResult {
	if _, err := os.Stat(fullPath); err == nil {
		return WriteResult{Path: relPath, Status: StatusSkipped}
	}
	if err := os.WriteFile(fullPath, content, mode); err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Failed to write %s: %v\n", relPath, err)
		return WriteResult{Path: relPath, Status: StatusFailed, Error: err.Error()}
	}
	return WriteResult{Path: relPath, Status: StatusCreated}
}

func writeOrMergeJSON(fullPath, relPath string, newContent []byte) WriteResult {
	existingData, err := os.ReadFile(fullPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return WriteResult{Path: relPath, Status: StatusFailed, Error: err.Error()}
		}
		// File doesn't exist — write new
		if err := os.WriteFile(fullPath, newContent, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  ❌ Failed to write %s: %v\n", relPath, err)
			return WriteResult{Path: relPath, Status: StatusFailed, Error: err.Error()}
		}
		return WriteResult{Path: relPath, Status: StatusCreated}
	}

	// File exists — merge JSON objects
	var existing, incoming map[string]interface{}
	if err := json.Unmarshal(existingData, &existing); err != nil {
		return WriteResult{Path: relPath, Status: StatusFailed, Error: err.Error()}
	}
	if err := json.Unmarshal(newContent, &incoming); err != nil {
		return WriteResult{Path: relPath, Status: StatusFailed, Error: err.Error()}
	}

	merged := mergeJSONMaps(existing, incoming)
	mergedBytes, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return WriteResult{Path: relPath, Status: StatusFailed, Error: err.Error()}
	}

	if err := os.WriteFile(fullPath, append(mergedBytes, '\n'), 0644); err != nil {
		return WriteResult{Path: relPath, Status: StatusFailed, Error: err.Error()}
	}
	return WriteResult{Path: relPath, Status: StatusMerged}
}

// mergeJSONMaps does a shallow merge: adds keys from src that don't exist in dst.
func mergeJSONMaps(dst, src map[string]interface{}) map[string]interface{} {
	for key, srcVal := range src {
		if _, exists := dst[key]; !exists {
			dst[key] = srcVal
		} else {
			// For nested maps (like mcpServers), merge one level deeper
			dstMap, dstOk := dst[key].(map[string]interface{})
			srcMap, srcOk := srcVal.(map[string]interface{})
			if dstOk && srcOk {
				dst[key] = mergeJSONMaps(dstMap, srcMap)
			}
			// For arrays (like recommendations), append missing values
			dstArr, dstOk := dst[key].([]interface{})
			srcArr, srcOk := srcVal.([]interface{})
			if dstOk && srcOk {
				dst[key] = mergeArrays(dstArr, srcArr)
			}
		}
	}
	return dst
}

func mergeArrays(dst, src []interface{}) []interface{} {
	existing := make(map[string]bool)
	for _, v := range dst {
		if s, ok := v.(string); ok {
			existing[s] = true
		}
	}
	for _, v := range src {
		if s, ok := v.(string); ok {
			if !existing[s] {
				dst = append(dst, v)
			}
		}
	}
	return dst
}

// ResultsToSubsteps converts scaffold write results into structured install substeps.
func ResultsToSubsteps(results []WriteResult) []installstate.InstallOutcome {
	substeps := make([]installstate.InstallOutcome, 0, len(results))
	for _, r := range results {
		substep := installstate.InstallOutcome{Step: r.Path}
		switch r.Status {
		case StatusCreated, StatusDirCreated:
			substep.Status = installstate.InstallStepDone
			substep.Detail = "created"
		case StatusMerged:
			substep.Status = installstate.InstallStepDone
			substep.Detail = "merged"
		case StatusSkipped:
			substep.Status = installstate.InstallStepSkipped
			substep.Detail = "exists"
			substep.SkipReason = installstate.SkipReasonAlreadyInstalled
		case StatusFailed:
			substep.Status = installstate.InstallStepFailed
			substep.Reason = r.Error
		}
		substeps = append(substeps, substep)
	}
	return substeps
}
