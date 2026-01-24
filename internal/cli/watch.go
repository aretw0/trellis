package cli

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

// RunWatch executes Trellis in development mode, reloading on file changes.
func RunWatch(opts RunOptions) {
	logger := createLogger(opts.Debug)
	tui.PrintBanner(trellis.Version)

	// Default session for watch mode to enable Stateful Hot Reload by default
	// We scope it by path hash to prevent collisions between projects.
	if opts.SessionID == "" {
		hash := md5.Sum([]byte(opts.RepoPath))
		opts.SessionID = fmt.Sprintf("watch-%x", hash[:4])
	}

	if opts.Fresh {
		ResetSession(opts.SessionID)
	}

	logger.Info("Starting Watcher", "path", opts.RepoPath, "session_id", opts.SessionID)
	printSystemMessage("Watcher at '%s' session.", opts.SessionID)

	sigCtx := NewSignalContext(context.Background())
	defer sigCtx.Cancel()

	// Reuse the same IO handler to avoid multiple Stdin Pumps (ghost readers)
	// We use the interruptible reader to stop blocking on Stdin during reload
	ioHandler := runner.NewTextHandler(os.Stdin, os.Stdout)
	ioHandler.Renderer = tui.NewRenderer()

	for {
		if !runWatchIteration(sigCtx, opts, ioHandler) {
			break
		}
		logger.Info("Watcher restarting")
	}

	// Graceful exit message for the outer loop (Interrupted message handled by logCompletion)
	// Ensure we exit cleanly
	os.Exit(0)
}

func runWatchIteration(parentCtx *SignalContext, opts RunOptions, ioHandler runner.IOHandler) bool {
	// Create a child context that can be cancelled by reload (without cancelling the parent signal context)
	// But catching parent signals
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	logger := createLogger(opts.Debug)

	// 1. Initialize Engine
	engine, err := createEngine(opts, logger)
	if err != nil {
		logger.Error("Engine initialization failed", "err", err)
		// We can't reuse waitBackoff easily with context, so manual check
		select {
		case <-parentCtx.Done():
			return false
		case <-time.After(2 * time.Second):
			return true
		}
	}

	// 2. Setup Persistence and Session Management
	store, sessionManager := setupPersistence(opts, logger)

	state, loaded, err := hydrateAndValidateState(ctx, engine, opts.SessionID, nil, sessionManager)
	if err != nil {
		logger.Error("State rehydration failed", "err", err)
		// Wait for fix
		watchCh, _ := engine.Watch(ctx)
		select {
		case <-parentCtx.Done():
			return false
		case _, ok := <-watchCh:
			return ok
		}
	}

	if loaded && opts.SessionID != "" {
		logger.Info("Session rehydrated", "session_id", opts.SessionID, "node_id", state.CurrentNodeID)
	}

	// 3. Setup Watcher & Runner
	watchCh, _ := engine.Watch(ctx)

	rOpts := createRunnerOptions(logger, false, opts.SessionID, store, false, &ioHandler)
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
		case event, ok := <-watchCh:
			if ok {
				logger.Info("Change detected, triggering reload", "event", event)
				if !opts.Debug {
					fmt.Printf("\n")
				}
				printSystemMessage("Change detected in '%s'.", event)
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
		printSystemMessage("Resuming at '%s' node...", state.CurrentNodeID)
	}

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
	case <-parentCtx.Done():
		runCancel() // Stop the runner
		<-doneCh    // Wait for it to exit
		logCompletion(state.CurrentNodeID, context.Canceled, opts.Debug, true, false, parentCtx.Signal())
		logger.Info("Stopping watcher (signal received)", "signal", parentCtx.Signal())
		return false
	case <-reloadCh:
		runCancel() // Stop the runner
		<-doneCh    // Wait for it to exit
		return true // Continue to next iteration
	case res := <-doneCh:
		return handleRunCompletion(runCtx, res.state, res.err, watchCh, parentCtx, logger, opts.Debug)
	}
}

func handleRunCompletion(ctx context.Context, finalState *domain.State, err error, watchCh <-chan string, parentCtx *SignalContext, logger *slog.Logger, debug bool) bool {
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
			logCompletion(nodeID, nil, debug, false, false, nil)
			printSystemMessage("Waiting for changes...")
		}
		logger.Info("Flow finished, waiting for changes")
		select {
		case <-parentCtx.Done():
			logCompletion(nodeID, context.Canceled, debug, false, false, parentCtx.Signal())
			logger.Info("Stopping watcher (signal received)")
			return false
		case <-ctx.Done():
			return true
		}
	}
	return true
}
