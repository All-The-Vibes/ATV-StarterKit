package gstack

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	// GstackRepo is the GitHub repository URL for gstack.
	GstackRepo = "https://github.com/garrytan/gstack.git"
)

// InstallMode controls how much of gstack to install.
type InstallMode int

const (
	// ModeMarkdownOnly copies SKILL.md files only, no binary build.
	ModeMarkdownOnly InstallMode = iota
	// ModeFullRuntime clones and runs ./setup for browser skills + binary.
	ModeFullRuntime
)

// InstallResult reports what happened during installation.
type InstallResult struct {
	Cloned    bool
	Built     bool
	Copied    bool
	SkillDirs []string
	Mode      InstallMode
	Error     error
	Warning   string
}

// Install clones gstack, runs its own ./setup, then copies generated skills to .github/skills/.
//
// Flow:
//  1. Clone gstack to <projectDir>/.gstack/ (gitignored staging area)
//  2. Run gstack's own ./setup script (handles deps, build, skill generation)
//  3. Copy generated skills from .gstack/.agents/skills/gstack-* → .github/skills/gstack-*
func Install(projectDir string, mode InstallMode) *InstallResult {
	result := &InstallResult{Mode: mode}
	stagingDir := filepath.Join(projectDir, ".gstack")
	skillsTargetDir := filepath.Join(projectDir, ".github", "skills")

	// Fix BOM-encoded package.json in the project root. Bun traverses up from
	// .gstack/ and will fail with "Unexpected ■" if it finds a UTF-16 BOM.
	stripBOMFromPackageJSON(filepath.Join(projectDir, "package.json"))

	// Idempotent: skip if .gstack/ already exists with SKILL.md
	if _, err := os.Stat(filepath.Join(stagingDir, "SKILL.md")); err == nil {
		// Already cloned — just re-copy skills
		copied, dirs := copyGeneratedSkills(stagingDir, skillsTargetDir)
		result.Copied = copied
		result.SkillDirs = dirs
		result.Warning = "gstack already cloned, re-synced skills"
		return result
	}

	// Step 1: Clone to .gstack/
	if err := clone(stagingDir); err != nil {
		result.Error = fmt.Errorf("failed to clone gstack: %w", err)
		return result
	}
	result.Cloned = true

	// Step 2: Run gstack's ./setup (full runtime) or just generate docs (markdown-only)
	if mode == ModeFullRuntime {
		if err := runSetup(stagingDir); err != nil {
			result.Warning = fmt.Sprintf("runtime build failed (%v); fell back to docs only", err)
			// Fallback: try generating docs without full build
			_ = generateDocs(stagingDir)
		} else {
			result.Built = true
		}
	} else {
		// Markdown-only: just install deps and generate skill docs
		if err := generateDocs(stagingDir); err != nil {
			result.Warning = fmt.Sprintf("doc generation failed (%v), copying raw SKILL.md files", err)
		}
	}

	// Step 3: Copy generated skills to .github/skills/
	copied, dirs := copyGeneratedSkills(stagingDir, skillsTargetDir)
	result.Copied = copied
	result.SkillDirs = dirs

	// Step 4: Prune non-GitHub platform dirs from staging
	pruneNonGitHubDirs(stagingDir)

	return result
}

// clone performs a shallow git clone of gstack.
func clone(targetDir string) error {
	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	cmd := exec.Command("git", "clone", "--single-branch", "--depth", "1", GstackRepo, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone failed (check network connection): %w", err)
	}
	return nil
}

