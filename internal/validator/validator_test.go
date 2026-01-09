package validator

import (
	"strings"
	"testing"

	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/compiler"
)

func TestValidateGraph(t *testing.T) {
	// 1. Setup
	parser := compiler.NewParser()
	loader := adapters.NewInMemoryLoader()

	// 2. Scenario A: Valid Graph
	// start -> a -> b (end)
	loader.AddNode("start", []byte(`{
		"id": "start",
		"type": "text",
		"transitions": [{"to_node_id": "a"}]
	}`))
	loader.AddNode("a", []byte(`{
		"id": "a",
		"type": "text",
		"transitions": [{"to_node_id": "b"}]
	}`))
	loader.AddNode("b", []byte(`{
		"id": "b",
		"type": "text"
	}`))

	if err := ValidateGraph(loader, parser, "start"); err != nil {
		t.Errorf("Scenario A (Valid) failed: %v", err)
	}

	// 3. Scenario B: Broken Link
	// start -> ghost
	loader.AddNode("broken_start", []byte(`{
		"id": "broken_start",
		"type": "text",
		"transitions": [{"to_node_id": "ghost_node"}]
	}`))

	err := ValidateGraph(loader, parser, "broken_start")
	if err == nil {
		t.Error("Scenario B (Broken) should have failed, but got nil")
	} else {
		if !strings.Contains(err.Error(), "Missing node") {
			t.Errorf("Expected 'Missing node' error, got: %v", err)
		}
	}
}
