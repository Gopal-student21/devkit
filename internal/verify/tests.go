package verify

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// TestResult captures an automated test run merged into the verdict.
type TestResult struct {
	Ran      bool   `json:"ran"`
	Command  string `json:"command,omitempty"`
	Passed   bool   `json:"passed"`
	Duration string `json:"duration,omitempty"`
	Output   string `json:"output,omitempty"`
}

// testCommand maps a detected stack type to the test command to run.
func testCommand(stackType string) (string, []string) {
	switch stackType {
	case "go":
		return "go", []string{"test", "./..."}
	case "node":
		return "npm", []string{"test", "--", "--watch=false"}
	case "python":
		return "pytest", []string{"-q"}
	case "rust":
		return "cargo", []string{"test", "--quiet"}
	case "java":
		return "mvn", []string{"test", "-q"}
	default:
		return "", nil
	}
}

// runTests executes the project test suite with a timeout and summarizes it.
func runTests(stackType string, timeout time.Duration) TestResult {
	bin, args := testCommand(stackType)
	if bin == "" {
		return TestResult{Ran: false, Output: "no known test framework for this stack"}
	}

	start := time.Now()
	cmd := exec.Command(bin, args...)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out

	done := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return TestResult{Ran: false, Output: fmt.Sprintf("could not start %s: %v", bin, err)}
	}
	go func() { done <- cmd.Wait() }()

	var err error
	select {
	case err = <-done:
	case <-time.After(timeout):
		cmd.Process.Kill()
		<-done
		err = fmt.Errorf("test run timed out after %s", timeout)
	}

	output := out.String()
	const maxOut = 4000
	if len(output) > maxOut {
		output = "…" + output[len(output)-maxOut:]
	}

	return TestResult{
		Ran:      true,
		Command:  fmt.Sprintf("%s %s", bin, strings.Join(args, " ")),
		Passed:   err == nil,
		Duration: time.Since(start).Round(100 * time.Millisecond).String(),
		Output:   strings.TrimSpace(output),
	}
}
