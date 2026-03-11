package scaffold

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteStatus indicates what happened when writing a file.
type WriteStatus int

const (
	StatusCreated WriteStatus = iota
	StatusSkipped
	StatusMerged
	StatusDirCreated
)

// WriteResult records what happened with a single file.
type WriteResult struct {
	Path   string
	Status WriteStatus
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
		os.MkdirAll(parentDir, 0755)

		if comp.MergeJSON {
			result := writeOrMergeJSON(destPath, comp.Path, comp.Content)
			results = append(results, result)
			continue
		}

		result := writeIfNotExists(destPath, comp.Path, comp.Content)
		results = append(results, result)
	}

	return results
}

func ensureDir(fullPath, relPath string) WriteResult {
	if _, err := os.Stat(fullPath); err == nil {
		return WriteResult{Path: relPath, Status: StatusSkipped}
	}
	os.MkdirAll(fullPath, 0755)
	return WriteResult{Path: relPath, Status: StatusDirCreated}
}

func writeIfNotExists(fullPath, relPath string, content []byte) WriteResult {
	if _, err := os.Stat(fullPath); err == nil {
		return WriteResult{Path: relPath, Status: StatusSkipped}
	}
	if err := os.WriteFile(fullPath, content, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "  ❌ Failed to write %s: %v\n", relPath, err)
		return WriteResult{Path: relPath, Status: StatusSkipped}
	}
	return WriteResult{Path: relPath, Status: StatusCreated}
}

func writeOrMergeJSON(fullPath, relPath string, newContent []byte) WriteResult {
	existingData, err := os.ReadFile(fullPath)
	if err != nil {
		// File doesn't exist — write new
		if err := os.WriteFile(fullPath, newContent, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "  ❌ Failed to write %s: %v\n", relPath, err)
			return WriteResult{Path: relPath, Status: StatusSkipped}
		}
		return WriteResult{Path: relPath, Status: StatusCreated}
	}

	// File exists — merge JSON objects
	var existing, incoming map[string]interface{}
	if err := json.Unmarshal(existingData, &existing); err != nil {
		return WriteResult{Path: relPath, Status: StatusSkipped}
	}
	if err := json.Unmarshal(newContent, &incoming); err != nil {
		return WriteResult{Path: relPath, Status: StatusSkipped}
	}

	merged := mergeJSONMaps(existing, incoming)
	mergedBytes, err := json.MarshalIndent(merged, "", "  ")
	if err != nil {
		return WriteResult{Path: relPath, Status: StatusSkipped}
	}

	if err := os.WriteFile(fullPath, append(mergedBytes, '\n'), 0644); err != nil {
		return WriteResult{Path: relPath, Status: StatusSkipped}
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
