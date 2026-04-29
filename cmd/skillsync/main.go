package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/All-The-Vibes/ATV-StarterKit/pkg/skillsync"
)

func main() {
	manifestPath := flag.String("manifest", "", "Path to pkg/scaffold/skill_sources.json. Defaults to the repo manifest.")
	localPath := flag.String("local", "", "Path to local ATV skill templates. Defaults to pkg/scaffold/templates/skills.")
	upstreamPath := flag.String("upstream", "", "Path to upstream Compound Engineering skills directory.")
	format := flag.String("format", "text", "Output format: text or json.")
	flag.Parse()

	if *upstreamPath == "" {
		exit("-upstream is required and should point at the Compound Engineering skills directory")
	}
	repoRoot, err := findRepoRoot()
	if err != nil {
		exit("locate repo root: %v", err)
	}
	if *manifestPath == "" {
		*manifestPath = filepath.Join(repoRoot, "pkg", "scaffold", "skill_sources.json")
	}
	if *localPath == "" {
		*localPath = filepath.Join(repoRoot, "pkg", "scaffold", "templates", "skills")
	}

	results, err := skillsync.Compare(*manifestPath, *localPath, *upstreamPath)
	if err != nil {
		exit("compare skills: %v", err)
	}
	switch *format {
	case "json":
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			exit("encode results: %v", err)
		}
		fmt.Println(string(data))
	case "text":
		printText(results)
	default:
		exit("unknown -format %q; use text or json", *format)
	}
}

func printText(results []skillsync.Result) {
	counts := map[string]int{}
	for _, result := range results {
		counts[result.Status]++
		localName := result.LocalName
		if localName == "" {
			localName = "-"
		}
		upstreamName := result.UpstreamName
		if upstreamName == "" {
			upstreamName = "-"
		}
		fmt.Printf("%-20s %-20s %-18s %s\n", localName, upstreamName, result.Status, result.Tracking)
	}
	fmt.Println()
	for _, status := range []string{
		skillsync.StatusStale,
		skillsync.StatusOverlay,
		skillsync.StatusInSync,
		skillsync.StatusMissingLocal,
		skillsync.StatusMissingUpstream,
		skillsync.StatusUntrackedUpstream,
	} {
		if counts[status] > 0 {
			fmt.Printf("%s=%d\n", status, counts[status])
		}
	}
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no go.mod found at or above %s", wd)
		}
		dir = parent
	}
}

func exit(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	message = strings.TrimSpace(message)
	fmt.Fprintf(os.Stderr, "skillsync: %s\n", message)
	os.Exit(1)
}
