package runtime_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestEngine_RenderAndNavigate(t *testing.T) {
	// Setup
	loader := adapters.NewInMemoryLoader()
	engine := runtime.NewEngine(loader, nil)

	// Node 1: Start
	startNode := domain.Node{
		ID:      "start",
		Type:    domain.NodeTypeQuestion,
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
		Type:    domain.NodeTypeText,
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
		Type:        domain.NodeTypeText,
		Content:     []byte("End Node"),
		Transitions: []domain.Transition{},
	}
	data3, _ := json.Marshal(endNode)
	loader.AddNode("end", data3)

	t.Run("Initial Render", func(t *testing.T) {
		state := domain.NewState("start")
		// 1. Initial Render (Start)
		actions, _, err := engine.Render(context.Background(), state)
		if err != nil {
			t.Fatalf("Render failed (start): %v", err)
		}

		if len(actions) != 1 {
			t.Errorf("Expected 1 action, got %d", len(actions))
		}
		if actions[0].Payload != "Start Node" {
			t.Errorf("Unexpected payload: %v", actions[0].Payload)
		}
		// Render doesn't change state, so we check state locally or skip usage of nextState here
		// Actually the original test asserted on nextState.CurrentNodeID
		// In Render only, state doesn't change.
		if state.CurrentNodeID != "start" {
			t.Errorf("Expected state to remain 'start', got '%s'", state.CurrentNodeID)
		}
	})

	t.Run("Transition Matching", func(t *testing.T) {
		state := domain.NewState("start")
		// Simulate input
		// 1. Render (Start)
		actions, _, err := engine.Render(context.Background(), state)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		// ... check actions if needed

		// 2. Navigate (Input: "YeS")
		nextState, err := engine.Navigate(context.Background(), state, "YeS") // Mixed case
		if err != nil {
			t.Fatalf("Navigate failed: %v", err)
		}

		// Check actions from Render (should be present now)
		if len(actions) == 0 {
			t.Errorf("Expected actions from Render, got 0")
		}

		if nextState.CurrentNodeID != "middle" {
			t.Errorf("Expected transition to 'middle', got '%s'", nextState.CurrentNodeID)
		}
	})

	t.Run("No Transition Match", func(t *testing.T) {
		state := domain.NewState("start")
		// Navigate (Input: "no")
		nextState, err := engine.Navigate(context.Background(), state, "no")
		if err != nil {
			t.Fatalf("Navigate failed: %v", err)
		}

		if nextState.CurrentNodeID != "start" {
			t.Errorf("Expected to stay in 'start', got '%s'", nextState.CurrentNodeID)
		}
	})

	t.Run("Default Transition", func(t *testing.T) {
		state := domain.NewState("middle")
		// Navigate (Input: "")
		nextState, err := engine.Navigate(context.Background(), state, "") // Empty input for auto transition
		if err != nil {
			t.Fatalf("Navigate failed: %v", err)
		}

		if nextState.CurrentNodeID != "end" {
			t.Errorf("Expected auto transition to 'end', got '%s'", nextState.CurrentNodeID)
		}
	})
}

func TestEngine_Render_Inputs(t *testing.T) {
	// Setup
	loader := adapters.NewInMemoryLoader()
	engine := runtime.NewEngine(loader, nil)

	// Node 1: Input Node
	node := domain.Node{
		ID:           "input",
		Type:         domain.NodeTypeQuestion,
		Content:      []byte("Question content"),
		Transitions:  []domain.Transition{},
		InputType:    "choice",
		InputOptions: []string{"A", "B"},
		InputDefault: "A",
	}
	data, _ := json.Marshal(node)
	loader.AddNode("input", data)

	// Render
	state := domain.NewState("input")
	actions, _, err := engine.Render(context.Background(), state)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	// Assert
	if len(actions) != 2 {
		t.Fatalf("Expected 2 actions, got %d", len(actions))
	}

	// Check text action
	if actions[0].Type != domain.ActionRenderContent {
		t.Errorf("Expected first action to be RENDER_CONTENT, got %s", actions[0].Type)
	}

	// Check input action
	inputAct := actions[1]
	if inputAct.Type != domain.ActionRequestInput {
		t.Errorf("Expected second action to be REQUEST_INPUT, got %s", inputAct.Type)
	}

	req, ok := inputAct.Payload.(domain.InputRequest)
	if !ok {
		t.Fatalf("Payload is NOT InputRequest, got %T", inputAct.Payload)
	}

	if req.Type != domain.InputChoice {
		t.Errorf("Expected input type 'choice', got '%s'", req.Type)
	}
	if len(req.Options) != 2 {
		t.Errorf("Expected 2 options, got %d", len(req.Options))
	}
}
