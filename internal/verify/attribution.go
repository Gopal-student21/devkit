package verify

import (
	"regexp"
	"strings"
)

// Attribution answers "how much of this change looks AI-authored?"
// Multiple weak signals beat a single marker: generated headers, comments
// that restate code, essay-style comment blocks, widened types and
// swallowed errors are all tell-tale LLM output patterns.
type attribution struct {
	AiLines int     `json:"aiLines"`
	Total   int     `json:"totalLines"`
	Pct     float64 `json:"pct"`
	Label   string  `json:"label"`
}

var (
	reComment = regexp.MustCompile(`^\s*(//|/\*|#|--|//\*|\*|<!--)`)
	reCodeTok = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]{3,}`)
)

// scoreAttribution returns a per-file attribution result.
func scoreAttribution(files map[string]fileContent) map[string]attribution {
	out := make(map[string]attribution, len(files))
	for name, fc := range files {
		out[name] = attributionFor(name, fc.contents)
	}
	return out
}

// attributionFor scores one file's added lines.
func attributionFor(name string, lines []string) attribution {
	a := attribution{Total: len(lines)}
	attr := map[int]bool{} // line indexes attributed to AI

	// signal 1: explicit generated markers (comment headers)
	markerRule, hasMarker := RuleByName("ai-generated-marker")

	for i, line := range lines {
		if hasMarker && markerRule.Match(line) {
			attr[i] = true
			a.AiLines++
			continue
		}
		if isEchoComment(lines, i) || isEssayComment(lines, i) {
			attr[i] = true
			a.AiLines++
		}
	}

	// signal 2: error-swallowing and type-widening findings on AI-attributed rules
	for _, ruleName := range []string{"ts-any", "ts-non-null", "empty-catch", "ignored-error"} {
		if rule, ok := RuleByName(ruleName); ok {
			for i, line := range lines {
				if attr[i] {
					continue
				}
				if rule.Match(line) {
					attr[i] = true
					a.AiLines++
				}
			}
		}
	}

	if a.Total > 0 {
		a.Pct = float64(a.AiLines) / float64(a.Total)
	}
	a.Label = attributionLabel(a.Pct, a.AiLines)
	return a
}

// isEchoComment flags comments that restate the very next code token — a
// classic LLM "explain what I just wrote" tell.
func isEchoComment(lines []string, i int) bool {
	line := lines[i]
	if !reComment.MatchString(line) {
		return false
	}
	words := wordsIn(commentText(line))
	if len(words) == 0 {
		return false
	}
	for j := i + 1; j < len(lines) && j <= i+3; j++ {
		next := lines[j]
		if reComment.MatchString(next) {
			continue
		}
		for _, w := range words {
			if strings.Contains(next, w) {
				return true
			}
		}
		return false
	}
	return false
}

// isEssayComment flags 3+ consecutive long prose comments (AI writes
// paragraph-length docblocks humans rarely do).
func isEssayComment(lines []string, i int) bool {
	if !reComment.MatchString(lines[i]) {
		return false
	}
	run := 0
	for j := i; j < len(lines) && j < i+5; j++ {
		if !reComment.MatchString(lines[j]) {
			break
		}
		if len(strings.TrimSpace(lines[j])) > 45 {
			run++
		}
	}
	return run >= 3
}

func commentText(line string) string {
	line = reComment.ReplaceAllString(line, "")
	return line
}

func wordsIn(s string) []string {
	var out []string
	for _, tok := range reCodeTok.FindAllString(s, -1) {
		if len(tok) > 4 && !isStopword(tok) {
			out = append(out, tok)
		}
	}
	return out
}

func isStopword(w string) bool {
	switch strings.ToLower(w) {
	case "the", "this", "that", "with", "from", "into", "check", "make", "return", "will", "used", "here", "below", "above", "each", "when":
		return true
	}
	return false
}

func attributionLabel(pct float64, lines int) string {
	if lines == 0 {
		return "None"
	}
	switch {
	case pct >= 0.35:
		return "High"
	case pct >= 0.15:
		return "Moderate"
	default:
		return "Low"
	}
}

// averageAttribution folds per-file results into a corpus-level summary.
func averageAttribution(perFile map[string]attribution) attribution {
	agg := attribution{}
	for _, a := range perFile {
		agg.AiLines += a.AiLines
		agg.Total += a.Total
	}
	if agg.Total > 0 {
		agg.Pct = float64(agg.AiLines) / float64(agg.Total)
	}
	agg.Label = attributionLabel(agg.Pct, agg.AiLines)
	return agg
}
