package verify

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/devkit/devkit/internal/detect"
	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

// Finding is a single verification result attached to a line.
type Finding struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Rule       string `json:"rule"`
	Category   string `json:"category"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
	Code       string `json:"code,omitempty"`
}

// FileReport groups findings per file.
type FileReport struct {
	File      string    `json:"file"`
	Findings  []Finding `json:"findings"`
	LineCount int       `json:"lines"`
}

// Report is the full verification output.
type Report struct {
	Mode         string                 `json:"mode"`
	Files        []FileReport           `json:"files"`
	Total        int                    `json:"totalFindings"`
	High         int                    `json:"high"`
	Medium       int                    `json:"medium"`
	Low          int                    `json:"low"`
	AIMarkers    int                    `json:"aiMarkers"`
	AIlikelihood string                 `json:"aiLikelihood"`
	RiskScore    int                    `json:"riskScore"`
	Risk         string                 `json:"risk"`
	Verdict      string                 `json:"verdict"`
	Strict       bool                   `json:"strict"`
	Attribution  map[string]attribution `json:"attribution,omitempty"`
	AiLines      int                    `json:"aiLines"`
	AiTotal      int                    `json:"aiTotal"`
	AiSignal     string                 `json:"aiSignal"`
	Tests        TestResult             `json:"tests,omitempty"`
}

const (
	weightHigh   = 10
	weightMedium = 3
	weightLow    = 1
)

// NewVerifyCommand builds the `dev verify` command.
func NewVerifyCommand() *cobra.Command {
	var all bool
	var file string
	var strict bool
	var ci bool
	var runTestsFlag bool
	var htmlOut string
	var openHTML bool

	cmd := &cobra.Command{
		Use:   "verify [path]",
		Short: "Verification layer for AI-generated code",
		Long: `Verify code before it ships: security checks, code-smell detection,
