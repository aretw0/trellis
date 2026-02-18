package process_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/aretw0/trellis/pkg/adapters/process"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildFixture compiles a Go program from fixtures/resilience subdirectories into a temp binary.
func buildFixture(t *testing.T, dirName string) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	root := wd
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Fatal("could not find project root (go.mod)")
		}
		root = parent
	}

	resilienceDir := filepath.Join(root, "tests", "fixtures", "resilience")
	sourcePath := filepath.Join(resilienceDir, dirName)

	exeName := dirName
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}

	destDir := t.TempDir()
	destPath := filepath.Join(destDir, exeName)

	// Build
	// Note: We build the package in the subdirectory
	cmd := exec.Command("go", "build", "-o", destPath, sourcePath)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build fixture %s: %s", dirName, string(out))

	return destPath
}

func TestResilience_GoodCitizen(t *testing.T) {
	exe := buildFixture(t, "good_citizen")

	r := process.NewRunner(process.WithInlineExecution(true))

	toolCall := domain.ToolCall{
		ID:   "test-good-citizen",
		Name: "good_citizen",
		Args: map[string]any{},
		Metadata: map[string]string{
			"x-exec-command": exe,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	startTime := time.Now()
	result, err := r.Execute(ctx, toolCall)
	duration := time.Since(startTime)

	t.Logf("Duration: %v, ExecError: %v, ResultError: %s, IsError: %v", duration, err, result.Error, result.IsError)

	// In Windows, graceful signal propagation is limited for background processes.
	// It will likely hit the force-kill grace period of 5s.
	if runtime.GOOS == "windows" {
		assert.Greater(t, duration, 2*time.Second)
		assert.Less(t, duration, 10*time.Second, "On Windows, it should fallback to force-kill")
	} else {
		assert.Greater(t, duration, 2*time.Second)
		assert.Less(t, duration, 5*time.Second, "On Unix, it should exit gracefully quickly after signal")
	}

	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Error, "deadline exceeded")
}

func TestResilience_BadCitizen_Ignore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}
	exe := buildFixture(t, "bad_citizen_ignore")

	r := process.NewRunner(process.WithInlineExecution(true))
	toolCall := domain.ToolCall{
		ID:       "test-ignore-citizen",
		Name:     "ignore_citizen",
		Metadata: map[string]string{"x-exec-command": exe},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	startTime := time.Now()
	result, err := r.Execute(ctx, toolCall)
	duration := time.Since(startTime)

	t.Logf("Duration: %v, ResultError: %s, ExecError: %v", duration, result.Error, err)

	// Should wait for grace period (5s)
	assert.Greater(t, duration, 5*time.Second)
	assert.NoError(t, err)
}

func TestResilience_BadCitizen_Slow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow test in short mode")
	}
	exe := buildFixture(t, "bad_citizen_slow")

	r := process.NewRunner(process.WithInlineExecution(true))

	toolCall := domain.ToolCall{
		ID:   "test-slow-citizen",
		Name: "slow_citizen",
		Args: map[string]any{},
		Metadata: map[string]string{
			"x-exec-command": exe,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	startTime := time.Now()
	result, err := r.Execute(ctx, toolCall)
	duration := time.Since(startTime)

	t.Logf("Duration: %v, ExecError: %v, ResultError: %s, IsError: %v", duration, err, result.Error, result.IsError)

	assert.Greater(t, duration, 5*time.Second, "Should wait for at least grace period (5s)")
	assert.NoError(t, err)
}

func TestResilience_Crashy(t *testing.T) {
	exe := buildFixture(t, "crashy")

	r := process.NewRunner(process.WithInlineExecution(true))
	toolCall := domain.ToolCall{
		ID:       "test-crashy",
		Name:     "crashy",
		Metadata: map[string]string{"x-exec-command": exe},
	}

	result, err := r.Execute(context.Background(), toolCall)
	assert.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Contains(t, result.Error, "exit status 123")
	assert.Contains(t, result.Error, "Something went terribly wrong")
}
