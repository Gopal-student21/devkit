package logger

import (
	"fmt"
	"os"
)

var (
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Red    = "\033[31m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
	Reset  = "\033[0m"
)

func init() {
	if os.Getenv("NO_COLOR") != "" {
		Green, Yellow, Red, Cyan, Bold, Reset = "", "", "", "", "", ""
	}
}

func Success(msg string)   { fmt.Printf("%s✓%s %s\n", Green, Reset, msg) }
func Info(msg string)      { fmt.Printf("%s→%s %s\n", Cyan, Reset, msg) }
func Warn(msg string)      { fmt.Printf("%s⚠%s %s\n", Yellow, Reset, msg) }
func Error(msg string)     { fmt.Printf("%s✗%s %s\n", Red, Reset, msg) }
func Step(msg string)      { fmt.Printf("%s▸%s %s\n", Cyan, Reset, msg) }
func Header(msg string)    { fmt.Printf("\n%s%s%s\n\n", Bold, msg, Reset) }
func Print(msg string)     { fmt.Println(msg) }
