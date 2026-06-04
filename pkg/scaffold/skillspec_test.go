package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// skillspec_test.go is a structural regression guard for ATV's Copilot skill
// and agent definitions. It asserts that the frontmatter on every shipped
// definition stays well-formed and complete, so the repo's growing set of
// definitions cannot silently drift out of the shape the Copilot harness
// expects (a missing closing `---`, an empty `description`, a `name` that no
// longer matches its folder, a duplicated key, etc.).
//
// It is deliberately MINIMAL and calibrated to ATV's existing conventions so
// it passes green on the current tree. It intentionally does NOT enforce
// stylistic policy that ATV does not already follow — there is no description
// length cap, no required body-section list, no naming-style rule beyond the
// `name == folder` invariant the skills already satisfy, and no banned-word
// list. Its only job is to catch accidental structural breakage.
//
// Scope: the canonical dogfood definitions under `.github/skills` and
// `.github/agents`, plus the consumer-facing skill templates under
// `pkg/scaffold/templates/skills`. The agent templates under
// `pkg/scaffold/templates/agents` are intentionally excluded: they are
// currently stored in a single-line minified form that is not standard
// line-delimited YAML frontmatter, which is tracked as separate follow-up
// debt rather than fixed (or papered over) here.

// parsedFrontmatter holds the top-level frontmatter keys and the document body
// that follow a strict `---` / `---` delimited block.
type parsedFrontmatter struct {
	keys map[string]string
	body string
}

// parseFrontmatter parses a strict, line-delimited YAML-style frontmatter block
// from raw markdown. Line endings are normalized first so CRLF files (`.github`)
// and LF files (templates) are treated identically. It returns a non-empty
// problem string describing the first structural defect found.
//
// Only unindented `key: value` lines are treated as frontmatter keys; indented
// lines and list items (continuations of a value) are ignored so that inline
// list values do not register spurious keys.
func parseFrontmatter(raw string) (*parsedFrontmatter, string) {
	s := strings.ReplaceAll(raw, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	if !strings.HasPrefix(s, "---\n") {
		return nil, "missing opening '---' delimiter on the first line"
	}
	rest := s[len("---\n"):]
	lines := strings.Split(rest, "\n")

	closing := -1
	for i, line := range lines {
		if line == "---" {
			closing = i
			break
		}
	}
	if closing == -1 {
		return nil, "missing closing '---' delimiter on its own line"
	}

	keys := make(map[string]string)
	for _, line := range lines[:closing] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Skip indented lines and list items: these are value continuations,
		// not top-level keys.
		if line[0] == ' ' || line[0] == '\t' || line[0] == '-' {
			continue
		}
		idx := strings.Index(line, ":")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if !isFrontmatterKey(key) {
			continue
		}
		if _, dup := keys[key]; dup {
			return nil, "duplicate frontmatter key '" + key + "'"
		}
		keys[key] = strings.TrimSpace(line[idx+1:])
	}

	body := ""
	if closing+1 < len(lines) {
		body = strings.Join(lines[closing+1:], "\n")
	}
	return &parsedFrontmatter{keys: keys, body: body}, ""
}

// isFrontmatterKey reports whether s is a plausible top-level YAML key
// (letters, digits, hyphen, underscore only).
func isFrontmatterKey(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}

// isPlaceholder reports whether a frontmatter value is an obvious unfilled
// placeholder rather than real content.
func isPlaceholder(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "todo", "tbd", "fixme", "changeme", "change me", "description":
		return true
	}
	return false
}

// TestParseFrontmatterRejectsMalformed proves the gate is not vacuous: each
// class of structural defect it is meant to catch is actually rejected. This
// guards the validator itself against silently degrading into a no-op.
func TestParseFrontmatterRejectsMalformed(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{
			name:    "valid",
			raw:     "---\nname: foo\ndescription: A real description.\n---\nbody text\n",
			wantErr: false,
		},
		{
			name:    "missing opening delimiter (single-line minified)",
			raw:     "---description: x user-invocable: true---body",
			wantErr: true,
		},
		{
			name:    "no opening delimiter at all",
			raw:     "name: foo\ndescription: x\nbody",
			wantErr: true,
		},
		{
			name:    "missing closing delimiter",
			raw:     "---\nname: foo\ndescription: x\nbody without close",
			wantErr: true,
		},
		{
			name:    "duplicate key",
			raw:     "---\nname: foo\nname: bar\ndescription: x\n---\nbody",
			wantErr: true,
		},
		{
			name:    "crlf is normalized and parses",
			raw:     "---\r\nname: foo\r\ndescription: A real description.\r\n---\r\nbody\r\n",
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fm, problem := parseFrontmatter(tc.raw)
			if tc.wantErr {
				if problem == "" {
					t.Fatalf("expected a structural problem, got none (parsed %+v)", fm)
				}
				return
			}
			if problem != "" {
				t.Fatalf("unexpected problem: %s", problem)
			}
			// A normalized value must not carry a trailing CR.
			if got := fm.keys["name"]; strings.ContainsRune(got, '\r') {
				t.Fatalf("name value retained a carriage return: %q", got)
			}
		})
	}
}

