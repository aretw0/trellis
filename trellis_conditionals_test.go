package trellis_test

import (
	"context"
	"testing"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/testutils"
)

func TestDelegatedLogic(t *testing.T) {
	// 1. Setup Temp Repo
	// 1. Setup Temp Repo
	tmpDir, repo := testutils.SetupTestRepo(t)

	// 2. Create nodes with condition
	// start -> (condition: secret_ok) -> secret_room
	// start -> (default) -> public_room
	startContent := `---
id: start
type: text
transitions:
  - to: secret_room
    condition: is_vip
---
Welcome.
`
	ctx := context.Background()
	if err := repo.Save(ctx, core.Document{ID: "start.md", Content: startContent}); err != nil {
		t.Fatal(err)
	}

	if err := repo.Save(ctx, core.Document{ID: "secret_room.md", Content: "---\nid: secret_room\n---\nVIP Area"}); err != nil {
		t.Fatal(err)
	}
	// public_room not needed for this flow anymore

	// 3. Define Evaluator that returns true for "is_vip" ONLY if input is "password"
	evaluator := func(ctx context.Context, condition string, input string) (bool, error) {
		if condition == "is_vip" {
			return input == "password", nil
		}
		return false, nil
	}

	// 4. Init Engine with Evaluator
	eng, err := trellis.New(tmpDir, trellis.WithConditionEvaluator(evaluator))
	if err != nil {
		t.Fatalf("Failed to init engine: %v", err)
	}

	// 5. Test Case 1: Wrong Password -> Should STAY at start
	state := eng.Start()
	// Step 1: Start node (Render)
	_, _, err = eng.Render(context.Background(), state)
	if err != nil {
		t.Fatalf("Failed to render start: %v", err)
	}
	if state.CurrentNodeID != "start" {
		t.Errorf("Expected start, got %s", state.CurrentNodeID)
	}

	// Step 2: Input "wrong"
	nextState, err := eng.Navigate(context.Background(), state, "wrong")
	if err != nil {
		t.Fatal(err)
	}

	// EXPECTATION: Evaluator returns false. No other transitions. Stays at start.
	if nextState.CurrentNodeID != "start" {
		t.Errorf("Expected stay at start for wrong password, got %s", nextState.CurrentNodeID)
	}

	// 6. Test Case 2: Right Password -> Should go to secret_room
	// Reuse state (it is still at start)

	// Step 3: Input "password" -> Navigate
	nextState, err = eng.Navigate(context.Background(), state, "password")
	if err != nil {
		t.Fatal(err)
	}

	if nextState.CurrentNodeID != "secret_room" {
		t.Fatalf("FAILED: Expected secret_room for correct password, got %s", nextState.CurrentNodeID)
	}
}
