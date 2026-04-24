# PRD: Rebrand memegen → memeIQ + Easter Egg Installer Category

## Goal

Rebrand the meme-generation skill and meme-creator agent to "memeIQ", add them as installer templates with a new Easter Egg category in the guided TUI, and verify it installs and generates a meme in a sandbox.

## Tasks

### 1. Create memeIQ skill template for the installer
- Copy `.github/skills/meme-generation/SKILL.md` to `pkg/scaffold/templates/skills/meme-iq/SKILL.md`
- Rebrand all references from "meme-generation" to "meme-iq" / "memeIQ"
- Update the skill frontmatter name to `meme-iq`

### 2. Create memeIQ agent template for the installer
- Copy `.github/agents/meme-creator.agent.md` to `pkg/scaffold/templates/agents/meme-iq.agent.md`
- Rebrand description and references from "meme-creator" to "meme-iq" / "memeIQ"

### 3. Rebrand existing project-root files
- Rename `.github/skills/meme-generation/` → `.github/skills/meme-iq/`
- Rename `.github/agents/meme-creator.agent.md` → `.github/agents/meme-iq.agent.md`
- Update all internal references in both files

### 4. Add Easter Egg category to installer TUI
- Add `CategoryEasterEgg = "easter-egg"` constant in `pkg/gstack/skills.go`
- Add it to `AllCategories()` list
- Add label `"🥚 Easter Eggs"` in `CategoryLabel()`
- Add description in `pkg/tui/categories.go` `categoryDescription()`
- Add memeIQ to `atvCategoryMapping` under the easter-egg category

### 5. Wire memeIQ into the catalog
- Add `"meme-iq"` to a skill directory list in `pkg/scaffold/catalog.go` (new `easterEggSkillDirectories`)
- Add new `LayerEasterEggs` constant in `pkg/tui/wizard.go`
- Add easter-egg layer handling in `BuildFilteredCatalog` in `catalog.go`
- Add easter egg to `InfraLayers` or as a separate category-driven install

### 6. Build and test in sandbox
- Run `go build ./...` to verify compilation
- Run `go test ./...` to verify existing tests pass
- Create a sandbox directory, run the installer in non-interactive mode targeting it
- Verify `.github/skills/meme-iq/SKILL.md` and `.github/agents/meme-iq.agent.md` are scaffolded
- Generate a test meme URL to confirm the skill content works

## Acceptance Criteria

- [ ] `pkg/scaffold/templates/skills/meme-iq/SKILL.md` exists with "memeIQ" branding
- [ ] `pkg/scaffold/templates/agents/meme-iq.agent.md` exists with "memeIQ" branding
- [ ] `.github/skills/meme-iq/SKILL.md` replaces old `.github/skills/meme-generation/SKILL.md`
- [ ] `.github/agents/meme-iq.agent.md` replaces old `.github/agents/meme-creator.agent.md`
- [ ] `CategoryEasterEgg` wired into gstack categories, TUI labels, and descriptions
- [ ] memeIQ appears in the guided wizard under "🥚 Easter Eggs" category
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes
- [ ] Sandbox install scaffolds the memeIQ skill and agent
- [ ] A meme URL can be generated using the installed skill content

## Completion Promise

DONE
