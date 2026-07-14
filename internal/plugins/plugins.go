package plugins

import (
	"github.com/devkit/devkit/internal/config"
)

func GetBuiltinPlugins() map[string]*config.Plugin {
	return map[string]*config.Plugin{
		"docker": DockerPlugin(),
		"postgres": PostgresPlugin(),
		"redis": RedisPlugin(),
		"git": GitPlugin(),
		"node": NodePlugin(),
	}
}

func DockerPlugin() *config.Plugin {
	return &config.Plugin{
		Name:        "docker",
		Label:       "Docker",
		Description: "Container management",
		Icon:        "🐳",
		Commands: []config.CommandDef{
			{Name: "ps", Label: "List Containers", Action: "exec", Args: []string{"docker", "ps", "-a", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}"}},
			{Name: "images", Label: "List Images", Action: "exec", Args: []string{"docker", "images", "--format", "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"}},
			{Name: "logs", Label: "Container Logs", Action: "exec", Args: []string{"docker", "logs", "--tail", "100"}},
			{Name: "stats", Label: "Resource Usage", Action: "exec", Args: []string{"docker", "stats", "--no-stream", "--format", "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"}},
			{Name: "start", Label: "Start Container", Action: "exec", Args: []string{"docker", "start"}, Confirm: "Start this container?", Params: []config.ParamDef{{Name: "container", Label: "Container Name", Type: "text"}}},
			{Name: "stop", Label: "Stop Container", Action: "exec", Args: []string{"docker", "stop"}, Confirm: "Stop this container?", Params: []config.ParamDef{{Name: "container", Label: "Container Name", Type: "text"}}},
			{Name: "rm", Label: "Remove Container", Action: "exec", Args: []string{"docker", "rm", "-f"}, Confirm: "Permanently remove this container?", Params: []config.ParamDef{{Name: "container", Label: "Container Name", Type: "text"}}},
			{Name: "pull", Label: "Pull Image", Action: "exec", Args: []string{"docker", "pull"}, Params: []config.ParamDef{{Name: "image", Label: "Image Name", Type: "text"}}},
		},
	}
}

func PostgresPlugin() *config.Plugin {
	return &config.Plugin{
		Name:        "postgres",
		Label:       "PostgreSQL",
		Description: "Database management",
		Icon:        "🐘",
		Commands: []config.CommandDef{
			{Name: "psql", Label: "Open Shell", Action: "terminal", Args: []string{"docker", "exec", "-it", "devkit-postgres", "psql", "-U", "postgres", "-d", "devkit"}},
			{Name: "tables", Label: "List Tables", Action: "exec", Args: []string{"docker", "exec", "devkit-postgres", "psql", "-U", "postgres", "-d", "devkit", "-c", "\\dt"}},
			{Name: "status", Label: "Server Status", Action: "exec", Args: []string{"docker", "exec", "devkit-postgres", "psql", "-U", "postgres", "-d", "devkit", "-c", "SELECT version()"}},
			{Name: "connections", Label: "Active Connections", Action: "exec", Args: []string{"docker", "exec", "devkit-postgres", "psql", "-U", "postgres", "-d", "devkit", "-c", "SELECT count(*) FROM pg_stat_activity"}},
			{Name: "size", Label: "Database Size", Action: "exec", Args: []string{"docker", "exec", "devkit-postgres", "psql", "-U", "postgres", "-d", "devkit", "-c", "SELECT pg_size_pretty(pg_database_size('devkit'))"}},
			{Name: "query", Label: "Run Query", Action: "exec", Args: []string{"docker", "exec", "devkit-postgres", "psql", "-U", "postgres", "-d", "devkit", "-c"}, Params: []config.ParamDef{{Name: "query", Label: "SQL Query", Type: "textarea"}}},
		},
	}
}

func RedisPlugin() *config.Plugin {
	return &config.Plugin{
		Name:        "redis",
		Label:       "Redis",
		Description: "Cache management",
		Icon:        "🔴",
		Commands: []config.CommandDef{
			{Name: "cli", Label: "Open CLI", Action: "terminal", Args: []string{"docker", "exec", "-it", "devkit-redis", "redis-cli"}},
			{Name: "ping", Label: "Ping Server", Action: "exec", Args: []string{"docker", "exec", "devkit-redis", "redis-cli", "ping"}},
			{Name: "info", Label: "Server Info", Action: "exec", Args: []string{"docker", "exec", "devkit-redis", "redis-cli", "info", "server"}},
			{Name: "keys", Label: "List Keys", Action: "exec", Args: []string{"docker", "exec", "devkit-redis", "redis-cli", "keys", "*"}},
			{Name: "memory", Label: "Memory Usage", Action: "exec", Args: []string{"docker", "exec", "devkit-redis", "redis-cli", "info", "memory"}},
			{Name: "dbsize", Label: "Key Count", Action: "exec", Args: []string{"docker", "exec", "devkit-redis", "redis-cli", "dbsize"}},
		},
	}
}

func GitPlugin() *config.Plugin {
	return &config.Plugin{
		Name:        "git",
		Label:       "Git",
		Description: "Version control",
		Icon:        "📦",
		Commands: []config.CommandDef{
			{Name: "status", Label: "Status", Action: "exec", Args: []string{"git", "status"}},
			{Name: "log", Label: "Commit History", Action: "exec", Args: []string{"git", "log", "--oneline", "-20"}},
			{Name: "diff", Label: "Uncommitted Changes", Action: "exec", Args: []string{"git", "diff"}},
			{Name: "branches", Label: "Branches", Action: "exec", Args: []string{"git", "branch", "-a"}},
			{Name: "add", Label: "Stage All", Action: "exec", Args: []string{"git", "add", "."}},
			{Name: "commit", Label: "Commit", Action: "exec", Args: []string{"git", "commit", "-m"}, Params: []config.ParamDef{{Name: "message", Label: "Commit Message", Type: "text"}}},
			{Name: "push", Label: "Push to Remote", Action: "exec", Args: []string{"git", "push"}},
			{Name: "pull", Label: "Pull from Remote", Action: "exec", Args: []string{"git", "pull"}},
		},
	}
}

func NodePlugin() *config.Plugin {
	return &config.Plugin{
		Name:        "node",
		Label:       "Node.js",
		Description: "JavaScript runtime",
		Icon:        "🟢",
		Commands: []config.CommandDef{
			{Name: "version", Label: "Node Version", Action: "exec", Args: []string{"node", "--version"}},
			{Name: "npm-version", Label: "npm Version", Action: "exec", Args: []string{"npm", "--version"}},
			{Name: "install", Label: "Install Dependencies", Action: "exec", Args: []string{"npm", "install"}},
			{Name: "update", Label: "Update Dependencies", Action: "exec", Args: []string{"npm", "update"}},
			{Name: "audit", Label: "Security Audit", Action: "exec", Args: []string{"npm", "audit"}},
			{Name: "outdated", Label: "Outdated Packages", Action: "exec", Args: []string{"npm", "outdated"}},
			{Name: "dev", Label: "Run Dev Server", Action: "terminal", Args: []string{"npm", "run", "dev"}},
			{Name: "test", Label: "Run Tests", Action: "exec", Args: []string{"npm", "test"}},
		},
	}
}

func GetPlugin(name string) *config.Plugin {
	builtin := GetBuiltinPlugins()
	if p, ok := builtin[name]; ok {
		return p
	}
	return nil
}

func ListPlugins() []*config.Plugin {
	builtin := GetBuiltinPlugins()
	var plugins []*config.Plugin
	for _, p := range builtin {
		plugins = append(plugins, p)
	}
	return plugins
}
