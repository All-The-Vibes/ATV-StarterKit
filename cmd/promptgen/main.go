// Command promptgen regenerates the dogfooded VS Code Copilot Chat prompt
// shims at .github/prompts/<skill>.prompt.md from the canonical
// scaffold.BuildPromptShim template.
//
// Usage:
//
//	go run ./cmd/promptgen          # regenerate .github/prompts/
//	go run ./cmd/promptgen -check   # CI mode: exit 1 if regeneration would change anything
//
// Run from any directory inside the repository — the tool walks up to find
// go.mod. Pair with TestDogfoodPromptParity in pkg/scaffold/parity_test.go.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/scaffold"
)

func main() {
	check := flag.Bool("check", false, "Verify committed .github/prompts/ matches the generator output. Exit 1 on drift.")
	flag.Parse()

	repoRoot, err := findRepoRoot()
	if err != nil {
		exit("locate repo root: %v", err)
	}

	plan := scaffold.PromptShimPlan()
	promptsDir := filepath.Join(repoRoot, ".github", "prompts")
	if err := os.MkdirAll(promptsDir, 0o755); err != nil {
		exit("create %s: %v", promptsDir, err)
	}

	drift := false
	for _, name := range plan {
		dest := filepath.Join(promptsDir, name+".prompt.md")
		want := scaffold.BuildPromptShim(name)
		if *check {
			got, err := os.ReadFile(dest)
			if err != nil || string(got) != string(want) {
				drift = true
				fmt.Fprintf(os.Stderr, "drift: %s\n", dest)
			}
			continue
		}
		if err := os.WriteFile(dest, want, 0o644); err != nil {
			exit("write %s: %v", dest, err)
		}
		fmt.Printf("wrote %s\n", dest)
	}

	if *check && drift {
		fmt.Fprintln(os.Stderr, "Run `go run ./cmd/promptgen` to regenerate.")
		os.Exit(1)
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found from %s upward", dir)
		}
		dir = parent
	}
}

func exit(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "promptgen: "+format+"\n", args...)
	os.Exit(1)
}
