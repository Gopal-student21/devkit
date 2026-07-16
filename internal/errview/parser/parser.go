package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// Error represents a single parsed error with location context
type Error struct {
	File    string
	Line    int
	Column  int
	Message string
	Level   string // error, warning, info
	Raw     string // original line
}

// Parser extracts errors from tool output
type Parser interface {
	Parse(lines []string) []Error
}

// Detect auto-detects which parser to use based on the output content
func Detect(lines []string) Parser {
	joined := strings.Join(lines, "\n")

	// Check for specific patterns in priority order
	goHints := []string{"undefined:", "cannot ", "not enough ", "too many ", "syntax error", "imported and not used", "overflow"}
	for _, hint := range goHints {
		if strings.Contains(joined, hint) && strings.Contains(joined, ".go:") {
			return &GoParser{}
		}
	}

	if strings.Contains(joined, "error TS") || strings.Contains(joined, ".ts(") || strings.Contains(joined, ".tsx(") {
		return &TSParser{}
	}

	if strings.Contains(joined, "Traceback") || (strings.Contains(joined, ".py") && strings.Contains(joined, "Error")) {
		return &PythonParser{}
	}

	if strings.Contains(joined, "error[") {
		return &RustParser{}
	}

	// ESLint: has file paths + indented lines with "error"/"warning" followed by rule names
	if hasESLintPattern(lines) {
		return &ESLintParser{}
	}

	// GCC/Clang: file:line:col: error/warning: message
	if strings.Contains(joined, "error:") || strings.Contains(joined, "warning:") || strings.Contains(joined, "fatal error:") {
		return &GCCParser{}
	}

	return &GenericParser{}
}

// hasESLintPattern checks for ESLint-style output: file path followed by indented lines
func hasESLintPattern(lines []string) bool {
	fileExts := []string{".js", ".ts", ".jsx", ".tsx", ".mjs", ".cjs", ".vue", ".svelte"}
	fileCount := 0
	errorLineCount := 0

	for _, line := range lines {
		s := strings.TrimSpace(line)
		if s == "" {
			continue
		}
		// Count file path lines
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			for _, ext := range fileExts {
				if strings.HasSuffix(s, ext) {
					fileCount++
					break
				}
			}
		}
		// Count indented error/warning lines with line:col pattern
		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			re := regexp.MustCompile(`^\s+\d+:\d+\s+(error|warning|info)\s+`)
			if re.MatchString(line) {
				errorLineCount++
			}
		}
	}

	return fileCount > 0 && errorLineCount > 0
}

// ReadInput reads from stdin or files
func ReadInput(args []string) []string {
	if len(args) == 0 {
		return readStdin()
	}
	var all []string
	for _, arg := range args {
		f, err := os.Open(arg)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			all = append(all, scanner.Text())
		}
		f.Close()
	}
	return all
}

func readStdin() []string {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil
	}
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// stripAnsi removes ANSI color codes from a string
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}
