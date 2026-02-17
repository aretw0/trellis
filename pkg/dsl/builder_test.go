package dsl

import (
	"encoding/json"
	"testing"

	"github.com/aretw0/trellis/pkg/domain"
)

func TestBuilder_SimpleFlow(t *testing.T) {
	// 1. Build the graph using DSL
	b := New()

	b.Add("start").
		Text("Hello, DSL!").
		Go("ask_name")

	b.Add("ask_name").
		Question("What is your name?").
		Input("text").
		SaveTo("user_name").
		Go("greet")

	b.Add("greet").
		Text("Nice to meet you, {{user_name}}!").
		Go("end")

	b.Add("end").
		Text("Goodbye!")

	// 2. Compile to Loader
	loader, err := b.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// 3. Verify specific nodes
	// Check Start Node
	startNodeBytes, err := loader.GetNode("start")
	if err != nil {
		t.Fatalf("GetNode('start') failed: %v", err)
	}

	var startNode domain.Node
	if err := json.Unmarshal(startNodeBytes, &startNode); err != nil {
		t.Fatalf("Failed to unmarshal start node: %v", err)
	}

	if startNode.Type != domain.NodeTypeText {
		t.Errorf("Expected start node type 'text', got '%s'", startNode.Type)
	}
	if string(startNode.Content) != "Hello, DSL!" {
		t.Errorf("Expected content 'Hello, DSL!', got '%s'", startNode.Content)
	}
	if len(startNode.Transitions) != 1 {
		t.Fatalf("Expected 1 transition, got %d", len(startNode.Transitions))
	}
	if startNode.Transitions[0].ToNodeID != "ask_name" {
		t.Errorf("Expected transition to 'ask_name', got '%s'", startNode.Transitions[0].ToNodeID)
	}

	// Check Question Node
	askNodeBytes, err := loader.GetNode("ask_name")
	if err != nil {
		t.Fatalf("GetNode('ask_name') failed: %v", err)
	}
	var askNode domain.Node
	if err := json.Unmarshal(askNodeBytes, &askNode); err != nil {
		t.Fatalf("Failed to unmarshal ask node: %v", err)
	}

	if askNode.Type != domain.NodeTypeQuestion {
		t.Errorf("Expected ask node type 'question', got '%s'", askNode.Type)
	}
	if !askNode.Wait {
		t.Error("Expected Wait=true for question node")
	}
	if askNode.SaveTo != "user_name" {
		t.Errorf("Expected SaveTo='user_name', got '%s'", askNode.SaveTo)
	}

	// Check ListNodes
	nodes, err := loader.ListNodes()
	if err != nil {
		t.Fatalf("ListNodes() failed: %v", err)
	}
	if len(nodes) != 4 {
		t.Errorf("Expected 4 nodes, got %d", len(nodes))
	}
}
