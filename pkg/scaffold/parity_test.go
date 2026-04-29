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
// pkg/scaffold/templates/skills/ is registered in exactly one of the three
// catalog slices (core, orchestrator, easter-egg). This catches the case
// where a skill template is added but the wiring step is forgotten, which
// would silently exclude it from --guided installs.
func TestSkillDirectoryParity(t *testing.T) {
	templateDirs := readEmbeddedSkillDirs(t)

	registered := make(map[string]string)
	for _, name := range coreSkillDirectories {
		if existing, ok := registered[name]; ok {
			t.Fatalf("skill %q is registered in both %q and core", name, existing)
		}
		registered[name] = "core"
	}
	for _, name := range orchestratorSkillDirectories {
		if existing, ok := registered[name]; ok {
			t.Fatalf("skill %q is registered in both %q and orchestrators", name, existing)
		}
		registered[name] = "orchestrators"
	}
	for _, name := range easterEggSkillDirectories {
		if existing, ok := registered[name]; ok {
			t.Fatalf("skill %q is registered in both %q and easter-eggs", name, existing)
		}
		registered[name] = "easter-eggs"
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
				"or easterEggSkillDirectories in pkg/scaffold/catalog.go.",
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

// TestDogfoodTemplateParity keeps the repository dogfood surface explicit.
// pkg/scaffold/templates/skills/ is the canonical source for installable ATV
// product skills. .github/skills/ may mirror a product skill, intentionally
// diverge for dogfooding, or contain dogfood-only skills, but every case must
// be recorded in pkg/scaffold/skill_sources.json.
func TestDogfoodTemplateParity(t *testing.T) {
	// The full validation lives in skill_sources_test.go so the inventory is
	// data-driven instead of being split across hard-coded allow-lists.
	_ = readSkillSourceManifest(t)
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
