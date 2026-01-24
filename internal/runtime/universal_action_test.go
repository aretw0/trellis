package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

// TestEngine_UniversalAction verifies that any node type can execute a tool if 'Do' is present.
func TestEngine_UniversalAction(t *testing.T) {
	// Scenario: A "Text" node (implicitly) that also executes a tool "init_db".
	// The transitions depend on the tool result.
	node := domain.Node{
		ID:      "universal_node",
		Type:    domain.NodeTypeText, // Explicitly Text, but has action
		Content: []byte("Initializing..."),
		Do: &domain.ToolCall{
			ID:   "call_1",
			Name: "init_db",
		},
		Transitions: []domain.Transition{
			{Condition: "input == 'success'", ToNodeID: "success_node"},
			{ToNodeID: "fail_node"},
		},
	}
	successNode := domain.Node{ID: "success_node", Content: []byte("Done")}
	failNode := domain.Node{ID: "fail_node", Content: []byte("Failed")}

	loader, _ := memory.NewFromNodes(node, successNode, failNode)
	engine := runtime.NewEngine(loader, nil, nil) // Default Evaluator

	// 1. Render Phase
	state := domain.NewState("test-session", "universal_node")
	actions, _, err := engine.Render(context.Background(), state)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Verify Implicit "WaitingForTool" status?
	// Note: Render does NOT change state status (it's pure).
	// But `transitionTo` creates the initial status.
	// We need to check if we can simulate the "WaitingForTool" check during Navigate.

	// Check Actions: Text AND Tool
	if len(actions) != 2 {
		t.Fatalf("Expected 2 actions (Text + Tool), got %d", len(actions))
	}
	if actions[0].Type != domain.ActionRenderContent {
		t.Errorf("Expected Action 1 to be RenderContent")
	}
	if actions[1].Type != domain.ActionCallTool {
		t.Errorf("Expected Action 2 to be CallTool")
	}

	// 2. Simulate Status Update (usually done by transitionTo when entering the node)
	// We manually set it to match what transitionTo would do given our code change.
	state.Status = domain.StatusWaitingForTool
	state.PendingToolCall = "call_1"

	// 3. Navigate Phase (Tool returns Success)
	toolResult := domain.ToolResult{
		ID:     "call_1",
		Result: "success",
	}

	nextState, err := engine.Navigate(context.Background(), state, toolResult)
	if err != nil {
		t.Fatalf("Navigate failed: %v", err)
	}

	if nextState.CurrentNodeID != "success_node" {
		t.Errorf("Expected transition to 'success_node', got '%s'", nextState.CurrentNodeID)
	}
}
