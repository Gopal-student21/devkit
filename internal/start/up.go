package start

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

var serviceCatalog = map[string]ServiceDef{
	"postgres": {
		Name:    "postgres",
		Image:   "postgres:16-alpine",
		Port:    5432,
		EnvVars: []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"},
	},
	"postgresql": {
		Name:    "postgres",
		Image:   "postgres:16-alpine",
		Port:    5432,
		EnvVars: []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB"},
	},
	"redis": {
		Name:    "redis",
		Image:   "redis:7-alpine",
		Port:    6379,
		EnvVars: []string{},
	},
	"mysql": {
		Name:    "mysql",
		Image:   "mysql:8-alpine",
		Port:    3306,
		EnvVars: []string{"MYSQL_ROOT_PASSWORD", "MYSQL_DATABASE", "MYSQL_USER", "MYSQL_PASSWORD"},
	},
	"mongo": {
		Name:    "mongo",
		Image:   "mongo:7",
		Port:    27017,
		EnvVars: []string{"MONGO_INITDB_ROOT_USERNAME", "MONGO_INITDB_ROOT_PASSWORD"},
	},
	"mongodb": {
		Name:    "mongo",
		Image:   "mongo:7",
		Port:    27017,
		EnvVars: []string{"MONGO_INITDB_ROOT_USERNAME", "MONGO_INITDB_ROOT_PASSWORD"},
	},
	"minio": {
		Name:    "minio",
		Image:   "minio/minio:latest",
		Port:    9000,
		EnvVars: []string{"MINIO_ROOT_USER", "MINIO_ROOT_PASSWORD"},
	},
}

type ServiceDef struct {
	Name    string
	Image   string
	Port    int
	EnvVars []string
}

func NewUpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up [service...]",
		Short: "Start development services",
		Long:  "Start specific services (dev up postgres redis) or all from devkit.json (dev up)",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				upFromConfig()
				return
			}

			started := 0
			for _, arg := range args {
				svcName := strings.ToLower(arg)
				def, ok := serviceCatalog[svcName]
				if !ok {
					logger.Error(fmt.Sprintf("Unknown service: %s", svcName))
					printAvailableServices()
					os.Exit(1)
				}
				if err := startService(def); err != nil {
					logger.Error(fmt.Sprintf("Failed to start %s: %s", def.Name, err))
				} else {
					started++
				}
			}
			logger.Success(fmt.Sprintf("Started %d service(s)", started))
		},
	}
	return cmd
}

func upFromConfig() {
	cfg, err := config.Load()
	if err != nil {
		logger.Info("No devkit.json found.")
		logger.Print("  Usage: dev up postgres redis")
		logger.Print("  Or run: dev init")
		os.Exit(1)
	}

	if len(cfg.Services) == 0 {
		logger.Info("No services in devkit.json.")
		logger.Print("  Usage: dev up postgres redis")
		return
	}

	logger.Header(fmt.Sprintf("Starting %s", cfg.Name))

	started := 0
	for _, svc := range cfg.Services {
		def, ok := serviceCatalog[svc.Name]
		if !ok {
			logger.Warn(fmt.Sprintf("Skipping unknown service: %s", svc.Name))
			continue
		}
		if svc.Port > 0 {
			def.Port = svc.Port
		}
		if err := startService(def); err != nil {
			logger.Error(fmt.Sprintf("Failed to start %s: %s", def.Name, err))
		} else {
			started++
		}
	}

	if started > 0 {
		logger.Success(fmt.Sprintf("All %d service(s) running!", started))
	}
}

