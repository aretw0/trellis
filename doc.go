/*
Package trellis is a deterministic state machine engine (DFA) designed for building robust conversational agents, CLIs, and automation workflows.

It implements a "Reentrant DFA with Controlled Side-Effects" architecture, separating the narrative graph (Logic) from the execution state (Context) and side-effects (Tools).

# Concept

Trellis treats your application flow as a graph of nodes. The engine manages the state transitions, data binding, and persistence, while your application ("Host") manages the I/O and external tool execution. This Hexagonal Architecture allows Trellis to be embedded in any interface: CLI, HTTP Server, or AI Agent infrastructure.

# Key Features

  - Deterministic Execution: Given the same state and input, the transition is always reproducible.
  - Hexagonal Architecture: Core logic is decoupled from adapters (Storage, UI, Tools).
  - State Persistence: Built-in support for long-running sessions ("Durable Execution").
  - Strict Contracts: Validates graph integrity and data types to prevent runtime surprises.

# Usage

Initialize the engine using the "Start" entrypoint. You can use the default filesystem loader (Loam) or inject a custom one.

	package main

	import (
		"context"
		"log"

		"github.com/aretw0/trellis"
	)

	func main() {
		// Initialize Engine with default settings (reads from ./my-flow)
		eng, err := trellis.New("./my-flow")
		if err != nil {
			log.Fatal(err)
		}

		// Start a new session
		ctx := context.Background()
		state, err := eng.Start(ctx, "session-123", nil)
		if err != nil {
			log.Fatal(err)
		}

		// Main Loop: Render -> Input -> Navigate
		for {
			// 1. Render View (What to show/do?)
			actions, terminal, valErr := eng.Render(ctx, state)
			if valErr != nil {
				log.Printf("Error: %v", valErr)
				break
			}

			// Handle Actions (Print text, Call tools...)
			for _, act := range actions {
				log.Println("Action:", act)
			}

			if terminal {
				log.Println("End of flow.")
				break
			}

			// 2. Navigate (Next Step)
			// In a real app, this input comes from User or Tool Result
			state, err = eng.Navigate(ctx, state, "user input")
			if err != nil {
				log.Fatal(err)
			}
		}
	}
*/
package trellis
