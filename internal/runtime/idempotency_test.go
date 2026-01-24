package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func TestEngine_IdempotencyKeys(t *testing.T) {
	// Setup: A flow with a tool call
	node := domain.Node{
		ID:   "start",
		Type: domain.NodeTypeTool,
		Do: &domain.ToolCall{
			Name: "my_tool",
			ID:   "call_1",
		},
	}
	loader, _ := memory.NewFromNodes(node)
	engine := runtime.NewEngine(loader, nil, nil)

	ctx := context.Background()

	t.Run("Deterministic Keys for Same Session", func(t *testing.T) {
		sessionID := "session-A"

		// Run 1
		state1, err := engine.Start(ctx, sessionID, nil)
		if err != nil {
			t.Fatalf("Start failed: %v", err)
		}
		actions1, _, err := engine.Render(ctx, state1)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		requireToolCall(t, actions1)
		key1 := actions1[0].Payload.(domain.ToolCall).IdempotencyKey

		// Run 2 (Same Session ID, Fresh Start)
		state2, _ := engine.Start(ctx, sessionID, nil)
		actions2, _, _ := engine.Render(ctx, state2)
		requireToolCall(t, actions2)
		key2 := actions2[0].Payload.(domain.ToolCall).IdempotencyKey

		assert.NotEmpty(t, key1)
		assert.Equal(t, key1, key2, "Keys must be identical for same session and step")
	})

	t.Run("Different Keys for Different Sessions", func(t *testing.T) {
		// Run 1
		state1, _ := engine.Start(ctx, "session-B", nil)
		actions1, _, _ := engine.Render(ctx, state1)
		requireToolCall(t, actions1)
		key1 := actions1[0].Payload.(domain.ToolCall).IdempotencyKey

		// Run 2
		state2, _ := engine.Start(ctx, "session-C", nil)
		actions2, _, _ := engine.Render(ctx, state2)
		requireToolCall(t, actions2)
		key2 := actions2[0].Payload.(domain.ToolCall).IdempotencyKey

		assert.NotEqual(t, key1, key2, "Keys must differ for different sessions")
	})

	t.Run("Different Keys for Different Steps", func(t *testing.T) {
		// Verify that if we hypothetically had the same tool call at step 1 and step 2, keys differ.
		// We simulate this by manually appending to history.

		sessionID := "session-D"
		state, _ := engine.Start(ctx, sessionID, nil)
		// Step 0
		actions1, _, _ := engine.Render(ctx, state)
		key1 := actions1[0].Payload.(domain.ToolCall).IdempotencyKey

		// Manually advance history (Hack for unit test without full transition)
		state.History = append(state.History, "step1")

		// Render again (Step 1)
		actions2, _, _ := engine.Render(ctx, state)
		key2 := actions2[0].Payload.(domain.ToolCall).IdempotencyKey

		assert.NotEqual(t, key1, key2, "Keys must differ for different history lengths")
	})
}

func requireToolCall(t *testing.T, actions []domain.ActionRequest) {
	if len(actions) == 0 || actions[0].Type != domain.ActionCallTool {
		t.Helper()
		t.Fatalf("Expected tool call action")
	}
}
