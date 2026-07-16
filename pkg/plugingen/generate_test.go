package plugingen

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot resolves to the top-level ATV-StarterKit checkout. Since
// this test file lives at pkg/plugingen/, the root is two levels up.
func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "..", ".."))
}

// regenerateInto runs Generate against a fresh tempdir mirror of the
// scaffold templates and returns the tempdir root.
func regenerateInto(t *testing.T) string {
	t.Helper()
	src := repoRoot(t)
	tmp := t.TempDir()
	if err := mirrorTree(filepath.Join(src, "pkg", "scaffold", "templates"),
		filepath.Join(tmp, "pkg", "scaffold", "templates")); err != nil {
		t.Fatalf("mirror templates: %v", err)
	}
	if err := Generate(Config{RepoRoot: tmp, KitVersion: "test-1.2.3"}); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return tmp
}

func TestGenerate_ProducesEveryPerSkillPlugin(t *testing.T) {
	tmp := regenerateInto(t)
	src := filepath.Join(repoRoot(t), "pkg", "scaffold", "templates", "skills")
	skillNames, err := listSkills(src)
	if err != nil {
		t.Fatalf("listSkills: %v", err)
	}
	if len(skillNames) == 0 {
		t.Fatal("no skills discovered — test setup is broken")
	}
	for _, name := range skillNames {
		dir := filepath.Join(tmp, "plugins", pluginNameForSkill(name))
		if _, err := os.Stat(filepath.Join(dir, "plugin.json")); err != nil {
			t.Errorf("missing plugin.json for %s: %v", pluginNameForSkill(name), err)
		}
		// Skill content directory inside the plugin keeps the original
		// template directory name (so the SKILL.md name: field which
		// drives the slash command stays intact).
		if _, err := os.Stat(filepath.Join(dir, "skills", name, "SKILL.md")); err != nil {
			t.Errorf("missing SKILL.md inside %s: %v", pluginNameForSkill(name), err)
		}
	}
}

// TestPluginNameForSkill_KebabCaseOnly is the regression guard for the
// 2.6.1 marketplace bug where `resolve_todo_parallel` produced an
// underscore-containing plugin name and the Copilot CLI rejected the
// whole marketplace.json with "Plugin name must be kebab-case".
func TestPluginNameForSkill_KebabCaseOnly(t *testing.T) {
	cases := map[string]string{
		"ce-plan":               "atv-skill-ce-plan",
		"resolve_todo_parallel": "atv-skill-resolve-todo-parallel",
		"atv-security":          "atv-skill-atv-security",
		"meme-iq":               "atv-skill-meme-iq",
	}
	for in, want := range cases {
		got := pluginNameForSkill(in)
		if got != want {
			t.Errorf("pluginNameForSkill(%q) = %q, want %q", in, got, want)
		}
		if strings.ContainsAny(got, "_ ./") {
			t.Errorf("plugin name %q contains forbidden character", got)
		}
	}
}

func TestGenerate_PacksContainExpectedSkills(t *testing.T) {
	tmp := regenerateInto(t)
	for _, p := range Packs() {
		dir := filepath.Join(tmp, "plugins", p.Name)
		for _, sn := range p.SkillNames {
			path := filepath.Join(dir, "skills", sn, "SKILL.md")
			if _, err := os.Stat(path); err != nil {
				t.Errorf("pack %s missing skill %s: %v", p.Name, sn, err)
			}
		}
		// Pack manifest should have category set so users can filter
		// with `copilot plugin marketplace browse`.
		var m PluginManifest
		readJSON(t, filepath.Join(dir, "plugin.json"), &m)
		if m.Category != p.Category {
			t.Errorf("pack %s category mismatch: got %q want %q", p.Name, m.Category, p.Category)
		}
		if len(m.Skills) != 1 || m.Skills[0] != "./skills" {
			t.Errorf("pack %s skills field should be [./skills], got %v", p.Name, m.Skills)
		}
	}
}

func TestGenerate_AtvEverythingBundlesAllSkillsAndAgents(t *testing.T) {
	tmp := regenerateInto(t)
	src := filepath.Join(repoRoot(t), "pkg", "scaffold", "templates")
	skillNames, _ := listSkills(filepath.Join(src, "skills"))
	agentFiles, _ := listAgents(filepath.Join(src, "agents"))

	for _, name := range skillNames {
		path := filepath.Join(tmp, "plugins", "atv-everything", "skills", name, "SKILL.md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("atv-everything missing skill %s: %v", name, err)
		}
	}
	for _, file := range agentFiles {
		path := filepath.Join(tmp, "plugins", "atv-everything", "agents", file)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("atv-everything missing agent %s: %v", file, err)
		}
	}

	var m PluginManifest
	readJSON(t, filepath.Join(tmp, "plugins", "atv-everything", "plugin.json"), &m)
	if len(m.Skills) == 0 || len(m.Agents) == 0 {
		t.Errorf("atv-everything manifest should declare both skills and agents, got %+v", m)
	}
}

