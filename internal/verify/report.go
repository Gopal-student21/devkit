package verify

import (
	"fmt"
	"strings"

	"github.com/devkit/devkit/pkg/logger"
)

// Render prints the visual verification report.
func Render(r Report) {
	logger.Header("Verification Report")

	fmt.Printf("%s▸%s Mode: %s\n", logger.Cyan, logger.Reset, r.Mode)

	if r.Total == 0 {
		fmt.Println()
		logger.Success("No issues found")
		fmt.Printf("  %sRisk%s: Low   %sAI%s: %s\n", logger.Bold, logger.Reset, logger.Bold, logger.Reset, r.AIlikelihood)
		printVerdict(r.Verdict)
		fmt.Println()
		return
	}

	fmt.Printf("  %s%d finding(s)%s  %s%d high%s  %s%d medium%s  %s%d low%s  %s%d AI marker(s)%s\n",
		logger.Bold, r.Total, logger.Reset,
		logger.Red, r.High, logger.Reset,
		logger.Yellow, r.Medium, logger.Reset,
		logger.Reset, r.Low, logger.Reset,
		logger.Cyan, r.AIMarkers, logger.Reset)

	fmt.Printf("  %sRisk:%s %-9s  %sAI:%s %s\n",
		logger.Bold, logger.Reset, r.Risk,
		logger.Bold, logger.Reset, r.AIlikelihood)

	for _, fr := range r.Files {
		fmt.Println()
		fmt.Printf("%s▸ %s%s\n", logger.Bold, fr.File, logger.Reset)
		for _, f := range fr.Findings {
			printFinding(f)
		}
	}

	fmt.Println()
	printVerdict(r.Verdict)
	fmt.Println()
}

func printFinding(f Finding) {
	badge := severityBadge(f.Severity)
	tag := ""
	if f.Category == "ai" {
		tag = fmt.Sprintf("  %s[AI]%s", logger.Cyan, logger.Reset)
	}
	fmt.Printf("  %s %s line %d%s%s\n", badge, strings.ToUpper(f.Message), f.Line, tag, codeSnippet(f.Code))
	if f.Suggestion != "" {
		fmt.Printf("      %s▸%s %s\n", logger.Cyan, logger.Reset, f.Suggestion)
	}
}

func codeSnippet(code string) string {
	if code == "" {
		return ""
	}
	return fmt.Sprintf("  ·  %s", truncate(code, 60))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func severityBadge(sev string) string {
	switch sev {
	case "high":
		return fmt.Sprintf("%s[!]%s", logger.Red, logger.Reset)
	case "medium":
		return fmt.Sprintf("%s[!]%s", logger.Yellow, logger.Reset)
	default:
		return fmt.Sprintf("[·]")
	}
}

func printVerdict(verdict string) {
	switch verdict {
	case "PASS":
		logger.Success(fmt.Sprintf("VERDICT: %s", verdict))
	case "ACTION NEEDED":
		logger.Warn(fmt.Sprintf("VERDICT: %s — review findings before merge", verdict))
	default:
		logger.Error(fmt.Sprintf("VERDICT: %s — security issues found, do not merge", verdict))
	}
}
