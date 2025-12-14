package tests

import (
	"context"
	"testing"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestInterpolation(t *testing.T) {
	// 1. Setup Temp Repo
	// 1. Setup Temp Repo
	tmpDir := t.TempDir()

	repo, err := loam.Init(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// 2. Create node with variable
	content := `---
id: welcome
type: text
---
Hello, {{ username }}!
`
	if err := repo.Save(context.Background(), core.Document{ID: "welcome.md", Content: content}); err != nil {
		t.Fatal(err)
	}

	// 3. Init Engine
	eng, err := trellis.New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to init engine: %v", err)
	}

	// 4. Manually seed state with memory
	state := domain.NewState("welcome")
	state.Memory["username"] = "Alice"

	// 5. Run Step
	actions, _, err := eng.Step(context.Background(), state, "")
	if err != nil {
		t.Fatal(err)
	}

	// 6. Verify Output
	if len(actions) == 0 {
		t.Fatal("Expected render action")
	}
	output := actions[0].Payload.(string)
	expected := "Hello, Alice!\n"
	if output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, output)
	}
}
