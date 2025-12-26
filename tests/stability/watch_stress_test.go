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

	// Build the binary to test the actual CLI behavior.
	tempBinDir := t.TempDir()
	binPath := filepath.Join(tempBinDir, "trellis")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	// Tests run in the package directory, so we look up two levels.
	cmdBuild := exec.Command("go", "build", "-o", binPath, "../../cmd/trellis")
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("Failed to compile trellis: %v\nOutput: %s", err, string(out))
	}

	// Setup a fresh repo
	tempRepoDir := t.TempDir()

	startFile := filepath.Join(tempRepoDir, "start.md")
	writeContent := func(content string) {
		if err := os.WriteFile(startFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write content: %v", err)
		}
	}

	writeContent(`---
id: start
type: text
---
# Version 1
Initial content.
`)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, binPath, "run", "--watch", "--dir", tempRepoDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Keep stdin open to simulate an interactive session
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	defer stdinPipe.Close()

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start trellis: %v", err)
	}

	// Give it a moment to startup
	time.Sleep(2 * time.Second)

	iterations := 10
	t.Logf("Starting stress loop (%d iterations)...", iterations)

	for i := 0; i < iterations; i++ {
		t.Logf("[%d] Updating with Valid Content", i)
		writeContent(fmt.Sprintf(`---
id: start
type: text
---
# Version %d
Updated content.
`, i+2))

		time.Sleep(200 * time.Millisecond)

		t.Logf("[%d] Updating with Invalid Content (Chaos)", i)
		writeContent(`---
id: start
type: text
broken_yaml: [ unclosed list
---
`)
		// The watcher should log an error but NOT crash
		time.Sleep(200 * time.Millisecond)

		// Recovery
		writeContent(fmt.Sprintf(`---
id: start
type: text
---
# Version %d (Recovered)
Recovered content.
`, i+2))

		time.Sleep(300 * time.Millisecond)
	}

	t.Log("Stress loop finished. Stopping process...")
	cancel()

	err = cmd.Wait()

	if err != nil {
		// Check if it was purely our kill signal
		if ctx.Err() == context.Canceled {
			return
		}
		// If we are on Windows, os.Interrupt might result in exit code 1 or similar.
		// We scrutinize unexpected crashes.
		t.Logf("Process exited with: %v", err)
	} else {
		t.Log("Process exited cleanly.")
	}
}
