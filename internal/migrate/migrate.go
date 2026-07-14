package migrate

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

func NewMigrateCommand() *cobra.Command {
	var action string

	cmd := &cobra.Command{
		Use:   "migrate [action]",
		Short: "Run database migrations",
		Long: `Run database migrations:
  dev migrate        — Run all pending migrations
  dev migrate up     — Run all pending migrations
  dev migrate down   — Rollback last migration
  dev migrate status — Show migration status`,
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				action = args[0]
			} else {
				action = "up"
			}

			cfg, err := config.Load()
			if err != nil {
				logger.Error("No devkit.json found. Run: dev init")
				os.Exit(1)
			}

			switch action {
			case "up", "run", "apply":
				runMigrate(cfg)
			case "down", "rollback":
				rollbackMigrate(cfg)
			case "status":
				showStatus(cfg)
			default:
				logger.Error(fmt.Sprintf("Unknown action: %s", action))
				logger.Print("  Actions: up, down, status")
			}
		},
	}

	return cmd
}

func runMigrate(cfg *config.ProjectConfig) {
	logger.Header("Running Migrations")

	dbType := getDBType(cfg)
	if dbType == "" {
		logger.Warn("No database found in services")
		return
	}

	// Check if migration file exists
	if _, err := os.Stat("src/db/migrate.js"); err == nil {
		logger.Step("Running Node.js migrations...")
		cmd := exec.Command("node", "src/db/migrate.js")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			logger.Error("Migration failed")
			os.Exit(1)
		}
	} else if _, err := os.Stat("migrate.js"); err == nil {
		logger.Step("Running migrations...")
		cmd := exec.Command("node", "migrate.js")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			logger.Error("Migration failed")
			os.Exit(1)
		}
	} else if _, err := os.Stat("app/main.py"); err == nil {
		logger.Step("Running Python migrations...")
		logger.Info("Run: flask db upgrade")
		return
	} else if _, err := os.Stat("cmd/server/main.go"); err == nil {
		logger.Step("Running Go migrations...")
		logger.Info("Run: go run cmd/migrate/main.go")
		return
	} else {
		logger.Warn("No migration files found")
		logger.Print("  Create src/db/migrate.js for auto-migration")
		return
	}

	logger.Success("Migrations completed!")
}

func rollbackMigrate(cfg *config.ProjectConfig) {
	logger.Header("Rolling Back Migrations")

	if _, err := os.Stat("src/db/rollback.js"); err == nil {
		cmd := exec.Command("node", "src/db/rollback.js")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			logger.Error("Rollback failed")
			os.Exit(1)
		}
	} else {
		logger.Warn("No rollback script found")
		logger.Print("  Create src/db/rollback.js for rollback support")
		return
	}

	logger.Success("Rollback completed!")
}

func showStatus(cfg *config.ProjectConfig) {
	logger.Header("Migration Status")

	dbType := getDBType(cfg)
	if dbType == "" {
		logger.Warn("No database configured")
		return
	}

	logger.Step(fmt.Sprintf("Database: %s", dbType))

	// Check for migration files
	files := []string{
		"src/db/migrate.js",
		"src/db/seed.js",
		"src/db/rollback.js",
		"migrate.js",
		"seed.js",
	}

	found := false
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			logger.Step(fmt.Sprintf("✓ %s", file))
			found = true
		}
	}

	if !found {
		logger.Info("No migration files found")
	}
}

func getDBType(cfg *config.ProjectConfig) string {
	for _, svc := range cfg.Services {
		switch svc.Name {
		case "postgres", "postgresql":
			return "postgresql"
		case "mysql":
			return "mysql"
		case "mongo", "mongodb":
			return "mongodb"
		}
	}
	return ""
}
