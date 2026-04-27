package sandbox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/detect"
	"github.com/All-The-Vibes/ATV-StarterKit/pkg/gstack"
	"github.com/All-The-Vibes/ATV-StarterKit/pkg/installstate"
	"github.com/All-The-Vibes/ATV-StarterKit/pkg/scaffold"
)

func TestAutoModeNoGstack(t *testing.T) {
	// Auto mode should install ATV files only, no gstack
	sandboxDir := t.TempDir()

	// Create a fake TypeScript project
	if err := os.WriteFile(filepath.Join(sandboxDir, "tsconfig.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	env := detect.DetectEnvironment(sandboxDir)
	if env.Stack != detect.StackTypeScript {
		t.Fatalf("expected TypeScript stack, got %s", env.Stack)
	}

	catalog := scaffold.BuildCatalog(env.Stack)
	results := scaffold.WriteAll(sandboxDir, catalog)

	if len(results) == 0 {
		t.Fatal("auto mode should produce results")
	}

	// ATV files should exist
	assertFileExists(t, filepath.Join(sandboxDir, ".github", "copilot-instructions.md"))
	assertFileExists(t, filepath.Join(sandboxDir, ".github", "copilot-setup-steps.yml"))

	// gstack should NOT exist
	gstackDir := filepath.Join(sandboxDir, ".github", "skills", "gstack")
	if _, err := os.Stat(gstackDir); !os.IsNotExist(err) {
		t.Error("auto mode should NOT install gstack")
	}

	manifestPath := filepath.Join(sandboxDir, ".atv", "install-manifest.json")
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Error("auto mode should NOT write a guided install manifest")
	}
}

func TestATVFilesCreatedCorrectly(t *testing.T) {
	sandboxDir := t.TempDir()

	// Create a fake Python project
	if err := os.WriteFile(filepath.Join(sandboxDir, "pyproject.toml"), []byte("[project]"), 0644); err != nil {
		t.Fatal(err)
	}

	env := detect.DetectEnvironment(sandboxDir)
	catalog := scaffold.BuildCatalog(env.Stack)
	results := scaffold.WriteAll(sandboxDir, catalog)

	// Count created files
	created := 0
	for _, r := range results {
		if r.Status == scaffold.StatusCreated || r.Status == scaffold.StatusDirCreated {
			created++
		}
	}

	if created == 0 {
		t.Error("should have created files")
	}

	// Check key ATV files
	assertFileExists(t, filepath.Join(sandboxDir, ".github", "copilot-instructions.md"))
	assertFileExists(t, filepath.Join(sandboxDir, ".github", "copilot-setup-steps.yml"))
	assertFileExists(t, filepath.Join(sandboxDir, ".github", "copilot-mcp-config.json"))
	assertDirExists(t, filepath.Join(sandboxDir, ".github", "skills"))
	assertDirExists(t, filepath.Join(sandboxDir, ".github", "agents"))
	assertDirExists(t, filepath.Join(sandboxDir, "docs", "plans"))
	assertDirExists(t, filepath.Join(sandboxDir, "docs", "brainstorms"))
}

func TestGstackInstallMarkdownOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network-dependent test in short mode")
	}

	sandboxDir := t.TempDir()

	// Create .github/skills/ so copy has a target
	if err := os.MkdirAll(filepath.Join(sandboxDir, ".github", "skills"), 0755); err != nil {
		t.Fatal(err)
	}

	result := gstack.Install(sandboxDir, gstack.ModeMarkdownOnly)

	if result.Error != nil {
		t.Fatalf("gstack install failed: %v", result.Error)
	}

	if !result.Cloned {
		t.Error("expected gstack to be cloned")
	}

	// .gstack/ staging dir should exist
	gstackDir := filepath.Join(sandboxDir, ".gstack")
	if _, err := os.Stat(gstackDir); os.IsNotExist(err) {
		t.Error(".gstack/ staging directory should exist")
	}

	// Should not build in markdown-only mode
	if result.Built {
		t.Error("markdown-only mode should not build")
	}

	// Should have copied skill dirs to .github/skills/gstack-*
	if len(result.SkillDirs) == 0 {
		t.Error("expected at least some skill directories")
	}

	// Verify skills are at the right level for Copilot discovery
	for _, dir := range result.SkillDirs {
		skillPath := filepath.Join(sandboxDir, ".github", "skills", dir, "SKILL.md")
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			t.Errorf("expected skill file at %s", skillPath)
		}
	}
}

func TestGstackIdempotent(t *testing.T) {
	sandboxDir := t.TempDir()
	gstackDir := filepath.Join(sandboxDir, ".gstack")

	// Create a fake existing install
	if err := os.MkdirAll(gstackDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gstackDir, "SKILL.md"), []byte("# gstack"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(sandboxDir, ".github", "skills"), 0755); err != nil {
		t.Fatal(err)
	}

	result := gstack.Install(sandboxDir, gstack.ModeMarkdownOnly)
	if result.Error != nil {
		t.Fatalf("idempotent install should not error: %v", result.Error)
	}
	if result.Cloned {
		t.Error("should not clone when already installed")
	}
}

