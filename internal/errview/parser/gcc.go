package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// GCC/Clang: file.c:42:5: error: undeclared identifier
var gccRegex = regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(error|warning|note|fatal error):\s*(.+)$`)

type GCCParser struct{}

func (p *GCCParser) Parse(lines []string) []Error {
	var errs []Error
	for _, line := range lines {
		stripped := stripAnsi(strings.TrimSpace(line))
		if m := gccRegex.FindStringSubmatch(stripped); m != nil {
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
		}
	}
	return errs
}
