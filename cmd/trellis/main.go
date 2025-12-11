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
		// Run Step with empty input first to render current node?
		// Or we loop: Render -> Input -> Step.
		// But Step does Render implicitly via Actions.
		// Let's call Step with empty input to 'Enter' the node if we just arrived?
		// Our Step logic checks "if input matches condition".
		// Use a loop that prompts.

		// This loop is a loose "Driver" implementation.

		// Execute Step (Processing input from previous iteration, or empty for first run)
		// ... Wait, Step function in my implementation takes input and decides transition.
		// If we are at "start", we need to SHOW "start" content.
		// If I pass empty input, "start" Node logic runs.
		// It says "question", prints "Welcome...", sees no input match, returns same state.

		// Logic:
		// 1. Run Step(state, input).
		// 2. Handle Actions (Print).
		// 3. If State changed, loop with empty input?
		//    If State didn't change, READ input, then loop with input.

		// Refined Loop:
		var input string
		// check if this is the first run

		actions, nextState, err := engine.Step(state, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}

		// Dispatch Actions (Driver Port Logic)
		for _, act := range actions {
			if act.Type == "CLI_PRINT" {
				fmt.Println(act.Payload)
			}
		}

		// Check if we need input
		if nextState.CurrentNodeID == state.CurrentNodeID {
			// We didn't move. Means we are waiting for input or stuck.
			// Read input
			fmt.Print("> ")
			text, _ := reader.ReadString('\n')
			input = strings.TrimSpace(text)

			// Run Step again with input
			actions, nextState, err = engine.Step(state, input)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				break
			}
		}

		// Update state
		state = nextState
		input = "" // Clear input for next step "entry"

		// Exit condition for demo
		if state.CurrentNodeID == "end" {
			// One last print?
			// The loop will run one more time for "end", print "Goodbye", then wait for input.
			// If "end" has no transitions, we start loop, Step prints "Goodbye", actions handled.
			// nextState == state. We wait for input.
			// User types anything. Step runs again. No transitions match. State same.
			// Loop continues.
			// It's fine for a REPL.
		}
	}
}
