package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// Generic: try to match any file:line:col pattern
var genericRegex = regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(.+)$`)
var genericShortRegex = regexp.MustCompile(`^(.+?):(\d+):\s*(.+)$`)

type GenericParser struct{}

func (p *GenericParser) Parse(lines []string) []Error {
	var errs []Error
	for _, line := range lines {
		stripped := stripAnsi(strings.TrimSpace(line))
		if stripped == "" {
			continue
		}

		if m := genericRegex.FindStringSubmatch(stripped); m != nil {
			lineNum, _ := strconv.Atoi(m[2])
			colNum, _ := strconv.Atoi(m[3])
			errs = append(errs, Error{
				File:    m[1],
				Line:    lineNum,
				Column:  colNum,
				Level:   "error",
				Message: m[4],
				Raw:     stripped,
			})
		} else if m := genericShortRegex.FindStringSubmatch(stripped); m != nil {
			lineNum, _ := strconv.Atoi(m[2])
			errs = append(errs, Error{
				File:    m[1],
				Line:    lineNum,
				Column:  0,
				Level:   "error",
				Message: m[3],
				Raw:     stripped,
			})
		}
	}
	return errs
}
