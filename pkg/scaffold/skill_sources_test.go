package scaffold

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

type skillSourceManifest struct {
	Version            int                `json:"version"`
	CanonicalSkillRoot string             `json:"canonicalSkillRoot"`
	DogfoodSkillRoot   string             `json:"dogfoodSkillRoot"`
	Skills             []skillSourceEntry `json:"skills"`
	DogfoodOnlySkills  []string           `json:"dogfoodOnlySkills"`
}

type skillSourceEntry struct {
	Name     string `json:"name"`
	Origin   string `json:"origin"`
	Tracking string `json:"tracking"`
	Upstream string `json:"upstream,omitempty"`
	Dogfood  string `json:"dogfood"`
}

func TestSkillSourceManifestCoversTemplateSkills(t *testing.T) {
	manifest := readSkillSourceManifest(t)
	if manifest.CanonicalSkillRoot != "pkg/scaffold/templates/skills" {
		t.Fatalf("canonicalSkillRoot = %q, want pkg/scaffold/templates/skills", manifest.CanonicalSkillRoot)
	}

	templateDirs := readEmbeddedSkillDirs(t)
	entries := make(map[string]skillSourceEntry, len(manifest.Skills))
	for _, entry := range manifest.Skills {
		if entry.Name == "" || entry.Origin == "" || entry.Tracking == "" || entry.Dogfood == "" {
			t.Fatalf("skill source entry must include name, origin, tracking, and dogfood: %+v", entry)
		}
		if existing, ok := entries[entry.Name]; ok {
			t.Fatalf("duplicate skill source entry for %q: %+v and %+v", entry.Name, existing, entry)
		}
		entries[entry.Name] = entry
		if entry.Origin == "compound-engineering" && entry.Upstream == "" {
			t.Fatalf("compound-engineering skill %q must declare upstream name", entry.Name)
		}
	}

	var missing []string
	for _, name := range templateDirs {
		if _, ok := entries[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("skill templates missing from pkg/scaffold/skill_sources.json: %v", missing)
	}

	templateSet := make(map[string]bool, len(templateDirs))
	for _, name := range templateDirs {
		templateSet[name] = true
	}
	var stale []string
	for name := range entries {
		if !templateSet[name] {
			stale = append(stale, name)
		}
	}
	if len(stale) > 0 {
		sort.Strings(stale)
		t.Fatalf("skill_sources.json contains entries missing from pkg/scaffold/templates/skills: %v", stale)
	}
}

func TestSkillSourceManifestClassifiesDogfoodSkills(t *testing.T) {
	manifest := readSkillSourceManifest(t)
	repoRoot := repoRoot(t)
	dogfoodRoot := filepath.Join(repoRoot, manifest.DogfoodSkillRoot)

	dogfoodDirs := readDiskSkillDirs(t, dogfoodRoot)
	dogfoodSet := make(map[string]bool, len(dogfoodDirs))
	for _, name := range dogfoodDirs {
		dogfoodSet[name] = true
	}

	templateSet := make(map[string]skillSourceEntry, len(manifest.Skills))
	for _, entry := range manifest.Skills {
		templateSet[entry.Name] = entry
		classifyDogfoodEntry(t, repoRoot, manifest, entry, dogfoodSet[entry.Name])
	}

	dogfoodOnly := make(map[string]bool, len(manifest.DogfoodOnlySkills))
	for _, name := range manifest.DogfoodOnlySkills {
		if templateSet[name].Name != "" {
			t.Fatalf("dogfoodOnlySkills contains %q, but it also exists in canonical templates", name)
		}
		if !dogfoodSet[name] {
			t.Fatalf("dogfoodOnlySkills contains %q, but .github/skills/%s does not exist", name, name)
		}
		dogfoodOnly[name] = true
	}

	var unclassified []string
	for name := range dogfoodSet {
		if templateSet[name].Name != "" || dogfoodOnly[name] {
			continue
		}
		unclassified = append(unclassified, name)
	}
	if len(unclassified) > 0 {
		sort.Strings(unclassified)
		t.Fatalf(".github/skills entries missing from skill_sources.json dogfoodOnlySkills: %v", unclassified)
	}
}

func classifyDogfoodEntry(t *testing.T, repoRoot string, manifest skillSourceManifest, entry skillSourceEntry, dogfoodExists bool) {
	t.Helper()
	switch entry.Dogfood {
	case "mirror":
		if !dogfoodExists {
			t.Fatalf("skill %q is classified dogfood=mirror but .github/skills/%s is missing", entry.Name, entry.Name)
		}
		templatePath := filepath.Join(repoRoot, manifest.CanonicalSkillRoot, entry.Name, "SKILL.md")
		dogfoodPath := filepath.Join(repoRoot, manifest.DogfoodSkillRoot, entry.Name, "SKILL.md")
		if !filesEqualAfterLineEndingNormalization(t, templatePath, dogfoodPath) {
			t.Fatalf("skill %q is classified dogfood=mirror but canonical template and dogfood copy differ", entry.Name)
		}
	case "divergent":
		if !dogfoodExists {
			t.Fatalf("skill %q is classified dogfood=divergent but .github/skills/%s is missing", entry.Name, entry.Name)
		}
	case "omitted":
		if dogfoodExists {
			t.Fatalf("skill %q is classified dogfood=omitted but .github/skills/%s exists", entry.Name, entry.Name)
		}
	default:
		t.Fatalf("skill %q has unknown dogfood classification %q", entry.Name, entry.Dogfood)
	}
}

func readSkillSourceManifest(t *testing.T) skillSourceManifest {
	t.Helper()
	path := filepath.Join(repoRoot(t), "pkg", "scaffold", "skill_sources.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var manifest skillSourceManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	if manifest.Version != 1 {
		t.Fatalf("skill source manifest version = %d, want 1", manifest.Version)
	}
	return manifest
}

func readDiskSkillDirs(t *testing.T, root string) []string {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("read %s: %v", root, err)
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)
	return dirs
}

func filesEqualAfterLineEndingNormalization(t *testing.T, leftPath, rightPath string) bool {
	t.Helper()
	left, err := os.ReadFile(leftPath)
	if err != nil {
		t.Fatalf("read %s: %v", leftPath, err)
	}
	right, err := os.ReadFile(rightPath)
	if err != nil {
		t.Fatalf("read %s: %v", rightPath, err)
	}
	left = bytes.ReplaceAll(left, []byte("\r\n"), []byte("\n"))
	right = bytes.ReplaceAll(right, []byte("\r\n"), []byte("\n"))
	return bytes.Equal(left, right)
}
