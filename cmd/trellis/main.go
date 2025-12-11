package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/domain"
)

func main() {
	// 1. Setup Loam Repository
	repo, err := loam.Init(".", loam.WithVersioning(false))
	if err != nil {
		fmt.Printf("Failed to init loam: %v\n", err)
		os.Exit(1)
	}

	// 2. Wrap with Typed Repository
	typedRepo := loam.NewTyped[adapters.NodeMetadata](repo)

	// 3. Setup Adapter
	loader := adapters.NewLoamLoader(typedRepo)

	// 4. Seed Data (using Typed Repo)
	ctx := context.TODO()

	// Node 1: Start
	node1Meta := adapters.NodeMetadata{
		ID:   "start",
		Type: "question",
		Transitions: []domain.Transition{
			{ToNodeID: "end", Condition: "input == 'go'"},
		},
	}

	// Save start (Loam defaults to .md)
	err = typedRepo.Save(ctx, &loam.DocumentModel[adapters.NodeMetadata]{
		ID:      "start",
		Content: "Welcome to Trellis! Type 'go' to continue.",
		Data:    node1Meta,
	})
	if err != nil {
		fmt.Printf("Save start failed: %v\n", err)
	}

	// Node 2: End
	node2Meta := adapters.NodeMetadata{
		ID:          "end",
		Type:        "text",
		Transitions: []domain.Transition{},
	}

	// Save end (Loam defaults to .md)
	err = typedRepo.Save(ctx, &loam.DocumentModel[adapters.NodeMetadata]{
		ID:      "end",
		Content: "You have reached the end. Goodbye!",
		Data:    node2Meta,
	})
	if err != nil {
		fmt.Printf("Save end failed: %v\n", err)
	}

	// 5. Setup Core (hex type)
	engine := runtime.NewEngine(loader)
	state := domain.NewState("start")

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("--- Trellis CLI (Bootstrapped) ---")

	// Simple Event Loop
	for {
		var input string

		// Run Step
		actions, nextState, err := engine.Step(state, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}

		// Dispatch Actions
		for _, act := range actions {
			if act.Type == "CLI_PRINT" {
				fmt.Println(act.Payload)
			}
		}

		// Exit condition for demo (Check AFTER render)
		if nextState.CurrentNodeID == "end" {
			break
		}

		// If state didn't change, we need input to proceed (or we are stuck)
		if nextState.CurrentNodeID == state.CurrentNodeID {
			fmt.Print("> ")
			text, _ := reader.ReadString('\n')
			input = strings.TrimSpace(text)

			// Check for explicit exit command
			if input == "exit" || input == "quit" {
				fmt.Println("Bye!")
				break
			}

			// Run Step again with input
			actions, nextState, err = engine.Step(state, input)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				break
			}

			// Dispatch any actions from the input-triggered step
			for _, act := range actions {
				if act.Type == "CLI_PRINT" {
					fmt.Println(act.Payload)
				}
			}

			// Update state after input processing
			state = nextState
		} else {
			// We moved to a new state automatically (e.g. forced transition)
			state = nextState
		}
	}
}
