package plugingen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// pluginNameForSkill converts a template skill directory name into a
// valid Copilot CLI plugin name. Plugin names must be kebab-case
// (letters, numbers, hyphens only) per the spec, but our template
// directories occasionally use underscores (e.g. resolve_todo_parallel).
//
// The skill's slash command and SKILL.md `name:` field are unaffected
// — they continue to use the underscore form. Only the marketplace-
// facing plugin name and plugin directory name get sanitized.
func pluginNameForSkill(skill string) string {
	return "atv-skill-" + strings.ReplaceAll(skill, "_", "-")
}

// KitVersion is the marketplace + per-plugin version. It is set by the
// CLI from the repo VERSION file at runtime.
//
// All ATV plugins share this version (they are released as a single
// kit). Users cannot pin individual plugin versions via the
// `copilot plugin install` command, so per-plugin semver would be
// metadata only.
type Config struct {
	// RepoRoot is the absolute path to the repository root.
	RepoRoot string

	// KitVersion is the version string written into every manifest.
	KitVersion string
}

// Generate produces the plugins/ tree, marketplace.json,
// .github/plugin/marketplace.json, and .claude-plugin/marketplace.json
// from pkg/scaffold/templates/.
// It always writes a deterministic byte sequence: sorted lists,
// slash-normalized paths, LF line endings,
// indented JSON with a trailing newline.
//
// Generate first runs Audit and returns an error containing every
// finding if any non-allow-listed marketplace-incompatible content is
// detected. This forces the operator to either fix the template or
// extend the allow-list with a documented reason.
func Generate(cfg Config) error {
	if cfg.RepoRoot == "" {
		return fmt.Errorf("plugingen: RepoRoot is required")
	}
	if cfg.KitVersion == "" {
		return fmt.Errorf("plugingen: KitVersion is required")
	}

	templatesDir := filepath.Join(cfg.RepoRoot, "pkg", "scaffold", "templates")
	skillsDir := filepath.Join(templatesDir, "skills")
	agentsDir := filepath.Join(templatesDir, "agents")

	skillNames, err := listSkills(skillsDir)
	if err != nil {
		return fmt.Errorf("list skills: %w", err)
	}
	agentFiles, err := listAgents(agentsDir)
	if err != nil {
		return fmt.Errorf("list agents: %w", err)
	}

	// Load every skill SKILL.md and agent .agent.md into memory once.
	// We need the contents for both audit and emission.
	skillBody := make(map[string]string, len(skillNames))
	for _, name := range skillNames {
		body, err := os.ReadFile(filepath.Join(skillsDir, name, "SKILL.md"))
		if err != nil {
			return fmt.Errorf("read skill %s: %w", name, err)
		}
		skillBody[name] = normalizeLineEndings(string(body))
	}
	agentBody := make(map[string]string, len(agentFiles))
	for _, file := range agentFiles {
		body, err := os.ReadFile(filepath.Join(agentsDir, file))
		if err != nil {
			return fmt.Errorf("read agent %s: %w", file, err)
		}
		agentBody[file] = normalizeLineEndings(string(body))
	}

	// Audit before writing anything. Fail fast on unexpected
	// marketplace-incompatible content.
	auditInput := make(map[string]string, len(skillBody)+len(agentBody))
	for name, body := range skillBody {
		auditInput["skill:"+name] = body
	}
	for file, body := range agentBody {
		id := strings.TrimSuffix(file, ".agent.md")
		auditInput["agent:"+id] = body
	}
	if findings := Audit(auditInput); len(findings) > 0 {
		var b strings.Builder
		b.WriteString("plugingen: audit found marketplace-incompatible content. Fix the template or add an allow-list entry in pkg/plugingen/audit.go.\n")
		for _, f := range findings {
			fmt.Fprintf(&b, "  - %s\n", f)
		}
		return fmt.Errorf("%s", b.String())
	}

	// Validate pack skill references and that no skill is double-
	// claimed by multiple packs.
	packs := Packs()
	skillSet := make(map[string]bool, len(skillNames))
	for _, n := range skillNames {
		skillSet[n] = true
	}
	claimed := make(map[string]string)
	for _, p := range packs {
		for _, sn := range p.SkillNames {
			if !skillSet[sn] {
				return fmt.Errorf("plugingen: pack %s references unknown skill %q", p.Name, sn)
			}
			if prev, ok := claimed[sn]; ok && prev != p.Name {
				return fmt.Errorf("plugingen: skill %q claimed by both %s and %s", sn, prev, p.Name)
			}
			claimed[sn] = p.Name
		}
	}
	// Every skill should either be in a pack or in MiscSkills.
	miscSet := make(map[string]bool, len(MiscSkills))
	for _, n := range MiscSkills {
		miscSet[n] = true
	}
	for _, sn := range skillNames {
		if _, inPack := claimed[sn]; !inPack && !miscSet[sn] {
			return fmt.Errorf("plugingen: skill %q is not assigned to any pack and not in MiscSkills — add it to one or the other", sn)
		}
	}

	pluginsDir := filepath.Join(cfg.RepoRoot, "plugins")
	if err := resetDir(pluginsDir); err != nil {
		return fmt.Errorf("reset plugins dir: %w", err)
	}

	// 1. Per-skill plugins (atv-skill-<name>).
	for _, name := range skillNames {
		pluginName := pluginNameForSkill(name)
		dir := filepath.Join(pluginsDir, pluginName)
		if err := writeSkillFile(dir, name, skillBody[name]); err != nil {
			return err
		}
		manifest := PluginManifest{
			Name:        pluginName,
			Description: skillPluginDescription(name),
			Version:     cfg.KitVersion,
			Author:      defaultAuthor(),
			Repository:  defaultRepository(),
			License:     "MIT",
			Keywords:    []string{"atv", "skill", strings.ReplaceAll(name, "_", "-")},
			Skills:      []string{"./skills"},
		}
		if err := writeManifest(dir, manifest); err != nil {
			return err
		}
		if err := writePluginReadme(dir, manifest); err != nil {
			return err
		}
	}

	// 2. Pack plugins (atv-pack-<category>).
	for _, p := range packs {
		dir := filepath.Join(pluginsDir, p.Name)
		for _, sn := range p.SkillNames {
			if err := writeSkillFile(dir, sn, skillBody[sn]); err != nil {
				return err
			}
		}
		manifest := PluginManifest{
			Name:        p.Name,
			Description: p.Description,
			Version:     cfg.KitVersion,
			Author:      defaultAuthor(),
			Repository:  defaultRepository(),
			License:     "MIT",
			Keywords:    p.Keywords,
			Category:    p.Category,
			Skills:      []string{"./skills"},
		}
		if err := writeManifest(dir, manifest); err != nil {
			return err
		}
		if err := writePluginReadme(dir, manifest); err != nil {
			return err
		}
	}

	// 3. atv-agents — single plugin bundling every reviewer/specialist agent.
	agentsPluginDir := filepath.Join(pluginsDir, "atv-agents")
	for _, file := range agentFiles {
		if err := writeAgentFile(agentsPluginDir, file, agentBody[file]); err != nil {
			return err
		}
	}
	agentsManifest := PluginManifest{
		Name:        "atv-agents",
		Description: fmt.Sprintf("All %d ATV reviewer and specialist agents — universal reviewers (security, performance, architecture, simplicity), stack-specific reviewers (Rails, Python, TypeScript, .NET, etc.), and Compound Engineering specialists. Install alongside any atv-skill-* or atv-pack-* plugin that dispatches agents.", len(agentFiles)),
		Version:     cfg.KitVersion,
		Author:      defaultAuthor(),
		Repository:  defaultRepository(),
		License:     "MIT",
		Keywords:    []string{"atv", "agents", "reviewers", "specialists"},
		Agents:      []string{"./agents"},
	}
	if err := writeManifest(agentsPluginDir, agentsManifest); err != nil {
		return err
	}
	if err := writePluginReadme(agentsPluginDir, agentsManifest); err != nil {
		return err
	}

	// 4. atv-everything — flagship bundle with every skill + every agent.
	everythingDir := filepath.Join(pluginsDir, "atv-everything")
	for _, name := range skillNames {
		if err := writeSkillFile(everythingDir, name, skillBody[name]); err != nil {
			return err
		}
	}
	for _, file := range agentFiles {
		if err := writeAgentFile(everythingDir, file, agentBody[file]); err != nil {
			return err
		}
	}
	everythingManifest := PluginManifest{
		Name:        "atv-everything",
		Description: fmt.Sprintf("ATV Starter Kit — install everything in one shot: all %d skills (slash commands like /ce-plan, /atv-security, /autoresearch) and all %d reviewer/specialist agents. Equivalent to the Full preset of `atv init`, scoped to skills + agents only (no MCP servers, hooks, or instructions templates — for those use `npx atv-starterkit init`).", len(skillNames), len(agentFiles)),
		Version:     cfg.KitVersion,
		Author:      defaultAuthor(),
		Repository:  defaultRepository(),
		License:     "MIT",
		Keywords:    []string{"atv", "starter-kit", "everything", "compound-engineering"},
		Skills:      []string{"./skills"},
		Agents:      []string{"./agents"},
	}
	if err := writeManifest(everythingDir, everythingManifest); err != nil {
		return err
	}
	if err := writeSourceInstallPluginManifest(everythingDir, cfg); err != nil {
		return err
	}
	if err := writePluginReadme(everythingDir, everythingManifest); err != nil {
		return err
	}

	// 5. .github/plugin/marketplace.json — Copilot CLI manifest enumerating every plugin.
	if err := writeMarketplace(cfg, skillNames, packs); err != nil {
		return err
	}
	// 6. marketplace.json + .claude-plugin/marketplace.json — curated source-install surface.
	if err := writeSourceInstallMarketplace(cfg); err != nil {
		return err
	}

	return nil
}

