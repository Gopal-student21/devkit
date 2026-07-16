package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// ESLint JSON output or line format:
// file.js
//   42:5  warning  Unexpected console statement  no-console
// Or compact: /path/file.js:42:5: warning - message [rule-id]
var eslintLineRegex = regexp.MustCompile(`^\s*(\d+):(\d+)\s+(error|warning|info)\s+(.+?)(?:\s+\S+\s*$|$)`)
var eslintCompactRegex = regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(error|warning|info)\s*[-–]\s*(.+)$`)

type ESLintParser struct{}

func (p *ESLintParser) Parse(lines []string) []Error {
	var errs []Error
	var currentFile string

	for _, line := range lines {
		stripped := stripAnsi(strings.TrimSpace(line))
		if stripped == "" {
			continue
		}

		// Compact format: file.js:42:5: warning - message
		if m := eslintCompactRegex.FindStringSubmatch(stripped); m != nil {
			lineNum, _ := strconv.Atoi(m[2])
			colNum, _ := strconv.Atoi(m[3])
			errs = append(errs, Error{
				File:    m[1],
				Line:    lineNum,
				Column:  colNum,
				Level:   m[4],
				Message: m[5],
				Raw:     stripped,
			})
			continue
		}

		// Check if this line is a file path (no indentation, ends with common ext)
		if !strings.HasPrefix(stripped, " ") && !strings.HasPrefix(stripped, "\t") {
			if strings.HasSuffix(stripped, ".js") || strings.HasSuffix(stripped, ".ts") ||
				strings.HasSuffix(stripped, ".jsx") || strings.HasSuffix(stripped, ".tsx") ||
				strings.HasSuffix(stripped, ".mjs") || strings.HasSuffix(stripped, ".cjs") ||
				strings.HasSuffix(stripped, ".vue") || strings.HasSuffix(stripped, ".svelte") {
				currentFile = stripped
				continue
			}
		}

		// Indented rule line: 42:5  warning  message  rule-name
		if m := eslintLineRegex.FindStringSubmatch(stripped); m != nil && currentFile != "" {
			lineNum, _ := strconv.Atoi(m[1])
			colNum, _ := strconv.Atoi(m[2])
			errs = append(errs, Error{
				File:    currentFile,
				Line:    lineNum,
				Column:  colNum,
				Level:   m[3],
				Message: strings.TrimSpace(m[4]),
				Raw:     stripped,
			})
		}
	}
	return errs
}
