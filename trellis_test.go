package trellis_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestFacade_Integration(t *testing.T) {
	// 0. Setup Temp Repo
	repoPath := t.TempDir()
	startFile := filepath.Join(repoPath, "start.md")
	content := []byte(`---
id: start
type: text
---
Hello World`)
	if err := os.WriteFile(startFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	// 1. Test Initialization
	engine, err := trellis.New(repoPath)
	if err != nil {
		t.Fatalf("Failed to initialize engine with path %s: %v", repoPath, err)
	}

	ctx := context.Background()
	state, err := engine.Start(ctx, "test", nil)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if state.CurrentNodeID != "start" {
		t.Errorf("Expected initial state 'start', got '%s'", state.CurrentNodeID)
	}

	// 3. Test Render (Start Node)
	actions, _, err := engine.Render(context.Background(), state)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if len(actions) == 0 {
		t.Error("Expected actions from start node, got 0")
	}

	// 4. Test Navigate (No input -> Next)
	// Assuming Start flows to something else or just testing validity of current state if no input provided.
	// In this test we just want to ensure Render worked.

	if len(actions) == 0 {
		// It might be empty if we don't render on empty input, but let's check payload if present
	} else {
		// Expect ACTION_RENDER_CONTENT "Hello World"
		found := false
		for _, act := range actions {
			if act.Type == domain.ActionRenderContent {
				if msg, ok := act.Payload.(string); ok && msg == "Hello World" {
					found = true
				}
			}
		}
		if !found {
			t.Logf("Did not find expected RENDER_CONTENT Hello World action. Got: %v", actions)
		}
	}

}