func TestGenerate_AtvAgentsBundlesAllAgents(t *testing.T) {
	tmp := regenerateInto(t)
	src := filepath.Join(repoRoot(t), "pkg", "scaffold", "templates", "agents")
	agentFiles, _ := listAgents(src)

	for _, file := range agentFiles {
		path := filepath.Join(tmp, "plugins", "atv-agents", "agents", file)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("atv-agents missing %s: %v", file, err)
		}
	}
}

func TestGenerate_CliMarketplaceListsEveryPlugin(t *testing.T) {
	tmp := regenerateInto(t)
	var mp Marketplace
	readJSON(t, filepath.Join(tmp, ".github", "plugin", "marketplace.json"), &mp)

	if mp.Name != "atv-starter-kit" {
		t.Errorf("marketplace name: got %q want atv-starter-kit", mp.Name)
	}
	if mp.Metadata.PluginRoot != "./plugins" {
		t.Errorf("pluginRoot: got %q want ./plugins", mp.Metadata.PluginRoot)
	}

	have := map[string]bool{}
	for _, e := range mp.Plugins {
		have[e.Name] = true
		if e.Source != e.Name {
			t.Errorf("entry %s source should match name, got %q", e.Name, e.Source)
		}
	}

	src := filepath.Join(repoRoot(t), "pkg", "scaffold", "templates", "skills")
	skillNames, _ := listSkills(src)

	wantNames := []string{"atv-everything", "atv-agents"}
	for _, p := range Packs() {
		wantNames = append(wantNames, p.Name)
	}
	for _, n := range skillNames {
		wantNames = append(wantNames, pluginNameForSkill(n))
	}
	for _, n := range wantNames {
		if !have[n] {
			t.Errorf("marketplace.json missing entry %q", n)
		}
	}
	if len(mp.Plugins) != len(wantNames) {
		t.Errorf("marketplace.json plugin count: got %d want %d", len(mp.Plugins), len(wantNames))
	}

	if len(mp.Plugins) == 0 || mp.Plugins[0].Name != "atv-everything" {
		t.Fatalf("CLI marketplace should put atv-everything first, got %+v", mp.Plugins)
	}

	// After the flagship entry, entries remain sorted alphabetically for deterministic output.
	for i := 2; i < len(mp.Plugins); i++ {
		if mp.Plugins[i-1].Name > mp.Plugins[i].Name {
			t.Errorf("marketplace.json plugins not sorted: %q > %q",
				mp.Plugins[i-1].Name, mp.Plugins[i].Name)
		}
	}
}

func TestGenerate_SourceInstallMarketplaceHasOneFlagshipPlugin(t *testing.T) {
	tmp := regenerateInto(t)
	for _, rel := range []string{"marketplace.json", filepath.Join(".claude-plugin", "marketplace.json")} {
		assertSourceInstallMarketplaceHasOneFlagshipPlugin(t, filepath.Join(tmp, rel))
	}
	assertFilesEqual(t, filepath.Join(tmp, "marketplace.json"), filepath.Join(tmp, ".claude-plugin", "marketplace.json"))

	sourceDir := filepath.Join(tmp, "plugins", "atv-everything")
	for _, path := range []string{
		filepath.Join(sourceDir, "skills"),
		filepath.Join(sourceDir, "agents"),
		filepath.Join(sourceDir, ".claude-plugin", "plugin.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("source-install entry points at incomplete plugin dir, missing %s: %v", path, err)
		}
	}

	var manifest PluginManifest
	readJSON(t, filepath.Join(sourceDir, ".claude-plugin", "plugin.json"), &manifest)
	if manifest.Name != "atv-starter-kit" {
		t.Errorf("source-install plugin manifest name: got %q want atv-starter-kit", manifest.Name)
	}
}

func TestGenerate_GranularSkillDescriptionsMentionAgentCompanion(t *testing.T) {
	tmp := regenerateInto(t)
	var manifest PluginManifest
	readJSON(t, filepath.Join(tmp, "plugins", "atv-skill-ce-review", "plugin.json"), &manifest)

	if !strings.Contains(manifest.Description, "installs may need `atv-agents` alongside them") {
		t.Errorf("granular skill description should tell users when atv-agents is needed, got %q", manifest.Description)
	}
	if strings.Contains(manifest.Description, "includes the relevant agents") {
		t.Errorf("granular skill description should not imply category packs include agents, got %q", manifest.Description)
	}
}

