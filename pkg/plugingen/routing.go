package plugingen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RoutingEntry is one skill's routing row: its slash-command name and the
// intent-rich description pulled from its SKILL.md frontmatter. This is the
// authoritative source for the /atv router's catalog (GATE 1): the template
// SKILL.md `description:` carries trigger phrases, unlike the install
// boilerplate in plugins/*/plugin.json.
type RoutingEntry struct {
	Name        string
	Description string
}

// parseSkillDescription extracts the `description:` value from a SKILL.md
// YAML frontmatter block. It handles bare, double-quoted, and single-quoted
// values (with ” escaping in single-quoted form), matching the mix used
// across ATV templates. Returns an error when no frontmatter or no
// description line is present so callers can skip malformed skills.
func parseSkillDescription(body []byte) (string, error) {
	text := string(body)
	if !strings.HasPrefix(text, "---") {
		return "", fmt.Errorf("no frontmatter")
	}
	// Isolate the frontmatter block between the first two --- fences.
	rest := text[len("---"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", fmt.Errorf("unterminated frontmatter")
	}
	front := rest[:end]
	for _, line := range strings.Split(front, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "description:") {
			continue
		}
		val := strings.TrimSpace(strings.TrimPrefix(trimmed, "description:"))
		return unquoteYAMLScalar(val), nil
	}
	return "", fmt.Errorf("no description")
}

// unquoteYAMLScalar unwraps a single-line YAML scalar: double-quoted,
// single-quoted (with ” -> ' unescaping), or bare.
func unquoteYAMLScalar(val string) string {
	if len(val) >= 2 && strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
		inner := val[1 : len(val)-1]
		inner = strings.ReplaceAll(inner, "\\\"", "\"")
		return inner
	}
	if len(val) >= 2 && strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'") {
		inner := val[1 : len(val)-1]
		return strings.ReplaceAll(inner, "''", "'")
	}
	return val
}

// BuildRoutingCatalog is the exported entrypoint: scans a templates/skills
// root and returns the deduped, name-sorted routing catalog. Used by the
// catalog sync test and any CLI that regenerates the /atv router menu.
func BuildRoutingCatalog(skillsRoot string) ([]RoutingEntry, error) {
	return buildRoutingCatalog(skillsRoot)
}

// RenderRoutingCatalog renders a catalog to the committed llms.txt artifact
// body (one line per skill). Exported for the sync test and CLI regen.
func RenderRoutingCatalog(cat []RoutingEntry) string {
	return renderRoutingCatalog(cat) + "\n"
}

// buildRoutingCatalog scans a templates/skills root, reads each skill's
// SKILL.md frontmatter description, and returns entries sorted by name.
// Skills with missing or malformed frontmatter are skipped (never abort):
// one bad manifest must not break the whole routing table.
func buildRoutingCatalog(skillsRoot string) ([]RoutingEntry, error) {
	entries, err := os.ReadDir(skillsRoot)
	if err != nil {
		return nil, err
	}
	var out []RoutingEntry
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		body, err := os.ReadFile(filepath.Join(skillsRoot, name, "SKILL.md"))
		if err != nil {
			// No SKILL.md — skip, don't abort.
			continue
		}
		desc, err := parseSkillDescription(body)
		if err != nil {
			// Malformed frontmatter — skip, don't abort.
			continue
		}
		out = append(out, RoutingEntry{Name: name, Description: desc})
	}
	out = dedupeByName(out)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// dedupeByName collapses entries sharing a name to the first occurrence.
// Guards against the triplication risk (C3) if a caller ever merges rows
// from multiple roots.
func dedupeByName(in []RoutingEntry) []RoutingEntry {
	seen := make(map[string]bool, len(in))
	var out []RoutingEntry
	for _, e := range in {
		if seen[e.Name] {
			continue
		}
		seen[e.Name] = true
		out = append(out, e)
	}
	return out
}

// renderRoutingCatalog emits an llms.txt-style catalog: one line per skill,
// `- [/name](skills/name/SKILL.md): description`. Deterministic (input is
// pre-sorted). This single artifact feeds both the human-readable route
// table and the machine-readable menu the router matches against.
func renderRoutingCatalog(cat []RoutingEntry) string {
	var b strings.Builder
	for i, e := range cat {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "- [/%s](skills/%s/SKILL.md): %s", e.Name, e.Name, e.Description)
	}
	return b.String()
}