// runSetup runs gstack's own ./setup script which handles deps, build, and skill generation.
func runSetup(gstackDir string) error {
	// Fix CRLF in scripts on Windows before running
	if runtime.GOOS == "windows" {
		fixCRLFInScripts(gstackDir)
	}

	setupPath := filepath.Join(gstackDir, "setup")

	// Find bash — on Windows it comes with Git (not WSL)
	bashPath := findBash()
	if bashPath == "" {
		return fmt.Errorf("bash not found; gstack's setup requires bash (included with Git for Windows)")
	}

	var stderr bytes.Buffer
	cmd := exec.Command(bashPath, setupPath, "--host", "codex")
	cmd.Dir = gstackDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr
	// Prevent bun from walking up to the parent project's package.json.
	// Without this, a BOM-encoded or malformed package.json in the project root
	// causes bun to fail inside .gstack/ even though .gstack has its own package.json.
	cmd.Env = append(os.Environ(), "NONINTERACTIVE=1", "BUN_WORKSPACE_ROOT="+gstackDir)

	if err := cmd.Run(); err != nil {
		errDetail := firstLine(stderr.String())
		// Exit 127 = "command not found" inside bash — usually means bun is missing
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 127 {
			hint := "a required command was not found (likely 'bun'); install bun: https://bun.sh"
			if runtime.GOOS == "windows" {
				hint += " (verify Git Bash is being used, not WSL)"
			}
			return fmt.Errorf("%s", hint)
		}
		if errDetail != "" {
			return fmt.Errorf("%s", errDetail)
		}
		return fmt.Errorf("exit status %d", exitErrorCode(err))
	}
	return nil
}

// firstLine returns the first meaningful error line from stderr output.
// Skips blank lines, caret markers (^), and code context lines (e.g. "1 |").
// Prefers lines starting with "error:" if present.
func firstLine(s string) string {
	var fallback string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "^" {
			continue
		}
		// Strip null bytes that appear in BOM-related errors
		clean := strings.Map(func(r rune) rune {
			if r == 0 {
				return -1
			}
			return r
		}, line)
		clean = strings.TrimSpace(clean)
		if clean == "" {
			continue
		}
		// Prefer actual error messages over code context dump lines
		if strings.HasPrefix(clean, "error:") {
			return clean
		}
		if fallback == "" {
			fallback = clean
		}
	}
	return fallback
}

// exitErrorCode extracts the exit code from an error, or returns -1.
func exitErrorCode(err error) int {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}

// generateDocs runs bun install + bun run gen:skill-docs without the full build.
func generateDocs(gstackDir string) error {
	if _, err := exec.LookPath("bun"); err != nil {
		return fmt.Errorf("bun not found; install bun (https://bun.sh)")
	}

	installCmd := exec.Command("bun", "install")
	installCmd.Dir = gstackDir
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("bun install failed: %w", err)
	}

	genCmd := exec.Command("bun", "run", "gen:skill-docs", "--host", "codex")
	genCmd.Dir = gstackDir
	genCmd.Stdout = os.Stdout
	genCmd.Stderr = os.Stderr
	if err := genCmd.Run(); err != nil {
		return fmt.Errorf("gen:skill-docs failed: %w", err)
	}
	return nil
}

// copyGeneratedSkills copies gstack-*/SKILL.md from .agents/skills/ to .github/skills/,
// and creates a runtime sidecar at .github/skills/gstack/ with bin/, browse/dist/, ETHOS.md.
// Falls back to copying raw skill dirs if .agents/ doesn't exist.
func copyGeneratedSkills(gstackDir, targetSkillsDir string) (bool, []string) {
	var dirs []string

	// Try generated .agents/skills/ first (produced by ./setup or gen:skill-docs)
	agentsDir := filepath.Join(gstackDir, ".agents", "skills")
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			// Only copy gstack-prefixed skill dirs (not the root "gstack" sidecar)
			if len(name) < 7 || name[:7] != "gstack-" {
				continue
			}
			srcSkill := filepath.Join(agentsDir, name, "SKILL.md")
			if _, err := os.Stat(srcSkill); err != nil {
				continue
			}
			destDir := filepath.Join(targetSkillsDir, name)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				continue
			}
			data, err := os.ReadFile(srcSkill)
			if err != nil {
				continue
			}
			if err := os.WriteFile(filepath.Join(destDir, "SKILL.md"), data, 0644); err != nil {
				continue
			}
			dirs = append(dirs, name)
		}
	}

	// Fallback: copy raw skill dirs directly from gstack root
	if len(dirs) == 0 {
		entries, err := os.ReadDir(gstackDir)
		if err != nil {
			return false, nil
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			srcSkill := filepath.Join(gstackDir, name, "SKILL.md")
			if _, err := os.Stat(srcSkill); err != nil {
				continue
			}
			destName := "gstack-" + name
			destDir := filepath.Join(targetSkillsDir, destName)
			if err := os.MkdirAll(destDir, 0755); err != nil {
				continue
			}
			data, err := os.ReadFile(srcSkill)
			if err != nil {
				continue
			}
			if err := os.WriteFile(filepath.Join(destDir, "SKILL.md"), data, 0644); err != nil {
				continue
			}
			dirs = append(dirs, destName)
		}
	}

	// Create runtime sidecar at .github/skills/gstack/
	// Skills reference $GSTACK_ROOT for binaries, ETHOS.md, and review assets
	createSidecar(gstackDir, targetSkillsDir)

	return len(dirs) > 0, dirs
}

