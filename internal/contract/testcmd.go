package contract

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/devkit/devkit/pkg/logger"
)

// runContractTest hits every endpoint in api.yaml and validates the JSON
// response against the declared response schema. Exits 1 on any failure.
func runContractTest(baseURL string, strict bool) {
	spec, err := LoadSpec("api.yaml")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	logger.Header("Contract Test")
	logger.Step(fmt.Sprintf("Base URL: %s", baseURL))
	logger.Step(fmt.Sprintf("Contract: api.yaml (%d endpoints)", len(spec.Endpoints())))
	if strict {
		logger.Step("Mode: strict (extra fields fail)")
	}
	logger.Print("")

	client := &http.Client{Timeout: 10 * time.Second}
	passed, failed := 0, 0
	totalIssues := 0

	for _, ep := range spec.Endpoints() {
		issues := testEndpoint(client, spec, ep, baseURL, strict)
		totalIssues += len(issues)

		if len(issues) == 0 {
			logger.Success(fmt.Sprintf("%-6s %s", strings.ToUpper(ep.Method), ep.Path))
			passed++
			continue
		}
		logger.Error(fmt.Sprintf("%-6s %s", strings.ToUpper(ep.Method), ep.Path))
		failed++
		for _, iss := range issues {
			prefix := logger.Yellow
			if iss.Kind == "extra" && !iss.Strict {
				prefix = logger.Reset
			}
			fmt.Printf("%s   [%s] %s: %s%s\n", prefix, iss.Kind, iss.Path, iss.Msg, logger.Reset)
		}
	}

	logger.Print("")
	logger.Header("Results")
	logger.Success(fmt.Sprintf("%d endpoint(s) passed", passed))
	if failed > 0 {
		logger.Error(fmt.Sprintf("%d endpoint(s) failed", failed))
		logger.Print(fmt.Sprintf("%d contract issue(s) found", totalIssues))
		os.Exit(1)
	}
}

var pathParam = regexp.MustCompile(`\{([^}]+)\}`)

// testEndpoint builds a request for ep, runs it, and validates the response.
func testEndpoint(client *http.Client, spec *Spec, ep Endpoint, baseURL string, strict bool) []Issue {
	url := baseURL + ep.Path
	body := strings.NewReader("")

	// substitute path params with values derived from the spec
	for _, m := range pathParam.FindAllStringSubmatch(ep.Path, -1) {
		paramName := m[1]
		val := "test-" + paramName
		for _, p := range ep.Operation.Parameters {
			if p.Name == paramName && p.Schema != nil && len(p.Schema.Enum) > 0 {
				val = fmt.Sprint(p.Schema.Enum[0])
			}
		}
		url = strings.ReplaceAll(url, "{"+paramName+"}", val)
	}

	// generate a request body when the spec declares one
	if rb := ep.Operation.RequestBody; rb != nil {
		if mt, ok := rb.Content["application/json"]; ok && mt.Schema != nil {
			if obj, ok := oneRequestBody(spec, mt.Schema); ok {
				if data, err := json.Marshal(obj); err == nil {
					body = strings.NewReader(string(data))
				}
			}
		}
	}

	req, err := http.NewRequest(strings.ToUpper(ep.Method), url, body)
	if err != nil {
		return []Issue{{Kind: "schema", Path: "$", Msg: err.Error()}}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return []Issue{{Kind: "status", Path: "$", Msg: "request failed: " + err.Error()}}
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	status := fmt.Sprintf("%d", resp.StatusCode)

	// 404 etc. that the spec declares are acceptable
	if specAccepts(ep, status) {
		if len(raw) == 0 {
			return nil
		}
	}

	schema, ok := ep.ResponseSchema(status)
	if !ok {
		// no schema declared for this status — accept any 2xx
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		return []Issue{{Kind: "status", Path: "$", Msg: fmt.Sprintf("unexpected HTTP %s (no schema declared)", status)}}
	}

	issues, _ := ValidateBytes(spec, raw, schema, strict)
	return issues
}

// specAccepts reports whether the spec explicitly declares this status.
func specAccepts(ep Endpoint, status string) bool {
	if _, ok := ep.Operation.Responses[status]; ok {
		return true
	}
	return false
}

// oneRequestBody builds a single object matching the request schema.
func oneRequestBody(spec *Spec, schema *Schema) (map[string]interface{}, bool) {
	seeder := NewSeeder(spec, SeedOptions{Count: 1, Seed: 7})
	resolved := schema
	if schema.Ref != "" {
		if r, ok := spec.Resolve(schema.Ref); ok {
			resolved = r
		}
	}
	obj := map[string]interface{}{}
	for _, k := range sortedKeys(resolved.Properties) {
		obj[k] = seeder.generateValue(resolved.Properties[k], k, "request", 0, 0)
	}
	return obj, true
}
