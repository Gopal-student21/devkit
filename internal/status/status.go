package status

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

func NewStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check status of all development services",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			logger.Header("DevKit Status")
			logger.Step(fmt.Sprintf("Project: %s", cfg.Name))

			if len(cfg.Services) == 0 {
				logger.Info("No services configured")
				return
			}

			for _, svc := range cfg.Services {
				status := getContainerStatus(svc)
				icon := "○"
				if status == "running" {
					icon = "●"
				}
				logger.Step(fmt.Sprintf("%s %s (%s)", icon, svc.Name, status))
			}
		},
	}
}

func getContainerStatus(svc config.ServiceConfig) string {
	if svc.Type == "runtime" {
		return "available"
	}

	containerName := fmt.Sprintf("devkit-%s", svc.Name)
	cmd := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "stopped"
	}
	return strings.TrimSpace(string(output))
}
