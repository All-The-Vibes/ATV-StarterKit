package scaffold

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestSummarizeResults(t *testing.T) {
	results := []WriteResult{
		{Status: StatusCreated},
		{Status: StatusCreated},
		{Status: StatusDirCreated},
		{Status: StatusMerged},
		{Status: StatusSkipped},
		{Status: StatusFailed, Error: "boom"},
	}

	summary := SummarizeResults(results)
	if summary.Created != 2 || summary.Directories != 1 || summary.Merged != 1 || summary.Skipped != 1 || summary.Failed != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.Successful() {
		t.Fatal("summary with failures should not be successful")
	}
	if !strings.Contains(summary.Detail(), "2 files created") || !strings.Contains(summary.Detail(), "1 writes failed") {
		t.Fatalf("unexpected detail string: %q", summary.Detail())
	}
	if summary.FailureReason() == "" {
		t.Fatal("failure reason should be populated when writes fail")
	}
}

func TestIsExecutableScript(t *testing.T) {
	cases := []struct {
		path    string
		content []byte
		want    bool
	}{
		// Known extensions under scripts/ → executable
		{".github/skills/git-worktree/scripts/worktree-manager.sh", nil, true},
		{".github/skills/rclone/scripts/check_setup.sh", nil, true},
		{".github/skills/skill-creator/scripts/init_skill.py", nil, true},
		{".github/skills/gemini-imagegen/scripts/run_gemini.py", nil, true},
		{".github/skills/onboarding/scripts/inventory.mjs", nil, true},
		// Extensionless under scripts/ with shebang → executable
		{".github/skills/git-clean-gone-branches/scripts/clean-gone", []byte("#!/usr/bin/env bash\n"), true},
		// Extensionless under scripts/ without shebang → not executable
		{".github/skills/foo/scripts/data", []byte("plain text"), false},
		// Known extension OUTSIDE scripts/ → not executable (path restriction)
		{".github/skills/foo/templates/run.sh", nil, false},
		{".github/skills/foo/references/example.py", nil, false},
		// Non-script files
		{".github/skills/foo/SKILL.md", nil, false},
		{".github/skills/foo/templates/example.md", nil, false},
		{".github/copilot-instructions.md", nil, false},
		{".github/copilot-mcp-config.json", nil, false},
		{"Dockerfile", nil, false},
	}
	for _, tc := range cases {
		got := isExecutableScript(tc.path, tc.content)
		if got != tc.want {
			t.Errorf("isExecutableScript(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestWriteExecutableScriptHasExecBit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exec-bit semantics differ on windows")
	}
	tmp := t.TempDir()
	comps := []Component{
		{Path: ".github/skills/foo/scripts/run.sh", Content: []byte("#!/bin/sh\necho hi\n")},
		{Path: ".github/skills/foo/SKILL.md", Content: []byte("---\nname: foo\n---\n")},
	}
	results := WriteAll(tmp, comps)
	for _, r := range results {
		if r.Status == StatusFailed {
			t.Fatalf("unexpected failure for %s: %s", r.Path, r.Error)
		}
	}
	scriptInfo, err := os.Stat(filepath.Join(tmp, ".github/skills/foo/scripts/run.sh"))
	if err != nil {
		t.Fatalf("stat script: %v", err)
	}
	if scriptInfo.Mode().Perm()&0o111 == 0 {
		t.Errorf("expected exec bit on run.sh, got mode %v", scriptInfo.Mode().Perm())
	}
	mdInfo, err := os.Stat(filepath.Join(tmp, ".github/skills/foo/SKILL.md"))
	if err != nil {
		t.Fatalf("stat md: %v", err)
	}
	if mdInfo.Mode().Perm()&0o111 != 0 {
		t.Errorf("did not expect exec bit on SKILL.md, got mode %v", mdInfo.Mode().Perm())
	}
}
