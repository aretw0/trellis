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
// Runner handler the execution loop of the Trellis engine.
type Runner struct {
	Handler     IOHandler
	Store       ports.StateStore
	Logger      *slog.Logger
	Headless    bool
	SessionID   string
	Renderer    ContentRenderer
	Interceptor ToolInterceptor
	ToolRunner  ToolRunner
}

// ContentRenderer is a function that transforms the content before outputting it.
// This allows for TUI rendering (markdown to ANSI) without coupling the core package.
type ContentRenderer func(string) (string, error)

// NewRunner creates a new Runner with options.
func NewRunner(opts ...Option) *Runner {
	r := &Runner{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// Run executes the engine loop until termination.
func (r *Runner) Run(ctx context.Context, engine *trellis.Engine, initialState *domain.State) (*domain.State, error) {
	// 1. Setup Phase
	handler := r.resolveHandler()
	interceptor := r.resolveInterceptor(handler)

	state, err := r.resolveInitialState(ctx, engine, initialState)
	if err != nil {
		return nil, err
	}

	signals := NewSignalManager()
	defer signals.Stop()

	lastRenderedID := ""

	// 2. Execution Loop
	for {
		// Check for cancellation before each step
		select {
		case <-ctx.Done():
			return state, ctx.Err()
		default:
		}

		// currentCtx combines signals and the runner context
		currentCtx, currentCancel := context.WithCancel(ctx)
		go func() {
			select {
			case <-signals.Context().Done():
				currentCancel()
			case <-ctx.Done():
				currentCancel()
			}
		}()
		defer currentCancel()

		// A. Render
		actions, _, err := engine.Render(currentCtx, state)
		if err != nil {
			return state, fmt.Errorf("render error: %w", err)
		}

		// B. Output
		stepTimeout := r.detectTimeout(actions)
		inputCtx, inputCancel := r.createInputContext(currentCtx, stepTimeout)
		defer inputCancel()

		needsInput, err := handler.Output(currentCtx, actions)
		if err != nil {
			return state, fmt.Errorf("output error: %w", err)
		}

		if state.CurrentNodeID != lastRenderedID {
			lastRenderedID = state.CurrentNodeID
		}

		// C. Input / Tool Execution
		var nextInput any
		var nextState *domain.State

		// If no input requested by actions (needsInput=false) and not waiting for tool,
		// we treat this as an auto-transition (pass-through).
		// We skip the Handler.Input blocking call entirely.
		if !needsInput && state.Status != domain.StatusWaitingForTool && state.Status != domain.StatusRollingBack {
			// Auto-transition with empty input
			nextInput = ""
		} else if state.Status == domain.StatusWaitingForTool || state.Status == domain.StatusRollingBack {
			nextInput, err = r.handleTool(currentCtx, actions, state, handler, interceptor)
		} else {
			nextInput, nextState, err = r.handleInput(inputCtx, handler, needsInput, signals, engine, state)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return state, err
		}

		// If a signal caused a transition, update state and loop immediately
		if nextState != nil {
			state = nextState
			// Auto-Save on Signal Transition
			if err := r.saveState(currentCtx, r.SessionID, state); err != nil {
				return state, err
			}
			continue
		}

		// 4. Navigate Phase (Controller)
		// nextInput is valid here.
		// Always call Navigate to ensure lifecycle hooks (Leave) are triggered and state is updated.
		nextState, err = engine.Navigate(currentCtx, state, nextInput)
		if err != nil {
			return state, fmt.Errorf("navigation error: %w", err)
		}

		// 5. Commit Phase (Persistence)
		if err := r.saveState(currentCtx, r.SessionID, nextState); err != nil {
			return nextState, fmt.Errorf("critical persistence error: %w", err)
		}

		if nextState == nil {
			// Should not happen if Navigate returns correctly
			break
		}

		if nextState.Terminated || nextState.Status == domain.StatusTerminated {
			state = nextState
			break
		}
		state = nextState
	}
	// Final cleanup if needed (e.g. remove session if complete? Optional logic)
	return state, nil
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
	th := NewTextHandler(os.Stdin, os.Stdout)
	th.Renderer = r.Renderer

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

func (r *Runner) resolveInitialState(ctx context.Context, engine *trellis.Engine, initial *domain.State) (*domain.State, error) {
	if initial != nil {
		return initial, nil
	}
	state, err := engine.Start(ctx, "runner-ephemeral", nil)
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

	// Priority: Explicit ToolRunner > Handler (Legacy/IO-bound)
	if r.ToolRunner != nil {
		return r.ToolRunner.Execute(ctx, *pendingCall)
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
	// If we got here, we are in Interactive Mode BUT needsInput is false?
	// This shouldn't happen if logic above handles auto-transition.
	// But in case handleInput IS called with needsInput=false:
	if !needsInput {
		return "", nil, nil
	}

	// We pass ctx to Input. If signal received, Input implies cancellation.
	// Note: TextHandler's Input (standard fmt.Scan) might not respect context immediately.
	val, err := handler.Input(ctx)
	if err != nil {
		// Check if error is due to signal cancellation
		signals.CheckRace()

		if ctx.Err() != nil {
			if ctx.Err() == context.Canceled {
				// Check if this cancellation came from the SignalManager (SIGINT) or external (Reload)
				if signals.Context().Err() != nil {
					r.Logger.Debug("Runner input: Interrupted (Clean Exit)")
				} else {
					// This means ctx was cancelled by the parent (e.g. Watcher reload), not by SIGINT.
					// We should not treat this as a signal that needs 'on_signal' handling.
					r.Logger.Debug("Runner input: Context Cancelled (Reload/Stop)")
					return nil, nil, fmt.Errorf("interrupted")
				}
			} else if ctx.Err() == context.DeadlineExceeded {
				r.Logger.Debug("Runner input: Context Expired (Timeout)")
			} else {
				r.Logger.Debug("Runner input: Context Error", "err", ctx.Err())
			}

			// Determine cause: Global Interrupt vs Local Timeout
			signalName := domain.SignalInterrupt
			if ctx.Err() == context.DeadlineExceeded {
				signalName = domain.SignalTimeout
			}

			// Attempt Signal Transition
			nextState, sigErr := engine.Signal(context.Background(), currentState, signalName)
			if sigErr == nil {
				r.Logger.Debug("Runner input: Signal transition success", "signal", signalName)
				// Re-arm signals for the new state
				signals.Reset()
				return nil, nextState, nil
			}

			// Default Behavior: If no custom handler is defined for the signal, we break execution gracefully.
			// For known signals like 'interrupt' or 'timeout', we log a helpful hint.
			switch signalName {
			case domain.SignalInterrupt:
				r.Logger.Debug("Runner input: Stopping (Default Exit)", "signal", signalName, "help", "Define 'on_signal: interrupt' to override")
				return nil, nil, fmt.Errorf("interrupted")
			case domain.SignalTimeout:
				r.Logger.Debug("Runner input: Stopping (Default Exit)", "signal", signalName, "help", "Define 'on_signal: timeout' to override")
				return nil, nil, fmt.Errorf("timeout exceeded and no 'on_signal.timeout' handler defined")
			default:
				r.Logger.Debug("Runner input: Stopping (Unhandled Signal)", "signal", signalName)
				return nil, nil, fmt.Errorf("interrupted")
			}
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
