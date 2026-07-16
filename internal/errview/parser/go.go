package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// Go compiler error format:
// file.go:42:5: undefined: foo
// file.go:42:5: cannot use ... (type mismatch)
var goRegex = regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(error|warning|note|panic):\s*(.+)$`)
var goShortRegex = regexp.MustCompile(`^(.+?):(\d+):(\d+):\s*(.+)$`)

type GoParser struct{}

func (p *GoParser) Parse(lines []string) []Error {
	var errs []Error
	for _, line := range lines {
		stripped := stripAnsi(strings.TrimSpace(line))
		if stripped == "" {
			continue
		}

		var m []string
		level := "error"

		if groups := goRegex.FindStringSubmatch(stripped); groups != nil {
			m = groups
			level = groups[4]
		} else if groups := goShortRegex.FindStringSubmatch(stripped); groups != nil {
			m = groups
			msg := strings.ToLower(groups[4])
			if strings.Contains(msg, "warning") {
				level = "warning"
			}
		}

		if m != nil {
			lineNum, _ := strconv.Atoi(m[2])
			colNum, _ := strconv.Atoi(m[3])
			errs = append(errs, Error{
				File:    m[1],
				Line:    lineNum,
				Column:  colNum,
				Message: m[len(m)-1],
				Level:   level,
				Raw:     stripped,
			})
		}
	}
	return errs
}
