package runtime_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestEngine_Step(t *testing.T) {
	// Setup
	loader := adapters.NewInMemoryLoader()
	engine := runtime.NewEngine(loader)

	// Node 1: Start
	startNode := domain.Node{
		ID:      "start",
		Type:    "question",
		Content: []byte("Start Node"),
		Transitions: []domain.Transition{
			{ToNodeID: "middle", Condition: "input == 'yes'"},
		},
	}
	data1, _ := json.Marshal(startNode)
	loader.AddNode("start", data1)

	// Node 2: Middle
	middleNode := domain.Node{
		ID:      "middle",
		Type:    "text",
		Content: []byte("Middle Node"),
		Transitions: []domain.Transition{
			{ToNodeID: "end", Condition: ""}, // Always
		},
	}
	data2, _ := json.Marshal(middleNode)
	loader.AddNode("middle", data2)

	// Node 3: End
	endNode := domain.Node{
		ID:          "end",
		Type:        "text",
		Content:     []byte("End Node"),
		Transitions: []domain.Transition{},
	}
	data3, _ := json.Marshal(endNode)
	loader.AddNode("end", data3)

	t.Run("Initial Render", func(t *testing.T) {
		state := domain.NewState("start")
		actions, nextState, err := engine.Step(context.Background(), state, "")
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}

		if len(actions) != 1 {
			t.Errorf("Expected 1 action, got %d", len(actions))
		}
		if actions[0].Payload != "Start Node" {
			t.Errorf("Unexpected payload: %v", actions[0].Payload)
		}
		if nextState.CurrentNodeID != "start" {
			t.Errorf("Expected to return state 'start', got '%s'", nextState.CurrentNodeID)
		}
	})

	t.Run("Transition Matching", func(t *testing.T) {
		state := domain.NewState("start")
		// Simulate input
		actions, nextState, err := engine.Step(context.Background(), state, "YeS") // Mixed case
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}

		// Node logic should NOT run when processing input
		if len(actions) != 0 {
			t.Errorf("Expected 0 actions (no re-render), got %d", len(actions))
		}

		if nextState.CurrentNodeID != "middle" {
			t.Errorf("Expected transition to 'middle', got '%s'", nextState.CurrentNodeID)
		}
	})

	t.Run("No Transition Match", func(t *testing.T) {
		state := domain.NewState("start")
		_, nextState, err := engine.Step(context.Background(), state, "no")
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}

		if nextState.CurrentNodeID != "start" {
			t.Errorf("Expected to stay in 'start', got '%s'", nextState.CurrentNodeID)
		}
	})

	t.Run("Default Transition", func(t *testing.T) {
		state := domain.NewState("middle")
		_, nextState, err := engine.Step(context.Background(), state, "") // Empty input for auto transition
		if err != nil {
			t.Fatalf("Step failed: %v", err)
		}

		if nextState.CurrentNodeID != "end" {
			t.Errorf("Expected auto transition to 'end', got '%s'", nextState.CurrentNodeID)
		}
	})
}
