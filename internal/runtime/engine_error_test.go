package runtime_test

import (
	"context"
	"strings"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestEngine_OnError_Transition(t *testing.T) {
	// Scenario: Tool returns error, Engine should transition to OnError node ignoring SaveTo

	// Node 1: Tool Node with OnError
	toolNode := domain.Node{
		ID:       "step1",
		Type:     domain.NodeTypeTool,
		ToolCall: &domain.ToolCall{ID: "t1", Name: "risky_tool"},
		SaveTo:   "result_data", // Should NOT be written on error
		OnError:  "recovery",
		Transitions: []domain.Transition{
			{ToNodeID: "success"},
		},
	}

	// Node 2: Recovery Node
	recoveryNode := domain.Node{
		ID:      "recovery",
		Type:    domain.NodeTypeText,
		Content: []byte("Recovery Mode"),
	}

	// Node 3: Success Node
	successNode := domain.Node{
		ID:      "success",
		Type:    domain.NodeTypeText,
		Content: []byte("Success Mode"),
	}

	loader, _ := inmemory.NewFromNodes(toolNode, recoveryNode, successNode)
	engine := runtime.NewEngine(loader, nil, nil)

	// A. Start at step1
	state := domain.NewState("step1")
	state.Status = domain.StatusWaitingForTool
	state.PendingToolCall = "t1"
	// Set some context to verify it persists
	state.Context["pre_existing"] = "safe"

	// B. Simulate Host returning Error
	toolResult := domain.ToolResult{
		ID:      "t1",
		IsError: true,
		Result:  "Connection Failed", // Error details
	}

	// C. Navigate with Error Result
	newState, err := engine.Navigate(context.Background(), state, toolResult)
	if err != nil {
		t.Fatalf("Navigate failed: %v", err)
	}

	// D. Verify Transition to Recovery
	if newState.CurrentNodeID != "recovery" {
		t.Fatalf("Expected transition to 'recovery', got '%s'", newState.CurrentNodeID)
	}

	// E. Verify SaveTo did NOT happen (context should not contain 'result_data')
	if _, ok := newState.Context["result_data"]; ok {
		t.Errorf("Expected 'result_data' to be absent (skipped SaveTo), but found it")
	}

	// F. Verify Context Persistence
	if val, ok := newState.Context["pre_existing"]; !ok || val != "safe" {
		t.Errorf("Expected 'pre_existing' context to be preserved")
	}
}

func TestEngine_OnError_Missing_FailFast(t *testing.T) {
	// Scenario: Tool returns error, but NO OnError is defined. Should FAIL FAST.

	toolNode := domain.Node{
		ID:       "step1",
		Type:     domain.NodeTypeTool,
		ToolCall: &domain.ToolCall{ID: "t1", Name: "risky_tool"},
		SaveTo:   "result_data",
		// OnError is MISSING
		Transitions: []domain.Transition{
			{ToNodeID: "next"},
		},
	}
	nextNode := domain.Node{
		ID:   "next",
		Type: domain.NodeTypeText,
	}

	loader, _ := inmemory.NewFromNodes(toolNode, nextNode)
	engine := runtime.NewEngine(loader, nil, nil)

	state := domain.NewState("step1")
	state.Status = domain.StatusWaitingForTool
	state.PendingToolCall = "t1"

	// Error Result
	toolResult := domain.ToolResult{
		ID:      "t1",
		IsError: true,
		Result:  "Critical Failure",
	}

	// Navigate
	_, err := engine.Navigate(context.Background(), state, toolResult)
	if err == nil {
		t.Fatal("Expected Navigate to fail due to missing on_error, but it succeeded")
	}

	// Verify Error Type and Message
	if !strings.Contains(err.Error(), "Execution halted") {
		t.Errorf("Expected helpful debug message, got: %v", err)
	}
	// Note: We can't check specific type easily here if we don't export/import it correctly in test package or use errors.As
	// Since test is in runtime_test package but UnhandledToolError is in runtime, we need to export it (which we did).
}
