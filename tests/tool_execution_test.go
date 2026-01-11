package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/domain"
)

// InMemoryLoader implements ports.GraphLoader for testing
type InMemoryLoader struct {
	Nodes map[string][]byte
}

func (l *InMemoryLoader) GetNode(id string) ([]byte, error) {
	if content, ok := l.Nodes[id]; ok {
		return content, nil
	}
	return nil, fmt.Errorf("node not found: %s", id)
}

func (l *InMemoryLoader) ListNodes() ([]string, error) {
	keys := make([]string, 0, len(l.Nodes))
	for k := range l.Nodes {
		keys = append(keys, k)
	}
	return keys, nil
}

// TestToolExecutionFlow verifies the full lifecycle of a side-effect:
// 1. Engine encounters Tool Node.
// 2. Engine status becomes WaitingForTool.
// 3. Render returns ActionCallTool.
// 4. Navigate resumes with ToolResult.
func TestToolExecutionFlow(t *testing.T) {
	// Define Graph
	toolCall := domain.ToolCall{
		ID:   "call_1",
		Name: "calculator",
		Args: map[string]interface{}{"op": "add", "a": 1, "b": 1},
	}

	nodes := map[string]domain.Node{
		"start": {
			ID:      "start",
			Type:    domain.NodeTypeText,
			Content: []byte("Welcome"),
			Transitions: []domain.Transition{
				{ToNodeID: "tool_node"},
			},
		},
		"tool_node": {
			ID:       "tool_node",
			Type:     domain.NodeTypeTool,
			ToolCall: &toolCall,
			Transitions: []domain.Transition{
				{ToNodeID: "success", Condition: "input == '2'"},
				{ToNodeID: "failure", Condition: ""}, // Default fallback
			},
		},
		"success": {
			ID:      "success",
			Type:    domain.NodeTypeText,
			Content: []byte("Success"),
		},
		"failure": {
			ID:      "failure",
			Type:    domain.NodeTypeText,
			Content: []byte("Failed"),
		},
	}

	rawNodes := make(map[string][]byte)
	for k, v := range nodes {
		b, _ := json.Marshal(v)
		rawNodes[k] = b
	}

	loader := &InMemoryLoader{Nodes: rawNodes}
	engine := runtime.NewEngine(loader, nil) // Default evaluator

	// 1. Start (Initial State)
	state := &domain.State{
		CurrentNodeID: "start",
		Status:        domain.StatusActive,
		Context:       make(map[string]any),
	}

	// 2. Navigate from Start -> Tool Node
	// This should set Status to WaitingForTool immediately because the target is a Tool Node.
	var err error
	state, err = engine.Navigate(context.Background(), state, "") // Empty input to proceed from text node logic?
	// Wait, ConditionEvaluator might treat "" as valid if Transition has no condition.

	if err != nil {
		t.Fatalf("Failed to navigate from start: %v", err)
	}

	if state.CurrentNodeID != "tool_node" {
		t.Fatalf("Expected current node to be 'tool_node', got %s", state.CurrentNodeID)
	}

	if state.Status != domain.StatusWaitingForTool {
		t.Fatalf("Expected status to be WaitingForTool, got %s", state.Status)
	}

	if state.PendingToolCall != "call_1" {
		t.Errorf("Expected PendingToolCall 'call_1', got %s", state.PendingToolCall)
	}

	// 3. Render (Should emit ActionCallTool)
	actions, _, err := engine.Render(context.Background(), state)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	foundTool := false
	for _, act := range actions {
		if act.Type == domain.ActionCallTool {
			foundTool = true
			if call, ok := act.Payload.(domain.ToolCall); ok {
				if call.ID != "call_1" {
					t.Errorf("Action Payload ID mismatch")
				}
			} else {
				t.Errorf("Action Payload type mismatch")
			}
		}
	}
	if !foundTool {
		t.Error("Render did not emit ActionCallTool")
	}

	// 4. Execute Tool (Simulated) -> "2"
	result := domain.ToolResult{
		ID:     "call_1",
		Result: 2, // Integer result
	}

	// 5. Navigate (Resume)
	state, err = engine.Navigate(context.Background(), state, result)
	if err != nil {
		t.Fatalf("Navigate (Resume) failed: %v", err)
	}

	// 6. Verify Transition to Success
	if state.CurrentNodeID != "success" {
		t.Errorf("Expected transition to 'success' (input='2'), got '%s'", state.CurrentNodeID)
		// Debug if blocked
		t.Logf("State History: %v", state.History)
	}

	if state.Status != domain.StatusActive {
		t.Errorf("Expected status Active after resume, got %s", state.Status)
	}
}