func assertSourceInstallMarketplaceHasOneFlagshipPlugin(t *testing.T, path string) {
	t.Helper()
	var mp SourceInstallMarketplace
	readJSON(t, path, &mp)

	if mp.Name != "atv-starter-kit" {
		t.Errorf("source-install marketplace name: got %q want atv-starter-kit", mp.Name)
	}
	if len(mp.Plugins) != 1 {
		t.Fatalf("source-install marketplace should expose exactly one plugin, got %d", len(mp.Plugins))
	}

	entry := mp.Plugins[0]
	if entry.Name != "atv-starter-kit" {
		t.Errorf("source-install entry name: got %q want atv-starter-kit", entry.Name)
	}
	if entry.Source != "./plugins/atv-everything" {
		t.Errorf("source-install entry source: got %q want ./plugins/atv-everything", entry.Source)
	}
	if len(entry.Description) > 120 {
		t.Errorf("source-install description should fit in a picker row, got %d chars", len(entry.Description))
	}
	if strings.Contains(entry.Description, "http") {
		t.Errorf("source-install description should not include URLs, got %q", entry.Description)
	}
	for _, noisyPrefix := range []string{"atv-skill-", "atv-pack-"} {
		if strings.HasPrefix(entry.Name, noisyPrefix) {
			t.Errorf("source-install entry should be the flagship plugin, got %q", entry.Name)
		}
	}
	if entry.Name == "atv-agents" {
		t.Errorf("source-install entry should not expose agents-only plugin")
	}
}

func TestGenerate_MaintenanceSkillsCoverSourceAgentPlugins(t *testing.T) {
	tmp := regenerateInto(t)

	doctorSnippets := []string{
		"VS Code source-installed AgentPlugins",
		"hasSourceAgentPlugins",
		"owner/repo",
		"ahead/behind",
		"Never update, reset, delete, or reinstall VS Code AgentPlugin folders from `/atv-doctor`",
	}
	updateSnippets := []string{
		"VS Code source-installed AgentPlugins",
		"clean and behind-only",
		"Do not run `git reset`, `git stash`",
		"Reload or restart VS Code",
		"Never remove an entire `agent-plugins`",
	}

	for _, plugin := range []string{"atv-everything", "atv-pack-maintenance", "atv-skill-atv-doctor"} {
		path := filepath.Join(tmp, "plugins", plugin, "skills", "atv-doctor", "SKILL.md")
		assertFileContainsAll(t, path, doctorSnippets)
	}
	for _, plugin := range []string{"atv-everything", "atv-pack-maintenance", "atv-skill-atv-update"} {
		path := filepath.Join(tmp, "plugins", plugin, "skills", "atv-update", "SKILL.md")
		assertFileContainsAll(t, path, updateSnippets)
	}
}

func TestGenerate_DeterministicAcrossRuns(t *testing.T) {
	src := repoRoot(t)
	tmpA := t.TempDir()
	tmpB := t.TempDir()
	for _, dst := range []string{tmpA, tmpB} {
		if err := mirrorTree(filepath.Join(src, "pkg", "scaffold", "templates"),
			filepath.Join(dst, "pkg", "scaffold", "templates")); err != nil {
			t.Fatalf("mirror: %v", err)
		}
		if err := Generate(Config{RepoRoot: dst, KitVersion: "det-1.0.0"}); err != nil {
			t.Fatalf("generate: %v", err)
		}
	}
	for _, sub := range []string{"plugins", filepath.Join(".github", "plugin"), ".claude-plugin"} {
		diffs, err := compareTrees(filepath.Join(tmpA, sub), filepath.Join(tmpB, sub))
		if err != nil {
			t.Fatalf("compareTrees %s: %v", sub, err)
		}
		if len(diffs) > 0 {
			t.Errorf("non-deterministic generation in %s, diffs: %v", sub, diffs)
		}
	}
	assertFilesEqual(t, filepath.Join(tmpA, "marketplace.json"), filepath.Join(tmpB, "marketplace.json"))
}

