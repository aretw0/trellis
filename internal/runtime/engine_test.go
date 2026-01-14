package runtime_test

import (
	"context"
	"testing"

	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestEngine_RenderAndNavigate(t *testing.T) {
	// Setup
	// Setup Declartive Nodes
	startNode := domain.Node{
		ID:      "start",
		Type:    domain.NodeTypeQuestion,
		Content: []byte("Start Node"),
		Transitions: []domain.Transition{
			{ToNodeID: "middle", Condition: "input == 'yes'"},
		},
	}
	middleNode := domain.Node{
		ID:      "middle",
		Type:    domain.NodeTypeText,
		Content: []byte("Middle Node"),
		Transitions: []domain.Transition{
			{ToNodeID: "end", Condition: ""}, // Always
		},
	}
	endNode := domain.Node{
		ID:          "end",
		Type:        domain.NodeTypeText,
		Content:     []byte("End Node"),
		Transitions: []domain.Transition{},
	}

	loader, _ := inmemory.NewFromNodes(startNode, middleNode, endNode)
	engine := runtime.NewEngine(loader, nil, nil)

	t.Run("Initial Render", func(t *testing.T) {
		state := domain.NewState("start")
		// 1. Initial Render (Start)
		actions, _, err := engine.Render(context.Background(), state)
		if err != nil {
			t.Fatalf("Render failed (start): %v", err)
		}

		if len(actions) != 2 {
			t.Errorf("Expected 2 actions, got %d", len(actions))
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

	loader, _ := inmemory.NewFromNodes(node)
	engine := runtime.NewEngine(loader, nil, nil)

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

func TestEngine_Interpolation(t *testing.T) {
	node := domain.Node{
		ID:      "tmpl",
		Type:    domain.NodeTypeText,
		Content: []byte("Hello {{ .Name }}! VIP: {{ if .VIP }}Yes{{ else }}No{{ end }}"),
	}
	loader, _ := inmemory.NewFromNodes(node)
	engine := runtime.NewEngine(loader, nil, nil) // Uses DefaultInterpolator (text/template)

	t.Run("Standard Template", func(t *testing.T) {
		state := domain.NewState("tmpl")
		state.Context["Name"] = "Alice"
		state.Context["VIP"] = true

		actions, _, err := engine.Render(context.Background(), state)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if len(actions) != 1 {
			t.Fatal("Expected 1 action")
		}
		if actions[0].Payload != "Hello Alice! VIP: Yes" {
			t.Errorf("Unexpected output: %s", actions[0].Payload)
		}
	})

	t.Run("Missing Variable", func(t *testing.T) {
		state := domain.NewState("tmpl")
		state.Context["VIP"] = false
		// Name is missing

		actions, _, err := engine.Render(context.Background(), state)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		expected := "Hello ! VIP: No"
		if actions[0].Payload != expected {
			t.Errorf("Expected '%s', got '%s'", expected, actions[0].Payload)
		}
	})
}

func TestEngine_LegacyInterpolation(t *testing.T) {
	node := domain.Node{
		ID:      "legacy",
		Type:    domain.NodeTypeText,
		Content: []byte("Hello {{ Name }}"), // Old syntax
	}
	loader, _ := inmemory.NewFromNodes(node)

	// Inject LegacyInterpolator
	engine := runtime.NewEngine(loader, nil, runtime.LegacyInterpolator)

	state := domain.NewState("legacy")
	state.Context["Name"] = "Bob"

	actions, _, err := engine.Render(context.Background(), state)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if actions[0].Payload != "Hello Bob" {
		t.Errorf("Expected 'Hello Bob', got '%s'", actions[0].Payload)
	}
}

func TestEngine_DataBinding(t *testing.T) {
	node := domain.Node{
		ID:      "ask_name",
		Type:    domain.NodeTypeQuestion,
		Content: []byte("What is your name?"),
		SaveTo:  "user_name",
		Transitions: []domain.Transition{
			{ToNodeID: "greet", Condition: ""},
		},
	}
	greetNode := domain.Node{
		ID:      "greet",
		Type:    domain.NodeTypeText,
		Content: []byte("Hello {{ .user_name }}"),
	}

	loader, _ := inmemory.NewFromNodes(node, greetNode)
	engine := runtime.NewEngine(loader, nil, nil) // Default Interpolator

	state := domain.NewState("ask_name")

	// Navigate with Input "Alice"
	nextState, err := engine.Navigate(context.Background(), state, "Alice")
	if err != nil {
		t.Fatalf("Navigate failed: %v", err)
	}

	// Verify Context update
	if val, ok := nextState.Context["user_name"]; !ok {
		t.Error("Expected 'user_name' to be in Context, but it was missing")
	} else if val != "Alice" {
		t.Errorf("Expected 'user_name' to be 'Alice', got '%v'", val)
	}

	// Verify Transition
	if nextState.CurrentNodeID != "greet" {
		t.Errorf("Expected transition to 'greet', got '%s'", nextState.CurrentNodeID)
	}

	// Render next node to verify interpolation works with the bound data
	actions, _, err := engine.Render(context.Background(), nextState)
	if err != nil {
		t.Fatalf("Render failed (greet): %v", err)
	}
	if len(actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(actions))
	} else if actions[0].Payload != "Hello Alice" {
		t.Errorf("Expected 'Hello Alice', got '%s'", actions[0].Payload)
	}
}

func TestEngine_Namespacing(t *testing.T) {
	node := domain.Node{
		ID:      "node1",
		Type:    domain.NodeTypeText,
		Content: []byte("System says: {{ .sys.message }}"),
		Transitions: []domain.Transition{
			{ToNodeID: "node2"},
		},
	}
	// Node trying to write to sys
	nodeViolation := domain.Node{
		ID:      "violation",
		Type:    domain.NodeTypeQuestion,
		SaveTo:  "sys.hacked",
		Content: []byte("Try to hack"),
	}

	loader, _ := inmemory.NewFromNodes(node, nodeViolation)
	engine := runtime.NewEngine(loader, nil, nil)

	t.Run("Read System Context", func(t *testing.T) {
		state := domain.NewState("node1")
		state.SystemContext["message"] = "Secure"

		actions, _, err := engine.Render(context.Background(), state)
		if err != nil {
			t.Fatalf("Render failed: %v", err)
		}
		if actions[0].Payload != "System says: Secure" {
			t.Errorf("Expected 'System says: Secure', got '%s'", actions[0].Payload)
		}
	})

	t.Run("Block System Write", func(t *testing.T) {
		state := domain.NewState("violation")
		_, err := engine.Navigate(context.Background(), state, "malicious_input")
		if err == nil {
			t.Fatal("Expected error when saving to 'sys.*', got nil")
		}
		if err.Error() != "security violation: cannot save to reserved namespace 'sys' in node violation" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}
