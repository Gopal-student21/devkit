package stop

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

func NewStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop [service...]",
		Short: "Stop DevKit services (all or specific)",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				stopAll()
			} else {
				for _, arg := range args {
					stopService(arg)
				}
			}
		},
	}
}

func stopAll() {
	containers := getDevKitContainers()
	if len(containers) == 0 {
		logger.Info("No running DevKit services found")
		return
	}

	logger.Header("Stopping All Services")

	for _, name := range containers {
		stopContainer(name)
	}

	logger.Success(fmt.Sprintf("Stopped %d service(s)", len(containers)))
}

func stopService(name string) {
	containerName := fmt.Sprintf("devkit-%s", name)
	stopContainer(containerName)
}

func stopContainer(containerName string) {
	stop := exec.Command("docker", "stop", containerName)
	if err := stop.Run(); err != nil {
		logger.Warn(fmt.Sprintf("Could not stop %s", containerName))
		return
	}

	rm := exec.Command("docker", "rm", containerName)
	rm.Run()

	logger.Success(fmt.Sprintf("Stopped and removed %s", containerName))
}

func getDevKitContainers() []string {
	cmd := exec.Command("docker", "ps", "-a", "--filter", "name=devkit-", "--format", "{{.Names}}")
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
