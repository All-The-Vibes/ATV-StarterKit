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
		// atv-doctor and atv-update ship only as templates (maintenance
		// skills added in v2.6.3). No dogfooded copy in this repo.
		"atv-doctor": true,
		"atv-update": true,
		// unslop ships only as a template (ATV quality skill).
		"unslop": true,
		// atv-security ships only as a template (security skill
		// added via the installer template tree, not dogfooded yet).
		// Note: the former `cso` template was folded into atv-security.
		"atv-security": true,
	}

	// pendingMirror: skills that exist under .github/skills/ but were never
	// copied into the installer template tree. This list freezes the current
	// state so future drift fails CI, but every entry here represents real
	// tech debt: a skill that this repo dogfoods but that --guided users
	// don't get. Shrink this list over time by either:
	//   1. Copying the skill into pkg/scaffold/templates/skills/<name>/ and
	//      registering it in catalog.go (then remove the entry here), or
	//   2. Removing the unused .github/skills/<name>/ directory entirely.
	//
	// Do NOT add entries here without also opening a tracking issue. The
	// presence of an entry should always be a question, not an answer.
	pendingMirror := map[string]bool{
		"agent-browser":             true,
		"agent-native-architecture": true,
		"agent-native-audit":        true,
		"andrew-kane-gem-writer":    true,
		"ce-work-beta":              true,
		"changelog":                 true,
		"compound-docs":             true,
		"create-agent-skill":        true,
		"create-agent-skills":       true,
		"deploy-docs":               true,
		"dhh-rails-style":           true,
		"dspy-ruby":                 true,
		"every-style-editor":        true,
		"file-todos":                true,
		"frontend-design":           true,
		"gemini-imagegen":           true,
		"generate_command":          true,
		"ghcp-review-resolve":       true,
		"git-clean-gone-branches":   true,
		"git-commit":                true,
		"git-commit-push-pr":        true,
		"git-worktree":              true,
		"heal-skill":                true,
		"onboarding":                true,
		"proof":                     true,
		"rclone":                    true,
		"report-bug":                true,
		"report-bug-ce":             true,
		"reproduce-bug":             true,
		"resolve-pr-feedback":       true,
		"resolve-pr-parallel":       true,
		"resolve_parallel":          true,
		"skill-creator":             true,
		"test-xcode":                true,
		"todo-create":               true,
		"todo-resolve":              true,
		"todo-triage":               true,
		"triage":                    true,
		"workflows-brainstorm":      true,
		"workflows-compound":        true,
		"workflows-plan":            true,
		"workflows-review":          true,
		"workflows-work":            true,
	}

	// Stale-entry check: every name listed in templateOnly must actually
	// exist under templates/skills/, and every name in pendingMirror must
	// actually exist under .github/skills/. If a skill is removed or
	// renamed and the allow-list isn't updated, fail loudly so the list
	// stays honest instead of silently accumulating dead entries.
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
	var stalePendingMirror []string
	for name := range pendingMirror {
		if !dogfoodSkills[name] {
			stalePendingMirror = append(stalePendingMirror, name)
		}
	}
	if len(stalePendingMirror) > 0 {
		sort.Strings(stalePendingMirror)
		t.Errorf(
			"pendingMirror contains entries no longer present under .github/skills/: %v\n"+
				"Remove each stale entry from this test (the underlying tech debt is gone).",
			stalePendingMirror,
		)
	}

	// Graduation check: a skill that has been mirrored into templates/skills/
	// should be removed from pendingMirror so the list shrinks over time
	// instead of fossilizing.
	var graduatedPendingMirror []string
	for name := range pendingMirror {
		if templateSkills[name] {
			graduatedPendingMirror = append(graduatedPendingMirror, name)
		}
	}
	if len(graduatedPendingMirror) > 0 {
		sort.Strings(graduatedPendingMirror)
		t.Errorf(
			"pendingMirror contains skills now mirrored under pkg/scaffold/templates/skills/: %v\n"+
				"Remove each graduated entry — the tech debt has been paid down.",
			graduatedPendingMirror,
		)
	}

	var missingFromTemplates []string
	for name := range dogfoodSkills {
		if templateSkills[name] || pendingMirror[name] {
			continue
		}
		missingFromTemplates = append(missingFromTemplates, name)
	}
	if len(missingFromTemplates) > 0 {
		sort.Strings(missingFromTemplates)
		t.Fatalf(
			"skills present in .github/skills/ but missing from pkg/scaffold/templates/skills/: %v\n"+
				"Copy each skill into pkg/scaffold/templates/skills/<name>/ so --guided installs ship it. "+
				"If a skill is intentionally dogfood-only and shouldn't ship, add it to pendingMirror in this test "+
				"(but treat that as recording tech debt, not closing the gap).",
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