func TestInstructionsContainGstackSection(t *testing.T) {
	sandboxDir := t.TempDir()

	env := detect.DetectEnvironment(sandboxDir) // general stack
	catalog := scaffold.BuildCatalog(env.Stack)
	scaffold.WriteAll(sandboxDir, catalog)

	// Read copilot-instructions.md
	content, err := os.ReadFile(filepath.Join(sandboxDir, ".github", "copilot-instructions.md"))
	if err != nil {
		t.Fatalf("failed to read copilot-instructions.md: %v", err)
	}

	contentStr := string(content)

	// Should contain gstack section
	if !containsStr(contentStr, "gstack Skills") {
		t.Error("copilot-instructions.md should contain gstack Skills section")
	}

	// Should contain ATV override rules
	if !containsStr(contentStr, "ATV Override Rules") {
		t.Error("copilot-instructions.md should contain ATV Override Rules")
	}

	// Should contain protected artifacts
	if !containsStr(contentStr, "Protected artifacts") {
		t.Error("copilot-instructions.md should contain protected artifacts rule")
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file to exist: %s", path)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("expected directory to exist: %s", path)
		return
	}
	if !info.IsDir() {
		t.Errorf("expected %s to be a directory", path)
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// --- Integration test scenarios from the plan ---

func TestPrereqSkipProducesCorrectOutcomeAndManifest(t *testing.T) {
	// "Missing prerequisites downgrade or skip runtime-dependent capabilities
	//  with clear telemetry and manifest output"
	root := t.TempDir()

	// Simulate guided install with gstack runtime requested but build failed (prereq missing)
	manifest := installstate.InstallManifest{
		Requested: installstate.RequestedState{
			StackPacks:          []installstate.StackPack{installstate.StackPackGeneral},
			GstackDirs:          []string{"review"},
			GstackRuntime:       true,
			IncludeAgentBrowser: true,
			PresetName:          "Full",
		},
		Outcomes: []installstate.InstallOutcome{
			{Step: "Scaffolding ATV files", Status: installstate.InstallStepDone, Detail: "5 files created"},
			{
				Step: "Syncing gstack skills", Status: installstate.InstallStepWarning,
				Detail: "3 skills synced, markdown-only", Reason: "./setup failed, falling back to docs",
				Substeps: []installstate.InstallOutcome{
					{Step: "git clone", Status: installstate.InstallStepDone},
					{Step: "runtime build", Status: installstate.InstallStepWarning, Reason: "bun not found",
						SkipReason: installstate.SkipReasonPrereqMissing},
					{Step: "copy skills", Status: installstate.InstallStepDone, Detail: "3 skill dirs"},
				},
			},
			{
				Step: "Installing agent-browser", Status: installstate.InstallStepWarning,
				Reason: "npm not found",
				Substeps: []installstate.InstallOutcome{
					{Step: "npm install", Status: installstate.InstallStepWarning, Reason: "npm not found",
						SkipReason: installstate.SkipReasonPrereqMissing},
				},
			},
		},
	}

	if err := installstate.WriteManifest(root, manifest); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}

	read, err := installstate.ReadManifest(root)
	if err != nil {
		t.Fatalf("ReadManifest() error = %v", err)
	}

	// Verify substeps survive round-trip
	if len(read.Outcomes) != 3 {
		t.Fatalf("expected 3 outcomes, got %d", len(read.Outcomes))
	}
	gstackOutcome := read.Outcomes[1]
	if len(gstackOutcome.Substeps) != 3 {
		t.Fatalf("expected 3 gstack substeps, got %d", len(gstackOutcome.Substeps))
	}
	if gstackOutcome.Substeps[1].SkipReason != installstate.SkipReasonPrereqMissing {
		t.Fatalf("expected prereq-missing skip reason, got %q", gstackOutcome.Substeps[1].SkipReason)
	}

	// Verify recommendations prioritize the install issue
	recs := installstate.BuildRecommendations(root, read)
	if len(recs) == 0 || recs[0].ID != "fix-install-issues" {
		t.Fatalf("expected fix-install-issues as first recommendation, got %+v", recs)
	}
}

func TestPartialInstallWritesValidManifest(t *testing.T) {
	// "Partial install success still writes a valid manifest and useful completion summary"
	root := t.TempDir()

	manifest := installstate.InstallManifest{
		Requested: installstate.RequestedState{
			StackPacks:          []installstate.StackPack{installstate.StackPackGeneral, installstate.StackPackTypeScript},
			GstackDirs:          []string{"review"},
			IncludeAgentBrowser: true,
			PresetName:          "Full",
		},
		Outcomes: []installstate.InstallOutcome{
			{Step: "Scaffolding ATV files", Status: installstate.InstallStepDone, Detail: "created"},
			{Step: "Syncing gstack skills", Status: installstate.InstallStepFailed, Reason: "git clone failed"},
			{Step: "Installing agent-browser", Status: installstate.InstallStepFailed, Reason: "npm not found"},
		},
	}

	if err := installstate.WriteManifest(root, manifest); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}

	read, err := installstate.ReadManifest(root)
	if err != nil {
		t.Fatalf("ReadManifest() error = %v", err)
	}

	if read.Version != installstate.ManifestVersion {
		t.Fatalf("expected manifest version %d, got %d", installstate.ManifestVersion, read.Version)
	}
	if read.RerunPolicy != installstate.RerunPolicyAdditiveOnly {
		t.Fatalf("expected additive-only rerun policy, got %q", read.RerunPolicy)
	}

	summary := installstate.SummarizeOutcomes(read.Outcomes)
	if summary.Done != 1 || summary.Failed != 2 {
		t.Fatalf("expected 1 done + 2 failed, got %+v", summary)
	}

	// Even with failures, recommendations should be generated
	recs := installstate.BuildRecommendations(root, read)
	if len(recs) == 0 {
		t.Fatal("expected recommendations even with partial install")
	}
	if recs[0].ID != "fix-install-issues" {
		t.Fatalf("expected fix-install-issues first, got %s", recs[0].ID)
	}
}

