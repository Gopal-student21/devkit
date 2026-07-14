package initcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/internal/detect"
	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

func NewInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new DevKit project",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			stack := detect.DetectStack()

			name := "."
			if len(args) > 0 {
				name = args[0]
			} else {
				dir, _ := os.Getwd()
				name = filepath.Base(dir)
			}

			if _, err := os.Stat(config.ConfigFile); err == nil {
				logger.Warn("devkit.json already exists in this directory")
				return
			}

			logger.Header("Initializing DevKit Project")

			logger.Step(fmt.Sprintf("Project: %s", name))

			if stack.Type == "unknown" {
				logger.Info("No language detected — creating generic project config")
			} else {
				logger.Info(fmt.Sprintf("Detected %s project", stack.Name))
			}

			cfg := &config.ProjectConfig{
				Name:     name,
				Type:     stack.Type,
				Services: detectServices(stack),
				Scripts:  defaultScripts(stack),
			}

			if err := config.Save(cfg); err != nil {
				logger.Error(fmt.Sprintf("Failed to save config: %s", err))
				os.Exit(1)
			}
			logger.Success("Created devkit.json")

			if err := createGitignore(); err == nil {
				logger.Success("Updated .gitignore")
			}

			logger.Header("Next Steps")
			logger.Print("  dev up       — Start development services")
			logger.Print("  dev test     — Run your test suite")
			logger.Print("  dev status   — Check service status")
			logger.Print("")
		},
	}
}

func detectServices(stack *detect.StackInfo) []config.ServiceConfig {
	var services []config.ServiceConfig

	switch stack.Type {
	case "node":
		services = append(services, config.ServiceConfig{
			Name: "node",
			Type: "runtime",
		})
	case "python":
		services = append(services, config.ServiceConfig{
			Name: "python",
			Type: "runtime",
		})
	case "go":
		services = append(services, config.ServiceConfig{
			Name: "go",
			Type: "runtime",
		})
	}

	return services
}

func defaultScripts(stack *detect.StackInfo) map[string]string {
	scripts := make(map[string]string)

	switch stack.Type {
	case "node":
		scripts["dev"] = "npm run dev"
		scripts["test"] = "npm test"
		scripts["build"] = "npm run build"
		scripts["lint"] = "npx next lint"
	case "python":
		scripts["dev"] = "python manage.py runserver"
		scripts["test"] = "pytest"
	case "go":
		scripts["dev"] = "go run ."
		scripts["test"] = "go test ./..."
		scripts["build"] = "go build -o bin/app ."
	}

	return scripts
}

func createGitignore() error {
	gitignore := ".gitignore"
	entries := []string{
		"# DevKit",
		".devkit/",
		"devkit.json",
		"",
	}

	existing, _ := os.ReadFile(gitignore)
	content := string(existing)

	for _, entry := range entries {
		if !strings.Contains(content, entry) {
			content += entry + "\n"
		}
	}

	return os.WriteFile(gitignore, []byte(content), 0644)
}
