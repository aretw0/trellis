package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/presentation/tui"
)

// RunSession executes a single session of Trellis.
func RunSession(repoPath string, headless bool, jsonMode bool) error {
	// Initialize Engine
	engine, err := trellis.New(repoPath)
	if err != nil {
		return fmt.Errorf("error initializing trellis: %w", err)
	}

	// Configure Runner
	runner := trellis.NewRunner()
	runner.Headless = headless

	if jsonMode {
		runner.Handler = trellis.NewJSONHandler(os.Stdin, os.Stdout)
	} else {
		// Default Text Handler
		// We can explicitly set it or let runner default fallback.
		// Explicit is better if we want to attach the renderer.
		// Wait, NewRunner defaults don't set Handler.
		// The Runner.Run logic handles fallback.
		// But we want to attach TUI renderer if not headless.

		// If we rely on fallback, we lose the Ability to set Renderer on the Handler?
		// TextHandler needs the Renderer.
		// So we should instantiate TextHandler here.

		th := trellis.NewTextHandler(os.Stdin, os.Stdout)
		if !headless {
			th.Renderer = tui.NewRenderer()
		}
		runner.Handler = th
	}

	// Execute
	if err := runner.Run(engine); err != nil {
		return fmt.Errorf("error running trellis: %w", err)
	}
	return nil
}

// RunWatch executes Trellis in development mode, reloading on file changes.
func RunWatch(repoPath string) {
	fmt.Printf("Starting Trellis Watcher in %s...\n", repoPath)

	// Catch OS signals for graceful shutdown of the Watch loop
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	for {
		// 1. Initialize Engine
		engine, err := trellis.New(repoPath)
		if err != nil {
			fmt.Printf("Error initializing trellis: %v\nRetrying in 2s...\n", err)

			// Allow exit during backoff
			select {
			case <-sigCh:
				fmt.Println("\nStopping watcher...")
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		// 2. Setup Watcher
		// Context for this specific session
		sessionCtx, cancelRequest := context.WithCancel(context.Background())

		watchCh, err := engine.Watch(sessionCtx)
		if err != nil {
			fmt.Printf("Warning: Watch failed (%v). Running in normal mode.\n", err)
		}

		// 3. Configure Runner with Interruptible Input
		interruptReader := NewInterruptibleReader(os.Stdin, sessionCtx.Done())

		runner := trellis.NewRunner()
		runner.Input = interruptReader
		runner.Output = os.Stdout
		runner.Renderer = tui.NewRenderer()

		// 4. Start Watcher Routine
		go func() {
			if watchCh == nil {
				return
			}
			// Wait for change OR context cancellation (re-render, exit)
			select {
			case <-sessionCtx.Done():
				return // Session ended normally or manually cancelled
			case _, ok := <-watchCh:
				if !ok {
					return // Channel closed (shouldn't happen if ctx is valid, but safety)
				}
				fmt.Println("\n\n>>> Change detected! Reloading... <<<")
				// Debounce slightly to let multiple file writes finish
				time.Sleep(100 * time.Millisecond)
				cancelRequest() // Cancel to stop the runner
			}
		}()

		// 5. Run Execution in a non-blocking check for signals? No, Run blocks.
		// We rely on interruptReader to unblock Run if sessionCtx is cancelled.
		// BUT we also need to handle global SIGINT.

		fmt.Println("--- Session Started ---")

		// Run logic
		doneCh := make(chan error, 1)
		go func() {
			doneCh <- runner.Run(engine)
		}()

		// Wait for Run completion OR Global Signal
		select {
		case <-sigCh:
			// User pressed Ctrl+C
			fmt.Println("\nStopping watcher...")
			cancelRequest() // Stop session
			return          // Exit loop
		case err := <-doneCh:
			// Session finished (either naturally, or via internal cancellation/reload)

			// Normalize error
			isInterrupted := false
			if err != nil {
				isInterrupted = errors.Is(err, context.Canceled) ||
					err.Error() == "input error: interrupted" ||
					errors.Is(err, io.EOF) ||
					// Check for wrapped error
					(errors.Unwrap(err) != nil && errors.Is(errors.Unwrap(err), context.Canceled))

				if !isInterrupted {
					fmt.Printf("Runtime error: %v\n", err)
					// Don't wait here, we will wait below if we decide not to restart immediately
				}
			}

			// If the session finished naturally (flow end), we should NOT restart immediately.
			// We should wait for a file change.
			if !isInterrupted && watchCh != nil {
				fmt.Println("\nFlow finished. Waiting for changes...")
				select {
				case <-sigCh:
					// User pressed Ctrl+C
					fmt.Println("\nStopping watcher...")
					cancelRequest()
					return
				case <-sessionCtx.Done():
					// Watcher signaled change (via cancelRequest)
					// Proceed to restart
				}
			}
		}

		// Cleanup before restart
		cancelRequest()

		// Visual separation
		fmt.Println("--- Restarting ---")
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
