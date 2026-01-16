package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

// RunSession executes a single session of Trellis.
func RunSession(repoPath string, headless bool, jsonMode bool, debug bool, initialContext map[string]any, sessionID string) error {
	logger := createLogger(debug)

	if !jsonMode && !headless {
		tui.PrintBanner()
	}

	// Initialize Engine
	engineOpts := []trellis.Option{}
	if debug {
		engineOpts = append(engineOpts, trellis.WithLogger(logger))
		engineOpts = append(engineOpts, trellis.WithLifecycleHooks(createDebugHooks(logger)))
	}

	engine, err := trellis.New(repoPath, engineOpts...)
	if err != nil {
		return fmt.Errorf("error initializing trellis: %w", err)
	}

	// Setup Persistence
	store, sessionManager := setupPersistence(sessionID)

	// Hydrate State
	ctx := context.Background()
	state, loaded, err := hydrateAndValidateState(ctx, engine, sessionID, initialContext, sessionManager)
	if err != nil {
		return fmt.Errorf("failed to init session: %w", err)
	}

	logSessionStatus(logger, sessionID, state.CurrentNodeID, loaded, jsonMode || headless)

	// Setup Runner
	runnerOpts := createRunnerOptions(logger, headless, sessionID, store, jsonMode)
	r := runner.NewRunner(runnerOpts...)

	// Execute
	finalState, runErr := r.Run(ctx, engine, state)

	// Log Completion
	// Use finalState ID if available, otherwise fallback to initial state (for early errors)
	completionNodeID := state.CurrentNodeID
	if finalState != nil {
		completionNodeID = finalState.CurrentNodeID
	}
	logCompletion(completionNodeID, sessionID, runErr, jsonMode || headless)

	return handleExecutionError(runErr)
}

func createLogger(debug bool) *slog.Logger {
	if debug {
		return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func logSessionStatus(logger *slog.Logger, sessionID, nodeID string, loaded, quiet bool) {
	if loaded {
		logger.Info("Session Resumed", "session_id", sessionID, "node", nodeID)
		if !quiet {
			fmt.Printf(">>> Resuming session '%s' at node '%s'...\n", sessionID, nodeID)
		}
	} else if sessionID != "" {
		logger.Info("Session Created", "session_id", sessionID)
		if !quiet {
			fmt.Printf(">>> Created new session '%s'...\n", sessionID)
		}
	}
}

// RunWatch executes Trellis in development mode, reloading on file changes.
func RunWatch(repoPath string, sessionID string, debug bool) {
	logger := createLogger(debug)
	tui.PrintBanner()

	// Default session for watch mode to enable Stateful Hot Reload by default
	if sessionID == "" {
		sessionID = "watch-dev"
	}

	logger.Info("Starting Watcher", "path", repoPath, "session_id", sessionID)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Reuse the same IO handler to avoid multiple Stdin Pumps (ghost readers)
	// We use the interruptible reader to stop blocking on Stdin during reload
	ioHandler := runner.NewTextHandler(os.Stdin, os.Stdout)
	ioHandler.Renderer = tui.NewRenderer()

	for {
		if !runWatchIteration(repoPath, sessionID, debug, sigCh, ioHandler) {
			break
		}
		logger.Info("Watcher restarting")
	}
}

func runWatchIteration(repoPath string, sessionID string, debug bool, sigCh chan os.Signal, ioHandler runner.IOHandler) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := createLogger(debug)

	// 1. Initialize Engine
	engineOpts := []trellis.Option{}
	if debug {
		engineOpts = append(engineOpts, trellis.WithLogger(logger))
		engineOpts = append(engineOpts, trellis.WithLifecycleHooks(createDebugHooks(logger)))
	}

	engine, err := trellis.New(repoPath, engineOpts...)
	if err != nil {
		logger.Error("Engine initialization failed", "error", err)
		return waitBackoff(sigCh, 2*time.Second)
	}

	// 2. Setup Persistence and Session Management
	store, sessionManager := setupPersistence(sessionID)

	state, loaded, err := hydrateAndValidateState(ctx, engine, sessionID, nil, sessionManager)
	if err != nil {
		logger.Error("State rehydration failed", "error", err)
		return waitForFix(ctx, engine, sigCh)
	}

	if loaded && sessionID != "" {
		logger.Info("Session rehydrated", "session_id", sessionID, "node_id", state.CurrentNodeID)
	}

	// 3. Setup Watcher & Runner
	watchCh, _ := engine.Watch(ctx)

	rOpts := createRunnerOptions(logger, false, sessionID, store, false)
	// Use the shared handler
	rOpts = append(rOpts, runner.WithInputHandler(ioHandler))

	r := runner.NewRunner(rOpts...)

	// 4. Start Watcher Routine
	reloadCh := make(chan struct{}, 1)
	go func() {
		if watchCh == nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case _, ok := <-watchCh:
			if ok {
				logger.Info("Change detected, triggering reload")
				// Delay slightly to ensure file system is stable
				time.Sleep(100 * time.Millisecond)
				reloadCh <- struct{}{}
				cancel()
			}
		}
	}()

	// 5. Run
	if loaded {
		logger.Debug("Resuming node execution", "node_id", state.CurrentNodeID)
		if debug {
			fmt.Printf(">>> Resuming at node '%s'...\n", state.CurrentNodeID)
		}
	}

	fmt.Printf("--- Hot Reload Active (Node: %s) ---\n", state.CurrentNodeID)

	// Use a dedicated context for this run iteration that can be cancelled by reloads
	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()

	doneCh := make(chan struct {
		state *domain.State
		err   error
	}, 1)
	go func() {
		s, err := r.Run(runCtx, engine, state)
		doneCh <- struct {
			state *domain.State
			err   error
		}{s, err}
	}()

	select {
	case <-sigCh:
		runCancel() // Stop the runner
		<-doneCh    // Wait for it to exit
		logger.Info("Stopping watcher (signal received)")
		return false
	case <-reloadCh:
		runCancel() // Stop the runner
		<-doneCh    // Wait for it to exit
		return true // Continue to next iteration
	case res := <-doneCh:
		return handleRunCompletion(runCtx, res.state, res.err, watchCh, sigCh, logger)
	}
}

