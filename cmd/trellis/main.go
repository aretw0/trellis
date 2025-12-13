package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aretw0/trellis"
)

func main() {
	// 0. Parse Args
	repoPath := "."
	if len(os.Args) > 1 {
		args := os.Args[1:]
		if len(args) > 0 {
			if args[0] == "--dir" && len(args) > 1 {
				repoPath = args[1]
			} else if !strings.HasPrefix(args[0], "-") {
				repoPath = args[0]
			}
		}
	}

	// 1. Initialize Engine via Public Facade
	engine, err := trellis.New(repoPath)
	if err != nil {
		fmt.Printf("Failed to initialize trellis: %v\n", err)
		os.Exit(1)
	}

	state := engine.Start()
	lastRenderedID := ""

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("--- Trellis CLI (via Facade) ---")

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
		if state.CurrentNodeID != lastRenderedID {
			for _, act := range actions {
				if act.Type == "CLI_PRINT" {
					if msg, ok := act.Payload.(string); ok {
						fmt.Println(strings.TrimSpace(msg))
					}
				}
			}
			lastRenderedID = state.CurrentNodeID
		}

		// Generic Exit condition
		if nextState.Terminated {
			break
		}

		// Input needed?
		if nextState.CurrentNodeID == state.CurrentNodeID {
			fmt.Print("> ")
			text, _ := reader.ReadString('\n')
			input = strings.TrimSpace(text)

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

			for _, act := range actions {
				if act.Type == "CLI_PRINT" {
					if msg, ok := act.Payload.(string); ok {
						fmt.Println(strings.TrimSpace(msg))
					}
				}
			}

			state = nextState
		} else {
			state = nextState
		}
	}
}
