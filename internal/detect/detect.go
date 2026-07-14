package detect

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
)

type StackInfo struct {
	Name        string
	Type        string
	ConfigFiles []string
	HasTests    bool
	HasDocker   bool
	HasDB       bool
	DBType      string
}

func DetectStack() *StackInfo {
	stack := &StackInfo{}
	entries, _ := os.ReadDir(".")

	files := make(map[string]bool)
	for _, e := range entries {
		files[e.Name()] = true
	}

	switch {
	case files["package.json"]:
		stack.Name = "Node.js"
		stack.Type = "node"
		stack.ConfigFiles = []string{"package.json", "package-lock.json"}
		stack.HasDocker = files["Dockerfile"] || files["docker-compose.yml"]
		stack.HasDB = files["prisma"] || files["drizzle"] || files[".env"]
		if files["prisma"] || dirExists("prisma") {
			stack.DBType = "postgresql"
		}

	case files["go.mod"]:
		stack.Name = "Go"
		stack.Type = "go"
		stack.ConfigFiles = []string{"go.mod", "go.sum"}

	case files["requirements.txt"] || files["pyproject.toml"] || files["setup.py"]:
		stack.Name = "Python"
		stack.Type = "python"
		stack.ConfigFiles = []string{"requirements.txt"}
		stack.HasDB = files["manage.py"] || files["settings.py"]
		stack.DBType = "sqlite"

	case files["Cargo.toml"]:
		stack.Name = "Rust"
		stack.Type = "rust"
		stack.ConfigFiles = []string{"Cargo.toml"}

	case files["pom.xml"] || files["build.gradle"]:
		stack.Name = "Java"
		stack.Type = "java"

	case files["Gemfile"]:
		stack.Name = "Ruby"
		stack.Type = "ruby"

	default:
		stack.Name = "Unknown"
		stack.Type = "unknown"
	}

	stack.HasTests = hasTestFiles()
	return stack
}

func hasTestFiles() bool {
	found := false
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := filepath.Base(path)
		if len(name) > 5 {
			switch {
			case name == "jest.config.js" || name == "jest.config.ts":
				found = true
			case name == "pytest.ini" || name == "conftest.py":
				found = true
			case name == "vitest.config.ts" || name == "vitest.config.js":
				found = true
			}
		}
		return nil
	})
	return found
}

func dirExists(name string) bool {
	info, err := os.Stat(name)
	return err == nil && info.IsDir()
}

func Register(root *cobra.Command) {
	root.AddCommand(&cobra.Command{
		Use:   "detect",
		Short: "Detect the project stack and services",
		Run: func(cmd *cobra.Command, args []string) {
			stack := DetectStack()
			logger.Header("Detected Stack")
			logger.Step(fmt.Sprintf("Language: %s", stack.Name))
			logger.Step(fmt.Sprintf("Type:     %s", stack.Type))
			if stack.HasTests {
				logger.Step("Tests:    detected")
			}
			if stack.HasDB {
				logger.Step(fmt.Sprintf("Database: %s", stack.DBType))
			}
			if stack.HasDocker {
				logger.Step("Docker:   detected")
			}
		},
	})
}