// CheckClean returns nil if running Generate would not change any
// committed file, or an error listing the changed paths otherwise.
//
// The check generates into a temporary directory, then compares every
// output file against the corresponding committed file. Line endings are
// normalized during comparison so Windows autocrlf checkouts do not create
// false drift reports. This is safer than mutating the working tree and
// diffing.
func CheckClean(cfg Config) error {
	if cfg.RepoRoot == "" {
		return fmt.Errorf("plugingen: RepoRoot is required")
	}
	tmp, err := os.MkdirTemp("", "plugingen-check-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	// Mirror just the inputs the generator reads, then run with
	// RepoRoot pointed at the temp clone.
	if err := mirrorTree(filepath.Join(cfg.RepoRoot, "pkg", "scaffold", "templates"),
		filepath.Join(tmp, "pkg", "scaffold", "templates")); err != nil {
		return err
	}
	tmpCfg := Config{RepoRoot: tmp, KitVersion: cfg.KitVersion}
	if err := Generate(tmpCfg); err != nil {
		return err
	}

	var diffs []string
	for _, sub := range []string{"plugins", filepath.Join(".github", "plugin"), ".claude-plugin"} {
		gotRoot := filepath.Join(tmp, sub)
		wantRoot := filepath.Join(cfg.RepoRoot, sub)
		more, err := compareTrees(gotRoot, wantRoot)
		if err != nil {
			return err
		}
		diffs = append(diffs, prefixDiffs(filepath.ToSlash(sub), more)...)
	}
	more, err := compareFile(filepath.Join(tmp, "marketplace.json"), filepath.Join(cfg.RepoRoot, "marketplace.json"), "marketplace.json")
	if err != nil {
		return err
	}
	diffs = append(diffs, more...)
	if len(diffs) > 0 {
		sort.Strings(diffs)
		return fmt.Errorf("plugin tree out of sync with templates. Run `go run ./cmd/plugingen` and commit the result. Differences:\n  - %s",
			strings.Join(diffs, "\n  - "))
	}
	return nil
}

// --- helpers -------------------------------------------------------------

func listSkills(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

func listAgents(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".agent.md") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

func writeSkillFile(pluginDir, skillName, body string) error {
	dest := filepath.Join(pluginDir, "skills", skillName, "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, []byte(body), 0o644)
}

func writeAgentFile(pluginDir, agentFile, body string) error {
	dest := filepath.Join(pluginDir, "agents", agentFile)
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dest, []byte(body), 0o644)
}

func writeManifest(pluginDir string, m PluginManifest) error {
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return err
	}
	data, err := marshalJSON(m)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(pluginDir, "plugin.json"), data, 0o644)
}

