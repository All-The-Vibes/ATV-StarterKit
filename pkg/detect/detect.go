package detect

import (
	"os"
	"path/filepath"
)

// Stack represents the detected project stack.
type Stack string

const (
	StackRails      Stack = "rails"
	StackPython     Stack = "python"
	StackTypeScript Stack = "typescript"
	StackGeneral    Stack = "general"
)

// Environment holds detection results.
type Environment struct {
	Stack     Stack
	IsGitRepo bool
	StackHint string // human-readable detection reason
}

// DetectEnvironment scans the target directory for stack indicators.
func DetectEnvironment(dir string) Environment {
	env := Environment{Stack: StackGeneral}

	// Check for git repo
	if exists(filepath.Join(dir, ".git")) {
		env.IsGitRepo = true
	}

	// Rails: Gemfile + config/routes.rb
	if exists(filepath.Join(dir, "Gemfile")) && exists(filepath.Join(dir, "config", "routes.rb")) {
		env.Stack = StackRails
		env.StackHint = "Gemfile + config/routes.rb found"
		return env
	}

	// TypeScript: tsconfig.json
	if exists(filepath.Join(dir, "tsconfig.json")) {
		env.Stack = StackTypeScript
		env.StackHint = "tsconfig.json found"
		return env
	}

	// Python: pyproject.toml or requirements.txt
	if exists(filepath.Join(dir, "pyproject.toml")) {
		env.Stack = StackPython
		env.StackHint = "pyproject.toml found"
		return env
	}
	if exists(filepath.Join(dir, "requirements.txt")) {
		env.Stack = StackPython
		env.StackHint = "requirements.txt found"
		return env
	}

	// General fallback
	env.StackHint = "no stack-specific files detected"
	return env
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
