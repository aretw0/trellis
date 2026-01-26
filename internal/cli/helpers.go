package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/aretw0/lifecycle"
	"github.com/aretw0/lifecycle/pkg/signal"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/logging"
	"github.com/aretw0/trellis/internal/presentation/tui"
	"github.com/aretw0/trellis/pkg/adapters/file"
	"github.com/aretw0/trellis/pkg/adapters/memory"
	"github.com/aretw0/trellis/pkg/adapters/redis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/ports"
	"github.com/aretw0/trellis/pkg/runner"
	"github.com/aretw0/trellis/pkg/session"
)

type SignalContext = signal.Context

// NewSignalContext creates a context that is cancelled on SIGTERM (standard termination).
// It captures SIGINT (Interrupt) separately to allow the state machine to handle it.
// Delegated to lifecycle library.
func NewSignalContext(parent context.Context) *SignalContext {
	return lifecycle.NewSignalContext(parent)
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
func createRunnerOptions(logger *slog.Logger, headless bool, sessionID string, store ports.StateStore, jsonMode bool, ioHandler *runner.IOHandler) []runner.Option {

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

func isInterrupted(err error) bool {
	return lifecycle.IsInterrupted(err)
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
func setupPersistence(opts RunOptions, logger *slog.Logger) (ports.StateStore, *session.Manager) {
	var store ports.StateStore
	var locker ports.DistributedLocker

	if opts.RedisURL != "" {
		// Use Redis Store & Locker
		storeOpts, err := redis.ParseURL(opts.RedisURL)
		if err == nil {
			rStore := redis.New(storeOpts.Addr, storeOpts.Password, storeOpts.DB)
			// Enable Distributed Locking by default for Redis
			locker = redis.NewLocker(rStore.Client(), "trellis:lock:")
			store = rStore
		} else {
			fmt.Printf("Warning: Invalid Redis URL %q, falling back to FileStore. Error: %v\n", opts.RedisURL, err)
		}
	}

	if store == nil {
		if opts.SessionID != "" {
			store = file.New("") // Uses default .trellis/sessions
		} else {
			// Ephemeral session: Use In-Memory store to prevent Panics when Session Manager tries to Load/Save
			store = memory.NewStore()
		}
	}

	managerOpts := []session.Option{
		session.WithLogger(logger),
	}
	if locker != nil {
		managerOpts = append(managerOpts, session.WithLocker(locker))
	}

	return store, session.NewManager(store, managerOpts...)
}

// ResetSession clears the session data for the given ID.
func ResetSession(sessionID string) {
	if sessionID == "" {
		sessionID = "watch-dev"
	}
	store := file.New("")
	_ = store.Delete(context.Background(), sessionID)
}

// hydrateAndValidateState handles session rehydration and reload guardrails.
func hydrateAndValidateState(ctx context.Context, engine *trellis.Engine, sessionID string, initialContext map[string]any, sessionManager *session.Manager) (*domain.State, bool, error) {
	// We need to know 'loaded' boolean for UI logs ("Resumed" vs "Created").
	// Since LoadOrStart atomically handles creation, we assume "Loaded" if the state exists in the store.

	var state *domain.State
	var loaded bool
	sessionManager.WithLock(ctx, sessionID, func(ctx context.Context) error {
		// 1. Try Load
		s, err := sessionManager.Store().Load(ctx, sessionID)
		if err == nil {
			state = s
			loaded = true
			return nil
		}
		// If not found, create new
		if err != domain.ErrSessionNotFound {
			return err
		}

		// 2. Start New Session
		s, err = engine.Start(ctx, sessionID, initialContext)
		if err != nil {
			return err
		}
		state = s
		loaded = false
		return sessionManager.Store().Save(ctx, sessionID, state)
	})

	// Validation logic...
	if loaded && sessionID != "" {
		// Reload Guardrail...

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
