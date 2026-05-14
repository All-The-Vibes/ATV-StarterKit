package plugingen

// Pack defines a category meta-plugin. A Pack bundles multiple skills
// (by their template directory name under pkg/scaffold/templates/skills/)
// into a single installable plugin.
//
// SkillNames must match directory names exactly. Generation will fail
// loudly if any name doesn't resolve to an existing template.
type Pack struct {
	// Name is the plugin name written into plugin.json and the
	// directory name under plugins/ (e.g. "atv-pack-planning").
	Name string

	// Category is a human-readable label for plugin metadata.
	Category string

	// Description appears in plugin.json and marketplace.json.
	Description string

	// Keywords help discoverability via `copilot plugin marketplace browse`.
	Keywords []string

	// SkillNames is the ordered list of template directory names under
	// pkg/scaffold/templates/skills/ to bundle. Order is preserved in
	// the generated plugin.json's skills array.
	SkillNames []string
}

// Packs returns the canonical set of category meta-plugins shipped by
// the ATV marketplace. The ordering here is the order they appear in
// marketplace.json (after sorting at the marketplace level — see
// generate.go for sort behaviour).
//
// The skill assignments mirror the categories surfaced by the guided
// installer's TUI (pkg/tui/categories.go) so users see consistent
// groupings across both install paths.
func Packs() []Pack {
	return []Pack{
		{
			Name:        "atv-pack-planning",
			Category:    "planning",
			Description: "Planning & design pack — Brainstorming, CE Brainstorm, CE Ideate, Plan, and Deepen Plan skills for shaping work before coding.",
			Keywords:    []string{"atv", "compound-engineering", "planning", "brainstorming"},
			SkillNames:  []string{"brainstorming", "ce-brainstorm", "ce-ideate", "ce-plan", "deepen-plan"},
		},
		{
			Name:        "atv-pack-review",
			Category:    "review",
			Description: "Code review pack — multi-agent CE Review and document review skills for catching issues early.",
			Keywords:    []string{"atv", "code-review", "compound-engineering"},
			SkillNames:  []string{"ce-review", "document-review"},
		},
		{
			Name:        "atv-pack-shipping",
			Category:    "shipping",
			Description: "Shipping & deploy pack — Takeoff, CE Work, Land, LFG/SLFG orchestrators, and CE Compound for executing and delivering work.",
			Keywords:    []string{"atv", "shipping", "deploy", "compound-engineering", "orchestrator"},
			SkillNames: []string{
				"ce-compound",
				"ce-compound-refresh",
				"ce-work",
				"land",
				"lfg",
				"slfg",
				"takeoff",
			},
		},
		{
			Name:        "atv-pack-security",
			Category:    "security",
			Description: "Security pack — unified ATV Security skill that audits both agentic configuration and application source code (OWASP Top 10 + STRIDE).",
			Keywords:    []string{"atv", "security", "owasp", "stride"},
			SkillNames:  []string{"atv-security"},
		},
		{
			Name:        "atv-pack-quality",
			Category:    "quality",
			Description: "Quality pack — Unslop and Ralph Loop for tightening up code and iterating on solutions.",
			Keywords:    []string{"atv", "quality", "ralph-loop"},
			SkillNames:  []string{"ralph-loop", "unslop"},
		},
		{
			Name:        "atv-pack-guidelines",
			Category:    "guidelines",
			Description: "Coding guidelines pack — Karpathy Guidelines (behavioral) plus Autoresearch (autonomous experimentation loop).",
			Keywords:    []string{"atv", "guidelines", "karpathy", "autoresearch"},
			SkillNames:  []string{"autoresearch", "karpathy-guidelines"},
		},
		{
			Name:        "atv-pack-easter-eggs",
			Category:    "easter-eggs",
			Description: "Easter eggs pack — fun extras like memeIQ for AI-powered meme generation.",
			Keywords:    []string{"atv", "easter-egg", "fun"},
			SkillNames:  []string{"meme-iq"},
		},
		{
			Name:        "atv-pack-learning",
			Category:    "learning",
			Description: "Learning pipeline pack — Learn, Instincts, Evolve, and Observe for compounding institutional knowledge across sessions.",
			Keywords:    []string{"atv", "learning", "memory", "compound-engineering"},
			SkillNames:  []string{"evolve", "instincts", "learn", "observe"},
		},
		{
			Name:        "atv-pack-maintenance",
			Category:    "maintenance",
			Description: "Maintenance & health pack — atv-doctor (diagnose ATV install health) and atv-update (update marketplace plugins; advisory for project scaffold). Keep your ATV install healthy and current.",
			Keywords:    []string{"atv", "maintenance", "doctor", "update", "health"},
			SkillNames:  []string{"atv-doctor", "atv-update"},
		},
	}
}

// MiscSkills are skills that exist in pkg/scaffold/templates/skills/
// but are not assigned to any pack. They still ship as their own
// atv-skill-* plugins and are bundled into atv-everything.
//
// Today: setup (project bootstrap), feature-video (demo recording),
// resolve_todo_parallel (orchestrator helper), test-browser
// (orchestrator helper). These are intentionally pack-less because
// they're internal helpers rather than user-facing workflows.
var MiscSkills = []string{
	"feature-video",
	"resolve_todo_parallel",
	"setup",
	"test-browser",
}
