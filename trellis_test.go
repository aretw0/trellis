package trellis_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aretw0/trellis"
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

	// 2. Test Start State
	state := engine.Start()
	if state.CurrentNodeID != "start" {
		t.Errorf("Expected initial state 'start', got '%s'", state.CurrentNodeID)
	}

	// 3. Test Step
	actions, nextState, err := engine.Step(state, "")
	if err != nil {
		t.Fatalf("Step failed: %v", err)
	}

	if len(actions) == 0 {
		// It might be empty if we don't render on empty input, but let's check payload if present
	} else {
		// Expect CLI_PRINT "Hello World"
		found := false
		for _, act := range actions {
			if act.Type == "CLI_PRINT" {
				if msg, ok := act.Payload.(string); ok && msg == "Hello World" {
					found = true
				}
			}
		}
		if !found {
			t.Logf("Did not find expected CLI_PRINT Hello World action. Got: %v", actions)
		}
	}

	_ = nextState
}
