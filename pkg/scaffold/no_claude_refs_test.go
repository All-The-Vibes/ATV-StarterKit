package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestNoClaudeCodeReferencesInSkills enforces that skill files under
// .github/skills/ and pkg/scaffold/templates/skills/ do not contain
// Claude Code-specific references that would confuse a GitHub Copilot
// agent loading these instructions at runtime.
//
// This is a Copilot harness. References to "Claude Code", "~/.claude/"
// paths, and Claude Code-only affordances should be removed or rephrased
// to be harness-neutral. Genuine multi-provider documentation, security
// rules that document Anthropic key patterns, and external SDK package
// names (e.g., "@anthropic-ai/claude-agent-sdk") are explicitly allowed.
//
// The allowlist below captures every legitimate retention. Add to it
// only when the reference is informative context, not Claude Code drift.
func TestNoClaudeCodeReferencesInSkills(t *testing.T) {
	root := repoRoot(t)

	// Forbidden patterns. Each one represents Claude Code drift that
	// should not appear in instructional text loaded by a Copilot agent.
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)claude code`),
		regexp.MustCompile(`(?i)\.claude/`),
		regexp.MustCompile(`(?i)CLAUDE\.md`),
		regexp.MustCompile(`(?i)anthropic`),
		// Claude-Code-specific environment variables (e.g.
		// CLAUDE_PLUGIN_ROOT, CLAUDE_HOME). These are unset in a
		// Copilot harness, so leaving them in skill instructions is a
		// runtime correctness bug, not just naming drift.
		regexp.MustCompile(`\bCLAUDE_[A-Z][A-Z0-9_]*\b`),
		// Claude-Code documentation domains. The plain `\.claude/`
		// pattern above does not match URL hosts like
		// `code.claude.com` or `platform.claude.com` because there is
		// no slash immediately after `claude`.
		regexp.MustCompile(`(?i)(code|platform)\.claude\.(com|ai)`),
	}

	// Files (or file:line ranges) that legitimately mention these
	// patterns. Justifications are tracked here so the allowlist is a
	// review artifact rather than a silent escape hatch.
	//
	// Format: "<repo-relative path>" -> reason
	//
	// Every entry is asserted to exist on disk below — a stale entry
	// (path moved/deleted) becomes a test failure rather than a silent
	// future exemption for whatever shows up at that path next.
	allowlist := map[string]string{
		// dspy-ruby is a multi-provider integration guide. Anthropic
		// appears alongside OpenAI, Gemini, Ollama, and OpenRouter as
		// one of N supported LLM providers. This is informative, not
		// drift.
		".github/skills/dspy-ruby/SKILL.md":                  "multi-provider documentation",
		".github/skills/dspy-ruby/references/providers.md":   "multi-provider documentation",
		".github/skills/dspy-ruby/assets/config-template.rb": "multi-provider configuration template",

		// atv-security documents secret-detection rules. The Anthropic
		// API key pattern (`sk-ant-*`) and the ${ANTHROPIC_API_KEY}
		// fix recommendation are legitimate security-rule content.
		"pkg/scaffold/templates/skills/atv-security/SKILL.md": "security-rule redaction patterns",

		// agent-native-architecture references documents agent-native
		// patterns that legitimately reference the Anthropic SDK
		// (`@anthropic-ai/claude-agent-sdk` is an actual published
		// package), CI workflows that pass ANTHROPIC_API_KEY as a
		// secret, and multi-provider compatibility tables.
		".github/skills/agent-native-architecture/references/mcp-tool-design.md":      "Anthropic SDK package name",
		".github/skills/agent-native-architecture/references/agent-native-testing.md": "CI secret env var",
		".github/skills/agent-native-architecture/references/mobile-patterns.md":      "multi-provider compat table",

		// onboarding inventory script detects API-key env var names
		// across providers; the regex pattern legitimately includes
		// CLAUDE/ANTHROPIC alongside OPENAI.
		".github/skills/onboarding/scripts/inventory.mjs": "secret-detection regex",
	}

	// Fail fast if any allowlisted path no longer exists on disk. A
	// stale allowlist silently exempts whatever future file lands at
	// that path, which is exactly the regression this test exists to
	// prevent.
	for rel := range allowlist {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(abs); err != nil {
			t.Errorf("allowlisted path %q does not exist: %v (drop the entry or restore the file)", rel, err)
		}
	}

	scanRoots := []string{
		filepath.Join(root, ".github", "skills"),
		filepath.Join(root, "pkg", "scaffold", "templates", "skills"),
	}

	// Only scan textual instruction surfaces. Skip binaries and asset
	// bundles by extension. Hoisted out of the WalkDir callback to
	// avoid reallocating per-file.
	textExts := map[string]bool{
		".md":   true,
		".mdx":  true,
		".rb":   true,
		".mjs":  true,
		".js":   true,
		".ts":   true,
		".json": true,
		".yml":  true,
		".yaml": true,
		".sh":   true,
		".py":   true,
	}

	type violation struct {
		path    string
		line    int
		match   string
		content string
	}
	var violations []violation

	for _, scanRoot := range scanRoots {
		if _, err := os.Stat(scanRoot); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(scanRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			// Only scan textual instruction surfaces. Skip binaries
			// and asset bundles by extension.
			ext := strings.ToLower(filepath.Ext(path))
			if !textExts[ext] {
				return nil
			}

			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			rel = filepath.ToSlash(rel)
			if _, allowed := allowlist[rel]; allowed {
				return nil
			}

			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			lines := strings.Split(string(content), "\n")
			for i, line := range lines {
				for _, pat := range patterns {
					if loc := pat.FindStringIndex(line); loc != nil {
						violations = append(violations, violation{
							path:    rel,
							line:    i + 1,
							match:   line[loc[0]:loc[1]],
							content: strings.TrimSpace(line),
						})
						break // one violation per line is enough
					}
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", scanRoot, err)
		}
	}

	if len(violations) > 0 {
		t.Errorf("Found %d Claude Code references in skill files. ATV is a GitHub Copilot harness; these need to be removed or added to the allowlist with justification.", len(violations))
		// Group by file for readability.
		byFile := map[string][]violation{}
		for _, v := range violations {
			byFile[v.path] = append(byFile[v.path], v)
		}
		for path, vs := range byFile {
			t.Errorf("\n  %s (%d hit(s)):", path, len(vs))
			for _, v := range vs {
				preview := v.content
				if len(preview) > 100 {
					preview = preview[:97] + "..."
				}
				t.Errorf("    L%d [%s]: %s", v.line, v.match, preview)
			}
		}
	}
}