func waitBackoff(sigCh chan os.Signal, d time.Duration) bool {
	select {
	case <-sigCh:
		fmt.Println("\nStopping watcher...")
		return false
	case <-time.After(d):
		return true
	}
}

func waitForFix(ctx context.Context, engine *trellis.Engine, sigCh chan os.Signal) bool {
	watchCh, _ := engine.Watch(ctx)
	select {
	case <-sigCh:
		fmt.Println("\nStopping watcher...")
		return false
	case _, ok := <-watchCh:
		return ok
	}
}

func handleRunCompletion(ctx context.Context, finalState *domain.State, err error, watchCh <-chan string, sigCh chan os.Signal, logger *slog.Logger) bool {
	nodeID := ""
	if finalState != nil {
		nodeID = finalState.CurrentNodeID
	}

	if err != nil {
		// If the context was cancelled, it's a reload request
		if errors.Is(err, context.Canceled) {
			return true // Continue to next iteration
		}

		if isInterrupted(err) {
			return false // User stop
		}
		logger.Error("Runtime error", "error", err)
	}

	if watchCh != nil {
		if err == nil {
			fmt.Printf("\n>>> Flow finished at node '%s'\n", nodeID)
		}
		logger.Info("Flow finished, waiting for changes")
		select {
		case <-sigCh:
			logger.Info("Stopping watcher (signal received)")
			fmt.Println("\nStopping watcher...")
			return false
		case <-ctx.Done():
			return true
		}
	}
	return true
}

func isInterrupted(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.Canceled) ||
		err.Error() == "interrupted" ||
		err.Error() == "input error: interrupted" ||
		errors.Is(err, io.EOF) ||
		(errors.Unwrap(err) != nil && isInterrupted(errors.Unwrap(err)))
}

func handleExecutionError(err error) error {
	if err == nil {
		return nil
	}
	if isInterrupted(err) {
		return nil // Exit 0 for interruptions
	}
	return err
}

func logCompletion(nodeID string, sessionID string, err error, quiet bool) {
	if quiet {
		return
	}
	if err == nil {
		fmt.Printf("\n>>> Flow finished at node '%s'\n", nodeID)
	} else if isInterrupted(err) {
		fmt.Print("\n")
		if sessionID != "" {
			fmt.Printf(">>> Session saved at node '%s'. Goodbye!\n", nodeID)
		} else {
			fmt.Println(">>> Interrupted. Goodbye!")
		}
	}
}