// createSidecar creates .github/skills/gstack/ with the runtime assets that skills reference.
// This mirrors what gstack's ./setup creates for Codex at .agents/skills/gstack/.
func createSidecar(gstackDir, targetSkillsDir string) {
	sidecar := filepath.Join(targetSkillsDir, "gstack")
	_ = os.MkdirAll(sidecar, 0755)

	// Copy root SKILL.md (the meta-skill for browse)
	copyFileIfExists(filepath.Join(gstackDir, "SKILL.md"), filepath.Join(sidecar, "SKILL.md"))

	// Copy ETHOS.md (referenced by "Search Before Building" in all skill preambles)
	copyFileIfExists(filepath.Join(gstackDir, "ETHOS.md"), filepath.Join(sidecar, "ETHOS.md"))

	// Copy AGENTS.md
	copyFileIfExists(filepath.Join(gstackDir, "AGENTS.md"), filepath.Join(sidecar, "AGENTS.md"))

	// Symlink or copy bin/ (gstack-config, gstack-update-check, etc.)
	linkOrCopyDir(filepath.Join(gstackDir, "bin"), filepath.Join(sidecar, "bin"))

	// Symlink or copy browse/dist/ (compiled browse binary)
	browseDist := filepath.Join(gstackDir, "browse", "dist")
	if _, err := os.Stat(browseDist); err == nil {
		_ = os.MkdirAll(filepath.Join(sidecar, "browse"), 0755)
		linkOrCopyDir(browseDist, filepath.Join(sidecar, "browse", "dist"))
	}

	// Symlink or copy browse/bin/ (helper scripts)
	browseBin := filepath.Join(gstackDir, "browse", "bin")
	if _, err := os.Stat(browseBin); err == nil {
		_ = os.MkdirAll(filepath.Join(sidecar, "browse"), 0755)
		linkOrCopyDir(browseBin, filepath.Join(sidecar, "browse", "bin"))
	}

	// Symlink or copy design/dist/ (design binary)
	designDist := filepath.Join(gstackDir, "design", "dist")
	if _, err := os.Stat(designDist); err == nil {
		_ = os.MkdirAll(filepath.Join(sidecar, "design"), 0755)
		linkOrCopyDir(designDist, filepath.Join(sidecar, "design", "dist"))
	}

	// Copy review runtime assets (checklists, not the SKILL.md)
	reviewAssets := []string{"checklist.md", "design-checklist.md", "greptile-triage.md", "TODOS-format.md"}
	reviewDir := filepath.Join(sidecar, "review")
	_ = os.MkdirAll(reviewDir, 0755)
	for _, f := range reviewAssets {
		copyFileIfExists(filepath.Join(gstackDir, "review", f), filepath.Join(reviewDir, f))
	}
}

func copyFileIfExists(src, dst string) {
	data, err := os.ReadFile(src)
	if err != nil {
		return
	}
	_ = os.WriteFile(dst, data, 0644)
}

