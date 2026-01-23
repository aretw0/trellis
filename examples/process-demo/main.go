package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/internal/adapters/process"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
)

func main() {
	// 1. Initialize Process Adapter
	processRunner := process.NewRunner()

	// Register a "safe" script.
	// For portability, we'll just use 'echo' if on linux, or 'cmd /c echo' on windows?
	// To make it truly cross-platform for this demo, let's try to run a go "script" (go run).
	// Actually, let's keep it simple: "echo" is standard enough (except on Windows usually requires cmd /c).
	// Let's assume the user has a modern shell or we use go run.

	// Let's create a temporary script file in the example to run.
	processRunner.Register("hello_script", "go", "run", "examples/process-demo/script/main.go")

	// 2. Custom Handler that delegates to Process Runner
	// We wrap standard IOHandler to intercept tools.
	stdHandler := runner.NewTextHandler(os.Stdin, os.Stdout)

	// This is a naive implementation: normally we'd have a router.
	// We'll use a ToolInterceptor or just wrap HandleTool.

	handler := &ProcessAwareHandler{
		IOHandler: stdHandler,
		Process:   processRunner,
	}

	// 3. Setup Engine
	r := runner.NewRunner(
		runner.WithInputHandler(handler),
	)

	// 4. Run Flow
	// We'll build the loader manually or load from file.
	// Let's assume loading from local dir.
	engine, err := trellis.New("examples/process-demo")
	if err != nil {
		panic(err)
	}

	state, err := r.Run(context.Background(), engine, nil)
	if err != nil {
		slog.Error("Run failed", "err", err)
		os.Exit(1)
	}

	if state.Status == domain.StatusTerminated {
		os.Exit(0)
	}
}

// ProcessAwareHandler delegates tool calls.
type ProcessAwareHandler struct {
	runner.IOHandler
	Process *process.Runner
}

func (h *ProcessAwareHandler) HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error) {
	// Naive router: check if Process runner handles it
	// In reality we should check if metadata has "type: process" or similar.
	// For now, we try process runner first.
	result, err := h.Process.Execute(ctx, call)
	if err == nil && !result.IsError {
		return result, nil
	}

	// Fallback to default (simulated)
	return h.IOHandler.HandleTool(ctx, call)
}
