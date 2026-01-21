package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/adapters"
	"github.com/aretw0/trellis/internal/logging"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

// SignalContext wraps a context and captures the signal that cancelled it.
type SignalContext struct {
	context.Context
	Cancel func()
	start  sync.Once
	stop   sync.Once
	sigCh  chan os.Signal
	sigVal os.Signal
	mu     sync.Mutex
}

// NewSignalContext creates a context that is cancelled on SIGINT or SIGTERM.
// It acts as a drop-in replacement for signal.NotifyContext but allows retrieving the signal.
func NewSignalContext(parent context.Context) *SignalContext {
	ctx, cancel := context.WithCancel(parent)
	sc := &SignalContext{
		Context: ctx,
		Cancel:  cancel,
		sigCh:   make(chan os.Signal, 1),
	}

	sc.start.Do(func() {
		signal.Notify(sc.sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			select {
			case sig := <-sc.sigCh:
				sc.mu.Lock()
				sc.sigVal = sig
				sc.mu.Unlock()
				sc.Cancel()
			case <-sc.Context.Done():
				// Context cancelled elsewhere
			}
			sc.stop.Do(func() {
				signal.Stop(sc.sigCh)
			})
		}()
	})

	return sc
}

// Signal returns the signal that caused the context to be cancelled, or nil.
func (sc *SignalContext) Signal() os.Signal {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.sigVal
}

// createLogger configures the application logger.
// In debug mode, it writes to Stderr (to separate from Stdout flow UI).
func createLogger(debug bool) *slog.Logger {
	if debug {
		return logging.New(slog.LevelDebug)
	}
	return logging.NewNop()
}

// printSystemMessage prints a standardized system message to stdout.
func printSystemMessage(format string, args ...any) {
	fmt.Printf(">>> %s\n", fmt.Sprintf(format, args...))
}

func logSessionStatus(logger *slog.Logger, sessionID, nodeID string, loaded, quiet bool) {
	if loaded {
		logger.Info("Session Resumed", "session_id", sessionID, "node", nodeID)
		if !quiet {
			printSystemMessage("Resuming at '%s' node...", nodeID)
		}
	} else if sessionID != "" {
		logger.Info("Session Created", "session_id", sessionID)
		if !quiet {
			printSystemMessage("Session '%s' active.", sessionID)
		}
	}
}

// createRunnerOptions prepares the functional options for the Runner.
func createRunnerOptions(logger *slog.Logger, headless bool, sessionID string, store *adapters.FileStore, jsonMode bool, ioHandler *runner.IOHandler) []runner.Option {
	opts := []runner.Option{
		runner.WithLogger(logger),
		runner.WithHeadless(headless),
	}

	if sessionID != "" {
		opts = append(opts, runner.WithSessionID(sessionID))
		opts = append(opts, runner.WithStore(store))
	}

	if jsonMode {
		opts = append(opts, runner.WithInputHandler(runner.NewJSONHandler(os.Stdin, os.Stdout)))
	} else if ioHandler != nil {
		opts = append(opts, runner.WithInputHandler(*ioHandler))
	} else if !headless {
		opts = append(opts, runner.WithRenderer(tui.NewRenderer()))
	}

	return opts
}

func createDebugHooks(logger *slog.Logger) domain.LifecycleHooks {
	return domain.LifecycleHooks{
		OnNodeEnter: func(ctx context.Context, e *domain.NodeEvent) {
			logger.Debug("Enter Node", "node_id", e.NodeID, "type", e.NodeType)
		},
		OnNodeLeave: func(ctx context.Context, e *domain.NodeEvent) {
			logger.Debug("Leave Node", "node_id", e.NodeID)
		},
		OnToolCall: func(ctx context.Context, e *domain.ToolEvent) {
			logger.Debug("Tool Call", "tool_name", e.ToolName)
		},
		OnToolReturn: func(ctx context.Context, e *domain.ToolEvent) {
			if e.IsError {
				logger.Debug("Tool Return (Error)", "tool_name", e.ToolName, "err", e.Output)
			} else {
				logger.Debug("Tool Return (Success)", "tool_name", e.ToolName)
			}
		},
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

func isInterrupted(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, context.Canceled) ||
		err.Error() == "interrupted" ||
		err.Error() == "input error: interrupted" ||
		errors.Is(err, io.EOF) ||
		(errors.Unwrap(err) != nil && isInterrupted(errors.Unwrap(err)))
}

func handleExecutionError(err error) error {
	if err == nil {
		return nil
	}
	if isInterrupted(err) {
		return nil // Exit 0 for interruptions
	}
	return err
}

func logCompletion(nodeID string, err error, debug bool, promptActive bool, quiet bool, sig os.Signal) {
	if quiet {
		return
	}
	if err == nil {
		printSystemMessage("Finished at '%s' node.", nodeID)
		return
	}

	if isInterrupted(err) {
		// Aesthetic: Print [CTRL+C] simulation if likely from user via SIGINT
		if sig == os.Interrupt {
			if debug {
				// Debug mode: Logs likely interrupted the prompt line. Restore context.
				fmt.Printf("> [CTRL+C]\n")
			} else {
				if promptActive {
					// Normal mode, Input active: Clean UI, append to existing prompt.
					fmt.Printf("[CTRL+C]\n")
				} else {
					// Normal mode, Idle: Create prompt for consistency.
					fmt.Printf("> [CTRL+C]\n")
				}
			}
			printSystemMessage("Interrupted at '%s' node.", nodeID)
		} else if sig != nil {
			// SIGTERM or others
			fmt.Printf("\n")
			printSystemMessage("Terminated at '%s' node.", nodeID)
		} else {
			// clean exit without specific signal (e.g. context cancel during reload)
			// usually handled elsewhere, but if we get here:
			fmt.Printf("\n")
			printSystemMessage("Interrupted at '%s' node.", nodeID)
		}
	}
}

// setupPersistence initializes the state store and session manager.
func setupPersistence(sessionID string) (*adapters.FileStore, *runner.SessionManager) {
	var store *adapters.FileStore
	if sessionID != "" {
		store = adapters.NewFileStore("") // Uses default .trellis/sessions
	}
	return store, runner.NewSessionManager(store)
}

// ResetSession clears the session data for the given ID.
func ResetSession(sessionID string) {
	if sessionID == "" {
		sessionID = "watch-dev"
	}
	store := adapters.NewFileStore("")
	_ = store.Delete(context.Background(), sessionID)
}

// hydrateAndValidateState handles session rehydration and reload guardrails.
func hydrateAndValidateState(ctx context.Context, engine *trellis.Engine, sessionID string, initialContext map[string]any, sessionManager *runner.SessionManager) (*domain.State, bool, error) {
	state, loaded, err := sessionManager.LoadOrStart(ctx, engine, sessionID, initialContext)
	if err != nil {
		return nil, false, err
	}

	if loaded && sessionID != "" {
		// Reload Guardrail: Check if node still exists and handle type mismatches
		nodes, _ := engine.Inspect()
		var node *domain.Node
		for _, n := range nodes {
			if n.ID == state.CurrentNodeID {
				node = &n
				break
			}
		}

		if state.Status == domain.StatusWaitingForTool && (node == nil || node.Type != domain.NodeTypeTool) {
			fmt.Printf(">>> Resetting status from WaitingForTool to Active (Node type changed or missing)\n")
			state.Status = domain.StatusActive
			state.PendingToolCall = ""
		}
	}

	return state, loaded, nil
}