// TestFrontmatterValueChecks covers the small value-level predicates.
func TestFrontmatterValueChecks(t *testing.T) {
	for _, v := range []string{"TODO", "tbd", "  FIXME ", "description", "change me"} {
		if !isPlaceholder(v) {
			t.Errorf("expected %q to be detected as a placeholder", v)
		}
	}
	for _, v := range []string{"A real, specific description.", "Run linting before push."} {
		if isPlaceholder(v) {
			t.Errorf("did not expect %q to be a placeholder", v)
		}
	}

	for _, k := range []string{"name", "user-invocable", "allowed_tools", "argument-hint"} {
		if !isFrontmatterKey(k) {
			t.Errorf("expected %q to be a valid key", k)
		}
	}
	for _, k := range []string{"", "- Bash(x)", "a key", "key:sub"} {
		if isFrontmatterKey(k) {
			t.Errorf("did not expect %q to be a valid key", k)
		}
	}
}

// TestSkillFrontmatterSpec validates the frontmatter of every shipped skill in
// the canonical dogfood tree and the consumer-facing template tree.
func TestSkillFrontmatterSpec(t *testing.T) {
	root := repoRoot(t)

	trees := []string{
		filepath.Join(".github", "skills"),
		filepath.Join("pkg", "scaffold", "templates", "skills"),
	}

	for _, tree := range trees {
		treeDir := filepath.Join(root, tree)
		entries, err := os.ReadDir(treeDir)
		if err != nil {
			t.Fatalf("reading skill tree %s: %v", tree, err)
		}

		seen := make(map[string]string) // name -> folder that first claimed it
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			folder := entry.Name()
			skillPath := filepath.Join(treeDir, folder, "SKILL.md")
			rel := filepath.ToSlash(filepath.Join(tree, folder, "SKILL.md"))

			raw, err := os.ReadFile(skillPath)
			if err != nil {
				t.Errorf("%s: skill directory %q has no readable SKILL.md: %v", tree, folder, err)
				continue
			}

			fm, problem := parseFrontmatter(string(raw))
			if problem != "" {
				t.Errorf("%s: %s", rel, problem)
				continue
			}

			name, ok := fm.keys["name"]
			if !ok || strings.TrimSpace(name) == "" {
				t.Errorf("%s: missing or empty required frontmatter key 'name'", rel)
			} else {
				if name != folder {
					t.Errorf("%s: name %q does not match its folder %q", rel, name, folder)
				}
				if prev, dup := seen[name]; dup {
					t.Errorf("%s: duplicate skill name %q (also defined by %q)", rel, name, prev)
				} else {
					seen[name] = folder
				}
			}

			desc := strings.TrimSpace(fm.keys["description"])
			if desc == "" {
				t.Errorf("%s: missing or empty required frontmatter key 'description'", rel)
			} else if isPlaceholder(desc) {
				t.Errorf("%s: description is an unfilled placeholder %q", rel, desc)
			}

			if strings.TrimSpace(fm.body) == "" {
				t.Errorf("%s: body after frontmatter is empty", rel)
			}
		}
	}
}

// TestAgentFrontmatterSpec validates the frontmatter of every canonical agent
// definition under .github/agents.
func TestAgentFrontmatterSpec(t *testing.T) {
	root := repoRoot(t)
	agentsDir := filepath.Join(root, ".github", "agents")

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("reading agents dir: %v", err)
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".agent.md") {
			continue
		}
		count++
		rel := filepath.ToSlash(filepath.Join(".github", "agents", entry.Name()))

		raw, err := os.ReadFile(filepath.Join(agentsDir, entry.Name()))
		if err != nil {
			t.Errorf("%s: unreadable: %v", rel, err)
			continue
		}

		fm, problem := parseFrontmatter(string(raw))
		if problem != "" {
			t.Errorf("%s: %s", rel, problem)
			continue
		}

		desc := strings.TrimSpace(fm.keys["description"])
		if desc == "" {
			t.Errorf("%s: missing or empty required frontmatter key 'description'", rel)
		} else if isPlaceholder(desc) {
			t.Errorf("%s: description is an unfilled placeholder %q", rel, desc)
		}

		ui, ok := fm.keys["user-invocable"]
		if !ok {
			t.Errorf("%s: missing required frontmatter key 'user-invocable'", rel)
		} else if ui != "true" && ui != "false" {
			t.Errorf("%s: user-invocable must be 'true' or 'false', got %q", rel, ui)
		}

		if strings.TrimSpace(fm.body) == "" {
			t.Errorf("%s: body after frontmatter is empty", rel)
		}
	}

	if count == 0 {
		t.Fatal("no .agent.md definitions found under .github/agents — path drift?")
	}
}