// linkOrCopyDir creates a symlink from src to dst. Falls back to copy on Windows if symlinks fail.
func linkOrCopyDir(src, dst string) {
	// Remove existing target
	os.RemoveAll(dst)

	// Try symlink first (works on most systems, needs dev mode on Windows)
	if err := os.Symlink(src, dst); err == nil {
		return
	}

	// Fallback: copy the directory
	copyDir(src, dst)
}

func copyDir(src, dst string) {
	_ = os.MkdirAll(dst, 0755)
	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyDir(srcPath, dstPath)
		} else {
			copyFileIfExists(srcPath, dstPath)
		}
	}
}

// pruneNonGitHubDirs removes platform-specific and development directories
// from the cloned .gstack/ staging area that are not needed for GitHub Copilot.
func pruneNonGitHubDirs(gstackDir string) {
	pruneList := []string{
		// Non-GitHub AI platform outputs
		".cursor", ".factory", ".kiro", ".openclaw", ".opencode", ".slate",
		"codex", "openclaw",
		// Build/dev artifacts
		"node_modules", ".git", ".github",
		// gstack internal infrastructure
		"extension", "hosts", "contrib", "supabase", "test", "scripts", "docs",
	}
	for _, dir := range pruneList {
		os.RemoveAll(filepath.Join(gstackDir, dir))
	}
}

// findBash locates a suitable bash executable.
// On Windows, Git Bash is preferred over WSL bash because gstack's setup script
// depends on native tools (bun, node) that are typically not installed inside WSL.
func findBash() string {
	if runtime.GOOS == "windows" {
		// Check Git for Windows locations first — these provide MSYS2 bash
		// which shares the native PATH (bun, node, etc.).
		candidates := []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Git", "bin", "bash.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Git", "bin", "bash.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Git", "bin", "bash.exe"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c
			}
		}
		// Fallback: try PATH, but skip WSL bash (System32\bash.exe)
		if path, err := exec.LookPath("bash"); err == nil {
			norm := strings.ToLower(filepath.Clean(path))
			if !strings.Contains(norm, "system32") && !strings.Contains(norm, "windowsapps") {
				return path
			}
		}
		return ""
	}
	// Non-Windows: just use PATH
	if path, err := exec.LookPath("bash"); err == nil {
		return path
	}
	return ""
}

// fixCRLFInScripts converts \r\n to \n in shell scripts to fix Windows git clone issues.
func fixCRLFInScripts(gstackDir string) {
	// Fix the main setup script
	setupPath := filepath.Join(gstackDir, "setup")
	fixCRLFInFile(setupPath)

	// Fix scripts in browse/scripts/
	scriptsDir := filepath.Join(gstackDir, "browse", "scripts")
	entries, err := os.ReadDir(scriptsDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			fixCRLFInFile(filepath.Join(scriptsDir, e.Name()))
		}
	}

	// Fix scripts in bin/
	binDir := filepath.Join(gstackDir, "bin")
	entries, err = os.ReadDir(binDir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			fixCRLFInFile(filepath.Join(binDir, e.Name()))
		}
	}
}

func fixCRLFInFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	fixed := bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	if !bytes.Equal(data, fixed) {
		_ = os.WriteFile(path, fixed, 0644)
	}
}

// stripBOMFromPackageJSON removes UTF-16/UTF-8 BOM from a package.json file.
// PowerShell's echo '{}' > file writes UTF-16 with BOM which breaks bun.
// Bun traverses up to the project root's package.json from .gstack/, so we
// need this file to be valid UTF-8 before running gstack's setup.
func stripBOMFromPackageJSON(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	// UTF-16 LE BOM: FF FE — need to convert entire file
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		// Decode UTF-16 LE to UTF-8: drop BOM, take every other byte (ASCII content)
		var utf8 []byte
		for i := 2; i+1 < len(data); i += 2 {
			if data[i+1] == 0 { // ASCII range
				utf8 = append(utf8, data[i])
			}
		}
		_ = os.WriteFile(path, utf8, 0644)
		return
	}
	// UTF-8 BOM: EF BB BF — just strip prefix
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		_ = os.WriteFile(path, data[3:], 0644)
	}
}
