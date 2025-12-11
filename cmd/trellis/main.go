package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aretw0/loam"
	"github.com/aretw0/loam/pkg/core"
	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/domain"
)

func main() {
	// 1. Setup Loam (Repository & Service)
	repo, err := loam.Init(".", loam.WithVersioning(false)) // Use current directory as data store
	if err != nil {
		fmt.Printf("Failed to init loam: %v\n", err)
		os.Exit(1)
	}
	svc := core.NewService(repo)

	// 2. Setup Adapter (driven port)
	// We use LoamLoader now.
	loader := adapters.NewLoamLoader(svc)

	// Seed some data (using Loam Service to persist)
	// We only seed if we want to ensure data exists for the demo.
	node1 := domain.Node{
		ID:      "start",
		Type:    "question",
		Content: []byte("Welcome to Trellis! Type 'go' to continue."),
		Transitions: []domain.Transition{
			{ToNodeID: "end", Condition: "input == 'go'"},
		},
	}

	// Save start.json
	ctx := context.TODO()
	tx, err := svc.Begin(ctx)
	if err != nil {
		fmt.Printf("Tx begin failed: %v\n", err)
		os.Exit(1)
	}
	// We need to construct a core.Document
	err = tx.Save(ctx, core.Document{
		ID:      node1.ID + ".json",
		Content: string(node1.Content),
		Metadata: core.Metadata{
			"ID":          node1.ID,
			"Type":        node1.Type,
			"Transitions": node1.Transitions,
		},
	})
	if err != nil {
		fmt.Printf("Save start failed: %v\n", err)
	}

	node2 := domain.Node{
		ID:          "end",
		Type:        "text",
		Content:     []byte("You have reached the end. Goodbye!"),
		Transitions: []domain.Transition{}, // No exit
	}

	err = tx.Save(ctx, core.Document{
		ID:      node2.ID + ".json",
		Content: string(node2.Content),
		Metadata: core.Metadata{
			"ID":          node2.ID,
			"Type":        node2.Type,
			"Transitions": node2.Transitions,
		},
	})
	if err != nil {
		fmt.Printf("Save end failed: %v\n", err)
	}

	if err := tx.Commit(ctx, "Seed data"); err != nil {
		fmt.Printf("Commit failed: %v\n", err)
	}

	// 3. Setup Core (hex type)
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
