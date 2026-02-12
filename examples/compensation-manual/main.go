package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/adapters/file"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/registry"
	"github.com/aretw0/trellis/pkg/runner"
)

// Mock Tools
func registerMockTools(reg *registry.Registry) {
	tools := map[string]registry.ToolFunction{
		"book_flight": func(ctx context.Context, args map[string]any) (any, error) {
			return "Flight-123", nil
		},
		"book_hotel": func(ctx context.Context, args map[string]any) (any, error) {
			return "Hotel-456", nil
		},
		"book_car": func(ctx context.Context, args map[string]any) (any, error) {
			// Simulate Failure
			// Returning an error here causes the Engine to check for 'on_error' transition
			return nil, fmt.Errorf("Service Unavailable: Car rental system down")
		},
		"cancel_car": func(ctx context.Context, args map[string]any) (any, error) {
			return "Car Cancelled", nil
		},
		"cancel_hotel": func(ctx context.Context, args map[string]any) (any, error) {
			return "Hotel Cancelled", nil
		},
		"cancel_flight": func(ctx context.Context, args map[string]any) (any, error) {
			return "Flight Cancelled", nil
		},
	}

	for name, fn := range tools {
		reg.Register(name, fn)
	}
}

func main() {
	// Setup Logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Setup Dependencies
	wd, _ := os.Getwd()
	flowDir := filepath.Join(wd, "examples", "compensation-manual")

	// 1. Initialize Engine
	// Define hooks for observability
	hooks := domain.LifecycleHooks{
		OnNodeEnter: func(ctx context.Context, e *domain.NodeEvent) {
			fmt.Printf("DEBUG: Entering Node %s (Type: %s)\n", e.NodeID, e.NodeType)
		},
		OnToolCall: func(ctx context.Context, e *domain.ToolEvent) {
			fmt.Printf("DEBUG: Calling Tool %s\n", e.ToolName)
		},
		OnToolReturn: func(ctx context.Context, e *domain.ToolEvent) {
			fmt.Printf("DEBUG: Tool Returned %s (Error: %v)\n", e.ToolName, e.IsError)
		},
	}

	engine, err := trellis.New(flowDir,
		trellis.WithLogger(logger),
		trellis.WithLifecycleHooks(hooks),
	)
	if err != nil {
		fmt.Printf("Failed to init engine: %v\n", err)
		os.Exit(1)
	}

	// 2. Setup Persistence
	store := file.New(filepath.Join(".trellis", "sessions"))

	// 3. Setup Tools Registry
	reg := registry.NewRegistry()
	registerMockTools(reg)

	// 4. Setup Handler with Registry
	handler := runner.NewTextHandler(
		os.Stdout,
		runner.WithTextHandlerRegistry(reg),
	)

	// 5. Initialize Runner
	r := runner.NewRunner(
		runner.WithInputHandler(handler),
		runner.WithStore(store),
		runner.WithLogger(logger),
	)

	// 6. Run Session
	sessionID := fmt.Sprintf("saga-demo-%d", time.Now().Unix())

	fmt.Printf("--- SAGA DEMO (Session: %s) ---\n", sessionID)
	fmt.Println("This demo simulates a booking flow failure and subsequent rollback.")

	_, err = r.Run(context.Background(), engine, nil)
	if err != nil {
		fmt.Printf("Execution Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- DEMO COMPLETE ---")
}