func TestCompareGeneratedContentIgnoresLineEndingDifferences(t *testing.T) {
	gotRoot := t.TempDir()
	wantRoot := t.TempDir()
	gotPath := filepath.Join(gotRoot, "plugin", "README.md")
	wantPath := filepath.Join(wantRoot, "plugin", "README.md")
	if err := os.MkdirAll(filepath.Dir(gotPath), 0o755); err != nil {
		t.Fatalf("mkdir got: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(wantPath), 0o755); err != nil {
		t.Fatalf("mkdir want: %v", err)
	}
	if err := os.WriteFile(gotPath, []byte("# plugin\n\nbody\n"), 0o644); err != nil {
		t.Fatalf("write got: %v", err)
	}
	if err := os.WriteFile(wantPath, []byte("# plugin\r\n\r\nbody\r\n"), 0o644); err != nil {
		t.Fatalf("write want: %v", err)
	}

	diffs, err := compareTrees(gotRoot, wantRoot)
	if err != nil {
		t.Fatalf("compareTrees: %v", err)
	}
	if len(diffs) != 0 {
		t.Fatalf("line-ending-only differences should not be reported, got %v", diffs)
	}

	diffs, err = compareFile(gotPath, wantPath, "plugin/README.md")
	if err != nil {
		t.Fatalf("compareFile: %v", err)
	}
	if len(diffs) != 0 {
		t.Fatalf("line-ending-only file differences should not be reported, got %v", diffs)
	}
}

func TestCheckCleanReportsPrefixedCatalogDrift(t *testing.T) {
	tmp := regenerateInto(t)
	for _, rel := range []string{
		"marketplace.json",
		filepath.Join(".claude-plugin", "marketplace.json"),
		filepath.Join(".github", "plugin", "marketplace.json"),
	} {
		path := filepath.Join(tmp, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", rel, err)
		}
		if err := os.WriteFile(path, append(data, []byte("\n")...), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}

	err := CheckClean(Config{RepoRoot: tmp, KitVersion: "test-1.2.3"})
	if err == nil {
		t.Fatal("expected CheckClean to report drift")
	}
	message := err.Error()
	for _, want := range []string{
		"content differs: marketplace.json",
		"content differs: .claude-plugin/marketplace.json",
		"content differs: .github/plugin/marketplace.json",
	} {
		if !strings.Contains(message, want) {
			t.Errorf("CheckClean error missing %q:\n%s", want, message)
		}
	}
}

func TestGenerate_FailsWhenSkillNotInPackOrMisc(t *testing.T) {
	src := repoRoot(t)
	tmp := t.TempDir()
	if err := mirrorTree(filepath.Join(src, "pkg", "scaffold", "templates"),
		filepath.Join(tmp, "pkg", "scaffold", "templates")); err != nil {
		t.Fatalf("mirror: %v", err)
	}
	// Create a fake new skill that's not in any pack or in MiscSkills.
	stray := filepath.Join(tmp, "pkg", "scaffold", "templates", "skills", "zz-stray-skill-for-test")
	if err := os.MkdirAll(stray, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(stray, "SKILL.md"),
		[]byte("---\nname: zz-stray\n---\nbody\n"), 0o644); err != nil {
		t.Fatalf("write skill: %v", err)
	}
	err := Generate(Config{RepoRoot: tmp, KitVersion: "stray-test"})
	if err == nil {
		t.Fatal("expected error for unassigned skill, got nil")
	}
	if !strings.Contains(err.Error(), "zz-stray-skill-for-test") {
		t.Errorf("error should mention the stray skill: %v", err)
	}
}

func TestAudit_FlagsNonAllowListedFindings(t *testing.T) {
	contents := map[string]string{
		"skill:test-bad": "this references mcp__plugin_evil__do_thing in a bad way",
	}
	findings := Audit(contents)
	if len(findings) == 0 {
		t.Fatal("expected at least one finding for evil mcp tool")
	}
}

func TestAudit_PassesAllowListedFindings(t *testing.T) {
	// Read the actual ce-plan body which has compound-engineering refs
	// allow-listed. Audit should return nothing for it.
	body, err := os.ReadFile(filepath.Join(repoRoot(t), "pkg", "scaffold", "templates", "skills", "ce-plan", "SKILL.md"))
	if err != nil {
		t.Fatalf("read ce-plan: %v", err)
	}
	contents := map[string]string{"skill:ce-plan": string(body)}
	findings := Audit(contents)
	if len(findings) != 0 {
		t.Errorf("ce-plan should be fully allow-listed, got findings: %v", findings)
	}
}

// readJSON loads a generated JSON file and decodes it into v.
func readJSON(t *testing.T, path string, v interface{}) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(v); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}

func assertFileContainsAll(t *testing.T, path string, snippets []string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	body := string(data)
	for _, snippet := range snippets {
		if !strings.Contains(body, snippet) {
			t.Errorf("%s missing snippet %q", path, snippet)
		}
	}
}

func assertFilesEqual(t *testing.T, leftPath, rightPath string) {
	t.Helper()
	left, err := os.ReadFile(leftPath)
	if err != nil {
		t.Fatalf("read %s: %v", leftPath, err)
	}
	right, err := os.ReadFile(rightPath)
	if err != nil {
		t.Fatalf("read %s: %v", rightPath, err)
	}
	if !bytes.Equal(left, right) {
		t.Errorf("files differ: %s and %s", leftPath, rightPath)
	}
}
