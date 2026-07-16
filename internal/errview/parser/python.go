package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// Python traceback:
//   File "app/main.py", line 42, in handler
//     result = foo(bar)
// Python error on single line:
//   app/main.py:42:5: Error: message
var pyFileRegex = regexp.MustCompile(`^\s*File "(.+?)", line (\d+)(?:, in (.+))?$`)
var pyColonRegex = regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(error|warning|Error|Warning):\s*(.+)$`)
var pySimpleRegex = regexp.MustCompile(`^(.+?):(\d+):\s*(error|warning|Error|Warning):\s*(.+)$`)

type PythonParser struct{}

func (p *PythonParser) Parse(lines []string) []Error {
	var errs []Error

	for i := 0; i < len(lines); i++ {
		stripped := stripAnsi(strings.TrimSpace(lines[i]))
		if stripped == "" {
			continue
		}

		// Traceback format: File "path", line N, in func
		if m := pyFileRegex.FindStringSubmatch(stripped); m != nil {
			lineNum, _ := strconv.Atoi(m[2])
			// Next line usually has the code, skip it, error message follows
			var msg string
			// Look ahead for IndentationError, NameError, etc.
			for j := i + 1; j < len(lines) && j <= i+5; j++ {
				next := strings.TrimSpace(stripAnsi(lines[j]))
				if strings.Contains(next, "Error") || strings.Contains(next, "Exception") || strings.Contains(next, "Traceback") {
					if !strings.HasPrefix(next, "File ") {
						msg = next
						break
					}
				}
			}
			if msg == "" && i+1 < len(lines) {
				msg = strings.TrimSpace(stripAnsi(lines[i+1]))
			}
			errs = append(errs, Error{
				File:    m[1],
				Line:    lineNum,
				Column:  0,
				Level:   "error",
				Message: msg,
				Raw:     stripped,
			})
			continue
		}

		// Colon format: file.py:42:5: Error: msg
		if m := pyColonRegex.FindStringSubmatch(stripped); m != nil {
			lineNum, _ := strconv.Atoi(m[2])
			colNum, _ := strconv.Atoi(m[3])
			errs = append(errs, Error{
				File:    m[1],
				Line:    lineNum,
				Column:  colNum,
				Level:   strings.ToLower(m[4]),
				Message: m[5],
				Raw:     stripped,
			})
			continue
		}

		// Simple format: file.py:42: Error: msg
		if m := pySimpleRegex.FindStringSubmatch(stripped); m != nil {
			lineNum, _ := strconv.Atoi(m[2])
			errs = append(errs, Error{
				File:    m[1],
				Line:    lineNum,
				Column:  0,
				Level:   strings.ToLower(m[3]),
				Message: m[4],
				Raw:     stripped,
			})
		}
	}
	return errs
}
