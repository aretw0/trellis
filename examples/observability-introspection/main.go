package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/lifecycle"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/observability"
	"github.com/aretw0/trellis/pkg/runner"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// 1. Initialize Engine
	// We point to current directory which contains start.md
	eng, err := trellis.New(".", trellis.WithLogger(logger))
	if err != nil {
		logger.Error("Failed to init engine", "error", err)
		os.Exit(1)
	}

	// 2. Initialize Runner (Observable)
	h := runner.NewTextHandler(os.Stdout)

	run := runner.NewRunner(
		runner.WithEngine(eng),
		runner.WithInputHandler(h),
		runner.WithLogger(logger),
	)

	// 3. Create context for lifecycle management
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup Lifecycle Integration for Input
	go func() {
		// Simple consolidated stdin reader (or use lifecycle if fully integrated)
		// For this example, we'll just read from stdin and feed the handler
		// In a real app, use internal/cli/router.go logic
		buf := make([]byte, 1024)
		for {
			// Check for cancellation to avoid goroutine leak
			select {
			case <-ctx.Done():
				h.FeedInput("", ctx.Err())
				return
			default:
			}
			
			n, err := os.Stdin.Read(buf)
			if err != nil {
				h.FeedInput("", err)
				return
			}
			h.FeedInput(string(buf[:n]), nil)
		}
	}()

	// 4. Create Introspection Watcher
	// Use our helper from pkg/observability to combine watchers
	aggregator := observability.NewAggregator()
	aggregator.AddWatcher(run)

	// 5. Start Monitoring (simulating an external observer like an HTTP endpoint)

	// Get the aggregated stream of changes
	changes := aggregator.Watch(ctx)

	go func() {
		fmt.Println("🔎 Observer started... (Press Ctrl+C to exit)")

		for {
			select {
			case <-ctx.Done():
				return
			case snap, ok := <-changes:
				if !ok {
					return
				}

				// Snapshot payload is *domain.State (if from Runner)
				if state, ok := snap.Payload.(*domain.State); ok {
					fmt.Printf("\n[👀 Introspection] Node: %-10s | Status: %-10s | History: %d steps\n",
						state.CurrentNodeID,
						state.Status,
						len(state.History),
					)
				}
			}
		}
	}()

	// 6. Run Execution
	// Using lifecycle context for proper signal handling
	runCtx := lifecycle.NewSignalContext(ctx)
	if err := run.Run(runCtx); err != nil {
		logger.Error("Execution failed", "error", err)
		os.Exit(1)
	}

	// Print final state
	finalState := run.State()
	fmt.Printf("\n[✅ Done] Final Node: %s\n", finalState.CurrentNodeID)
}
