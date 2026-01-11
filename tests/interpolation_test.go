package tests

import (
	"context"
	"testing"

	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/testutils"
	"github.com/aretw0/trellis/pkg/domain"
)

func TestInterpolation(t *testing.T) {
	// 1. Setup Temp Repo
	tmpDir, repo := testutils.SetupTestRepo(t)

	// Since we are using testutils, we don't need loam import directly unless used elsewhere
	// But repo is *git.Repository.
	// We need loam.Init return value if we use it.
	// Oh, I see the original code:
	// repo, err := loam.Init(tmpDir)
	// if err != nil { t.Fatal(err) }

	// Check if `repo` variable usage matches.
	// In original: `repo` is `*git.Repository`.
	// My helper returns `*git.Repository`.
	// Correct.

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

	// 4. Manually seed state with context
	state := domain.NewState("welcome")
	state.Context["username"] = "Alice"

	// 5. Run Render
	// We only care about the Output actions here.
	actions, _, err := eng.Render(context.Background(), state)
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
