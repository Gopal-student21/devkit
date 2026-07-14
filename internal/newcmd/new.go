package newcmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/devkit/devkit/internal/config"
	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

func NewNewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "new <name> [template]",
		Short: "Scaffold a new project with best practices",
		Long: `Create a new project with pre-configured:
  - Database (PostgreSQL/MySQL/MongoDB)
  - Environment variables
  - Docker setup
  - Git repository
  - Basic API structure`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			projectName := args[0]
			template := "node"
			if len(args) > 1 {
				template = args[1]
			}

			if _, err := os.Stat(projectName); err == nil {
				logger.Error(fmt.Sprintf("Directory '%s' already exists", projectName))
				os.Exit(1)
			}

			logger.Header(fmt.Sprintf("Creating Project: %s", projectName))

			// Create project directory
			if err := os.MkdirAll(projectName, 0755); err != nil {
				logger.Error(fmt.Sprintf("Failed to create directory: %s", err))
				os.Exit(1)
			}

			// Change to project directory
			originalDir, _ := os.Getwd()
			if err := os.Chdir(projectName); err != nil {
				logger.Error(fmt.Sprintf("Failed to enter directory: %s", err))
				os.Exit(1)
			}
			defer os.Chdir(originalDir)

			// Scaffold based on template
			switch template {
			case "node", "nodejs", "nextjs":
				scaffoldNode(projectName)
			case "python", "django", "flask":
				scaffoldPython(projectName)
			case "go", "golang":
				scaffoldGo(projectName)
			default:
				scaffoldGeneric(projectName)
			}

			// Initialize git
			initGit()

			// Create devkit.json
			createConfig(projectName, template)

			// Create .env
			createEnvFile(template)

			// Create .gitignore
			createGitignore(template)

			logger.Header("Project Ready!")
			logger.Print(fmt.Sprintf("  cd %s", projectName))
			logger.Print("  dev up          — Start database")
			logger.Print("  dev env         — Show credentials")
			logger.Print("  dev migrate     — Run migrations")
			logger.Print("  dev test        — Run tests")
			logger.Print("")
		},
	}
}

func scaffoldNode(name string) {
	logger.Step("Scaffolding Node.js project...")

	// package.json
	packageJSON := `{
  "name": "` + name + `",
  "version": "1.0.0",
  "private": true,
  "scripts": {
    "dev": "node --watch src/index.js",
    "start": "node src/index.js",
    "test": "node --test src/**/*.test.js",
    "migrate": "node src/db/migrate.js",
    "seed": "node src/db/seed.js"
  },
  "dependencies": {
    "express": "^4.18.0",
    "pg": "^8.11.0",
    "dotenv": "^16.3.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0"
  }
}`
	os.WriteFile("package.json", []byte(packageJSON), 0644)

	// Create src directory structure
	dirs := []string{"src", "src/db", "src/routes", "src/middleware"}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	// Main entry point
	indexJS := `require('dotenv').config();
const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

// Health check
app.get('/health', (req, res) => {
  res.json({ status: 'ok', timestamp: new Date().toISOString() });
});

// Routes
app.get('/', (req, res) => {
  res.json({ message: 'Welcome to ' + process.env.npm_package_name });
});

app.listen(PORT, () => {
  console.log('Server running on port ' + PORT);
});
`
	os.WriteFile("src/index.js", []byte(indexJS), 0644)

	// Database connection
	dbJS := `const { Pool } = require('pg');

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
});

module.exports = pool;
`
	os.WriteFile("src/db/connection.js", []byte(dbJS), 0644)

	// Migration script
	migrateJS := `const pool = require('./connection');

async function migrate() {
  const client = await pool.connect();
  try {
    await client.query('CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, email VARCHAR(255) UNIQUE NOT NULL, created_at TIMESTAMP DEFAULT NOW())');
    console.log('Migration completed successfully');
  } finally {
    client.release();
    process.exit(0);
  }
}

migrate().catch(err => {
  console.error('Migration failed:', err);
  process.exit(1);
});
`
	os.WriteFile("src/db/migrate.js", []byte(migrateJS), 0644)

	// Seed script
	seedJS := `const pool = require('./connection');

async function seed() {
  const client = await pool.connect();
  try {
    await client.query('INSERT INTO users (email) VALUES ($1) ON CONFLICT DO NOTHING', ['test@example.com']);
    console.log('Seed data inserted');
  } finally {
    client.release();
    process.exit(0);
  }
}

seed().catch(err => {
  console.error('Seed failed:', err);
  process.exit(1);
});
`
	os.WriteFile("src/db/seed.js", []byte(seedJS), 0644)

	logger.Success("Node.js project scaffolded")
}

