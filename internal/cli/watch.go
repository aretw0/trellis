package cli

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aretw0/lifecycle"
	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/adapters/process"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

// RunWatch executes Trellis in development mode, reloading on file changes.
func RunWatch(ctx context.Context, opts RunOptions) {
	// Wrap execution in Lifecycle Job
	// Watch mode uses WithCancelOnInterrupt(false), delegating interrupt handling to the lifecycle router.
	// The router implements escalation: first interrupt triggers the primary handler (e.g., suspend if declared in flow),
	// subsequent interrupts force shutdown. Actual behavior depends on flow-level signal handlers.
	lifecycle.Run(lifecycle.Job(func(ctx context.Context) error {
		logger := createLogger(opts.Debug)
		tui.PrintBanner(trellis.Version)

		// Default session for watch mode to enable Stateful Hot Reload by default
		if opts.SessionID == "" {
			hash := md5.Sum([]byte(opts.RepoPath))
			opts.SessionID = fmt.Sprintf("watch-%x", hash[:4])
		}

		if opts.Fresh {
			ResetSession(opts.SessionID)
		}

		logger.Info("Starting Watcher", "path", opts.RepoPath, "session_id", opts.SessionID)
		printSystemMessage("Watcher at '%s' session.", opts.SessionID)

		// Reuse the same IO handler to avoid multiple Stdin Pumps (ghost readers)
		ioHandler := runner.NewTextHandler(os.Stdout, runner.WithTextHandlerRenderer(tui.NewRenderer()))

		// Setup Lifecycle Router (uses shared factory)
		interruptSource := make(chan struct{}, 1)
		watcherMux := createInteractiveRouter(ctx, ioHandler, "watch", interruptSource)

		lifecycle.Go(ctx, func(ctx context.Context) error {
			return watcherMux.Start(ctx)
		})

		// Watch Loop
		for {
			if !runWatchIteration(ctx, opts, ioHandler, interruptSource) {
				break
			}
			if ctx.Err() != nil {
				break
			}
			logger.Info("Watcher restarting")
		}

		return nil
	}), lifecycle.WithCancelOnInterrupt(false))
}

func runWatchIteration(parentCtx context.Context, opts RunOptions, ioHandler runner.IOHandler, interruptSource <-chan struct{}) bool {
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
		// Try to wait for a fix if the engine supports watching
		watchCh, wErr := engine.Watch(ctx)
		if wErr != nil {
			logger.Error("Cannot wait for fix: engine does not support watching", "err", wErr)
			return false
		}
		select {
		case <-parentCtx.Done():
			return false
		case event, ok := <-watchCh:
			if ok {
				logger.Info("Attempting recovery after change", "file", event)
			}
			return ok
		}
	}

	if loaded && opts.SessionID != "" {
		logger.Info("Session rehydrated", "session_id", opts.SessionID, "node_id", state.CurrentNodeID)
	}

	// 3. Setup Watcher & Runner
	watchCh, err := engine.Watch(ctx)
	if err != nil {
		logger.Warn("Hot-reload is disabled: engine does not support watching", "err", err)
	}

	// Setup Process Runner
	toolConfig, err := process.LoadTools(opts.ToolsPath)
	if err != nil {
		logger.Warn("Failed to load tools configuration", "path", opts.ToolsPath, "err", err)
	}
	procRunner := process.NewRunner(
		process.WithRegistry(toolConfig),
		process.WithInlineExecution(opts.UnsafeInline),
		process.WithBaseDir(filepath.Dir(opts.ToolsPath)),
	)

	rOpts := createRunnerOptions(logger, false, opts.SessionID, store, ioHandler, interruptSource)
	// Use the shared handler and tool runner
	rOpts = append(rOpts,
		runner.WithInputHandler(ioHandler),
		runner.WithToolRunner(procRunner),
	)

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
		sig := lifecycle.Signal(parentCtx)
		logCompletion(state.CurrentNodeID, context.Canceled, false, sig)
		logger.Info("Stopping watcher (signal received)", "signal", sig)
		return false
	case <-reloadCh:
		runCancel() // Stop the runner
		<-doneCh    // Wait for it to exit
		return true // Continue to next iteration
	case res := <-doneCh:
		return handleRunCompletion(res.state, res.err, watchCh, parentCtx, logger)
	}
}

func handleRunCompletion(finalState *domain.State, err error, watchCh <-chan string, parentCtx context.Context, logger *slog.Logger) bool {
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
			logCompletion(nodeID, nil, false, nil)
			printSystemMessage("Waiting for changes...")
		}
		logger.Info("Flow finished, waiting for changes")
		select {
		case <-parentCtx.Done():
			sig := lifecycle.Signal(parentCtx)
			logCompletion(nodeID, context.Canceled, false, sig)
			logger.Info("Stopping watcher (signal received)")
			return false
		case <-watchCh:
			// File change detected, re-run
			return true
		}
	}
	return true
}
