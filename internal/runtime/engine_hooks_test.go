package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestEngine_LifecycleHooks(t *testing.T) {
	// Setup
	nodes := map[string]string{
		"start": `{
			"id": "start",
			"type": "text",
			"transitions": [
				{ "to_node_id": "step_2" }
			]
		}`,
		"step_2": `{
			"id": "step_2",
			"type": "text"
		}`,
	}
	loader := inmemory.New(nodes)

	// Capture events
	var entered []string
	var left []string

	hooks := domain.LifecycleHooks{
		OnNodeEnter: func(ctx context.Context, e *domain.NodeEvent) {
			entered = append(entered, e.NodeID)
		},
		OnNodeLeave: func(ctx context.Context, e *domain.NodeEvent) {
			left = append(left, e.NodeID)
		},
	}

	// Initialize Engine with Hooks
	// engine := runtime.NewEngine(loader, nil, nil, runtime.WithLifecycleHooks(hooks))
	// Note: Engine is in internal/runtime, so we call runtime.NewEngine
	engine := runtime.NewEngine(loader, nil, nil, runtime.WithLifecycleHooks(hooks))

	// Execution
	ctx := context.Background()

	// Verify OnNodeEnter for initial state (Now triggered by Start)
	state, err := engine.Start(ctx, nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if len(entered) != 1 || entered[0] != "start" {
		t.Errorf("Expected enter 'start' on Start(), got: %v", entered)
	}

	// Move to step_2
	var nextState *domain.State
	nextState, err = engine.Navigate(ctx, state, "")
	if err != nil {
		t.Fatalf("Navigate failed: %v", err)
	}
	_ = nextState

	// Move from step_2 to Termination
	var termState *domain.State
	termState, err = engine.Navigate(ctx, nextState, "")
	if err != nil {
		t.Fatalf("Navigate from step_2 failed: %v", err)
	}
	if termState.Status != domain.StatusTerminated {
		t.Errorf("Expected terminated status, got %s", termState.Status)
	}

	// Verify Events
	// We should have:
	// - Returned from "start" (OnNodeLeave)
	// - Returned from "step_2" (OnNodeLeave on termination)
	// - Entered "start" (OnNodeEnter on Start)
	// - Entered "step_2" (OnNodeEnter on Navigate)

	if len(left) != 2 || left[0] != "start" || left[1] != "step_2" {
		t.Errorf("Expected leave [start, step_2], got: %v", left)
	}

	if len(entered) != 2 || entered[0] != "start" || entered[1] != "step_2" {
		t.Errorf("Expected enter [start, step_2], got: %v", entered)
	}
}
