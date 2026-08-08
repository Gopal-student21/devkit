package contract

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/devkit/devkit/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type APISpec struct {
	OpenAPI string                 `json:"openapi"`
	Info    map[string]interface{} `json:"info"`
	Paths   map[string]interface{} `json:"paths"`
}

func NewContractCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract",
		Short: "API contract testing tools",
		Long: `Manage API contracts between frontend and backend:
  dev contract init     — Create API contract file
  dev contract test     — Test API matches contract
  dev contract mock     — Start mock server
  dev contract types    — Generate TypeScript types
  dev contract validate — Validate contract file`,
	}

	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newTestCommand())
	cmd.AddCommand(newMockCommand())
	cmd.AddCommand(newTypesCommand())
	cmd.AddCommand(newValidateCommand())

	return cmd
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create API contract file (OpenAPI 3.0)",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat("api.yaml"); err == nil {
				logger.Warn("api.yaml already exists")
				return
			}

			spec := `openapi: "3.0.0"
info:
  title: "My API"
  version: "1.0.0"
  description: "API contract for frontend-backend alignment"

paths:
  /api/users:
    get:
      summary: "Get all users"
      responses:
        "200":
          description: "Success"
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/User"
    post:
      summary: "Create user"
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateUserRequest"
      responses:
        "201":
          description: "Created"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"

  /api/users/{id}:
    get:
      summary: "Get user by ID"
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: "Success"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
        "404":
          description: "Not found"

components:
  schemas:
    User:
      type: object
      required:
        - id
        - email
        - name
      properties:
        id:
          type: string
        email:
          type: string
          format: email
        name:
          type: string
        createdAt:
          type: string
          format: date-time

    CreateUserRequest:
      type: object
      required:
        - email
        - name
      properties:
        email:
          type: string
          format: email
        name:
          type: string
        password:
          type: string
          minLength: 8
`

			if err := os.WriteFile("api.yaml", []byte(spec), 0644); err != nil {
				logger.Error(fmt.Sprintf("Failed to create api.yaml: %s", err))
				return
			}

			logger.Success("Created api.yaml")
			logger.Print("")
			logger.Print("Edit api.yaml to define your API contract")
			logger.Print("Then run: dev contract test")
		},
	}
}

func newTestCommand() *cobra.Command {
	var baseURL string
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Test API matches contract",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat("api.yaml"); err != nil {
				logger.Error("No api.yaml found. Run: dev contract init")
				return
			}

			if baseURL == "" {
				baseURL = "http://localhost:3000"
			}

			logger.Header("Testing API Contract")
			logger.Step(fmt.Sprintf("Base URL: %s", baseURL))
			logger.Step("Contract: api.yaml")
			logger.Print("")

			// Parse the contract and test endpoints
			testAPIContract(baseURL)
		},
	}
	cmd.Flags().StringVar(&baseURL, "url", "http://localhost:3000", "API base URL")
	return cmd
}

func testAPIContract(baseURL string) {
	data, err := os.ReadFile("api.yaml")
	if err != nil {
		logger.Error("Failed to read api.yaml")
		return
	}

	content := string(data)
	passed := 0
	failed := 0

	// Simple parsing - find endpoints and test them
	lines := strings.Split(content, "\n")
	currentPath := ""
	currentMethod := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "/") && strings.HasSuffix(line, ":") {
			currentPath = strings.TrimSuffix(line, ":")
		}

		if line == "get:" || line == "post:" || line == "put:" || line == "delete:" {
			currentMethod = strings.TrimSuffix(line, ":")
		}

		if currentPath != "" && currentMethod != "" {
			url := baseURL + currentPath
			logger.Step(fmt.Sprintf("Testing %s %s", strings.ToUpper(currentMethod), currentPath))

			cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "-X", strings.ToUpper(currentMethod), url)
			output, err := cmd.CombinedOutput()

			statusCode := strings.TrimSpace(string(output))
			if err != nil || (statusCode != "200" && statusCode != "201" && statusCode != "404") {
				logger.Error(fmt.Sprintf("  Failed: HTTP %s", statusCode))
				failed++
			} else {
				logger.Success(fmt.Sprintf("  Passed: HTTP %s", statusCode))
				passed++
			}

			currentPath = ""
			currentMethod = ""
		}
	}

	logger.Print("")
	logger.Header("Results")
	logger.Success(fmt.Sprintf("%d passed", passed))
	if failed > 0 {
		logger.Error(fmt.Sprintf("%d failed", failed))
	}
}

