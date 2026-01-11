package runner

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestRunner_Run_BasicFlow(t *testing.T) {
	// 1. Setup Engine with Memory Loader
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:      "start",
			Type:    domain.NodeTypeText,
			Content: []byte("Welcome to Trellis"),
			Transitions: []domain.Transition{
				{ToNodeID: "end"},
			},
		},
		domain.Node{
			ID:      "end",
			Type:    domain.NodeTypeText,
			Content: []byte("Goodbye"),
			// No transitions -> implicit termination or sink state?
			// Trellis engine doesn't auto-terminate on leaf nodes unless configured.
			// Let's make "end" clearly terminal or just check we reached it.
		},
	)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	// trellis.New signature: func New(repoPath string, opts ...Option) (*Engine, error)
	// When using WithLoader, repoPath can be empty.
	engine, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// 2. Setup Runner with TextHandler (Inputs pre-filled)
	// We need to provide input to advance "start" -> "end".
	// "start" is text node, waits for default input (enter) to proceed.
	inputBuf := bytes.NewBufferString("\n\n") // Enter for start
	outputBuf := &bytes.Buffer{}

	r := NewRunner()
	r.Handler = NewTextHandler(inputBuf, outputBuf)
	r.Interceptor = AutoApproveMiddleware() // Just in case

	// 3. Run in a goroutine to prevent deadlock if it hangs,
	// but for a unit test we want to control execution.
	// The Runner.Run loop blocks until termination.
	// We need a way to stop it. "exit" input or termination state.
	// If "end" node is a sink, the engine might just stay there?
	// The Engine defaults: Text nodes wait for input.
	// So at "end", it will wait for input. If we send "exit", it breaks.
	inputBuf.WriteString("exit\n") // Ensure we exit at the end

	done := make(chan error)
	go func() {
		done <- r.Run(engine)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Runner failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Runner timed out")
	}

	// 4. Verify Output
	out := outputBuf.String()
	if !strings.Contains(out, "Welcome to Trellis") {
		t.Error("Expected welcome message in output")
	}
	if !strings.Contains(out, "Goodbye") {
		t.Error("Expected goodbye message in output")
	}
}

func TestRunner_Run_Headless(t *testing.T) {
	// 1. Setup Engine
	loader, _ := memory.NewFromNodes(
		domain.Node{
			ID:      "start",
			Type:    domain.NodeTypeText,
			Content: []byte("Headless Mode"),
		},
	)
	engine, _ := trellis.New("", trellis.WithLoader(loader))

	// 2. Setup Runner (Headless)
	// For Headless, we use JSONHandler usually, but let's test the Headless flag behavior on Reader/Writer
	// Actually, `Headless` field in Runner is deprecated in favor of Handler strategy.
	// But let's use NewJSONHandler which serves the headless purpose.

	inBuf := bytes.NewBufferString("\"exit\"\n")
	outBuf := &bytes.Buffer{}

	r := NewRunner()
	r.Handler = NewJSONHandler(inBuf, outBuf)
	r.Headless = true // Legacy flag, might still be used by Interceptor defaults if not set

	// 3. Run
	done := make(chan error)
	go func() {
		done <- r.Run(engine)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Runner failed: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Runner timed out")
	}

	// 4. Verify JSON Output
	out := outBuf.String()
	if !strings.Contains(out, "Headless Mode") {
		t.Errorf("Expected 'Headless Mode' in JSON output, got: %s", out)
	}
}
