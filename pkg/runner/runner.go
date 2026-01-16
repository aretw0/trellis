package runner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
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

	// Store is the persistence adapter for durable execution.
	// If nil, sessions are ephemeral.
	Store ports.StateStore

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
// If sessionID is provided (via RunSession wrapper or custom logic), it is used for persistence.
// Note: backward compatible signature, but we might want `RunSession` in future.
// For now, we assume explicit session management is done before calling Run, OR we rely on Runner.Store.
func (r *Runner) Run(engine *trellis.Engine, initialState *domain.State, sessionID string) error {
	// 1. Setup Phase
	handler := r.resolveHandler()
	interceptor := r.resolveInterceptor(handler)

	state, err := r.resolveInitialState(engine, initialState)
	if err != nil {
		return err
	}

	signals := NewSignalManager()
	defer signals.Stop()

	lastRenderedID := ""

	// 2. Execution Loop
	for {
		currentCtx := signals.Context()

		// A. Render
		actions, _, err := engine.Render(currentCtx, state)
		if err != nil {
			return fmt.Errorf("render error: %w", err)
		}

		// B. Output
		stepTimeout := r.detectTimeout(actions)
		inputCtx, cancel := r.createInputContext(currentCtx, stepTimeout)

		needsInput, err := handler.Output(currentCtx, actions)
		if err != nil {
			cancel()
			return fmt.Errorf("output error: %w", err)
		}

		if state.CurrentNodeID != lastRenderedID {
			lastRenderedID = state.CurrentNodeID
		}

		// C. Input / Tool Execution
		var nextInput any
		var nextState *domain.State

		if state.Status == domain.StatusWaitingForTool {
			nextInput, err = r.handleTool(currentCtx, actions, state, handler, interceptor)
		} else {
			nextInput, nextState, err = r.handleInput(inputCtx, handler, needsInput, signals, engine, state)
		}

		cancel() // Ensure context is cancelled after input/tool phase

		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// If a signal caused a transition, update state and loop immediately
		if nextState != nil {
			state = nextState
			// Auto-Save on Signal Transition
			if err := r.saveState(currentCtx, sessionID, state); err != nil {
				return err
			}
			continue
		}

		// 4. Navigate Phase (Controller)
		// nextInput is valid here
		nextState, err = engine.Navigate(context.Background(), state, nextInput)
		if err != nil {
			return fmt.Errorf("navigation error: %w", err)
		}

		// 5. Commit Phase (Persistence)
		if err := r.saveState(context.Background(), sessionID, nextState); err != nil {
			return fmt.Errorf("critical persistence error: %w", err)
		}

		if nextState.Terminated || nextState.Status == domain.StatusTerminated {
			break
		}
		state = nextState
	}
	// Final cleanup if needed (e.g. remove session if complete? Optional logic)
	return nil
}

func (r *Runner) saveState(ctx context.Context, sessionID string, state *domain.State) error {
	if r.Store != nil && sessionID != "" {
		if err := r.Store.Save(ctx, sessionID, state); err != nil {
			return err
		}
		r.Logger.Debug("state saved", "session_id", sessionID, "node_id", state.CurrentNodeID)
	}
	return nil
}

// resolveHandler ensures a valid IOHandler is set.
func (r *Runner) resolveHandler() IOHandler {
	if r.Handler != nil {
		return r.Handler
	}
	th := NewTextHandler(r.Input, r.Output)
	th.Renderer = r.Renderer
	if !r.Headless && r.Output != nil {
		fmt.Fprintln(r.Output, "--- Trellis CLI (Runner) ---")
	}
	// Memoize to prevent creating new Pumps on subsequent Run() calls
	r.Handler = th
	return th
}

// resolveInterceptor returns the configured or default interceptor.
func (r *Runner) resolveInterceptor(h IOHandler) ToolInterceptor {
	if r.Interceptor != nil {
		return r.Interceptor
	}
	if r.Headless {
		return AutoApproveMiddleware()
	}
	return ConfirmationMiddleware(h)
}

func (r *Runner) resolveInitialState(engine *trellis.Engine, initial *domain.State) (*domain.State, error) {
	if initial != nil {
		return initial, nil
	}
	state, err := engine.Start(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create initial state: %w", err)
	}
	return state, nil
}

func (r *Runner) detectTimeout(actions []domain.ActionRequest) time.Duration {
	for _, act := range actions {
		if act.Type == domain.ActionRequestInput {
			if req, ok := act.Payload.(domain.InputRequest); ok {
				return req.Timeout
			}
		}
	}
	return 0
}

func (r *Runner) createInputContext(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(parent, timeout)
	}
	return parent, func() {}
}

// handleTool manages the execution of pending tools.
func (r *Runner) handleTool(
	ctx context.Context,
	actions []domain.ActionRequest,
	state *domain.State,
	handler IOHandler,
	interceptor ToolInterceptor,
) (any, error) {
	var pendingCall *domain.ToolCall
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
		return nil, fmt.Errorf("state is waiting for tool %s but no corresponding action produced", state.PendingToolCall)
	}

	allowed, policyResult, err := interceptor(ctx, *pendingCall)
	if err != nil {
		return nil, fmt.Errorf("tool interceptor error: %w", err)
	}

	if !allowed {
		return policyResult, nil
	}

	result, err := handler.HandleTool(ctx, *pendingCall)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}
	return result, nil
}

// handleInput manages standard user interaction and signal recovery.
// Returns:
// - input: The value to pass to Navigate (if success)
// - nextState: A new state if a signal caused a transition (input should be ignored)
// - error: If input failed or signal execution failed
func (r *Runner) handleInput(
	ctx context.Context,
	handler IOHandler,
	needsInput bool,
	signals *SignalManager,
	engine *trellis.Engine,
	currentState *domain.State,
) (any, *domain.State, error) {
	if !needsInput && r.Headless {
		// No input needed (auto-transition?), but usually this implies a Wait or immediate move.
		// If headless and no input needed, we might Loop forever if logic is broken.
		// For now, pass empty string.
		return "", nil, nil
	}

	// We pass ctx to Input. If signal received, Input implies cancellation.
	// Note: TextHandler's Input (standard fmt.Scan) might not respect context immediately.
	val, err := handler.Input(ctx)
	if err != nil {
		// Check if error is due to signal cancellation
		signals.CheckRace()

		if ctx.Err() != nil {
			r.Logger.Debug("Runner input: Context cancelled", "err", ctx.Err())

			// Determine cause: Global Interrupt vs Local Timeout
			signalName := "interrupt"
			if ctx.Err() == context.DeadlineExceeded {
				signalName = "timeout"
			}

			// Attempt Signal Transition
			nextState, sigErr := engine.Signal(context.Background(), currentState, signalName)
			if sigErr == nil {
				r.Logger.Debug("Runner input: Signal transition success", "signal", signalName)
				// Re-arm signals for the new state
				signals.Reset()
				return nil, nextState, nil
			}

			if signalName == "timeout" {
				// If unhandled timeout, we might just loop? Or error?
				// For now, if unhandled timeout, we return error to stop execution
				return nil, nil, fmt.Errorf("timeout exceeded and no 'on_signal.timeout' handler defined")
			}

			r.Logger.Debug("Runner input: Signal failed", "error", sigErr)
			// If signal unhandled, fallthrough to exit
			return nil, nil, fmt.Errorf("interrupted")
		}

		if err == io.EOF {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("input error: %w", err)
	}

	if val == "exit" || val == "quit" {
		return nil, nil, io.EOF
	}

	return val, nil, nil
}
