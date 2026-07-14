package testcmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

func NewTestCommand() *cobra.Command {
	var watch bool
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run the project test suite",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			testScript, ok := cfg.Scripts["test"]
			if !ok {
				logger.Error("No test script defined in devkit.json")
				os.Exit(1)
			}

			logger.Header("Running Tests")

			var execCmd *exec.Cmd
			if watch {
				logger.Info("Running in watch mode...")
				execCmd = exec.Command("sh", "-c", testScript+" --watch")
			} else {
				execCmd = exec.Command("sh", "-c", testScript)
			}

			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			execCmd.Stdin = os.Stdin

			if err := execCmd.Run(); err != nil {
				logger.Error("Tests failed")
				os.Exit(1)
			}

			logger.Success("All tests passed!")
		},
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "Run tests in watch mode")
	return cmd
}