func TestRerunPreservesManifestAdditively(t *testing.T) {
	// "Re-running guided mode after a prior install preserves or updates state"
	root := t.TempDir()

	// First run
	first := installstate.InstallManifest{
		Requested: installstate.RequestedState{
			StackPacks: []installstate.StackPack{installstate.StackPackGeneral},
			PresetName: "Starter",
		},
		Outcomes: []installstate.InstallOutcome{
			{Step: "Scaffolding ATV files", Status: installstate.InstallStepDone, Detail: "10 files created"},
		},
	}
	if err := installstate.WriteManifest(root, first); err != nil {
		t.Fatalf("first WriteManifest() error = %v", err)
	}

	read1, err := installstate.ReadManifest(root)
	if err != nil {
		t.Fatalf("first ReadManifest() error = %v", err)
	}
	gen1 := read1.GeneratedAt

	// Second run (simulates rerun with more packs)
	second := installstate.InstallManifest{
		Requested: installstate.RequestedState{
			StackPacks:    []installstate.StackPack{installstate.StackPackGeneral, installstate.StackPackTypeScript},
			GstackDirs:    []string{"review"},
			GstackRuntime: false,
			PresetName:    "Pro",
		},
		Outcomes: []installstate.InstallOutcome{
			{Step: "Scaffolding ATV files", Status: installstate.InstallStepDone,
				Detail: "2 files created",
				Substeps: []installstate.InstallOutcome{
					{Step: "copilot-instructions.md", Status: installstate.InstallStepSkipped,
						SkipReason: installstate.SkipReasonAlreadyInstalled},
					{Step: "tsconfig-agent.md", Status: installstate.InstallStepDone, Detail: "created"},
				}},
			{Step: "Syncing gstack skills", Status: installstate.InstallStepDone, Detail: "3 skills synced"},
		},
	}
	if err := installstate.WriteManifest(root, second); err != nil {
		t.Fatalf("second WriteManifest() error = %v", err)
	}

	read2, err := installstate.ReadManifest(root)
	if err != nil {
		t.Fatalf("second ReadManifest() error = %v", err)
	}

	// Manifest should reflect the latest run
	if len(read2.Requested.StackPacks) != 2 {
		t.Fatalf("expected 2 stack packs after rerun, got %d", len(read2.Requested.StackPacks))
	}
	if read2.Requested.PresetName != "Pro" {
		t.Fatalf("expected Pro preset after rerun, got %q", read2.Requested.PresetName)
	}
	if !read2.GeneratedAt.After(gen1) {
		t.Fatal("second manifest should have a later timestamp")
	}
	if read2.RerunPolicy != installstate.RerunPolicyAdditiveOnly {
		t.Fatalf("rerun policy should remain additive-only, got %q", read2.RerunPolicy)
	}
}