func newMockCommand() *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "mock",
		Short: "Start mock server from contract",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat("api.yaml"); err != nil {
				logger.Error("No api.yaml found. Run: dev contract init")
				return
			}

			if port == 0 {
				port = 4000
			}

			logger.Header("Starting Mock Server")
			logger.Step(fmt.Sprintf("Port: %d", port))
			logger.Step("Contract: api.yaml")
			logger.Print("")
			logger.Print("Mock responses based on your API contract")
			logger.Print("Press Ctrl+C to stop")
			logger.Print("")

			startMockServer(port)
		},
	}
	cmd.Flags().IntVarP(&port, "port", "p", 4000, "Mock server port")
	return cmd
}

func startMockServer(port int) {
	// Generate mock server code
	mockCode := fmt.Sprintf(`const http = require('http');
const fs = require('fs');
const yaml = require('js-yaml');

const spec = yaml.load(fs.readFileSync('api.yaml', 'utf8'));
const PORT = %d;

function resolveRef(schema) {
  if (!schema || typeof schema.$ref !== 'string') return schema;
  const parts = schema.$ref.split('/');
  const name = parts[parts.length - 1];
  return spec.components?.schemas?.[name] || schema;
}

function generateMockValue(prop, seen) {
  if (!prop) return null;
  prop = resolveRef(prop);
  seen = seen || [];
  if (prop.enum && prop.enum.length) return prop.enum[0];
  if (prop.type === 'string') {
    if (prop.format === 'email') return 'test@example.com';
    if (prop.format === 'date-time') return new Date().toISOString();
    if (prop.format === 'url' || prop.format === 'uri') return 'https://example.com';
    return 'mock-value';
  }
  if (prop.type === 'number' || prop.type === 'integer') return 42;
  if (prop.type === 'boolean') return true;
  if (prop.type === 'array') {
    const item = generateMockValue(prop.items, seen);
    return [item].filter((v) => v !== null);
  }
  if (prop.type === 'object') return generateMockResponse(prop, seen);
  return null;
}

function generateMockResponse(schema, seen) {
  if (!schema) return {};
  schema = resolveRef(schema);
  const name = schema.$ref ? schema.$ref.split('/').pop() : schema.title;
  if (seen && name && seen.includes(name)) return {};
  if (schema.type === 'object' && schema.properties) {
    const obj = {};
    const nextSeen = (seen || []).concat(name ? [name] : []);
    for (const [key, prop] of Object.entries(schema.properties)) {
      obj[key] = generateMockValue(prop, nextSeen);
    }
    return obj;
  }
  if (schema.type === 'array') {
    return [generateMockResponse(schema.items, seen)].filter((v) => v && Object.keys(v).length > 0);
  }
  return generateMockValue(schema, seen) || {};
}

const server = http.createServer((req, res) => {
  res.setHeader('Content-Type', 'application/json');
  res.setHeader('Access-Control-Allow-Origin', '*');
  
  const path = spec.paths[req.url.split('?')[0]];
  if (!path) {
    res.writeHead(404);
    res.end(JSON.stringify({ error: 'Not found' }));
    return;
  }

  const method = req.method.toLowerCase();
  const operation = path[method];
  if (!operation) {
    res.writeHead(405);
    res.end(JSON.stringify({ error: 'Method not allowed' }));
    return;
  }

  const response = operation.responses['200'] || operation.responses['201'];
  if (response?.content?.['application/json']?.schema) {
    const mockData = generateMockResponse(response.content['application/json'].schema);
    res.writeHead(200);
    res.end(JSON.stringify(mockData));
  } else {
    res.writeHead(200);
    res.end(JSON.stringify({ message: 'Mock response' }));
  }
});

server.listen(PORT, () => {
  console.log('Mock server running on http://localhost:' + PORT);
});
`, port)

	os.WriteFile("mock-server.js", []byte(mockCode), 0644)

	// Install js-yaml if needed
	exec.Command("npm", "install", "js-yaml", "--save-dev").Run()

	// Start the mock server
	cmd := exec.Command("node", "mock-server.js")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func newTypesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "types",
		Short: "Generate TypeScript types from contract",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat("api.yaml"); err != nil {
				logger.Error("No api.yaml found. Run: dev contract init")
				return
			}

			logger.Header("Generating TypeScript Types")

			data, err := os.ReadFile("api.yaml")
			if err != nil {
				logger.Error("Failed to read api.yaml")
				return
			}

			var spec struct {
				Components struct {
					Schemas map[string]map[string]interface{} `yaml:"schemas"`
				} `yaml:"components"`
			}
			if err := yaml.Unmarshal(data, &spec); err != nil {
				logger.Error("Failed to parse api.yaml: " + err.Error())
				return
			}

			if len(spec.Components.Schemas) == 0 {
				logger.Warn("No schemas found under components.schemas")
			}

			requiredFields := map[string]map[string]bool{}
			for name, schema := range spec.Components.Schemas {
				req := map[string]bool{}
				if required, ok := schema["required"].([]interface{}); ok {
					for _, r := range required {
						if s, ok := r.(string); ok {
							req[s] = true
						}
					}
				}
				requiredFields[name] = req
			}

			names := sortedKeys(spec.Components.Schemas)

			var b strings.Builder
			b.WriteString("// Auto-generated from api.yaml\n// Do not edit manually\n\n")

			for _, name := range names {
				schema := spec.Components.Schemas[name]
				props, _ := schema["properties"].(map[string]interface{})

				b.WriteString(fmt.Sprintf("export interface %s {\n", name))
				if len(props) == 0 {
					b.WriteString("  [key: string]: unknown;\n")
				}
				for _, key := range sortedKeys(props) {
					prop, _ := props[key].(map[string]interface{})
					opt := ""
					if !requiredFields[name][key] {
						opt = "?"
					}
					b.WriteString(fmt.Sprintf("  %s%s: %s;\n", key, opt, contractToTS(prop)))
				}
				b.WriteString("}\n\n")
			}

			os.WriteFile("types.ts", []byte(b.String()), 0644)
			logger.Success("Generated types.ts")
		},
	}
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func contractToTS(prop map[string]interface{}) string {
	if prop == nil {
		return "unknown"
	}

	if ref, ok := prop["$ref"].(string); ok {
		return tsRefName(ref)
	}

	if refs, ok := prop["anyOf"].([]interface{}); ok && len(refs) > 0 {
		parts := make([]string, 0, len(refs))
		for _, r := range refs {
			if m, ok := r.(map[string]interface{}); ok {
				parts = append(parts, contractToTS(m))
			}
		}
		return strings.Join(parts, " | ")
	}

	t, _ := prop["type"].(string)
	switch t {
	case "string":
		if enums, ok := prop["enum"].([]interface{}); ok && len(enums) > 0 {
			parts := make([]string, 0, len(enums))
			for _, e := range enums {
				parts = append(parts, fmt.Sprintf("%q", fmt.Sprint(e)))
			}
			return strings.Join(parts, " | ")
		}
		return "string"
	case "number", "integer":
		return "number"
	case "boolean":
		return "boolean"
	case "array":
		items, _ := prop["items"].(map[string]interface{})
		inner := contractToTS(items)
		if strings.HasPrefix(inner, "{ ") || strings.Contains(inner, "|") {
			return "Array<" + inner + ">"
		}
		return inner + "[]"
	case "object":
		if props, ok := prop["properties"].(map[string]interface{}); ok && len(props) > 0 {
			required := map[string]bool{}
			if req, ok := prop["required"].([]interface{}); ok {
				for _, r := range req {
					if s, ok := r.(string); ok {
						required[s] = true
					}
				}
			}
			var inner strings.Builder
			inner.WriteString("{ ")
			for _, key := range sortedKeys(props) {
				p, _ := props[key].(map[string]interface{})
				opt := ""
				if !required[key] {
					opt = "?"
				}
				inner.WriteString(fmt.Sprintf("%s%s: %s; ", key, opt, contractToTS(p)))
			}
			inner.WriteString("}")
			return inner.String()
		}
		return "Record<string, unknown>"
	default:
		return "unknown"
	}
}

func tsRefName(ref string) string {
	if idx := strings.LastIndex(ref, "/"); idx >= 0 {
		return ref[idx+1:]
	}
	return ref
}

func newValidateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate API contract file",
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := os.Stat("api.yaml"); err != nil {
				logger.Error("No api.yaml found. Run: dev contract init")
				return
			}

			logger.Header("Validating api.yaml")

			data, err := os.ReadFile("api.yaml")
			if err != nil {
				logger.Error("Failed to read api.yaml")
				return
			}

			content := string(data)
			errors := 0

			if !strings.Contains(content, "openapi:") {
				logger.Error("Missing 'openapi' version")
				errors++
			}
			if !strings.Contains(content, "info:") {
				logger.Error("Missing 'info' section")
				errors++
			}
			if !strings.Contains(content, "paths:") {
				logger.Error("Missing 'paths' section")
				errors++
			}

			if errors == 0 {
				logger.Success("api.yaml is valid")
			} else {
				logger.Error(fmt.Sprintf("Found %d error(s)", errors))
			}
		},
	}
}
