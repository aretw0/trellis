package main

import (
	"context"
	"log"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

func main() {
	// 1. Define the Graph using Go Structs
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:      "start",
			Type:    "question",
			Content: []byte("Welcome to the **Trellis TUI Demo**, {{ .User }}!\nDo you see this in *rich text*? [yes] [no]"),
			Transitions: []domain.Transition{
				{ToNodeID: "yes", Condition: "input == 'yes'"},
				{ToNodeID: "no"},
			},
		},
		domain.Node{
			ID:      "yes",
			Type:    "text",
			Wait:    true,
			Content: []byte("## Great!\n\nThat means `glamour` is working."),
		},
		domain.Node{
			ID:      "no",
			Type:    "text",
			Wait:    true,
			Content: []byte("## Oops.\n\nSomething is wrong with the renderer."),
		},
	)
	if err != nil {
		log.Fatalf("Failed to create loader: %v", err)
	}

	// 2. Initialize Engine
	eng, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		log.Fatalf("Failed to init engine: %v", err)
	}

	// 3. Configure Runner with TUI
	r := runner.NewRunner(
		runner.WithRenderer(tui.NewRenderer()),
	)

	// 4. Create initial state and seed data
	ctx := context.Background()
	state, err := eng.Start(ctx, "hello-world", nil)
	if err != nil {
		panic(err)
	}
	state.Context["User"] = "World" // Seed Data (Verification of Interpolation)

	// 5. Run!
	if _, err := r.Run(context.Background(), eng, state); err != nil {
		log.Fatalf("Error running: %v", err)
	}
}