func writePluginReadme(pluginDir string, m PluginManifest) error {
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", m.Name)
	fmt.Fprintf(&b, "%s\n\n", m.Description)
	b.WriteString("## Install\n\n")
	b.WriteString("```bash\n")
	b.WriteString("copilot plugin marketplace add All-The-Vibes/ATV-StarterKit\n")
	fmt.Fprintf(&b, "copilot plugin install %s@atv-starter-kit\n", m.Name)
	b.WriteString("```\n\n")
	b.WriteString("Generated by `pkg/plugingen` from `pkg/scaffold/templates/`. Do not edit by hand — run `go run ./cmd/plugingen` to regenerate.\n")
	return os.WriteFile(filepath.Join(pluginDir, "README.md"), []byte(b.String()), 0o644)
}

func writeSourceInstallPluginManifest(pluginDir string, cfg Config) error {
	manifestDir := filepath.Join(pluginDir, ".claude-plugin")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		return err
	}
	manifest := PluginManifest{
		Name:        "atv-starter-kit",
		Description: sourceInstallDescription(),
		Version:     cfg.KitVersion,
		Author:      defaultAuthor(),
		Homepage:    defaultRepository(),
		Repository:  defaultRepository(),
		License:     "MIT",
		Keywords:    []string{"atv", "starter-kit", "vscode", "agent-plugin"},
	}
	data, err := marshalJSON(manifest)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(manifestDir, "plugin.json"), data, 0o644)
}

