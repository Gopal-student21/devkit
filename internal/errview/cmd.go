package errview

import (
	"fmt"
	"os"

	"github.com/devkit/devkit/internal/errview/parser"
	"github.com/devkit/devkit/internal/errview/render"
	"github.com/spf13/cobra"
)

func NewErrviewCommand() *cobra.Command {
	var compact bool

	cmd := &cobra.Command{
		Use:   "errview [files...]",
		Short: "Visual error viewer for any language",
		Long: `Parse compiler/linter output and display errors visually
with code context, line numbers, and error pointers.

Examples:
  go build ./... 2>&1 | dev errview
  npx tsc --noEmit 2>&1 | dev errview
  eslint src/ 2>&1 | dev errview
  dev errview errors.txt
  dev errview -c errors.txt`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			lines := parser.ReadInput(args)
			if len(lines) == 0 {
				return fmt.Errorf("no input — pipe errors or pass files:\n  go build ./... 2>&1 | dev errview\n  dev errview errors.txt")
			}

			p := parser.Detect(lines)
			errs := p.Parse(lines)

			if len(errs) == 0 {
				return fmt.Errorf("could not parse any errors from input")
			}

			if compact {
				render.PrintCompact(errs)
			} else {
				render.Display(errs)
			}

			os.Exit(1)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&compact, "compact", "c", false, "One line per error (no code context)")

	return cmd
}
