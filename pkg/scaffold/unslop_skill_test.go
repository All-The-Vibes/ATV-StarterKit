package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnslopSkillDocumentsFullSurfaceWorkflow(t *testing.T) {
	root := repoRoot(t)
	skillPath := filepath.Join(root, "pkg", "scaffold", "templates", "skills", "unslop", "SKILL.md")
	contentBytes, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read unslop skill template: %v", err)
	}
	content := string(contentBytes)

	required := []string{
		"code hygiene",
		"comments/docs",
		"frontend/design",
		"architecture",
		"Architecture Slop Detector",
		"Matt-style",
		"depth",
		"seam",
		"locality",
		"leverage",
		"/unslop fix all",
		"/unslop fix --all",
		"high-priority, auto-fix-eligible fixes across all four lanes",
		"Priority selectors are exact",
		"risk",
		"effort",
		"candidate",
		"Multiple lane selectors",
	}
	for _, want := range required {
		if !strings.Contains(content, want) {
			t.Errorf("unslop skill is missing required v2 contract text %q", want)
		}
	}
}

func TestUnslopSkillCopiesStayInSync(t *testing.T) {
	root := repoRoot(t)
	templatePath := filepath.Join(root, "pkg", "scaffold", "templates", "skills", "unslop", "SKILL.md")
	template, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	copyPaths := []string{
		filepath.Join(root, "plugins", "atv-skill-unslop", "skills", "unslop", "SKILL.md"),
		filepath.Join(root, "plugins", "atv-pack-quality", "skills", "unslop", "SKILL.md"),
		filepath.Join(root, "plugins", "atv-everything", "skills", "unslop", "SKILL.md"),
	}
	for _, path := range copyPaths {
		got, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatalf("read %s: %v", path, readErr)
		}
		if string(got) != string(template) {
			t.Errorf("%s drifted from pkg/scaffold/templates/skills/unslop/SKILL.md", path)
		}
	}
}

func TestUnslopPromptShimDoesNotPointAtMissingSkillPath(t *testing.T) {
	root := repoRoot(t)
	promptPath := filepath.Join(root, ".github", "prompts", "unslop.prompt.md")
	contentBytes, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("read unslop prompt shim: %v", err)
	}
	content := string(contentBytes)

	missingPath := ".github/skills/unslop/SKILL.md"
	if strings.Contains(content, missingPath) {
		t.Fatalf("unslop prompt shim references missing path %q; it should invoke the installed unslop skill without a brittle missing path", missingPath)
	}
	if !strings.Contains(content, "installed `unslop` skill") {
		t.Fatalf("unslop prompt shim should tell the harness to invoke the installed `unslop` skill")
	}

	generated := string(BuildPromptShim("unslop"))
	if strings.Contains(generated, missingPath) {
		t.Fatalf("generated unslop prompt shim references missing path %q", missingPath)
	}
	if !strings.Contains(generated, "installed `unslop` skill") {
		t.Fatalf("generated unslop prompt shim should tell the harness to invoke the installed `unslop` skill")
	}
}

func TestReadmeDocumentsUnslopV2Commands(t *testing.T) {
	root := repoRoot(t)
	readmePath := filepath.Join(root, "README.md")
	contentBytes, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	content := string(contentBytes)

	required := []string{
		"`/unslop` reports code hygiene, comments/docs, frontend/design, and architecture slop",
		"`/unslop fix` applies safe hygiene and comments/docs cleanup",
		"`/unslop fix all` applies high-priority eligible fixes across all lanes",
	}
	for _, want := range required {
		if !strings.Contains(content, want) {
			t.Errorf("README missing unslop v2 documentation %q", want)
		}
	}
}
