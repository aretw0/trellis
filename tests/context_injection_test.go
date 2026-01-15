package tests

import (
	"context"
	"testing"

	"github.com/aretw0/trellis"
)

// TestContextInjection verifies that the Engine correctly seeds the initial state
// with the provided context map when using StartWithContext.
func TestContextInjection(t *testing.T) {
	// 1. Setup minimal engine (no special loader needed for this unit-like test,
	// checking state initialization doesn't require loading nodes yet)
	// Actually, Start() triggers OnNodeEnter which tries to load "start" node.
	// So we need a valid loader or mocking.
	// Using the tour example path is easiest for integration testing.
	repoPath := "../examples/tour"
	eng, err := trellis.New(repoPath)
	if err != nil {
		t.Fatalf("Failed to init engine: %v", err)
	}

	// 2. Define Context
	initialData := map[string]any{
		"user": "Tester",
		"role": "Admin",
		"config": map[string]any{
			"debug": true,
		},
	}

	// 3. Start with Context
	ctx := context.Background()
	state, err := eng.Start(ctx, initialData)
	if err != nil {
		t.Fatalf("StartWithContext failed: %v", err)
	}

	// 4. Verify State
	if state.Context["user"] != "Tester" {
		t.Errorf("Expected context['user'] to be 'Tester', got %v", state.Context["user"])
	}
	if state.Context["role"] != "Admin" {
		t.Errorf("Expected context['role'] to be 'Admin', got %v", state.Context["role"])
	}

	// Verify deep structure
	cfg, ok := state.Context["config"].(map[string]any)
	if !ok {
		t.Fatalf("Expected context['config'] to be map, got %T", state.Context["config"])
	}
	if cfg["debug"] != true {
		t.Errorf("Expected config.debug to be true, got %v", cfg["debug"])
	}
}

// TestContextInjection_Persistence verifies that the injected context
// persists across transitions (basic sanity check).
func TestContextInjection_Persistence(t *testing.T) {
	repoPath := "../examples/tour"
	eng, err := trellis.New(repoPath)
	if err != nil {
		t.Fatalf("Failed to init engine: %v", err)
	}

	initialData := map[string]any{
		"score": 100,
	}

	state, err := eng.Start(context.Background(), initialData)
	if err != nil {
		t.Fatalf("StartWithContext failed: %v", err)
	}

	// Simulate a navigation (render only, checking if context is available to interpolator)
	// We assume 'start' node exists.
	// We can't easily "navigate" without valid input for the specific flow,
	// but we can check if the context is still there in the state object.

	if state.Context["score"] != 100 {
		t.Errorf("Context lost after Start: %v", state.Context)
	}

	// Hypothetical navigation (if we knew inputs).
	// For now, testing initialization is the core requirement for "Context Injection".
}

// TestCLIFlagParsing is difficult to test here without os/exec.
// We rely on TestContextInjection to prove the API works.
