package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/aretw0/lifecycle"

	"github.com/aretw0/introspection"
	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
)

// Runner handles the execution loop of the Trellis engine using provided IO.
// This allows for easy testing and integration with different frontends (CLI, TUI, etc).
//
// Runner implements the lifecycle.Worker interface.
//
// Note: Runner is designed for single-use execution. It is NOT thread-safe for concurrent usage.
// Create a new Runner instance for each execution.
type Runner struct {
	Handler         IOHandler
	Store           ports.StateStore
	Logger          *slog.Logger
	Headless        bool
	SessionID       string
	Renderer        ContentRenderer
	Interceptor     ToolInterceptor
	ToolRunner      ToolRunner
	InterruptSource <-chan struct{}

	// Worker Pattern: Self-contained execution context
	engine       *trellis.Engine
	initialState *domain.State
	finalState   *domain.State

	// State Watching
	stateMu   sync.RWMutex
	lastState *domain.State
	watchers  []chan introspection.StateChange[*domain.State]
}

// State returns a snapshot of the current execution state.
// It implements the introspection.TypedWatcher interface.
func (r *Runner) State() *domain.State {
	r.stateMu.RLock()
	defer r.stateMu.RUnlock()
	if r.lastState == nil {
		return r.initialState
	}
	// lastState is already a snapshot, no need to clone again
	return r.lastState
}

// Watch returns a channel of state changes for the introspection system.
func (r *Runner) Watch(ctx context.Context) <-chan introspection.StateChange[*domain.State] {
	ch := make(chan introspection.StateChange[*domain.State], 10)
	
	r.stateMu.Lock()
	r.watchers = append(r.watchers, ch)
	r.stateMu.Unlock()
	
	go func() {
		<-ctx.Done()
		
		// Safe removal using copy-and-swap to avoid race conditions
		r.stateMu.Lock()
		newWatchers := make([]chan introspection.StateChange[*domain.State], 0, len(r.watchers)-1)
		for _, w := range r.watchers {
			if w != ch {
				newWatchers = append(newWatchers, w)
			}
		}
		r.watchers = newWatchers
		r.stateMu.Unlock()
		
		close(ch)
	}()
	return ch
}

