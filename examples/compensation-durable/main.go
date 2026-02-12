package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/file"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
	"github.com/aretw0/trellis/pkg/session"
)

func main() {
	// Flags to control the simulation
	action := flag.String("action", "run", "Action to perform: run, approve, reject")
	fresh := flag.Bool("fresh", false, "Force a fresh session (delete existing state)")
	flag.Parse()

	// 1. Setup Logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// 2. Define Persistence (File Store)
	wd, _ := os.Getwd()
	// Store sessions in the example directory for visibility
	storePath := filepath.Join(wd, "examples", "compensation-durable", ".sessions")
	fileStore := file.New(storePath)

	// Session Manager (implements ports.StateStore)
	sessMgr := session.NewManager(fileStore)

	// 3. Define Mock Tools
	tools := map[string]func(map[string]any) (any, error){
		"reserve_hotel": func(args map[string]any) (any, error) {
			logger.Info("🏨 ACT: Reserving Hotel", "city", args["city"])
			return map[string]string{"id": "HTL-999", "status": "reserved"}, nil
		},
		"cancel_hotel": func(args map[string]any) (any, error) {
			logger.Info("🏨 UNDO: Cancelling Hotel", "id", args["id"])
			return "cancelled", nil
		},
		"book_flight": func(args map[string]any) (any, error) {
			logger.Info("✈️  ACT: Booking Flight", "destination", args["destination"])
			return map[string]string{"id": "FLT-888", "status": "confirmed"}, nil
		},
		"cancel_flight": func(args map[string]any) (any, error) {
			logger.Info("✈️  UNDO: Cancelling Flight", "id", args["id"])
			return "cancelled", nil
		},
		"reject_request": func(args map[string]any) (any, error) {
			logger.Info("🚫 ACT: Rejecting Request... Triggering Rollback")
			return "rejected", nil
		},
	}

	// 4. Initialize Engine
	repoPath := filepath.Join(wd, "examples", "compensation-durable")
	eng, err := trellis.New(repoPath,
		trellis.WithLogger(logger),
		trellis.WithLifecycleHooks(domain.LifecycleHooks{
			OnToolCall: func(ctx context.Context, ev *domain.ToolEvent) {
				logger.Info("🛠️  ENGINE: Tool Called", "tool", ev.ToolName)
			},
			OnToolReturn: func(ctx context.Context, ev *domain.ToolEvent) {
				if ev.IsError {
					logger.Error("❌ ENGINE: Tool Failed", "tool", ev.ToolName, "error", ev.Output)
				} else {
					logger.Info("✅ ENGINE: Tool Returned", "tool", ev.ToolName, "output", ev.Output)
				}
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	// 5. Run Logic
	sessionID := "durable-saga-session"
	ctx := context.Background()

	// Tool Middleware setup
	handler := runner.NewTextHandler(os.Stdout)
	toolHandler := &ToolMiddleware{Next: handler, Tools: tools}

	r := runner.NewRunner(
		runner.WithInputHandler(toolHandler),
		// Use the session manager as the store (it implements StateStore)
		runner.WithStore(sessMgr),
		runner.WithSessionID(sessionID),
		runner.WithHeadless(true), // Auto-approve tools
	)

	fmt.Printf("\n--- DURABLE SAGA DEMO: Action='%s' ---\n", *action)

	switch *action {
	case "run":
		// Initial Run: Start the flow
		fmt.Println("🚀 Starting Flow... (Will pause at Approval)")

		if *fresh {
			fmt.Println("🧹 Force Fresh: Deleting existing session.")
			_ = sessMgr.Delete(ctx, sessionID) // Ignore error if not exists
		}

		// Load state if exists, or start new
		state, err := sessMgr.LoadOrStart(ctx, sessionID, "start")
		if err != nil {
			panic(err)
		}

		// Check if terminated (finished previously)
		if state.Status == domain.StatusTerminated {
			fmt.Println("🔄 Previous session finished. Starting fresh...")
			// Start new session
			state, err = eng.Start(ctx, sessionID, nil)
			if err != nil {
				panic(err)
			}
			// Save initial state so the runner picks it up (though LoadOrStart usually handles this, we are overriding)
			if err := sessMgr.Save(ctx, sessionID, state); err != nil {
				panic(err)
			}
		}

		// Run with the loaded/new state
		_, err = r.Run(ctx, eng, state)
		if err != nil {
			// Ignore interrupt error as we expect to stop at prompts or signals
			if err.Error() != "interrupt" && err.Error() != "interrupted" {
				fmt.Printf("Run finished with: %v\n", err)
			}
		}
		fmt.Println("\n💾 Session interrupted/paused.")

	case "approve":
		// Resume and signal approval
		fmt.Println("👍 Sending Approval Signal...")

		state, err := sessMgr.Load(ctx, sessionID)
		if err != nil {
			fmt.Printf("Error loading session (did you run --action=run first?): %v\n", err)
			return
		}

		// Apply Signal
		state, err = eng.Signal(ctx, state, "manager_approval")
		if err != nil {
			fmt.Printf("Error sending signal: %v\n", err)
			return
		}

		// Save state after signal (optional, but good for durability)
		if err := sessMgr.Save(ctx, sessionID, state); err != nil {
			panic(err)
		}

		fmt.Println("✅ Signal 'manager_approval' applied to state.")
		fmt.Println("💾 State saved. Run 'go run main.go' to RESUME execution from the next step.")

	case "reject":
		// Resume and signal rejection
		fmt.Println("👎 Sending Rejection Signal...")

		state, err := sessMgr.Load(ctx, sessionID)
		if err != nil {
			fmt.Printf("Error loading session (did you run --action=run first?): %v\n", err)
			return
		}

		// Apply Signal
		state, err = eng.Signal(ctx, state, "manager_rejection")
		if err != nil {
			fmt.Printf("Error sending signal: %v\n", err)
			return
		}

		// Save state after signal (optional, but good for durability)
		if err := sessMgr.Save(ctx, sessionID, state); err != nil {
			panic(err)
		}

		fmt.Println("✅ Signal 'manager_rejection' applied to state.")
		fmt.Println("💾 State saved. Run 'go run main.go' to RESUME execution (Rollback phase).")
	}
}

type ToolMiddleware struct {
	Next  runner.IOHandler
	Tools map[string]func(map[string]any) (any, error)
}

func (m *ToolMiddleware) Output(ctx context.Context, actions []domain.ActionRequest) (bool, error) {
	return m.Next.Output(ctx, actions)
}
func (m *ToolMiddleware) Input(ctx context.Context) (string, error) { return m.Next.Input(ctx) }
func (m *ToolMiddleware) SystemOutput(ctx context.Context, msg string) error {
	return m.Next.SystemOutput(ctx, msg)
}
func (m *ToolMiddleware) Signal(ctx context.Context, name string, args map[string]any) error {
	return m.Next.Signal(ctx, name, args)
}
func (m *ToolMiddleware) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	fn, ok := m.Tools[call.Name]
	if !ok {
		return domain.ToolResult{ID: call.ID, IsError: true, Result: "tool not found"}, nil
	}
	time.Sleep(100 * time.Millisecond)
	res, err := fn(call.Args)
	if err != nil {
		return domain.ToolResult{ID: call.ID, IsError: true, Result: err.Error()}, nil
	}
	return domain.ToolResult{ID: call.ID, IsError: false, Result: res}, nil
}

type AutoInputMiddleware struct {
	Next          runner.IOHandler
	InputToInject string
}

// Delegate methods...
func (m *AutoInputMiddleware) Output(ctx context.Context, a []domain.ActionRequest) (bool, error) {
	return m.Next.Output(ctx, a)
}
func (m *AutoInputMiddleware) SystemOutput(ctx context.Context, msg string) error {
	return m.Next.SystemOutput(ctx, msg)
}
func (m *AutoInputMiddleware) Signal(ctx context.Context, name string, args map[string]any) error {
	return m.Next.Signal(ctx, name, args)
}
func (m *AutoInputMiddleware) HandleTool(ctx context.Context, c domain.ToolCall) (domain.ToolResult, error) {
	return m.Next.HandleTool(ctx, c)
}
func (m *AutoInputMiddleware) Input(ctx context.Context) (string, error) {
	// Auto inject input
	fmt.Printf("(Auto-Input: %s)\n", m.InputToInject)
	// We only inject once. Subsequent calls should maybe eof or block?
	// For demo, just returning it immediately.
	return m.InputToInject, nil
}
