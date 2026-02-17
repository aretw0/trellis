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

func TestBuilder_ToolFlow(t *testing.T) {
	b := New()

	b.Add("start").
		Text("Executing tool...").
		Go("run_tool")

	b.Add("run_tool").
		Do("http_get", map[string]any{"url": "https://api.example.com"}).
		Undo("http_delete", map[string]any{"id": "123"}).
		Go("end")

	b.Add("end").
		Text("Done!").
		Terminal()

	loader, err := b.Build()
	if err != nil {
		t.Fatalf("Build() failed: %v", err)
	}

	// Verify Tool Node
	toolNodeBytes, err := loader.GetNode("run_tool")
	if err != nil {
		t.Fatalf("GetNode('run_tool') failed: %v", err)
	}

	var toolNode domain.Node
	if err := json.Unmarshal(toolNodeBytes, &toolNode); err != nil {
		t.Fatalf("Failed to unmarshal tool node: %v", err)
	}

	if toolNode.Type != domain.NodeTypeTool {
		t.Errorf("Expected node type 'tool', got '%s'", toolNode.Type)
	}
	if toolNode.Do == nil || toolNode.Do.Name != "http_get" {
		t.Errorf("Expected tool 'http_get', got %+v", toolNode.Do)
	}
	if toolNode.Undo == nil || toolNode.Undo.Name != "http_delete" {
		t.Errorf("Expected undo tool 'http_delete', got %+v", toolNode.Undo)
	}

	// Verify Terminal Node
	endNodeBytes, err := loader.GetNode("end")
	if err != nil {
		t.Fatalf("GetNode('end') failed: %v", err)
	}
	var endNode domain.Node
	if err := json.Unmarshal(endNodeBytes, &endNode); err != nil {
		t.Fatalf("Failed to unmarshal end node: %v", err)
	}
	if len(endNode.Transitions) != 0 {
		t.Errorf("Expected 0 transitions for terminal node, got %d", len(endNode.Transitions))
	}
}
