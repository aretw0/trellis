package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aretw0/loam"
	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/runtime"
	"github.com/aretw0/trellis/pkg/domain"
)

func main() {
	// 0. Parse Args
	repoPath := "."
	if len(os.Args) > 1 {
		// Simple arg parsing. Could use flags but user asked for "trellis.exe --dir" or just arg.
		// Let's support just the arg for simplicity as per "trellis.exe ./data-folder" in plan.
		// If user passes flags, we might need more logic.
		// For now, treat first non-flag arg as path, or just first arg.
		// Plan said: "Accept a --dir flag (or argument)".
		// Let's try to match user expectation: "no inicializa-lo com versionamento".
		// Loam Init takes path.
		// Let's look for --dir explicitly or just check args.

		args := os.Args[1:]
		if len(args) > 0 {
			if args[0] == "--dir" && len(args) > 1 {
				repoPath = args[1]
			} else if !strings.HasPrefix(args[0], "-") {
				repoPath = args[0]
			}
		}
	}

	// 1. Setup Loam Repository
	absPath, _ := filepath.Abs(repoPath)
	// Note: We use WithVersioning(false) to treat it as a player/viewer without side-effects.
	repo, err := loam.Init(absPath, loam.WithVersioning(false))
	if err != nil {
		fmt.Printf("Failed to init loam: %v\n", err)
		os.Exit(1)
	}

	// 2. Wrap with Typed Repository (Read-Only usage essentially)
	typedRepo := loam.NewTypedRepository[adapters.NodeMetadata](repo)

	// 3. Setup Adapter
	loader := adapters.NewLoamLoader(typedRepo)

	// 4. Seeding Removed (Using Golden Path or existing data)

	// 5. Setup Core (hex type)
	engine := runtime.NewEngine(loader)
	state := domain.NewState("start")
	lastRenderedID := "" // Track to avoid re-printing static content

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

		// Dispatch Actions (Only if we haven't rendered this node yet, or it's new feedback)
		// For the initial "Render" step (input=""), we suppress output if we are still on the same node.
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

		// Generic Exit condition (Sink Node)
		if nextState.Terminated {
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
					if msg, ok := act.Payload.(string); ok {
						fmt.Println(strings.TrimSpace(msg))
					}
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
