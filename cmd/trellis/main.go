package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/domain"
)

func main() {
	// 1. Setup Adapter (driven port)
	loader := adapters.NewInMemoryLoader()

	// Seed some data
	node1 := domain.Node{
		ID:      "start",
		Type:    "question",
		Content: []byte("Welcome to Trellis! Type 'go' to continue."),
		Transitions: []domain.Transition{
			{ToNodeID: "end", Condition: "input == 'go'"},
		},
	}
	data1, _ := json.Marshal(node1)
	loader.AddNode("start", data1)

	node2 := domain.Node{
		ID:          "end",
		Type:        "text",
		Content:     []byte("You have reached the end. Goodbye!"),
		Transitions: []domain.Transition{}, // No exit
	}
	data2, _ := json.Marshal(node2)
	loader.AddNode("end", data2)

	// 2. Setup Core (hex type)
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
