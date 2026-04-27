package gstack

// Category constants for function-based TUI grouping.
const (
	CategoryPlanning      = "planning"
	CategoryReview        = "review"
	CategoryQATesting     = "qa-testing"
	CategorySecurity      = "security"
	CategoryShipping      = "shipping"
	CategorySafety        = "safety"
	CategoryDebugging     = "debugging"
	CategoryRetrospective = "retrospective"
	CategoryGuidelines    = "guidelines"
	CategoryEasterEgg     = "easter-egg"
)

// GstackSkill describes a single gstack skill with its metadata.
type GstackSkill struct {
	Name            string
	Dir             string // directory name in gstack repo (e.g., "review", "qa")
	Category        string
	RequiresRuntime bool // true if skill needs bun/browser to function
	Description     string
}

// AllSkills returns the full catalog of gstack skills.
func AllSkills() []GstackSkill {
	return []GstackSkill{
		// Planning
		{Name: "Office Hours", Dir: "office-hours", Category: CategoryPlanning, Description: "YC-style forcing questions to reframe your product before coding"},
		{Name: "Plan CEO Review", Dir: "plan-ceo-review", Category: CategoryPlanning, Description: "Rethink the problem; find the 10-star product"},
		{Name: "Plan Eng Review", Dir: "plan-eng-review", Category: CategoryPlanning, Description: "Lock architecture, data flow, diagrams, edge cases, tests"},
		{Name: "Plan Design Review", Dir: "plan-design-review", Category: CategoryPlanning, Description: "Rate design dimensions 0-10, edit plan to improve"},
		{Name: "Design Consultation", Dir: "design-consultation", Category: CategoryPlanning, Description: "Build a complete design system from scratch"},
		{Name: "Autoplan", Dir: "autoplan", Category: CategoryPlanning, Description: "One command, fully reviewed plan with encoded decision principles"},

		// Review
		{Name: "Review", Dir: "review", Category: CategoryReview, Description: "Staff-level code review; auto-fix obvious issues, flag gaps"},
		{Name: "Design Review", Dir: "design-review", Category: CategoryReview, Description: "Design audit then fix what it finds with atomic commits"},
		{Name: "Design Shotgun", Dir: "design-shotgun", Category: CategoryReview, Description: "Generate multiple AI design variants and comparison board"},
		{Name: "Codex", Dir: "codex", Category: CategoryReview, Description: "Independent code review from OpenAI Codex CLI"},

		// QA & Testing
		{Name: "QA", Dir: "qa", Category: CategoryQATesting, RequiresRuntime: true, Description: "Test app in real browser, find and fix bugs with regression tests"},
		{Name: "QA Only", Dir: "qa-only", Category: CategoryQATesting, RequiresRuntime: true, Description: "Same QA methodology, report only without code changes"},
		{Name: "Benchmark", Dir: "benchmark", Category: CategoryQATesting, RequiresRuntime: true, Description: "Baseline page load times, Core Web Vitals, resource sizes"},
		{Name: "Browse", Dir: "browse", Category: CategoryQATesting, RequiresRuntime: true, Description: "Give the agent eyes via real Chromium browser"},

		// Security
		{Name: "CSO", Dir: "cso", Category: CategorySecurity, Description: "OWASP Top 10 + STRIDE threat model with zero false-positive noise"},

		// Shipping
		{Name: "Ship", Dir: "ship", Category: CategoryShipping, Description: "Sync main, run tests, audit coverage, push, open PR"},
		{Name: "Land and Deploy", Dir: "land-and-deploy", Category: CategoryShipping, Description: "Merge PR, wait for CI and deploy, verify production health"},
		{Name: "Canary", Dir: "canary", Category: CategoryShipping, Description: "Post-deploy monitoring for console errors and performance regressions"},
		{Name: "Document Release", Dir: "document-release", Category: CategoryShipping, Description: "Update all project docs to match what was shipped"},

		// Safety
		{Name: "Careful", Dir: "careful", Category: CategorySafety, Description: "Warn before destructive commands (rm -rf, DROP TABLE, force-push)"},
		{Name: "Freeze", Dir: "freeze", Category: CategorySafety, Description: "Restrict file edits to one directory while debugging"},
		{Name: "Guard", Dir: "guard", Category: CategorySafety, Description: "Careful + Freeze combined; maximum safety for prod work"},
		{Name: "Unfreeze", Dir: "unfreeze", Category: CategorySafety, Description: "Remove the freeze boundary"},

		// Debugging
		{Name: "Investigate", Dir: "investigate", Category: CategoryDebugging, Description: "Systematic root-cause debugging; no fixes without investigation"},

		// Retrospective
		{Name: "Retro", Dir: "retro", Category: CategoryRetrospective, Description: "Team-aware weekly retro with per-person breakdowns and trends"},
	}
}

// SkillsByCategory groups all gstack skills by their functional category.
func SkillsByCategory() map[string][]GstackSkill {
	result := make(map[string][]GstackSkill)
	for _, s := range AllSkills() {
		result[s.Category] = append(result[s.Category], s)
	}
	return result
}

// FilterSkills returns only the skills whose directories are in the selected set,
// optionally excluding skills that require runtime when runtime is unavailable.
func FilterSkills(selected []string, hasRuntime bool) []GstackSkill {
	selectedSet := make(map[string]bool)
	for _, s := range selected {
		selectedSet[s] = true
	}

	var result []GstackSkill
	for _, skill := range AllSkills() {
		if !selectedSet[skill.Dir] {
			continue
		}
		if skill.RequiresRuntime && !hasRuntime {
			continue
		}
		result = append(result, skill)
	}
	return result
}

// AllCategories returns the ordered list of category keys for TUI display.
func AllCategories() []string {
	return []string{
		CategoryPlanning,
		CategoryReview,
		CategoryQATesting,
		CategorySecurity,
		CategoryShipping,
		CategorySafety,
		CategoryDebugging,
		CategoryRetrospective,
		CategoryGuidelines,
		CategoryEasterEgg,
	}
}

// CategoryLabel returns a human-readable label for a category key.
func CategoryLabel(cat string) string {
	labels := map[string]string{
		CategoryPlanning:      "📋 Planning & Design",
		CategoryReview:        "🔍 Code Review",
		CategoryQATesting:     "🧪 QA & Testing",
		CategorySecurity:      "🔒 Security",
		CategoryShipping:      "🚀 Shipping & Deploy",
		CategorySafety:        "🛡️ Safety Guardrails",
		CategoryDebugging:     "🐛 Debugging",
		CategoryRetrospective: "📊 Retrospective",
		CategoryGuidelines:    "📐 Coding Guidelines",
		CategoryEasterEgg:     "🥚 Easter Eggs",
	}
	if l, ok := labels[cat]; ok {
		return l
	}
	return cat
}
