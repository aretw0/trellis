package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestEngine_SAGA_ExplicitRollback(t *testing.T) {
	// Scenario:
	// 1. Step 1 (Tool): executes "charge" (Undo: "refund")
	// 2. Step 2 (Tool): executes "logic_check" -> returns "rejected"
	// 3. Step 2 (Transition): "to: rollback" triggered by logic
	// 4. Engine unwinds to Step 1

	step1 := domain.Node{
		ID:   "step1",
		Type: domain.NodeTypeTool,
		Do:   &domain.ToolCall{ID: "charge", Name: "charge_card"},
		Undo: &domain.ToolCall{ID: "refund", Name: "refund_card"},
		Transitions: []domain.Transition{
			{ToNodeID: "step2"},
		},
	}

	step2 := domain.Node{
		ID:   "step2",
		Type: domain.NodeTypeTool,
		Do:   &domain.ToolCall{ID: "check", Name: "check_fraud"},
		// No OnError here!
		Transitions: []domain.Transition{
			{
				Condition: "input == 'rejected'",
				ToNodeID:  "rollback", // Explicit transition
			},
		},
	}

	// Capture Leave Events
	var leaveEvents []*domain.NodeEvent
	hooks := domain.LifecycleHooks{
		OnNodeLeave: func(ctx context.Context, e *domain.NodeEvent) {
			leaveEvents = append(leaveEvents, e)
		},
	}

	loader, _ := memory.NewFromNodes(step1, step2)
	engine := runtime.NewEngine(loader, nil, nil, runtime.WithLifecycleHooks(hooks))

	ctx := context.Background()
	state := domain.NewState("sess-explicit", "step1")

	// 1. Skip to Step 2 Waiting
	state.History = []string{"step1", "step2"}
	state.CurrentNodeID = "step2"
	state.Status = domain.StatusWaitingForTool
	state.PendingToolCall = "check"

	// 2. Return "rejected" from tool
	rollbackState, err := engine.Navigate(ctx, state, domain.ToolResult{
		ID:     "check",
		Result: "rejected",
	})
	if err != nil {
		t.Fatalf("Navigate failed: %v", err)
	}

	// 3. Verify Rollback started
	if rollbackState.Status != domain.StatusRollingBack {
		t.Errorf("Expected StatusRollingBack, got %s", rollbackState.Status)
	}
	// It should have popped step2?
	// startRollback calls continueRollback(popCurrent=true)
	// check_fraud (step2) has NO undo.
	// charge_card (step1) HAS undo.
	// So step2 is popped, we see step1.
	if rollbackState.CurrentNodeID != "step1" {
		t.Errorf("Expected CurrentNodeID 'step1', got %s", rollbackState.CurrentNodeID)
	}
	if rollbackState.PendingToolCall != "refund" {
		t.Errorf("Expected PendingToolCall 'refund', got %s", rollbackState.PendingToolCall)
	}

	// 4. Verify Lifecycle Events (Sober Analysis)
	// We expect 'step2' to emit OnNodeLeave when entering rollback state.
	// We might also expect others if the rollback touches them, but step2 is critical.
	foundStep2Leave := false
	for _, evt := range leaveEvents {
		if evt.NodeID == "step2" && evt.Type == domain.EventNodeLeave {
			foundStep2Leave = true
			break
		}
	}
	if !foundStep2Leave {
		t.Errorf("Expected OnNodeLeave event for 'step2' during rollback transition, but it was missing. Events: %v", leaveEvents)
	}
}
