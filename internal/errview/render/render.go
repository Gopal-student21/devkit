package render

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/devkit/devkit/internal/errview/parser"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorWhite   = "\033[37m"
	colorDim     = "\033[2m"
	colorBold    = "\033[1m"
	colorBgRed   = "\033[41m"
	colorBgYellow = "\033[43m"
)

var levelColors = map[string]string{
	"error":   colorRed,
	"warning": colorYellow,
	"note":    colorCyan,
	"info":    colorBlue,
}

var levelIcons = map[string]string{
	"error":   "✗",
	"warning": "⚠",
	"note":    "ℹ",
	"info":    "ℹ",
}

const contextLines = 3

// Display renders all errors visually
func Display(errs []parser.Error) {
	if len(errs) == 0 {
		fmt.Printf("%s%sNo errors found.%s\n", colorGreen, colorBold, colorReset)
		return
	}

	// Print summary header
	errors, warnings := countByLevel(errs)
	fmt.Println()
	fmt.Printf("%s%s━━━ Error Summary ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%s\n", colorBold, colorWhite, colorReset)
	if errors > 0 {
		fmt.Printf("  %s%s%d error(s)%s", colorRed, colorBold, errors, colorReset)
	}
	if warnings > 0 {
		if errors > 0 {
			fmt.Print("  ")
		}
		fmt.Printf("%s%s%d warning(s)%s", colorYellow, colorBold, warnings, colorReset)
	}

	// Count unique files
	files := make(map[string]bool)
	for _, e := range errs {
		files[e.File] = true
	}
	fmt.Printf("  %s%d file(s) affected%s\n", colorDim, len(files), colorReset)
	fmt.Println()

	// Group errors by file
	byFile := groupByFile(errs)

	for file, fileErrs := range byFile {
		fmt.Printf("%s%s▸ %s%s\n", colorBold, colorCyan, file, colorReset)

		for _, err := range fileErrs {
			renderError(err)
		}
		fmt.Println()
	}
}

func renderError(err parser.Error) {
	levelColor := levelColors[err.Level]
	icon := levelIcons[err.Level]

	// Error header line
	fmt.Printf("  %s%s %s line %d:%d%s\n",
		levelColor, icon, strings.ToUpper(err.Level),
		err.Line, err.Column, colorReset)

	// Error message
	fmt.Printf("  %s%s%s\n", colorBold, err.Message, colorReset)

	// Read source code context
	lines := readFileContext(err.File, err.Line, contextLines)
	if len(lines) == 0 {
		fmt.Printf("  %s(could not read source)%s\n", colorDim, colorReset)
		fmt.Println()
		return
	}

	// Find the gutter width for line numbers
	maxLineNum := err.Line + contextLines
	if err.Line-contextLines > 1 {
		maxLineNum = err.Line + contextLines
	}
	gutterWidth := len(fmt.Sprintf("%d", maxLineNum))
	if gutterWidth < 3 {
		gutterWidth = 3
	}

	// Separator
	fmt.Printf("  %s%s┌─%s\n", colorDim, colorWhite, colorReset)

	// Render each line
	for _, cl := range lines {
		isErrorLine := cl.num == err.Line
		lineNumStr := fmt.Sprintf("%*d", gutterWidth, cl.num)

		if isErrorLine {
			// Error line: colored line number + highlighted code
			fmt.Printf("  %s%s%s │%s %s%s%s%s\n",
				colorBold, colorWhite, lineNumStr, colorReset,
				levelColor, cl.text, colorReset, colorReset)

			// Draw the pointer ^^^^
			if err.Column > 0 {
				pointer := buildPointer(err.Column, gutterWidth, err.Message, levelColor)
				fmt.Print(pointer)
			}
		} else {
			// Context line: dimmed
			fmt.Printf("  %s%s%s │%s %s%s%s\n",
				colorDim, colorWhite, lineNumStr, colorReset,
				colorDim, cl.text, colorReset)
		}
	}

	fmt.Printf("  %s%s└─%s\n", colorDim, colorWhite, colorReset)
}

func buildPointer(col, gutterWidth int, msg string, color string) string {
	// Spaces before the pointer: gutter + " │ " + (col-1 spaces)
	spaces := strings.Repeat(" ", gutterWidth) + " │ "
	pointerIndent := strings.Repeat(" ", col-1)
	pointer := fmt.Sprintf("%s%s%s%s%s^ %s%s%s\n",
		colorDim, spaces, colorReset,
		color, pointerIndent,
		color, msg, colorReset)
	return pointer
}

type codeLine struct {
	num  int
	text string
}

func readFileContext(filename string, targetLine, context int) []codeLine {
	f, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []codeLine
	scanner := bufio.NewScanner(f)
	lineNum := 0
	startLine := targetLine - context
	if startLine < 1 {
		startLine = 1
	}
	endLine := targetLine + context

	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && lineNum <= endLine {
			text := scanner.Text()
			// Trim trailing whitespace but preserve indentation
			text = strings.TrimRight(text, " \t")
			if text == "" {
				text = " "
			}
			lines = append(lines, codeLine{num: lineNum, text: text})
		}
		if lineNum > endLine {
			break
		}
	}
	return lines
}

func groupByFile(errs []parser.Error) map[string][]parser.Error {
	byFile := make(map[string][]parser.Error)
	for _, e := range errs {
		byFile[e.File] = append(byFile[e.File], e)
	}
	return byFile
}

func countByLevel(errs []parser.Error) (errors, warnings int) {
	for _, e := range errs {
		switch e.Level {
		case "error", "fatal error":
			errors++
		case "warning":
			warnings++
		}
	}
	return
}

// PrintCompact renders a single-line-per-error compact view
func PrintCompact(errs []parser.Error) {
	for _, err := range errs {
		levelColor := levelColors[err.Level]
		icon := levelIcons[err.Level]
		fmt.Printf("%s%s%s %s%s:%d:%d%s %s%s%s\n",
			levelColor, icon, colorReset,
			colorCyan, err.File, err.Line, err.Column, colorReset,
			colorBold, err.Message, colorReset)
	}
}
