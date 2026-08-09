package verify

import "regexp"

// Rule defines a single verification check.
type Rule struct {
	Name       string
	Category   string // security | smell | ai | style
	Severity   string // high | medium | low
	Regex      string
	Suggestion string
}

// compiled caches precompiled regexes for each rule name.
var compiled = map[string]*regexp.Regexp{}

func init() {
	for _, r := range Rules {
		compiled[r.Name] = regexp.MustCompile(r.Regex)
	}
}

// Match reports whether the rule matches the given line.
func (r Rule) Match(line string) bool {
	re, ok := compiled[r.Name]
	if !ok {
		return false
	}
	return re.MatchString(line)
}

// Rules are the verification checks applied to code.
// Modeled on the research: ~89.3% of AI-introduced issues are code smells,
// 21% of agent trajectories produce insecure code, and teams cannot
// attribute AI-authored lines. `dev verify` covers all three.
var Rules = []Rule{
	// --- Security (HIGH) -------------------------------------------------
	{
		Name:       "hardcoded-secret",
		Category:   "security",
		Severity:   "high",
		Regex:      `(?i)(password|passwd|secret|api[_-]?key|access[_-]?token|private[_-]?key)\s*[=:]\s*["'][A-Za-z0-9_\-\.]{8,}["']`,
		Suggestion: "Move secrets to environment variables or a secret manager",
	},
	{
		Name:       "credential-value",
		Category:   "security",
		Severity:   "high",
		Regex:      `(?i)["'](sk_live_|sk_test_|pk_live_|pk_test_|ghp_[A-Za-z0-9]{20,}|github_pat_|AKIA[0-9A-Z]{16}|eyJ[A-Za-z0-9_-]{20,}|xox[baprs]-)[A-Za-z0-9_\-\.]*["']`,
		Suggestion: "Known credential format detected — move to a secret manager",
	},
	{
		Name:       "code-execution",
		Category:   "security",
		Severity:   "high",
		Regex:      `\b(eval|exec|system|child_process\.exec|spawn)\s*\(`,
		Suggestion: "Avoid dynamic code/command execution — validate and sandbox input",
	},
	{
		Name:       "sql-injection",
		Category:   "security",
		Severity:   "high",
		Regex:      `(?i)\b(select|insert|update|delete)\b[^\n;]*\b(from|into|set|values)\b[^\n;]*(query\(|exec\(|\$|" \+|%[sd])`,
		Suggestion: "Use parameterized queries / prepared statements instead of string concatenation",
	},
	{
		Name:       "xss-innerhtml",
		Category:   "security",
		Severity:   "high",
		Regex:      `\.innerHTML\s*=`,
		Suggestion: "Use textContent or a safe DOM API to avoid XSS",
	},
	{
		Name:       "insecure-shell",
		Category:   "security",
		Severity:   "high",
		Regex:      `(os\.system|shell\.run|subprocess\.call|!\[)`,
		Suggestion: "Avoid direct shell calls with unsanitized input",
	},
	// --- AI attribution (marker that code is machine-generated) ----------
	{
		Name:       "ai-generated-marker",
		Category:   "ai",
		Severity:   "low",
		Regex:      `(?i)(generated\s+by|generated\s+with|ai-?generated|auto-?generated|automatically\s+generated|copilot|created\s+with|do\s+not\s+edit\s+manually|this\s+(file|code)\s+was\s+generated)`,
		Suggestion: "AI-authored code detected — verify logic and edge cases manually",
	},
	// --- Code smells (MEDIUM / LOW) --------------------------------------
	{
		Name:       "ts-any",
		Category:   "smell",
		Severity:   "medium",
		Regex:      `:\s*any\b|\bas\s+any\b|<any>`,
		Suggestion: "Replace 'any' with a proper type — AI output frequently widens types",
	},
	{
		Name:       "ts-non-null",
		Category:   "smell",
		Severity:   "medium",
		Regex:      `\w+\s*!\s*\.`,
		Suggestion: "Non-null assertion may mask null/undefined bugs introduced by AI",
	},
	{
		Name:       "empty-catch",
		Category:   "smell",
		Severity:   "medium",
		Regex:      `catch\s*\([^)]*\)\s*\{\s*\}`,
		Suggestion: "Empty catch silently swallows errors — handle or rethrow",
	},
	{
		Name:       "ignored-error",
		Category:   "smell",
		Severity:   "medium",
		Regex:      `\b_\s*=\s*\w+\(|catch\s*\(\s*\)\s*\{\s*\}`,
		Suggestion: "Ignored error return — check it before proceeding",
	},
	{
		Name:       "prod-log",
		Category:   "smell",
		Severity:   "low",
		Regex:      `\b(console\.log|print|println)\s*\(`,
		Suggestion: "Debug logging left in production code",
	},
	{
		Name:       "todo",
		Category:   "smell",
		Severity:   "low",
		Regex:      `\b(TODO|FIXME|XXX)\b`,
		Suggestion: "Unfinished work left in the change",
	},
	{
		Name:       "magic-number",
		Category:   "smell",
		Severity:   "low",
		Regex:      `[^.\w]\d{3,}`,
		Suggestion: "Magic number — extract to a named constant for clarity",
	},
	{
		Name:       "long-line",
		Category:   "style",
		Severity:   "low",
		Regex:      `^.{121,}$`,
		Suggestion: "Line exceeds 120 characters — break it up",
	},
}

// RuleByName returns the rule matching the given name.
func RuleByName(name string) (Rule, bool) {
	for _, r := range Rules {
		if r.Name == name {
			return r, true
		}
	}
	return Rule{}, false
}
