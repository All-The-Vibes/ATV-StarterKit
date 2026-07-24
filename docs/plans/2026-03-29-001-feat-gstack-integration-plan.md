---
title: "feat: Integrate gstack into ATV StarterKit Unified Installer"
type: feat
status: active
date: 2026-03-29
origin: docs/brainstorms/2026-03-29-gstack-integration-brainstorm.md
---

# feat: Integrate gstack into ATV StarterKit Unified Installer

> **Historical note (2026-07):** This plan predates the binary rename in PR #58.
> The shipped binary file is now `atv-inst-tool` (renamed from `atv-installer`
> to avoid a Windows UAC installer-detection heuristic), and the user-facing
> command is `atv-starterkit`. References to `atv-installer` below are retained
> as the original record; the `atv-installer` command name also survives as an
> npm `bin` alias.

## Overview

Add [garrytan/gstack](https://github.com/garrytan/gstack) as a bundled skill source in the ATV StarterKit installer. The guided TUI wizard presents a single categorized list of skills organized by **function** (Planning, Review, QA, Security, etc.) — mixing ATV and gstack skills together. Users don't need to know which system a skill came from. ATV skills take priority where both provide overlapping functionality.

The installer gains new capabilities: network operations (git clone), runtime detection (Bun), and external process execution (bun run build). A sandbox environment enables testing the full install flow without polluting real projects.

## Problem Statement

ATV StarterKit currently bundles ~14 workflow skills and ~28 agents, all embedded at compile time. gstack offers 29 complementary skills covering areas ATV doesn't: browser-based QA, real deployment automation, safety guardrails, and a structured sprint process. Users must install gstack separately today, manually managing compatibility. A unified installer eliminates this friction and presents the best of both systems.

## Proposed Solution

Extend the Go installer with:
1. A new `pkg/gstack` package handling clone, build, and file placement
2. Reorganized TUI wizard with function-based categories instead of source-based layers
3. A sandbox test environment using temp directories for integration testing
4. Updated copilot-instructions templates that list the combined skill set

## Technical Approach

### Architecture

```
┌──────────────────────────────────────────────────────────────┐
│ atv-installer init --guided                                  │
└─────────────┬────────────────────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────────────────────┐
│ detect.go: Detect stack + detect Bun/Git availability        │
└─────────────┬────────────────────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────────────────────┐
│ wizard.go: Function-based TUI                                │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │ Planning:  ce-brainstorm (ATV), office-hours (gstack)   │ │
│  │ Review:    ce-review (ATV), review (gstack)             │ │
│  │ QA:        qa, browse (gstack) [requires Bun]           │ │
│  │ Security:  security-sentinel (ATV), cso (gstack)        │ │
│  │ Shipping:  ce-work (ATV), ship (gstack)                 │ │
│  │ Safety:    careful, freeze, guard (gstack)              │ │
│  │ ...                                                     │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────┬────────────────────────────────────────────────┘
              │
              ├── ATV components ──► scaffold.WriteAll() (embedded)
              │
              └── gstack components ──► gstack.Install() (network)
                    │
                    ├── git clone --depth 1
                    ├── rm -rf .git
                    └── bun run build (if full runtime selected)
```

### Implementation Phases

#### Phase 1: Foundation — `pkg/gstack` Package

**Goal:** Isolated package that clones, configures, and builds gstack.

**Tasks:**

- [ ] Create `pkg/gstack/gstack.go` — core install logic
  - `DetectPrerequisites() Prerequisites` — check git, bun, node availability via `exec.LookPath`
  - `Clone(targetDir string) error` — `git clone --depth 1 https://github.com/garrytan/gstack.git <targetDir>`
  - `StripGit(targetDir string) error` — remove `.git/` directory after clone
  - `Build(targetDir string) error` — `bun run build` in the gstack directory
  - `Install(targetDir string, mode InstallMode) (*InstallResult, error)` — orchestrates clone → strip → build
- [ ] Create `pkg/gstack/prereqs.go` — prerequisite detection
  - `Prerequisites` struct: `HasGit bool`, `HasBun bool`, `HasNode bool`, `GitVersion string`, `BunVersion string`
  - Run `git --version`, `bun --version`, `node --version` and parse output
- [ ] Create `pkg/gstack/skills.go` — gstack skill catalog and metadata
  - Define `GstackSkill` struct: `Name`, `Dir`, `Category`, `RequiresRuntime bool`, `Description string`
  - Hardcoded catalog of gstack skills with their categories and runtime requirements
  - `SkillsByCategory() map[string][]GstackSkill` — group skills for TUI display
  - `FilterSkills(selected []string, hasRuntime bool) []GstackSkill` — filter by user selection and runtime availability
  - **Assumption:** gstack skill SKILL.md paths follow `{skill-name}/SKILL.md` pattern (e.g., `review/SKILL.md`, `qa/SKILL.md`). Verify against actual clone output during Phase 1 development.

**Files:**

```
pkg/gstack/
├── gstack.go      # Clone, strip, build orchestration
├── prereqs.go     # Git/Bun/Node detection
└── skills.go      # Skill catalog and metadata
```

**Acceptance criteria:**
- [ ] `gstack.Install()` clones to a target directory, strips `.git`, and optionally builds
- [ ] `DetectPrerequisites()` correctly identifies git, bun, node on Windows, macOS, Linux
- [ ] If git is missing, Install returns a clear error
- [ ] If bun is missing and full runtime requested, falls back to markdown-only with a warning
- [ ] Network errors produce user-friendly messages, not raw Go errors
- [ ] `Install()` is idempotent — running twice doesn't duplicate files

#### Phase 2: TUI Reorganization — Function-Based Categories

**Goal:** Replace source-based layers (core-skills, orchestrators, etc.) with function-based categories.

**Tasks:**

- [ ] Define new category constants in `pkg/tui/wizard.go`:
  ```
  CategoryPlanning      = "planning"
  CategoryReview        = "review"
  CategoryQATesting     = "qa-testing"
  CategorySecurity      = "security"
  CategoryShipping      = "shipping"
  CategorySafety        = "safety"
  CategoryDebugging     = "debugging"
  CategoryRetrospective = "retrospective"
  CategoryInfra         = "infrastructure"  // MCP, extensions, instructions, setup-steps
  ```
- [ ] Create `pkg/tui/categories.go` — category-to-skill mapping
  - Map each category to a mix of ATV and gstack skill identifiers
  - ATV skills are default-selected (checked by default)
  - gstack skills are default-selected where no ATV overlap exists
  - Browser-requiring skills are disabled (grayed out) when Bun is not detected
- [ ] Update `RunWizard()` in `wizard.go`:
  - Step 1: Stack selection (unchanged)
  - Step 2 (new): Detect prerequisites, show Bun/runtime status
  - Step 3: Category-based multi-select with descriptions showing ATV vs gstack skill names
  - Step 4 (new): If gstack skills selected and no Bun detected, show warning + markdown-only confirmation
- [ ] Update `WizardResult` struct to carry both ATV layers and gstack selections:
  ```go
  type WizardResult struct {
      Stack           detect.Stack
      ATVComponents   []string   // existing layer keys
      GstackSkills    []string   // selected gstack skill dirs
      GstackRuntime   bool       // whether to build TS binary
  }
  ```

**Acceptance criteria:**
- [ ] TUI shows categories like "Planning", "Review", "QA & Testing" with mixed ATV + gstack skills
- [ ] Skills requiring Bun runtime are visibly marked and disabled when Bun is absent
- [ ] Wizard returns separate ATV and gstack selections for the scaffold phase
- [ ] Existing `--guided` flag behavior preserved for ATV-only installs
- [ ] Default auto-mode (`init` without `--guided`) installs ATV-only (no gstack, preserves current behavior)
- [ ] Backward compat: existing `--guided` users who don't select gstack categories see no change in output

#### Phase 3: Scaffold Integration

**Goal:** Wire gstack installation into the existing scaffold flow.

**Tasks:**

- [ ] Update `cmd/init.go` `runInit()`:
  - After `scaffold.WriteAll()`, check if gstack skills were selected
  - Call `gstack.Install()` with the appropriate mode
  - Merge gstack results into the overall results for printing
- [ ] Update `pkg/scaffold/catalog.go` `BuildFilteredCatalog()`:
  - Accept new category keys alongside legacy layer keys
  - Map categories back to existing ATV components
  - Return gstack skill list as a separate slice (not embedded, since they're cloned)
- [ ] Update `copilot-instructions.md` templates:
  - Add conditional gstack skill section listing
  - ATV skills listed first in each category
  - Format: `Available skills: /ce-brainstorm (planning), /office-hours (planning, gstack), ...`
- [ ] Update `pkg/output/printer.go`:
  - New status indicator for gstack operations: 🔗 Cloned, 🔨 Built
  - Show gstack install progress (clone → strip → build)
  - Show prerequisite status in detection output

**Acceptance criteria:**
- [ ] `atv-installer init --guided` with gstack selection clones gstack to `.github/skills/gstack/`
- [ ] `atv-installer init` (auto mode) does NOT install gstack (backward compatible)
- [ ] Printer shows gstack clone/build status alongside ATV file creation
- [ ] copilot-instructions.md includes gstack skills when selected

#### Phase 4: Sandbox Testing Environment

**Goal:** A test harness that exercises the full install flow in isolation.

**Tasks:**

- [ ] Create `pkg/gstack/gstack_test.go` — unit tests
  - `TestDetectPrerequisites` — verify git/bun/node detection
  - `TestClone` — clone to temp dir, verify skill files exist
  - `TestStripGit` — verify `.git` removed after strip
  - `TestBuild` — verify `bun run build` produces binary (skip if no Bun)
  - `TestInstallIdempotent` — run Install twice, verify no errors
  - `TestInstallNoGit` — verify clear error when git missing
  - `TestInstallNoBun` — verify fallback to markdown-only
- [ ] Create `test/sandbox/sandbox_test.go` — integration tests
  - `TestFullInstallFlow` — creates temp dir, runs full init with gstack, verifies output structure
  - `TestGuidedModeWithGstack` — simulates wizard selections, verifies correct files placed
  - `TestAutoModeNoGstack` — verifies auto mode doesn't install gstack
  - `TestGstackOverlapResolution` — verifies ATV skills take priority
- [ ] Create `Makefile` target: `test-sandbox`
  - Runs integration tests in isolated temp directories
  - Cleans up after each test
  - Can be run locally or in CI
- [ ] Add `test/sandbox/testdata/` — expected output fixtures
  - Golden files for expected directory structures
  - Expected copilot-instructions.md content with gstack skills

**Sandbox design:**

```go
// test/sandbox/sandbox_test.go
func TestFullInstallFlow(t *testing.T) {
    // Create isolated temp directory
    sandboxDir := t.TempDir()
    
    // Create a fake project (e.g., tsconfig.json for TypeScript detection)
    os.WriteFile(filepath.Join(sandboxDir, "tsconfig.json"), []byte("{}"), 0644)
    
    // Run full install with gstack
    env := detect.DetectEnvironment(sandboxDir)
    catalog := scaffold.BuildCatalog(env.Stack)
    results := scaffold.WriteAll(sandboxDir, catalog)
    
    // Install gstack
    gstackResult, err := gstack.Install(
        filepath.Join(sandboxDir, ".github", "skills", "gstack"),
        gstack.ModeMarkdownOnly,
    )
    
    // Verify ATV files exist
    assert.FileExists(t, filepath.Join(sandboxDir, ".github", "copilot-instructions.md"))
    assert.FileExists(t, filepath.Join(sandboxDir, ".github", "skills", "ce-brainstorm", "SKILL.md"))
    
    // Verify gstack files exist
    assert.DirExists(t, filepath.Join(sandboxDir, ".github", "skills", "gstack"))
    assert.FileExists(t, filepath.Join(sandboxDir, ".github", "skills", "gstack", "review", "SKILL.md"))
    
    // Verify no .git in gstack
    assert.NoDirExists(t, filepath.Join(sandboxDir, ".github", "skills", "gstack", ".git"))
}
```

**Running the sandbox:**

```bash
# Run all tests including sandbox integration tests
go test ./...

# Run only sandbox tests (slower, requires network)
go test ./test/sandbox/ -v -run TestFullInstallFlow

# Run with short flag to skip network-dependent tests
go test ./test/sandbox/ -short
```

## Context & Memory Integration

**Principle: ATV overrides gstack in all conflicts.** ATV's file-based, git-tracked memory system takes priority over gstack's runtime-based state.

### ATV's Existing Memory System (Preserved As-Is)

| Layer | Mechanism | Location |
|-------|-----------|----------|
| Institutional knowledge | `/ce-compound` writes structured solution docs with YAML frontmatter | `docs/solutions/**/*.md` |
| Knowledge retrieval | `learnings-researcher` agent greps `docs/solutions/` before new work | Invoked by `/ce-plan`, `/ce-review` |
| Project config | `/setup` writes review agent config | `compound-engineering.local.md` |
| Design context | `/ce-brainstorm` writes design docs, `/ce-plan` auto-discovers them | `docs/brainstorms/*.md` |
| Plan context | `/ce-plan` writes plans, `/ce-work` reads them | `docs/plans/*.md` |
| Protected artifacts | `/ce-review` prevents deletion of `docs/plans/`, `docs/solutions/` | Enforced in SKILL.md rules |

### gstack's Memory System (New, Additive)

| Layer | Mechanism | Location | Integration Rule |
|-------|-----------|----------|------------------|
| Session tracking | Counts active sessions, triggers ELI16 mode at 3+ | `~/.gstack/sessions/$PPID` | No conflict — user-global, not project-scoped |
| Config persistence | Prefix choice, telemetry opt-in, contributor mode | `~/.gstack/config.yaml` | No conflict — user-global |
| Telemetry | Opt-in anonymous usage data (skill name, duration, success/fail) | Supabase (remote) | No conflict — opt-in only |
| Local analytics | Usage dashboard data | `~/.gstack/*.jsonl` | No conflict — user-global |
| Browser state | PID, port, auth token for browse daemon | `.gstack/browse.json` | Add `.gstack/` to generated `.gitignore` |
| GStack Learns | Per-project self-learning infrastructure | Project-scoped files | **Override: route through `docs/solutions/` format** |
| Binary versioning | Auto-restart on version mismatch | `browse/dist/.version` | No conflict — internal to gstack dir |

### Conflict Resolution Rules

These rules are enforced in the scaffolded `copilot-instructions.md`:

1. **Design docs**: ATV's `docs/brainstorms/*.md` takes priority over gstack's `DESIGN.md`. gstack skills that write `DESIGN.md` should be instructed to write to `docs/brainstorms/` instead.
2. **Solution docs**: ATV's `docs/solutions/` takes priority over gstack's `/retro` output. gstack's `/retro` findings that warrant documentation should flow through `/ce-compound` into `docs/solutions/`.
3. **Review config**: ATV's `compound-engineering.local.md` governs which review agents run. gstack's `/review` runs alongside but does not override ATV's `/ce-review` agent selection.
4. **Plans**: ATV's `docs/plans/` is the canonical plan location. gstack skills that generate plans should target `docs/plans/` with ATV's naming convention (`YYYY-MM-DD-NNN-<type>-<name>-plan.md`).
5. **Protected artifacts**: gstack agents must respect ATV's protected artifact rules — never flag `docs/plans/`, `docs/solutions/`, `docs/brainstorms/`, or `compound-engineering.local.md` for deletion.

### Implementation Tasks

- [ ] Add `.gstack/` entry to the generated `.gitignore` template
- [ ] Add conflict resolution rules to `copilot-instructions.md` templates (all stacks)
- [ ] Document `~/.gstack/` user-global state in installer output (next steps section)
- [ ] Add gstack protected paths to `/ce-review` SKILL.md's protected artifacts list: `.github/skills/gstack/`

## Alternative Approaches Considered

(see brainstorm: docs/brainstorms/2026-03-29-gstack-integration-brainstorm.md)

1. **Live Git Clone with .git intact** — Rejected: couples to git availability, `.git` overhead in project
2. **Vendor into ATV Templates** — Rejected: 29 skills bloat binary, can't replicate TypeScript build in Go
3. **Release Tarball** — Evolved into shallow clone + strip, since gstack has no formal releases

## System-Wide Impact

### Interaction Graph

- `cmd/init.go` → `pkg/gstack/gstack.go` (new dependency)
- `pkg/tui/wizard.go` → `pkg/gstack/prereqs.go` (for runtime detection display)
- `pkg/tui/wizard.go` → `pkg/gstack/skills.go` (for category-skill mapping)
- `pkg/scaffold/catalog.go` → category mapping (new grouping logic)
- `pkg/output/printer.go` → new status types for clone/build operations

### Error Propagation

| Error | Source | Handling |
|-------|--------|----------|
| Git not installed | `exec.LookPath("git")` | Fatal error with install instructions |
| Git clone fails (network) | `exec.Command("git", "clone", ...)` | Warn, skip gstack, continue ATV-only |
| Bun not installed | `exec.LookPath("bun")` | Fallback to markdown-only mode |
| `bun run build` fails | `exec.Command("bun", "run", "build")` | Warn, skip binary, SKILL.md files still usable |
| Target directory exists | `os.Stat()` | Skip clone, warn user |

### State Lifecycle Risks

- **Partial clone failure**: If clone succeeds but strip or build fails, a half-installed gstack directory remains. Mitigation: `Install()` cleans up partially-written directories on error.
- **No rollback for ATV files**: If gstack install fails after ATV files are written, ATV files remain (acceptable — they're still valid).

### API Surface Parity

- Auto mode (`init`): ATV-only, unchanged behavior
- Guided mode (`init --guided`): Extended with gstack categories
- New flag consideration: `init --with-gstack` for auto mode + gstack (future, not in this plan)

### Integration Test Scenarios

1. Fresh project + guided mode + all categories selected + Bun available → ATV + gstack full runtime installed
2. Fresh project + guided mode + all categories selected + no Bun → ATV + gstack markdown-only installed
3. Fresh project + auto mode → ATV only, no gstack (backward compatible)
4. Existing ATV project + guided mode + gstack selected → gstack added, ATV files skipped (writeIfNotExists)
5. No git on system + guided mode + gstack selected → clear error, ATV installs normally

## Acceptance Criteria

### Functional Requirements

- [ ] `atv-installer init --guided` shows categories with mixed ATV + gstack skills
- [ ] Selecting gstack skills clones gstack to `.github/skills/gstack/`
- [ ] ATV skills override gstack for overlapping functions
- [ ] Browser-based skills are disabled when Bun is not detected
- [ ] `atv-installer init` (auto mode) behavior is unchanged (no gstack)
- [ ] Sandbox tests verify the full flow in temp directories

### Non-Functional Requirements

- [ ] Clone + build completes in <30 seconds on reasonable network
- [ ] Install is idempotent — running twice produces no errors
- [ ] All error messages include actionable next steps
- [ ] Windows support: git/bun detection works on Windows

### Quality Gates

- [ ] Unit tests for `pkg/gstack/` pass
- [ ] Integration tests in `test/sandbox/` pass
- [ ] Existing tests (`go test ./...`) continue to pass
- [ ] `golangci-lint` passes (use `install-mode: goinstall` per CI convention)
- [ ] Manual test: guided mode on a fresh directory produces working Copilot setup

## Success Metrics

- Users can install ATV + gstack with a single `atv-installer init --guided` command
- All 29 gstack skills are accessible after full runtime install
- Markdown-only mode works for users without Bun
- Zero regressions in existing ATV-only install flow

## Dependencies & Prerequisites

| Dependency | Required By | Status |
|-----------|-------------|--------|
| Go 1.26.1 | Build | ✅ Available |
| charmbracelet/huh | TUI | ✅ Already in go.mod |
| git (user's machine) | gstack clone | Required at install time |
| Bun (user's machine) | Full runtime | Optional, fallback to markdown-only |
| Network access | gstack clone | Required at install time |

## Risk Analysis & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| gstack repo goes offline | Clone fails | Low | Graceful fallback to ATV-only with clear message |
| gstack breaks backward compat | Skills stop working | Medium | Pin to a known-good commit SHA (future enhancement) |
| Bun install complexity on Windows | Build fails | Medium | Markdown-only fallback, clear error messages |
| TUI reorganization breaks existing users | Confusion | Low | Auto mode unchanged, guided mode clearly labeled |
| Large clone size slows install | Bad UX | Low | `--depth 1` keeps clone <10MB |

## Future Considerations

- **Version pinning**: Pin to specific gstack commit SHA instead of `main` HEAD
- **`--with-gstack` flag**: Auto mode + gstack without guided TUI
- **Offline mode**: Bundle gstack SKILL.md files as fallback when no network
- **Update command**: `atv-installer update-gstack` to re-clone latest

## Sources & References

### Origin

- **Brainstorm document:** [docs/brainstorms/2026-03-29-gstack-integration-brainstorm.md](docs/brainstorms/2026-03-29-gstack-integration-brainstorm.md) — Key decisions carried forward: single Go installer, ATV overrides gstack, categorized TUI by function, shallow clone + strip .git, build at install time

### Internal References

- Scaffold system: `pkg/scaffold/catalog.go` — BuildCatalog, BuildFilteredCatalog
- Write strategies: `pkg/scaffold/scaffold.go` — writeIfNotExists, writeOrMergeJSON
- TUI wizard: `pkg/tui/wizard.go` — RunWizard, WizardResult
- Init flow: `cmd/init.go` — runInit
- Detection: `pkg/detect/detect.go` — DetectEnvironment

### External References

- gstack README: https://github.com/garrytan/gstack
- gstack Architecture: https://github.com/garrytan/gstack/blob/main/ARCHITECTURE.md
- gstack requires: Bun v1.0+, Git, Node.js (Windows only)
- gstack install target: `.github/skills/gstack/` or `.claude/skills/gstack/`
