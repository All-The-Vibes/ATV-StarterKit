package monitor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// allowedPrefixes is the command allowlist. Only commands starting with these
// prefixes may be executed. This prevents arbitrary command execution.
var allowedPrefixes = []string{
	"atv-inst-tool ",
	"atv-installer ",
	"atv ",
	"gstack ",
	"git ",
	"copilot ",
}

// dangerousChars are shell metacharacters that could enable injection.
var dangerousChars = []string{";", "|", "&&", "||", "$(", "`", ">", "<", "\n", "\r"}

// Executor runs approved actions sequentially with safety validation.
type Executor struct {
	root    string
	mu      sync.Mutex
	active  bool
	timeout time.Duration
}

// NewExecutor creates an action executor for the given workspace root.
func NewExecutor(root string) *Executor {
	return &Executor{
		root:    root,
		timeout: 2 * time.Minute,
	}
}

// IsActive reports whether an action is currently executing.
func (e *Executor) IsActive() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.active
}

// Execute validates and runs a proposed action, returning the result.
// Only one action may execute at a time; concurrent calls will be rejected.
func (e *Executor) Execute(ctx context.Context, action ProposedAction) (*ActionResult, error) {
	if err := ValidateAction(action); err != nil {
		return &ActionResult{
			Action:     action,
			Success:    false,
			Error:      err.Error(),
			ExecutedAt: time.Now(),
		}, err
	}

	e.mu.Lock()
	if e.active {
		e.mu.Unlock()
		return nil, fmt.Errorf("another action is already executing")
	}
	e.active = true
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.active = false
		e.mu.Unlock()
	}()

	// Build command
	ctx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	parts := splitCommand(action.Command, action.Args)
	if len(parts) == 0 {
		return &ActionResult{
			Action:     action,
			Success:    false,
			Error:      "empty command",
			ExecutedAt: time.Now(),
		}, fmt.Errorf("empty command")
	}

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Dir = e.root

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := &ActionResult{
		Action:     action,
		ExecutedAt: time.Now(),
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		if stderr.Len() > 0 {
			result.Output = stderr.String()
		}
	} else {
		result.Success = true
		result.Output = stdout.String()
	}

	return result, nil
}

// ValidateAction checks whether an action is safe to execute.
func ValidateAction(action ProposedAction) error {
	cmd := action.Command

	// Check against allowlist
	allowed := false
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(cmd, prefix) || strings.TrimSpace(cmd)+" " == prefix {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("command %q not in allowlist", cmd)
	}

	// Check for shell metacharacters in command and args
	allParts := append([]string{cmd}, action.Args...)
	for _, part := range allParts {
		for _, ch := range dangerousChars {
			if strings.Contains(part, ch) {
				return fmt.Errorf("rejected: dangerous character %q in %q", ch, part)
			}
		}
	}

	// Check for path traversal
	for _, part := range allParts {
		if strings.Contains(part, "..") {
			return fmt.Errorf("rejected: path traversal in %q", part)
		}
	}

	return nil
}

// splitCommand splits a command string plus extra args into exec-ready parts.
func splitCommand(cmd string, args []string) []string {
	parts := strings.Fields(cmd)
	parts = append(parts, args...)
	return parts
}
