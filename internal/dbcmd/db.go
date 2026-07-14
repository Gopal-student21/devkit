package dbcmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

func NewDBCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "db [shell|url]",
		Short: "Database utilities",
		Long: `Database utilities:
  dev db shell — Open database shell
  dev db url   — Print connection URL`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				logger.Error("No devkit.json found. Run: dev init")
				os.Exit(1)
			}

			action := "url"
			if len(args) > 0 {
				action = args[0]
			}

			switch action {
			case "shell", "connect", "psql":
				openShell(cfg)
			case "url":
				printURL(cfg)
			default:
				logger.Error(fmt.Sprintf("Unknown action: %s", action))
				logger.Print("  Actions: shell, url")
			}
		},
	}
}

func openShell(cfg *config.ProjectConfig) {
	for _, svc := range cfg.Services {
		switch svc.Name {
		case "postgres", "postgresql":
			logger.Step("Connecting to PostgreSQL...")
			cmd := exec.Command("docker", "exec", "-it", "devkit-postgres", "psql", "-U", "postgres", "-d", "devkit")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			return
		case "mysql":
			logger.Step("Connecting to MySQL...")
			cmd := exec.Command("docker", "exec", "-it", "devkit-mysql", "mysql", "-u", "devkit", "-pdevkit123", "devkit")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			return
		case "mongo", "mongodb":
			logger.Step("Connecting to MongoDB...")
			cmd := exec.Command("docker", "exec", "-it", "devkit-mongo", "mongosh", "-u", "devkit", "-p", "devkit123")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			return
		}
	}

	logger.Warn("No database service found")
}

func printURL(cfg *config.ProjectConfig) {
	for _, svc := range cfg.Services {
		switch svc.Name {
		case "postgres", "postgresql":
			logger.Step("postgresql://postgres:devkit123@localhost:5432/devkit")
			return
		case "mysql":
			logger.Step("mysql://devkit:devkit123@localhost:3306/devkit")
			return
		case "mongo", "mongodb":
			logger.Step("mongodb://devkit:devkit123@localhost:27017/devkit?authSource=admin")
			return
		}
	}

	logger.Warn("No database service found")
}