// setupPersistence initializes the state store and session manager.
func setupPersistence(sessionID string) (*adapters.FileStore, *runner.SessionManager) {
	var store *adapters.FileStore
	if sessionID != "" {
		store = adapters.NewFileStore("") // Uses default .trellis/sessions
	}
	return store, runner.NewSessionManager(store)
}

// ResetSession clears the session data for the given ID.
func ResetSession(sessionID string) {
	if sessionID == "" {
		sessionID = "watch-dev"
	}
	store := adapters.NewFileStore("")
	_ = store.Delete(context.Background(), sessionID)
}

// hydrateAndValidateState handles session rehydration and reload guardrails.
func hydrateAndValidateState(ctx context.Context, engine *trellis.Engine, sessionID string, initialContext map[string]any, sessionManager *runner.SessionManager) (*domain.State, bool, error) {
	state, loaded, err := sessionManager.LoadOrStart(ctx, engine, sessionID, initialContext)
	if err != nil {
		return nil, false, err
	}

	if loaded && sessionID != "" {
		// Reload Guardrail: Check if node still exists and handle type mismatches
		nodes, _ := engine.Inspect()
		var node *domain.Node
		for _, n := range nodes {
			if n.ID == state.CurrentNodeID {
				node = &n
				break
			}
		}

		if state.Status == domain.StatusWaitingForTool && (node == nil || node.Type != domain.NodeTypeTool) {
			fmt.Printf(">>> Resetting status from WaitingForTool to Active (Node type changed or missing)\n")
			state.Status = domain.StatusActive
			state.PendingToolCall = ""
		}
	}

	return state, loaded, nil
}

// createRunnerOptions prepares the functional options for the Runner.
func createRunnerOptions(logger *slog.Logger, headless bool, sessionID string, store *adapters.FileStore, jsonMode bool) []runner.Option {
	opts := []runner.Option{
		runner.WithLogger(logger),
		runner.WithHeadless(headless),
	}

	if sessionID != "" {
		opts = append(opts, runner.WithSessionID(sessionID))
		opts = append(opts, runner.WithStore(store))
	}

	if jsonMode {
		opts = append(opts, runner.WithInputHandler(runner.NewJSONHandler(os.Stdin, os.Stdout)))
	} else if !headless {
		opts = append(opts, runner.WithRenderer(tui.NewRenderer()))
	}

	return opts
}

func createDebugHooks(logger *slog.Logger) domain.LifecycleHooks {
	return domain.LifecycleHooks{
		OnNodeEnter: func(ctx context.Context, e *domain.NodeEvent) {
			logger.Debug("Enter Node", "node_id", e.NodeID, "type", e.NodeType)
		},
		OnNodeLeave: func(ctx context.Context, e *domain.NodeEvent) {
			logger.Debug("Leave Node", "node_id", e.NodeID)
		},
		OnToolCall: func(ctx context.Context, e *domain.ToolEvent) {
			logger.Debug("Tool Call", "tool_name", e.ToolName)
		},
		OnToolReturn: func(ctx context.Context, e *domain.ToolEvent) {
			if e.IsError {
				logger.Debug("Tool Return (Error)", "tool_name", e.ToolName, "error", e.Output)
			} else {
				logger.Debug("Tool Return (Success)", "tool_name", e.ToolName)
			}
		},
	}
}

// InterruptibleReader wraps an io.Reader (like os.Stdin) and checks for a cancellation signal.
type InterruptibleReader struct {
	base   io.Reader
	cancel <-chan struct{}
}

func NewInterruptibleReader(base io.Reader, cancel <-chan struct{}) *InterruptibleReader {
	return &InterruptibleReader{
		base:   base,
		cancel: cancel,
	}
}

func (r *InterruptibleReader) Read(p []byte) (n int, err error) {
	// Check before blocking
	select {
	case <-r.cancel:
		return 0, errors.New("interrupted")
	default:
	}

	// Read (This blocks!)
	n, err = r.base.Read(p)

	// Check after returning
	select {
	case <-r.cancel:
		return 0, errors.New("interrupted")
	default:
	}
	return n, err
}
