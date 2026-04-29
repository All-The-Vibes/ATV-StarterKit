package skillsync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCompareReportsTrackedSkillStatuses(t *testing.T) {
	root := t.TempDir()
	manifestPath := writeManifest(t, root, Manifest{
		Version:            1,
		CanonicalSkillRoot: "local",
		Skills: []SkillSource{
			{Name: "ce-plan", Origin: "compound-engineering", Tracking: "upstream", Upstream: "ce-plan", Dogfood: "mirror"},
			{Name: "ce-review", Origin: "compound-engineering", Tracking: "alias", Upstream: "ce-code-review", Dogfood: "mirror"},
			{Name: "lfg", Origin: "compound-engineering", Tracking: "overlay", Upstream: "lfg", Dogfood: "divergent"},
			{Name: "atv-doctor", Origin: "atv", Tracking: "native", Dogfood: "omitted"},
		},
	})
	localRoot := filepath.Join(root, "local")
	upstreamRoot := filepath.Join(root, "upstream")
	writeSkill(t, localRoot, "ce-plan", "same\n")
	writeSkill(t, upstreamRoot, "ce-plan", "same\r\n")
	writeSkill(t, localRoot, "ce-review", "old\n")
	writeSkill(t, upstreamRoot, "ce-code-review", "new\n")
	writeSkill(t, localRoot, "lfg", "atv overlay\n")
	writeSkill(t, upstreamRoot, "lfg", "upstream\n")
	writeSkill(t, upstreamRoot, "ce-work", "untracked\n")

	results, err := Compare(manifestPath, localRoot, upstreamRoot)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	statuses := map[string]string{}
	for _, result := range results {
		key := result.LocalName
		if key == "" {
			key = result.UpstreamName
		}
		statuses[key] = result.Status
	}
	for name, want := range map[string]string{
		"ce-plan":   StatusInSync,
		"ce-review": StatusStale,
		"lfg":       StatusOverlay,
		"ce-work":   StatusUntrackedUpstream,
	} {
		if statuses[name] != want {
			t.Errorf("status for %s = %q, want %q", name, statuses[name], want)
		}
	}
	if _, ok := statuses["atv-doctor"]; ok {
		t.Errorf("ATV-native skills should not appear in CE comparison results")
	}
}

func TestCompareReportsMissingSkillFiles(t *testing.T) {
	root := t.TempDir()
	manifestPath := writeManifest(t, root, Manifest{
		Version:            1,
		CanonicalSkillRoot: "local",
		Skills: []SkillSource{
			{Name: "ce-plan", Origin: "compound-engineering", Tracking: "upstream", Upstream: "ce-plan", Dogfood: "mirror"},
			{Name: "ce-work", Origin: "compound-engineering", Tracking: "upstream", Upstream: "ce-work", Dogfood: "mirror"},
		},
	})
	localRoot := filepath.Join(root, "local")
	upstreamRoot := filepath.Join(root, "upstream")
	writeSkill(t, localRoot, "ce-work", "local\n")
	writeSkill(t, upstreamRoot, "ce-plan", "upstream\n")

	results, err := Compare(manifestPath, localRoot, upstreamRoot)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	statuses := map[string]string{}
	for _, result := range results {
		statuses[result.LocalName] = result.Status
	}
	if statuses["ce-plan"] != StatusMissingLocal {
		t.Errorf("ce-plan status = %q, want %q", statuses["ce-plan"], StatusMissingLocal)
	}
	if statuses["ce-work"] != StatusMissingUpstream {
		t.Errorf("ce-work status = %q, want %q", statuses["ce-work"], StatusMissingUpstream)
	}
}

func writeManifest(t *testing.T, root string, manifest Manifest) string {
	t.Helper()
	path := filepath.Join(root, "skill_sources.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

func writeSkill(t *testing.T, root, name, body string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
}
