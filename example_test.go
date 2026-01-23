package trellis_test

import (
	"context"
	"fmt"
	"log"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/inmemory"
	"github.com/aretw0/trellis/pkg/domain"
)

// ExampleNew_memory demonstrates how to use the Engine with an in-memory graph definition.
// This is useful for testing, embedded scenarios, or when you don't want to rely on the file system.
func ExampleNew_memory() {
	// 1. Define your graph using helper NewFromNodes for clean, type-safe construction.
	loader, err := inmemory.NewFromNodes(
		domain.Node{
			ID:      "start",
			Type:    "question",
			Content: []byte("Hello! Do you want to proceed? [yes] [no]"),
			Transitions: []domain.Transition{
				{ToNodeID: "yes", Condition: "input == 'yes'"},
				{ToNodeID: "no"},
			},
		},
		domain.Node{
			ID:      "yes",
			Type:    "text",
			Content: []byte("Great! You moved forward."),
		},
		domain.Node{
			ID:      "no",
			Type:    "text",
			Content: []byte("Okay, bye."),
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Initialize Trellis with the custom loader
	// Note: We leave path empty ("") because we are providing a loader.
	engine, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		log.Fatal(err)
	}

	// 4. Start the flow
	ctx := context.Background()
	state, err := engine.Start(ctx, "test", nil)
	if err != nil {
		panic(err)
	}

	// 5. Navigate (Input: "yes")
	// "start" -> (input: yes) -> "yes"
	// Note: Example previously captured actions from Step.
	// Render to show we can get actions, then Navigate.
	actions, _, err := engine.Render(ctx, state)
	if err != nil {
		log.Fatal(err)
	}
	// Verify actions from start node (optional in example, but good for completeness)

	nextState, err := engine.Navigate(ctx, state, "yes")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Current Node: %s\n", nextState.CurrentNodeID)
	for _, action := range actions {
		fmt.Printf("Action: %s\n", action.Type)
	}
	// Output:
	// Current Node: yes
	// Action: RENDER_CONTENT
	// Action: REQUEST_INPUT
}
