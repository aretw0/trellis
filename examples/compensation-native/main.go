package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

func main() {
	// 1. Setup Logger (Structured for visibility)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// 2. Define Mock Tools
	tools := map[string]func(map[string]any) (any, error){
		"reserve_hotel": func(args map[string]any) (any, error) {
			logger.Info("🏨 ACT: Reserving Hotel", "city", args["city"])
			return map[string]string{"id": "HTL-123", "status": "reserved"}, nil
		},
		"cancel_hotel": func(args map[string]any) (any, error) {
			logger.Info("🏨 UNDO: Cancelling Hotel", "id", args["id"])
			return "cancelled", nil
		},
		"book_flight": func(args map[string]any) (any, error) {
			logger.Info("✈️  ACT: Booking Flight", "destination", args["destination"])
			return map[string]string{"id": "FLT-456", "status": "confirmed"}, nil
		},
		"cancel_flight": func(args map[string]any) (any, error) {
			logger.Info("✈️  UNDO: Cancelling Flight", "id", args["id"])
			return "cancelled", nil
		},
		"rent_car": func(args map[string]any) (any, error) {
			logger.Info("🚗 ACT: Renting Car... 💥")
			return nil, errors.New("car rental service unavailable")
		},
	}

	// 3. Create Custom Tool Handler
	handler := runner.NewTextHandler(os.Stdin, os.Stdout)
	// We wrap the default handler to inject our tools
	toolHandler := &ToolMiddleware{
		Next:  handler,
		Tools: tools,
	}

	// 4. Initialize Engine
	wd, _ := os.Getwd()
	repoPath := filepath.Join(wd, "examples", "compensation-native")
	eng, err := trellis.New(repoPath,
		trellis.WithLifecycleHooks(domain.LifecycleHooks{
			OnToolCall: func(ctx context.Context, ev *domain.ToolEvent) {
				logger.Info("🛠️  ENGINE: Tool Called", "tool", ev.ToolName, "input", ev.Input)
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

	// 5. Run
	fmt.Println("--- STARTING NATIVE SAGA DEMO ---")
	r := runner.NewRunner(
		runner.WithInputHandler(toolHandler),
		runner.WithHeadless(true),
	)

	_, err = r.Run(context.Background(), eng, nil)
	if err != nil {
		fmt.Printf("\n❌ Demo Finished with error: %v\n", err)
	} else {
		fmt.Println("\n✅ Demo Finished Successfully")
	}
}

// ToolMiddleware intercepts tool calls to run our mock functions
type ToolMiddleware struct {
	Next  runner.IOHandler
	Tools map[string]func(map[string]any) (any, error)
}

func (m *ToolMiddleware) Output(ctx context.Context, actions []domain.ActionRequest) (bool, error) {
	return m.Next.Output(ctx, actions)
}
func (m *ToolMiddleware) Input(ctx context.Context) (string, error) {
	return m.Next.Input(ctx)
}
func (m *ToolMiddleware) SystemOutput(ctx context.Context, msg string) error {
	return m.Next.SystemOutput(ctx, msg)
}

func (m *ToolMiddleware) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	fn, ok := m.Tools[call.Name]
	if !ok {
		return domain.ToolResult{ID: call.ID, IsError: true, Result: fmt.Sprintf("tool %s not found", call.Name)}, nil
	}

	// Simulate latency
	time.Sleep(500 * time.Millisecond)

	res, err := fn(call.Args)
	if err != nil {
		return domain.ToolResult{ID: call.ID, IsError: true, Result: err.Error()}, nil
	}
	return domain.ToolResult{ID: call.ID, IsError: false, Result: res}, nil
}
