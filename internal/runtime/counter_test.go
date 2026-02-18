package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestCounterPersistence(t *testing.T) {
	// Setup nodes simulating the reactivity-demo loop
	menuNode := domain.Node{
		ID:   "menu",
		Wait: true,
		Transitions: []domain.Transition{
			{ToNodeID: "loop", Condition: "input == '1'"},
		},
	}
	loopNode := domain.Node{
		ID:     "loop",
		Do:     &domain.ToolCall{ID: "t1", Name: "increment", Args: map[string]any{"count": "{{ .count }}"}},
		SaveTo: "count",
		Transitions: []domain.Transition{
			{ToNodeID: "loop-result"},
		},
		DefaultContext: map[string]any{"count": 0},
	}
	resultNode := domain.Node{
		ID:   "loop-result",
		Wait: true,
		Transitions: []domain.Transition{
			{ToNodeID: "menu"},
		},
	}

	loader, _ := memory.NewFromNodes(menuNode, loopNode, resultNode)
	engine := runtime.NewEngine(loader, nil, nil)

	ctx := context.Background()
	state, _ := engine.Start(ctx, "test", nil)

	// 1. Initial State (at menu via start.md -> menu skip for brevity here)
	state.CurrentNodeID = "menu"

	// 2. Navigate to loop
	state, err := engine.Navigate(ctx, state, "1")
	if err != nil {
		t.Fatalf("Navigate to loop failed: %v", err)
	}
	if state.CurrentNodeID != "loop" {
		t.Fatalf("Expected loop node, got %s", state.CurrentNodeID)
	}

	// 3. Simulate Tool Success (increment 0 -> 1)
	state, err = engine.Navigate(ctx, state, domain.ToolResult{
		ID:     "t1",
		Result: 1,
	})
	if err != nil {
		t.Fatalf("Navigate after tool failed: %v", err)
	}
	if state.Context["count"] != 1 {
		t.Errorf("Expected count 1, got %v", state.Context["count"])
	}

	// 4. Navigate back to menu from result
	state, err = engine.Navigate(ctx, state, "")
	if err != nil {
		t.Fatalf("Navigate to menu failed: %v", err)
	}

	// 5. Navigate to loop again
	state, err = engine.Navigate(ctx, state, "1")
	if err != nil {
		t.Fatalf("Navigate to loop 2 failed: %v", err)
	}

	// Check if count remained 1 before the next tool call
	if state.Context["count"] != 1 {
		t.Errorf("Expected count 1 to persist in menu, got %v", state.Context["count"])
	}

	// 6. Simulate Tool Success (increment 1 -> 2)
	state, err = engine.Navigate(ctx, state, domain.ToolResult{
		ID:     state.PendingToolCall,
		Result: 2,
	})
	if err != nil {
		t.Fatalf("Navigate after tool 2 failed: %v", err)
	}
	if state.Context["count"] != 2 {
		t.Errorf("Expected count 2, got %v", state.Context["count"])
	}
}
