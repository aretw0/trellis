package cli

import (
	"context"
	"fmt"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/runner"
)

// RunSession executes a single session of Trellis.
func RunSession(repoPath string, headless bool, jsonMode bool, debug bool, initialContext map[string]any, sessionID string) error {
	logger := createLogger(debug)

	if !jsonMode && !headless {
		tui.PrintBanner(trellis.Version)
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

	// Setup signal handling
	// Use the unified SignalContext helper
	sigCtx := NewSignalContext(context.Background())
	defer sigCtx.Cancel()

	// Hydrate State
	state, loaded, err := hydrateAndValidateState(sigCtx, engine, sessionID, initialContext, sessionManager)
	if err != nil {
		return fmt.Errorf("failed to init session: %w", err)
	}

	logSessionStatus(logger, sessionID, state.CurrentNodeID, loaded, jsonMode || headless)

	// Setup Runner
	runnerOpts := createRunnerOptions(logger, headless, sessionID, store, jsonMode, nil)
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

	logCompletion(completionNodeID, runErr, debug, true, jsonMode || headless, sigCtx.Signal())

	return handleExecutionError(runErr)
}
