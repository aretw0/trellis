package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/domain"
)

func main() {
	// 1. Define the Graph using Go Structs (Clean & Type-Safe)
	loader, err := memory.NewFromNodes(
		domain.Node{
			ID:      "start",
			Type:    "question",
			Content: []byte("Welcome to the Pure Go Trellis Demo!\nDo you want to see how easy this is? [yes] [no]"),
			Transitions: []domain.Transition{
				{ToNodeID: "yes", Condition: "input == 'yes'"},
				{ToNodeID: "no"},
			},
		},
		domain.Node{
			ID:      "yes",
			Type:    "text",
			Content: []byte("Correct! We just in-memory structs and now we have a state machine."),
		},
		domain.Node{
			ID:      "no",
			Type:    "text",
			Content: []byte("Oh well. It is easy though."),
		},
	)
	if err != nil {
		log.Fatalf("Failed to create loader: %v", err)
	}

	// 2. Initialize Engine with MemoryLoader
	eng, err := trellis.New("", trellis.WithLoader(loader))
	if err != nil {
		log.Fatalf("Failed to init engine: %v", err)
	}

	// 3. Start the State
	state := eng.Start()
	ctx := context.Background()

	// 4. Simple REPL Loop
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("--- Trellis Memory Demo ---")

	for !state.Terminated {
		// Run a Step (with empty input to get the initial content)
		// Note: In a real app we would manage "visiting" vs "submitting" better.
		// Here we just simulate: Get Content -> Wait for Input -> Submit Input.

		// First, get content for current node (blind step with no input)
		// Wait, Step logic: output content if input is empty? Yes.
		actions, _, err := eng.Step(ctx, state, "")
		if err != nil {
			log.Fatalf("Error rendering: %v", err)
		}

		// Print Content
		for _, act := range actions {
			if act.Type == domain.ActionRenderContent {
				fmt.Println(act.Payload)
			}
		}

		// Read Input
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Submit Input to Transition
		_, nextState, err := eng.Step(ctx, state, input)
		if err != nil {
			log.Fatalf("Error transitioning: %v", err)
		}

		// Update State
		state = nextState
		fmt.Printf("[Debug] Transitioned to: %s\n", state.CurrentNodeID)
	}

	fmt.Println("Flow Terminated.")
}
