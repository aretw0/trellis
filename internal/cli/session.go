package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aretw0/lifecycle"
	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/adapters/process"
	"github.com/aretw0/trellis/pkg/runner"
)

// RunSession executes a single session of Trellis.
func RunSession(ctx context.Context, opts RunOptions, initialContext map[string]any) error {
	logger := createLogger(opts.Debug)

	// Unified Logging
	lifecycle.SetLogger(logger)

	// Wrap execution in Lifecycle Job
	return lifecycle.Run(lifecycle.Job(func(ctx context.Context) error {
		// Initialize Engine
		engine, err := createEngine(opts, logger)
		if err != nil {
			return err
		}

		// ---------------------------------------------------------
		// 1. Setup IO Handler (The "Sink")
		// ---------------------------------------------------------
		var ioHandler runner.IOHandler

		if opts.JSON {
			ioHandler = runner.NewJSONHandler(os.Stdout)
		} else if !opts.Headless {
			tui.PrintBanner(trellis.Version)
			ioHandler = runner.NewTextHandler(os.Stdout,
				runner.WithTextHandlerRenderer(tui.NewRenderer()),
			)
		}

		// ---------------------------------------------------------
		// 2. Lifecycle Integration (The "Source")
		// ---------------------------------------------------------
		var mode string
		if opts.Headless && !opts.JSON {
			mode = "headless"
		} else if opts.JSON {
			mode = "json"
		} else {
			mode = "interactive"
		}

		// Prepare Interrupt Bridge
		interruptSource := make(chan struct{}, 1)

		mux := createInteractiveRouter(ctx, ioHandler, mode, interruptSource)

		// Start Lifecycle in Background
		lifecycle.Go(ctx, func(ctx context.Context) error {
			return mux.Start(ctx)
		})

		// ---------------------------------------------------------
		// 5. App Initialization
		// ---------------------------------------------------------

		// Setup Persistence
		store, sessionManager := setupPersistence(opts, logger)

		// Hydrate State
		state, loaded, err := hydrateAndValidateState(ctx, engine, opts.SessionID, initialContext, sessionManager)
		if err != nil {
			return fmt.Errorf("failed to init session: %w", err)
		}

		logSessionStatus(logger, opts.SessionID, state.CurrentNodeID, loaded, opts.JSON || opts.Headless)

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

		// Setup Runner
		runnerOpts := createRunnerOptions(logger, opts.Headless, opts.SessionID, store, ioHandler, interruptSource)
		runnerOpts = append(runnerOpts, runner.WithToolRunner(procRunner))

		r := runner.NewRunner(runnerOpts...)

		// Execute
		finalState, runErr := r.Run(ctx, engine, state)

		// Log Completion
		completionNodeID := state.CurrentNodeID
		if finalState != nil {
			completionNodeID = finalState.CurrentNodeID
		}
		if ctx.Err() != nil && runErr == nil {
			runErr = ctx.Err()
		}
		sig := lifecycle.Signal(ctx)
		logCompletion(completionNodeID, runErr, opts.JSON || opts.Headless, sig)

		return handleExecutionError(runErr)
	}), lifecycle.WithCancelOnInterrupt(false))
}
