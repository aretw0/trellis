package cli

import (
	"context"

	"github.com/aretw0/lifecycle"
	"github.com/aretw0/trellis/pkg/runner"
)

// createInteractiveRouter creates a lifecycle router configured for Trellis.
// This centralizes the router setup logic shared between RunSession and RunWatch.
//
// Mode options:
//   - "interactive": Standard TUI mode with explicit command mappings (q/quit only)
//   - "headless": No input source (signals only)
//   - "json": Raw input mode for JSONL processing
//   - "watch": Same as interactive but semantically for watch mode
func createInteractiveRouter(ctx context.Context, ioHandler runner.IOHandler, mode string, interruptSource chan<- struct{}) *lifecycle.Router {
	var routerOpts []lifecycle.InteractiveOption

	// ---------------------------------------------------------
	// 1. Input/Signal Bridging (if IOHandler exists)
	// ---------------------------------------------------------
	if ioHandler != nil {
		feed := func(cmd string) {
			if th, ok := ioHandler.(*runner.TextHandler); ok {
				th.FeedInput(cmd, nil)
			} else if jh, ok := ioHandler.(*runner.JSONHandler); ok {
				jh.FeedInput(cmd, nil)
			}
		}

		// A. Input Bridge: Lifecycle -> IOHandler.FeedInput
		routerOpts = append(routerOpts, lifecycle.WithDefaultHandler(lifecycle.HandlerFunc(func(ctx context.Context, e lifecycle.Event) error {
			switch ev := e.(type) {
			case lifecycle.InputEvent:
				feed(ev.Command)
				return nil
			case lifecycle.LineEvent:
				feed(ev.Line)
				return nil
			case lifecycle.UnknownCommandEvent:
				feed(ev.Command)
				return nil
			}
			return lifecycle.ErrNotHandled
		})))

		// B. Signal Bridge: Visual Feedback + Engine Notification
		// lifecycle manages the actual signal (Cancel/Exit). We notify the IOHandler and the Runner.
		routerOpts = append(routerOpts, lifecycle.WithInterruptHandler(lifecycle.HandlerFunc(func(ctx context.Context, _ lifecycle.Event) error {
			if interruptSource != nil {
				select {
				case interruptSource <- struct{}{}:
				default:
					// Already signaled
				}
			}
			return ioHandler.Signal(ctx, "interrupt", nil)
		})))
	}

	// C. Shutdown Bridge: Lifecycle -> Context Cancellation
	// When "q", "quit", or "exit" is triggered, we need to explicitly shut down the context.
	routerOpts = append(routerOpts, lifecycle.WithShutdown(func() {
		lifecycle.Shutdown(ctx)
	}))

	// ---------------------------------------------------------
	// 2. Mode-Specific Input Configuration
	// ---------------------------------------------------------
	switch mode {
	case "headless":
		// No input, signals only
		routerOpts = append(routerOpts, lifecycle.WithInput(false))

	case "json":
		// RAW MODE: Treat everything as data (JSONL)
		routerOpts = append(routerOpts,
			lifecycle.WithInputOptions(
				lifecycle.WithRawInput(func(line string) {
					lifecycle.Dispatch(ctx, lifecycle.InputEvent{Command: line})
				}),
			),
		)

	case "interactive", "watch":
		// EXPLICIT command mapping: Only q/quit for shutdown
		// This avoids polluting the command space with lifecycle defaults (suspend, resume, exit, terminate, etc.)
		// Trellis flows should own their command space.
		routerOpts = append(routerOpts,
			lifecycle.WithInputOptions(
				lifecycle.WithInputMappings(map[string]lifecycle.Event{
					"q":    lifecycle.ShutdownEvent{Reason: "manual"},
					"quit": lifecycle.ShutdownEvent{Reason: "manual"},
					"exit": lifecycle.ShutdownEvent{Reason: "manual"},
				}),
			),
		)
	}

	return lifecycle.NewInteractiveRouter(routerOpts...)
}
