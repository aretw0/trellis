package validator

import (
	"context"
	"strings"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
)

func TestValidateGraph(t *testing.T) {
	// 1. Setup
	tmpDir := t.TempDir()

	repo, err := loam.Init(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	// 2. Scenario A: Valid Graph
	// start -> a -> b (end)
	validDocs := []core.Document{
		{
			ID: "start.md",
			Content: `---
id: start
type: text
transitions:
  - to: a
---
Start`,
		},
		{
			ID: "a.md",
			Content: `---
id: a
type: text
transitions:
  - to: b
---
Node A`,
		},
		{
			ID: "b.md",
			Content: `---
id: b
type: text
---
End`,
		},
	}

	for _, d := range validDocs {
		if err := repo.Save(ctx, d); err != nil {
			t.Fatal(err)
		}
	}

	if err := ValidateGraph(repo, "start"); err != nil {
		t.Errorf("Scenario A (Valid) failed: %v", err)
	}

	// 3. Scenario B: Broken Link
	// start -> ghost
	brokenDoc := core.Document{
		ID: "broken_start.md",
		Content: `---
id: broken_start
type: text
transitions:
  - to: ghost_node
---
Start`,
	}
	if err := repo.Save(ctx, brokenDoc); err != nil {
		t.Fatal(err)
	}

	err = ValidateGraph(repo, "broken_start")
	if err == nil {
		t.Error("Scenario B (Broken) should have failed, but got nil")
	} else {
		if !strings.Contains(err.Error(), "Missing node") {
			t.Errorf("Expected 'Missing node' error, got: %v", err)
		}
	}
}