func scaffoldPython(name string) {
	logger.Step("Scaffolding Python project...")

	dirs := []string{"app", "app/routes", "app/models"}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	requirements := `flask>=3.0.0
psycopg2-binary>=2.9.0
python-dotenv>=1.0.0
`
	os.WriteFile("requirements.txt", []byte(requirements), 0644)

	mainPy := `import os
from flask import Flask, jsonify
from dotenv import load_dotenv

load_dotenv()
app = Flask(__name__)

@app.route('/health')
def health():
    return jsonify(status='ok')

@app.route('/')
def index():
    return jsonify(message='Welcome to ` + name + `')

if __name__ == '__main__':
    app.run(debug=True, port=int(os.getenv('PORT', 3000)))
`
	os.WriteFile("app/main.py", []byte(mainPy), 0644)

	logger.Success("Python project scaffolded")
}

func scaffoldGo(name string) {
	logger.Step("Scaffolding Go project...")

	dirs := []string{"cmd", "internal", "pkg"}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	mainGo := `package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"message": "Welcome to ` + name + `"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Server running on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}
`
	os.WriteFile("cmd/server/main.go", []byte(mainGo), 0644)

	logger.Success("Go project scaffolded")
}

func scaffoldGeneric(name string) {
	logger.Step("Creating generic project...")

	dirs := []string{"src", "docs"}
	for _, dir := range dirs {
		os.MkdirAll(dir, 0755)
	}

	readme := `# ` + name + `

## Getting Started

1. Start services: dev up postgres
2. Generate .env: dev env
3. Run migrations: dev migrate
4. Start development: dev run
`
	os.WriteFile("README.md", []byte(readme), 0644)

	logger.Success("Generic project scaffolded")
}

func initGit() {
	logger.Step("Initializing git repository...")
	cmd := exec.Command("git", "init")
	cmd.Run()

	cmd = exec.Command("git", "add", ".")
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Run()

	logger.Success("Git repository initialized")
}

func createConfig(name, template string) {
	dbType := "postgres"
	switch template {
	case "python":
		dbType = "postgres"
	case "go":
		dbType = "postgres"
	}

	cfg := &config.ProjectConfig{
		Name: name,
		Type: template,
		Services: []config.ServiceConfig{
			{Name: dbType, Type: "database", Port: 5432},
			{Name: "redis", Type: "cache", Port: 6379},
		},
		Scripts: map[string]string{
			"dev":     "npm run dev",
			"test":    "npm test",
			"build":   "npm run build",
			"migrate": "npm run migrate",
			"seed":    "npm run seed",
		},
	}

	config.Save(cfg)
	logger.Success("Created devkit.json")
}

func createEnvFile(template string) {
	env := `# Generated by DevKit
DATABASE_URL=postgresql://postgres:devkit123@localhost:5432/devkit
REDIS_URL=redis://localhost:6379
PORT=3000
NODE_ENV=development
`
	os.WriteFile(".env", []byte(env), 0644)
	logger.Success("Created .env")
}

func createGitignore(template string) {
	gitignore := `# Dependencies
node_modules/
venv/
__pycache__/

# Environment
.env
.env.local

# DevKit
.devkit/

# Build
dist/
build/
*.exe

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
`
	os.WriteFile(".gitignore", []byte(gitignore), 0644)
	logger.Success("Created .gitignore")
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func createDir(path string) {
	if !dirExists(path) {
		os.MkdirAll(path, 0755)
	}
}

func writeTemplate(name, content string) {
	createDir(filepath.Dir(name))
	os.WriteFile(name, []byte(content), 0644)
}

func trimSpace(s string) string {
	return strings.TrimSpace(s)
}
