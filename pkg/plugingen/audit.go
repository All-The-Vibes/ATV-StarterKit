package plugingen

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// AuditFinding describes one marketplace-incompatible string located
// inside a SKILL.md or agent file.
type AuditFinding struct {
	SkillOrAgent string // "skill:ce-plan" or "agent:security-reviewer"
	Pattern      string // human-readable pattern label
	Match        string // first matched text (truncated)
	LineNumber   int    // 1-based; 0 if line tracking not applicable
}

func (f AuditFinding) String() string {
	if f.LineNumber > 0 {
		return fmt.Sprintf("[%s] %s @ line %d: %q", f.SkillOrAgent, f.Pattern, f.LineNumber, f.Match)
	}
	return fmt.Sprintf("[%s] %s: %q", f.SkillOrAgent, f.Pattern, f.Match)
}

// auditPattern represents one rule applied to every audited file.
type auditPattern struct {
	label string
	re    *regexp.Regexp
}

// auditPatterns is the full ruleset. Each pattern catches a class of
// marketplace-incompatible content that would break for users
// installing via the Copilot CLI marketplace (because they may not
// have the referenced plugin/agent/cache path available).
var auditPatterns = []auditPattern{
	{
		label: "compound-engineering plugin cache path",
		re:    regexp.MustCompile(`(?i)(?:~/\.copilot/plugins/cache/[^"\s]*compound-engineering|plugins/compound-engineering/)`),
	},
	{
		label: "compound-engineering namespaced agent or tool",
		re:    regexp.MustCompile(`compound-engineering[:_][a-z][a-z0-9-]*(?::[a-z0-9*-]+)?`),
	},
	{
		label: "MCP plugin tool name (mcp__plugin_...)",
		re:    regexp.MustCompile(`mcp__plugin_[a-zA-Z0-9_-]+__[a-zA-Z0-9_-]+`),
	},
}

// auditAllowList is the set of subjects (skill or agent IDs) that are
// known to reference the compound-engineering plugin. They still ship
// in the marketplace, but with explicit documentation that
// compound-engineering needs to be installed alongside them for full
// functionality. Without compound-engineering, these skills/agents
// degrade gracefully (the missing-plugin discovery returns fewer
// results) rather than crashing.
//
// Adding to this list is allowed but should always be paired with a
// comment explaining why the dependency exists. The fix is to
// eventually rewrite the templates to be truly marketplace-agnostic.
//
// The reason field is informational — it is not enforced but provides
// a paper trail for future cleanup work.
var auditAllowList = map[string]string{
	"skill:deepen-plan":                "References compound-engineering plugin cache and research agents during the discovery phase. Falls back to limited discovery when compound-engineering is absent.",
	"skill:ce-ideate":                  "References compound-engineering research agents for multi-perspective ideation. Falls back to single-perspective when absent.",
	"skill:ce-plan":                    "Delegates deep research to compound-engineering's research agents when available. In-process planning otherwise.",
	"skill:ce-review":                  "Dispatches the full compound-engineering reviewer fleet (correctness, testing, maintainability, project-standards, agent-native, security, performance, api-contract, data-migrations, reliability, adversarial, cli-readiness, previous-comments, language-specific reviewers). Without compound-engineering, ce-review uses only the agents bundled in atv-agents.",
	"skill:document-review":            "Dispatches compound-engineering document-review reviewers (coherence, feasibility, product-lens, design-lens, security-lens, scope-guardian, adversarial). Falls back to atv-agents when absent.",
	"skill:frontend-design":            "References compound-engineering:design:design-iterator for iterative UI refinement. Falls back to single-pass design when compound-engineering is absent.",
	"skill:ghcp-review-resolve":        "Delegates merge-conflict resolution to compound-engineering:ce-work as an optional escape hatch. The skill still functions without compound-engineering — users resolve conflicts manually.",
	"skill:todo-resolve":               "Delegates per-comment PR review-thread resolution to compound-engineering:workflow:pr-comment-resolver. Falls back to inline resolution when compound-engineering is absent.",
	"agent:project-standards-reviewer": "References compound-engineering:research:learnings-researcher and the compound-engineering plugin cache for standards lookup. Skips the lookup gracefully when compound-engineering is not installed.",
}

// Audit scans a set of files (paths to SKILL.md or .agent.md) and
// returns any non-allow-listed marketplace-incompatible findings.
//
// The contents map keys are "skill:<name>" or "agent:<id>" (used in
// findings and for allow-list lookup). Values are the file contents.
//
// An empty return slice means the audit passed.
func Audit(contents map[string]string) []AuditFinding {
	keys := make([]string, 0, len(contents))
	for k := range contents {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var findings []AuditFinding
	for _, key := range keys {
		if _, allowed := auditAllowList[key]; allowed {
			continue
		}
		body := contents[key]
		lines := strings.Split(body, "\n")
		for _, p := range auditPatterns {
			for i, line := range lines {
				match := p.re.FindString(line)
				if match == "" {
					continue
				}
				if len(match) > 80 {
					match = match[:77] + "..."
				}
				findings = append(findings, AuditFinding{
					SkillOrAgent: key,
					Pattern:      p.label,
					Match:        match,
					LineNumber:   i + 1,
				})
			}
		}
	}
	return findings
}
