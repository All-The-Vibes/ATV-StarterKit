package skillsync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const (
	StatusInSync            = "in_sync"
	StatusStale             = "stale"
	StatusOverlay           = "atv_overlay"
	StatusMissingLocal      = "missing_local"
	StatusMissingUpstream   = "missing_upstream"
	StatusUntrackedUpstream = "untracked_upstream"
)

type Manifest struct {
	Version            int           `json:"version"`
	CanonicalSkillRoot string        `json:"canonicalSkillRoot"`
	DogfoodSkillRoot   string        `json:"dogfoodSkillRoot"`
	Skills             []SkillSource `json:"skills"`
	DogfoodOnlySkills  []string      `json:"dogfoodOnlySkills"`
}

type SkillSource struct {
	Name     string `json:"name"`
	Origin   string `json:"origin"`
	Tracking string `json:"tracking"`
	Upstream string `json:"upstream,omitempty"`
	Dogfood  string `json:"dogfood"`
}

type Result struct {
	LocalName    string `json:"localName,omitempty"`
	UpstreamName string `json:"upstreamName,omitempty"`
	Origin       string `json:"origin,omitempty"`
	Tracking     string `json:"tracking,omitempty"`
	Status       string `json:"status"`
	LocalPath    string `json:"localPath,omitempty"`
	UpstreamPath string `json:"upstreamPath,omitempty"`
}

func Compare(manifestPath, localSkillsRoot, upstreamSkillsRoot string) ([]Result, error) {
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	if localSkillsRoot == "" {
		return nil, fmt.Errorf("local skills root is required")
	}
	if upstreamSkillsRoot == "" {
		return nil, fmt.Errorf("upstream skills root is required")
	}
	if err := requireDir(localSkillsRoot); err != nil {
		return nil, fmt.Errorf("local skills root: %w", err)
	}
	if err := requireDir(upstreamSkillsRoot); err != nil {
		return nil, fmt.Errorf("upstream skills root: %w", err)
	}

	trackedUpstream := make(map[string]bool)
	results := make([]Result, 0, len(manifest.Skills))
	for _, source := range manifest.Skills {
		if source.Origin != "compound-engineering" {
			continue
		}
		upstreamName := source.Upstream
		if upstreamName == "" {
			upstreamName = source.Name
		}
		trackedUpstream[upstreamName] = true
		result := Result{
			LocalName:    source.Name,
			UpstreamName: upstreamName,
			Origin:       source.Origin,
			Tracking:     source.Tracking,
			LocalPath:    filepath.ToSlash(filepath.Join(localSkillsRoot, source.Name, "SKILL.md")),
			UpstreamPath: filepath.ToSlash(filepath.Join(upstreamSkillsRoot, upstreamName, "SKILL.md")),
		}

		localBody, localErr := readSkillBody(localSkillsRoot, source.Name)
		upstreamBody, upstreamErr := readSkillBody(upstreamSkillsRoot, upstreamName)
		switch {
		case os.IsNotExist(localErr):
			result.Status = StatusMissingLocal
		case localErr != nil:
			return nil, fmt.Errorf("read local skill %s: %w", source.Name, localErr)
		case os.IsNotExist(upstreamErr):
			result.Status = StatusMissingUpstream
		case upstreamErr != nil:
			return nil, fmt.Errorf("read upstream skill %s: %w", upstreamName, upstreamErr)
		case source.Tracking == "overlay":
			result.Status = StatusOverlay
		case bytes.Equal(normalizeLineEndings(localBody), normalizeLineEndings(upstreamBody)):
			result.Status = StatusInSync
		default:
			result.Status = StatusStale
		}
		results = append(results, result)
	}

	upstreamNames, err := listSkillDirs(upstreamSkillsRoot)
	if err != nil {
		return nil, err
	}
	for _, upstreamName := range upstreamNames {
		if trackedUpstream[upstreamName] {
			continue
		}
		results = append(results, Result{
			UpstreamName: upstreamName,
			Status:       StatusUntrackedUpstream,
			UpstreamPath: filepath.ToSlash(filepath.Join(upstreamSkillsRoot, upstreamName, "SKILL.md")),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		left := results[i].LocalName
		if left == "" {
			left = results[i].UpstreamName
		}
		right := results[j].LocalName
		if right == "" {
			right = results[j].UpstreamName
		}
		return left < right
	})
	return results, nil
}

func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	if manifest.Version != 1 {
		return Manifest{}, fmt.Errorf("unsupported skill source manifest version %d", manifest.Version)
	}
	return manifest, nil
}

func readSkillBody(root, name string) ([]byte, error) {
	return os.ReadFile(filepath.Join(root, name, "SKILL.md"))
}

func requireDir(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	return nil
}

func listSkillDirs(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

func normalizeLineEndings(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
}
