package logs

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

func NewLogsCommand() *cobra.Command {
	var follow bool
	var tail string

	cmd := &cobra.Command{
		Use:   "logs [service]",
		Short: "Show logs from DevKit services",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				showAllLogs(follow, tail)
			} else {
				showServiceLogs(args[0], follow, tail)
			}
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().StringVarP(&tail, "tail", "t", "50", "Number of lines to show")
	return cmd
}

func showAllLogs(follow bool, tail string) {
	containers := getDevKitContainers()
	if len(containers) == 0 {
		logger.Warn("No running DevKit services found")
		return
	}

	logger.Header("DevKit Service Logs")

	for _, name := range containers {
		logger.Step(fmt.Sprintf("--- %s ---", name))
		args := []string{"logs", "--tail", tail}
		if follow {
			args = append(args, "-f")
		}
		args = append(args, name)

		cmd := exec.Command("docker", args...)
		output, err := cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			fmt.Println(string(output))
		}
	}
}

func showServiceLogs(service string, follow bool, tail string) {
	containerName := fmt.Sprintf("devkit-%s", service)

	check := exec.Command("docker", "inspect", containerName)
	if check.Run() != nil {
		logger.Error(fmt.Sprintf("Service '%s' is not running", service))
		return
	}

	args := []string{"logs", "--tail", tail}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, containerName)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func getDevKitContainers() []string {
	cmd := exec.Command("docker", "ps", "--filter", "name=devkit-", "--format", "{{.Names}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	var containers []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			containers = append(containers, line)
		}
	}
	return containers
}
