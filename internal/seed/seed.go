package seed

import (
	"os"
	"os/exec"

	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

func NewSeedCommand() *cobra.Command {
	var clear bool

	cmd := &cobra.Command{
		Use:   "seed",
		Short: "Load test data into database",
		Long: `Load test data:
  dev seed       — Insert test data
  dev seed --clear — Remove all test data`,
		Args: cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				logger.Error("No devkit.json found. Run: dev init")
				os.Exit(1)
			}

			if clear {
				clearData(cfg)
			} else {
				seedData(cfg)
			}
		},
	}

	cmd.Flags().BoolVar(&clear, "clear", false, "Remove all test data")
	return cmd
}

func seedData(cfg *config.ProjectConfig) {
	logger.Header("Seeding Database")

	if _, err := os.Stat("src/db/seed.js"); err == nil {
		logger.Step("Running seed script...")
		cmd := exec.Command("node", "src/db/seed.js")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			logger.Error("Seed failed")
			os.Exit(1)
		}
	} else if _, err := os.Stat("seed.js"); err == nil {
		logger.Step("Running seed script...")
		cmd := exec.Command("node", "seed.js")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			logger.Error("Seed failed")
			os.Exit(1)
		}
	} else {
		logger.Warn("No seed file found")
		logger.Print("  Create src/db/seed.js to add test data")
		return
	}

	logger.Success("Database seeded!")
}

func clearData(cfg *config.ProjectConfig) {
	logger.Header("Clearing Test Data")

	if _, err := os.Stat("src/db/clear.js"); err == nil {
		cmd := exec.Command("node", "src/db/clear.js")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			logger.Error("Clear failed")
			os.Exit(1)
		}
	} else {
		logger.Warn("No clear script found")
		logger.Print("  Create src/db/clear.js to clear test data")
		return
	}

	logger.Success("Test data cleared!")
}
