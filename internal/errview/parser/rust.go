package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// Rust: error[E0425]: cannot find value `foo` in this scope
// --> src/main.rs:42:5
var rustLocRegex = regexp.MustCompile(`^-->\s*(.+?):(\d+):(\d+)$`)
var rustErrRegex = regexp.MustCompile(`^(error|warning)\[?(E?\d+)?\]?:\s*(.+)$`)

type RustParser struct{}

func (p *RustParser) Parse(lines []string) []Error {
	var errs []Error
	for i := 0; i < len(lines); i++ {
		stripped := stripAnsi(strings.TrimSpace(lines[i]))
		if m := rustErrRegex.FindStringSubmatch(stripped); m != nil {
			err := Error{
				Level:   m[1],
				Message: stripped,
				Raw:     stripped,
			}
			// Look ahead for location
			if i+1 < len(lines) {
				next := strings.TrimSpace(stripAnsi(lines[i+1]))
				if loc := rustLocRegex.FindStringSubmatch(next); loc != nil {
					err.File = loc[1]
					err.Line, _ = strconv.Atoi(loc[2])
					err.Column, _ = strconv.Atoi(loc[3])
				}
			}
			if err.File != "" {
				errs = append(errs, err)
			}
		}
	}
	return errs
}
