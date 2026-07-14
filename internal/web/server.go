package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/internal/plugins"
)

func StartServer(port int) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", serveDashboard)
	mux.HandleFunc("/api/plugins", handleListPlugins)
	mux.HandleFunc("/api/services", handleServices)
	mux.HandleFunc("/api/exec/", handleExec)
	mux.HandleFunc("/api/start/", handleStartService)
	mux.HandleFunc("/api/stop/", handleStopService)
	mux.HandleFunc("/api/stop-all", handleStopAll)
	mux.HandleFunc("/api/logs/", handleLogs)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("\n  DevKit Dashboard: http://localhost%s\n\n", addr)
	return http.ListenAndServe(addr, mux)
}

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, dashboardHTML)
}

func handleListPlugins(w http.ResponseWriter, r *http.Request) {
	cfg, _ := config.Load()
	var enabled []*config.Plugin
	all := plugins.GetBuiltinPlugins()

	for _, name := range cfg.Plugins {
		if p, ok := all[name]; ok {
			enabled = append(enabled, p)
		}
	}

	if len(enabled) == 0 {
		for _, p := range all {
			enabled = append(enabled, p)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"plugins": enabled})
}

func handleExec(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/exec/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		http.Error(w, "Invalid path", 400)
		return
	}
	pluginName := parts[0]
	cmdName := parts[1]

	var body map[string]string
	json.NewDecoder(r.Body).Decode(&body)

	plugin := plugins.GetPlugin(pluginName)
	if plugin == nil {
		http.Error(w, "Plugin not found", 404)
		return
	}

	var cmdDef *config.CommandDef
	for _, c := range plugin.Commands {
		if c.Name == cmdName {
			cmdDef = &c
			break
		}
	}
	if cmdDef == nil {
		http.Error(w, "Command not found", 404)
		return
	}

	args := append([]string{}, cmdDef.Args...)

	if len(cmdDef.Params) > 0 {
		for _, param := range cmdDef.Params {
			if val, ok := body[param.Name]; ok {
				args = append(args, val)
			}
		}
	} else if body["args"] != "" {
		for _, a := range strings.Fields(body["args"]) {
			args = append(args, a)
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		output = append(output, []byte("\nError: "+err.Error())...)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"output": string(output)})
}

func handleStartService(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/start/")
	if _, ok := plugins.GetBuiltinPlugins()[name]; !ok {
		http.Error(w, "Unknown service", 404)
		return
	}

	containerName := "devkit-" + name
	check := exec.Command("docker", "inspect", containerName)
	if check.Run() == nil {
		exec.Command("docker", "start", containerName).Run()
	} else {
		args := []string{"run", "-d", "--name", containerName, "--restart", "unless-stopped"}
		switch name {
		case "postgres":
			args = append(args, "-p", "5432:5432", "-e", "POSTGRES_USER=postgres", "-e", "POSTGRES_PASSWORD=devkit123", "-e", "POSTGRES_DB=devkit", "postgres:16-alpine")
		case "redis":
			args = append(args, "-p", "6379:6379", "redis:7-alpine")
		case "mysql":
			args = append(args, "-p", "3306:3306", "-e", "MYSQL_ROOT_PASSWORD=devkit123", "-e", "MYSQL_USER=devkit", "-e", "MYSQL_PASSWORD=devkit123", "-e", "MYSQL_DATABASE=devkit", "mysql:8-alpine")
		case "mongo":
			args = append(args, "-p", "27017:27017", "-e", "MONGO_INITDB_ROOT_USERNAME=devkit", "-e", "MONGO_INITDB_ROOT_PASSWORD=devkit123", "mongo:7")
		default:
			http.Error(w, "No default image for "+name, 400)
			return
		}
		cmd := exec.Command("docker", args...)
		cmd.Run()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "started", "name": name})
}

func handleStopService(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/stop/")
	container := "devkit-" + name
	exec.Command("docker", "stop", container).Run()
	exec.Command("docker", "rm", container).Run()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func handleStopAll(w http.ResponseWriter, r *http.Request) {
	containers := getDevKitContainers()
	for _, c := range containers {
		exec.Command("docker", "stop", c).Run()
		exec.Command("docker", "rm", c).Run()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stopped"})
}

func handleLogs(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/api/logs/")
	container := "devkit-" + name
	cmd := exec.Command("docker", "logs", "--tail", "100", container)
	output, _ := cmd.CombinedOutput()
	lines := strings.Split(string(output), "\n")
	var filtered []string
	for _, l := range lines {
		if l = strings.TrimSpace(l); l != "" {
			filtered = append(filtered, l)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"lines": filtered})
}

func handleServices(w http.ResponseWriter, r *http.Request) {
	containers := getDevKitContainers()
	type svc struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		Port   int    `json:"port"`
	}
	var services []svc
	for _, c := range containers {
		name := strings.TrimPrefix(c, "devkit-")
		port := getContainerPort(c)
		services = append(services, svc{Name: name, Status: "running", Port: port})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"services": services})
}

func getDevKitContainers() []string {
	cmd := exec.Command("docker", "ps", "--filter", "name=devkit-", "--format", "{{.Names}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}
	var containers []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			containers = append(containers, line)
		}
	}
	return containers
}

func getContainerPort(container string) int {
	cmd := exec.Command("docker", "port", container)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) >= 2 {
			port := 0
			fmt.Sscanf(parts[len(parts)-1], "%d", &port)
			return port
		}
	}
	return 0
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
