package runtime_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestEngine_Extreme_StateIntegrity(t *testing.T) {
	node := domain.Node{ID: "start", Content: []byte("Start")}
	loader, _ := memory.NewFromNodes(node)
	engine := runtime.NewEngine(loader, nil, nil)

	t.Run("Nil State", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Engine panicked on nil state: %v", r)
			}
		}()
		_, _, err := engine.Render(context.Background(), nil)
		if err == nil {
			t.Error("Expected error for nil state in Render, got nil")
		}
		_, err = engine.Navigate(context.Background(), nil, "go")
		if err == nil {
			t.Error("Expected error for nil state in Navigate, got nil")
		}
	})

	t.Run("Invalid Node ID in State", func(t *testing.T) {
		state := domain.NewState("sid", "ghost_node")
		_, _, err := engine.Render(context.Background(), state)
		if err == nil || !strings.Contains(err.Error(), "failed to load node") {
			t.Errorf("Expected node loading error, got: %v", err)
		}
	})
}

func TestEngine_Extreme_InterpolationStress(t *testing.T) {
	t.Run("Deeply Nested Context Access", func(t *testing.T) {
		node := domain.Node{
			ID:      "stress",
			Content: []byte("Value: {{ .a.b.c.d.e.f.g }}"),
		}
		loader, _ := memory.NewFromNodes(node)
		engine := runtime.NewEngine(loader, nil, nil)

		state := domain.NewState("sid", "stress")
		state.Context["a"] = map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": map[string]any{
						"e": map[string]any{
							"f": map[string]any{
								"g": "found_me",
							},
						},
					},
				},
			},
		}

		actions, _, err := engine.Render(context.Background(), state)
		if err != nil {
			t.Fatalf("Deep interpolation failed: %v", err)
		}
		if actions[0].Payload != "Value: found_me" {
			t.Errorf("Expected 'Value: found_me', got '%v'", actions[0].Payload)
		}
	})

	t.Run("Large Payload Interpolation", func(t *testing.T) {
		largeStr := strings.Repeat("A", 10000)
		node := domain.Node{
			ID:      "large",
			Content: []byte("Data: {{ .data }}"),
		}
		loader, _ := memory.NewFromNodes(node)
		engine := runtime.NewEngine(loader, nil, nil)

		state := domain.NewState("sid", "large")
		state.Context["data"] = largeStr

		actions, _, err := engine.Render(context.Background(), state)
		if err != nil {
			t.Fatalf("Large interpolation failed: %v", err)
		}
		expected := "Data: " + largeStr
		if actions[0].Payload != expected {
			t.Errorf("Payload mismatch. Lengths: expected %d, got %d", len(expected), len(actions[0].Payload.(string)))
		}
	})
}

func TestEngine_Extreme_PrioritySynergy(t *testing.T) {
	// Goal: Verify hierarchy: Conditional > OnDenied > Fallback
	nodes := []domain.Node{
		{
			ID:        "complex",
			InputType: "confirm",
			Transitions: []domain.Transition{
				{ToNodeID: "override", Condition: "input == 'yes'"}, // Specific Logic
				{ToNodeID: "default"},                               // Fallback
			},
			OnDenied: "denial_handler",
		},
		{ID: "override"},
		{ID: "denial_handler"},
		{ID: "default"},
	}

	loader, _ := memory.NewFromNodes(nodes...)
	engine := runtime.NewEngine(loader, nil, nil)

	t.Run("Logic Overrides Denial", func(t *testing.T) {
		state := domain.NewState("sid", "complex")
		// User replies 'yes'. Condition matches.
		// Priority 1 (Conditional) should win even though it's not a refusal.
		next, err := engine.Navigate(context.Background(), state, "yes")
		if err != nil {
			t.Fatalf("Navigate failed: %v", err)
		}
		if next.CurrentNodeID != "override" {
			t.Errorf("Priority failure: logic should win. Got: %s", next.CurrentNodeID)
		}
	})

	t.Run("Denial Overrides Default", func(t *testing.T) {
		state := domain.NewState("sid", "complex")
		// User replies 'no'. Logic doesn't match (input == 'yes').
		// Priority 2 (OnDenied) should win over Priority 3 (Default 'to').
		next, err := engine.Navigate(context.Background(), state, "no")
		if err != nil {
			t.Fatalf("Navigate failed: %v", err)
		}
		if next.CurrentNodeID != "denial_handler" {
			t.Errorf("Priority failure: on_denied should win over default. Got: %s", next.CurrentNodeID)
		}
	})

	t.Run("Error Handlers Priority", func(t *testing.T) {
		// Scenario: Tool Fails. Check prioritized handlers.
		errorNodes := []domain.Node{
			{
				ID:      "error_test",
				Do:      &domain.ToolCall{ID: "t1", Name: "t"},
				OnError: "node_error",
			},
			{ID: "node_error"},
			{ID: "global_error"},
		}
		l, _ := memory.NewFromNodes(errorNodes...)
		e := runtime.NewEngine(l, nil, nil, runtime.WithDefaultErrorNode("global_error"))

		state := domain.NewState("sid", "error_test")
		state.Status = domain.StatusWaitingForTool
		state.PendingToolCall = "t1"

		res := domain.ToolResult{ID: "t1", IsError: true}
		next, _ := e.Navigate(context.Background(), state, res)
		if next.CurrentNodeID != "node_error" {
			t.Errorf("Expected node-level error handler to win, got %s", next.CurrentNodeID)
		}
	})
}

func TestEngine_Extreme_Concurrency(t *testing.T) {
	node := domain.Node{ID: "shared", Content: []byte("Data: {{ .val }}")}
	loader, _ := memory.NewFromNodes(node)
	engine := runtime.NewEngine(loader, nil, nil)

	t.Run("Parallel Rendering", func(t *testing.T) {
		const goroutines = 20
		errs := make(chan error, goroutines)

		for i := 0; i < goroutines; i++ {
			go func(val int) {
				state := domain.NewState("sid", "shared")
				state.Context["val"] = val
				actions, _, err := engine.Render(context.Background(), state)
				if err != nil {
					errs <- err
					return
				}
				expected := fmt.Sprintf("Data: %d", val)
				if actions[0].Payload != expected {
					errs <- fmt.Errorf("concurrency corruption: expected %s, got %s", expected, actions[0].Payload)
					return
				}
				errs <- nil
			}(i)
		}

		for i := 0; i < goroutines; i++ {
			if err := <-errs; err != nil {
				t.Errorf("Concurrent render failed: %v", err)
			}
		}
	})
}
