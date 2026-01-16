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
