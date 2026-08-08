package main

import (
	"fmt"
	"os"

	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/internal/contract"
	"github.com/devkit/devkit/internal/dbcmd"
	"github.com/devkit/devkit/internal/detect"
	"github.com/devkit/devkit/internal/env"
	"github.com/devkit/devkit/internal/errview"
	"github.com/devkit/devkit/internal/initcmd"
	"github.com/devkit/devkit/internal/logs"
	"github.com/devkit/devkit/internal/migrate"
	"github.com/devkit/devkit/internal/newcmd"
	"github.com/devkit/devkit/internal/qa"
	"github.com/devkit/devkit/internal/review"
	"github.com/devkit/devkit/internal/seed"
	"github.com/devkit/devkit/internal/servcmd"
	"github.com/devkit/devkit/internal/start"
	"github.com/devkit/devkit/internal/status"
	"github.com/devkit/devkit/internal/stop"
	"github.com/devkit/devkit/internal/testcmd"
	"github.com/spf13/cobra"
)

var version = "0.6.1"

func main() {
	rootCmd := &cobra.Command{
		Use:   "dev",
		Short: "DevKit — Universal CLI Dashboard",
		Long:  `Turn any CLI tool into a visual dashboard.`,
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("devkit v%s\n", version)
		},
	})

	rootCmd.AddCommand(servcmd.NewServeCommand())
	rootCmd.AddCommand(initcmd.NewInitCommand())
	rootCmd.AddCommand(newcmd.NewNewCommand())
	rootCmd.AddCommand(contract.NewContractCommand())
	rootCmd.AddCommand(qa.NewQACommand())
	rootCmd.AddCommand(review.NewReviewCommand())
	rootCmd.AddCommand(start.NewUpCommand())
	rootCmd.AddCommand(stop.NewStopCommand())
	rootCmd.AddCommand(testcmd.NewTestCommand())
	rootCmd.AddCommand(status.NewStatusCommand())
	rootCmd.AddCommand(env.NewEnvCommand())
	rootCmd.AddCommand(logs.NewLogsCommand())
	rootCmd.AddCommand(migrate.NewMigrateCommand())
	rootCmd.AddCommand(seed.NewSeedCommand())
	rootCmd.AddCommand(dbcmd.NewDBCommand())
	rootCmd.AddCommand(errview.NewErrviewCommand())
	config.Register(rootCmd)
	detect.Register(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