func TestInstallSnapshotMatchesManifestAndRepoMemory(t *testing.T) {
	// "Install snapshot recommendation output matches manifest + repo memory for empty, partial, and mature repos"
	root := t.TempDir()

	// --- Empty repo (no manifest, no docs) ---
	snapshot, err := installstate.BuildInstallSnapshot(root)
	if err != nil {
		t.Fatalf("empty snapshot error = %v", err)
	}
	if snapshot.HasManifest {
		t.Fatal("empty repo should not have manifest")
	}
	if len(snapshot.Recommendations) == 0 || snapshot.Recommendations[0].ID != "start-brainstorm" {
		t.Fatalf("empty repo should recommend start-brainstorm, got %+v", snapshot.Recommendations)
	}

	// --- Partial repo (has brainstorm and manifest with warnings) ---
	mustWriteFile(t, filepath.Join(root, "docs", "brainstorms", "idea.md"), "# Idea\n")
	manifest := installstate.InstallManifest{
		Requested: installstate.RequestedState{
			StackPacks:          []installstate.StackPack{installstate.StackPackGeneral, installstate.StackPackRails},
			GstackDirs:          []string{"review"},
			IncludeAgentBrowser: true,
			PresetName:          "Full",
		},
		Outcomes: []installstate.InstallOutcome{
			{Step: "Scaffolding ATV files", Status: installstate.InstallStepDone},
			{Step: "Syncing gstack skills", Status: installstate.InstallStepWarning, Reason: "setup failed"},
			{Step: "Installing agent-browser", Status: installstate.InstallStepFailed, Reason: "npm not found"},
		},
	}
	if err := installstate.WriteManifest(root, manifest); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}

	snapshot, err = installstate.BuildInstallSnapshot(root)
	if err != nil {
		t.Fatalf("partial snapshot error = %v", err)
	}
	if !snapshot.HasManifest {
		t.Fatal("should find manifest")
	}
	if snapshot.OutcomeSummary.Done != 1 || snapshot.OutcomeSummary.Warning != 1 || snapshot.OutcomeSummary.Failed != 1 {
		t.Fatalf("unexpected outcome summary: %+v", snapshot.OutcomeSummary)
	}
	if len(snapshot.Recommendations) == 0 || snapshot.Recommendations[0].ID != "fix-install-issues" {
		t.Fatalf("partial repo should prioritize fix-install-issues, got %+v", snapshot.Recommendations)
	}

	// --- Mature repo (manifest clean, brainstorms, plans with unchecked items) ---
	mustWriteFile(t, filepath.Join(root, "docs", "plans", "work.md"), "status: active\n- [ ] finish feature\n")
	matureManifest := installstate.InstallManifest{
		Requested: installstate.RequestedState{
			StackPacks: []installstate.StackPack{installstate.StackPackGeneral, installstate.StackPackRails},
			GstackDirs: []string{"review"},
			PresetName: "Pro",
		},
		Outcomes: []installstate.InstallOutcome{
			{Step: "Scaffolding ATV files", Status: installstate.InstallStepDone},
			{Step: "Syncing gstack skills", Status: installstate.InstallStepDone},
		},
	}
	if err := installstate.WriteManifest(root, matureManifest); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}

	snapshot, err = installstate.BuildInstallSnapshot(root)
	if err != nil {
		t.Fatalf("mature snapshot error = %v", err)
	}
	if snapshot.OutcomeSummary.Failed != 0 && snapshot.OutcomeSummary.Warning != 0 {
		t.Fatalf("mature repo should have clean outcomes: %+v", snapshot.OutcomeSummary)
	}
	// Should recommend executing the active plan, not fixing install issues
	foundPlan := false
	for _, rec := range snapshot.Recommendations {
		if rec.ID == "execute-active-plan" {
			foundPlan = true
		}
		if rec.ID == "fix-install-issues" {
			t.Fatal("mature repo with clean install should not recommend fix-install-issues")
		}
	}
	if !foundPlan {
		t.Fatalf("mature repo with active plan should recommend execute-active-plan, got %+v", snapshot.Recommendations)
	}

	// Verify determinism
	snap1, _ := installstate.BuildInstallSnapshot(root)
	snap2, _ := installstate.BuildInstallSnapshot(root)
	if len(snap1.Recommendations) != len(snap2.Recommendations) {
		t.Fatal("snapshot should be deterministic")
	}
	for i := range snap1.Recommendations {
		if snap1.Recommendations[i].ID != snap2.Recommendations[i].ID {
			t.Fatalf("recommendation order differs: %+v vs %+v", snap1.Recommendations, snap2.Recommendations)
		}
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// =============================================================================
// Comprehensive E2E: Full guided install lifecycle
// =============================================================================

// TestE2EFullGuidedInstallLifecycle simulates a complete guided install:
//   - detect stack → build filtered catalog → scaffold files → write manifest
//   - verify every expected file is installed in the right location
//   - verify memory index picks up installed intelligence
//   - verify install snapshot reads the correct state
//   - verify determinism across repeated calls
func TestE2EFullGuidedInstallLifecycle(t *testing.T) {
	root := t.TempDir()

	// --- Step 1: Set up a Rails project ---
	mustWriteFile(t, filepath.Join(root, "Gemfile"), `gem "rails"`)
	mustWriteFile(t, filepath.Join(root, "config", "routes.rb"), `Rails.application.routes.draw { }`)

	// --- Step 2: Detect environment ---
	env := detect.DetectEnvironment(root)
	if env.Stack != detect.StackRails {
		t.Fatalf("expected Rails stack, got %s", env.Stack)
	}
	if len(env.DetectedPacks) == 0 {
		t.Fatal("should detect at least one pack")
	}

	// --- Step 3: Build multi-pack filtered catalog (simulating guided "Pro" preset) ---
	selectedPacks := []installstate.StackPack{installstate.StackPackGeneral, installstate.StackPackRails}
	selectedLayers := []string{
		"copilot-instructions", "setup-steps", "mcp-servers",
		"core-skills", "orchestrators", "universal-agents", "stack-agents",
		"file-instructions", "vscode-extensions", "docs-structure",
	}
	catalog := scaffold.BuildFilteredCatalogForPacks(selectedPacks, env.Stack, selectedLayers)

	if len(catalog) == 0 {
		t.Fatal("filtered catalog should produce components")
	}

	// --- Step 4: Write all scaffold files ---
	results := scaffold.WriteAll(root, catalog)
	summary := scaffold.SummarizeResults(results)

	if summary.Created == 0 {
		t.Fatal("scaffold should create files")
	}
	if summary.Failed > 0 {
		t.Fatalf("scaffold should not fail: %d failures", summary.Failed)
	}

	// --- Step 5: Verify core Copilot lifecycle hooks installed ---
	t.Run("hook1_system_instructions", func(t *testing.T) {
		path := filepath.Join(root, ".github", "copilot-instructions.md")
		assertFileExists(t, path)
		content, _ := os.ReadFile(path)
		if !strings.Contains(string(content), "ATV Override Rules") {
			t.Error("instructions should contain ATV Override Rules")
		}
	})

	t.Run("hook2_setup_steps", func(t *testing.T) {
		assertFileExists(t, filepath.Join(root, ".github", "copilot-setup-steps.yml"))
	})

	t.Run("hook3_mcp_config", func(t *testing.T) {
		path := filepath.Join(root, ".github", "copilot-mcp-config.json")
		assertFileExists(t, path)
		content, _ := os.ReadFile(path)
		var parsed map[string]interface{}
		if err := json.Unmarshal(content, &parsed); err != nil {
			t.Fatalf("MCP config should be valid JSON: %v", err)
		}
	})

	t.Run("hook4_skills_installed", func(t *testing.T) {
		skillsDir := filepath.Join(root, ".github", "skills")
		assertDirExists(t, skillsDir)
		// Verify at least core skills exist
		for _, skill := range []string{"brainstorming", "ce-brainstorm", "ce-plan", "ce-review", "ce-work"} {
			skillPath := filepath.Join(skillsDir, skill, "SKILL.md")
			assertFileExists(t, skillPath)
		}
		// Verify learning pipeline skills exist
		for _, skill := range []string{"learn", "instincts", "evolve", "observe"} {
			skillPath := filepath.Join(skillsDir, skill, "SKILL.md")
			assertFileExists(t, skillPath)
		}
		// Verify behavioral guidelines skill exists
		karpathyPath := filepath.Join(skillsDir, "karpathy-guidelines", "SKILL.md")
		assertFileExists(t, karpathyPath)
		// Verify autoresearch skill exists (sourced from github/awesome-copilot, MIT)
		autoresearchPath := filepath.Join(skillsDir, "autoresearch", "SKILL.md")
		assertFileExists(t, autoresearchPath)
	})

	t.Run("hook5_agents_installed", func(t *testing.T) {
		agentsDir := filepath.Join(root, ".github", "agents")
		assertDirExists(t, agentsDir)
		// Universal agents
		for _, agent := range []string{"code-simplicity-reviewer", "security-sentinel", "architecture-strategist"} {
			assertFileExists(t, filepath.Join(agentsDir, agent+".agent.md"))
		}
		// Stack-specific agents for Rails
		for _, agent := range []string{"kieran-rails-reviewer", "dhh-rails-reviewer", "data-integrity-guardian"} {
			assertFileExists(t, filepath.Join(agentsDir, agent+".agent.md"))
		}
	})

	t.Run("hook6_file_instructions", func(t *testing.T) {
		assertFileExists(t, filepath.Join(root, ".github", "rails.instructions.md"))
	})

	t.Run("vscode_extensions", func(t *testing.T) {
		path := filepath.Join(root, ".vscode", "extensions.json")
		assertFileExists(t, path)
		content, _ := os.ReadFile(path)
		var parsed map[string]interface{}
		if err := json.Unmarshal(content, &parsed); err != nil {
			t.Fatalf("extensions.json should be valid JSON: %v", err)
		}
	})

	t.Run("docs_structure", func(t *testing.T) {
		for _, dir := range []string{"docs/plans", "docs/brainstorms", "docs/solutions"} {
			assertDirExists(t, filepath.Join(root, dir))
		}
		// Learning pipeline directories
		assertDirExists(t, filepath.Join(root, ".atv", "instincts"))
	})

	t.Run("observer_hooks_in_lifecycle", func(t *testing.T) {
		assertFileExists(t, filepath.Join(root, ".github", "hooks", "copilot-hooks.json"))
		assertFileExists(t, filepath.Join(root, ".github", "hooks", "scripts", "observe.js"))
	})

	// --- Step 6: Write install manifest (simulating guided mode completion) ---
	substeps := scaffold.ResultsToSubsteps(results)
	manifest := installstate.InstallManifest{
		Requested: installstate.RequestedState{
			StackPacks:          selectedPacks,
			ATVLayers:           selectedLayers,
			GstackDirs:          []string{"review", "plan"},
			GstackRuntime:       false,
			IncludeAgentBrowser: true,
			PresetName:          "Pro",
		},
		Outcomes: []installstate.InstallOutcome{
			{
				Step:     "Scaffolding ATV files",
				Status:   installstate.InstallStepDone,
				Detail:   summary.Detail(),
				Duration: "340ms",
				Substeps: substeps,
			},
			{
				Step:   "Syncing gstack skills",
				Status: installstate.InstallStepWarning,
				Detail: "2 skills synced, markdown-only",
				Reason: "setup script failed, fell back to doc generation",
				Substeps: []installstate.InstallOutcome{
					{Step: "git clone", Status: installstate.InstallStepDone, Detail: "shallow clone"},
					{Step: "runtime build", Status: installstate.InstallStepWarning,
						Reason: "bun not found", SkipReason: installstate.SkipReasonPrereqMissing},
					{Step: "copy skills", Status: installstate.InstallStepDone, Detail: "2 skill dirs"},
				},
			},
			{
				Step:   "Installing agent-browser",
				Status: installstate.InstallStepDone,
				Detail: "CLI ready, skill copied",
				Substeps: []installstate.InstallOutcome{
					{Step: "npm install", Status: installstate.InstallStepDone, Detail: "agent-browser CLI available"},
					{Step: "copy SKILL.md", Status: installstate.InstallStepDone, Detail: "skill registered"},
				},
			},
		},
	}
	manifest.Recommendations = installstate.BuildRecommendations(root, manifest)

	if err := installstate.WriteManifest(root, manifest); err != nil {
		t.Fatalf("WriteManifest() error = %v", err)
	}

	// Simulate gstack staging and agent-browser skill presence (as if install ran)
	if err := os.MkdirAll(filepath.Join(root, ".gstack"), 0o755); err != nil {
		t.Fatalf("MkdirAll(.gstack) error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".github", "skills", "agent-browser"), 0o755); err != nil {
		t.Fatalf("MkdirAll(agent-browser) error = %v", err)
	}
	mustWriteFile(t, filepath.Join(root, ".github", "skills", "agent-browser", "SKILL.md"), "# agent-browser")

	// --- Step 7: Verify manifest round-trip ---
	t.Run("manifest_roundtrip", func(t *testing.T) {
		read, err := installstate.ReadManifest(root)
		if err != nil {
			t.Fatalf("ReadManifest() error = %v", err)
		}
		if read.Version != installstate.ManifestVersion {
			t.Fatalf("version = %d, want %d", read.Version, installstate.ManifestVersion)
		}
		if read.RerunPolicy != installstate.RerunPolicyAdditiveOnly {
			t.Fatalf("rerun policy = %q, want additive-only", read.RerunPolicy)
		}
		if read.Requested.PresetName != "Pro" {
			t.Fatalf("preset = %q, want Pro", read.Requested.PresetName)
		}
		if len(read.Outcomes) != 3 {
			t.Fatalf("expected 3 outcomes, got %d", len(read.Outcomes))
		}
		// Substeps survive serialization
		if len(read.Outcomes[0].Substeps) == 0 {
			t.Fatal("scaffold substeps should survive JSON round-trip")
		}
		if len(read.Outcomes[1].Substeps) != 3 {
			t.Fatalf("gstack substeps = %d, want 3", len(read.Outcomes[1].Substeps))
		}
		// Skip reason survives
		if read.Outcomes[1].Substeps[1].SkipReason != installstate.SkipReasonPrereqMissing {
			t.Fatalf("skip reason = %q, want prereq-missing", read.Outcomes[1].Substeps[1].SkipReason)
		}
	})

	// --- Step 8: Verify memory index ---
	t.Run("memory_index_classification", func(t *testing.T) {
		state := installstate.ScanRepoState(root)

		if !state.HasCopilotInstructions {
			t.Error("should detect copilot-instructions.md")
		}
		if state.InstalledAgents == 0 {
			t.Error("should detect installed agents")
		}
		if state.InstalledSkills == 0 {
			t.Error("should detect installed skills")
		}
		if !state.HasGstackStaging {
			t.Error("should detect .gstack/ staging directory")
		}
		if !state.HasAgentBrowserSkill {
			t.Error("should detect agent-browser skill")
		}
	})

	// --- Step 9: Verify install snapshot ---
	t.Run("install_snapshot", func(t *testing.T) {
		snapshot, err := installstate.BuildInstallSnapshot(root)
		if err != nil {
			t.Fatalf("BuildInstallSnapshot() error = %v", err)
		}
		if !snapshot.HasManifest {
			t.Fatal("should find manifest")
		}
		if snapshot.Requested.PresetName != "Pro" {
			t.Fatalf("preset = %q, want Pro", snapshot.Requested.PresetName)
		}
		labels := snapshot.StackPackLabels()
		if len(labels) != 2 || labels[1] != "Rails" {
			t.Fatalf("stack pack labels = %v, want [General, Rails]", labels)
		}
		if !snapshot.HasGstack() {
			t.Error("should report gstack requested")
		}
		if !snapshot.HasAgentBrowser() {
			t.Error("should report agent-browser requested")
		}
		if snapshot.OutcomeSummary.Done != 2 || snapshot.OutcomeSummary.Warning != 1 {
			t.Fatalf("outcomes = %+v, want 2 done + 1 warning", snapshot.OutcomeSummary)
		}
		if len(snapshot.Recommendations) == 0 {
			t.Fatal("should have recommendations")
		}
	})

	// --- Step 10: Verify memory index detail ---
	t.Run("memory_index_detail", func(t *testing.T) {
		state := installstate.ScanRepoState(root)
		snapshot, _ := installstate.BuildInstallSnapshot(root)

		// Verify manifest data flows through to snapshot
		if !snapshot.HasManifest {
			t.Fatal("should have manifest")
		}
		if snapshot.Requested.PresetName != "Pro" {
			t.Fatalf("preset = %q, want Pro", snapshot.Requested.PresetName)
		}
		if snapshot.OutcomeSummary.Done != 2 || snapshot.OutcomeSummary.Warning != 1 {
			t.Fatalf("outcomes = %+v, want 2 done + 1 warning", snapshot.OutcomeSummary)
		}

		// Verify all memory index fields populated
		if !state.HasCopilotInstructions {
			t.Error("should detect copilot-instructions.md")
		}
		if state.InstalledAgents == 0 {
			t.Error("should detect installed agents")
		}
		if state.InstalledSkills == 0 {
			t.Error("should detect installed skills")
		}
		if !state.HasGstackStaging {
			t.Error("should detect gstack staging")
		}
		if !state.HasAgentBrowserSkill {
			t.Error("should detect agent-browser skill")
		}

		// Recommendations determinism: call 10 times, always same order
		first := snapshot.Recommendations
		for i := 0; i < 10; i++ {
			check, _ := installstate.BuildInstallSnapshot(root)
			if len(check.Recommendations) != len(first) {
				t.Fatalf("iteration %d: length differs", i)
			}
			for j := range check.Recommendations {
				if check.Recommendations[j].ID != first[j].ID {
					t.Fatalf("iteration %d, index %d: order differs", i, j)
				}
			}
		}

		// First recommendation should be fix-install-issues (we have a warning)
		if len(snapshot.Recommendations) == 0 || snapshot.Recommendations[0].ID != "fix-install-issues" {
			t.Fatalf("first rec = %+v, want fix-install-issues", snapshot.Recommendations)
		}
	})
}

// =============================================================================
// Learning Pipeline E2E: Verify observer hooks, skills, and instinct support
// =============================================================================

func TestE2ELearningPipelineInstalled(t *testing.T) {
	root := t.TempDir()

	// Build full catalog (general stack) — should include learning pipeline
	env := detect.DetectEnvironment(root)
	catalog := scaffold.BuildCatalog(env.Stack)
	results := scaffold.WriteAll(root, catalog)

	summary := scaffold.SummarizeResults(results)
	if summary.Failed > 0 {
		t.Fatalf("scaffold should not fail: %d failures", summary.Failed)
	}

	// --- Observer hooks installed ---
	t.Run("observer_hooks_config", func(t *testing.T) {
		path := filepath.Join(root, ".github", "hooks", "copilot-hooks.json")
		assertFileExists(t, path)
		content, _ := os.ReadFile(path)
		var parsed map[string]interface{}
		if err := json.Unmarshal(content, &parsed); err != nil {
			t.Fatalf("copilot-hooks.json should be valid JSON: %v", err)
		}
		hooks, ok := parsed["hooks"]
		if !ok {
			t.Fatal("copilot-hooks.json should contain 'hooks' key")
		}
		hooksMap, ok := hooks.(map[string]interface{})
		if !ok {
			t.Fatal("hooks should be a JSON object")
		}
		// Verify session hook types are configured (trimmed to reduce noise)
		for _, hookType := range []string{"sessionStart", "sessionEnd"} {
			if _, exists := hooksMap[hookType]; !exists {
				t.Errorf("missing hook type: %s", hookType)
			}
		}
	})

	t.Run("observer_script", func(t *testing.T) {
		path := filepath.Join(root, ".github", "hooks", "scripts", "observe.js")
		assertFileExists(t, path)
		content, _ := os.ReadFile(path)
		if !strings.Contains(string(content), "observations.jsonl") {
			t.Error("observe.js should reference observations.jsonl")
		}
		if !strings.Contains(string(content), "updateInstinctConfidence") {
			t.Error("observe.js should contain updateInstinctConfidence function")
		}
	})

	// --- Learning skills installed ---
	t.Run("learning_skills", func(t *testing.T) {
		skillsDir := filepath.Join(root, ".github", "skills")
		for _, skill := range []string{"learn", "instincts", "evolve", "observe"} {
			skillPath := filepath.Join(skillsDir, skill, "SKILL.md")
			assertFileExists(t, skillPath)
		}
	})

	// --- Pattern observer agent installed ---
	t.Run("pattern_observer_agent", func(t *testing.T) {
		assertFileExists(t, filepath.Join(root, ".github", "agents", "pattern-observer.agent.md"))
	})

	// --- Instructions mention learning pipeline ---
	t.Run("instructions_mention_learning", func(t *testing.T) {
		content, _ := os.ReadFile(filepath.Join(root, ".github", "copilot-instructions.md"))
		contentStr := string(content)
		if !strings.Contains(contentStr, "Continuous Learning Pipeline") {
			t.Error("copilot-instructions.md should mention Continuous Learning Pipeline")
		}
		if !strings.Contains(contentStr, "/learn") {
			t.Error("copilot-instructions.md should mention /learn command")
		}
		if !strings.Contains(contentStr, "/instincts") {
			t.Error("copilot-instructions.md should mention /instincts command")
		}
		if !strings.Contains(contentStr, "/evolve") {
			t.Error("copilot-instructions.md should mention /evolve command")
		}
	})

	// --- Instinct directories scaffolded ---
	t.Run("instinct_directories", func(t *testing.T) {
		assertDirExists(t, filepath.Join(root, ".atv", "instincts"))
	})
}

func TestLearningPipelineRepoState(t *testing.T) {
	root := t.TempDir()

	// --- No instincts, no observations ---
	state := installstate.ScanRepoState(root)
	if state.HasInstincts {
		t.Error("should not have instincts in empty repo")
	}
	if state.InstinctCount != 0 {
		t.Errorf("instinct count should be 0, got %d", state.InstinctCount)
	}
	if state.ObservationCount != 0 {
		t.Errorf("observation count should be 0, got %d", state.ObservationCount)
	}

	// --- Create observations ---
	mustWriteFile(t, filepath.Join(root, ".atv", "observations.jsonl"),
		`{"ts":"2026-04-06T10:00:00Z","hook":"preToolUse","tool":"Edit"}
{"ts":"2026-04-06T10:01:00Z","hook":"postToolUse","tool":"Edit"}
{"ts":"2026-04-06T10:02:00Z","hook":"preToolUse","tool":"Bash"}
`)
	state = installstate.ScanRepoState(root)
	if state.ObservationCount != 3 {
		t.Errorf("observation count should be 3, got %d", state.ObservationCount)
	}

	// --- Create instincts ---
	mustWriteFile(t, filepath.Join(root, ".atv", "instincts", "project.yaml"),
		`instincts:
  - id: always-wrap-errors
    trigger: "when returning an error"
    behavior: "wrap with fmt.Errorf using %w"
    confidence: 0.85
    domain: error-handling
    observations: 12
  - id: table-driven-tests
    trigger: "when writing tests"
    behavior: "use table-driven test pattern"
    confidence: 0.7
    domain: testing
    observations: 8
`)
	state = installstate.ScanRepoState(root)
	if !state.HasInstincts {
		t.Error("should have instincts")
	}
	if state.InstinctCount != 2 {
		t.Errorf("instinct count should be 2, got %d", state.InstinctCount)
	}

	// --- Observer hooks detection ---
	mustWriteFile(t, filepath.Join(root, ".github", "hooks", "copilot-hooks.json"), `{"hooks":{}}`)
	state = installstate.ScanRepoState(root)
	if !state.HasObserverHooks {
		t.Error("should detect observer hooks")
	}
}

func TestLearningPipelineRecommendations(t *testing.T) {
	root := t.TempDir()

	// Setup: install manifest with clean outcomes, observer hooks, observations, no instincts
	mustWriteFile(t, filepath.Join(root, ".github", "hooks", "copilot-hooks.json"), `{"hooks":{}}`)
	mustWriteFile(t, filepath.Join(root, ".github", "copilot-instructions.md"), "# Test")
	mustWriteFile(t, filepath.Join(root, ".github", "copilot-mcp-config.json"), `{"servers":{}}`)

	// Write 15 observations (exceeds threshold of 10)
	var obsLines string
	for i := 0; i < 15; i++ {
		obsLines += `{"ts":"2026-04-06T10:00:00Z","hook":"preToolUse","tool":"Edit"}` + "\n"
	}
	mustWriteFile(t, filepath.Join(root, ".atv", "observations.jsonl"), obsLines)

	manifest := installstate.InstallManifest{
		Requested: installstate.RequestedState{
			StackPacks: []installstate.StackPack{installstate.StackPackGeneral},
			PresetName: "Starter",
		},
		Outcomes: []installstate.InstallOutcome{
			{Step: "Scaffolding ATV files", Status: installstate.InstallStepDone},
		},
	}
	recs := installstate.BuildRecommendations(root, manifest)

	// Should recommend /learn when observations exist but no instincts
	foundLearn := false
	for _, rec := range recs {
		if rec.ID == "run-learn" {
			foundLearn = true
			if !strings.Contains(rec.Reason, "15 observations") {
				t.Errorf("run-learn reason should mention observation count, got %q", rec.Reason)
			}
		}
	}
	if !foundLearn {
		t.Fatalf("should recommend /learn when observations > 10 but no instincts, got %+v", recs)
	}

	// Now add instincts — should recommend /instincts instead
	mustWriteFile(t, filepath.Join(root, ".atv", "instincts", "project.yaml"),
		`instincts:
  - id: test-instinct
    confidence: 0.85
`)
	recs = installstate.BuildRecommendations(root, manifest)
	foundInstincts := false
	for _, rec := range recs {
		if rec.ID == "check-instincts" {
			foundInstincts = true
		}
	}
	if !foundInstincts {
		t.Fatalf("should recommend /instincts when instincts exist, got %+v", recs)
	}
}

func TestE2EMultiStackDeterminism(t *testing.T) {
	// Order A: General, TypeScript, Rails
	catalogA := scaffold.BuildFilteredCatalogForPacks(
		[]installstate.StackPack{installstate.StackPackGeneral, installstate.StackPackTypeScript, installstate.StackPackRails},
		detect.StackRails,
		[]string{"core-skills", "universal-agents", "stack-agents", "file-instructions"},
	)

	// Order B: Rails, General, TypeScript (reversed)
	catalogB := scaffold.BuildFilteredCatalogForPacks(
		[]installstate.StackPack{installstate.StackPackRails, installstate.StackPackGeneral, installstate.StackPackTypeScript},
		detect.StackRails,
		[]string{"core-skills", "universal-agents", "stack-agents", "file-instructions"},
	)

	if len(catalogA) != len(catalogB) {
		t.Fatalf("catalog lengths differ: %d vs %d", len(catalogA), len(catalogB))
	}

	pathsA := make(map[string]bool, len(catalogA))
	for _, c := range catalogA {
		pathsA[c.Path] = true
	}

	for _, c := range catalogB {
		if !pathsA[c.Path] {
			t.Fatalf("catalog B has path %q not in catalog A", c.Path)
		}
	}

	// Write both catalogs and verify identical output
	rootA := t.TempDir()
	rootB := t.TempDir()

	scaffold.WriteAll(rootA, catalogA)
	scaffold.WriteAll(rootB, catalogB)

	// Verify same files exist
	var filesA, filesB []string
	if err := filepath.Walk(rootA, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(rootA, path)
			filesA = append(filesA, filepath.ToSlash(rel))
		}
		return nil
	}); err != nil {
		t.Fatalf("Walk(rootA) error = %v", err)
	}
	if err := filepath.Walk(rootB, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			rel, _ := filepath.Rel(rootB, path)
			filesB = append(filesB, filepath.ToSlash(rel))
		}
		return nil
	}); err != nil {
		t.Fatalf("Walk(rootB) error = %v", err)
	}

	if len(filesA) != len(filesB) {
		t.Fatalf("file count differs: %d vs %d", len(filesA), len(filesB))
	}

	setA := make(map[string]bool, len(filesA))
	for _, f := range filesA {
		setA[f] = true
	}
	for _, f := range filesB {
		if !setA[f] {
			t.Fatalf("file %q in B but not A", f)
		}
	}
}