func writeMarketplace(cfg Config, skillNames []string, packs []Pack) error {
	mpDir := filepath.Join(cfg.RepoRoot, ".github", "plugin")
	if err := os.MkdirAll(mpDir, 0o755); err != nil {
		return err
	}

	var entries []MarketplaceEntry
	entries = append(entries, MarketplaceEntry{
		Name:        "atv-everything",
		Source:      "atv-everything",
		Description: "ATV Starter Kit — install everything in one shot: all skills + all reviewer/specialist agents.",
		Version:     cfg.KitVersion,
		Keywords:    []string{"atv", "starter-kit", "everything"},
		Category:    "starter-kit",
	})
	entries = append(entries, MarketplaceEntry{
		Name:        "atv-agents",
		Source:      "atv-agents",
		Description: "All ATV reviewer and specialist agents (universal + stack-specific).",
		Version:     cfg.KitVersion,
		Keywords:    []string{"atv", "agents", "reviewers"},
		Category:    "agents",
	})
	for _, p := range packs {
		entries = append(entries, MarketplaceEntry{
			Name:        p.Name,
			Source:      p.Name,
			Description: p.Description,
			Version:     cfg.KitVersion,
			Keywords:    p.Keywords,
			Category:    p.Category,
		})
	}
	for _, name := range skillNames {
		pluginName := pluginNameForSkill(name)
		entries = append(entries, MarketplaceEntry{
			Name:        pluginName,
			Source:      pluginName,
			Description: skillPluginDescription(name),
			Version:     cfg.KitVersion,
			Keywords:    []string{"atv", "skill", strings.ReplaceAll(name, "_", "-")},
			Category:    "skill",
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		leftRank := cliMarketplaceRank(entries[i].Name)
		rightRank := cliMarketplaceRank(entries[j].Name)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return entries[i].Name < entries[j].Name
	})

	mp := Marketplace{
		Name:  "atv-starter-kit",
		Owner: defaultMarketplaceOwner(),
		Metadata: MarketplaceMeta{
			Description: "ATV Starter Kit — agentic coding skills, agents, and orchestrators for GitHub Copilot CLI. See https://github.com/All-The-Vibes/ATV-StarterKit.",
			Version:     cfg.KitVersion,
			PluginRoot:  "./plugins",
		},
		Plugins: entries,
	}
	data, err := marshalJSON(mp)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(mpDir, "marketplace.json"), data, 0o644)
}

func cliMarketplaceRank(name string) int {
	if name == "atv-everything" {
		return 0
	}
	return 1
}

func writeSourceInstallMarketplace(cfg Config) error {
	mpDir := filepath.Join(cfg.RepoRoot, ".claude-plugin")
	if err := os.MkdirAll(mpDir, 0o755); err != nil {
		return err
	}

	mp := SourceInstallMarketplace{
		Name:  "atv-starter-kit",
		Owner: defaultMarketplaceOwner(),
		Metadata: SourceInstallMarketplaceMeta{
			Description: "Curated VS Code source-install catalog for ATV Starter Kit.",
			Version:     cfg.KitVersion,
		},
		Plugins: []SourceInstallEntry{
			{
				Name:        "atv-starter-kit",
				Description: sourceInstallDescription(),
				Author:      defaultAuthor(),
				Homepage:    defaultRepository(),
				Tags:        []string{"atv", "starter-kit", "vscode", "skills", "agents"},
				Source:      "./plugins/atv-everything",
			},
		},
	}
	data, err := marshalJSON(mp)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cfg.RepoRoot, "marketplace.json"), data, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(mpDir, "marketplace.json"), data, 0o644)
}

func sourceInstallDescription() string {
	return "ATV Starter Kit for VS Code: all ATV skills and reviewer/specialist agents in one personal install."
}

