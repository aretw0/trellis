package runner

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestRunner_Run_BasicFlow(t *testing.T) {
	// 1. Setup Engine with Memory Loader
	loader, err := inmemory.NewFromNodes(
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
		},
	)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	engine, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// 2. Setup Runner with TextHandler (Inputs pre-filled)
	inputBuf := bytes.NewBufferString("\n\n") // Enter for start
	outputBuf := &bytes.Buffer{}

	r := NewRunner()
	r.Handler = NewTextHandler(inputBuf, outputBuf)
	r.Interceptor = AutoApproveMiddleware()

	// 3. Run in a goroutine to prevent deadlock
	inputBuf.WriteString("exit\n") // Ensure we exit at the end

	done := make(chan error)
	go func() {
		_, err := r.Run(t.Context(), engine, nil)
		done <- err
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
	loader, _ := inmemory.NewFromNodes(
		domain.Node{
			ID:      "start",
			Type:    domain.NodeTypeText,
			Content: []byte("Headless Mode"),
		},
	)
	engine, _ := trellis.New("", trellis.WithLoader(loader))

	// 2. Setup Runner (Headless)
	inBuf := bytes.NewBufferString("\"exit\"\n")
	outBuf := &bytes.Buffer{}

	r := NewRunner()
	r.Handler = NewJSONHandler(inBuf, outBuf)
	r.Headless = true

	// 3. Run
	done := make(chan error)
	go func() {
		_, err := r.Run(t.Context(), engine, nil)
		done <- err
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
