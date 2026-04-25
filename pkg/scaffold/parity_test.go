package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/detect"
)

// TestCoreLayerShipsLandAndTakeoff verifies that selecting the core-skills
// layer in --guided mode produces components for both the takeoff and land
// session-lifecycle skills. Regression guard for the gap this test file
// was created to close.
func TestCoreLayerShipsLandAndTakeoff(t *testing.T) {
	components := BuildFilteredCatalog(detect.StackGeneral, []string{"core-skills"})

	want := map[string]bool{
		".github/skills/land/SKILL.md":    false,
		".github/skills/takeoff/SKILL.md": false,
	}
	for _, c := range components {
		// Filepath separator may be OS-specific; normalize.
		p := filepath.ToSlash(c.Path)
		if _, ok := want[p]; ok {
			want[p] = true
		}
	}
	for path, found := range want {
		if !found {
			t.Errorf("expected %q in core-skills layer output, not found", path)
		}
	}

	// Negative: without core-skills, neither file should appear (proves
	// they are not smuggled in via another layer such as orchestrators).
	other := BuildFilteredCatalog(detect.StackGeneral, []string{"orchestrators", "easter-eggs"})
	for _, c := range other {
		p := filepath.ToSlash(c.Path)
		if strings.HasSuffix(p, "/skills/land/SKILL.md") || strings.HasSuffix(p, "/skills/takeoff/SKILL.md") {
			t.Errorf("did not expect %q without core-skills layer selected", p)
		}
	}
}

// TestSkillDirectoryParity ensures every skill directory under
// pkg/scaffold/templates/skills/ is registered in exactly one catalog
// slice. This catches the case where a skill template is added but the
// wiring step is forgotten, which would silently exclude it from --guided
// installs.
func TestSkillDirectoryParity(t *testing.T) {
	templateDirs := readEmbeddedSkillDirs(t)

	registered := make(map[string]string)
	register := func(name, layer string) {
		if existing, ok := registered[name]; ok {
			t.Fatalf("skill %q is registered in both %q and %q", name, existing, layer)
		}
		registered[name] = layer
	}
	for _, name := range coreSkillDirectories {
		register(name, "core")
	}
	for _, name := range orchestratorSkillDirectories {
		register(name, "orchestrators")
	}
	for _, name := range easterEggSkillDirectories {
		register(name, "easter-eggs")
	}
	for _, name := range devToolsSkillDirectories {
		register(name, "dev-tools")
	}
	for _, name := range styleSkillDirectories {
		register(name, "style-skills")
	}
	for _, name := range mediaSkillDirectories {
		register(name, "media-skills")
	}

	var unregistered []string
	for _, dir := range templateDirs {
		if _, ok := registered[dir]; !ok {
			unregistered = append(unregistered, dir)
		}
	}
	if len(unregistered) > 0 {
		t.Fatalf(
			"skill template directories not registered in any catalog slice: %v\n"+
				"Add each name to coreSkillDirectories, orchestratorSkillDirectories, "+
				"easterEggSkillDirectories, devToolsSkillDirectories, styleSkillDirectories, "+
				"or mediaSkillDirectories in pkg/scaffold/catalog.go.",
			unregistered,
		)
	}

	templateSet := make(map[string]bool, len(templateDirs))
	for _, d := range templateDirs {
		templateSet[d] = true
	}
	var orphans []string
	for name := range registered {
		if !templateSet[name] {
			orphans = append(orphans, name)
		}
	}
	if len(orphans) > 0 {
		sort.Strings(orphans)
		t.Fatalf(
			"skill names registered in catalog.go but missing from pkg/scaffold/templates/skills/: %v\n"+
				"Add the SKILL.md template or remove the catalog entry.",
			orphans,
		)
	}
}

