package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/process"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProcessAdapter_Integration verifies that the Engine can execute a process tool
// defined in tools.yaml via the configured Process Runner.
func TestProcessAdapter_Integration(t *testing.T) {
	// 1. Setup Temporary Workspace
	tmpDir := t.TempDir()

	// 2. Creates tools.yaml
	// We use 'echo' as a cross-platform command (available in Windows Powershell too? Yes, usually aliases to Write-Output or is a binary)
	// For robustness on Windows, maybe we should use 'cmd /c echo' or rely on 'echo'.
	// Go's exec.Command handles 'echo' on Linux/Mac. On Windows, it needs help if it's a builtin.
	// Safe bet for portable test: "go version"
	toolsConfig := `
tools:
  - name: check_go
    command: go
    args: ["version"]
    description: Check Go Version
`
	err := os.WriteFile(filepath.Join(tmpDir, "tools.yaml"), []byte(toolsConfig), 0644)
	require.NoError(t, err)

	// 3. Create Flow (start.md) using the Universal Action syntax
	// Note: We use 'x-exec-command' to test the flattening logic if we wanted inline,
	// but here we test the Registry (which is the safe path).
	flowContent := `---
id: start
do:
  name: check_go
save_to: version_output
---
Checking version...
`
	err = os.WriteFile(filepath.Join(tmpDir, "start.md"), []byte(flowContent), 0644)
	require.NoError(t, err)

	// 4. Initialize Config
	loadedConfig, err := process.LoadTools(filepath.Join(tmpDir, "tools.yaml"))
	require.NoError(t, err)

	// 5. Initialize Process Runner
	procRunner := process.NewRunner(process.WithRegistry(loadedConfig))

	// 6. Initialize Engine
	// We use a mock IOHandler to capture output/input if needed, but here we just run.
	engine, err := trellis.New(tmpDir)
	require.NoError(t, err)

	// 7. Initialize Core Runner with ToolRunner
	r := runner.NewRunner(
		runner.WithHeadless(true),
		runner.WithToolRunner(procRunner),
		runner.WithEngine(engine),
	)

	// 8. Run
	ctx := context.Background()
	err = r.Run(ctx)
	require.NoError(t, err)

	// 9. Verify Result
	// The output of 'go version' should be in context["version_output"]
	finalState := r.State()
	require.NotNil(t, finalState)
	val, exists := finalState.Context["version_output"]
	assert.True(t, exists, "Expected version_output in context")
	assert.IsType(t, "", val)
	assert.Contains(t, val.(string), "go version", "Tool output should contain 'go version'")
}

// TestProcessAdapter_Inline_Flattening verifies that nested x-exec metadata in YAML
// is correctly flattened by the Loader and executed by the Runner.
func TestProcessAdapter_Inline_Flattening(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create Flow with Nested YAML Metadata
	// This tests the 'loam' adapter flattening fix.
	flowContent := `---
id: start
# Universal Syntax: Action Node
do:
  name: dynamic_echo
  metadata:
    x-exec:
      command: go
      args: ["env", "GOOS"]
save_to: os_name
---
Checking OS...
`
	err := os.WriteFile(filepath.Join(tmpDir, "start.md"), []byte(flowContent), 0644)
	require.NoError(t, err)

	// 2. Initialize Engine & Process Runner with Inline Enabled
	engine, err := trellis.New(tmpDir)
	require.NoError(t, err)

	procRunner := process.NewRunner(
		process.WithInlineExecution(true), // DANGEROUS! Enabled for test.
	)

	r := runner.NewRunner(
		runner.WithHeadless(true),
		runner.WithToolRunner(procRunner),
		runner.WithEngine(engine),
	)

	// 3. Run
	err = r.Run(context.Background())
	require.NoError(t, err)

	//  4. Verify
	finalState := r.State()
	val, exists := finalState.Context["os_name"]
	assert.True(t, exists, "Expected os_name in context")
	// On Windows, should be windows. On Linux, linux.
	t.Logf("Detected OS: %v", val)
	assert.NotEmpty(t, val)
}

// TestProcessAdapter_ComplexArgs verifies that complex arguments (Maps/Slices)
// are serialized as valid JSON strings in environment variables.
func TestProcessAdapter_ComplexArgs(t *testing.T) {
	// 1. Setup
	procRunner := process.NewRunner(process.WithInlineExecution(true))

	// 2. Define complex args
	args := map[string]any{
		"data": map[string]any{
			"foo":  "bar",
			"list": []int{1, 2, 3},
		},
	}

	toolCall := domain.ToolCall{
		Name: "json_check",
		Args: args,
		Metadata: map[string]string{
			"x-exec-command": "go",
			"x-exec-args":    "run ./testdata/json_echo.go", // We will mock this command logic differently to keep test self-contained
		},
	}

	// Because creating a separate Go file for the test command is complex,
	// let's use a simple shell approach or just rely on 'echo'.
	// Windows 'echo' is tricky with quotes.
	// Alternative: Inspect the Runner logic directly? No, we want integration.
	// Let's use `go run` with a tiny inline helper if possible, or write a temporary handled script.

	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "inspector.go")
	scriptContent := `package main
import (
	"fmt"
	"os"
)
func main() {
	// Print the env var raw
	fmt.Print(os.Getenv("TRELLIS_ARGS"))
}`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	require.NoError(t, err)

	toolCall.Metadata["x-exec-command"] = "go"
	toolCall.Metadata["x-exec-args"] = "run " + scriptPath

	// 3. Execute
	result, err := procRunner.Execute(context.Background(), toolCall)
	require.NoError(t, err)

	// 4. Verify Output
	// Should be valid JSON: {"foo":"bar","list":[1,2,3]}
	// Note: keys in map are unordered, so exact string match is flaky.
	// But since it's small, it might be stable or we parse it back.

	// The runner now AUTO-PARSES JSON output.
	// If the script prints JSON, the runner parses it.
	// Our script prints the ENV VAR. If the ENV VAR is JSON, the script prints JSON.
	// So the runner should parse it back to a map.

	assert.IsType(t, map[string]any{}, result.Result, "Expected result to be parsed JSON object")
	fullRes := result.Result.(map[string]any)
	resMap := fullRes["data"].(map[string]any)

	assert.Equal(t, "bar", resMap["foo"])
	// JSON numbers are float64 by default in unmarshal
	list := resMap["list"].([]any)
	assert.Equal(t, 3, len(list))
}
