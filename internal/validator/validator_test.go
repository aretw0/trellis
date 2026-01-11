package validator

import (
	"strings"
	"testing"

	"github.com/aretw0/trellis/internal/compiler"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
)

func TestValidateGraph(t *testing.T) {
	// 1. Setup
	parser := compiler.NewParser()

	// 2. Scenario A: Valid Graph
	// start -> a -> b (end)
	loader := inmemory.New(map[string]string{
		"start": `{
			"id": "start",
			"type": "text",
			"transitions": [{"to_node_id": "a"}]
		}`,
		"a": `{
			"id": "a",
			"type": "text",
			"transitions": [{"to_node_id": "b"}]
		}`,
		"b": `{
			"id": "b",
			"type": "text"
		}`,
	})

	if err := ValidateGraph(loader, parser, "start"); err != nil {
		t.Errorf("Scenario A (Valid) failed: %v", err)
	}

	// 3. Scenario B: Broken Link
	// start -> ghost
	// 3. Scenario B: Broken Link
	// start -> ghost
	// We create a NEW loader for the second scenario since memory.New is immutable/static
	loaderBroken := inmemory.New(map[string]string{
		"broken_start": `{
			"id": "broken_start",
			"type": "text",
			"transitions": [{"to_node_id": "ghost_node"}]
		}`,
	})

	err := ValidateGraph(loaderBroken, parser, "broken_start")
	if err == nil {
		t.Error("Scenario B (Broken) should have failed, but got nil")
	} else {
		if !strings.Contains(err.Error(), "Missing node") {
			t.Errorf("Expected 'Missing node' error, got: %v", err)
		}
	}
}