// TestDogfoodTemplateParity ensures that every skill present in
// .github/skills/ (the dogfooding source-of-truth used by this repo's own
// Copilot configuration) is also present under pkg/scaffold/templates/skills/
// (the embedded copy the installer ships). Without this, a skill added to
// .github/skills/ would silently miss the --guided install pipeline.
//
// This is a presence check only. Content drift between the two copies is
// accepted: .github/skills/<name>/ is the editable source, and the template
// is a periodic snapshot.
func TestDogfoodTemplateParity(t *testing.T) {
	repoRoot := repoRoot(t)

	dogfoodRoot := filepath.Join(repoRoot, ".github", "skills")
	dogfoodEntries, err := os.ReadDir(dogfoodRoot)
	if err != nil {
		t.Fatalf("reading %s: %v", dogfoodRoot, err)
	}

	dogfoodSkills := make(map[string]bool)
	for _, e := range dogfoodEntries {
		if e.IsDir() {
			dogfoodSkills[e.Name()] = true
		}
	}

	templateSkills := make(map[string]bool)
	for _, d := range readEmbeddedSkillDirs(t) {
		templateSkills[d] = true
	}

	// Skills intentionally living in only one location.
	//
	// templateOnly: skills that ship via the installer but are not used to
	// dogfood this repo. Small, justified per-entry.
	templateOnly := map[string]bool{
		// karpathy-guidelines ships only as a template; there is no
		// .github/skills/karpathy-guidelines/ in this repo.
		"karpathy-guidelines": true,
		// autoresearch ships only as a template (sourced from
		// github/awesome-copilot, MIT). No dogfooded copy in this repo.
		"autoresearch": true,
		// unslop ships only as a template (ATV quality skill).
		"unslop": true,
		// atv-security ships only as a template (security skill
		// added via the installer template tree, not dogfooded yet).
		// Note: the former `cso` template was folded into atv-security.
		"atv-security": true,
	}

	// dogfoodOnly: skills intentionally kept under .github/skills/ for the
	// repo's own Copilot configuration but deliberately NOT mirrored into
	// the installer template tree. Each entry must carry a one-line reason
	// so the list stays honest. Treat membership as a deliberate decision,
	// not parking-lot tech debt.
	//
	// To shrink this list:
	//   1. Mirror the skill into pkg/scaffold/templates/skills/<name>/ and
	//      register it in catalog.go (then remove the entry here), or
	//   2. Remove the .github/skills/<name>/ directory entirely.
	dogfoodOnly := map[string]bool{
		// CE-internal architecture audit; not user-facing.
		"agent-native-audit": true,
		// Beta variant of ce-work; will be deleted once external-delegate
		// mode lands in ce-work proper.
		"ce-work-beta": true,
		// CE-internal documentation skill superseded for users by
		// ce-compound / ce-compound-refresh.
		"compound-docs": true,
		// CE-internal deploy-docs tool for the plugin docs site.
		"deploy-docs": true,
		// Repo-internal todo tracking superseded for users by todo-create
		// / todo-resolve / todo-triage.
		"file-todos": true,
		// Meta-skill for repairing other skills; not user-facing.
		"heal-skill": true,
		// CE-internal swarm orchestration superseded for users by slfg / lfg.
		"orchestrating-swarms": true,
		// CE-only bug-report skill; ATV users should not file bugs in the
		// CE plugin.
		"report-bug-ce": true,
		// Superseded for users by ghcp-review-resolve; kept for
		// repo-internal use during PR review experiments.
		"resolve-pr-feedback": true,
		// iOS-specific build/test skill; not relevant to the current ATV
		// stack-detection set. Revisit if/when an iOS pack ships.
		"test-xcode": true,
		// Lower-level CLI todo triage primitive; user-facing equivalent
		// is todo-triage.
		"triage": true,
	}

	// Stale-entry checks: every name in templateOnly must exist in
	// templates/skills/, and every name in dogfoodOnly must exist in
	// .github/skills/. Without these, the allow-lists silently rot.
	var staleTemplateOnly []string
	for name := range templateOnly {
		if !templateSkills[name] {
			staleTemplateOnly = append(staleTemplateOnly, name)
		}
	}
	if len(staleTemplateOnly) > 0 {
		sort.Strings(staleTemplateOnly)
		t.Errorf(
			"templateOnly contains entries no longer present under pkg/scaffold/templates/skills/: %v\n"+
				"Remove each stale entry from this test.",
			staleTemplateOnly,
		)
	}
	var staleDogfoodOnly []string
	for name := range dogfoodOnly {
		if !dogfoodSkills[name] {
			staleDogfoodOnly = append(staleDogfoodOnly, name)
		}
	}
	if len(staleDogfoodOnly) > 0 {
		sort.Strings(staleDogfoodOnly)
		t.Errorf(
			"dogfoodOnly contains entries no longer present under .github/skills/: %v\n"+
				"Remove each stale entry from this test (the skill has been deleted or renamed).",
			staleDogfoodOnly,
		)
	}

	// Conflict check: a skill in dogfoodOnly should not also exist in
	// templates/skills/. Pick one.
	var conflictingDogfoodOnly []string
	for name := range dogfoodOnly {
		if templateSkills[name] {
			conflictingDogfoodOnly = append(conflictingDogfoodOnly, name)
		}
	}
	if len(conflictingDogfoodOnly) > 0 {
		sort.Strings(conflictingDogfoodOnly)
		t.Errorf(
			"skills listed in dogfoodOnly but also present under pkg/scaffold/templates/skills/: %v\n"+
				"Pick one — either remove the dogfoodOnly entry (the skill is now shipping), "+
				"or remove the template directory (the skill is deliberately repo-local).",
			conflictingDogfoodOnly,
		)
	}

	var missingFromTemplates []string
	for name := range dogfoodSkills {
		if templateSkills[name] || dogfoodOnly[name] {
			continue
		}
		missingFromTemplates = append(missingFromTemplates, name)
	}
	if len(missingFromTemplates) > 0 {
		sort.Strings(missingFromTemplates)
		t.Fatalf(
			"skills present in .github/skills/ but missing from pkg/scaffold/templates/skills/: %v\n"+
				"Mirror each skill into pkg/scaffold/templates/skills/<name>/ so --guided installs ship it, "+
				"or — if the skill is intentionally repo-local — add it to dogfoodOnly in this test "+
				"with a one-line rationale.",
			missingFromTemplates,
		)
	}

	var missingFromDogfood []string
	for name := range templateSkills {
		if dogfoodSkills[name] || templateOnly[name] {
			continue
		}
		missingFromDogfood = append(missingFromDogfood, name)
	}
	if len(missingFromDogfood) > 0 {
		sort.Strings(missingFromDogfood)
		t.Fatalf(
			"skills present in pkg/scaffold/templates/skills/ but missing from .github/skills/: %v\n"+
				"Either mirror the skill into .github/skills/<name>/, or add it to templateOnly in this test.",
			missingFromDogfood,
		)
	}
}

