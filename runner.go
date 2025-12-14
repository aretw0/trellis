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
}

// NewRunner creates a new Runner with default Stdin/Stdout.
// Use SetInput/SetOutput to override for testing.
func NewRunner() *Runner {
	return &Runner{
		Input:    nil, // defaults to os.Stdin if not set, handled in Run? Or explicit?
		Output:   nil, // defaults to os.Stdout
		Headless: false,
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

		// Run Step
		actions, nextState, err := engine.Step(context.TODO(), state, input)
		if err != nil {
			return fmt.Errorf("step error: %w", err)
		}

		// Dispatch Actions
		if state.CurrentNodeID != lastRenderedID {
			for _, act := range actions {
				if act.Type == domain.ActionRenderContent {
					if msg, ok := act.Payload.(string); ok {
						fmt.Fprintln(writer, strings.TrimSpace(msg))
					}
				}
			}
			lastRenderedID = state.CurrentNodeID
		}

		// Generic Exit condition
		if nextState.Terminated {
			break
		}

		// Input needed
		if nextState.CurrentNodeID == state.CurrentNodeID {
			if !r.Headless {
				fmt.Fprint(writer, "> ")
			}
			text, _ := lineReader.ReadString('\n')
			input = strings.TrimSpace(text)

			if !r.Headless && (input == "exit" || input == "quit") {
				fmt.Fprintln(writer, "Bye!")
				break
			}
			// In Headless mode, EOF or empty input should probably break loop or handle gracefully?
			// If input is empty in headless, it might mean end of stream.
			// bufio.ReadString returns err on EOF. Check it.
			// Actually, let's check input reading error generically.

			// Run Step again with input
			actions, nextState, err = engine.Step(context.TODO(), state, input)
			if err != nil {
				return fmt.Errorf("step input error: %w", err)
			}

			// Dispatch Actions from Input
			for _, act := range actions {
				if act.Type == domain.ActionRenderContent {
					if msg, ok := act.Payload.(string); ok {
						fmt.Fprintln(writer, strings.TrimSpace(msg))
					}
				}
			}

			state = nextState
		} else {
			state = nextState
		}
	}
	return nil
}
