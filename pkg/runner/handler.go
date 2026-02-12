package runner

import (
	"context"

	"github.com/aretw0/trellis/pkg/domain"
)

// IOHandler defines the strategy for interacting with the user.
// This allows switching between Text (CLI/TUI) and JSON (Structured) modes.
type IOHandler interface {
	// Output presents the actions to the user.
	// Returns true if the output requires user input (e.g. asking a question),
	// or if the handler expects to read input after this.
	Output(ctx context.Context, actions []domain.ActionRequest) (bool, error)

	// Input reads a response from the user.
	Input(ctx context.Context) (string, error)

	// Signal notifies the handler of an event (e.g. "thinking", "typing").
	// This is used for visual feedback without blocking input.
	Signal(ctx context.Context, name string, args map[string]any) error

	// HandleTool executes a side-effect requested by the engine.
	// In a text/CLI context, this might just log the request or ask for confirmation.
	HandleTool(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error)

	// SystemOutput presents a meta-message to the user (e.g. system logs, status updates).
	// This is distinct from content rendering.
	SystemOutput(ctx context.Context, msg string) error
}

// ToolRunner defines the strategy for executing side-effects.
// It decouples execution logic (Process, HTTP, MCP) from IO logic (Text, JSON).
type ToolRunner interface {
	Execute(ctx context.Context, call domain.ToolCall) (domain.ToolResult, error)
}