func startService(def ServiceDef) error {
	containerName := fmt.Sprintf("devkit-%s", def.Name)

	// Check if already running
	running := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
	if running.Run() == nil {
		logger.Success(fmt.Sprintf("%s already running (port %d)", def.Name, def.Port))
		return nil
	}

	// Check if exists but stopped
	exists := exec.Command("docker", "inspect", containerName)
	if exists.Run() == nil {
		startCmd := exec.Command("docker", "start", containerName)
		if startCmd.Run() == nil {
			logger.Success(fmt.Sprintf("%s restarted (port %d)", def.Name, def.Port))
			return nil
		}
	}

	// Pull image
	logger.Step(fmt.Sprintf("Pulling %s...", def.Image))
	pull := exec.Command("docker", "pull", def.Image)
	pull.Stdout = os.Stdout
	pull.Stderr = os.Stderr
	if err := pull.Run(); err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	// Build docker run command
	args := []string{
		"run", "-d",
		"--name", containerName,
		"--restart", "unless-stopped",
		"-p", fmt.Sprintf("%d:%d", def.Port, def.Port),
	}

	// Add environment variables
	envMap := generateEnv(def)
	for key, val := range envMap {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, val))
	}

	// Special handling for MinIO
	if def.Name == "minio" {
		args = append(args, def.Image, "server", "/data", "--console-address", ":9001")
	} else {
		args = append(args, def.Image)
	}

	// Create and start container
	logger.Step(fmt.Sprintf("Starting %s on port %d...", def.Name, def.Port))
	create := exec.Command("docker", args...)
	output, err := create.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run failed: %s — %s", err, string(output))
	}

	logger.Success(fmt.Sprintf("%s running on port %d", def.Name, def.Port))
	printConnectionInfo(def, envMap)

	return nil
}

func generateEnv(def ServiceDef) map[string]string {
	env := map[string]string{}

	switch def.Name {
	case "postgres":
		env["POSTGRES_USER"] = getEnvOrDefault("POSTGRES_USER", "postgres")
		env["POSTGRES_PASSWORD"] = getEnvOrDefault("POSTGRES_PASSWORD", "devkit123")
		env["POSTGRES_DB"] = getEnvOrDefault("POSTGRES_DB", "devkit")

	case "redis":
		// Redis needs no env vars for basic setup

	case "mysql":
		pass := getEnvOrDefault("MYSQL_ROOT_PASSWORD", "devkit123")
		env["MYSQL_ROOT_PASSWORD"] = pass
		env["MYSQL_USER"] = getEnvOrDefault("MYSQL_USER", "devkit")
		env["MYSQL_PASSWORD"] = pass
		env["MYSQL_DATABASE"] = getEnvOrDefault("MYSQL_DATABASE", "devkit")

	case "mongo":
		env["MONGO_INITDB_ROOT_USERNAME"] = getEnvOrDefault("MONGO_INITDB_ROOT_USERNAME", "devkit")
		env["MONGO_INITDB_ROOT_PASSWORD"] = getEnvOrDefault("MONGO_INITDB_ROOT_PASSWORD", "devkit123")

	case "minio":
		env["MINIO_ROOT_USER"] = getEnvOrDefault("MINIO_ROOT_USER", "devkit")
		env["MINIO_ROOT_PASSWORD"] = getEnvOrDefault("MINIO_ROOT_PASSWORD", "devkit123")
	}

	return env
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func printConnectionInfo(def ServiceDef, env map[string]string) {
	logger.Print("")
	switch def.Name {
	case "postgres":
		user := env["POSTGRES_USER"]
		pass := env["POSTGRES_PASSWORD"]
		db := env["POSTGRES_DB"]
		logger.Step(fmt.Sprintf("postgresql://%s:%s@localhost:%d/%s", user, pass, def.Port, db))
	case "redis":
		logger.Step(fmt.Sprintf("redis://localhost:%d", def.Port))
	case "mysql":
		user := env["MYSQL_USER"]
		pass := env["MYSQL_PASSWORD"]
		db := env["MYSQL_DATABASE"]
		logger.Step(fmt.Sprintf("mysql://%s:%s@localhost:%d/%s", user, pass, def.Port, db))
	case "mongo":
		user := env["MONGO_INITDB_ROOT_USERNAME"]
		pass := env["MONGO_INITDB_ROOT_PASSWORD"]
		logger.Step(fmt.Sprintf("mongodb://%s:%s@localhost:%d", user, pass, def.Port))
	case "minio":
		logger.Step(fmt.Sprintf("API: http://localhost:%d", def.Port))
		logger.Step(fmt.Sprintf("Console: http://localhost:9001"))
	}
	logger.Print("")
}

func printAvailableServices() {
	logger.Print("Available services:")
	for name := range serviceCatalog {
		logger.Step(name)
	}
}

func parsePort(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
