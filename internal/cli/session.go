package cli

import (
	"context"
	"fmt"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/adapters/process"
	"github.com/aretw0/trellis/pkg/runner"
)

// RunSession executes a single session of Trellis.
func RunSession(opts RunOptions, initialContext map[string]any) error {
	logger := createLogger(opts.Debug)

	if !opts.JSON && !opts.Headless {
		tui.PrintBanner(trellis.Version)
	}

	// Initialize Engine
	engineOpts := []trellis.Option{
		trellis.WithDefaultErrorNode("error"),
	}
	if opts.Debug {
		engineOpts = append(engineOpts, trellis.WithLogger(logger))
		engineOpts = append(engineOpts, trellis.WithLifecycleHooks(createDebugHooks(logger)))
	}

	engine, err := trellis.New(opts.RepoPath, engineOpts...)
	if err != nil {
		return fmt.Errorf("error initializing trellis: %w", err)
	}

	// Setup Persistence
	store, sessionManager := setupPersistence(opts, logger)

	// Setup signal handling
	// Use the unified SignalContext helper
	sigCtx := NewSignalContext(context.Background())
	defer sigCtx.Cancel()

	// Hydrate State
	state, loaded, err := hydrateAndValidateState(sigCtx, engine, opts.SessionID, initialContext, sessionManager)
	if err != nil {
		return fmt.Errorf("failed to init session: %w", err)
	}

	logSessionStatus(logger, opts.SessionID, state.CurrentNodeID, loaded, opts.JSON || opts.Headless)

	// Setup Process Runner (Tooling)
	toolConfig, err := process.LoadTools(opts.ToolsPath)
	if err != nil {
		// Only fail if user explicitly pointed to a file that is broken.
		// If default missing, LoadTools returns empty.
		// However, process.LoadTools logic I implemented returns error for any read failure except IsNotExist.
		// Let's assume critical failure if config is bad.
		logger.Warn("Failed to load tools config", "path", opts.ToolsPath, "error", err)
	}

	procRunner := process.NewRunner(
		process.WithRegistry(toolConfig),
		process.WithInlineExecution(opts.UnsafeInline),
	)

	// Setup Runner
	runnerOpts := createRunnerOptions(logger, opts.Headless, opts.SessionID, store, opts.JSON, nil)
	runnerOpts = append(runnerOpts, runner.WithToolRunner(procRunner))

	r := runner.NewRunner(runnerOpts...)

	// Execute
	finalState, runErr := r.Run(sigCtx, engine, state)

	// Log Completion
	completionNodeID := state.CurrentNodeID
	if finalState != nil {
		completionNodeID = finalState.CurrentNodeID
	}

	// If context was canceled (signal received), ensure runErr reflects it if it doesn't already
	if sigCtx.Err() != nil && runErr == nil {
		runErr = sigCtx.Err()
	}

	logCompletion(completionNodeID, runErr, opts.Debug, true, opts.JSON || opts.Headless, sigCtx.Signal())

	return handleExecutionError(runErr)
}
