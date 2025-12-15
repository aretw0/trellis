package trellis

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aretw0/trellis/pkg/domain"
)

// Runner handles the execution loop of the Trellis engine using provided IO.
// This allows for easy testing and integration with different frontends (CLI, TUI, etc).
type Runner struct {
	Input    io.Reader
	Output   io.Writer
	Headless bool
	Renderer ContentRenderer
}

// ContentRenderer is a function that transforms the content before outputting it.
// This allows for TUI rendering (markdown to ANSI) without coupling the core package.
type ContentRenderer func(string) (string, error)

// NewRunner creates a new Runner with default Stdin/Stdout.
// Use SetInput/SetOutput to override for testing.
func NewRunner() *Runner {
	return &Runner{
		Input:    nil, // defaults to os.Stdin if not set, handled in Run? Or explicit?
		Output:   nil, // defaults to os.Stdout
		Headless: false,
		Renderer: nil,
	}
}

// Run executes the engine loop until termination.
func (r *Runner) Run(engine *Engine) error {
	// Resolve IO
	reader := r.Input
	// We need a bufio Reader for line reading
	var lineReader *bufio.Reader
	if reader == nil {
		// We can't default to os.Stdin here easily without importing os?
		// Ideally the caller sets it. But for DX, let's allow nil.
		// Wait, importing "os" in root is fine.
		return fmt.Errorf("input reader must be set (use os.Stdin)")
	}
	lineReader = bufio.NewReader(reader)

	writer := r.Output
	if writer == nil {
		return fmt.Errorf("output writer must be set (use os.Stdout)")
	}

	state := engine.Start()
	lastRenderedID := ""

	if !r.Headless {
		fmt.Fprintln(writer, "--- Trellis CLI (Runner) ---")
	}

	for {
		var input string

		// 1. Render Phase (View)
		// We always ask the engine what to show.
		actions, isTerminal, err := engine.Render(context.Background(), state)
		if err != nil {
			return fmt.Errorf("render error: %w", err)
		}

		// 2. Display Phase & Input Decision
		hasContent := false
		if state.CurrentNodeID != lastRenderedID {
			for _, act := range actions {
				if act.Type == domain.ActionRenderContent {
					if msg, ok := act.Payload.(string); ok {
						hasContent = true
						output := msg
						if r.Renderer != nil {
							rendered, err := r.Renderer(msg)
							if err == nil {
								output = rendered
							}
						}
						// Ensure we print a newline after content
						fmt.Fprintln(writer, strings.TrimSpace(output))
					}
				}
			}
			lastRenderedID = state.CurrentNodeID
		}

		// 3. Wait Phase (Input)
		// Policy: If we showed content, we must wait for user acknowledgment/input.
		// If we didn't show anything (e.g. logic node), we proceed with empty input (auto-skip).
		// EXCEPTION: If isTerminal is true, we are done. We showed the final message (if any) and we exit.
		if isTerminal {
			break
		}

		needsInput := hasContent && !r.Headless

		if needsInput {
			fmt.Fprint(writer, "> ")
			text, _ := lineReader.ReadString('\n')
			input = strings.TrimSpace(text)

			if input == "exit" || input == "quit" {
				fmt.Fprintln(writer, "Bye!")
				break
			}
		}

		// 4. Navigate Phase (Controller)
		nextState, err := engine.Navigate(context.Background(), state, input)
		if err != nil {
			return fmt.Errorf("navigation error: %w", err)
		}

		// Check Exit Condition
		if nextState.Terminated {
			break
		}

		state = nextState
	}
	return nil
}
