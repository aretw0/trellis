package trellis

import (
	"context"
	"fmt"
	"io"
	"os"
)

// Runner handles the execution loop of the Trellis engine using provided IO.
// This allows for easy testing and integration with different frontends (CLI, TUI, etc).
// Runner handles the execution loop of the Trellis engine using provided IO.
// It uses an IOHandler strategy to abstract the interaction mode (Text vs JSON).
type Runner struct {
	// Handler is the strategy for IO. If nil, it falls back to legacy fields.
	Handler IOHandler

	// Deprecated: Use Handler instead. These are kept for backward compatibility.
	Input    io.Reader
	Output   io.Writer
	Headless bool
	Renderer ContentRenderer
}

// ContentRenderer is a function that transforms the content before outputting it.
// This allows for TUI rendering (markdown to ANSI) without coupling the core package.
type ContentRenderer func(string) (string, error)

// NewRunner creates a new Runner with default Stdin/Stdout.
func NewRunner() *Runner {
	return &Runner{
		Input:    os.Stdin,
		Output:   os.Stdout,
		Headless: false,
	}
}

// Run executes the engine loop until termination.
func (r *Runner) Run(engine *Engine) error {
	// Resolve Strategy
	handler := r.Handler
	if handler == nil {
		// Fallback to legacy TextHandler behavior using struct fields
		th := NewTextHandler(r.Input, r.Output)
		th.Renderer = r.Renderer

		// Legacy headless support: suppress welcome message in TextHandler fallback
		if !r.Headless && r.Output != nil {
			fmt.Fprintln(r.Output, "--- Trellis CLI (Runner) ---")
		}

		handler = th
	}

	state := engine.Start()
	lastRenderedID := ""

	for {
		// 1. Render Phase (View)
		actions, isTerminal, err := engine.Render(context.Background(), state)
		if err != nil {
			return fmt.Errorf("render error: %w", err)
		}

		// 2. Output Phase
		needsInput, err := handler.Output(context.Background(), actions)
		if err != nil {
			return fmt.Errorf("output error: %w", err)
		}

		// Optimization: Only update lastRendered if we actually moved?
		// Logic preserved from original: if we output something, we usually moved.
		if state.CurrentNodeID != lastRenderedID {
			lastRenderedID = state.CurrentNodeID
		}

		// 3. Wait Phase (Input)
		if isTerminal {
			break
		}

		var input string

		// Logic Check: If the handler says it *needs* or *expects* input (or if it's generic text mode)
		// For TextHandler, Output returns true if it printed content.
		if needsInput || !r.Headless {
			// In legacy headless, we still read input, just didn't prompt "> ".
			// TextHandler.Input adds prompt.
			// We might need to adjust TextHandler to respect "Headless" or let the user config it.
			// For now, let's assume the Handler handles the prompting.

			val, err := handler.Input(context.Background())
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("input error: %w", err)
			}
			input = val

			if input == "exit" || input == "quit" {
				// TODO: Handler should probably handle "Bye!" message too?
				break
			}
		}

		// 4. Navigate Phase (Controller)
		nextState, err := engine.Navigate(context.Background(), state, input)
		if err != nil {
			return fmt.Errorf("navigation error: %w", err)
		}

		if nextState.Terminated {
			break
		}
		state = nextState
	}
	return nil
}
