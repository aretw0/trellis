package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

// RunWatch executes Trellis in development mode, reloading on file changes.
func RunWatch(repoPath string, sessionID string, debug bool) {
	logger := createLogger(debug)
	tui.PrintBanner(trellis.Version)

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

	// Graceful exit message for the outer loop
	fmt.Println(">>> Watcher stopped. Goodbye!")
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
		logger.Error("Engine initialization failed", "err", err)
		return waitBackoff(sigCh, 2*time.Second)
	}

	// 2. Setup Persistence and Session Management
	store, sessionManager := setupPersistence(sessionID)

	state, loaded, err := hydrateAndValidateState(ctx, engine, sessionID, nil, sessionManager)
	if err != nil {
		logger.Error("State rehydration failed", "err", err)
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
		// Let the outer loop handle the message
		return false
	case <-time.After(d):
		return true
	}
}

func waitForFix(ctx context.Context, engine *trellis.Engine, sigCh chan os.Signal) bool {
	watchCh, _ := engine.Watch(ctx)
	select {
	case <-sigCh:
		// Let the outer loop handle the message
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

		// Only log actual errors, not cancellation noise
		logger.Error("Runtime error", "err", err)
	}

	if watchCh != nil {
		if err == nil {
			fmt.Printf("\n>>> Flow finished at node '%s'\n", nodeID)
		}
		logger.Info("Flow finished, waiting for changes")
		select {
		case <-sigCh:
			logger.Info("Stopping watcher (signal received)")
			return false
		case <-ctx.Done():
			return true
		}
	}
	return true
}
