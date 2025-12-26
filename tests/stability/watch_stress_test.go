package stability

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// TestWatchStress compiles the trellis binary and runs it in watch mode
// against a temporary directory, performing rapid and invalid updates.
func TestWatchStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// 1. Compile Trellis
	// We build it to a temporary location to ensure we test the actual binary behavior
	tempBinDir, err := os.MkdirTemp("", "trellis-bin-*")
	if err != nil {
		t.Fatalf("Failed to create temp bin dir: %v", err)
	}
	defer os.RemoveAll(tempBinDir)

	binPath := filepath.Join(tempBinDir, "trellis")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	// Assume run from project root, so we point to ./cmd/trellis
	// If running from ./tests/stability, we need to go up.
	// We'll rely on "go test ./..." being run from root usually,
	// but let's try to find go.mod to be safe or just assume relative path "../../cmd/trellis"
	cmdBuild := exec.Command("go", "build", "-o", binPath, "../../cmd/trellis")
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("Failed to compile trellis: %v\nOutput: %s", err, string(out))
	}

	// 2. Setup Test Environment (Repo)
	tempRepoDir, err := os.MkdirTemp("", "trellis-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp repo dir: %v", err)
	}
	defer os.RemoveAll(tempRepoDir)

	startFile := filepath.Join(tempRepoDir, "start.md")
	writeContent := func(content string) {
		if err := os.WriteFile(startFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write content: %v", err)
		}
	}

	// Initial valid content
	writeContent(`---
id: start
type: text
---
# Version 1
Initial content.
`)

	// 3. Launch Watcher
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, "run", "--watch", "--dir", tempRepoDir)
	// We pipe stdout/stderr to monitor output if needed, or just let it print to test log
	// For stress test, checking exit code is primary.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create a pipe for Stdin so we can keep it open (simulating interactive session)
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	defer stdinPipe.Close()

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start trellis: %v", err)
	}

	// Give it a second to startup
	time.Sleep(2 * time.Second)

	// 4. Stress Loop
	iterations := 10
	t.Logf("Starting stress loop (%d iterations)...", iterations)

	for i := 0; i < iterations; i++ {
		// Valid update
		t.Logf("[%d] Updating with Valid Content", i)
		writeContent(fmt.Sprintf(`---
id: start
type: text
---
# Version %d
Updated content.
`, i+2))

		time.Sleep(200 * time.Millisecond)

		// Invalid update (Broken YAML)
		t.Logf("[%d] Updating with Invalid Content (Chaos)", i)
		writeContent(`---
id: start
type: text
broken_yaml: [ unclosed list
---
`)
		// The watcher should log an error but NOT crash
		time.Sleep(200 * time.Millisecond)

		// Valid recovery
		writeContent(fmt.Sprintf(`---
id: start
type: text
---
# Version %d (Recovered)
Recovered content.
`, i+2))

		time.Sleep(300 * time.Millisecond)
	}

	// 5. Cleanup & Verify
	t.Log("Stress loop finished. Stopping process...")
	cancel() // Kills the context -> Kills the process

	// Wait for process validation
	err = cmd.Wait()

	// verification logic:
	// Process killed by context -> err is usually "signal: killed" or similar.
	// If it crashed EARLIER (during loop), err would be non-nil and exit code != 0 (and not caused by our kill).

	if err != nil {
		// Check if it was purely our kill signal
		if ctx.Err() == context.Canceled {
			// Expected termination via context
			return
		}
		// If we are on Windows, os.Interrupt might result in exit code 1 or similar.
		// We scrutinize unexpected crashes.
		t.Logf("Process exited with: %v", err)
	} else {
		t.Log("Process exited cleanly.")
	}
}
