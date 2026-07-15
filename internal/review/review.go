package review

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

func NewReviewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "AI-powered code review",
		Long: `Automated code review:
  dev review          — Review staged changes
  dev review --all    — Review all uncommitted changes
  dev review --file   — Review specific file`,
	}

	var all bool
	var file string

	cmd.Run = func(cmd *cobra.Command, args []string) {
		logger.Header("Code Review")

		if file != "" {
			reviewFile(file)
			return
		}

		if all {
			reviewAllChanges()
			return
		}

		reviewStagedChanges()
	}

	cmd.Flags().BoolVar(&all, "all", false, "Review all uncommitted changes")
	cmd.Flags().StringVar(&file, "file", "", "Review specific file")

	return cmd
}

func reviewStagedChanges() {
	logger.Step("Reviewing staged changes...")

	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("No staged changes. Use: git add <files>")
		return
	}

	if len(output) == 0 {
		logger.Warn("No staged changes to review")
		logger.Print("  Stage changes first: git add .")
		return
	}

	analyzeDiff(string(output))
}

func reviewAllChanges() {
	logger.Step("Reviewing all uncommitted changes...")

	cmd := exec.Command("git", "diff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error("No changes to review")
		return
	}

	if len(output) == 0 {
		logger.Warn("No uncommitted changes")
		return
	}

	analyzeDiff(string(output))
}

func reviewFile(filename string) {
	logger.Step(fmt.Sprintf("Reviewing %s...", filename))

	data, err := os.ReadFile(filename)
	if err != nil {
		logger.Error(fmt.Sprintf("Could not read %s", filename))
		return
	}

	content := string(data)
	issues := analyzeCode(content, filename)

	if len(issues) == 0 {
		logger.Success("No issues found")
		return
	}

	logger.Print("")
	for _, issue := range issues {
		logger.Warn(fmt.Sprintf("Line %d: %s", issue.line, issue.message))
		if issue.suggestion != "" {
			logger.Step(fmt.Sprintf("  Suggestion: %s", issue.suggestion))
		}
	}
}

type Issue struct {
	line       int
	message    string
	suggestion string
	severity   string
}

func analyzeCode(content string, filename string) []Issue {
	var issues []Issue
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		lineNum := i + 1

		if strings.Contains(line, "password") && strings.Contains(line, "=") && !strings.Contains(line, "env") {
			issues = append(issues, Issue{
				line:       lineNum,
				message:    "Possible hardcoded password",
				suggestion: "Use environment variables for secrets",
				severity:   "high",
			})
		}

		if strings.Contains(line, "TODO") || strings.Contains(line, "FIXME") {
			issues = append(issues, Issue{
				line:       lineNum,
				message:    "TODO/FIXME found",
				suggestion: "Address before merging",
				severity:   "low",
			})
		}

		if strings.Contains(line, "console.log") && filename != "" {
			issues = append(issues, Issue{
				line:       lineNum,
				message:    "console.log in production code",
				suggestion: "Remove or use a proper logger",
				severity:   "low",
			})
		}

		if strings.Contains(line, "eval(") {
			issues = append(issues, Issue{
				line:       lineNum,
				message:    "eval() usage detected",
				suggestion: "Avoid eval() for security reasons",
				severity:   "high",
			})
		}

		if strings.Contains(line, "any") && filename != "" && strings.HasSuffix(filename, ".ts") {
			issues = append(issues, Issue{
				line:       lineNum,
				message:    "TypeScript 'any' type used",
				suggestion: "Use proper type definitions",
				severity:   "medium",
			})
		}

		if strings.Contains(line, "catch") && strings.Contains(line, "{}") {
			issues = append(issues, Issue{
				line:       lineNum,
				message:    "Empty catch block",
				suggestion: "Handle errors properly",
				severity:   "medium",
			})
		}

		if len(line) > 120 {
			issues = append(issues, Issue{
				line:       lineNum,
				message:    "Line exceeds 120 characters",
				suggestion: "Break into multiple lines",
				severity:   "low",
			})
		}
	}

	return issues
}

func analyzeDiff(diff string) {
	lines := strings.Split(diff, "\n")
	issues := 0
	additions := 0
	deletions := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			additions++

			if strings.Contains(line, "password") && strings.Contains(line, "=") {
				logger.Warn("Added hardcoded password")
				issues++
			}
			if strings.Contains(line, "console.log") {
				logger.Warn("Added console.log")
				issues++
			}
			if strings.Contains(line, "eval(") {
				logger.Warn("Added eval()")
				issues++
			}
		}
		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			deletions++
		}
	}

	logger.Print("")
	logger.Step(fmt.Sprintf("Changes: +%d -%d lines", additions, deletions))

	if issues == 0 {
		logger.Success("No issues found in changes")
	} else {
		logger.Warn(fmt.Sprintf("Found %d potential issue(s)", issues))
	}
}