func (r *Runner) broadcastState(newState *domain.State) {
	if newState == nil {
		return
	}
	
	// Snapshot for isolation
	snapshot := newState.Snapshot()
	timestamp := time.Now()
	
	r.stateMu.Lock()
	defer r.stateMu.Unlock()
	
	oldState := r.lastState
	r.lastState = snapshot
	
	change := introspection.StateChange[*domain.State]{
		ComponentID:   r.SessionID,
		ComponentType: "runner",
		OldState:      oldState,
		NewState:      snapshot,
		Timestamp:     timestamp,
	}
	
	droppedCount := 0
	for _, ch := range r.watchers {
		select {
		case ch <- change:
		default:
			// Non-blocking send to avoid stalling runner
			droppedCount++
		}
	}
	
	// Log dropped events for observability (only if drops occurred)
	if droppedCount > 0 {
		// Use a non-blocking log to avoid recursion if logger is also being watched
		// This is informational, not critical
		_ = droppedCount // Placeholder: integrate with metrics/logging later
	}
}

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
// This method implements the lifecycle.Worker interface (Run(context.Context) error).
func (r *Runner) Run(ctx context.Context) error {
	// 1. Setup Phase
	engine := r.engine
	initialState := r.initialState

	if engine == nil {
		return fmt.Errorf("runner: Engine is required (use WithEngine option)")
	}

	handler := r.resolveHandler()
	interceptor := r.resolveInterceptor(handler)

	state, err := r.resolveInitialState(ctx, engine, initialState)
	if err != nil {
		return err
	}

	lastRenderedID := ""

	// 2. Execution Loop
	for {
		// Check for cancellation before each step
		select {
		case <-ctx.Done():
			r.stateMu.Lock()
			r.finalState = state
			// Don't broadcast on cancellation to avoid deadlocks or closed channels
			r.lastState = state.Snapshot()
			r.stateMu.Unlock()
			return ctx.Err()
		default:
		}

		// Update Observability State
		r.broadcastState(state)

		// A. Render
		actions, _, err := engine.Render(ctx, state)
		if err != nil {
			r.finalState = state
			return fmt.Errorf("render error: %w", err)
		}

		// B. Output
		stepTimeout := r.detectTimeout(actions)
		inputCtx, inputCancel := r.createInputContext(ctx, stepTimeout)
		defer inputCancel()

		needsInput, err := handler.Output(ctx, actions)
		if err != nil {
			r.finalState = state
			return fmt.Errorf("output error: %w", err)
		}

		if state.CurrentNodeID != lastRenderedID {
			lastRenderedID = state.CurrentNodeID
		}

		// C. Input / Tool Execution
		var nextInput any
		var nextState *domain.State

		// If no input requested by actions (needsInput=false) and not waiting for tool,
		// we treat this as an auto-transition (pass-through).
		if !needsInput && state.Status != domain.StatusWaitingForTool && state.Status != domain.StatusRollingBack {
			// Auto-transition with empty input
			nextInput = ""
		} else if state.Status == domain.StatusWaitingForTool || state.Status == domain.StatusRollingBack {
			nextInput, err = r.handleTool(ctx, actions, state, handler, interceptor)
		} else {
			nextInput, nextState, err = r.handleInput(inputCtx, handler, needsInput, engine, state)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			r.finalState = state
			return err
		}

		// If a signal caused a transition, update state and loop immediately
		if nextState != nil {
			state = nextState
			// Auto-Save on Signal Transition
			if err := r.saveState(ctx, r.SessionID, state); err != nil {
				r.finalState = state
				return err
			}
			continue
		}

		// 4. Navigate Phase (Controller)
		nextState, err = engine.Navigate(ctx, state, nextInput)
		if err != nil {
			r.finalState = state
			return fmt.Errorf("navigation error: %w", err)
		}

		if nextState.CurrentNodeID != state.CurrentNodeID {
			r.Logger.Debug("Runner: transition", "from", state.CurrentNodeID, "to", nextState.CurrentNodeID, "input", nextInput)
		}

		// 5. Commit Phase (Persistence)
		if err := r.saveState(ctx, r.SessionID, nextState); err != nil {
			r.finalState = nextState
			return fmt.Errorf("critical persistence error: %w", err)
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
	r.finalState = state
	r.stateMu.Lock()
	r.lastState = state.Snapshot()
	r.stateMu.Unlock()
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
	th := NewTextHandler(os.Stdout)
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

func (r *Runner) handleInput(
	ctx context.Context,
	handler IOHandler,
	needsInput bool,
	engine *trellis.Engine,
	currentState *domain.State,
) (any, *domain.State, error) {
	// If we got here, we are in Interactive Mode BUT needsInput is false?
	// This shouldn't happen if logic above handles auto-transition.
	// But in case handleInput IS called with needsInput=false:
	if !needsInput {
		return "", nil, nil
	}

	// Result channels for the input goroutine
	type inputResult struct {
		val any
		err error
	}
	resChan := make(chan inputResult, 1)

	// Create a cancelable context for the input goroutine
	// This allows us to cancel the goroutine when an interrupt happens
	inputCtx, cancelInput := context.WithCancel(ctx)
	defer cancelInput() // Always clean up

	// Launch blocking input in a separate goroutine
	go func() {
		val, err := handler.Input(inputCtx)
		resChan <- inputResult{val, err}
	}()

	var val any
	var err error
	var intercepted bool

	// Wait for input, interruption, or context cancellation
	select {
	case res := <-resChan:
		val, err = res.val, res.err
		// If handler returned context error, treat as a signal situation
		if err != nil && (errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded)) {
			return r.handleSignal(context.Background(), engine, currentState, domain.SignalTimeout)
		}
		if err != nil && (errors.Is(err, context.Canceled) || errors.Is(ctx.Err(), context.Canceled)) {
			// Check if this was an Intercept or just regular cancel
			select {
			case <-r.InterruptSource:
				intercepted = true
			default:
				return nil, nil, err
			}
		}
	case <-r.InterruptSource:
		// CRITICAL: Cancel the input goroutine's context to prevent it from consuming the next input
		cancelInput()
		intercepted = true
	case <-ctx.Done():
		// CRITICAL: Cancel the input goroutine's context
		cancelInput()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return r.handleSignal(context.Background(), engine, currentState, domain.SignalTimeout)
		}
		return nil, nil, ctx.Err()
	}

	// Handle Interception (Lifecycle InterceptEvent)
	if intercepted {
		return r.handleSignal(context.Background(), engine, currentState, domain.SignalInterrupt)
	}

	// Handle Input Result
	if err != nil {
		if err == io.EOF {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("input error: %w", err)
	}

	return val, nil, nil
}

// handleSignal encapsulates the logic for triggering a signal and handling fallbacks.
func (r *Runner) handleSignal(ctx context.Context, engine *trellis.Engine, state *domain.State, signalName string) (any, *domain.State, error) {
	r.Logger.Debug("Runner: Triggering signal", "signal", signalName)
	nextState, sigErr := engine.Signal(ctx, state, signalName)
	if sigErr == nil {
		r.Logger.Debug("Runner: Signal transition success", "signal", signalName)
		if signalName == domain.SignalInterrupt {
			lifecycle.ResetSignalCount(ctx)
		}
		return nil, nextState, nil
	}

	// Fallback messages for unhandled signals
	switch signalName {
	case domain.SignalTimeout:
		return nil, nil, fmt.Errorf("timeout exceeded and no 'on_signal.timeout' handler defined")
	case domain.SignalInterrupt:
		return nil, nil, fmt.Errorf("interrupted")
	default:
		return nil, nil, fmt.Errorf("signal %s failed: %w", signalName, sigErr)
	}
}
