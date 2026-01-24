package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestEngine_SAGA_Rollback(t *testing.T) {
	// Scenario:
	// 1. Step 1 (Tool): Successfully executes "charge_card" (Undo: "refund_card")
	// 2. Step 2 (Tool): Fails executing "ship_item" -> Triggers Rollback
	// 3. Engine unwinds to Step 1, executes "refund_card"
	// 4. Engine terminates

	// DO: Charge Card / UNDO: Refund
	step1 := domain.Node{
		ID:   "step1",
		Type: domain.NodeTypeTool,
		Do:   &domain.ToolCall{ID: "charge", Name: "charge_card"},
		Undo: &domain.ToolCall{ID: "refund", Name: "refund_card"},
		Transitions: []domain.Transition{
			{ToNodeID: "step2"},
		},
	}

	// DO: Ship Item (Fails)
	step2 := domain.Node{
		ID:      "step2",
		Type:    domain.NodeTypeTool,
		Do:      &domain.ToolCall{ID: "ship", Name: "ship_item"},
		OnError: "rollback", // Magic keyword
	}

	loader, _ := memory.NewFromNodes(step1, step2)
	engine := runtime.NewEngine(loader, nil, nil)

	ctx := context.Background()
	state := domain.NewState("sess-1", "step1")

	// 1. Start -> Wait for Charge
	// (Simulation: We skip the Render call and assume we are WaitingForTool)
	state.Status = domain.StatusWaitingForTool
	state.PendingToolCall = "charge"

	// 2. Complete Step 1 (Charge Success)
	state, err := engine.Navigate(ctx, state, domain.ToolResult{
		ID:     "charge",
		Result: "charged_ok",
	})
	if err != nil {
		t.Fatalf("Navigate (Step 1) failed: %v", err)
	}

	if state.CurrentNodeID != "step2" {
		t.Fatalf("Expected transition to step2, got %s", state.CurrentNodeID)
	}
	if state.Status != domain.StatusWaitingForTool {
		t.Fatalf("Expected WaitingForTool at step2")
	}

	// 3. Fail Step 2 (Ship Failure) -> Trigger Rollback
	rollbackState, err := engine.Navigate(ctx, state, domain.ToolResult{
		ID:      "ship",
		IsError: true,
		Result:  "shipping_unavailable",
	})
	if err != nil {
		t.Fatalf("Navigate (Step 2 Failure) failed: %v", err)
	}

	// EXPECTATIONS:
	// - Status: RollingBack
	// - CurrentNode: step1 (Popped step2)
	// - PendingToolCall: refund (Undo action of step1)

	if rollbackState.Status != domain.StatusRollingBack {
		t.Errorf("Expected StatusRollingBack, got %s", rollbackState.Status)
	}
	if rollbackState.CurrentNodeID != "step1" {
		t.Errorf("Expected CurrentNodeID to unwind to step1, got %s", rollbackState.CurrentNodeID)
	}
	if rollbackState.PendingToolCall != "refund" {
		t.Errorf("Expected PendingToolCall 'refund', got %s", rollbackState.PendingToolCall)
	}

	// 4. Render during Rollback (Should verify ActionCallTool for Undo)
	actions, _, err := engine.Render(ctx, rollbackState)
	if err != nil {
		t.Fatalf("Render (Rollback) failed: %v", err)
	}
	if len(actions) == 0 || actions[0].Type != domain.ActionCallTool {
		t.Fatalf("Expected ActionCallTool for Undo")
	}
	undoCall := actions[0].Payload.(domain.ToolCall)
	if undoCall.Name != "refund_card" {
		t.Errorf("Expected undo tool 'refund_card', got %s", undoCall.Name)
	}

	// 5. Complete Undo (Refund Success) -> Continue Unwind
	finalState, err := engine.Navigate(ctx, rollbackState, domain.ToolResult{
		ID:     "refund",
		Result: "refunded_ok",
	})
	if err != nil {
		t.Fatalf("Navigate (Undo) failed: %v", err)
	}

	// EXPECTATIONS:
	// - History empty (or start only) -> Terminated
	// - Since step1 was the start, popping it leaves empty history?
	// The logic: popCurrent=true after undo.
	// step1 is popped. History empty.
	// Result -> Terminated.

	if finalState.Status != domain.StatusTerminated {
		t.Errorf("Expected StatusTerminated after full rollback, got %s", finalState.Status)
	}
}
