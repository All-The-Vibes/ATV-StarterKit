package monitor

import (
	"context"
	"testing"
)

func TestValidateAction_Allowed(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"atv-inst-tool", "atv-inst-tool init"},
		{"atv-installer", "atv-installer init"},
		{"atv", "atv dashboard"},
		{"gstack", "gstack build"},
		{"git", "git status"},
		{"copilot", "copilot auth"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAction(ProposedAction{Command: tt.cmd})
			if err != nil {
				t.Errorf("ValidateAction(%q) = %v, want nil", tt.cmd, err)
			}
		})
	}
}

func TestValidateAction_NotAllowed(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
	}{
		{"rm", "rm -rf /"},
		{"bash", "bash -c 'echo pwned'"},
		{"curl", "curl evil.com"},
		{"python", "python -c 'import os'"},
		{"empty", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAction(ProposedAction{Command: tt.cmd})
			if err == nil {
				t.Errorf("ValidateAction(%q) = nil, want error", tt.cmd)
			}
		})
	}
}

func TestValidateAction_DangerousChars(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		args []string
	}{
		{"semicolon", "git status; rm -rf", nil},
		{"pipe", "git log | head", nil},
		{"and", "git pull && echo done", nil},
		{"backtick in arg", "git commit", []string{"-m", "`whoami`"}},
		{"dollar-paren in arg", "git commit", []string{"-m", "$(id)"}},
		{"redirect", "git log > /tmp/out", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAction(ProposedAction{Command: tt.cmd, Args: tt.args})
			if err == nil {
				t.Errorf("ValidateAction(%q, %v) = nil, want error", tt.cmd, tt.args)
			}
		})
	}
}

func TestValidateAction_PathTraversal(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		args []string
	}{
		{"dotdot in cmd", "git checkout ../../etc/passwd", nil},
		{"dotdot in arg", "git checkout", []string{"../../etc/passwd"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAction(ProposedAction{Command: tt.cmd, Args: tt.args})
			if err == nil {
				t.Errorf("ValidateAction(%q, %v) = nil, want error", tt.cmd, tt.args)
			}
		})
	}
}

func TestNewExecutor(t *testing.T) {
	e := NewExecutor(t.TempDir())
	if e.active {
		t.Error("new executor should not be active")
	}
	if e.timeout == 0 {
		t.Error("timeout should be set")
	}
}

func TestExecutor_RejectsInvalidAction(t *testing.T) {
	e := NewExecutor(t.TempDir())
	result, err := e.Execute(context.Background(), ProposedAction{
		Command: "rm -rf /",
	})
	if err == nil {
		t.Error("expected error for invalid command")
	}
	if result == nil {
		t.Fatal("expected result even on validation failure")
	}
	if result.Success {
		t.Error("result should not be successful")
	}
}

func TestExecutor_SplitCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		args []string
		want int
	}{
		{"git status", nil, 2},
		{"git", []string{"status"}, 2},
		{"git commit", []string{"-m", "test"}, 4},
	}
	for _, tt := range tests {
		parts := splitCommand(tt.cmd, tt.args)
		if len(parts) != tt.want {
			t.Errorf("splitCommand(%q, %v) = %d parts, want %d", tt.cmd, tt.args, len(parts), tt.want)
		}
	}
}

func TestExecutor_IsActive(t *testing.T) {
	e := NewExecutor(t.TempDir())
	if e.IsActive() {
		t.Error("should not be active initially")
	}
}
