package runner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
)

// Runner handles the execution loop of the Trellis engine using provided IO.
// This allows for easy testing and integration with different frontends (CLI, TUI, etc).
// Runner handles the execution loop of the Trellis engine using provided IO.
// It uses an IOHandler strategy to abstract the interaction mode (Text vs JSON).
type Runner struct {
	// Handler is the strategy for IO. If nil, it falls back to legacy fields.
	Handler IOHandler

	// Interceptor is a middleware for tool execution policy.
	// If nil, defaults to AutoApprove (Phase 1 behavior).
	Interceptor ToolInterceptor

	// Logger is used for internal debug logging.
	// If nil, a no-op logger is used.
	Logger *slog.Logger

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
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

// Run executes the engine loop until termination.
// If initialState is nil, engine.Start() is called to create a new state.
func (r *Runner) Run(engine *trellis.Engine, initialState *domain.State) error {
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

	var state *domain.State
	if initialState != nil {
		state = initialState
	} else {
		// If no initial state provided, creating default "start" state
		var err error
		state, err = engine.Start(context.Background())
		if err != nil {
			return fmt.Errorf("failed to create initial state: %w", err)
		}
	}
	lastRenderedID := ""

	// Setup Signal Manager
	signals := NewSignalManager()
	defer signals.Stop()

	// Resolve Interceptor
	interceptor := r.Interceptor
	if interceptor == nil {
		// Default Policy:
		// - Headless: Auto-approve all tool calls for automation.
		// - Interactive: Require user confirmation for safety.
		if r.Headless {
			interceptor = AutoApproveMiddleware()
		} else {
			interceptor = ConfirmationMiddleware(handler)
		}
	}

	for {
		// 1. Render Phase (View)
		// We use the signals.Context() because it might be reset
		currentCtx := signals.Context()
		actions, isTerminal, err := engine.Render(currentCtx, state)
		if err != nil {
			return fmt.Errorf("render error: %w", err)
		}

		// 2. Output Phase
		needsInput, err := handler.Output(currentCtx, actions)
		if err != nil {
			return fmt.Errorf("output error: %w", err)
		}

		// Optimization: Update lastRendered
		if state.CurrentNodeID != lastRenderedID {
			lastRenderedID = state.CurrentNodeID
		}

		if isTerminal {
			// If the node is terminal but requested input (e.g. wait: true),
			// we must honor that request (Pause before Exit).
			if needsInput {
				_, err := handler.Input(currentCtx)
				// Race Mitigation
				if err != nil {
					signals.CheckRace()
				}

				if currentCtx.Err() != nil {
					r.Logger.Debug("Runner terminal wait: Context cancelled", "err", currentCtx.Err())
					// Attempt Global Signal Transition even at terminal node
					nextState, sigErr := engine.Signal(context.Background(), state, "interrupt")
					if sigErr == nil {
						r.Logger.Debug("Runner terminal wait: Signal transition success")
						// Re-arm signals
						signals.Reset()
						state = nextState
						continue
					}
					r.Logger.Debug("Runner terminal wait: Signal failed", "error", sigErr)
				}
				// If normal input (Enter) or unhandled signal, we proceed to exit
				if err != nil && err != io.EOF && currentCtx.Err() == nil {
					r.Logger.Error("Error during wait", "error", err)
				}
			}
			break
		}

		// 3. Input/Action Phase
		var nextInput any

		if state.Status == domain.StatusWaitingForTool {
			// Find the ToolCall in actions that matches the pending call
			var pendingCall *domain.ToolCall

			// We iterate actions to find the one that triggered this wait.
			// Ideally, Render should return it, or we find it in the actions payload.
			for _, act := range actions {
				if act.Type == domain.ActionCallTool {
					if call, ok := act.Payload.(domain.ToolCall); ok {
						if call.ID == state.PendingToolCall {
							pendingCall = &call
							break
						}
					}
				}
			}

			if pendingCall == nil {
				return fmt.Errorf("state is waiting for tool %s but no corresponding action produced", state.PendingToolCall)
			}

			// Interceptor / Policy Check
			allowed, policyResult, err := interceptor(currentCtx, *pendingCall)
			if err != nil {
				return fmt.Errorf("tool interceptor error: %w", err)
			}

			if !allowed {
				// Blocked by policy, return the policy result (denial)
				nextInput = policyResult
			} else {
				// Approved: Execute Tool (Mechanic: Pause -> host executes -> Result)
				result, err := handler.HandleTool(currentCtx, *pendingCall)
				if err != nil {
					return fmt.Errorf("tool execution failed: %w", err)
				}
				nextInput = result
			}

		} else {
			// Active State: Standard User Input
			if needsInput || !r.Headless {
				// We pass ctx to Input. If signal received, Input implies cancellation.
				// Note: TextHandler's Input (standard fmt.Scan) might not respect context immediately.
				val, err := handler.Input(currentCtx)
				if err != nil {
					// Check if error is due to signal cancellation
					signals.CheckRace()

					if currentCtx.Err() != nil {
						r.Logger.Debug("Runner input: Context cancelled", "err", currentCtx.Err())
						// Attempt Global Signal Transition
						nextState, sigErr := engine.Signal(context.Background(), state, "interrupt")
						if sigErr == nil {
							r.Logger.Debug("Runner input: Signal transition success")
							// Re-arm signals for the new state
							signals.Reset()

							// Successfully handled signal! Update state and loop.
							state = nextState
							continue
						}
						r.Logger.Debug("Runner input: Signal failed", "error", sigErr)
						// If signal unhandled, fallthrough to exit
						return fmt.Errorf("interrupted")
					}

					if err == io.EOF {
						break
					}
					return fmt.Errorf("input error: %w", err)
				}

				if val == "exit" || val == "quit" {
					break
				}
				nextInput = val
			} else {
				// No input needed (auto-transition?), but usually this implies a Wait or immediate move.
				// If headless and no input needed, we might Loop forever if logic is broken.
				// For now, pass empty string.
				nextInput = ""
			}
		}

		// 4. Navigate Phase (Controller)
		nextState, err := engine.Navigate(context.Background(), state, nextInput)
		if err != nil {
			return fmt.Errorf("navigation error: %w", err)
		}

		if nextState.Terminated || nextState.Status == domain.StatusTerminated {
			break
		}
		state = nextState
	}
	return nil
}
