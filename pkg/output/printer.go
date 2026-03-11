package output

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/github/atv-installer/pkg/detect"
	"github.com/github/atv-installer/pkg/scaffold"
)

//go:embed banner.txt
var bannerText string

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")). // bright blue
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")). // green
			Bold(true)

	skipStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243")) // dim gray

	mergeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")) // orange
)

// Printer handles terminal output with colored status indicators.
type Printer struct{}

// NewPrinter creates a new Printer.
func NewPrinter() *Printer {
	return &Printer{}
}

// PrintBanner shows the ATV ASCII art logo.
func (p *Printer) PrintBanner() {
	art := strings.TrimRight(bannerText, "\n\r ")

	bannerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true)

	fmt.Println(bannerStyle.Render(art))
	fmt.Println()
	fmt.Println(titleStyle.Render("  ATV Starter Kit — Agentic Tool & Workflow"))
	fmt.Println()
}

// PrintDetection shows the detected environment.
func (p *Printer) PrintDetection(env detect.Environment) {
	repoType := "empty directory"
	if env.IsGitRepo {
		repoType = "existing git repo"
	}
	fmt.Printf("  Auto-detected: %s project (%s, %s)\n\n",
		titleStyle.Render(string(env.Stack)), env.StackHint, repoType)
}

// PrintResults shows what was created, skipped, or merged.
func (p *Printer) PrintResults(results []scaffold.WriteResult) {
	created := 0
	skipped := 0
	merged := 0
	dirs := 0

	for _, r := range results {
		switch r.Status {
		case scaffold.StatusCreated:
			fmt.Printf("  %s %s\n", successStyle.Render("✅"), r.Path)
			created++
		case scaffold.StatusSkipped:
			fmt.Printf("  %s %s\n", skipStyle.Render("⏭️  "+r.Path+" (exists)"), "")
			skipped++
		case scaffold.StatusMerged:
			fmt.Printf("  %s %s\n", mergeStyle.Render("🔀 "+r.Path+" (merged)"), "")
			merged++
		case scaffold.StatusDirCreated:
			fmt.Printf("  %s %s\n", successStyle.Render("📁"), r.Path)
			dirs++
		}
	}

	fmt.Println()
	if skipped > 0 {
		fmt.Printf("  %s\n", skipStyle.Render(fmt.Sprintf("⏭️  Skipped %d existing files", skipped)))
	}
	if merged > 0 {
		fmt.Printf("  %s\n", mergeStyle.Render(fmt.Sprintf("🔀 Merged %d JSON configs", merged)))
	}
	fmt.Printf("  %s\n", successStyle.Render(fmt.Sprintf("✅ Created %d files, %d directories", created, dirs)))
}

// PrintNextSteps shows post-install guidance.
func (p *Printer) PrintNextSteps(stack detect.Stack) {
	fmt.Println()
	fmt.Println(successStyle.Render("  🎉 ATV Starter Kit ready!"))
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Println(titleStyle.Render("    1.") + " Open this folder in VS Code")
	fmt.Println(titleStyle.Render("    2.") + " Install recommended extensions when prompted")
	fmt.Println(titleStyle.Render("    3.") + ` Try: /ce-brainstorm "your first feature idea"`)
	fmt.Println()
}

// Info prints an informational message.
func (p *Printer) Info(msg string) {
	fmt.Printf("  %s\n", msg)
}
