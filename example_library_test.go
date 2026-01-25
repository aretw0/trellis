package trellis_test

import (
	"context"
	"fmt"
	"log"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

// ExampleNew_library demonstrates how to use Trellis purely as a Go library,
// injecting an in-memory graph without reading from the filesystem.
func ExampleNew_library() {
	// 1. Define your graph using pure Go structs
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:      "start",
			Type:    "text",
			Content: []byte("Hello from Memory!"),
			Transitions: []domain.Transition{
				{ToNodeID: "finish"},
			},
		},
		domain.Node{
			ID:      "finish",
			Type:    "text",
			Content: []byte("Goodbye."),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Initialize the Engine with the custom loader
	// No file path needed ("") because we are providing a loader.
	eng, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		log.Fatal(err)
	}

	// 3. Start a session
	ctx := context.Background()
	state, err := eng.Start(ctx, "session-mem", nil)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Run the loop (simplified for example)
	for {
		// Render current state
		actions, terminal, _ := eng.Render(ctx, state)

		// Print content
		for _, act := range actions {
			if act.Type == domain.ActionRenderContent {
				fmt.Println(act.Payload)
			}
		}

		if terminal {
			break
		}

		// Move to next state
		state, err = eng.Navigate(ctx, state, nil)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Output:
	// Hello from Memory!
	// Goodbye.
}