// TestNewLayersExposeSkills exercises each skill layer key, asserting that
// BuildFilteredCatalog emits exactly the templates registered in the
// matching catalog slice and nothing else. Regression guard for the
// dev-tools / style-skills / media-skills layer wiring.
func TestNewLayersExposeSkills(t *testing.T) {
	cases := []struct {
		layer string
		want  []string
	}{
		{"dev-tools", devToolsSkillDirectories},
		{"style-skills", styleSkillDirectories},
		{"media-skills", mediaSkillDirectories},
	}

	for _, tc := range cases {
		t.Run(tc.layer, func(t *testing.T) {
			components := BuildFilteredCatalog(detect.StackGeneral, []string{tc.layer})

			gotSkills := make(map[string]bool)
			for _, c := range components {
				p := filepath.ToSlash(c.Path)
				const prefix = ".github/skills/"
				if !strings.HasPrefix(p, prefix) {
					continue
				}
				rest := strings.TrimPrefix(p, prefix)
				if !strings.Contains(rest, "/") {
					continue
				}
				skill := rest[:strings.Index(rest, "/")]
				gotSkills[skill] = true
			}

			for _, skill := range tc.want {
				if !gotSkills[skill] {
					t.Errorf("layer %q did not emit skill %q", tc.layer, skill)
				}
			}

			// Negative: skills from other layers should not appear when only
			// this layer is selected.
			otherLayers := []struct {
				layer string
				dirs  []string
			}{
				{"core-skills", coreSkillDirectories},
				{"orchestrators", orchestratorSkillDirectories},
				{"easter-eggs", easterEggSkillDirectories},
			}
			for _, other := range otherLayers {
				if other.layer == tc.layer {
					continue
				}
				for _, foreign := range other.dirs {
					if gotSkills[foreign] {
						t.Errorf("layer %q leaked skill %q from %q", tc.layer, foreign, other.layer)
					}
				}
			}
		})
	}

	t.Run("empty layer list returns no skills", func(t *testing.T) {
		components := BuildFilteredCatalog(detect.StackGeneral, []string{})
		for _, c := range components {
			p := filepath.ToSlash(c.Path)
			if strings.HasPrefix(p, ".github/skills/") && strings.Contains(strings.TrimPrefix(p, ".github/skills/"), "/") {
				t.Errorf("expected no skill components for empty layer list, got %q", p)
			}
		}
	})

	t.Run("union of layers deduplicates", func(t *testing.T) {
		components := BuildFilteredCatalog(detect.StackGeneral, []string{"dev-tools", "style-skills", "media-skills"})

		seen := make(map[string]int)
		for _, c := range components {
			p := filepath.ToSlash(c.Path)
			if strings.HasPrefix(p, ".github/skills/") {
				seen[p]++
			}
		}
		for path, count := range seen {
			if count > 1 {
				t.Errorf("path %q emitted %d times — expected exactly 1", path, count)
			}
		}
	})
}

// readEmbeddedSkillDirs returns the immediate subdirectory names under
// templates/skills/ in the embedded template FS.
func readEmbeddedSkillDirs(t *testing.T) []string {
	t.Helper()

	entries, err := fs.ReadDir(templateFS, "templates/skills")
	if err != nil {
		t.Fatalf("reading embedded templates/skills: %v", err)
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	sort.Strings(dirs)
	return dirs
}

// repoRoot returns the repository root, derived from this test file's
// location, so the parity check works regardless of cwd.
func repoRoot(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// thisFile is .../pkg/scaffold/parity_test.go → climb two levels.
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repoRoot %q does not contain go.mod — was the package moved?: %v", root, err)
	}
	return root
}