and AI-author attribution with a visual risk report.

  dev verify           — verify staged changes (default)
  dev verify --all     — verify all uncommitted changes
  dev verify --file x  — verify a specific file
  dev verify --tests   — also run the project test suite (verify before merge)
  dev verify --html r  — write a visual HTML report
  dev verify --open    — write and open the HTML report in your browser
  dev verify --ci      — machine-readable JSON, exit 1 on issues (CI gate)
  dev verify --strict  — fail on any finding (not just high severity)`,
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var report Report

			switch {
			case file != "":
				report = verifyFile(file, strict)
			case all:
				report = verifyDiff(runGit("git", "diff"), "uncommitted", strict)
			default:
				report = verifyDiff(runGit("git", "diff", "--cached"), "staged", strict)
			}

			if runTestsFlag {
				stack := detect.DetectStack()
				report.Tests = runTests(stack.Type, 90*time.Second)
				finalizeVerdict(&report)
			}

			if htmlOut != "" || openHTML {
				path := htmlOut
				if path == "" {
					path = "verify-report.html"
				}
				var err error
				if openHTML {
					err = OpenHTML(report, path)
				} else {
					err = WriteHTML(report, path)
				}
				if err != nil {
					logger.Error(fmt.Sprintf("Could not write report: %v", err))
				} else {
					logger.Success(fmt.Sprintf("Visual report: %s", path))
				}
			}

			if ci {
				out, _ := json.MarshalIndent(report, "", "  ")
				fmt.Println(string(out))
			} else {
				Render(report)
			}

			if report.Verdict == "BLOCK" {
				os.Exit(1)
			}
			if strict && report.Total > 0 {
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Verify all uncommitted changes")
	cmd.Flags().StringVar(&file, "file", "", "Verify a specific file")
	cmd.Flags().BoolVar(&strict, "strict", false, "Exit non-zero on any finding")
	cmd.Flags().BoolVar(&ci, "ci", false, "JSON output for CI pipelines")
	cmd.Flags().BoolVar(&runTestsFlag, "tests", false, "Run the project test suite as part of verification")
	cmd.Flags().StringVar(&htmlOut, "html", "", "Write a visual HTML report to this path")
	cmd.Flags().BoolVar(&openHTML, "open", false, "Write and open the visual HTML report")

	return cmd
}

func runGit(args ...string) string {
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return string(out)
}

// scanLines applies all rules to the given (file, line) pairs.
func scanLines(file string, lineNums []int, lines []string) []Finding {
	var findings []Finding
	for idx, line := range lines {
		for _, rule := range Rules {
			if rule.Match(line) {
				findings = append(findings, Finding{
					File:       file,
					Line:       lineNums[idx],
					Rule:       rule.Name,
					Category:   rule.Category,
					Severity:   rule.Severity,
					Message:    rule.Name,
					Suggestion: rule.Suggestion,
					Code:       strings.TrimSpace(line),
				})
			}
		}
	}
	return findings
}

// fileContent holds added lines and their real file line numbers.
type fileContent struct {
	contents []string
	nums     []int
}

// analyze takes per-file content and produces a full report.
func analyze(mode string, files map[string]fileContent, strict bool) Report {
	var report Report
	report.Mode = mode
	report.Strict = strict

	for f, fc := range files {
		if len(fc.contents) == 0 {
			continue
		}
		findings := scanLines(f, fc.nums, fc.contents)
		if len(findings) == 0 {
			continue
		}
		report.Files = append(report.Files, FileReport{
			File:      f,
			Findings:  findings,
			LineCount: len(fc.contents),
		})
	}

	sort.Slice(report.Files, func(i, j int) bool {
		return report.Files[i].File < report.Files[j].File
	})

	for _, fr := range report.Files {
		for _, f := range fr.Findings {
			report.Total++
			switch f.Severity {
			case "high":
				report.High++
				report.RiskScore += weightHigh
			case "medium":
				report.Medium++
				report.RiskScore += weightMedium
			default:
				report.Low++
				report.RiskScore += weightLow
			}
			if f.Category == "ai" {
				report.AIMarkers++
			}
		}
	}

	report.Risk = riskLabel(report.RiskScore)
	report.AIlikelihood = aiLabel(report.AIMarkers, report.Total)

	perFile := scoreAttribution(files)
	agg := averageAttribution(perFile)
	report.Attribution = perFile
	report.AiLines = agg.AiLines
	report.AiTotal = agg.Total
	report.AiSignal = agg.Label

	finalizeVerdict(&report)

	return report
}

// finalizeVerdict derives the gate verdict from findings + test results.
func finalizeVerdict(r *Report) {
	if r.High > 0 {
		r.Verdict = "BLOCK"
		return
	}
	if r.Tests.Ran && !r.Tests.Passed {
		r.Verdict = "BLOCK"
		return
	}
	if r.Total > 0 {
		r.Verdict = "ACTION NEEDED"
		return
	}
	r.Verdict = "PASS"
}

func riskLabel(score int) string {
	switch {
	case score >= 20:
		return "Critical"
	case score >= 10:
		return "High"
	case score >= 4:
		return "Moderate"
	default:
		return "Low"
	}
}

func aiLabel(markers, total int) string {
	if markers == 0 {
		return "No AI markers"
	}
	ratio := float64(markers) / float64(max(total, 1))
	switch {
	case ratio >= 0.5:
		return "Heavily AI-generated"
	case ratio >= 0.25:
		return "Likely AI-generated"
	default:
		return "Partially AI-generated"
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// verifyDiff analyzes added lines of a git diff.
func verifyDiff(diff, mode string, strict bool) Report {
	files := parseDiff(diff)
	if len(files) == 0 {
		logger.Warn(fmt.Sprintf("No %s changes to verify", mode))
		logger.Print("  Stage changes first: git add .")
		return Report{Mode: mode, Verdict: "PASS", Strict: strict}
	}
	return analyze(mode, files, strict)
}

// verifyFile analyzes a single file on disk.
func verifyFile(filename string, strict bool) Report {
	data, err := os.ReadFile(filename)
	if err != nil {
		logger.Error(fmt.Sprintf("Could not read %s", filename))
		os.Exit(1)
	}
	lines := strings.Split(string(data), "\n")
	nums := make([]int, len(lines))
	for i := range lines {
		nums[i] = i + 1
	}
	files := map[string]fileContent{
		filename: {contents: lines, nums: nums},
	}
	return analyze(filename, files, strict)
}

// parseDiff converts a unified git diff into file → added lines with
// real new-file line numbers (parsed from @@ hunk headers).
func parseDiff(diff string) map[string]fileContent {
	files := map[string]fileContent{}
	var current string
	newLine := 0

	hunkRe := regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@`)

	for _, raw := range strings.Split(diff, "\n") {
		line := strings.TrimRight(raw, "\r")

		if m := hunkRe.FindStringSubmatch(line); m != nil {
			fmt.Sscanf(m[1], "%d", &newLine)
			continue
		}
		if strings.HasPrefix(line, "+++ ") {
			p := strings.TrimPrefix(line, "+++ ")
			p = strings.TrimPrefix(p, "b/")
			current = p
			files[current] = fileContent{}
			continue
		}
		if strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "diff --git ") ||
			strings.HasPrefix(line, "index ") || strings.HasPrefix(line, "new file mode ") ||
			strings.HasPrefix(line, "deleted file mode ") {
			continue
		}
		if current == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "+"):
			fc := files[current]
			fc.contents = append(fc.contents, strings.TrimPrefix(line, "+"))
			fc.nums = append(fc.nums, newLine)
			files[current] = fc
			newLine++
		case strings.HasPrefix(line, " "):
			newLine++
		}
	}
	return files
}
