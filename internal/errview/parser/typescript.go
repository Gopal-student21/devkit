package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// TypeScript: file.ts(42,5): error TS2345: ...
// tsc: src/file.ts:42:5 - error TS2345: ...
var tsRegex = regexp.MustCompile(`^(.+?)\((\d+),(\d+)\):\s*(error|warning)\s+(TS\d+):\s*(.+)$`)
var tsFileRegex = regexp.MustCompile(`^(.+?):(\d+):(\d+)\s*-\s*(error|warning)\s+(TS\d+):\s*(.+)$`)

type TSParser struct{}

func (p *TSParser) Parse(lines []string) []Error {
	var errs []Error
	for _, line := range lines {
		stripped := stripAnsi(strings.TrimSpace(line))
		if stripped == "" {
			continue
		}

		var m []string

		if groups := tsRegex.FindStringSubmatch(stripped); groups != nil {
			m = groups
		} else if groups := tsFileRegex.FindStringSubmatch(stripped); groups != nil {
			m = groups
		}

		if m != nil {
			lineNum, _ := strconv.Atoi(m[2])
			colNum, _ := strconv.Atoi(m[3])
			errs = append(errs, Error{
				File:    m[1],
				Line:    lineNum,
				Column:  colNum,
				Level:   m[4],
				Message: m[4] + " " + m[5] + ": " + m[6],
				Raw:     stripped,
			})
		}
	}
	return errs
}
