package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/process"
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
	)

	// 8. Run
	ctx := context.Background()
	finalState, err := r.Run(ctx, engine, nil)
	require.NoError(t, err)

	// 9. Verify Result
	// The output of 'go version' should be in context["version_output"]
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
	)

	// 3. Run
	finalState, err := r.Run(context.Background(), engine, nil)
	require.NoError(t, err)

	// 4. Verify
	val, exists := finalState.Context["os_name"]
	assert.True(t, exists, "Expected os_name in context")
	// On Windows, should be windows. On Linux, linux.
	t.Logf("Detected OS: %v", val)
	assert.NotEmpty(t, val)
}