func skillPluginDescription(name string) string {
	return fmt.Sprintf("ATV `%s` skill — single-skill plugin (granular install). Some ATV skills dispatch agents bundled separately in `atv-agents`, so single-skill and category-pack installs may need `atv-agents` alongside them. For the full standalone kit, install `atv-everything`. See https://github.com/All-The-Vibes/ATV-StarterKit/blob/main/docs/marketplace.md.", name)
}

func defaultAuthor() *Author {
	return &Author{
		Name: "All The Vibes",
		URL:  "https://github.com/All-The-Vibes",
	}
}

func defaultMarketplaceOwner() Author {
	return Author{
		Name: "All The Vibes",
		URL:  "https://github.com/All-The-Vibes",
	}
}

func defaultRepository() string {
	return "https://github.com/All-The-Vibes/ATV-StarterKit"
}

// marshalJSON renders v with deterministic formatting: 2-space indent,
// LF line endings, and a trailing newline. It uses encoding/json for
// the heavy lifting, which is map-iteration-order safe for structs.
func marshalJSON(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	// json.Encoder always appends a newline; we want exactly one.
	out := bytes.ReplaceAll(buf.Bytes(), []byte("\r\n"), []byte("\n"))
	return out, nil
}

// normalizeLineEndings converts CRLF to LF for text comparisons across
// Git autocrlf checkouts.
func normalizeLineEndings(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

func equalGeneratedContent(a, b []byte) bool {
	if bytes.Equal(a, b) {
		return true
	}
	return normalizeLineEndings(string(a)) == normalizeLineEndings(string(b))
}

func prefixDiffs(prefix string, diffs []string) []string {
	out := make([]string, 0, len(diffs))
	for _, diff := range diffs {
		out = append(out, prefixDiff(prefix, diff))
	}
	return out
}

func prefixDiff(prefix, diff string) string {
	for _, marker := range []string{
		"missing on disk: ",
		"stale on disk (not produced by generator): ",
		"content differs: ",
	} {
		if strings.HasPrefix(diff, marker) {
			rel := strings.TrimPrefix(diff, marker)
			return marker + filepath.ToSlash(filepath.Join(prefix, rel))
		}
	}
	return diff
}

// resetDir removes dir entirely if it exists, then recreates it empty.
// This guarantees we never leave stale files behind from previous
// generator runs (e.g. after a skill is renamed or deleted).
func resetDir(dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	return os.MkdirAll(dir, 0o755)
}

// mirrorTree copies src into dst recursively. Used by CheckClean to
// build a reproducible input tree in a tempdir.
func mirrorTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

// compareTrees returns a list of paths (relative to the common parent)
// that differ between gotRoot and wantRoot. Paths missing on either
// side are reported.
func compareTrees(gotRoot, wantRoot string) ([]string, error) {
	gotFiles, err := walkFiles(gotRoot)
	if err != nil {
		return nil, err
	}
	wantFiles, err := walkFiles(wantRoot)
	if err != nil {
		return nil, err
	}
	all := make(map[string]bool)
	for k := range gotFiles {
		all[k] = true
	}
	for k := range wantFiles {
		all[k] = true
	}
	var diffs []string
	for rel := range all {
		gp, gok := gotFiles[rel]
		wp, wok := wantFiles[rel]
		switch {
		case gok && !wok:
			diffs = append(diffs, "missing on disk: "+rel)
		case !gok && wok:
			diffs = append(diffs, "stale on disk (not produced by generator): "+rel)
		default:
			a, _ := os.ReadFile(gp)
			b, _ := os.ReadFile(wp)
			if !equalGeneratedContent(a, b) {
				diffs = append(diffs, "content differs: "+rel)
			}
		}
	}
	return diffs, nil
}

func compareFile(gotPath, wantPath, rel string) ([]string, error) {
	got, err := os.ReadFile(gotPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"missing generated file: " + rel}, nil
		}
		return nil, err
	}
	want, err := os.ReadFile(wantPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{"missing on disk: " + rel}, nil
		}
		return nil, err
	}
	if !equalGeneratedContent(got, want) {
		return []string{"content differs: " + rel}, nil
	}
	return nil, nil
}

// walkFiles returns a map from forward-slash relative path to absolute
// path for every regular file under root. Returns an empty map (no
// error) if root does not exist — that case is treated as "all files
// missing on disk" and reported by compareTrees.
func walkFiles(root string) (map[string]string, error) {
	out := map[string]string{}
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return out, nil
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = path
		return nil
	})
	return out, err
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
